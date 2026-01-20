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

package controllers

import (
	"errors"

	"github.com/casdoor/casdoor/util"
	"github.com/go-git/go-git/v5"
)

// GetSystemInfo
// @Title GetSystemInfo
// @Tag System API
// @Description get system info like CPU and memory usage
// @Success 200 {object} util.SystemInfo The Response object
// @router /get-system-info [get]
func (c *ApiController) GetSystemInfo() {
	_, ok := c.RequireAdmin()
	if !ok {
		return
	}

	systemInfo, err := util.GetSystemInfo()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(systemInfo)
}

// GetVersionInfo
// @Title GetVersionInfo
// @Tag System API
// @Description get version info like Casdoor release version and commit ID
// @Success 200 {object} util.VersionInfo The Response object
// @router /get-version-info [get]
func (c *ApiController) GetVersionInfo() {
	versionInfo, err := util.GetVersionInfo()
	if err != nil && !errors.Is(err, git.ErrRepositoryNotExists) {
		c.ResponseError(err.Error())
		return
	}

	if versionInfo.Version != "" {
		c.ResponseOk(versionInfo)
		return
	}

	c.ResponseOk(util.GetBuiltInVersionInfo())
}

// Health
// @Title Health
// @Tag System API
// @Description check if the system is live
// @Success 200 {object} controllers.Response The Response object
// @router /health [get]
func (c *ApiController) Health() {
	c.ResponseOk()
}
