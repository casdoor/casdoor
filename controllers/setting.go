// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/object"
)

// GetSetting
// @Title GetSetting
// @Tag Setting API
// @Description get the site-wide setting
// @Success 200 {object} object.Setting The Response object
// @router /get-setting [get]
func (c *ApiController) GetSetting() {
	if !c.RequireGlobalAdmin() {
		return
	}

	setting, err := object.GetSetting()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(setting)
}

// UpdateSetting
// @Title UpdateSetting
// @Tag Setting API
// @Description update the site-wide setting
// @Param   body    body   object.Setting  true        "The details of the setting"
// @Success 200 {object} controllers.Response The Response object
// @router /update-setting [post]
func (c *ApiController) UpdateSetting() {
	if !c.RequireGlobalAdmin() {
		return
	}

	var setting object.Setting
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &setting)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateSetting(&setting))
	c.ServeJSON()
}
