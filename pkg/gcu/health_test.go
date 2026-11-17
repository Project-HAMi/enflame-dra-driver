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
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckHealthHealthy(t *testing.T) {
	root := t.TempDir()
	writeHealthFiles(t, root, 0, "healthy", "working")

	h := CheckHealth(0, WithHostRoot(root))
	assert.Equal(t, HealthLevelHealthy, h.Level)
	assert.Equal(t, "healthy", h.Health)
	assert.Equal(t, "working", h.State)
}

func TestCheckHealthUnhealthy(t *testing.T) {
	root := t.TempDir()
	writeHealthFiles(t, root, 1, "unhealthy", "working")
	h := CheckHealth(1, WithHostRoot(root))
	assert.Equal(t, HealthLevelUnhealthy, h.Level)
	assert.Contains(t, h.Message, "unhealthy")

	writeHealthFiles(t, root, 2, "healthy", "error")
	h = CheckHealth(2, WithHostRoot(root))
	assert.Equal(t, HealthLevelUnhealthy, h.Level)
	assert.Contains(t, h.Message, "error")
}

func TestCheckHealthUnknown(t *testing.T) {
	root := t.TempDir()
	h := CheckHealth(0, WithHostRoot(root))
	assert.Equal(t, HealthLevelUnknown, h.Level)

	deviceDir := filepath.Join(root, "sys/class/enflame/gcu0/device")
	require.NoError(t, os.MkdirAll(deviceDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(deviceDir, "device_health"), []byte("healthy\n"), 0o640))
	h = CheckHealth(0, WithHostRoot(root))
	assert.Equal(t, HealthLevelUnknown, h.Level)
	assert.Equal(t, "healthy", h.Health)
}

func writeHealthFiles(t *testing.T, hostRoot string, index int, health, state string) {
	t.Helper()
	deviceDir := filepath.Join(hostRoot, "sys/class/enflame", "gcu"+strconv.Itoa(index), "device")
	require.NoError(t, os.MkdirAll(deviceDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(deviceDir, "device_health"), []byte(health+"\n"), 0o640))
	require.NoError(t, os.WriteFile(filepath.Join(deviceDir, "device_state"), []byte(state+"\n"), 0o640))
}
