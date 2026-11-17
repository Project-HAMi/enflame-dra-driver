/*
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

package gcu

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/dynamic-resource-allocation/deviceattribute"
	"k8s.io/dynamic-resource-allocation/resourceslice"
	"k8s.io/utils/ptr"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"github.com/Project-HAMi/enflame-dra-driver/internal/profiles"
	"github.com/Project-HAMi/enflame-dra-driver/internal/profiles/helpers"
	pkggcu "github.com/Project-HAMi/enflame-dra-driver/pkg/gcu"
)

const (
	// ProfileName is the device profile name used by the kubelet plugin and CDI class.
	ProfileName = "gcu"
	// DefaultDriverName uniquely identifies this DRA driver in a cluster.
	DefaultDriverName = "gcu.enflame.com"
	// DeviceType is the ResourceSlice "type" attribute value (aligned with
	// HAMi nvidia-dra-driver's "gpu" / "hami-gpu" convention).
	DeviceType = "gcu"

	attrType            resourceapi.QualifiedName = "type"
	attrUUID            resourceapi.QualifiedName = "uuid"
	attrProductName     resourceapi.QualifiedName = "productName"
	attrBrand           resourceapi.QualifiedName = "brand"
	attrSerialNumber    resourceapi.QualifiedName = "serialNumber"
	attrPCIBusID        resourceapi.QualifiedName = "pcieBusID"
	attrDriverVersion   resourceapi.QualifiedName = "driverVersion"
	attrFirmwareVersion resourceapi.QualifiedName = "firmwareVersion"
	capacityMemory      resourceapi.QualifiedName = "memory"
	brandEnflame                                  = "Enflame"
)

// discoverFunc is overwritten in tests.
var discoverFunc = pkggcu.Discover

// Profile advertises GCU devices discovered through EFML.
type Profile struct {
	nodeName string
	hostRoot string
}

// NewProfile returns a GCU profile for the given node.
// stubCount is retained for call-site compatibility but ignored: discovery is EFML-only.
func NewProfile(nodeName, hostRoot string, stubCount int) Profile {
	_ = stubCount
	return Profile{
		nodeName: nodeName,
		hostRoot: hostRoot,
	}
}

// EnumerateDevices implements [profiles.Profile].
func (p Profile) EnumerateDevices() (resourceslice.DriverResources, error) {
	devices, err := discoverFunc(pkggcu.WithHostRoot(p.hostRoot))
	if err != nil {
		return resourceslice.DriverResources{}, err
	}

	out := make([]resourceapi.Device, 0, len(devices))
	for _, d := range devices {
		out = append(out, toResourceDevice(d))
	}

	return resourceslice.DriverResources{
		Pools: map[string]resourceslice.Pool{
			p.nodeName: {Slices: []resourceslice.Slice{{Devices: out}}},
		},
	}, nil
}

// toResourceDevice maps a discovered GCU into a ResourceSlice device.
// Attribute names follow HAMi nvidia-dra-driver GpuInfo (type/uuid/productName/
// brand/pcieBusID/driverVersion + standard pcieRoot), with GCU extras that the
// official Enflame GCU DRA also publishes (serialNumber, firmwareVersion) and
// the standard NUMA attribute when topology is known.
func toResourceDevice(d pkggcu.Device) resourceapi.Device {
	attrs := map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
		attrType:  {StringValue: ptr.To(DeviceType)},
		attrBrand: {StringValue: ptr.To(brandEnflame)},
	}
	if d.UUID != "" {
		attrs[attrUUID] = resourceapi.DeviceAttribute{StringValue: ptr.To(d.UUID)}
	}
	if d.Model != "" {
		attrs[attrProductName] = resourceapi.DeviceAttribute{StringValue: ptr.To(d.Model)}
	}
	if d.Serial != "" {
		attrs[attrSerialNumber] = resourceapi.DeviceAttribute{StringValue: ptr.To(d.Serial)}
	}
	if d.PCIBusID != "" {
		attrs[attrPCIBusID] = resourceapi.DeviceAttribute{StringValue: ptr.To(d.PCIBusID)}
	}
	if d.DriverVersion != "" {
		attrs[attrDriverVersion] = versionOrStringAttribute(d.DriverVersion)
	}
	if d.FirmwareVersion != "" {
		attrs[attrFirmwareVersion] = resourceapi.DeviceAttribute{StringValue: ptr.To(d.FirmwareVersion)}
	}
	if d.PCIeRoot != "" {
		attrs[deviceattribute.StandardDeviceAttributePCIeRoot] = resourceapi.DeviceAttribute{
			StringValue: ptr.To(d.PCIeRoot),
		}
	}
	if d.NUMANode >= 0 {
		attrs[deviceattribute.StandardDeviceAttributeNUMANode] = resourceapi.DeviceAttribute{
			IntValue: ptr.To(int64(d.NUMANode)),
		}
	}

	dev := resourceapi.Device{
		Name:       d.Name,
		Attributes: attrs,
	}
	if d.MemoryBytes > 0 {
		dev.Capacity = map[resourceapi.QualifiedName]resourceapi.DeviceCapacity{
			capacityMemory: {Value: *resource.NewQuantity(int64(d.MemoryBytes), resource.BinarySI)},
		}
	}
	return dev
}

// versionOrStringAttribute prefers VersionValue (HAMi nvidia style) when the
// string is valid semver; otherwise falls back to StringValue.
func versionOrStringAttribute(v string) resourceapi.DeviceAttribute {
	if parsed, err := semver.StrictNewVersion(v); err == nil {
		return resourceapi.DeviceAttribute{VersionValue: ptr.To(parsed.String())}
	}
	if parsed, err := semver.NewVersion(v); err == nil {
		return resourceapi.DeviceAttribute{VersionValue: ptr.To(parsed.String())}
	}
	return resourceapi.DeviceAttribute{StringValue: ptr.To(v)}
}

// ApplyConfig implements [profiles.ConfigHandler]. Opaque configuration is
// rejected until GcuConfig (stage 6). ExtraEdits inject allocation identity
// env vars used by demos and unit tests.
func (p Profile) ApplyConfig(config runtime.Object, results []*resourceapi.DeviceRequestAllocationResult) (profiles.PerDeviceCDIContainerEdits, error) {
	if config != nil {
		return nil, fmt.Errorf("configuration not allowed")
	}

	edits := make(profiles.PerDeviceCDIContainerEdits, len(results))
	for _, result := range results {
		deviceID := helpers.GetCDIDeviceID(result.Device, (*string)(result.ShareID))
		edits[deviceID] = &cdiapi.ContainerEdits{ContainerEdits: &cdispec.ContainerEdits{
			Env: []string{
				fmt.Sprintf("GCU_DEVICE_%s=%s", result.Device, result.Device),
			},
		}}
	}
	return edits, nil
}

// Validate implements [profiles.ConfigHandler].
func (p Profile) Validate(config runtime.Object) error {
	if config != nil {
		return fmt.Errorf("configuration not allowed")
	}
	return nil
}

// SchemeBuilder implements [profiles.ConfigHandler].
func (p Profile) SchemeBuilder() runtime.SchemeBuilder {
	return runtime.NewSchemeBuilder()
}
