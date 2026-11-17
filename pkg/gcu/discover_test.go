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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Project-HAMi/enflame-dra-driver/pkg/gcu/efml"
)

func TestDiscoverUsesEFML(t *testing.T) {
	orig := discoverEFML
	t.Cleanup(func() { discoverEFML = orig })
	discoverEFML = func() ([]efml.Info, error) {
		return []efml.Info{{
			// EFML index 1 lives on char minor 4 (common on multi-socket S60).
			Index: 1, Minor: 4, UUID: "TR6W17100808", Model: "Enflame S60",
			Serial: "SN1", PartNumber: "PN1", PCIBusID: "0000:9d:00.0",
			MemoryBytes: 45063602176, MemoryReserved: 100, MemoryUsed: 0, MemoryFree: 44963602176,
			DriverVersion: "1.9.2401", FirmwareVersion: "33.6.16.5",
		}, {
			Index: 0, Minor: 0, UUID: "TR6W18220404", Model: "Enflame S60",
			Serial: "SN0", PartNumber: "PN0", PCIBusID: "0000:1b:00.0",
			MemoryBytes: 51506053120, DriverVersion: "1.9.2401", FirmwareVersion: "33.6.16.5",
		}}, nil
	}

	hostRoot := t.TempDir()
	writeTopology(t, hostRoot, 0, "pci0000:16", "0000:1b:00.0", "0")
	writeTopology(t, hostRoot, 4, "pci0000:97", "0000:9d:00.0", "1")

	devices, err := Discover(WithHostRoot(hostRoot))
	require.NoError(t, err)
	require.Len(t, devices, 2)

	assert.Equal(t, 0, devices[0].Index)
	assert.Equal(t, 0, devices[0].Minor)
	assert.Equal(t, "gcu-0", devices[0].Name)
	assert.Equal(t, "TR6W18220404", devices[0].UUID)
	assert.Equal(t, "Enflame S60", devices[0].Model)
	assert.Equal(t, "SN0", devices[0].Serial)
	assert.Equal(t, "PN0", devices[0].PartNumber)
	assert.Equal(t, "0000:1b:00.0", devices[0].PCIBusID)
	assert.Equal(t, "pci0000:16", devices[0].PCIeRoot)
	assert.Equal(t, 0, devices[0].NUMANode)
	assert.Equal(t, uint64(51506053120), devices[0].MemoryBytes)
	assert.Equal(t, "/dev/gcu0", devices[0].Path)

	assert.Equal(t, 1, devices[1].Index)
	assert.Equal(t, 4, devices[1].Minor)
	assert.Equal(t, "gcu-4", devices[1].Name)
	assert.Equal(t, "/dev/gcu4", devices[1].Path)
	assert.Equal(t, "pci0000:97", devices[1].PCIeRoot)
	assert.Equal(t, 1, devices[1].NUMANode)
}

func TestDiscoverPropagatesEFMLError(t *testing.T) {
	orig := discoverEFML
	t.Cleanup(func() { discoverEFML = orig })
	discoverEFML = func() ([]efml.Info, error) {
		return nil, fmt.Errorf("efml boom")
	}

	_, err := Discover()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "efml boom")
}

func writeTopology(t *testing.T, hostRoot string, minor int, pciRoot, pciSlot, numa string) {
	t.Helper()
	slotDir := filepath.Join(hostRoot, "sys/devices", pciRoot, pciSlot)
	require.NoError(t, os.MkdirAll(slotDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(slotDir, "numa_node"), []byte(numa+"\n"), 0o640))
	gcuDir := filepath.Join(slotDir, "enflame", fmt.Sprintf("gcu%d", minor))
	require.NoError(t, os.MkdirAll(gcuDir, 0o750))
	require.NoError(t, os.Symlink("../..", filepath.Join(gcuDir, "device")))
	classDir := filepath.Join(hostRoot, "sys/class/enflame")
	require.NoError(t, os.MkdirAll(classDir, 0o750))
	rel, err := filepath.Rel(classDir, gcuDir)
	require.NoError(t, err)
	require.NoError(t, os.Symlink(rel, filepath.Join(classDir, fmt.Sprintf("gcu%d", minor))))
}
