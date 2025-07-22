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

package controllers

import (
	"encoding/json"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetModels
// @Title GetModels
// @Tag Model API
// @Description get models
// @Param   owner     query    string  true        "The owner of models"
// @Success 200 {array} object.Model The Response object
// @router /get-models [get]
func (c *ApiController) GetModels() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		models, err := object.GetModels(owner)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		c.ResponseOk(models)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetModelCount(owner, field, value)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		models, err := object.GetPaginationModels(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		c.ResponseOk(models, paginator.Nums())
	}
}

// GetModel
// @Title GetModel
// @Tag Model API
// @Description get model
// @Param   id     query    string  true        "The id ( owner/name ) of the model"
// @Success 200 {object} object.Model The Response object
// @router /get-model [get]
func (c *ApiController) GetModel() {
	id := c.Input().Get("id")

	model, err := object.GetModel(id)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.ResponseOk(model)
}

// UpdateModel
// @Title UpdateModel
// @Tag Model API
// @Description update model
// @Param   id     query    string  true        "The id ( owner/name ) of the model"
// @Param   body    body   object.Model  true        "The details of the model"
// @Success 200 {object} controllers.Response The Response object
// @router /update-model [post]
func (c *ApiController) UpdateModel() {
	id := c.Input().Get("id")

	var model object.Model
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &model)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapErrorResponse(object.UpdateModelWithCheck(id, &model))
	c.ServeJSON()
}

// AddModel
// @Title AddModel
// @Tag Model API
// @Description add model
// @Param   body    body   object.Model  true        "The details of the model"
// @Success 200 {object} controllers.Response The Response object
// @router /add-model [post]
func (c *ApiController) AddModel() {
	var model object.Model
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &model)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddModel(&model))
	c.ServeJSON()
}

// DeleteModel
// @Title DeleteModel
// @Tag Model API
// @Description delete model
// @Param   body    body   object.Model  true        "The details of the model"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-model [post]
func (c *ApiController) DeleteModel() {
	var model object.Model
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &model)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteModel(&model))
	c.ServeJSON()
}
