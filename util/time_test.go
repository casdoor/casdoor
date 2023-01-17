// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetCurrentTime(t *testing.T) {
	test := GetCurrentTime()
	expected := time.Now().Format(time.RFC3339)

	assert.Equal(t, test, expected, "The times not are equals")

	types := reflect.TypeOf(test).Kind()
	assert.Equal(t, types, reflect.String, "GetCurrentUnixTime should be return string")
}

func Test_GetCurrentUnixTime_Shoud_Return_String(t *testing.T) {
	test := GetCurrentUnixTime()
	types := reflect.TypeOf(test).Kind()
	assert.Equal(t, types, reflect.String, "GetCurrentUnixTime should be return string")
}

func Test_IsTokenExpired(t *testing.T) {
	type input struct {
		createdTime string
		expiresIn   int
	}

	type testCases struct {
		description string
		input       input
		expected    bool
	}

	for _, scenario := range []testCases{
		{
			description: "Token emitted now is valid for 60 minutes",
			input: input{
				createdTime: time.Now().Format(time.RFC3339),
				expiresIn:   60,
			},
			expected: false,
		},
		{
			description: "Token emitted 60 minutes before now is valid for 60 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Minute * 60).Format(time.RFC3339),
				expiresIn:   61,
			},
			expected: false,
		},
		{
			description: "Token emitted 2 hours before now is Expired after 60 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
				expiresIn:   60,
			},
			expected: true,
		},
		{
			description: "Token emitted 61 minutes before now is Expired after 60 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Minute * 61).Format(time.RFC3339),
				expiresIn:   60,
			},
			expected: true,
		},
		{
			description: "Token emitted 2 hours before now  is valid for 120 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
				expiresIn:   121,
			},
			expected: false,
		},
		{
			description: "Token emitted 159 minutes before now is Expired after 60 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Minute * 159).Format(time.RFC3339),
				expiresIn:   120,
			},
			expected: true,
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			result := IsTokenExpired(scenario.input.createdTime, scenario.input.expiresIn)
			assert.Equal(t, scenario.expected, result, fmt.Sprintf("Expected %t, but was founded %t", scenario.expected, result))
		})
	}
}
