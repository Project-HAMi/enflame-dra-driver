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

const (
	defaultCDIRoot    = "/var/run/cdi"
	defaultHostRoot   = "/"
	defaultDevRoot    = "/dev"
	defaultDriverName = "gcu.enflame.com"
	defaultClass      = "gcu"
)

// Option configures a Manager.
type Option func(*Manager)

// WithCDIRoot sets the directory where CDI spec files are written.
func WithCDIRoot(root string) Option {
	return func(m *Manager) {
		if root != "" {
			m.cdiRoot = root
		}
	}
}

// WithHostRoot sets the host filesystem root used when discovering driver
// dependencies (efsmi, libefml). Paths written into CDI specs are
// still absolute host paths without this prefix.
func WithHostRoot(root string) Option {
	return func(m *Manager) {
		if root != "" {
			m.hostRoot = root
		}
	}
}

// WithDevRoot sets the device node root (typically "/dev"). When HostRoot is
// not "/", device nodes are looked up under HostRoot+DevRoot.
func WithDevRoot(root string) Option {
	return func(m *Manager) {
		if root != "" {
			m.devRoot = root
		}
	}
}

// WithDriverName sets the DRA driver name used for CDI vendor
// (k8s.<driverName>) and environment variables.
func WithDriverName(name string) Option {
	return func(m *Manager) {
		if name != "" {
			m.driverName = name
		}
	}
}

// WithClass sets the CDI class (default "gcu").
func WithClass(class string) Option {
	return func(m *Manager) {
		if class != "" {
			m.class = class
		}
	}
}

// WithNodeName sets the node name injected as KUBERNETES_NODE_NAME.
func WithNodeName(name string) Option {
	return func(m *Manager) {
		m.nodeName = name
	}
}
