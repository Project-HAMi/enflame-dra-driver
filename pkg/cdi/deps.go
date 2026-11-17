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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/klog/v2"
	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

const (
	gcuctlDeviceName = "gcuctl"
	firmwareRelPath  = "lib/firmware/enflame_ucode_bin"
	efsmiRelPath     = "usr/bin/efsmi"
	libefmlGlob      = "usr/lib/libefml*"
)

var mountOptions = []string{"ro", "nosuid", "nodev", "bind"}

// hostPath joins hostRoot with an absolute-looking path, treating path as
// relative to hostRoot even when it starts with "/".
func hostPath(hostRoot, path string) string {
	cleaned := strings.TrimPrefix(filepath.Clean(path), string(filepath.Separator))
	if hostRoot == "" || hostRoot == "/" {
		return filepath.Clean("/" + cleaned)
	}
	return filepath.Join(hostRoot, cleaned)
}

// hostDevPath returns the on-disk path of a device node under the configured
// host and device roots.
func (m *Manager) hostDevPath(name string) string {
	dev := filepath.Join(m.devRoot, name)
	return hostPath(m.hostRoot, dev)
}

// hostAbsPath returns the absolute host path that should appear in CDI specs
// (without the hostRoot prefix used for discovery inside a container).
func hostAbsPath(path string) string {
	cleaned := filepath.Clean(path)
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	return cleaned
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func bindMount(hostPath string) *cdispec.Mount {
	abs := hostAbsPath(hostPath)
	return &cdispec.Mount{
		HostPath:      abs,
		ContainerPath: abs,
		Options:       append([]string(nil), mountOptions...),
		Type:          "bind",
	}
}

func deviceNode(hostPath string) *cdispec.DeviceNode {
	return &cdispec.DeviceNode{
		Path: hostAbsPath(hostPath),
	}
}

// discoverBaseEdits builds the common container edits for all GCU claims:
// /dev/gcuctl plus optional firmware/efsmi/libefml mounts.
func (m *Manager) discoverBaseEdits() (cdispec.ContainerEdits, error) {
	edits := cdispec.ContainerEdits{}

	gcuctl := m.hostDevPath(gcuctlDeviceName)
	if !exists(gcuctl) {
		return edits, fmt.Errorf("required device node missing: %s", hostAbsPath(filepath.Join(m.devRoot, gcuctlDeviceName)))
	}
	edits.DeviceNodes = append(edits.DeviceNodes, deviceNode(filepath.Join(m.devRoot, gcuctlDeviceName)))

	firmware := hostPath(m.hostRoot, firmwareRelPath)
	if exists(firmware) {
		edits.Mounts = append(edits.Mounts, bindMount(firmwareRelPath))
	} else {
		klog.Background().Info("optional GCU firmware directory missing; skipping mount", "path", firmware)
	}

	efsmi := hostPath(m.hostRoot, efsmiRelPath)
	if exists(efsmi) {
		edits.Mounts = append(edits.Mounts, bindMount(efsmiRelPath))
	} else {
		klog.Background().Info("optional efsmi binary missing; skipping mount", "path", efsmi)
	}

	libMatches, err := filepath.Glob(hostPath(m.hostRoot, libefmlGlob))
	if err != nil {
		return edits, fmt.Errorf("glob libefml: %w", err)
	}
	sort.Strings(libMatches)
	if len(libMatches) == 0 {
		klog.Background().Info("optional libefml libraries missing; skipping mounts", "pattern", hostPath(m.hostRoot, libefmlGlob))
	}
	for _, match := range libMatches {
		rel, err := filepath.Rel(m.hostRoot, match)
		if err != nil || strings.HasPrefix(rel, "..") {
			// Fall back to stripping a leading hostRoot prefix for absolute paths.
			rel = strings.TrimPrefix(match, strings.TrimSuffix(m.hostRoot, "/")+"/")
			if m.hostRoot == "/" {
				rel = strings.TrimPrefix(match, "/")
			}
		}
		edits.Mounts = append(edits.Mounts, bindMount(rel))
	}

	return edits, nil
}

// discoverDeviceEdits builds per-device container edits for /dev/gcuN and
// optionally /dev/gcuNvid*. minor is the char-device minor; < 0 skips nodes.
func (m *Manager) discoverDeviceEdits(minor int, videoNodes bool) (cdispec.ContainerEdits, error) {
	edits := cdispec.ContainerEdits{}
	if minor < 0 {
		return edits, nil
	}

	name := fmt.Sprintf("gcu%d", minor)
	path := m.hostDevPath(name)
	if !exists(path) {
		return edits, fmt.Errorf("required device node missing: %s", hostAbsPath(filepath.Join(m.devRoot, name)))
	}
	edits.DeviceNodes = append(edits.DeviceNodes, deviceNode(filepath.Join(m.devRoot, name)))

	if !videoNodes {
		return edits, nil
	}

	pattern := m.hostDevPath(fmt.Sprintf("gcu%dvid*", minor))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return edits, fmt.Errorf("glob video nodes: %w", err)
	}
	sort.Strings(matches)
	for _, match := range matches {
		base := filepath.Base(match)
		edits.DeviceNodes = append(edits.DeviceNodes, deviceNode(filepath.Join(m.devRoot, base)))
	}
	if len(matches) == 0 {
		klog.Background().Info("no video device nodes found for GCU", "minor", minor, "pattern", pattern)
	}

	return edits, nil
}
