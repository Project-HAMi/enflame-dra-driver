//go:build linux && cgo

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

package efml

/*
#cgo CFLAGS: -I${SRCDIR}/../../../third_party/efml/include
#cgo LDFLAGS: -L${SRCDIR}/../../../third_party/efml/lib -lefml
#include <stdlib.h>
#include <efml/efml.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	fwBufSize = 64
)

// Devices enumerates GCUs through libefml. The caller must not assume devices
// remain valid after a later library re-init; each call init/shutdowns EFML.
func Devices() ([]Info, error) {
	if ret := C.efmlInit_v2(); ret != C.EFML_SUCCESS_V2 {
		return nil, fmt.Errorf("efmlInit_v2: %w", statusError(ret))
	}
	defer C.efmlShutdown()

	var count C.uint
	if ret := C.efmlDeviceGetCount_v2(&count); ret != C.EFML_SUCCESS_V2 {
		return nil, fmt.Errorf("efmlDeviceGetCount_v2: %w", statusError(ret))
	}

	driverVersion, err := systemDriverVersion()
	if err != nil {
		return nil, err
	}

	out := make([]Info, 0, int(count))
	for i := C.uint(0); i < count; i++ {
		info, err := readDevice(i, driverVersion)
		if err != nil {
			return nil, fmt.Errorf("device index %d: %w", int(i), err)
		}
		out = append(out, info)
	}
	return out, nil
}

func readDevice(index C.uint, driverVersion string) (Info, error) {
	var dev C.efmlDevice_t
	if ret := C.efmlDeviceGetHandleByIndex_v2(index, &dev); ret != C.EFML_SUCCESS_V2 {
		return Info{}, fmt.Errorf("efmlDeviceGetHandleByIndex_v2: %w", statusError(ret))
	}

	var resolvedIndex C.uint
	if ret := C.efmlDeviceGetIndex(dev, &resolvedIndex); ret != C.EFML_SUCCESS_V2 {
		return Info{}, fmt.Errorf("efmlDeviceGetIndex: %w", statusError(ret))
	}

	var minor C.uint
	if ret := C.efmlDeviceGetMinorNumber(dev, &minor); ret != C.EFML_SUCCESS_V2 {
		return Info{}, fmt.Errorf("efmlDeviceGetMinorNumber: %w", statusError(ret))
	}

	name, err := stringCall(int(C.EFML_DEVICE_NAME_V2_BUFFER_SIZE), func(buf *C.char, n C.uint) C.efmlReturn_v1_t {
		return C.efmlDeviceGetName(dev, buf, n)
	})
	if err != nil {
		return Info{}, fmt.Errorf("efmlDeviceGetName: %w", err)
	}

	uuid, err := stringCall(int(C.EFML_DEVICE_UUID_BUFFER_SIZE), func(buf *C.char, n C.uint) C.efmlReturn_v1_t {
		return C.efmlDeviceGetUUID(dev, buf, n)
	})
	if err != nil {
		return Info{}, fmt.Errorf("efmlDeviceGetUUID: %w", err)
	}

	serial, _ := stringCall(int(C.EFML_DEVICE_SERIAL_BUFFER_SIZE), func(buf *C.char, n C.uint) C.efmlReturn_v1_t {
		return C.efmlDeviceGetSerial(dev, buf, n)
	})
	partNumber, _ := stringCall(int(C.EFML_DEVICE_PART_NUMBER_BUFFER_SIZE), func(buf *C.char, n C.uint) C.efmlReturn_v1_t {
		return C.efmlDeviceGetBoardPartNumber(dev, buf, n)
	})

	var pci C.efmlPciInfo_t
	pciBusID := ""
	if ret := C.efmlDeviceGetPciInfo_v3(dev, &pci); ret == C.EFML_SUCCESS_V2 {
		pciBusID = C.GoString(&pci.busId[0])
	}

	var mem C.efmlMemory_v2_t
	mem.version = C.efmlMemory_v2
	memoryTotal := uint64(0)
	memoryReserved := uint64(0)
	memoryUsed := uint64(0)
	memoryFree := uint64(0)
	if ret := C.efmlDeviceGetMemoryInfo_v2(dev, &mem); ret == C.EFML_SUCCESS_V2 {
		memoryTotal = uint64(mem.total)
		memoryReserved = uint64(mem.reserved)
		memoryUsed = uint64(mem.used)
		memoryFree = uint64(mem.free)
	} else {
		var memV1 C.efmlMemory_t
		if ret := C.efmlDeviceGetMemoryInfo(dev, &memV1); ret != C.EFML_SUCCESS_V2 {
			return Info{}, fmt.Errorf("efmlDeviceGetMemoryInfo: %w", statusError(ret))
		}
		memoryTotal = uint64(memV1.total)
		memoryUsed = uint64(memV1.used)
		memoryFree = uint64(memV1.free)
	}

	fw := C.malloc(fwBufSize)
	if fw == nil {
		return Info{}, fmt.Errorf("allocate firmware buffer")
	}
	defer C.free(fw)
	firmware := ""
	if ret := C.efmlDeviceGetSsmFirmwareVersion(dev, (*C.char)(fw)); ret == C.EFML_SUCCESS_V2 {
		firmware = C.GoString((*C.char)(fw))
	} else if ret := C.efmlDeviceGetVbiosVersion(dev, (*C.char)(fw), fwBufSize); ret == C.EFML_SUCCESS_V2 {
		firmware = C.GoString((*C.char)(fw))
	}

	return Info{
		Index:           int(resolvedIndex),
		Minor:           int(minor),
		UUID:            uuid,
		Model:           name,
		Serial:          serial,
		PartNumber:      partNumber,
		PCIBusID:        pciBusID,
		MemoryBytes:     memoryTotal,
		MemoryReserved:  memoryReserved,
		MemoryUsed:      memoryUsed,
		MemoryFree:      memoryFree,
		DriverVersion:   driverVersion,
		FirmwareVersion: firmware,
	}, nil
}

func systemDriverVersion() (string, error) {
	return stringCall(64, func(buf *C.char, n C.uint) C.efmlReturn_v1_t {
		return C.efmlSystemGetDriverVersion(buf, n)
	})
}

func stringCall(size int, fn func(*C.char, C.uint) C.efmlReturn_v1_t) (string, error) {
	buf := make([]byte, size)
	ret := fn((*C.char)(unsafe.Pointer(&buf[0])), C.uint(size))
	if ret != C.EFML_SUCCESS_V2 {
		return "", statusError(ret)
	}
	return C.GoString((*C.char)(unsafe.Pointer(&buf[0]))), nil
}

func statusError(ret C.efmlReturn_v1_t) error {
	return fmt.Errorf("%s (%d)", C.GoString(C.efmlErrorString(ret)), int(ret))
}
