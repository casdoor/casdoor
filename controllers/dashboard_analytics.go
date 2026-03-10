// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

// Package controllers. dashboard_analytics exposes HTTP handlers for the
// analytics dashboard endpoints (admin and user variants).
package controllers

import "github.com/casdoor/casdoor/object"

// -------------------------------------------- Public Functions --------------------------------------------

// GetAdminDashboardAnalytics
// @Title GetAdminDashboardAnalytics
// @Tag System API
// @Description Get aggregated analytics for the admin dashboard:
//
//	total users, weekly login trend, top N apps, and real-time activity.
//
// @Param owner         query string false "Organization name; use 'All' or omit for global"
// @Param topAppsLimit  query int    false "Number of top apps to return (default: 5)"
// @Param period        query string false "Time window for top apps: 'day', 'week' (default), 'month'"
// @Success 200 {object} object.AdminDashboardAnalytics
// @router /get-admin-dashboard-analytics [get]
func (c *ApiController) GetAdminDashboardAnalytics() {
	owner := c.Ctx.Input.Query("owner")
	topAppsLimit := c.GetIntWithDefault("topAppsLimit", 5)
	period := object.Period(c.Ctx.Input.Query("period"))

	data, err := object.GetAdminDashboardAnalytics(owner, topAppsLimit, period)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}

// GetUserDashboardAnalytics
// @Title GetUserDashboardAnalytics
// @Tag System API
// @Description Get analytics scoped to a single user:
//
//	top used applications and 24-hour activity heatmap.
//
// @Param owner   query string true  "Organization name"
// @Param userId  query string true  "The user's 'name' field"
// @Param period  query string false "Time window for top apps: 'day', 'week' (default), 'month'"
// @Success 200 {object} object.UserDashboardAnalytics
// @router /get-user-dashboard-analytics [get]
func (c *ApiController) GetUserDashboardAnalytics() {
	owner := c.Ctx.Input.Query("owner")
	userId := c.Ctx.Input.Query("userId")
	period := object.Period(c.Ctx.Input.Query("period"))

	if userId == "" {
		c.ResponseError("userId is required")
		return
	}

	data, err := object.GetUserDashboardAnalytics(owner, userId, period)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(data)
}
