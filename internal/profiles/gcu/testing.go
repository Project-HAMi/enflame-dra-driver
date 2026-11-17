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

	pkggcu "github.com/Project-HAMi/enflame-dra-driver/pkg/gcu"
)

// SetDiscoverForTest replaces the discovery function used by Profile.
// The returned restore function puts the previous implementation back.
func SetDiscoverForTest(f func(...pkggcu.Option) ([]pkggcu.Device, error)) (restore func()) {
	prev := discoverFunc
	discoverFunc = f
	return func() { discoverFunc = prev }
}

// StubDevices returns synthetic GCUs for unit tests that cannot link libefml.
func StubDevices(n int) []pkggcu.Device {
	if n <= 0 {
		return nil
	}
	devices := make([]pkggcu.Device, 0, n)
	for i := 0; i < n; i++ {
		devices = append(devices, pkggcu.Device{
			Name:     pkggcu.DRAName(i),
			Index:    i,
			Minor:    i,
			NUMANode: -1,
			Path:     fmt.Sprintf("/dev/gcu%d", i),
		})
	}
	return devices
}
