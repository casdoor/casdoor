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
	"fmt"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// PlaceOrder
// @Title PlaceOrder
// @Tag Order API
// @Description place an order for a product
// @Param   productId     query    string  true        "The id ( owner/name ) of the product"
// @Param   pricingName   query    string  false       "The name of the pricing (for subscription)"
// @Param   planName      query    string  false       "The name of the plan (for subscription)"
// @Param   customPrice   query    number  false       "Custom price for recharge products"
// @Param   userName      query    string  false       "The username to place order for (admin only)"
// @Success 200 {object} object.Order The Response object
// @router /place-order [post]
func (c *ApiController) PlaceOrder() {
	owner := c.Ctx.Input.Query("owner")
	pricingName := c.Ctx.Input.Query("pricingName")
	planName := c.Ctx.Input.Query("planName")
	paidUserName := c.Ctx.Input.Query("userName")

	var req struct {
		ProductInfos []object.ProductInfo `json:"productInfos"`
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	productInfos := req.ProductInfos
	if len(productInfos) == 0 {
		c.ResponseError(c.T("product:Product list cannot be empty"))
		return
	}

	var userId string
	if paidUserName != "" {
		userId = util.GetId(owner, paidUserName)
		if userId != c.GetSessionUsername() && !c.IsAdmin() && userId != c.GetPaidUsername() {
			c.ResponseError(c.T("general:Only admin user can specify user"))
			return
		}

		c.SetSession("paidUsername", "")
	} else {
		userId = c.GetSessionUsername()
	}

	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), userId))
		return
	}

	order, err := object.PlaceOrder(owner, productInfos, user, pricingName, planName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(order)
}

// PayOrder
// @Title PayOrder
// @Tag Order API
// @Description pay an existing order
// @Param   id     query    string  true        "The id ( owner/name ) of the order"
// @Param   providerName    query    string  true  "The name of the provider"
// @Success 200 {object} controllers.Response The Response object
// @router /pay-order [post]
func (c *ApiController) PayOrder() {
	id := c.Ctx.Input.Query("id")
	host := c.Ctx.Request.Host
	providerName := c.Ctx.Input.Query("providerName")
	paymentEnv := c.Ctx.Input.Query("paymentEnv")

	order, err := object.GetOrder(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if order == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The order: %s does not exist"), id))
		return
	}

	userId := c.GetSessionUsername()
	orderUserId := util.GetId(order.Owner, order.User)
	if userId != orderUserId && !c.IsAdmin() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	payment, attachInfo, err := object.PayOrder(providerName, host, paymentEnv, order, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(payment, attachInfo)
}

// CancelOrder
// @Title CancelOrder
// @Tag Order API
// @Description cancel an order
// @Param   id     query    string  true        "The id ( owner/name ) of the order"
// @Success 200 {object} controllers.Response The Response object
// @router /cancel-order [post]
func (c *ApiController) CancelOrder() {
	id := c.Ctx.Input.Query("id")
	order, err := object.GetOrder(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if order == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The order: %s does not exist"), id))
		return
	}

	userId := c.GetSessionUsername()
	orderUserId := util.GetId(order.Owner, order.User)
	if userId != orderUserId && !c.IsAdmin() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	c.Data["json"] = wrapActionResponse(object.CancelOrder(order))
	c.ServeJSON()
}
