// Copyright 2021 The casbin Authors. All Rights Reserved.
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

func TestGetUploadXlsxPath(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"scenery one", "casdoor", "tmpFiles/casdoor.xlsx"},
		{"scenery two", "casbin", "tmpFiles/casbin.xlsx"},
		{"scenery three", "loremIpsum", "tmpFiles/loremIpsum.xlsx"},
		{"scenery four", "", "tmpFiles/.xlsx"},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := GetUploadXlsxPath(scenery.input)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}
