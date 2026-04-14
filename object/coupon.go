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

package object

import (
	"fmt"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Coupon struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(500)" json:"description"`

	// Coupon code (the redemption code users enter)
	Code string `xorm:"varchar(100) unique" json:"code"`

	// Discount type: "percentage" or "fixed"
	DiscountType string  `xorm:"varchar(20)" json:"discountType"`
	Discount     float64 `json:"discount"`    // Discount value: percentage (0-100) or fixed amount
	MaxDiscount  float64 `json:"maxDiscount"` // Max discount for percentage type, 0 = unlimited

	// Scope: "universal" / "product" / "user"
	Scope    string   `xorm:"varchar(20)" json:"scope"`
	Products []string `xorm:"varchar(2000)" json:"products"` // Applicable when scope="product"
	Users    []string `xorm:"varchar(2000)" json:"users"`    // Applicable when scope="user"

	// Usage limits
	Quantity        int `json:"quantity"`        // Total issued count, 0 = unlimited
	UsedCount       int `json:"usedCount"`       // Number of times used
	MaxUsagePerUser int `json:"maxUsagePerUser"` // Max usage per user, 0 = unlimited

	// Validity period
	StartTime  string `xorm:"varchar(100)" json:"startTime"`
	ExpireTime string `xorm:"varchar(100)" json:"expireTime"`

	// Minimum order amount
	MinOrderAmount float64 `json:"minOrderAmount"`
	Currency       string  `xorm:"varchar(100)" json:"currency"`

	// State: "Active" / "Inactive" / "Expired"
	State string `xorm:"varchar(20)" json:"state"`
}

type CouponUsage struct {
	Id          int     `xorm:"int notnull pk autoincr" json:"id"`
	Owner       string  `xorm:"varchar(100)" json:"owner"`
	CouponOwner string  `xorm:"varchar(100)" json:"couponOwner"`
	CouponName  string  `xorm:"varchar(100)" json:"couponName"`
	User        string  `xorm:"varchar(100)" json:"user"`
	Order       string  `xorm:"varchar(100)" json:"order"`
	CreatedTime string  `xorm:"varchar(100)" json:"createdTime"`
	Amount      float64 `json:"amount"` // Actual discount amount applied
}

func (coupon *Coupon) GetId() string {
	return fmt.Sprintf("%s/%s", coupon.Owner, coupon.Name)
}

func GetCouponCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Coupon{Owner: owner})
}

func GetCoupons(owner string) ([]*Coupon, error) {
	coupons := []*Coupon{}
	err := ormer.Engine.Desc("created_time").Find(&coupons, &Coupon{Owner: owner})
	if err != nil {
		return nil, err
	}
	return coupons, nil
}

func GetPaginationCoupons(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Coupon, error) {
	coupons := []*Coupon{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&coupons, &Coupon{Owner: owner})
	if err != nil {
		return nil, err
	}
	return coupons, nil
}

func getCoupon(owner string, name string) (*Coupon, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	coupon := Coupon{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&coupon)
	if err != nil {
		return nil, err
	}

	if existed {
		return &coupon, nil
	} else {
		return nil, nil
	}
}

func GetCoupon(id string) (*Coupon, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getCoupon(owner, name)
}

func GetCouponByCode(owner, code string) (*Coupon, error) {
	if code == "" {
		return nil, nil
	}

	coupon := Coupon{Owner: owner, Code: code}
	existed, err := ormer.Engine.Get(&coupon)
	if err != nil {
		return nil, err
	}

	if existed {
		return &coupon, nil
	} else {
		return nil, nil
	}
}

func UpdateCoupon(id string, coupon *Coupon) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}

	if c, err := getCoupon(owner, name); err != nil {
		return false, err
	} else if c == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(coupon)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddCoupon(coupon *Coupon) (bool, error) {
	affected, err := ormer.Engine.Insert(coupon)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteCoupon(coupon *Coupon) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{coupon.Owner, coupon.Name}).Delete(&Coupon{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
