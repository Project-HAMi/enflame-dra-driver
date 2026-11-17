/*
 * Copyright 2025 The Kubernetes Authors.
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

package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	drapbv1 "k8s.io/kubelet/pkg/apis/dra/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
	cdiparser "tags.cncf.io/container-device-interface/pkg/parser"
	cdispec "tags.cncf.io/container-device-interface/specs-go"

	"github.com/Project-HAMi/enflame-dra-driver/internal/profiles/gcu"
	gcucdi "github.com/Project-HAMi/enflame-dra-driver/pkg/cdi"
)

var (
	testShareId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
)

func TestPreparedDevicesGetDevices(t *testing.T) {
	tests := map[string]struct {
		preparedDevices PreparedDevices
		expected        []*drapbv1.Device
	}{
		"nil PreparedDevices": {
			preparedDevices: nil,
			expected:        nil,
		},
		"several PreparedDevices": {
			preparedDevices: PreparedDevices{
				{Device: drapbv1.Device{DeviceName: "dev1"}},
				{Device: drapbv1.Device{DeviceName: "dev2"}},
				{Device: drapbv1.Device{DeviceName: "dev3"}},
			},
			expected: []*drapbv1.Device{
				{DeviceName: "dev1"},
				{DeviceName: "dev2"},
				{DeviceName: "dev3"},
			},
		},
		"preparedDevice with shareId": {
			preparedDevices: PreparedDevices{
				{Device: drapbv1.Device{DeviceName: "dev1", ShareId: &testShareId}},
				{Device: drapbv1.Device{DeviceName: "dev2", ShareId: &testShareId}},
			},
			expected: []*drapbv1.Device{
				{DeviceName: "dev1", ShareId: &testShareId},
				{DeviceName: "dev2", ShareId: &testShareId},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			devices := test.preparedDevices.GetDevices()
			assert.Equal(t, test.expected, devices)
		})
	}
}

func TestComputeDeviceConfigShareID(t *testing.T) {
	const (
		nodeName   = "test-node"
		driverName = gcu.DefaultDriverName
	)

	flags := &Flags{
		cdiRoot:    t.TempDir(),
		hostRoot:   setupFakeHostRoot(t, 1),
		driverName: driverName,
		profile:    gcu.ProfileName,
		nodeName:   nodeName,
		numDevices: 1,
	}
	useStubGCUDiscover(t, 1)
	state, err := NewDeviceState(&Config{
		flags:   flags,
		profile: gcu.NewProfile(nodeName, flags.hostRoot, flags.numDevices),
	})
	require.NoError(t, err)

	result := func(request string, shareID *types.UID) resourceapi.DeviceRequestAllocationResult {
		return resourceapi.DeviceRequestAllocationResult{
			Request: request,
			Driver:  driverName,
			Pool:    nodeName,
			Device:  "gcu-0",
			ShareID: shareID,
		}
	}

	tests := map[string]struct {
		results             []resourceapi.DeviceRequestAllocationResult
		expectedShareIDs    []*types.UID
		expectedCDISuffixes []string
	}{
		"no ShareID": {
			results:             []resourceapi.DeviceRequestAllocationResult{result("gcu", nil)},
			expectedShareIDs:    []*types.UID{nil},
			expectedCDISuffixes: []string{"claim-uid-gcu-0"},
		},
		"distinct ShareIDs sharing one device": {
			results: []resourceapi.DeviceRequestAllocationResult{
				result("gcu-0", ptr.To(types.UID("share-0"))),
				result("gcu-1", ptr.To(types.UID("share-1"))),
			},
			expectedShareIDs:    []*types.UID{ptr.To(types.UID("share-0")), ptr.To(types.UID("share-1"))},
			expectedCDISuffixes: []string{"claim-uid-gcu-0-share-0", "claim-uid-gcu-0-share-1"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			claim := &resourceapi.ResourceClaim{
				ObjectMeta: metav1.ObjectMeta{UID: "claim-uid"},
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						Devices: resourceapi.DeviceAllocationResult{Results: test.results},
					},
				},
			}

			prepared, err := state.computeDeviceConfig(claim)
			require.NoError(t, err)
			require.Len(t, prepared, len(test.expectedShareIDs))

			var shareIDs []*types.UID
			var cdiDeviceIDs []string
			for _, device := range prepared {
				shareIDs = append(shareIDs, device.ShareID)
				cdiDeviceIDs = append(cdiDeviceIDs, device.CdiDeviceIds...)
			}
			assert.ElementsMatch(t, test.expectedShareIDs, shareIDs)

			for _, suffix := range test.expectedCDISuffixes {
				matched := slices.ContainsFunc(cdiDeviceIDs, func(id string) bool {
					return strings.HasSuffix(id, suffix)
				})
				assert.True(t, matched, "expected a CDI device ID ending with %q in %v", suffix, cdiDeviceIDs)
			}
		})
	}
}

func TestComputeDeviceConfigSharedDeviceContainerEdits(t *testing.T) {
	const (
		nodeName   = "test-node"
		driverName = gcu.DefaultDriverName
	)

	flags := &Flags{
		cdiRoot:    t.TempDir(),
		hostRoot:   setupFakeHostRoot(t, 1),
		driverName: driverName,
		profile:    gcu.ProfileName,
		nodeName:   nodeName,
		numDevices: 1,
	}
	useStubGCUDiscover(t, 1)
	state, err := NewDeviceState(&Config{
		flags:   flags,
		profile: gcu.NewProfile(nodeName, flags.hostRoot, flags.numDevices),
	})
	require.NoError(t, err)

	result := func(request string, shareID types.UID) resourceapi.DeviceRequestAllocationResult {
		return resourceapi.DeviceRequestAllocationResult{
			Request: request,
			Driver:  driverName,
			Pool:    nodeName,
			Device:  "gcu-0",
			ShareID: ptr.To(shareID),
		}
	}

	claim := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{UID: "claim-uid"},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{Results: []resourceapi.DeviceRequestAllocationResult{
					result("gcu0", "share-0"),
					result("gcu1", "share-1"),
				}},
			},
		},
	}

	prepared, err := state.computeDeviceConfig(claim)
	require.NoError(t, err)
	require.Len(t, prepared, 2)

	for _, device := range prepared {
		require.NotNil(t, device.ShareID)
		require.NotNil(t, device.ContainerEdits)
		require.NotNil(t, device.ContainerEdits.ContainerEdits)
		assert.Contains(t, device.ContainerEdits.Env, "GCU_DEVICE_gcu-0=gcu-0")
	}
}

func TestPreparePerDeviceAdminAccess(t *testing.T) {
	const (
		nodeName   = "test-node"
		driverName = gcu.DefaultDriverName
	)
	for _, tc := range []struct {
		name    string
		results []resourceapi.DeviceRequestAllocationResult
		want    map[string]bool
	}{
		{
			name: "mixed admin and ordinary allocations",
			results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "admin", Driver: driverName, Pool: nodeName, Device: "gcu-0", AdminAccess: ptr.To(true)},
				{Request: "ordinary", Driver: driverName, Pool: nodeName, Device: "gcu-1"},
				{Request: "explicit-false", Driver: driverName, Pool: nodeName, Device: "gcu-2", AdminAccess: ptr.To(false)},
			},
			want: map[string]bool{"gcu-0": true, "gcu-1": false, "gcu-2": false},
		},
		{
			name: "admin allocation for another driver",
			results: []resourceapi.DeviceRequestAllocationResult{
				{Request: "ordinary", Driver: driverName, Pool: nodeName, Device: "gcu-0"},
				{Request: "admin", Driver: "other.example.com", Pool: nodeName, Device: "other-0", AdminAccess: ptr.To(true)},
			},
			want: map[string]bool{"gcu-0": false},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			flags := &Flags{
				cdiRoot: root, hostRoot: setupFakeHostRoot(t, 3), driverName: driverName, profile: gcu.ProfileName, nodeName: nodeName,
				numDevices: 3, kubeletPluginsDirectoryPath: root,
			}
			require.NoError(t, os.MkdirAll(filepath.Join(root, driverName), 0750))
			useStubGCUDiscover(t, 3)
			state, err := NewDeviceState(&Config{
				flags:   flags,
				profile: gcu.NewProfile(nodeName, flags.hostRoot, flags.numDevices),
			})
			require.NoError(t, err)
			claim := testGCUClaim(driverName, nodeName)
			claim.Status.Allocation.Devices.Results = tc.results

			for range 2 {
				prepared, err := state.Prepare(t.Context(), claim)
				require.NoError(t, err)
				require.Len(t, prepared, len(tc.want))
				for _, device := range prepared {
					want, ok := tc.want[device.DeviceName]
					require.True(t, ok)
					assert.Equal(t, want, device.AdminAccess, "allocation %s", device.DeviceName)
				}
				specBytes, err := os.ReadFile(claimSpecPath(state, claim.UID))
				require.NoError(t, err)
				var spec cdispec.Spec
				require.NoError(t, yaml.Unmarshal(specBytes, &spec))
				require.Len(t, spec.Devices, len(tc.want))
				for _, device := range spec.Devices {
					deviceName := strings.TrimPrefix(device.Name, string(claim.UID)+"-")
					want, ok := tc.want[deviceName]
					require.True(t, ok)
					if want {
						assert.Contains(t, device.ContainerEdits.Env, "DRA_ADMIN_ACCESS=true")
					} else {
						assert.Contains(t, device.ContainerEdits.Env, "DRA_ADMIN_ACCESS=false")
					}
				}
			}
		})
	}
}

func TestUnprepareReturnsErrorOnUnreadableCheckpoint(t *testing.T) {
	const (
		nodeName   = "test-node"
		driverName = gcu.DefaultDriverName
		claimUID   = "some-claim-uid"
	)

	tmpDir := t.TempDir()
	flags := &Flags{
		cdiRoot:                     tmpDir,
		hostRoot:                    setupFakeHostRoot(t, 1),
		driverName:                  driverName,
		profile:                     gcu.ProfileName,
		nodeName:                    nodeName,
		numDevices:                  1,
		kubeletPluginsDirectoryPath: tmpDir,
	}

	useStubGCUDiscover(t, 1)
	state, err := NewDeviceState(&Config{
		flags:   flags,
		profile: gcu.NewProfile(nodeName, flags.hostRoot, flags.numDevices),
	})
	require.NoError(t, err)

	pluginDir := filepath.Join(tmpDir, driverName)
	require.NoError(t, os.MkdirAll(pluginDir, 0750))

	cdiSpecPath := filepath.Join(tmpDir, "k8s.gcu.enflame.com-gcu-"+claimUID+"-0.yaml")
	require.NoError(t, os.WriteFile(cdiSpecPath, []byte("placeholder"), 0600))

	corruptData := []byte("THIS IS NOT VALID CHECKPOINT JSON")
	require.NoError(t, os.WriteFile(
		filepath.Join(pluginDir, DriverPluginCheckpointFile),
		corruptData,
		0600,
	))

	err = state.Unprepare(types.UID(claimUID))
	require.Error(t, err, "Unprepare must return an error on an unreadable checkpoint")
	assert.Contains(t, err.Error(), "unable to read checkpoint")

	data, readErr := os.ReadFile(state.checkpointPath)
	require.NoError(t, readErr)
	assert.Equal(t, corruptData, data, "corrupt checkpoint file must not be modified by Unprepare")

	_, statErr := os.Stat(cdiSpecPath)
	assert.NoError(t, statErr, "CDI spec file must still be present after Unprepare returns an error")
}

func TestPrepareRestoredClaimRecreatesMissingClaimSpec(t *testing.T) {
	const (
		nodeName   = "test-node"
		driverName = gcu.DefaultDriverName
	)

	root := t.TempDir()
	state := newTestGCUDeviceState(t, root, nodeName, driverName, 1)
	claim := testGCUClaim(driverName, nodeName)

	prepared, err := state.Prepare(context.Background(), claim)
	require.NoError(t, err)
	require.NotEmpty(t, prepared)
	assertClaimSpecResolvesPreparedDevices(t, state, claim.UID, prepared)

	require.NoError(t, os.Remove(claimSpecPath(state, claim.UID)))

	restartedState := newTestGCUDeviceState(t, root, nodeName, driverName, 1)
	restored, err := restartedState.Prepare(context.Background(), claim)
	require.NoError(t, err)
	require.NotEmpty(t, restored)

	assert.Equal(t, prepared.GetDevices(), restored.GetDevices())
	assertClaimSpecResolvesPreparedDevices(t, restartedState, claim.UID, restored)

	checkpoint, err := readCheckpoint(restartedState.checkpointPath, restartedState.checkpointDecoder)
	require.NoError(t, err)
	require.Len(t, checkpoint.PreparedClaims, 1)
	assert.Equal(t, claim.UID, checkpoint.PreparedClaims[0].UID)
}

func TestPrepareRestoredClaimIsIdempotentWhenClaimSpecExists(t *testing.T) {
	const (
		nodeName   = "test-node"
		driverName = gcu.DefaultDriverName
	)

	root := t.TempDir()
	state := newTestGCUDeviceState(t, root, nodeName, driverName, 1)
	claim := testGCUClaim(driverName, nodeName)

	prepared, err := state.Prepare(context.Background(), claim)
	require.NoError(t, err)
	require.NotEmpty(t, prepared)
	assertClaimSpecResolvesPreparedDevices(t, state, claim.UID, prepared)

	restartedState := newTestGCUDeviceState(t, root, nodeName, driverName, 1)
	restored, err := restartedState.Prepare(context.Background(), claim)
	require.NoError(t, err)
	require.NotEmpty(t, restored)

	assert.Equal(t, prepared.GetDevices(), restored.GetDevices())
	assertClaimSpecResolvesPreparedDevices(t, restartedState, claim.UID, restored)
}

func TestPrepareRestoredClaimFailsWhenClaimSpecCannotBeRecreated(t *testing.T) {
	const (
		nodeName   = "test-node"
		driverName = gcu.DefaultDriverName
	)

	root := t.TempDir()
	state := newTestGCUDeviceState(t, root, nodeName, driverName, 1)
	claim := testGCUClaim(driverName, nodeName)

	_, err := state.Prepare(context.Background(), claim)
	require.NoError(t, err)

	require.NoError(t, os.Remove(claimSpecPath(state, claim.UID)))

	restartedState := newTestGCUDeviceState(t, root, nodeName, driverName, 1)
	require.NoError(t, os.Mkdir(claimSpecPath(restartedState, claim.UID), 0750))

	restored, err := restartedState.Prepare(context.Background(), claim)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to recreate CDI spec file for claim from checkpoint")
	assert.Nil(t, restored)

	checkpoint, readErr := readCheckpoint(restartedState.checkpointPath, restartedState.checkpointDecoder)
	require.NoError(t, readErr)
	require.Len(t, checkpoint.PreparedClaims, 1)
	assert.Equal(t, claim.UID, checkpoint.PreparedClaims[0].UID)
}

func newTestGCUDeviceState(t *testing.T, root, nodeName, driverName string, numDevices int) *DeviceState {
	t.Helper()
	useStubGCUDiscover(t, numDevices)

	flags := &Flags{
		cdiRoot:                     root,
		hostRoot:                    setupFakeHostRoot(t, numDevices),
		driverName:                  driverName,
		profile:                     gcu.ProfileName,
		nodeName:                    nodeName,
		numDevices:                  numDevices,
		kubeletPluginsDirectoryPath: root,
	}
	require.NoError(t, os.MkdirAll(filepath.Join(root, driverName), 0750))

	state, err := NewDeviceState(&Config{
		flags:   flags,
		profile: gcu.NewProfile(nodeName, flags.hostRoot, flags.numDevices),
	})
	require.NoError(t, err)
	return state
}

func testGCUClaim(driverName, nodeName string) *resourceapi.ResourceClaim {
	return &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "claim-uid",
			Name:      "claim",
			Namespace: "default",
		},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{
					Results: []resourceapi.DeviceRequestAllocationResult{
						{
							Request: "gcu",
							Driver:  driverName,
							Pool:    nodeName,
							Device:  "gcu-0",
						},
					},
				},
			},
		},
	}
}

func claimSpecPath(state *DeviceState, claimUID types.UID) string {
	specName := cdiapi.GenerateTransientSpecName(state.cdi.Vendor(), state.cdi.Class(), string(claimUID))
	return filepath.Join(filepath.Dir(state.checkpointPath), "..", specName+".yaml")
}

func assertClaimSpecResolvesPreparedDevices(t *testing.T, state *DeviceState, claimUID types.UID, prepared PreparedDevices) {
	t.Helper()

	specBytes, err := os.ReadFile(claimSpecPath(state, claimUID))
	require.NoError(t, err)

	var spec cdispec.Spec
	require.NoError(t, yaml.Unmarshal(specBytes, &spec))
	require.Equal(t, state.cdi.Kind(), spec.Kind)

	specDevices := map[string]cdispec.Device{}
	for _, device := range spec.Devices {
		specDevices[device.Name] = device
	}

	resolved := 0
	for _, device := range prepared {
		for _, cdiDeviceID := range device.CdiDeviceIds {
			_, _, deviceName, err := cdiparser.ParseQualifiedName(cdiDeviceID)
			require.NoError(t, err)
			if deviceName == cdiBaseDeviceName {
				continue
			}
			require.Contains(t, specDevices, deviceName)
			cdiDevice := specDevices[deviceName]
			require.NotEmpty(t, cdiDevice.ContainerEdits.DeviceNodes, "claim device %s should inject /dev/gcuN", deviceName)
			paths := make([]string, 0, len(cdiDevice.ContainerEdits.DeviceNodes))
			for _, node := range cdiDevice.ContainerEdits.DeviceNodes {
				paths = append(paths, node.Path)
			}
			assert.Contains(t, paths, "/dev/gcu0")
			resolved++
		}
	}
	assert.Greater(t, resolved, 0, "expected at least one claim-specific CDI device ID")
}

func TestNewDeviceStateCleansOrphanedClaimSpecs(t *testing.T) {
	const (
		nodeName   = "test-node"
		driverName = gcu.DefaultDriverName
	)

	root := t.TempDir()
	state := newTestGCUDeviceState(t, root, nodeName, driverName, 1)
	claim := testGCUClaim(driverName, nodeName)

	_, err := state.Prepare(context.Background(), claim)
	require.NoError(t, err)
	require.FileExists(t, claimSpecPath(state, claim.UID))

	orphanUID := types.UID("orphan-claim-uid")
	require.NoError(t, state.cdi.CreateClaimSpec(string(orphanUID), []gcucdi.Device{{
		Name:  "gcu-0",
		Index: 0,
	}}))
	require.FileExists(t, claimSpecPath(state, orphanUID))

	restarted := newTestGCUDeviceState(t, root, nodeName, driverName, 1)
	require.FileExists(t, claimSpecPath(restarted, claim.UID))
	_, err = os.Stat(claimSpecPath(restarted, orphanUID))
	assert.True(t, os.IsNotExist(err), "orphaned claim CDI spec should be removed on startup")
}

func TestMergeDeviceStatusPreservesConditions(t *testing.T) {
	cond := metav1.Condition{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Ready"}
	existing := []resourceapi.AllocatedDeviceStatus{
		{Driver: gcu.DefaultDriverName, Pool: "node-a", Device: "gcu-0", Conditions: []metav1.Condition{cond}},
	}
	updates := []resourceapi.AllocatedDeviceStatus{
		{Driver: gcu.DefaultDriverName, Pool: "node-a", Device: "gcu-0", Data: &runtime.RawExtension{Raw: []byte(`{"uuid":"x"}`)}},
		{Driver: gcu.DefaultDriverName, Pool: "node-a", Device: "gcu-1", Data: &runtime.RawExtension{Raw: []byte(`{"uuid":"y"}`)}},
	}

	merged := mergeDeviceStatus(existing, updates)

	require.Len(t, merged, 2)
	assert.Equal(t, "gcu-0", merged[0].Device)
	assert.Equal(t, []metav1.Condition{cond}, merged[0].Conditions, "existing conditions must be preserved")
	assert.Equal(t, updates[0].Data, merged[0].Data)
	assert.Equal(t, updates[1], merged[1], "unmatched entries are appended")
	assert.Len(t, existing[0].Conditions, 1, "input must not be mutated")
}

func TestMergeDeviceStatusDistinguishesShareID(t *testing.T) {
	shareA, shareB := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	existing := []resourceapi.AllocatedDeviceStatus{
		{Driver: gcu.DefaultDriverName, Pool: "node-a", Device: "gcu-0", ShareID: &shareA, Conditions: []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue}}},
	}
	updates := []resourceapi.AllocatedDeviceStatus{
		{Driver: gcu.DefaultDriverName, Pool: "node-a", Device: "gcu-0", ShareID: &shareB, Data: &runtime.RawExtension{Raw: []byte(`{}`)}},
	}

	merged := mergeDeviceStatus(existing, updates)

	require.Len(t, merged, 2)
	assert.Equal(t, &shareA, merged[0].ShareID)
	assert.Len(t, merged[0].Conditions, 1)
	assert.Equal(t, &shareB, merged[1].ShareID)
}
