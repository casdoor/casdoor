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
	"fmt"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
)

// GetEnforcers
// @Title GetEnforcers
// @Tag Enforcer API
// @Description get enforcers
// @Param   owner     query    string  true        "The owner of enforcers"
// @Success 200 {array} object.Enforcer
// @router /get-enforcers [get]
func (c *ApiController) GetEnforcers() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		enforcers, err := object.GetEnforcers(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(enforcers)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetEnforcerCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		enforcers, err := object.GetPaginationEnforcers(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(enforcers, paginator.Nums())
	}
}

// GetEnforcer
// @Title GetEnforcer
// @Tag Enforcer API
// @Description get enforcer
// @Param   id     query    string  true        "The id ( owner/name )  of enforcer"
// @Success 200 {object} object.Enforcer
// @router /get-enforcer [get]
func (c *ApiController) GetEnforcer() {
	id := c.Input().Get("id")
	loadModelCfg := c.Input().Get("loadModelCfg")

	enforcer, err := object.GetEnforcer(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if loadModelCfg == "true" && enforcer.Model != "" {
		err := enforcer.LoadModelCfg()
		if err != nil {
			return
		}
	}

	c.ResponseOk(enforcer)
}

// UpdateEnforcer
// @Title UpdateEnforcer
// @Tag Enforcer API
// @Description update enforcer
// @Param   id     query    string  true        "The id ( owner/name )  of enforcer"
// @Param   enforcer     body    object  true        "The enforcer object"
// @Success 200 {object} object.Enforcer
// @router /update-enforcer [post]
func (c *ApiController) UpdateEnforcer() {
	id := c.Input().Get("id")

	enforcer := object.Enforcer{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &enforcer)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateEnforcer(id, &enforcer))
	c.ServeJSON()
}

// AddEnforcer
// @Title AddEnforcer
// @Tag Enforcer API
// @Description add enforcer
// @Param   enforcer     body    object  true        "The enforcer object"
// @Success 200 {object} object.Enforcer
// @router /add-enforcer [post]
func (c *ApiController) AddEnforcer() {
	enforcer := object.Enforcer{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &enforcer)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddEnforcer(&enforcer))
	c.ServeJSON()
}

// DeleteEnforcer
// @Title DeleteEnforcer
// @Tag Enforcer API
// @Description delete enforcer
// @Param   body    body    object.Enforcer  true      "The enforcer object"
// @Success 200 {object} object.Enforcer
// @router /delete-enforcer [post]
func (c *ApiController) DeleteEnforcer() {
	var enforcer object.Enforcer
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &enforcer)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteEnforcer(&enforcer))
	c.ServeJSON()
}

func (c *ApiController) GetPolicies() {
	id := c.Input().Get("id")
	adapterId := c.Input().Get("adapterId")

	if adapterId != "" {
		adapter, err := object.GetAdapter(adapterId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if adapter == nil {
			c.ResponseError(fmt.Sprintf(c.T("the adapter: %s is not found"), adapterId))
			return
		}

		err = adapter.InitAdapter()
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk()
		return
	}

	policies, err := object.GetPolicies(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(policies)
}

func (c *ApiController) UpdatePolicy() {
	id := c.Input().Get("id")

	var policies []xormadapter.CasbinRule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &policies)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.UpdatePolicy(id, policies[0].Ptype, util.CasbinToSlice(policies[0]), util.CasbinToSlice(policies[1]))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

func (c *ApiController) AddPolicy() {
	id := c.Input().Get("id")

	var policy xormadapter.CasbinRule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &policy)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.AddPolicy(id, policy.Ptype, util.CasbinToSlice(policy))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

func (c *ApiController) RemovePolicy() {
	id := c.Input().Get("id")

	var policy xormadapter.CasbinRule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &policy)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.RemovePolicy(id, policy.Ptype, util.CasbinToSlice(policy))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}
