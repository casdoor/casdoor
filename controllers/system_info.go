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
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type SystemInfo struct {
	MemoryUsed  uint64    `json:"memory_used"`
	MemoryTotal uint64    `json:"memory_total"`
	CpuUsage    []float64 `json:"cpu_usage"`
}

// GetSystemInfo
// @Title GetSystemInfo
// @Tag System API
// @Description get user's system info
// @Param   id    query    string  true        "The id of the user"
// @Success 200 {object} object.SystemInfo The Response object
// @router /get-system-info [get]
func (c *ApiController) GetSystemInfo() {
	id := c.GetString("id")
	if id == "" {
		id = c.GetSessionUsername()
	}

	user := object.GetUser(id)
	if user == nil || !user.IsGlobalAdmin {
		c.ResponseError(c.T("ResourceErr.NotAuthorized"))
		return
	}

	cpuUsage, err := util.GetCpuUsage()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	memoryUsed, memoryTotal, err := util.GetMemoryUsage()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = SystemInfo{
		CpuUsage:    cpuUsage,
		MemoryUsed:  memoryUsed,
		MemoryTotal: memoryTotal,
	}
	c.ServeJSON()
}

// GitRepoVersion
// @Title GitRepoVersion
// @Tag System API
// @Description get local github repo's latest release version info
// @Success 200 {string} local latest version hash of casdoor
// @router /get-release [get]
func (c *ApiController) GitRepoVersion() {
	version, err := util.GetGitRepoVersion()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = version
	c.ServeJSON()
}
