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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDRAName(t *testing.T) {
	idx, err := ParseDRAName("gcu-0")
	require.NoError(t, err)
	assert.Equal(t, 0, idx)

	idx, err = ParseDRAName("gcu-12")
	require.NoError(t, err)
	assert.Equal(t, 12, idx)

	_, err = ParseDRAName("gcu-")
	require.Error(t, err)
	_, err = ParseDRAName("gcu")
	require.Error(t, err)
	_, err = ParseDRAName("gcu-0vid0")
	require.Error(t, err)
	_, err = ParseDRAName("gpu-0")
	require.Error(t, err)
}
