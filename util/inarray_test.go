// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInArray(t *testing.T) {

	scenarios := []struct {
		description string
		input       []interface{}
		expected    []interface{}
	}{
		{"scenery one", []interface{}{"str1", []string{"str1", "str2", "str3", "str4"}}, []interface{}{true, 0}},
		{"scenery two", []interface{}{"str", []string{"str1", "str2", "str3", "str4"}}, []interface{}{false, -1}},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			exists, index := InArray(scenery.input[0], scenery.input[1])
			actual := []interface{}{exists, index}
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}
