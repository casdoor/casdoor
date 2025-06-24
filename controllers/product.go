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
	"fmt"
	"strconv"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetProducts
// @Title GetProducts
// @Tag Product API
// @Description get products
// @Param   owner     query    string  true        "The owner of products"
// @Success 200 {array} object.Product The Response object
// @router /get-products [get]
func (c *ApiController) GetProducts() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		products, err := object.GetProducts(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(products)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetProductCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		products, err := object.GetPaginationProducts(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(products, paginator.Nums())
	}
}

// GetProduct
// @Title GetProduct
// @Tag Product API
// @Description get product
// @Param   id     query    string  true        "The id ( owner/name ) of the product"
// @Success 200 {object} object.Product The Response object
// @router /get-product [get]
func (c *ApiController) GetProduct() {
	id := c.Input().Get("id")

	product, err := object.GetProduct(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	err = object.ExtendProductWithProviders(product)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(product)
}

// UpdateProduct
// @Title UpdateProduct
// @Tag Product API
// @Description update product
// @Param   id     query    string  true        "The id ( owner/name ) of the product"
// @Param   body    body   object.Product  true        "The details of the product"
// @Success 200 {object} controllers.Response The Response object
// @router /update-product [post]
func (c *ApiController) UpdateProduct() {
	id := c.Input().Get("id")

	var product object.Product
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &product)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateProduct(id, &product))
	c.ServeJSON()
}

// AddProduct
// @Title AddProduct
// @Tag Product API
// @Description add product
// @Param   body    body   object.Product  true        "The details of the product"
// @Success 200 {object} controllers.Response The Response object
// @router /add-product [post]
func (c *ApiController) AddProduct() {
	var product object.Product
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &product)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddProduct(&product))
	c.ServeJSON()
}

// DeleteProduct
// @Title DeleteProduct
// @Tag Product API
// @Description delete product
// @Param   body    body   object.Product  true        "The details of the product"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-product [post]
func (c *ApiController) DeleteProduct() {
	var product object.Product
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &product)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteProduct(&product))
	c.ServeJSON()
}

// BuyProduct
// @Title BuyProduct
// @Tag Product API
// @Description buy product
// @Param   id     query    string  true        "The id ( owner/name ) of the product"
// @Param   providerName    query    string  true  "The name of the provider"
// @Success 200 {object} controllers.Response The Response object
// @router /buy-product [post]
func (c *ApiController) BuyProduct() {
	id := c.Input().Get("id")
	host := c.Ctx.Request.Host
	providerName := c.Input().Get("providerName")
	paymentEnv := c.Input().Get("paymentEnv")
	customPriceStr := c.Input().Get("customPrice")
	if customPriceStr == "" {
		customPriceStr = "0"
	}

	customPrice, err := strconv.ParseFloat(customPriceStr, 64)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// buy `pricingName/planName` for `paidUserName`
	pricingName := c.Input().Get("pricingName")
	planName := c.Input().Get("planName")
	paidUserName := c.Input().Get("userName")
	owner, _ := util.GetOwnerAndNameFromId(id)
	userId := util.GetId(owner, paidUserName)
	if paidUserName == "" {
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

	payment, attachInfo, err := object.BuyProduct(id, user, providerName, pricingName, planName, host, paymentEnv, customPrice)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(payment, attachInfo)
}
