// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

// GetPayments
// @Title GetPayments
// @Tag Payment API
// @Description get payments
// @Param   owner     query    string  true        "The owner of payments"
// @Success 200 {array} object.Payment The Response object
// @router /get-payments [get]
func (c *ApiController) GetPayments() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetPayments(owner)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetPaymentCount(owner, field, value)))
		payments := object.GetPaginationPayments(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(payments, paginator.Nums())
	}
}

// @Title GetPayment
// @Tag Payment API
// @Description get payment
// @Param   id    query    string  true        "The id of the payment"
// @Success 200 {object} object.Payment The Response object
// @router /get-payment [get]
func (c *ApiController) GetPayment() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetPayment(id)
	c.ServeJSON()
}

// @Title UpdatePayment
// @Tag Payment API
// @Description update payment
// @Param   id    query    string  true        "The id of the payment"
// @Param   body    body   object.Payment  true        "The details of the payment"
// @Success 200 {object} controllers.Response The Response object
// @router /update-payment [post]
func (c *ApiController) UpdatePayment() {
	id := c.Input().Get("id")

	var payment object.Payment
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &payment)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdatePayment(id, &payment))
	c.ServeJSON()
}

// @Title AddPayment
// @Tag Payment API
// @Description add payment
// @Param   body    body   object.Payment  true        "The details of the payment"
// @Success 200 {object} controllers.Response The Response object
// @router /add-payment [post]
func (c *ApiController) AddPayment() {
	var payment object.Payment
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &payment)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddPayment(&payment))
	c.ServeJSON()
}

// @Title DeletePayment
// @Tag Payment API
// @Description delete payment
// @Param   body    body   object.Payment  true        "The details of the payment"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-payment [post]
func (c *ApiController) DeletePayment() {
	var payment object.Payment
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &payment)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeletePayment(&payment))
	c.ServeJSON()
}
