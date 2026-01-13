// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

// GetDashboard
// @Title GetDashboard
// @Tag System API
// @Description get information of dashboard
// @Success 200 {object} controllers.Response The Response object
// @router /get-dashboard [get]
func (c *ApiController) GetDashboard() {
	owner := c.Ctx.Input.Query("owner")

	data, err := object.GetDashboard(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}

// GetDashboardUsersByProvider
// @Title GetDashboardUsersByProvider
// @Tag System API
// @Description get user count by provider
// @Success 200 {object} controllers.Response The Response object
// @router /get-dashboard-users-by-provider [get]
func (c *ApiController) GetDashboardUsersByProvider() {
	owner := c.Ctx.Input.Query("owner")

	data, err := object.GetDashboardUsersByProvider(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}

// GetDashboardLoginHeatmap
// @Title GetDashboardLoginHeatmap
// @Tag System API
// @Description get login count heatmap by weekday and hour
// @Success 200 {object} controllers.Response The Response object
// @router /get-dashboard-login-time-heatmap [get]
func (c *ApiController) GetDashboardLoginHeatmap() {
	owner := c.Ctx.Input.Query("owner")

	data, err := object.GetDashboardLoginHeatmap(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}

// GetDashboardResourcesByProvider
// @Title GetDashboardResourcesByProvider
// @Tag System API
// @Description get resource count by provider
// @Success 200 {object} controllers.Response The Response object
// @router /get-dashboard-resources-by-provider [get]
func (c *ApiController) GetDashboardResourcesByProvider() {
	owner := c.Ctx.Input.Query("owner")

	data, err := object.GetDashboardResourcesByProvider(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}
