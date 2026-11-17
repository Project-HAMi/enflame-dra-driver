/*
 * Copyright 2026 The Kubernetes Authors.
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
 * See the License for the specific license governing permissions and
 * limitations under the License.
 */

package cdi

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
	cdiparser "tags.cncf.io/container-device-interface/pkg/parser"
	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

const (
	// BaseDeviceName is the CDI device name for common GCU container edits
	// (gcuctl, firmware, efsmi, libefml). Every claim references this device.
	BaseDeviceName = "base"
)

var nonWord = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// Device describes one allocated GCU to inject into a claim's CDI spec.
type Device struct {
	// Name is the DRA device name (for example "gcu-0").
	Name string
	// Index is the char-device minor for /dev/gcu<Index> (DRA name gcu-N).
	// A negative value skips host device-node injection (mock / unit tests).
	Index int
	// VideoNodes controls injection of /dev/gcu<Index>vid* nodes.
	VideoNodes bool
	// ExtraEdits are additional container edits from opaque claim config.
	ExtraEdits *cdiapi.ContainerEdits
}

// Manager writes Enflame GCU CDI specs for the DRA kubelet plugin.
type Manager struct {
	cache      *cdiapi.Cache
	cdiRoot    string
	hostRoot   string
	devRoot    string
	driverName string
	class      string
	nodeName   string
}

// NewManager creates a CDI manager. Specs are written under the configured CDI root.
func NewManager(opts ...Option) (*Manager, error) {
	m := &Manager{
		cdiRoot:    defaultCDIRoot,
		hostRoot:   defaultHostRoot,
		devRoot:    defaultDevRoot,
		driverName: defaultDriverName,
		class:      defaultClass,
		nodeName:   os.Getenv("NODE_NAME"),
	}
	for _, opt := range opts {
		opt(m)
	}

	if err := os.MkdirAll(m.cdiRoot, 0o750); err != nil {
		return nil, fmt.Errorf("create CDI root %q: %w", m.cdiRoot, err)
	}

	cache, err := cdiapi.NewCache(cdiapi.WithSpecDirs(m.cdiRoot))
	if err != nil {
		return nil, fmt.Errorf("create CDI cache: %w", err)
	}
	m.cache = cache
	return m, nil
}

// Vendor returns the CDI vendor string (k8s.<driverName>).
func (m *Manager) Vendor() string {
	return "k8s." + m.driverName
}

// Class returns the CDI class.
func (m *Manager) Class() string {
	return m.class
}

// Kind returns the CDI kind (vendor/class).
func (m *Manager) Kind() string {
	return m.Vendor() + "/" + m.class
}

// DriverName returns the DRA driver name.
func (m *Manager) DriverName() string {
	return m.driverName
}

// CreateBaseSpec writes the common GCU dependency spec referenced by every claim.
func (m *Manager) CreateBaseSpec() error {
	edits, err := m.discoverBaseEdits()
	if err != nil {
		return err
	}

	edits.Env = append(edits.Env,
		fmt.Sprintf("KUBERNETES_NODE_NAME=%s", m.nodeName),
		fmt.Sprintf("DRA_RESOURCE_DRIVER_NAME=%s", m.driverName),
	)

	spec := &cdispec.Spec{
		Kind: m.Kind(),
		Devices: []cdispec.Device{
			{
				Name:           BaseDeviceName,
				ContainerEdits: edits,
			},
		},
	}

	minVersion, err := cdispec.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("minimum CDI version: %w", err)
	}
	spec.Version = minVersion

	specName, err := cdiapi.GenerateNameForTransientSpec(spec, BaseDeviceName)
	if err != nil {
		return fmt.Errorf("generate base spec name: %w", err)
	}

	if err := m.cache.WriteSpec(spec, specName); err != nil {
		return fmt.Errorf("write base CDI spec: %w", err)
	}
	klog.Background().V(4).Info("wrote GCU CDI base spec", "name", specName, "kind", m.Kind())
	return nil
}

// CreateClaimSpec writes a per-claim CDI spec for the allocated devices.
func (m *Manager) CreateClaimSpec(claimUID string, devices []Device) error {
	specName := cdiapi.GenerateTransientSpecName(m.Vendor(), m.class, claimUID)

	spec := &cdispec.Spec{
		Kind:    m.Kind(),
		Devices: []cdispec.Device{},
	}

	for _, device := range devices {
		deviceEnvKey := strings.ToUpper(nonWord.ReplaceAllString(device.Name, "_"))
		claimEdits := cdiapi.ContainerEdits{
			ContainerEdits: &cdispec.ContainerEdits{
				Env: []string{
					fmt.Sprintf("%s_DEVICE_%s_RESOURCE_CLAIM=%s", strings.ToUpper(m.class), deviceEnvKey, claimUID),
				},
			},
		}

		deviceEdits, err := m.discoverDeviceEdits(device.Index, device.VideoNodes)
		if err != nil {
			return fmt.Errorf("device %s: %w", device.Name, err)
		}
		if len(deviceEdits.DeviceNodes) > 0 || len(deviceEdits.Mounts) > 0 || len(deviceEdits.Env) > 0 {
			claimEdits.Append(&cdiapi.ContainerEdits{ContainerEdits: &deviceEdits})
		}
		claimEdits.Append(device.ExtraEdits)

		spec.Devices = append(spec.Devices, cdispec.Device{
			Name:           claimDeviceName(claimUID, device.Name),
			ContainerEdits: *claimEdits.ContainerEdits,
		})
	}

	minVersion, err := cdispec.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("minimum CDI version: %w", err)
	}
	spec.Version = minVersion

	if err := m.cache.WriteSpec(spec, specName); err != nil {
		return fmt.Errorf("write claim CDI spec: %w", err)
	}
	klog.Background().V(4).Info("wrote GCU CDI claim spec", "claimUID", claimUID, "devices", len(devices))
	return nil
}

// DeleteClaimSpec removes the CDI spec for a claim.
func (m *Manager) DeleteClaimSpec(claimUID string) error {
	specName := cdiapi.GenerateTransientSpecName(m.Vendor(), m.class, claimUID)
	if err := m.cache.RemoveSpec(specName); err != nil {
		return fmt.Errorf("remove claim CDI spec %q: %w", specName, err)
	}
	return nil
}

// ClaimDeviceIDs returns qualified CDI device IDs for a claim: the base device
// plus one entry per allocated device name.
func (m *Manager) ClaimDeviceIDs(claimUID string, names []string) []string {
	ids := []string{
		cdiparser.QualifiedName(m.Vendor(), m.class, BaseDeviceName),
	}
	for _, name := range names {
		ids = append(ids, cdiparser.QualifiedName(m.Vendor(), m.class, claimDeviceName(claimUID, name)))
	}
	return ids
}

// CleanupOrphanedClaimSpecs removes transient claim specs whose claim UIDs are
// not in active. The base spec is never removed.
func (m *Manager) CleanupOrphanedClaimSpecs(active sets.Set[string]) error {
	entries, err := os.ReadDir(m.cdiRoot)
	if err != nil {
		return fmt.Errorf("read CDI root: %w", err)
	}

	vendor := m.Vendor()
	class := m.class
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}
		specName := strings.TrimSuffix(name, ext)

		// Transient claim specs are named via GenerateTransientSpecName(vendor, class, claimUID).
		claimUID, ok := claimUIDFromTransientSpecName(specName, vendor, class)
		if !ok {
			continue
		}
		if active.Has(claimUID) {
			continue
		}
		klog.Background().Info("removing orphaned CDI claim spec", "spec", specName, "claimUID", claimUID)
		if err := m.cache.RemoveSpec(specName); err != nil {
			return fmt.Errorf("remove orphaned CDI spec %q: %w", specName, err)
		}
	}
	return nil
}

func claimDeviceName(claimUID, deviceName string) string {
	return fmt.Sprintf("%s-%s", claimUID, deviceName)
}

// claimUIDFromTransientSpecName extracts the claim UID from a name produced by
// cdiapi.GenerateTransientSpecName(vendor, class, claimUID).
func claimUIDFromTransientSpecName(specName, vendor, class string) (string, bool) {
	// GenerateTransientSpecName format: vendor_class_claimUID with '/' and '.' replaced.
	expectedKind := cdiapi.GenerateTransientSpecName(vendor, class, "PLACEHOLDER")
	prefix := strings.TrimSuffix(expectedKind, "PLACEHOLDER")
	if !strings.HasPrefix(specName, prefix) {
		return "", false
	}
	uid := strings.TrimPrefix(specName, prefix)
	if uid == "" || uid == BaseDeviceName {
		return "", false
	}
	return uid, true
}
