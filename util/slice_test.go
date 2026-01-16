// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasTagInSlice(t *testing.T) {
	scenarios := []struct {
		description string
		slice       []string
		userTag     string
		expected    bool
	}{
		{
			description: "Should return true when single tag matches",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "default-policy",
			expected:    true,
		},
		{
			description: "Should return true when comma-separated tags contain a match",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "default-policy,project-admin",
			expected:    true,
		},
		{
			description: "Should return true when comma-separated tags with spaces contain a match",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "default-policy, project-admin",
			expected:    true,
		},
		{
			description: "Should return true when one of multiple tags matches",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "other-tag,default-policy",
			expected:    true,
		},
		{
			description: "Should return false when no tags match",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "other-tag",
			expected:    false,
		},
		{
			description: "Should return false when no comma-separated tags match",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "other-tag,another-tag",
			expected:    false,
		},
		{
			description: "Should return false when userTag is empty",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "",
			expected:    false,
		},
		{
			description: "Should return false when slice is empty",
			slice:       []string{},
			userTag:     "default-policy",
			expected:    false,
		},
		{
			description: "Should return true when last tag in comma-separated list matches",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "other-tag,project-admin",
			expected:    true,
		},
		{
			description: "Should handle tags with extra spaces",
			slice:       []string{"default-policy", "project-admin"},
			userTag:     "  default-policy  ,  project-admin  ",
			expected:    true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			actual := HasTagInSlice(scenario.slice, scenario.userTag)
			assert.Equal(t, scenario.expected, actual, "The returned value is not as expected")
		})
	}
}

func TestInSlice(t *testing.T) {
	scenarios := []struct {
		description string
		slice       []string
		elem        string
		expected    bool
	}{
		{
			description: "Should return true when element is in slice",
			slice:       []string{"a", "b", "c"},
			elem:        "b",
			expected:    true,
		},
		{
			description: "Should return false when element is not in slice",
			slice:       []string{"a", "b", "c"},
			elem:        "d",
			expected:    false,
		},
		{
			description: "Should return false when slice is empty",
			slice:       []string{},
			elem:        "a",
			expected:    false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			actual := InSlice(scenario.slice, scenario.elem)
			assert.Equal(t, scenario.expected, actual, "The returned value is not as expected")
		})
	}
}

func TestHaveIntersection(t *testing.T) {
	scenarios := []struct {
		description string
		arr1        []string
		arr2        []string
		expected    bool
	}{
		{
			description: "Should return true when arrays have common elements",
			arr1:        []string{"a", "b", "c"},
			arr2:        []string{"c", "d", "e"},
			expected:    true,
		},
		{
			description: "Should return false when arrays have no common elements",
			arr1:        []string{"a", "b", "c"},
			arr2:        []string{"d", "e", "f"},
			expected:    false,
		},
		{
			description: "Should return false when one array is empty",
			arr1:        []string{"a", "b", "c"},
			arr2:        []string{},
			expected:    false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			actual := HaveIntersection(scenario.arr1, scenario.arr2)
			assert.Equal(t, scenario.expected, actual, "The returned value is not as expected")
		})
	}
}
