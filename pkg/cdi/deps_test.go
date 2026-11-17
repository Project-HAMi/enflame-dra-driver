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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o640))
}

func TestDiscoverBaseEditsFullDeps(t *testing.T) {
	host := t.TempDir()
	writeFile(t, filepath.Join(host, "dev", "gcuctl"), "ctl")
	writeFile(t, filepath.Join(host, "lib", "firmware", "enflame_ucode_bin", "x.bin"), "fw")
	writeFile(t, filepath.Join(host, "usr", "bin", "efsmi"), "efsmi")
	writeFile(t, filepath.Join(host, "usr", "lib", "libefml.so"), "lib")
	writeFile(t, filepath.Join(host, "usr", "lib", "libefml.so.1"), "lib")
	writeFile(t, filepath.Join(host, "usr", "lib", "libefml.so.1.2.3"), "lib")

	m, err := NewManager(WithHostRoot(host), WithCDIRoot(t.TempDir()), WithDevRoot("/dev"))
	require.NoError(t, err)

	edits, err := m.discoverBaseEdits()
	require.NoError(t, err)
	require.Len(t, edits.DeviceNodes, 1)
	require.NotNil(t, edits.DeviceNodes[0])
	assert.Equal(t, "/dev/gcuctl", edits.DeviceNodes[0].Path)

	mountPaths := make([]string, 0, len(edits.Mounts))
	for _, mount := range edits.Mounts {
		require.NotNil(t, mount)
		assert.Equal(t, []string{"ro", "nosuid", "nodev", "bind"}, mount.Options)
		assert.Equal(t, "bind", mount.Type)
		assert.Equal(t, mount.HostPath, mount.ContainerPath)
		mountPaths = append(mountPaths, mount.HostPath)
	}
	assert.Contains(t, mountPaths, "/lib/firmware/enflame_ucode_bin")
	assert.Contains(t, mountPaths, "/usr/bin/efsmi")
	assert.Contains(t, mountPaths, "/usr/lib/libefml.so")
	assert.Contains(t, mountPaths, "/usr/lib/libefml.so.1")
	assert.Contains(t, mountPaths, "/usr/lib/libefml.so.1.2.3")
}

func TestDiscoverBaseEditsOptionalMissing(t *testing.T) {
	host := t.TempDir()
	writeFile(t, filepath.Join(host, "dev", "gcuctl"), "ctl")

	m, err := NewManager(WithHostRoot(host), WithCDIRoot(t.TempDir()))
	require.NoError(t, err)

	edits, err := m.discoverBaseEdits()
	require.NoError(t, err)
	require.Len(t, edits.DeviceNodes, 1)
	assert.Empty(t, edits.Mounts)
}

func TestDiscoverBaseEditsMissingGcuctl(t *testing.T) {
	host := t.TempDir()
	m, err := NewManager(WithHostRoot(host), WithCDIRoot(t.TempDir()))
	require.NoError(t, err)

	_, err = m.discoverBaseEdits()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "/dev/gcuctl")
}

func TestDiscoverDeviceEdits(t *testing.T) {
	host := t.TempDir()
	writeFile(t, filepath.Join(host, "dev", "gcu0"), "gcu")
	writeFile(t, filepath.Join(host, "dev", "gcu0vid0"), "vid")
	writeFile(t, filepath.Join(host, "dev", "gcu0vid1"), "vid")

	m, err := NewManager(WithHostRoot(host), WithCDIRoot(t.TempDir()))
	require.NoError(t, err)

	edits, err := m.discoverDeviceEdits(0, false)
	require.NoError(t, err)
	require.Len(t, edits.DeviceNodes, 1)
	require.NotNil(t, edits.DeviceNodes[0])
	assert.Equal(t, "/dev/gcu0", edits.DeviceNodes[0].Path)

	edits, err = m.discoverDeviceEdits(0, true)
	require.NoError(t, err)
	require.Len(t, edits.DeviceNodes, 3)
	require.NotNil(t, edits.DeviceNodes[0])
	require.NotNil(t, edits.DeviceNodes[1])
	require.NotNil(t, edits.DeviceNodes[2])
	assert.Equal(t, "/dev/gcu0", edits.DeviceNodes[0].Path)
	assert.Equal(t, "/dev/gcu0vid0", edits.DeviceNodes[1].Path)
	assert.Equal(t, "/dev/gcu0vid1", edits.DeviceNodes[2].Path)
}

func TestDiscoverDeviceEditsMissing(t *testing.T) {
	m, err := NewManager(WithHostRoot(t.TempDir()), WithCDIRoot(t.TempDir()))
	require.NoError(t, err)

	_, err = m.discoverDeviceEdits(3, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "/dev/gcu3")
}

func TestDiscoverDeviceEditsMockIndex(t *testing.T) {
	m, err := NewManager(WithHostRoot(t.TempDir()), WithCDIRoot(t.TempDir()))
	require.NoError(t, err)

	edits, err := m.discoverDeviceEdits(-1, true)
	require.NoError(t, err)
	assert.Empty(t, edits.DeviceNodes)
}
