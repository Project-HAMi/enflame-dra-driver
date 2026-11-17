/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadyEndpoint(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(readyHandler))
	t.Cleanup(s.Close)

	res, err := http.Get(s.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestGCUProfile(t *testing.T) {
	require.Contains(t, validProfiles, "gcu")
	assert.Equal(t, "enflame-dra-webhook", newApp().Name)
}
