/*
 * Copyright The Kubernetes Authors.
 * Copyright The Project-HAMi Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/klog/v2"

	pkggcu "github.com/Project-HAMi/enflame-dra-driver/pkg/gcu"
	"github.com/Project-HAMi/enflame-dra-driver/pkg/metrics"
)

type driver struct {
	client coreclientset.Interface
	helper *kubeletplugin.Helper
	state  *DeviceState
	// healthcheck is the driver-process gRPC liveness probe. It is unrelated to
	// the per-device KEP-4680 health reporting below.
	healthcheck *healthcheck
	cancelCtx   func(error)

	config   *Config
	poolName string
	hostRoot string

	// deviceHealth reflects the --device-health flag. When false the driver
	// advertises no health service, starts no health goroutines, and
	// WatchHealthStatus returns ErrHealthNotSupported.
	deviceHealth bool
	// devices holds the device names, sorted, for stable report iteration.
	// It is immutable after startup and so needs no lock.
	devices []string
	// checkHealth is overwritten in unit tests.
	checkHealth func(index int) pkggcu.Health

	// clientsMu guards subscribers.
	clientsMu sync.RWMutex
	// subscribers are per-WatchHealthStatus wake channels. A send is a
	// coalescing "rebuild and resend" signal rather than the report itself.
	subscribers map[chan struct{}]struct{}

	stopHealthCh   chan struct{}
	stopHealthOnce sync.Once
	healthWg       sync.WaitGroup

	healthPollInterval time.Duration
}

const (
	deviceHealthPollInterval = 30 * time.Second
	deviceHealthCheckTimeout = 60 * time.Second
)

func NewDriver(ctx context.Context, config *Config) (*driver, error) {
	driver := &driver{
		client:             config.coreclient,
		cancelCtx:          config.cancelMainCtx,
		config:             config,
		poolName:           config.flags.nodeName,
		hostRoot:           config.flags.hostRoot,
		subscribers:        make(map[chan struct{}]struct{}),
		stopHealthCh:       make(chan struct{}),
		healthPollInterval: deviceHealthPollInterval,
	}
	driver.checkHealth = func(index int) pkggcu.Health {
		return pkggcu.CheckHealth(index, pkggcu.WithHostRoot(driver.hostRoot))
	}

	state, err := NewDeviceState(config)
	if err != nil {
		return nil, err
	}
	driver.state = state

	logger := klog.FromContext(ctx)
	driver.deviceHealth = config.flags.deviceHealth
	if driver.deviceHealth {
		for deviceName := range state.allocatable {
			driver.devices = append(driver.devices, deviceName)
		}
		sort.Strings(driver.devices)
		logger.Info("Device health reporting enabled", "devices", len(driver.devices))
	} else {
		logger.Info("Device health reporting disabled; WatchHealthStatus will report ErrHealthNotSupported",
			"flag", "--device-health=false")
	}

	helper, err := kubeletplugin.Start(ctx, driver,
		kubeletplugin.KubeClient(config.coreclient),
		kubeletplugin.NodeName(config.flags.nodeName),
		kubeletplugin.DriverName(config.flags.driverName),
		kubeletplugin.RegistrarDirectoryPath(config.flags.kubeletRegistrarDirectoryPath),
		kubeletplugin.PluginDataDirectoryPath(config.DriverPluginPath()),
		kubeletplugin.RollingUpdate(types.UID(config.flags.podUID)),
		kubeletplugin.HealthService(config.flags.deviceHealth),
	)
	if err != nil {
		return nil, err
	}
	driver.helper = helper

	driver.healthcheck, err = startHealthcheck(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("start healthcheck: %w", err)
	}

	if err := helper.PublishResources(ctx, state.driverResources); err != nil {
		return nil, err
	}

	if driver.deviceHealth {
		driver.healthWg.Add(1)
		go driver.deviceHealthLoop(ctx)
	}

	return driver, nil
}

func (d *driver) Shutdown(logger klog.Logger) error {
	if d.healthcheck != nil {
		d.healthcheck.Stop(logger)
	}

	if d.deviceHealth {
		logger.Info("Stopping device health monitoring")
		d.stopHealthOnce.Do(func() { close(d.stopHealthCh) })
		d.healthWg.Wait()
	}

	d.helper.Stop()
	return nil
}

func (d *driver) PrepareResourceClaims(ctx context.Context, claims []*resourceapi.ResourceClaim) (map[types.UID]kubeletplugin.PrepareResult, error) {
	logger := klog.FromContext(ctx)
	logger.Info("PrepareResourceClaims is called", "numClaims", len(claims))
	result := make(map[types.UID]kubeletplugin.PrepareResult)

	for _, claim := range claims {
		result[claim.UID] = d.prepareResourceClaim(ctx, claim)
	}

	return result, nil
}

func (d *driver) prepareResourceClaim(ctx context.Context, claim *resourceapi.ResourceClaim) (result kubeletplugin.PrepareResult) {
	logger := klog.FromContext(ctx)
	logger.Info("Preparing claim", "uid", claim.UID, "namespace", claim.Namespace, "name", claim.Name)

	start := time.Now()
	defer func() {
		metrics.ObservePrepareClaim(result.Err, time.Since(start))
	}()

	preparedDevices, err := d.state.Prepare(ctx, claim)
	if err != nil {
		logger.Error(err, "Error preparing devices for claim", "uid", claim.UID)
		result = kubeletplugin.PrepareResult{
			Err: fmt.Errorf("error preparing devices for claim %v: %w", claim.UID, err),
		}
		return result
	}
	var prepared []kubeletplugin.Device
	for _, preparedDevice := range preparedDevices {
		prepared = append(prepared, kubeletplugin.Device{
			Requests:     preparedDevice.GetRequestNames(),
			PoolName:     preparedDevice.GetPoolName(),
			DeviceName:   preparedDevice.GetDeviceName(),
			CDIDeviceIDs: preparedDevice.GetCdiDeviceIds(),
			ShareID:      preparedDevice.ShareID,
		})
	}

	logger.Info("Returning newly prepared devices for claim", "uid", claim.UID, "devices", prepared)
	result = kubeletplugin.PrepareResult{Devices: prepared}
	return result
}

func (d *driver) UnprepareResourceClaims(ctx context.Context, claims []kubeletplugin.NamespacedObject) (map[types.UID]error, error) {
	logger := klog.FromContext(ctx)
	logger.Info("UnprepareResourceClaims is called", "numClaims", len(claims))
	result := make(map[types.UID]error)

	for _, claim := range claims {
		result[claim.UID] = d.unprepareResourceClaim(ctx, claim)
	}

	return result, nil
}

func (d *driver) unprepareResourceClaim(_ context.Context, claim kubeletplugin.NamespacedObject) (err error) {
	start := time.Now()
	defer func() {
		metrics.ObserveUnprepareClaim(err, time.Since(start))
	}()

	if err = d.state.Unprepare(claim.UID); err != nil {
		return fmt.Errorf("error unpreparing devices for claim %v: %w", claim.UID, err)
	}

	return nil
}

func (d *driver) HandleError(ctx context.Context, err error, msg string) {
	utilruntime.HandleErrorWithContext(ctx, err, msg)
	if !errors.Is(err, kubeletplugin.ErrRecoverable) {
		metrics.FatalBackgroundErrorsTotal.Inc()
		if d.cancelCtx != nil {
			d.cancelCtx(fmt.Errorf("fatal background error: %w", err))
		}
	}
}

func (d *driver) deviceHealthLoop(ctx context.Context) {
	defer d.healthWg.Done()

	logger := klog.FromContext(ctx)
	ticker := time.NewTicker(d.healthPollInterval)
	defer ticker.Stop()

	logger.Info("Starting device health monitoring loop")
	d.notifySubscribers()

	for {
		select {
		case <-d.stopHealthCh:
			logger.Info("Health monitoring loop stopped")
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Refresh reports so the kubelet does not mark devices stale, and
			// pick up sysfs device_health / device_state changes.
			d.notifySubscribers()
		}
	}
}

func (d *driver) buildHealthReport() kubeletplugin.DeviceHealthReport {
	now := time.Now()
	devices := make([]kubeletplugin.DeviceHealth, 0, len(d.devices))
	for _, deviceName := range d.devices {
		devices = append(devices, d.deviceHealthEntry(deviceName, now))
	}
	return kubeletplugin.DeviceHealthReport{Devices: devices}
}

func (d *driver) deviceHealthEntry(deviceName string, now time.Time) kubeletplugin.DeviceHealth {
	entry := kubeletplugin.DeviceHealth{
		PoolName:           d.poolName,
		DeviceName:         deviceName,
		LastUpdated:        now,
		HealthCheckTimeout: deviceHealthCheckTimeout,
		Health:             kubeletplugin.HealthStatusUnknown,
		Message:            "unknown",
	}

	minor, err := pkggcu.ParseDRAName(deviceName)
	if err != nil {
		entry.Message = err.Error()
		return entry
	}

	h := d.checkHealth(minor)
	entry.Message = h.Message
	switch h.Level {
	case pkggcu.HealthLevelHealthy:
		entry.Health = kubeletplugin.HealthStatusHealthy
	case pkggcu.HealthLevelUnhealthy:
		entry.Health = kubeletplugin.HealthStatusUnhealthy
	default:
		entry.Health = kubeletplugin.HealthStatusUnknown
	}
	return entry
}

func (d *driver) notifySubscribers() {
	d.clientsMu.RLock()
	defer d.clientsMu.RUnlock()
	for sig := range d.subscribers {
		select {
		case sig <- struct{}{}:
		default:
		}
	}
}

// WatchHealthStatus implements [kubeletplugin.DRAPlugin].
func (d *driver) WatchHealthStatus(ctx context.Context, reports chan<- kubeletplugin.DeviceHealthReport) error {
	if !d.deviceHealth {
		return kubeletplugin.ErrHealthNotSupported
	}

	logger := klog.FromContext(ctx)
	logger.Info("New health monitoring client connected")

	sig := make(chan struct{}, 1)
	d.clientsMu.Lock()
	d.subscribers[sig] = struct{}{}
	d.clientsMu.Unlock()

	defer func() {
		d.clientsMu.Lock()
		delete(d.subscribers, sig)
		d.clientsMu.Unlock()
		logger.Info("Health monitoring client disconnected")
	}()

	if d.sendReport(ctx, reports) {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-d.stopHealthCh:
			return nil
		case <-sig:
			if d.sendReport(ctx, reports) {
				return nil
			}
		}
	}
}

func (d *driver) sendReport(ctx context.Context, reports chan<- kubeletplugin.DeviceHealthReport) (stopped bool) {
	report := d.buildHealthReport()
	select {
	case <-ctx.Done():
		return true
	case <-d.stopHealthCh:
		return true
	case reports <- report:
		return false
	}
}
