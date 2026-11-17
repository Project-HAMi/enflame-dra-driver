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

func TestParseDevID(t *testing.T) {
	uuid, err := parseDevID("84 82 54 87 49 56 50 50 48 52 48 52 0 0 0 0")
	require.NoError(t, err)
	assert.Equal(t, "TR6W18220404", uuid)

	uuid, err = parseDevID("TR6W18220404")
	require.NoError(t, err)
	assert.Equal(t, "TR6W18220404", uuid)

	_, err = parseDevID("")
	require.Error(t, err)
}

func TestParseMemoryBytes(t *testing.T) {
	n, err := parseMemoryBytes("0xbfe000000")
	require.NoError(t, err)
	assert.Equal(t, uint64(51506053120), n)

	n, err = parseMemoryBytes("1024")
	require.NoError(t, err)
	assert.Equal(t, uint64(1024), n)
}

func TestParseGCUIndex(t *testing.T) {
	idx, ok := parseGCUIndex("gcu0")
	assert.True(t, ok)
	assert.Equal(t, 0, idx)

	idx, ok = parseGCUIndex("gcu12")
	assert.True(t, ok)
	assert.Equal(t, 12, idx)

	_, ok = parseGCUIndex("gcu0vid0")
	assert.False(t, ok)
	_, ok = parseGCUIndex("gcuctl")
	assert.False(t, ok)
	_, ok = parseGCUIndex("gcu")
	assert.False(t, ok)
}

func TestParseFirmwareVersion(t *testing.T) {
	assert.Equal(t, "33.6.5", parseFirmwareVersion("firmware version: 33.6.5"))
	assert.Equal(t, "1.2.3", parseFirmwareVersion("1.2.3"))
}
