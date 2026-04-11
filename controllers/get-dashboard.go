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

// GetDashboardProviderDistribution
// @Title GetDashboardProviderDistribution
// @Tag System API
// @Description get provider type distribution for dashboard
// @Success 200 {object} controllers.Response The Response object
// @router /get-dashboard-providers [get]
func (c *ApiController) GetDashboardProviderDistribution() {
	owner := c.Ctx.Input.Query("owner")

	data, err := object.GetDashboardProviderDistribution(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}

// GetDashboardMfaCoverage
// @Title GetDashboardMfaCoverage
// @Tag System API
// @Description get MFA adoption coverage stats for dashboard
// @Success 200 {object} controllers.Response The Response object
// @router /get-dashboard-mfa [get]
func (c *ApiController) GetDashboardMfaCoverage() {
	owner := c.Ctx.Input.Query("owner")

	data, err := object.GetDashboardMfaCoverage(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}

// GetDashboardLoginHeatmap
// @Title GetDashboardLoginHeatmap
// @Tag System API
// @Description get system activity heatmap data for dashboard
// @Success 200 {object} controllers.Response The Response object
// @router /get-dashboard-heatmap [get]
func (c *ApiController) GetDashboardLoginHeatmap() {
	owner := c.Ctx.Input.Query("owner")

	data, err := object.GetDashboardLoginHeatmap(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}
