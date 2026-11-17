/*
 * Copyright The Kubernetes Authors.
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

package helm

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v4/pkg/chart/common"
	chartutil "helm.sh/helm/v4/pkg/chart/common/util"
	"helm.sh/helm/v4/pkg/chart/loader"
	"helm.sh/helm/v4/pkg/engine"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func TestDeviceHealthEnv(t *testing.T) {
	for _, tc := range []struct {
		name    string
		values  map[string]any
		enabled bool
	}{
		{name: "default", enabled: true},
		{name: "enabled", values: map[string]any{"deviceHealth": true}, enabled: true},
		{name: "disabled", values: map[string]any{"deviceHealth": false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := loader.Load("../../deployments/helm/enflame-dra-driver")
			require.NoError(t, err)
			overrides := map[string]any{}
			if tc.values != nil {
				overrides["kubeletPlugin"] = tc.values
			}
			values, err := chartutil.ToRenderValues(chart, overrides, common.ReleaseOptions{
				Name: "test", Namespace: "driver-test", IsInstall: true,
			}, common.DefaultCapabilities)
			require.NoError(t, err)
			rendered, err := engine.Render(chart, values)
			require.NoError(t, err)

			pluginPath := "enflame-dra-driver/templates/kubeletplugin.yaml"
			require.Contains(t, rendered, pluginPath)
			var ds appsv1.DaemonSet
			require.NoError(t, yaml.Unmarshal([]byte(rendered[pluginPath]), &ds))
			env := pluginContainerEnv(t, ds)
			assert.Equal(t, strconv.FormatBool(tc.enabled), env["DEVICE_HEALTH"])
			assert.Equal(t, "gcu.enflame.com", env["DRIVER_NAME"])
			assert.Equal(t, "gcu", env["DEVICE_PROFILE"])
			assert.Equal(t, "/host", env["HOST_ROOT"])
			assert.Equal(t, "/opt/enflame/lib", env["LD_LIBRARY_PATH"])
			assert.NotContains(t, env, "SIMULATE_HEALTH_CHANGES")
			assert.NotContains(t, env, "BINDING_CONDITIONS")
			assert.NotContains(t, env, "NUM_DEVICES")
			assert.Equal(t, []string{"enflame-dra-kubeletplugin"}, ds.Spec.Template.Spec.Containers[0].Command)
		})
	}
}

func pluginContainerEnv(t *testing.T, ds appsv1.DaemonSet) map[string]string {
	t.Helper()
	var plugin *corev1.Container
	for i := range ds.Spec.Template.Spec.Containers {
		if ds.Spec.Template.Spec.Containers[i].Name == "plugin" {
			plugin = &ds.Spec.Template.Spec.Containers[i]
		}
	}
	require.NotNil(t, plugin, "plugin container not found in rendered DaemonSet")
	env := map[string]string{}
	for _, e := range plugin.Env {
		if e.ValueFrom == nil {
			env[e.Name] = e.Value
		}
	}
	return env
}
