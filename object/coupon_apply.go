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
	"math"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// CalculateDiscount computes the discount amount based on the coupon type and order amount.
// The result is always >= 0 and never exceeds orderAmount.
func CalculateDiscount(coupon *Coupon, orderAmount float64) float64 {
	if orderAmount <= 0 {
		return 0
	}

	var discount float64
	switch coupon.DiscountType {
	case "percentage":
		discount = orderAmount * coupon.Discount / 100.0
		if coupon.MaxDiscount > 0 && discount > coupon.MaxDiscount {
			discount = coupon.MaxDiscount
		}
	case "fixed":
		discount = coupon.Discount
	default:
		return 0
	}

	// Discount cannot exceed order amount
	if discount > orderAmount {
		discount = orderAmount
	}

	// Round to 2 decimal places to avoid floating point issues
	discount = math.Round(discount*100) / 100

	if discount < 0 {
		return 0
	}

	return discount
}

// ValidateCoupon validates whether a coupon code can be applied to a given order.
// Returns the Coupon if valid, or an error describing why it's not valid.
func ValidateCoupon(owner, couponCode, userName string, productNames []string, orderAmount float64, currency string) (*Coupon, error) {
	coupon, err := GetCouponByCode(owner, couponCode)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, fmt.Errorf("the coupon code: %s does not exist", couponCode)
	}

	// Check state
	if coupon.State != "Active" {
		return nil, fmt.Errorf("the coupon: %s is not active (current state: %s)", coupon.Name, coupon.State)
	}

	// Check validity period
	now := time.Now().UTC()
	if coupon.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, coupon.StartTime)
		if err != nil {
			return nil, fmt.Errorf("the coupon: %s has an invalid start time %q: %w", coupon.Name, coupon.StartTime, err)
		}
		if now.Before(startTime) {
			return nil, fmt.Errorf("the coupon: %s is not yet active (starts at %s)", coupon.Name, coupon.StartTime)
		}
	}
	if coupon.ExpireTime != "" {
		expireTime, err := time.Parse(time.RFC3339, coupon.ExpireTime)
		if err != nil {
			return nil, fmt.Errorf("the coupon: %s has an invalid expire time %q: %w", coupon.Name, coupon.ExpireTime, err)
		}
		if now.After(expireTime) {
			return nil, fmt.Errorf("the coupon: %s has expired (expired at %s)", coupon.Name, coupon.ExpireTime)
		}
	}

	// Check stock (quantity=0 means unlimited)
	if coupon.Quantity > 0 && coupon.UsedCount >= coupon.Quantity {
		return nil, fmt.Errorf("the coupon: %s has been fully redeemed (%d/%d)", coupon.Name, coupon.UsedCount, coupon.Quantity)
	}

	// Check scope
	switch coupon.Scope {
	case "product":
		if !hasOverlap(coupon.Products, productNames) {
			return nil, fmt.Errorf("the coupon: %s is not applicable to the selected products", coupon.Name)
		}
	case "user":
		if !containsString(coupon.Users, userName) {
			return nil, fmt.Errorf("the coupon: %s is not applicable to user: %s", coupon.Name, userName)
		}
	case "universal":
		// No restriction
	default:
		return nil, fmt.Errorf("the coupon: %s has an invalid scope: %s", coupon.Name, coupon.Scope)
	}

	// Check minimum order amount
	if coupon.MinOrderAmount > 0 && orderAmount < coupon.MinOrderAmount {
		return nil, fmt.Errorf("the order amount %.2f is below the minimum %.2f required for coupon: %s", orderAmount, coupon.MinOrderAmount, coupon.Name)
	}

	// Check currency (empty coupon currency means any currency)
	if coupon.Currency != "" {
		if currency == "" {
			return nil, fmt.Errorf("the coupon: %s requires currency %s, but order currency is missing", coupon.Name, coupon.Currency)
		}
		if coupon.Currency != currency {
			return nil, fmt.Errorf("the coupon: %s requires currency %s, but order currency is %s", coupon.Name, coupon.Currency, currency)
		}
	}

	// Check per-user usage limit (maxUsagePerUser=0 means unlimited)
	if coupon.MaxUsagePerUser > 0 {
		usageCount, err := GetUserCouponUsageCount(coupon.Owner, coupon.Name, userName)
		if err != nil {
			return nil, err
		}
		if usageCount >= coupon.MaxUsagePerUser {
			return nil, fmt.Errorf("the user: %s has reached the usage limit (%d/%d) for coupon: %s", userName, usageCount, coupon.MaxUsagePerUser, coupon.Name)
		}
	}

	return coupon, nil
}

// ApplyCoupon records a coupon usage and increments the used count within a transaction.
// Re-checks per-user usage limit for concurrent safety.
// Should be called after payment is confirmed.
func ApplyCoupon(couponOwner, couponName, userName, orderName string, discountAmount float64) error {
	// Use a transaction to ensure atomicity of usedCount increment + usage record insert
	session := ormer.Engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// Lock the coupon row to serialize concurrent ApplyCoupon calls and get latest state
	coupon := Coupon{Owner: couponOwner, Name: couponName}
	existed, err := session.ForUpdate().Get(&coupon)
	if err != nil {
		_ = session.Rollback()
		return err
	}
	if !existed {
		_ = session.Rollback()
		return fmt.Errorf("the coupon: %s/%s does not exist", couponOwner, couponName)
	}

	// Re-check per-user usage limit inside transaction using the same session
	if coupon.MaxUsagePerUser > 0 {
		usageCount, err := session.
			Where("coupon_owner = ? AND coupon_name = ?", couponOwner, couponName).
			And("`user` = ?", userName).
			Count(&CouponUsage{})
		if err != nil {
			_ = session.Rollback()
			return err
		}
		if int(usageCount) >= coupon.MaxUsagePerUser {
			_ = session.Rollback()
			return fmt.Errorf("the user: %s has reached the usage limit (%d/%d) for coupon: %s/%s", userName, int(usageCount), coupon.MaxUsagePerUser, couponOwner, couponName)
		}
	}

	// Increment usedCount atomically; if quantity > 0, ensure we don't exceed it
	var affected int64
	if coupon.Quantity > 0 {
		affected, err = session.ID(core.PK{couponOwner, couponName}).
			Where("used_count < quantity").
			Incr("used_count", 1).
			Update(&Coupon{})
	} else {
		affected, err = session.ID(core.PK{couponOwner, couponName}).
			Incr("used_count", 1).
			Update(&Coupon{})
	}
	if err != nil {
		_ = session.Rollback()
		return err
	}
	if affected == 0 {
		_ = session.Rollback()
		return fmt.Errorf("failed to apply coupon: %s/%s (may be fully redeemed)", couponOwner, couponName)
	}

	// Record usage
	usage := &CouponUsage{
		Owner:       couponOwner,
		CouponOwner: couponOwner,
		CouponName:  couponName,
		User:        userName,
		Order:       orderName,
		CreatedTime: util.GetCurrentTime(),
		Amount:      discountAmount,
	}
	_, err = session.Insert(usage)
	if err != nil {
		_ = session.Rollback()
		return err
	}

	return session.Commit()
}

// GetUserCouponUsageCount returns how many times a specific user has used a specific coupon.
func GetUserCouponUsageCount(couponOwner, couponName, userName string) (int, error) {
	count, err := ormer.Engine.
		Where("coupon_owner = ? AND coupon_name = ?", couponOwner, couponName).
		And("`user` = ?", userName).
		Count(&CouponUsage{})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// hasOverlap checks if two string slices share at least one common element.
func hasOverlap(a, b []string) bool {
	set := make(map[string]bool, len(a))
	for _, s := range a {
		set[s] = true
	}
	for _, s := range b {
		if set[s] {
			return true
		}
	}
	return false
}

// containsString checks if a string slice contains a specific string.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
