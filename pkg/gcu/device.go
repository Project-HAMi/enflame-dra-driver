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
	"strconv"
	"strings"
)

// Device is one physical GCU discovered through EFML.
type Device struct {
	// Name is the DRA device name, keyed by char-device minor: "gcu-{Minor}".
	Name string
	// Index is the EFML / efsmi logical device index. It may differ from Minor.
	Index int
	// Minor is the character-device minor for Path and sysfs class enflame/gcuN.
	Minor int
	// UUID is the board UUID from EFML.
	UUID string
	// Model is the product name, e.g. "Enflame S60".
	Model string
	// Serial is the board serial number.
	Serial string
	// PartNumber is the board part number.
	PartNumber string
	// PCIBusID is the PCI slot name, e.g. "0000:1b:00.0".
	PCIBusID string
	// PCIeRoot is the PCI root complex, e.g. "pci0000:16" (from sysfs topology).
	PCIeRoot string
	// NUMANode is the NUMA node for the PCI device (-1 if unknown).
	NUMANode int
	// MemoryBytes is total device HBM capacity in bytes.
	MemoryBytes uint64
	// MemoryReserved is bytes reserved by driver/firmware.
	MemoryReserved uint64
	// MemoryUsed is currently allocated device memory.
	MemoryUsed uint64
	// MemoryFree is unallocated device memory.
	MemoryFree uint64
	// DriverVersion is the enflame driver version.
	DriverVersion string
	// FirmwareVersion is the SSM (or VBIOS) firmware version.
	FirmwareVersion string
	// Path is the host device node path, e.g. "/dev/gcu0" (/dev/gcu{Minor}).
	Path string
}

// DRAName returns the DRA device name for a char-device minor.
func DRAName(minor int) string {
	return fmt.Sprintf("gcu-%d", minor)
}

// ParseDRAName extracts the char-device minor from a DRA name such as "gcu-0".
func ParseDRAName(name string) (int, error) {
	const prefix = "gcu-"
	if !strings.HasPrefix(name, prefix) {
		return 0, fmt.Errorf("invalid GCU DRA name %q", name)
	}
	rest := name[len(prefix):]
	if rest == "" {
		return 0, fmt.Errorf("invalid GCU DRA name %q", name)
	}
	for i := 0; i < len(rest); i++ {
		if rest[i] < '0' || rest[i] > '9' {
			return 0, fmt.Errorf("invalid GCU DRA name %q", name)
		}
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0, fmt.Errorf("invalid GCU DRA name %q: %w", name, err)
	}
	return n, nil
}
