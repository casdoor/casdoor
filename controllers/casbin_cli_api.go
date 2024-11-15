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

package controllers

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunCasbinCommand
// @Title RunCasbinCommand
// @Tag Enforcer API
// @Description Call Casbin CLI commands
// @Success 200 {object} controllers.Response The Response object
// @router /run-casbin-command [get]
func (c *ApiController) RunCasbinCommand() {
	language := c.Input().Get("language")
	argString := c.Input().Get("args")

	if language == "" {
		language = "go"
	}
	// use "casbin-go-cli" by default, can be also "casbin-java-cli", "casbin-node-cli", etc.
	// the pre-built binary of "casbin-go-cli" can be found at: https://github.com/casbin/casbin-go-cli/releases
	binaryName := fmt.Sprintf("casbin-%s-cli", language)

	_, err := exec.LookPath(binaryName)
	if err != nil {
		c.ResponseError(fmt.Sprintf("executable file: %s not found in PATH", binaryName))
		return
	}

	// argString's example:
	// enforce -m "examples/rbac_model.conf" -p "examples/rbac_policy.csv" "alice" "data1" "read"
	// see: https://github.com/jcasbin/casbin-java-cli?tab=readme-ov-file#get-started
	args := strings.Split(argString, " ")

	command := exec.Command(binaryName, args...)
	outputBytes, err := command.CombinedOutput()
	if outputBytes != nil {
		output := string(outputBytes)
		c.ResponseError(output)
		return
	}

	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	output := string(outputBytes)
	c.ResponseOk(output)
}
