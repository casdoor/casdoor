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

import (
	"encoding/json"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetPlans
// @Title GetPlans
// @Tag Plan API
// @Description get plans
// @Param   owner     query    string  true        "The owner of plans"
// @Success 200 {array} object.Plan The Response object
// @router /get-plans [get]
func (c *ApiController) GetPlans() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		plans, err := object.GetPlans(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(plans)
		return
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetPlanCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		plan, err := object.GetPaginatedPlans(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(plan, paginator.Nums())
	}
}

// GetPlan
// @Title GetPlan
// @Tag Plan API
// @Description get plan
// @Param   id     query    string  true        "The id ( owner/name ) of the plan"
// @Param   includeOption     query    bool  false        "Should include plan's option"
// @Success 200 {object} object.Plan The Response object
// @router /get-plan [get]
func (c *ApiController) GetPlan() {
	id := c.Input().Get("id")
	includeOption := c.Input().Get("includeOption") == "true"

	plan, err := object.GetPlan(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if includeOption {
		options, err := object.GetPermissionsByRole(plan.Role)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		for _, option := range options {
			plan.Options = append(plan.Options, option.DisplayName)
		}

		c.ResponseOk(plan)
	} else {
		c.ResponseOk(plan)
	}
}

// UpdatePlan
// @Title UpdatePlan
// @Tag Plan API
// @Description update plan
// @Param   id     query    string  true        "The id ( owner/name ) of the plan"
// @Param   body    body   object.Plan  true        "The details of the plan"
// @Success 200 {object} controllers.Response The Response object
// @router /update-plan [post]
func (c *ApiController) UpdatePlan() {
	id := c.Input().Get("id")

	var plan object.Plan
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &plan)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := wrapActionResponse(object.UpdatePlan(id, &plan))
	c.ResponseOk(resp)
}

// AddPlan
// @Title AddPlan
// @Tag Plan API
// @Description add plan
// @Param   body    body   object.Plan  true        "The details of the plan"
// @Success 200 {object} controllers.Response The Response object
// @router /add-plan [post]
func (c *ApiController) AddPlan() {
	var plan object.Plan
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &plan)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := wrapActionResponse(object.AddPlan(&plan))
	c.ResponseOk(resp)
}

// DeletePlan
// @Title DeletePlan
// @Tag Plan API
// @Description delete plan
// @Param   body    body   object.Plan  true        "The details of the plan"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-plan [post]
func (c *ApiController) DeletePlan() {
	var plan object.Plan
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &plan)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := wrapActionResponse(object.DeletePlan(&plan))
	c.ResponseOk(resp)
}
