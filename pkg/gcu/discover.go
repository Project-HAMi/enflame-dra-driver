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
	"path/filepath"
	"sort"
	"strconv"

	"github.com/Project-HAMi/enflame-dra-driver/pkg/gcu/efml"
)

// discoverEFML is overwritten in unit tests.
var discoverEFML = efml.Devices

// Discover enumerates physical GCU devices through libefml (cgo).
// There is no sysfs fallback: EFML errors are returned to the caller.
//
// When host sysfs is present, NUMA node and PCIe root are filled in from the
// PCI topology for each discovered device.
func Discover(opts ...Option) ([]Device, error) {
	d := newDiscoverer(opts...)

	infos, err := discoverEFML()
	if err != nil {
		return nil, err
	}

	devices := make([]Device, 0, len(infos))
	for _, info := range infos {
		if info.Minor < 0 {
			return nil, fmt.Errorf("device EFML index %d: invalid minor %d", info.Index, info.Minor)
		}
		dev := Device{
			Name:            DRAName(info.Minor),
			Index:           info.Index,
			Minor:           info.Minor,
			UUID:            info.UUID,
			Model:           info.Model,
			Serial:          info.Serial,
			PartNumber:      info.PartNumber,
			PCIBusID:        info.PCIBusID,
			NUMANode:        -1,
			MemoryBytes:     info.MemoryBytes,
			MemoryReserved:  info.MemoryReserved,
			MemoryUsed:      info.MemoryUsed,
			MemoryFree:      info.MemoryFree,
			DriverVersion:   info.DriverVersion,
			FirmwareVersion: info.FirmwareVersion,
			Path:            filepath.Join(d.devRoot, "gcu"+strconv.Itoa(info.Minor)),
		}
		d.enrichTopology(&dev)
		devices = append(devices, dev)
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Minor < devices[j].Minor
	})
	return devices, nil
}

func (d *discoverer) enrichTopology(dev *Device) {
	// sysfs class names follow the char-device minor (/sys/class/enflame/gcuN).
	classEntry := hostPath(d.hostRoot, filepath.Join(d.classPath, fmt.Sprintf("gcu%d", dev.Minor)))
	pciDeviceDir := filepath.Join(classEntry, "device")
	resolved, err := filepath.EvalSymlinks(pciDeviceDir)
	if err != nil {
		return
	}
	dev.NUMANode = parseNUMANode(readTrimmedOptional(filepath.Join(resolved, "numa_node")))
	dev.PCIeRoot = findPCIeRoot(resolved)
}
