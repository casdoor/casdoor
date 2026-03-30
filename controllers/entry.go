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

import "github.com/casdoor/casdoor/object"

// GetEntries
// @Title GetEntries
// @Tag Entry API
// @Description get entries by agent name
// @Param   agentName     query    string  false       "The name of the agent"
// @Param   owner         query    string  false       "The owner of the entries"
// @Success 200 {array} object.Entry The Response object
// @router /get-entries [get]
func (c *ApiController) GetEntries() {
	agentName := c.Ctx.Input.Query("agentName")
	owner := c.Ctx.Input.Query("owner")
	entries, err := object.GetEntries(owner, agentName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(entries)
}
