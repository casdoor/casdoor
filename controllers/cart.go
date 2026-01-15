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
	"encoding/json"

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetCarts
// @Title GetCarts
// @Tag Cart API
// @Description get carts
// @Param   owner     query    string  true        "The owner of carts"
// @Success 200 {array} object.Cart The Response object
// @router /get-carts [get]
func (c *ApiController) GetCarts() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		var carts []*object.Cart
		var err error

		if c.IsAdmin() {
			if field == "user" && value != "" {
				carts, err = object.GetUserCarts(owner, value)
			} else {
				carts, err = object.GetCarts(owner)
			}
		} else {
			user := c.GetSessionUsername()
			_, userName, userErr := util.GetOwnerAndNameFromIdWithError(user)
			if userErr != nil {
				c.ResponseError(userErr.Error())
				return
			}
			carts, err = object.GetUserCarts(owner, userName)
		}

		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		for _, cart := range carts {
			err = object.ExtendCartWithProduct(cart)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}

		c.ResponseOk(carts)
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

		count, err := object.GetCartCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		carts, err := object.GetPaginationCarts(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		for _, cart := range carts {
			err = object.ExtendCartWithProduct(cart)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}

		c.ResponseOk(carts, paginator.Nums())
	}
}

// GetCart
// @Title GetCart
// @Tag Cart API
// @Description get cart
// @Param   id     query    string  true        "The id ( owner/name ) of the cart"
// @Success 200 {object} object.Cart The Response object
// @router /get-cart [get]
func (c *ApiController) GetCart() {
	id := c.Ctx.Input.Query("id")

	cart, err := object.GetCart(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if cart == nil {
		c.ResponseOk(nil)
		return
	}

	if !c.IsAdmin() {
		user := c.GetSessionUsername()
		_, userName, userErr := util.GetOwnerAndNameFromIdWithError(user)
		if userErr != nil {
			c.ResponseError(userErr.Error())
			return
		}

		if cart.User != userName {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
	}

	err = object.ExtendCartWithProduct(cart)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(cart)
}

// UpdateCart
// @Title UpdateCart
// @Tag Cart API
// @Description update cart
// @Param   id     query    string  true        "The id ( owner/name ) of the cart"
// @Param   body    body   object.Cart  true        "The details of the cart"
// @Success 200 {object} controllers.Response The Response object
// @router /update-cart [post]
func (c *ApiController) UpdateCart() {
	id := c.Ctx.Input.Query("id")

	var cart object.Cart
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cart)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateCart(id, &cart))
	c.ServeJSON()
}

// AddCart
// @Title AddCart
// @Tag Cart API
// @Description add cart
// @Param   body    body   object.Cart  true        "The details of the cart"
// @Success 200 {object} controllers.Response The Response object
// @router /add-cart [post]
func (c *ApiController) AddCart() {
	var cart object.Cart
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cart)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddCart(&cart))
	c.ServeJSON()
}

// DeleteCart
// @Title DeleteCart
// @Tag Cart API
// @Description delete cart
// @Param   body    body   object.Cart  true        "The details of the cart"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-cart [post]
func (c *ApiController) DeleteCart() {
	var cart object.Cart
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cart)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteCart(&cart))
	c.ServeJSON()
}
