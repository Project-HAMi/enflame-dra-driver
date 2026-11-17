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

const (
	defaultHostRoot  = "/"
	defaultClassPath = "/sys/class/enflame"
	defaultDevRoot   = "/dev"
)

// Option configures discovery.
type Option func(*discoverer)

type discoverer struct {
	hostRoot  string
	classPath string
	devRoot   string
}

func newDiscoverer(opts ...Option) *discoverer {
	d := &discoverer{
		hostRoot:  defaultHostRoot,
		classPath: defaultClassPath,
		devRoot:   defaultDevRoot,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// WithHostRoot sets the host filesystem root used when enriching EFML results
// with NUMA / PCIe root from sysfs topology. EFML itself always talks to the
// real driver on the running host.
func WithHostRoot(root string) Option {
	return func(d *discoverer) {
		if root != "" {
			d.hostRoot = root
		}
	}
}

// WithClassPath overrides the enflame class directory used for topology enrichment.
func WithClassPath(path string) Option {
	return func(d *discoverer) {
		if path != "" {
			d.classPath = path
		}
	}
}

// WithDevRoot sets the device node root used to build Device.Path (default /dev).
func WithDevRoot(root string) Option {
	return func(d *discoverer) {
		if root != "" {
			d.devRoot = root
		}
	}
}
