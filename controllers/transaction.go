// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

// GetTransactions
// @Title GetTransactions
// @Tag Transaction API
// @Description get transactions
// @Param   owner     query    string  true        "The owner of transactions"
// @Success 200 {array} object.Transaction The Response object
// @router /get-transactions [get]
func (c *ApiController) GetTransactions() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		transactions, err := object.GetTransactions(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(transactions)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetTransactionCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		transactions, err := object.GetPaginationTransactions(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(transactions, paginator.Nums())
	}
}

// GetUserTransactions
// @Title GetUserTransaction
// @Tag Transaction API
// @Description get transactions for a user
// @Param   owner     query    string  true        "The owner of transactions"
// @Param   organization    query   string  true   "The organization of the user"
// @Param   user    query   string  true           "The username of the user"
// @Success 200 {array} object.Transaction The Response object
// @router /get-user-transactions [get]
func (c *ApiController) GetUserTransactions() {
	owner := c.Input().Get("owner")
	user := c.Input().Get("user")

	transactions, err := object.GetUserTransactions(owner, user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(transactions)
}

// GetTransaction
// @Title GetTransaction
// @Tag Transaction API
// @Description get transaction
// @Param   id     query    string  true        "The id ( owner/name ) of the transaction"
// @Success 200 {object} object.Transaction The Response object
// @router /get-transaction [get]
func (c *ApiController) GetTransaction() {
	id := c.Input().Get("id")

	transaction, err := object.GetTransaction(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(transaction)
}

// UpdateTransaction
// @Title UpdateTransaction
// @Tag Transaction API
// @Description update transaction
// @Param   id     query    string  true        "The id ( owner/name ) of the transaction"
// @Param   body    body   object.Transaction  true        "The details of the transaction"
// @Success 200 {object} controllers.Response The Response object
// @router /update-transaction [post]
func (c *ApiController) UpdateTransaction() {
	id := c.Input().Get("id")

	var transaction object.Transaction
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &transaction)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateTransaction(id, &transaction))
	c.ServeJSON()
}

// AddTransaction
// @Title AddTransaction
// @Tag Transaction API
// @Description add transaction
// @Param   body    body   object.Transaction  true        "The details of the transaction"
// @Success 200 {object} controllers.Response The Response object
// @router /add-transaction [post]
func (c *ApiController) AddTransaction() {
	var transaction object.Transaction
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &transaction)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddTransaction(&transaction))
	c.ServeJSON()
}

// DeleteTransaction
// @Title DeleteTransaction
// @Tag Transaction API
// @Description delete transaction
// @Param   body    body   object.Transaction  true        "The details of the transaction"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-transaction [post]
func (c *ApiController) DeleteTransaction() {
	var transaction object.Transaction
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &transaction)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteTransaction(&transaction))
	c.ServeJSON()
}
