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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/dynamic-resource-allocation/deviceattribute"
	"k8s.io/utils/ptr"

	pkggcu "github.com/Project-HAMi/enflame-dra-driver/pkg/gcu"
)

func TestEnumerateDevicesFromDiscover(t *testing.T) {
	orig := discoverFunc
	t.Cleanup(func() { discoverFunc = orig })
	discoverFunc = func(opts ...pkggcu.Option) ([]pkggcu.Device, error) {
		return []pkggcu.Device{{
			Name: "gcu-0", Index: 0, Minor: 0, UUID: "TR6W18220404", Model: "Enflame S60",
			Serial: "SN0", PartNumber: "PN0",
			PCIBusID: "0000:1b:00.0", PCIeRoot: "pci0000:16", NUMANode: 0,
			MemoryBytes: 51506053120, DriverVersion: "1.9.2401",
			FirmwareVersion: "33.6.16.5",
		}}, nil
	}

	p := NewProfile("node-a", "/", 0)
	resources, err := p.EnumerateDevices()
	require.NoError(t, err)
	devices := resources.Pools["node-a"].Slices[0].Devices
	require.Len(t, devices, 1)
	d := devices[0]

	assert.Equal(t, DeviceType, *d.Attributes[attrType].StringValue)
	assert.Equal(t, brandEnflame, *d.Attributes[attrBrand].StringValue)
	assert.Equal(t, "TR6W18220404", *d.Attributes[attrUUID].StringValue)
	assert.Equal(t, "Enflame S60", *d.Attributes[attrProductName].StringValue)
	assert.Equal(t, "SN0", *d.Attributes[attrSerialNumber].StringValue)
	assert.Equal(t, "0000:1b:00.0", *d.Attributes[attrPCIBusID].StringValue)
	assert.Equal(t, "1.9.2401", *d.Attributes[attrDriverVersion].VersionValue)
	assert.Nil(t, d.Attributes[attrDriverVersion].StringValue)
	assert.Equal(t, "33.6.16.5", *d.Attributes[attrFirmwareVersion].StringValue)
	assert.Equal(t, "pci0000:16", *d.Attributes[deviceattribute.StandardDeviceAttributePCIeRoot].StringValue)
	assert.Equal(t, int64(0), *d.Attributes[deviceattribute.StandardDeviceAttributeNUMANode].IntValue)
	assert.NotContains(t, d.Attributes, resourceapi.QualifiedName("model"))
	assert.NotContains(t, d.Attributes, resourceapi.QualifiedName("index"))
	assert.NotContains(t, d.Attributes, resourceapi.QualifiedName("partNumber"))
	assert.NotContains(t, d.Attributes, resourceapi.QualifiedName("numaNode"))

	require.NotNil(t, d.AllowMultipleAllocations)
	assert.True(t, *d.AllowMultipleAllocations)

	require.Contains(t, d.Capacity, capacityCores)
	coresCap := d.Capacity[capacityCores]
	coresVal := coresCap.Value
	assert.Equal(t, int64(100), coresVal.Value())
	require.NotNil(t, coresCap.RequestPolicy)
	require.NotNil(t, coresCap.RequestPolicy.Default)
	assert.Equal(t, int64(100), coresCap.RequestPolicy.Default.Value())
	require.Len(t, coresCap.RequestPolicy.ValidValues, 1)
	vv := coresCap.RequestPolicy.ValidValues[0]
	assert.Equal(t, int64(100), vv.Value())

	require.Contains(t, d.Capacity, capacityMemory)
	memCap := d.Capacity[capacityMemory]
	memVal := memCap.Value
	assert.Equal(t, int64(51506053120), memVal.Value())
	require.NotNil(t, memCap.RequestPolicy)
	require.NotNil(t, memCap.RequestPolicy.Default)
	assert.Equal(t, int64(51506053120), memCap.RequestPolicy.Default.Value())
	require.Len(t, memCap.RequestPolicy.ValidValues, 1)
	mv := memCap.RequestPolicy.ValidValues[0]
	assert.Equal(t, int64(51506053120), mv.Value())
}

func TestEnumerateDevicesOmitsUnknownNUMA(t *testing.T) {
	orig := discoverFunc
	t.Cleanup(func() { discoverFunc = orig })
	discoverFunc = func(opts ...pkggcu.Option) ([]pkggcu.Device, error) {
		return []pkggcu.Device{{
			Name: "gcu-0", Index: 0, Minor: 0, UUID: "u", Model: "Enflame S60",
			NUMANode: -1, DriverVersion: "not-a-version",
		}}, nil
	}

	resources, err := NewProfile("node-a", "/", 0).EnumerateDevices()
	require.NoError(t, err)
	d := resources.Pools["node-a"].Slices[0].Devices[0]
	_, ok := d.Attributes[deviceattribute.StandardDeviceAttributeNUMANode]
	assert.False(t, ok)
	assert.Equal(t, "not-a-version", *d.Attributes[attrDriverVersion].StringValue)
	assert.Nil(t, d.Attributes[attrDriverVersion].VersionValue)
}

func TestEnumerateDevicesPropagatesError(t *testing.T) {
	orig := discoverFunc
	t.Cleanup(func() { discoverFunc = orig })
	discoverFunc = func(opts ...pkggcu.Option) ([]pkggcu.Device, error) {
		return nil, assert.AnError
	}
	_, err := NewProfile("node-a", "/", 0).EnumerateDevices()
	require.Error(t, err)
}

func TestApplyConfigRejectsOpaqueConfig(t *testing.T) {
	p := NewProfile("node-a", "/", 1)
	_, err := p.ApplyConfig(&resourceapi.ResourceClaim{}, nil)
	require.Error(t, err)
}

func TestApplyConfigShareAwareKeys(t *testing.T) {
	p := NewProfile("node-a", "/", 1)
	shareID := types.UID("share-0")
	edits, err := p.ApplyConfig(nil, []*resourceapi.DeviceRequestAllocationResult{{
		Request: "gcu",
		Device:  "gcu-0",
		ShareID: ptr.To(shareID),
	}})
	require.NoError(t, err)
	require.Contains(t, edits, "gcu-0-share-0")
	require.NotNil(t, edits["gcu-0-share-0"].ContainerEdits)
	assert.Equal(t, []string{"GCU_DEVICE_gcu-0=gcu-0"}, edits["gcu-0-share-0"].Env)
}
