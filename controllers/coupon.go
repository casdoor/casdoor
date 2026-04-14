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

// ValidateCouponRequest is the request body for the validate-coupon API.
type ValidateCouponRequest struct {
	Owner      string   `json:"owner"`
	CouponCode string   `json:"couponCode"`
	Products   []string `json:"products"`
	Amount     float64  `json:"amount"`
	Currency   string   `json:"currency"`
}

// GetCoupons
// @Title GetCoupons
// @Tag Coupon API
// @Description get coupons
// @Param   owner     query    string  true        "The owner of coupons"
// @Success 200 {array} object.Coupon The Response object
// @router /get-coupons [get]
func (c *ApiController) GetCoupons() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		coupons, err := object.GetCoupons(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(coupons)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetCouponCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		coupons, err := object.GetPaginationCoupons(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(coupons, paginator.Nums())
	}
}

// GetCoupon
// @Title GetCoupon
// @Tag Coupon API
// @Description get coupon
// @Param   id     query    string  true        "The id ( owner/name ) of the coupon"
// @Success 200 {object} object.Coupon The Response object
// @router /get-coupon [get]
func (c *ApiController) GetCoupon() {
	id := c.Ctx.Input.Query("id")

	coupon, err := object.GetCoupon(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(coupon)
}

// UpdateCoupon
// @Title UpdateCoupon
// @Tag Coupon API
// @Description update coupon
// @Param   id     query    string  true        "The id ( owner/name ) of the coupon"
// @Param   body    body   object.Coupon  true        "The details of the coupon"
// @Success 200 {object} controllers.Response The Response object
// @router /update-coupon [post]
func (c *ApiController) UpdateCoupon() {
	id := c.Ctx.Input.Query("id")

	var coupon object.Coupon
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &coupon)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateCoupon(id, &coupon))
	c.ServeJSON()
}

// AddCoupon
// @Title AddCoupon
// @Tag Coupon API
// @Description add coupon
// @Param   body    body   object.Coupon  true        "The details of the coupon"
// @Success 200 {object} controllers.Response The Response object
// @router /add-coupon [post]
func (c *ApiController) AddCoupon() {
	var coupon object.Coupon
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &coupon)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddCoupon(&coupon))
	c.ServeJSON()
}

// DeleteCoupon
// @Title DeleteCoupon
// @Tag Coupon API
// @Description delete coupon
// @Param   body    body   object.Coupon  true        "The details of the coupon"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-coupon [post]
func (c *ApiController) DeleteCoupon() {
	var coupon object.Coupon
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &coupon)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteCoupon(&coupon))
	c.ServeJSON()
}

// ValidateCoupon
// @Title ValidateCoupon
// @Tag Coupon API
// @Description validate a coupon code for an order
// @Param   body    body   controllers.ValidateCouponRequest  true        "The coupon validation request"
// @Success 200 {object} controllers.Response The Response object
// @router /validate-coupon [post]
func (c *ApiController) ValidateCoupon() {
	var req ValidateCouponRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	_, userName, err := util.GetOwnerAndNameFromIdWithError(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	coupon, err := object.ValidateCoupon(req.Owner, req.CouponCode, userName, req.Products, req.Amount, req.Currency)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	discount := object.CalculateDiscount(coupon, req.Amount)
	c.ResponseOk(map[string]interface{}{
		"coupon": map[string]interface{}{
			"name":         coupon.Name,
			"displayName":  coupon.DisplayName,
			"discountType": coupon.DiscountType,
			"discount":     coupon.Discount,
			"maxDiscount":  coupon.MaxDiscount,
			"scope":        coupon.Scope,
		},
		"discount": discount,
	})
}
