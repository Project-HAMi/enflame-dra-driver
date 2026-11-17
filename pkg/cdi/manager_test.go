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
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

func setupHost(t *testing.T) (hostRoot, cdiRoot string) {
	t.Helper()
	hostRoot = t.TempDir()
	cdiRoot = t.TempDir()
	writeFile(t, filepath.Join(hostRoot, "dev", "gcuctl"), "ctl")
	writeFile(t, filepath.Join(hostRoot, "dev", "gcu0"), "gcu")
	writeFile(t, filepath.Join(hostRoot, "dev", "gcu0vid0"), "vid")
	writeFile(t, filepath.Join(hostRoot, "usr", "bin", "efsmi"), "efsmi")
	writeFile(t, filepath.Join(hostRoot, "usr", "lib", "libefml.so"), "lib")
	return hostRoot, cdiRoot
}

func readSpec(t *testing.T, dir, name string) cdispec.Spec {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name+".yaml"))
	require.NoError(t, err)
	var spec cdispec.Spec
	require.NoError(t, yaml.Unmarshal(data, &spec))
	return spec
}

func TestCreateBaseAndClaimSpecs(t *testing.T) {
	hostRoot, cdiRoot := setupHost(t)
	m, err := NewManager(
		WithHostRoot(hostRoot),
		WithCDIRoot(cdiRoot),
		WithDriverName("gcu.enflame.com"),
		WithClass("gcu"),
		WithNodeName("node-a"),
	)
	require.NoError(t, err)
	require.Equal(t, "k8s.gcu.enflame.com", m.Vendor())
	require.Equal(t, "k8s.gcu.enflame.com/gcu", m.Kind())

	require.NoError(t, m.CreateBaseSpec())
	baseName, err := cdiapi.GenerateNameForTransientSpec(&cdispec.Spec{Kind: m.Kind()}, BaseDeviceName)
	require.NoError(t, err)
	base := readSpec(t, cdiRoot, baseName)
	require.Len(t, base.Devices, 1)
	assert.Equal(t, BaseDeviceName, base.Devices[0].Name)
	assert.Contains(t, base.Devices[0].ContainerEdits.Env, "KUBERNETES_NODE_NAME=node-a")
	assert.Contains(t, base.Devices[0].ContainerEdits.Env, "DRA_RESOURCE_DRIVER_NAME=gcu.enflame.com")
	require.Len(t, base.Devices[0].ContainerEdits.DeviceNodes, 1)
	require.NotNil(t, base.Devices[0].ContainerEdits.DeviceNodes[0])
	assert.Equal(t, "/dev/gcuctl", base.Devices[0].ContainerEdits.DeviceNodes[0].Path)
	require.NotEmpty(t, base.Devices[0].ContainerEdits.Mounts)
	for _, mount := range base.Devices[0].ContainerEdits.Mounts {
		require.NotNil(t, mount)
		assert.NotEqual(t, "/lib/firmware/enflame_ucode_bin", mount.ContainerPath)
	}

	claimUID := "11111111-2222-3333-4444-555555555555"
	require.NoError(t, m.CreateClaimSpec(claimUID, []Device{{
		Name:       "gcu-0",
		Index:      0,
		VideoNodes: true,
		ExtraEdits: &cdiapi.ContainerEdits{ContainerEdits: &cdispec.ContainerEdits{
			Env: []string{"EXTRA=1"},
		}},
	}}))

	claimName := cdiapi.GenerateTransientSpecName(m.Vendor(), m.Class(), claimUID)
	claim := readSpec(t, cdiRoot, claimName)
	require.Len(t, claim.Devices, 1)
	assert.Equal(t, claimUID+"-gcu-0", claim.Devices[0].Name)
	assert.Contains(t, claim.Devices[0].ContainerEdits.Env, "GCU_DEVICE_GCU_0_RESOURCE_CLAIM="+claimUID)
	assert.Contains(t, claim.Devices[0].ContainerEdits.Env, "EXTRA=1")
	paths := make([]string, 0, len(claim.Devices[0].ContainerEdits.DeviceNodes))
	for _, n := range claim.Devices[0].ContainerEdits.DeviceNodes {
		require.NotNil(t, n)
		paths = append(paths, n.Path)
	}
	assert.Equal(t, []string{"/dev/gcu0", "/dev/gcu0vid0"}, paths)

	ids := m.ClaimDeviceIDs(claimUID, []string{"gcu-0"})
	assert.Equal(t, []string{
		"k8s.gcu.enflame.com/gcu=base",
		"k8s.gcu.enflame.com/gcu=" + claimUID + "-gcu-0",
	}, ids)

	require.NoError(t, m.DeleteClaimSpec(claimUID))
	_, err = os.Stat(filepath.Join(cdiRoot, claimName+".yaml"))
	assert.True(t, os.IsNotExist(err))
}

func TestCreateClaimSpecMockDevice(t *testing.T) {
	hostRoot, cdiRoot := setupHost(t)
	m, err := NewManager(WithHostRoot(hostRoot), WithCDIRoot(cdiRoot), WithDriverName("gpu.example.com"), WithClass("gpu"))
	require.NoError(t, err)
	require.NoError(t, m.CreateBaseSpec())

	claimUID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	require.NoError(t, m.CreateClaimSpec(claimUID, []Device{{
		Name:  "gpu-0",
		Index: -1,
		ExtraEdits: &cdiapi.ContainerEdits{ContainerEdits: &cdispec.ContainerEdits{
			Env: []string{"GPU_DEVICE_0=gpu-0"},
		}},
	}}))

	claim := readSpec(t, cdiRoot, cdiapi.GenerateTransientSpecName(m.Vendor(), m.Class(), claimUID))
	require.Len(t, claim.Devices, 1)
	assert.Empty(t, claim.Devices[0].ContainerEdits.DeviceNodes)
	assert.Contains(t, claim.Devices[0].ContainerEdits.Env, "GPU_DEVICE_0=gpu-0")
}

func TestCleanupOrphanedClaimSpecs(t *testing.T) {
	hostRoot, cdiRoot := setupHost(t)
	m, err := NewManager(WithHostRoot(hostRoot), WithCDIRoot(cdiRoot), WithDriverName("gcu.enflame.com"))
	require.NoError(t, err)
	require.NoError(t, m.CreateBaseSpec())

	activeUID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	orphanUID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	require.NoError(t, m.CreateClaimSpec(activeUID, []Device{{Name: "gcu-0", Index: 0}}))
	require.NoError(t, m.CreateClaimSpec(orphanUID, []Device{{Name: "gcu-0", Index: 0}}))

	require.NoError(t, m.CleanupOrphanedClaimSpecs(sets.New(activeUID)))

	_, err = os.Stat(filepath.Join(cdiRoot, cdiapi.GenerateTransientSpecName(m.Vendor(), m.Class(), activeUID)+".yaml"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(cdiRoot, cdiapi.GenerateTransientSpecName(m.Vendor(), m.Class(), orphanUID)+".yaml"))
	assert.True(t, os.IsNotExist(err))

	baseName, err := cdiapi.GenerateNameForTransientSpec(&cdispec.Spec{Kind: m.Kind()}, BaseDeviceName)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(cdiRoot, baseName+".yaml"))
	require.NoError(t, err)
}
