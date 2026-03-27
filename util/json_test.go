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

package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTryJsonToAnonymousStructConvertsSnakeCaseFields(t *testing.T) {
	value, err := TryJsonToAnonymousStruct(`{"division_guid":"div-123","has_branch":true,"child":{"branch_guid":"br-1"}}`)
	require.NoError(t, err)

	structValue := reflect.ValueOf(value)
	require.Equal(t, reflect.Ptr, structValue.Kind())

	elem := structValue.Elem()
	assert.Equal(t, "div-123", elem.FieldByName("DivisionGuid").String())
	assert.True(t, elem.FieldByName("HasBranch").Bool())

	child := elem.FieldByName("Child")
	require.Equal(t, reflect.Ptr, child.Kind())
	assert.Equal(t, "br-1", child.Elem().FieldByName("BranchGuid").String())
}

func TestTryJsonToAnonymousStructRejectsScalarValues(t *testing.T) {
	value, err := TryJsonToAnonymousStruct(`"alice"`)
	assert.Nil(t, value)
	assert.EqualError(t, err, "JSON value is not an object or array")
}

func TestConvertInterfaceArrayConvertsMapValues(t *testing.T) {
	result := ConvertInterfaceArray([]interface{}{
		map[string]interface{}{"is_admin": true, "division_guid": "div-123"},
		"read",
	})

	require.Len(t, result, 2)
	assert.Equal(t, "read", result[1])

	converted := reflect.ValueOf(result[0])
	require.Equal(t, reflect.Ptr, converted.Kind())
	assert.True(t, converted.Elem().FieldByName("IsAdmin").Bool())
	assert.Equal(t, "div-123", converted.Elem().FieldByName("DivisionGuid").String())
}
