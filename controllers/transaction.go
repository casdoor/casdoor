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

	"github.com/beego/beego/v2/core/utils/pagination"
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
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		var transactions []*object.Transaction
		var err error

		if c.IsAdmin() {
			// If field is "user", filter by that user even for admins
			if field == "user" && value != "" {
				transactions, err = object.GetUserTransactions(owner, value)
			} else {
				transactions, err = object.GetTransactions(owner)
			}
		} else {
			user := c.GetSessionUsername()
			_, userName, userErr := util.GetOwnerAndNameFromIdWithError(user)
			if userErr != nil {
				c.ResponseError(userErr.Error())
				return
			}
			transactions, err = object.GetUserTransactions(owner, userName)
		}

		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(transactions)
	} else {
		limit := util.ParseInt(limit)

		// Apply user filter for non-admin users
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

		count, err := object.GetTransactionCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		transactions, err := object.GetPaginationTransactions(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(transactions, paginator.Nums())
	}
}

// GetTransaction
// @Title GetTransaction
// @Tag Transaction API
// @Description get transaction
// @Param   id     query    string  true        "The id ( owner/name ) of the transaction"
// @Success 200 {object} object.Transaction The Response object
// @router /get-transaction [get]
func (c *ApiController) GetTransaction() {
	id := c.Ctx.Input.Query("id")

	transaction, err := object.GetTransaction(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if transaction == nil {
		c.ResponseOk(nil)
		return
	}

	// Check if non-admin user is trying to access someone else's transaction
	if !c.IsAdmin() {
		user := c.GetSessionUsername()
		_, userName, userErr := util.GetOwnerAndNameFromIdWithError(user)
		if userErr != nil {
			c.ResponseError(userErr.Error())
			return
		}

		// Only allow users to view their own transactions
		if transaction.User != userName {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
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
	id := c.Ctx.Input.Query("id")

	var transaction object.Transaction
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &transaction)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateTransaction(id, &transaction, c.GetAcceptLanguage()))
	c.ServeJSON()
}

// AddTransaction
// @Title AddTransaction
// @Tag Transaction API
// @Description add transaction
// @Param   body    body   object.Transaction  true        "The details of the transaction"
// @Param   dryRun  query  string  false       "Dry run mode: set to 'true' or '1' to validate without committing"
// @Success 200 {object} controllers.Response The Response object
// @router /add-transaction [post]
func (c *ApiController) AddTransaction() {
	var transaction object.Transaction
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &transaction)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	dryRunParam := c.Ctx.Input.Query("dryRun")
	dryRun := dryRunParam != ""

	affected, transactionId, err := object.AddTransaction(&transaction, c.GetAcceptLanguage(), dryRun)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Data["json"] = wrapActionResponse(false)
		c.ServeJSON()
		return
	}

	c.ResponseOk(transactionId)
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

	c.Data["json"] = wrapActionResponse(object.DeleteTransaction(&transaction, c.GetAcceptLanguage()))
	c.ServeJSON()
}
