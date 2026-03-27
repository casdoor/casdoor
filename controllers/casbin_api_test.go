// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnforceRequestBodySupportsLegacyArray(t *testing.T) {
	request, err := parseEnforceRequestBody([]byte(`["alice","data1","read"]`), []string{"sub", "obj", "act"})
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"alice", "data1", "read"}, request)
}

func TestParseEnforceRequestBodySupportsNamedObject(t *testing.T) {
	request, err := parseEnforceRequestBody([]byte(`{"sub":{"division_guid":"div-123"},"obj":{"division_guid":"div-123"},"act":"read"}`), []string{"sub", "obj", "act"})
	require.NoError(t, err)
	require.Len(t, request, 3)

	sub, ok := request[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "div-123", sub["division_guid"])

	obj, ok := request[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "div-123", obj["division_guid"])

	assert.Equal(t, "read", request[2])
}

func TestParseEnforceRequestBodyRejectsMissingToken(t *testing.T) {
	request, err := parseEnforceRequestBody([]byte(`{"sub":"alice","act":"read"}`), []string{"sub", "obj", "act"})
	assert.Nil(t, request)
	assert.EqualError(t, err, `the request body is missing "obj"`)
}

func TestParseBatchEnforceRequestBodySupportsNamedObjects(t *testing.T) {
	requests, err := parseBatchEnforceRequestBody([]byte(`[
		{"sub":{"division_guid":"div-123"},"obj":{"division_guid":"div-123"},"act":"read"},
		{"sub":{"division_guid":"div-123"},"obj":{"division_guid":"div-999"},"act":"read"}
	]`), []string{"sub", "obj", "act"})
	require.NoError(t, err)
	require.Len(t, requests, 2)
	require.Len(t, requests[0], 3)

	firstObj, ok := requests[0][1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "div-123", firstObj["division_guid"])
	assert.Equal(t, "read", requests[1][2])
}
