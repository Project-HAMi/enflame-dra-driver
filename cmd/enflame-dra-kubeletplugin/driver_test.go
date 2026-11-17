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
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/klog/v2"

	"github.com/Project-HAMi/enflame-dra-driver/internal/profiles/gcu"
	pkggcu "github.com/Project-HAMi/enflame-dra-driver/pkg/gcu"
)

const (
	testNodeName   = "test-node"
	testDriverName = gcu.DefaultDriverName
)

func newDriverTestConfig(t *testing.T, deviceHealth bool, healthcheckPort int) (*Config, *error) {
	t.Helper()

	tmp, err := os.MkdirTemp("/tmp", "dra")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })

	flags := &Flags{
		nodeName:                      testNodeName,
		driverName:                    testDriverName,
		profile:                       gcu.ProfileName,
		cdiRoot:                       filepath.Join(tmp, "cdi"),
		hostRoot:                      setupFakeHostRootAt(t, tmp, 1),
		kubeletRegistrarDirectoryPath: filepath.Join(tmp, "r"),
		kubeletPluginsDirectoryPath:   filepath.Join(tmp, "p"),
		healthcheckPort:               healthcheckPort,
		deviceHealth:                  deviceHealth,
		numDevices:                    1,
	}
	useStubGCUDiscover(t, flags.numDevices)
	for _, dir := range []string{flags.cdiRoot, flags.kubeletRegistrarDirectoryPath, filepath.Join(flags.kubeletPluginsDirectoryPath, testDriverName)} {
		require.NoError(t, os.MkdirAll(dir, 0750))
	}

	var fatal error
	return &Config{
		flags:         flags,
		coreclient:    fakeclient.NewClientset(),
		cancelMainCtx: func(err error) { fatal = err },
		profile:       gcu.NewProfile(testNodeName, flags.hostRoot, flags.numDevices),
	}, &fatal
}

func receiveReport(t *testing.T, reports <-chan kubeletplugin.DeviceHealthReport) kubeletplugin.DeviceHealthReport {
	t.Helper()
	select {
	case report := <-reports:
		return report
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for health report")
		return kubeletplugin.DeviceHealthReport{}
	}
}

func healthOf(report kubeletplugin.DeviceHealthReport, device string) kubeletplugin.HealthStatus {
	for _, d := range report.Devices {
		if d.DeviceName == device {
			return d.Health
		}
	}
	return ""
}

func waitForHealthGoroutines(t *testing.T, d *driver) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		d.healthWg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("health goroutines did not exit")
	}
}

func TestLivenessRegistrationSocketName(t *testing.T) {
	const podUID = "f358c92b-e974-4f29-a84f-6c64bec0949d"

	for name, tc := range map[string]struct {
		podUID         string
		registrySubdir string
		wantHashed     bool
	}{
		"without pod UID": {},
		"with pod UID": {
			podUID: podUID,
		},
		"with pod UID and registration path too long for the full UID": {
			podUID:         podUID,
			registrySubdir: "long-registration-directory-for-liveness",
			wantHashed:     true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			config, _ := newDriverTestConfig(t, false, 0)
			config.flags.podUID = tc.podUID
			registryDir := filepath.Join(config.flags.kubeletRegistrarDirectoryPath, tc.registrySubdir)
			config.flags.kubeletRegistrarDirectoryPath = registryDir
			require.NoError(t, os.MkdirAll(registryDir, 0750))

			unhashedFilename := testDriverName + "-reg.sock"
			if tc.podUID != "" {
				unhashedFilename = testDriverName + "-" + tc.podUID + "-reg.sock"
			}
			if tc.wantHashed {
				require.Greater(t, len(filepath.Join(registryDir, unhashedFilename)), 108,
					"the path must exceed the helper's limit to trigger a hashed filename")
			}

			d, err := NewDriver(ctx, config)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, d.Shutdown(klog.FromContext(ctx))) })

			entries, err := os.ReadDir(registryDir)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			if tc.wantHashed {
				assert.NotEqual(t, unhashedFilename, entries[0].Name())
			} else {
				assert.Equal(t, unhashedFilename, entries[0].Name())
			}

			resp, err := d.healthcheck.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "liveness"})
			require.NoError(t, err)
			assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())
		})
	}
}

func TestNewDriverLifecycle(t *testing.T) {
	for name, deviceHealth := range map[string]bool{
		"device health enabled":  true,
		"device health disabled": false,
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			config, _ := newDriverTestConfig(t, deviceHealth, 0)

			d, err := NewDriver(ctx, config)
			require.NoError(t, err)
			require.NotNil(t, d.helper)
			require.NotNil(t, d.healthcheck)
			assert.Equal(t, deviceHealth, d.deviceHealth)

			resp, err := d.healthcheck.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "liveness"})
			require.NoError(t, err)
			assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())

			reports := make(chan kubeletplugin.DeviceHealthReport)
			done := make(chan error, 1)
			go func() { done <- d.WatchHealthStatus(ctx, reports) }()
			if deviceHealth {
				report := receiveReport(t, reports)
				assert.Equal(t, []string{"gcu-0"}, d.devices)
				assert.Equal(t, kubeletplugin.HealthStatusHealthy, healthOf(report, "gcu-0"))
			} else {
				assert.Empty(t, d.devices)
			}

			require.NoError(t, d.Shutdown(klog.FromContext(ctx)))

			select {
			case err := <-done:
				if deviceHealth {
					assert.NoError(t, err)
				} else {
					assert.ErrorIs(t, err, kubeletplugin.ErrHealthNotSupported)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("WatchHealthStatus did not return after Shutdown")
			}
			waitForHealthGoroutines(t, d)
		})
	}
}

func TestBuildHealthReportReflectsSysfs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, _ := newDriverTestConfig(t, true, -1)

	d, err := NewDriver(ctx, config)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, d.Shutdown(klog.FromContext(ctx))) })

	report := d.buildHealthReport()
	assert.Equal(t, kubeletplugin.HealthStatusHealthy, healthOf(report, "gcu-0"))

	d.checkHealth = func(index int) pkggcu.Health {
		return pkggcu.Health{
			Level:   pkggcu.HealthLevelUnhealthy,
			Health:  "unhealthy",
			State:   "working",
			Message: "device_health=\"unhealthy\" device_state=\"working\"",
		}
	}
	report = d.buildHealthReport()
	assert.Equal(t, kubeletplugin.HealthStatusUnhealthy, healthOf(report, "gcu-0"))
}

func TestNewDriverWithoutHealthcheck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, _ := newDriverTestConfig(t, true, -1)

	d, err := NewDriver(ctx, config)
	require.NoError(t, err)
	assert.Nil(t, d.healthcheck, "a negative port disables the healthcheck service")
	require.NoError(t, d.Shutdown(klog.FromContext(ctx)))
}

func TestNewDriverErrors(t *testing.T) {
	for name, breakConfig := range map[string]func(t *testing.T, config *Config){
		"device state cannot be created": func(t *testing.T, config *Config) {
			require.NoError(t, os.RemoveAll(config.flags.cdiRoot))
			require.NoError(t, os.WriteFile(config.flags.cdiRoot, nil, 0600))
		},
		"kubeletplugin helper cannot start": func(_ *testing.T, config *Config) {
			config.flags.driverName = ""
		},
		"healthcheck port in use": func(t *testing.T, config *Config) {
			lis, err := net.Listen("tcp", ":0")
			require.NoError(t, err)
			t.Cleanup(func() { _ = lis.Close() })
			addr, ok := lis.Addr().(*net.TCPAddr)
			require.True(t, ok)
			config.flags.healthcheckPort = addr.Port
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			config, _ := newDriverTestConfig(t, true, 0)
			breakConfig(t, config)

			_, err := NewDriver(ctx, config)
			require.Error(t, err)
		})
	}
}

func TestHealthcheckNotServingWithoutPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, _ := newDriverTestConfig(t, true, 0)

	h, err := startHealthcheck(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, h)
	defer h.Stop(klog.FromContext(ctx))

	for _, service := range []string{"", "liveness"} {
		resp, err := h.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
		require.NoError(t, err)
		assert.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, resp.GetStatus(), "service %q", service)
	}

	_, err = h.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "bogus"})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestStartHealthcheckFailsOnUsedPort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, _ := newDriverTestConfig(t, true, 0)

	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer func() { _ = lis.Close() }()
	addr, ok := lis.Addr().(*net.TCPAddr)
	require.True(t, ok)
	config.flags.healthcheckPort = addr.Port

	_, err = startHealthcheck(ctx, config)
	require.Error(t, err)
}

func TestPrepareAndUnprepareResourceClaims(t *testing.T) {
	ctx := context.Background()
	config, _ := newDriverTestConfig(t, true, -1)
	state, err := NewDeviceState(config)
	require.NoError(t, err)
	d := &driver{state: state, poolName: testNodeName}

	claim := func(uid types.UID, device string) *resourceapi.ResourceClaim {
		return &resourceapi.ResourceClaim{
			ObjectMeta: metav1.ObjectMeta{UID: uid, Namespace: "default", Name: string(uid)},
			Status: resourceapi.ResourceClaimStatus{
				Allocation: &resourceapi.AllocationResult{
					Devices: resourceapi.DeviceAllocationResult{
						Results: []resourceapi.DeviceRequestAllocationResult{{
							Request: "gcu",
							Driver:  testDriverName,
							Pool:    testNodeName,
							Device:  device,
						}},
					},
				},
			},
		}
	}

	good := claim("claim-good", "gcu-0")
	bad := claim("claim-bad", "gcu-9")

	results, err := d.PrepareResourceClaims(ctx, []*resourceapi.ResourceClaim{good, bad})
	require.NoError(t, err)
	require.Len(t, results, 2)

	require.NoError(t, results[good.UID].Err)
	require.Len(t, results[good.UID].Devices, 1)
	assert.Equal(t, "gcu-0", results[good.UID].Devices[0].DeviceName)
	assert.Equal(t, testNodeName, results[good.UID].Devices[0].PoolName)
	assert.Equal(t, []string{"gcu"}, results[good.UID].Devices[0].Requests)
	assert.NotEmpty(t, results[good.UID].Devices[0].CDIDeviceIDs)

	require.Error(t, results[bad.UID].Err)
	assert.Empty(t, results[bad.UID].Devices)

	unprepare := func() map[types.UID]error {
		errs, err := d.UnprepareResourceClaims(ctx, []kubeletplugin.NamespacedObject{
			{UID: good.UID, NamespacedName: types.NamespacedName{Namespace: good.Namespace, Name: good.Name}},
		})
		require.NoError(t, err)
		require.Len(t, errs, 1)
		return errs
	}

	checkpoint, err := os.ReadFile(state.checkpointPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(state.checkpointPath, []byte("{not json"), 0600))
	require.Error(t, unprepare()[good.UID])

	require.NoError(t, os.WriteFile(state.checkpointPath, checkpoint, 0600))
	assert.NoError(t, unprepare()[good.UID])
}

func TestHandleError(t *testing.T) {
	var fatal error
	d := &driver{cancelCtx: func(err error) { fatal = err }}
	ctx := context.Background()

	d.HandleError(ctx, fmt.Errorf("transient: %w", kubeletplugin.ErrRecoverable), "recoverable")
	assert.NoError(t, fatal, "recoverable errors must not stop the driver")

	boom := errors.New("boom")
	d.HandleError(ctx, boom, "fatal")
	require.ErrorIs(t, fatal, boom, "fatal errors must cancel the main context")

	(&driver{}).HandleError(ctx, boom, "fatal")
}
