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
package conf

import (
	"os"
	"testing"

	"github.com/astaxie/beego"
	"github.com/stretchr/testify/assert"
)

func TestGetConfString(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"Should be return casbin", "appname", "casbin"},
		{"Should be return 8000", "httpport", "8000"},
		{"Should be return  value", "key", "value"},
	}

	//do some set up job

	os.Setenv("appname", "casbin")
	os.Setenv("key", "value")

	err := beego.LoadAppConfig("ini", "app.conf")
	assert.Nil(t, err)

	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := GetConfigString(scenery.input)
			assert.Equal(t, scenery.expected, actual)
		})
	}
}

func TestGetConfInt(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"Should be return 8000", "httpport", 8001},
		{"Should be return 8000", "verificationCodeTimeout", 10},
	}

	//do some set up job
	os.Setenv("httpport", "8001")

	err := beego.LoadAppConfig("ini", "app.conf")
	assert.Nil(t, err)

	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual, err := GetConfigInt64(scenery.input)
			assert.Nil(t, err)
			assert.Equal(t, scenery.expected, int(actual))
		})
	}
}

func TestGetConfBool(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"Should be return false", "SessionOn", false},
		{"Should be return false", "copyrequestbody", true},
	}

	//do some set up job
	os.Setenv("SessionOn", "false")

	err := beego.LoadAppConfig("ini", "app.conf")
	assert.Nil(t, err)
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual, err := GetConfigBool(scenery.input)
			assert.Nil(t, err)
			assert.Equal(t, scenery.expected, actual)
		})
	}
}
