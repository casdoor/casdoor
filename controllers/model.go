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

	"github.com/astaxie/beego/utils/pagination"
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
		c.Data["json"] = object.GetModels(owner)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetModelCount(owner, field, value)))
		models := object.GetPaginationModels(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(models, paginator.Nums())
	}
}

// @Title GetModel
// @Tag Model API
// @Description get model
// @Param   id    query    string  true        "The id of the model"
// @Success 200 {object} object.Model The Response object
// @router /get-model [get]
func (c *ApiController) GetModel() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetModel(id)
	c.ServeJSON()
}

// @Title UpdateModel
// @Tag Model API
// @Description update model
// @Param   id    query    string  true        "The id of the model"
// @Param   body    body   object.Model  true        "The details of the model"
// @Success 200 {object} controllers.Response The Response object
// @router /update-model [post]
func (c *ApiController) UpdateModel() {
	id := c.Input().Get("id")

	var model object.Model
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &model)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateModel(id, &model))
	c.ServeJSON()
}

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
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddModel(&model))
	c.ServeJSON()
}

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
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteModel(&model))
	c.ServeJSON()
}
