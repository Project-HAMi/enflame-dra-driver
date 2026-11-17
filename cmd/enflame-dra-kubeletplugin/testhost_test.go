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
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Project-HAMi/enflame-dra-driver/internal/profiles/gcu"
	gcucdi "github.com/Project-HAMi/enflame-dra-driver/pkg/cdi"
	pkggcu "github.com/Project-HAMi/enflame-dra-driver/pkg/gcu"
)

// cdiBaseDeviceName is the common CDI device referenced by every claim.
const cdiBaseDeviceName = gcucdi.BaseDeviceName

func setupFakeHostRoot(t *testing.T, numDevices int) string {
	t.Helper()
	hostRoot := t.TempDir()
	writeFakeHostDevices(t, hostRoot, numDevices)
	return hostRoot
}

func setupFakeHostRootAt(t *testing.T, dir string, numDevices int) string {
	t.Helper()
	hostRoot := filepath.Join(dir, "host")
	writeFakeHostDevices(t, hostRoot, numDevices)
	return hostRoot
}

func writeFakeHostDevices(t *testing.T, hostRoot string, numDevices int) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(hostRoot, "dev"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(hostRoot, "dev", "gcuctl"), []byte("ctl"), 0o640))
	for i := 0; i < numDevices; i++ {
		path := filepath.Join(hostRoot, "dev", fmt.Sprintf("gcu%d", i))
		require.NoError(t, os.WriteFile(path, []byte("dev"), 0o640))
		deviceDir := filepath.Join(hostRoot, "sys/class/enflame", fmt.Sprintf("gcu%d", i), "device")
		require.NoError(t, os.MkdirAll(deviceDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(deviceDir, "device_health"), []byte("healthy\n"), 0o640))
		require.NoError(t, os.WriteFile(filepath.Join(deviceDir, "device_state"), []byte("working\n"), 0o640))
	}
}

// useStubGCUDiscover installs a synthetic EFML-free discoverer for unit tests.
func useStubGCUDiscover(t *testing.T, numDevices int) {
	t.Helper()
	devices := gcu.StubDevices(numDevices)
	restore := gcu.SetDiscoverForTest(func(...pkggcu.Option) ([]pkggcu.Device, error) {
		return devices, nil
	})
	t.Cleanup(restore)
}
