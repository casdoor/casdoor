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

	"github.com/beego/beego/utils/pagination"
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

// GetUserPayments
// @Title GetUserPayments
// @Tag Payment API
// @Description get payments for a user
// @Param   owner     query    string  true        "The owner of payments"
// @Param   organization    query   string  true   "The organization of the user"
// @Param   user    query   string  true           "The username of the user"
// @Success 200 {array} object.Payment The Response object
// @router /get-user-payments [get]
func (c *ApiController) GetUserPayments() {
	owner := c.Input().Get("owner")
	organization := c.Input().Get("organization")
	user := c.Input().Get("user")

	payments := object.GetUserPayments(owner, organization, user)
	c.ResponseOk(payments)
}

// GetPayment
// @Title GetPayment
// @Tag Payment API
// @Description get payment
// @Param   id     query    string  true        "The id ( owner/name ) of the payment"
// @Success 200 {object} object.Payment The Response object
// @router /get-payment [get]
func (c *ApiController) GetPayment() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetPayment(id)
	c.ServeJSON()
}

// UpdatePayment
// @Title UpdatePayment
// @Tag Payment API
// @Description update payment
// @Param   id     query    string  true        "The id ( owner/name ) of the payment"
// @Param   body    body   object.Payment  true        "The details of the payment"
// @Success 200 {object} controllers.Response The Response object
// @router /update-payment [post]
func (c *ApiController) UpdatePayment() {
	id := c.Input().Get("id")

	var payment object.Payment
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &payment)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdatePayment(id, &payment))
	c.ServeJSON()
}

// AddPayment
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
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddPayment(&payment))
	c.ServeJSON()
}

// DeletePayment
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
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeletePayment(&payment))
	c.ServeJSON()
}

// NotifyPayment
// @Title NotifyPayment
// @Tag Payment API
// @Description notify payment
// @Param   body    body   object.Payment  true        "The details of the payment"
// @Success 200 {object} controllers.Response The Response object
// @router /notify-payment [post]
func (c *ApiController) NotifyPayment() {
	owner := c.Ctx.Input.Param(":owner")
	providerName := c.Ctx.Input.Param(":provider")
	productName := c.Ctx.Input.Param(":product")
	paymentName := c.Ctx.Input.Param(":payment")

	body := c.Ctx.Input.RequestBody

	err, errorResponse := object.NotifyPayment(c.Ctx.Request, body, owner, providerName, productName, paymentName)

	_, err2 := c.Ctx.ResponseWriter.Write([]byte(errorResponse))
	if err2 != nil {
		panic(err2)
	}

	if err != nil {
		panic(err)
	}
}

// InvoicePayment
// @Title InvoicePayment
// @Tag Payment API
// @Description invoice payment
// @Param   id     query    string  true        "The id ( owner/name ) of the payment"
// @Success 200 {object} controllers.Response The Response object
// @router /invoice-payment [post]
func (c *ApiController) InvoicePayment() {
	id := c.Input().Get("id")

	payment := object.GetPayment(id)
	invoiceUrl, err := object.InvoicePayment(payment)
	if err != nil {
		c.ResponseError(err.Error())
	}
	c.ResponseOk(invoiceUrl)
}
