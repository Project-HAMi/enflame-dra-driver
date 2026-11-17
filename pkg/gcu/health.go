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
	"strings"
)

// HealthLevel is the DRA-facing health classification for a GCU.
type HealthLevel string

const (
	// HealthLevelHealthy means device_health=healthy and device_state=working.
	HealthLevelHealthy HealthLevel = "Healthy"
	// HealthLevelUnhealthy means at least one health signal is bad.
	HealthLevelUnhealthy HealthLevel = "Unhealthy"
	// HealthLevelUnknown means sysfs health inputs could not be read.
	HealthLevelUnknown HealthLevel = "Unknown"
)

const (
	expectedDeviceHealth = "healthy"
	expectedDeviceState  = "working"
)

// Health is the result of reading sysfs device_health / device_state for one GCU.
type Health struct {
	Level   HealthLevel
	Health  string // raw device_health
	State   string // raw device_state
	Message string
}

// CheckHealth reads /sys/class/enflame/gcu<minor>/device/{device_health,device_state}.
// minor is the char-device minor (same N as /dev/gcuN and DRA name gcu-N).
// Discovery remains EFML-only; sysfs is used here only as the health signal source.
func CheckHealth(minor int, opts ...Option) Health {
	d := newDiscoverer(opts...)
	deviceDir := hostPath(d.hostRoot, filepath.Join(d.classPath, fmt.Sprintf("gcu%d", minor), "device"))

	healthRaw, healthErr := readTrimmed(filepath.Join(deviceDir, "device_health"))
	stateRaw, stateErr := readTrimmed(filepath.Join(deviceDir, "device_state"))

	if healthErr != nil && stateErr != nil {
		return Health{
			Level:   HealthLevelUnknown,
			Message: fmt.Sprintf("device_health: %v; device_state: %v", healthErr, stateErr),
		}
	}

	h := Health{Health: healthRaw, State: stateRaw}
	switch {
	case healthErr != nil:
		h.Level = HealthLevelUnknown
		h.Message = fmt.Sprintf("device_health: %v; device_state=%q", healthErr, stateRaw)
	case stateErr != nil:
		h.Level = HealthLevelUnknown
		h.Message = fmt.Sprintf("device_health=%q; device_state: %v", healthRaw, stateErr)
	case strings.EqualFold(healthRaw, expectedDeviceHealth) && strings.EqualFold(stateRaw, expectedDeviceState):
		h.Level = HealthLevelHealthy
		h.Message = fmt.Sprintf("device_health=%s device_state=%s", healthRaw, stateRaw)
	default:
		h.Level = HealthLevelUnhealthy
		h.Message = fmt.Sprintf("device_health=%q device_state=%q", healthRaw, stateRaw)
	}
	return h
}
