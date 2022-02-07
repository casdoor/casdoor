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

func TestParseInt(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"Should be return zero when value is empty", "", 0},
		{"Should be return 0", "0", 0},
		{"Should be return 5", "5", 5},
		{"Should be return 10", "10", 10},
		{"Should be return -1", "-1", -1},
		{"Should be return -5", "-5", -5},
		{"Should be return -10", "-10", -10},
		{"Should be return -10", "string", "panic"},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			if scenery.expected == "panic" {
				defer func() {
					if r := recover(); r == nil {
						t.Error("function should panic")
					}
				}()
				ParseInt(scenery.input)

			} else {
				actual := ParseInt(scenery.input)
				assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"Should be return false", "0", false},
		{"Should be return true", "5", true},
		{"Should be return true", "10", true},
		{"Should be return true", "-1", true},
		{"Should be return false", "", false},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := ParseBool(scenery.input)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}

func TestBoolToString(t *testing.T) {
	scenarios := []struct {
		description string
		input       bool
		expected    interface{}
	}{
		{"Should be return 1", true, "1"},
		{"Should be return 0", false, "0"},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := BoolToString(scenery.input)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}

func TestCamelToSnakeCase(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"Should be return casdor_is_the_best", "CasdoorIsTheBest", "casdoor_is_the_best"},
		{"Should be return Lorem Ipsum", "Lorem Ipsum", "lorem_ipsum"},
		{"Should be return Lorem Ipsum", "lorem Ipsum", "lorem_ipsum"},
		{"Should be return Lorem Ipsum", "lorem ipsum", "loremipsum"},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := CamelToSnakeCase(scenery.input)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}

func TestIsStrsEmpty(t *testing.T) {
	scenarios := []struct {
		description string
		input       []string
		expected    interface{}
	}{
		{"Should be return true if one is empty", []string{"", "lorem", "ipsum"}, true},
		{"Should be return true if is empty", []string{""}, true},
		{"Should be return false all is a valid string", []string{"lorem", "ipsum"}, false},
		{"Should be return false is function called with empty parameters", []string{}, false},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := IsStrsEmpty(scenery.input...)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}

func TestGetMaxLenStr(t *testing.T) {
	scenarios := []struct {
		description string
		input       []string
		expected    interface{}
	}{
		{"Should be return casdoor", []string{"", "casdoor", "casbin"}, "casdoor"},
		{"Should be return casdoor_jdk", []string{"", "casdoor", "casbin", "casdoor_jdk"}, "casdoor_jdk"},
		{"Should be return empty string", []string{""}, ""},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := GetMaxLenStr(scenery.input...)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}

func TestGetMinLenStr(t *testing.T) {
	scenarios := []struct {
		description string
		input       []string
		expected    interface{}
	}{
		{"Should be return casbin", []string{"casdoor", "casbin"}, "casbin"},
		{"Should be return casbin", []string{"casdoor", "casbin", "casdoor_jdk"}, "casbin"},
		{"Should be return empty string", []string{"a", "", "casbin"}, ""},
		{"Should be return a", []string{"a", "casdoor", "casbin"}, "a"},
		{"Should be return a", []string{"casdoor", "a", "casbin"}, "a"},
		{"Should be return a", []string{"casbin", "casdoor", "a"}, "a"},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := GetMinLenStr(scenery.input...)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}

func TestSnakeString(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"Should be return casdor_is_the_best", "CasdoorIsTheBest", "casdoor_is_the_best"},
		{"Should be return Lorem Ipsum", "Lorem Ipsum", "lorem_ipsum"},
		{"Should be return Lorem Ipsum", "lorem Ipsum", "lorem_ipsum"},
		{"Should be return Lorem Ipsum", "lorem ipsum", "loremipsum"},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := SnakeString(scenery.input)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}
