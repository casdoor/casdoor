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
	"encoding/json"
	"errors"
	"strings"

	"github.com/casdoor/casdoor/upgrade"
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
	if !c.RequireGlobalAdmin() {
		return
	}

	systemInfo, err := util.GetSystemInfo()
	if err != nil {
		c.ResponseErrorFrom(err)
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
		c.ResponseErrorFrom(err)
		return
	}

	if versionInfo.Version != "" {
		c.ResponseOk(versionInfo)
		return
	}

	c.ResponseOk(util.GetBuiltInVersionInfo())
}

// GetUpgradeInfo
// @Title GetUpgradeInfo
// @Tag System API
// @Description get the latest Casdoor release, how this Casdoor is deployed and the progress of an upgrade
// @Success 200 {object} upgrade.UpgradeInfo The Response object
// @router /get-upgrade-info [get]
func (c *ApiController) GetUpgradeInfo() {
	if !c.RequireGlobalAdmin() {
		return
	}

	c.ResponseOk(upgrade.GetUpgradeInfo())
}

// UpgradeSystem
// @Title UpgradeSystem
// @Tag System API
// @Description upgrade an official release binary of Casdoor to the latest release and restart it
// @Param   body    body   string  true        "The JSON {version}, the release to upgrade to, which must be the latest one"
// @Success 200 {object} controllers.Response The Response object
// @router /upgrade-system [post]
func (c *ApiController) UpgradeSystem() {
	if !c.RequireGlobalAdmin() {
		return
	}

	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"), "application/json") {
		c.ResponseError(c.T("general:the request body must be JSON"))
		return
	}

	var req struct {
		Version string `json:"version"`
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseErrorFrom(err)
		return
	}

	err = upgrade.StartUpgrade(req.Version, c.GetSessionUsername())
	if err != nil {
		c.ResponseErrorFrom(err)
		return
	}

	c.ResponseOk()
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
