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
	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetOrganizationUsers
// @Title GetOrganizationUsers
// @Tag User API
// @Description search users of an organization with multiple filters, sorting and pagination
// @Param   owner         path    string  true   "The organization id"
// @Param   name          query   string  false  "Filter by login (substring match)"
// @Param   displayName   query   string  false  "Filter by display name (substring match)"
// @Param   email         query   string  false  "Filter by email (substring match)"
// @Param   phone         query   string  false  "Filter by phone (substring match)"
// @Param   affiliation   query   string  false  "Filter by affiliation (substring match)"
// @Param   pageSize      query   int     false  "Page size"
// @Param   p             query   int     false  "Page number"
// @Param   sortField     query   string  false  "Sort field (same as GetUsers)"
// @Param   sortOrder     query   string  false  "Sort order: ascend or descend (same as GetUsers)"
// @Success 200 {array} object.User The Response object
// @router /:owner/get-users [get]
func (c *ApiController) GetOrganizationUsers() {
	owner := c.Ctx.Input.Param(":owner")
	if owner == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	filters := object.UserSearchFilters{
		Name:        c.Ctx.Input.Query("name"),
		DisplayName: c.Ctx.Input.Query("displayName"),
		Email:       c.Ctx.Input.Query("email"),
		Phone:       c.Ctx.Input.Query("phone"),
		Affiliation: c.Ctx.Input.Query("affiliation"),
	}

	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		users, err := object.GetMaskedUsers(object.GetUsersWithSearchFilters(owner, filters, sortField, sortOrder))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(users)
		return
	}

	limitInt := util.ParseInt(limit)
	count, err := object.GetUserCountWithSearchFilters(owner, filters)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	paginator := pagination.NewPaginator(c.Ctx.Request, limitInt, count)
	users, err := object.GetPaginationUsersWithSearchFilters(owner, paginator.Offset(), limitInt, filters, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	users, err = object.GetMaskedUsers(users)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(users, paginator.Nums())
}
