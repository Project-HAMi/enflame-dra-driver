//go:build e2e

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

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
)

// The mock gpu/cpu/net e2e suite was removed with the example profiles.
// Stage 8 will restore Enflame GCU end-to-end coverage against real hardware.
var _ = Describe("Enflame GCU DRA e2e", func() {
	PIt("is reserved for stage 8", func() {})
})
