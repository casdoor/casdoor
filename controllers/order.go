// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetOrders
// @Title GetOrders
// @Tag Order API
// @Description get orders
// @Param   owner     query    string  true        "The owner of orders"
// @Success 200 {array} object.Order The Response object
// @router /get-orders [get]
func (c *ApiController) GetOrders() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		var orders []*object.Order
		var err error

		if c.IsAdmin() {
			// If field is "user", filter by that user even for admins
			if field == "user" && value != "" {
				orders, err = object.GetUserOrders(owner, value)
			} else {
				orders, err = object.GetOrders(owner)
			}
		} else {
			user := c.GetSessionUsername()
			_, userName, userErr := util.GetOwnerAndNameFromIdWithError(user)
			if userErr != nil {
				c.ResponseError(userErr.Error())
				return
			}
			orders, err = object.GetUserOrders(owner, userName)
		}

		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(orders)
	} else {
		limit := util.ParseInt(limit)
		if !c.IsAdmin() {
			user := c.GetSessionUsername()
			_, userName, userErr := util.GetOwnerAndNameFromIdWithError(user)
			if userErr != nil {
				c.ResponseError(userErr.Error())
				return
			}
			field = "user"
			value = userName
		}
		count, err := object.GetOrderCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		orders, err := object.GetPaginationOrders(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(orders, paginator.Nums())
	}
}

// GetUserOrders
// @Title GetUserOrders
// @Tag Order API
// @Description get orders for a user
// @Param   owner     query    string  true        "The owner of orders"
// @Param   user    query   string  true           "The username of the user"
// @Success 200 {array} object.Order The Response object
// @router /get-user-orders [get]
func (c *ApiController) GetUserOrders() {
	owner := c.Ctx.Input.Query("owner")
	user := c.Ctx.Input.Query("user")

	orders, err := object.GetUserOrders(owner, user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(orders)
}

// GetOrder
// @Title GetOrder
// @Tag Order API
// @Description get order
// @Param   id     query    string  true        "The id ( owner/name ) of the order"
// @Success 200 {object} object.Order The Response object
// @router /get-order [get]
func (c *ApiController) GetOrder() {
	id := c.Ctx.Input.Query("id")

	order, err := object.GetOrder(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(order)
}

// UpdateOrder
// @Title UpdateOrder
// @Tag Order API
// @Description update order
// @Param   id     query    string  true        "The id ( owner/name ) of the order"
// @Param   body    body   object.Order  true        "The details of the order"
// @Success 200 {object} controllers.Response The Response object
// @router /update-order [post]
func (c *ApiController) UpdateOrder() {
	id := c.Ctx.Input.Query("id")

	var order object.Order
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &order)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateOrder(id, &order))
	c.ServeJSON()
}

// AddOrder
// @Title AddOrder
// @Tag Order API
// @Description add order
// @Param   body    body   object.Order  true        "The details of the order"
// @Success 200 {object} controllers.Response The Response object
// @router /add-order [post]
func (c *ApiController) AddOrder() {
	var order object.Order
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &order)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddOrder(&order))
	c.ServeJSON()
}

// DeleteOrder
// @Title DeleteOrder
// @Tag Order API
// @Description delete order
// @Param   body    body   object.Order  true        "The details of the order"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-order [post]
func (c *ApiController) DeleteOrder() {
	var order object.Order
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &order)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteOrder(&order))
	c.ServeJSON()
}
