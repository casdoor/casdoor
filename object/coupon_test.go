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

//go:build !skipCi

package object

import (
	"fmt"
	"testing"
	"time"

	"github.com/casdoor/casdoor/util"
)

// ============================================================
// Module D: CRUD tests (require database)
// ============================================================

func newTestCoupon(owner, name, code string) *Coupon {
	return &Coupon{
		Owner:           owner,
		Name:            name,
		CreatedTime:     util.GetCurrentTime(),
		DisplayName:     fmt.Sprintf("Test Coupon %s", name),
		Description:     "Test coupon for unit testing",
		Code:            code,
		DiscountType:    "percentage",
		Discount:        20,
		MaxDiscount:     0,
		Scope:           "universal",
		Products:        []string{},
		Users:           []string{},
		Quantity:        100,
		UsedCount:       0,
		MaxUsagePerUser: 1,
		StartTime:       "2020-01-01T00:00:00Z",
		ExpireTime:      "2099-12-31T23:59:59Z",
		MinOrderAmount:  0,
		Currency:        "USD",
		State:           "Active",
	}
}

// D1: AddCoupon - create coupon and verify all fields persisted
func TestCouponCRUD_Add(t *testing.T) {
	InitConfig()

	coupon := newTestCoupon("admin", "test_coupon_add_"+util.GenerateTimeId(), "CODE_ADD_"+util.GenerateTimeId())

	t.Cleanup(func() {
		_, _ = DeleteCoupon(coupon)
	})

	affected, err := AddCoupon(coupon)
	if err != nil {
		t.Fatalf("AddCoupon() error: %v", err)
	}
	if !affected {
		t.Fatal("AddCoupon() returned false, expected true")
	}

	// Verify persisted
	got, err := GetCoupon(coupon.GetId())
	if err != nil {
		t.Fatalf("GetCoupon() error: %v", err)
	}
	if got == nil {
		t.Fatal("GetCoupon() returned nil after AddCoupon")
	}
	if got.Code != coupon.Code {
		t.Errorf("Code mismatch: got %s, want %s", got.Code, coupon.Code)
	}
	if got.DiscountType != coupon.DiscountType {
		t.Errorf("DiscountType mismatch: got %s, want %s", got.DiscountType, coupon.DiscountType)
	}
	if got.Discount != coupon.Discount {
		t.Errorf("Discount mismatch: got %v, want %v", got.Discount, coupon.Discount)
	}
	if got.Scope != coupon.Scope {
		t.Errorf("Scope mismatch: got %s, want %s", got.Scope, coupon.Scope)
	}
	if got.Quantity != coupon.Quantity {
		t.Errorf("Quantity mismatch: got %d, want %d", got.Quantity, coupon.Quantity)
	}
	if got.State != coupon.State {
		t.Errorf("State mismatch: got %s, want %s", got.State, coupon.State)
	}
}

// D2: GetCoupon - query by owner/name
func TestCouponCRUD_Get(t *testing.T) {
	InitConfig()

	coupon := newTestCoupon("admin", "test_coupon_get_"+util.GenerateTimeId(), "CODE_GET_"+util.GenerateTimeId())
	t.Cleanup(func() {
		_, _ = DeleteCoupon(coupon)
	})

	_, err := AddCoupon(coupon)
	if err != nil {
		t.Fatalf("AddCoupon() error: %v", err)
	}

	got, err := GetCoupon(coupon.GetId())
	if err != nil {
		t.Fatalf("GetCoupon() error: %v", err)
	}
	if got == nil {
		t.Fatal("GetCoupon() returned nil")
	}
	if got.Name != coupon.Name {
		t.Errorf("Name mismatch: got %s, want %s", got.Name, coupon.Name)
	}

	// Query non-existent
	got2, err := GetCoupon("admin/non_existent_coupon")
	if err != nil {
		t.Fatalf("GetCoupon() error for non-existent: %v", err)
	}
	if got2 != nil {
		t.Error("GetCoupon() should return nil for non-existent coupon")
	}
}

// D3: GetCouponByCode - query by code
func TestCouponCRUD_GetByCode(t *testing.T) {
	InitConfig()

	code := "CODE_BYCODE_" + util.GenerateTimeId()
	coupon := newTestCoupon("admin", "test_coupon_bycode_"+util.GenerateTimeId(), code)
	t.Cleanup(func() {
		_, _ = DeleteCoupon(coupon)
	})

	_, err := AddCoupon(coupon)
	if err != nil {
		t.Fatalf("AddCoupon() error: %v", err)
	}

	got, err := GetCouponByCode("admin", code)
	if err != nil {
		t.Fatalf("GetCouponByCode() error: %v", err)
	}
	if got == nil {
		t.Fatal("GetCouponByCode() returned nil")
	}
	if got.Name != coupon.Name {
		t.Errorf("Name mismatch: got %s, want %s", got.Name, coupon.Name)
	}

	// Non-existent code
	got2, err := GetCouponByCode("admin", "NON_EXISTENT_CODE")
	if err != nil {
		t.Fatalf("GetCouponByCode() error for non-existent: %v", err)
	}
	if got2 != nil {
		t.Error("GetCouponByCode() should return nil for non-existent code")
	}
}

// D4: UpdateCoupon - modify fields and verify
func TestCouponCRUD_Update(t *testing.T) {
	InitConfig()

	coupon := newTestCoupon("admin", "test_coupon_update_"+util.GenerateTimeId(), "CODE_UPDATE_"+util.GenerateTimeId())
	t.Cleanup(func() {
		_, _ = DeleteCoupon(coupon)
	})

	_, err := AddCoupon(coupon)
	if err != nil {
		t.Fatalf("AddCoupon() error: %v", err)
	}

	// Update fields
	coupon.DisplayName = "Updated Display Name"
	coupon.Discount = 30
	coupon.State = "Inactive"

	affected, err := UpdateCoupon(coupon.GetId(), coupon)
	if err != nil {
		t.Fatalf("UpdateCoupon() error: %v", err)
	}
	if !affected {
		t.Fatal("UpdateCoupon() returned false")
	}

	got, err := GetCoupon(coupon.GetId())
	if err != nil {
		t.Fatalf("GetCoupon() error: %v", err)
	}
	if got.DisplayName != "Updated Display Name" {
		t.Errorf("DisplayName not updated: got %s", got.DisplayName)
	}
	if got.Discount != 30 {
		t.Errorf("Discount not updated: got %v", got.Discount)
	}
	if got.State != "Inactive" {
		t.Errorf("State not updated: got %s", got.State)
	}
}

// D5: DeleteCoupon - delete and verify nil
func TestCouponCRUD_Delete(t *testing.T) {
	InitConfig()

	coupon := newTestCoupon("admin", "test_coupon_delete_"+util.GenerateTimeId(), "CODE_DELETE_"+util.GenerateTimeId())

	_, err := AddCoupon(coupon)
	if err != nil {
		t.Fatalf("AddCoupon() error: %v", err)
	}

	affected, err := DeleteCoupon(coupon)
	if err != nil {
		t.Fatalf("DeleteCoupon() error: %v", err)
	}
	if !affected {
		t.Fatal("DeleteCoupon() returned false")
	}

	got, err := GetCoupon(coupon.GetId())
	if err != nil {
		t.Fatalf("GetCoupon() after delete error: %v", err)
	}
	if got != nil {
		t.Error("GetCoupon() should return nil after delete")
	}
}

// D6: GetPaginationCoupons - paginated query
func TestCouponCRUD_Pagination(t *testing.T) {
	InitConfig()

	owner := "admin"
	var coupons []*Coupon
	for i := 0; i < 5; i++ {
		c := newTestCoupon(owner, fmt.Sprintf("test_coupon_page_%d_%s", i, util.GenerateTimeId()), fmt.Sprintf("CODE_PAGE_%d_%s", i, util.GenerateTimeId()))
		coupons = append(coupons, c)
	}

	t.Cleanup(func() {
		for _, c := range coupons {
			_, _ = DeleteCoupon(c)
		}
	})

	for _, c := range coupons {
		_, err := AddCoupon(c)
		if err != nil {
			t.Fatalf("AddCoupon() error: %v", err)
		}
	}

	// Get first page (limit 3)
	result, err := GetPaginationCoupons(owner, 0, 3, "", "", "", "")
	if err != nil {
		t.Fatalf("GetPaginationCoupons() error: %v", err)
	}
	if len(result) > 3 {
		t.Errorf("Expected at most 3 coupons, got %d", len(result))
	}

	// Get count
	count, err := GetCouponCount(owner, "", "")
	if err != nil {
		t.Fatalf("GetCouponCount() error: %v", err)
	}
	if count < 5 {
		t.Errorf("Expected at least 5 coupons in count, got %d", count)
	}
}

// D7: Code uniqueness - duplicate code should fail
func TestCouponCRUD_CodeUniqueness(t *testing.T) {
	InitConfig()

	sharedCode := "CODE_UNIQUE_" + util.GenerateTimeId()
	coupon1 := newTestCoupon("admin", "test_coupon_unique1_"+util.GenerateTimeId(), sharedCode)
	coupon2 := newTestCoupon("admin", "test_coupon_unique2_"+util.GenerateTimeId(), sharedCode)

	t.Cleanup(func() {
		_, _ = DeleteCoupon(coupon1)
		_, _ = DeleteCoupon(coupon2)
	})

	_, err := AddCoupon(coupon1)
	if err != nil {
		t.Fatalf("AddCoupon(coupon1) error: %v", err)
	}

	_, err = AddCoupon(coupon2)
	if err == nil {
		t.Error("AddCoupon(coupon2) should fail with duplicate code, but got no error")
	}
}

// ============================================================
// Module B: ValidateCoupon tests (require database)
// ============================================================

func TestValidateCoupon(t *testing.T) {
	InitConfig()

	now := time.Now()
	pastTime := now.Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	futureTime := now.Add(24 * time.Hour).UTC().Format(time.RFC3339)
	farFutureTime := now.Add(365 * 24 * time.Hour).UTC().Format(time.RFC3339)

	// Helper to create and add a coupon for test, returns cleanup func
	addTestCoupon := func(t *testing.T, coupon *Coupon) {
		t.Helper()
		_, err := AddCoupon(coupon)
		if err != nil {
			t.Fatalf("setup: AddCoupon() error: %v", err)
		}
		t.Cleanup(func() {
			_, _ = DeleteCoupon(coupon)
		})
	}

	// B1: Valid universal coupon
	t.Run("B1 valid universal coupon", func(t *testing.T) {
		code := "B1_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b1_"+util.GenerateTimeId(), code)
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		got, err := ValidateCoupon("admin", code, "alice", []string{"any_product"}, 100, "USD")
		if err != nil {
			t.Fatalf("ValidateCoupon() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("ValidateCoupon() returned nil, expected coupon")
		}
		if got.Code != code {
			t.Errorf("Code mismatch: got %s, want %s", got.Code, code)
		}
	})

	// B2: Expired coupon
	t.Run("B2 expired coupon", func(t *testing.T) {
		code := "B2_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b2_"+util.GenerateTimeId(), code)
		coupon.StartTime = "2020-01-01T00:00:00Z"
		coupon.ExpireTime = pastTime
		addTestCoupon(t, coupon)

		_, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error for expired coupon")
		}
	})

	// B3: Not yet active coupon
	t.Run("B3 not yet active coupon", func(t *testing.T) {
		code := "B3_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b3_"+util.GenerateTimeId(), code)
		coupon.StartTime = futureTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		_, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error for not-yet-active coupon")
		}
	})

	// B4: Inactive coupon
	t.Run("B4 inactive coupon", func(t *testing.T) {
		code := "B4_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b4_"+util.GenerateTimeId(), code)
		coupon.State = "Inactive"
		addTestCoupon(t, coupon)

		_, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error for inactive coupon")
		}
	})

	// B5: Fully redeemed coupon
	t.Run("B5 fully redeemed", func(t *testing.T) {
		code := "B5_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b5_"+util.GenerateTimeId(), code)
		coupon.Quantity = 10
		coupon.UsedCount = 10
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		_, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error for fully redeemed coupon")
		}
	})

	// B6: Unlimited quantity (quantity=0)
	t.Run("B6 unlimited quantity", func(t *testing.T) {
		code := "B6_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b6_"+util.GenerateTimeId(), code)
		coupon.Quantity = 0
		coupon.UsedCount = 9999
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		got, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err != nil {
			t.Fatalf("ValidateCoupon() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("ValidateCoupon() returned nil for unlimited coupon")
		}
	})

	// B7: Product scope - matching
	t.Run("B7 product scope matching", func(t *testing.T) {
		code := "B7_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b7_"+util.GenerateTimeId(), code)
		coupon.Scope = "product"
		coupon.Products = []string{"prod_A", "prod_B"}
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		got, err := ValidateCoupon("admin", code, "alice", []string{"prod_A"}, 100, "USD")
		if err != nil {
			t.Fatalf("ValidateCoupon() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("ValidateCoupon() returned nil for matching product")
		}
	})

	// B8: Product scope - not matching
	t.Run("B8 product scope not matching", func(t *testing.T) {
		code := "B8_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b8_"+util.GenerateTimeId(), code)
		coupon.Scope = "product"
		coupon.Products = []string{"prod_A"}
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		_, err := ValidateCoupon("admin", code, "alice", []string{"prod_B"}, 100, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error for non-matching product")
		}
	})

	// B9: User scope - matching
	t.Run("B9 user scope matching", func(t *testing.T) {
		code := "B9_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b9_"+util.GenerateTimeId(), code)
		coupon.Scope = "user"
		coupon.Users = []string{"alice", "bob"}
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		got, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err != nil {
			t.Fatalf("ValidateCoupon() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("ValidateCoupon() returned nil for matching user")
		}
	})

	// B10: User scope - not matching
	t.Run("B10 user scope not matching", func(t *testing.T) {
		code := "B10_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b10_"+util.GenerateTimeId(), code)
		coupon.Scope = "user"
		coupon.Users = []string{"alice"}
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		_, err := ValidateCoupon("admin", code, "bob", []string{"any"}, 100, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error for non-matching user")
		}
	})

	// B11: Min order amount not met
	t.Run("B11 min order amount not met", func(t *testing.T) {
		code := "B11_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b11_"+util.GenerateTimeId(), code)
		coupon.MinOrderAmount = 100
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		_, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 50, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error when order amount below minimum")
		}
	})

	// B12: Min order amount exactly met
	t.Run("B12 min order amount exactly met", func(t *testing.T) {
		code := "B12_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b12_"+util.GenerateTimeId(), code)
		coupon.MinOrderAmount = 100
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		got, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err != nil {
			t.Fatalf("ValidateCoupon() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("ValidateCoupon() returned nil when min order met exactly")
		}
	})

	// B13: Currency mismatch
	t.Run("B13 currency mismatch", func(t *testing.T) {
		code := "B13_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b13_"+util.GenerateTimeId(), code)
		coupon.Currency = "USD"
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		_, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "CNY")
		if err == nil {
			t.Error("ValidateCoupon() should return error for currency mismatch")
		}
	})

	// B14: User usage limit exceeded (requires CouponUsage record)
	t.Run("B14 user usage limit exceeded", func(t *testing.T) {
		code := "B14_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b14_"+util.GenerateTimeId(), code)
		coupon.MaxUsagePerUser = 1
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		// Record one usage for this user
		usage := &CouponUsage{
			Owner:       "admin",
			CouponOwner: coupon.Owner,
			CouponName:  coupon.Name,
			User:        "alice",
			Order:       "order_test",
			CreatedTime: util.GetCurrentTime(),
			Amount:      20,
		}
		_, err := ormer.Engine.Insert(usage)
		if err != nil {
			t.Fatalf("Insert CouponUsage error: %v", err)
		}
		t.Cleanup(func() {
			_, _ = ormer.Engine.Where("coupon_owner = ? AND coupon_name = ?", coupon.Owner, coupon.Name).Delete(&CouponUsage{})
		})

		_, err = ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error when user usage limit exceeded")
		}
	})

	// B15: Unlimited per-user usage (maxUsagePerUser=0)
	t.Run("B15 unlimited per-user usage", func(t *testing.T) {
		code := "B15_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b15_"+util.GenerateTimeId(), code)
		coupon.MaxUsagePerUser = 0
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		// Add several usages
		for i := 0; i < 5; i++ {
			usage := &CouponUsage{
				Owner:       "admin",
				CouponOwner: coupon.Owner,
				CouponName:  coupon.Name,
				User:        "alice",
				Order:       fmt.Sprintf("order_%d", i),
				CreatedTime: util.GetCurrentTime(),
				Amount:      20,
			}
			_, _ = ormer.Engine.Insert(usage)
		}
		t.Cleanup(func() {
			_, _ = ormer.Engine.Where("coupon_owner = ? AND coupon_name = ?", coupon.Owner, coupon.Name).Delete(&CouponUsage{})
		})

		got, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "USD")
		if err != nil {
			t.Fatalf("ValidateCoupon() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("ValidateCoupon() returned nil for unlimited per-user usage")
		}
	})

	// B16: Non-existent coupon code
	t.Run("B16 non-existent code", func(t *testing.T) {
		_, err := ValidateCoupon("admin", "NON_EXISTENT_CODE_"+util.GenerateTimeId(), "alice", []string{"any"}, 100, "USD")
		if err == nil {
			t.Error("ValidateCoupon() should return error for non-existent code")
		}
	})

	// B17: Product scope - partial match (at least one product matches)
	t.Run("B17 product scope partial match", func(t *testing.T) {
		code := "B17_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b17_"+util.GenerateTimeId(), code)
		coupon.Scope = "product"
		coupon.Products = []string{"prod_A", "prod_B"}
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		got, err := ValidateCoupon("admin", code, "alice", []string{"prod_A", "prod_C"}, 100, "USD")
		if err != nil {
			t.Fatalf("ValidateCoupon() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("ValidateCoupon() should allow partial product match")
		}
	})

	// B18: Empty currency - matches any
	t.Run("B18 empty currency matches any", func(t *testing.T) {
		code := "B18_" + util.GenerateTimeId()
		coupon := newTestCoupon("admin", "test_b18_"+util.GenerateTimeId(), code)
		coupon.Currency = ""
		coupon.StartTime = pastTime
		coupon.ExpireTime = farFutureTime
		addTestCoupon(t, coupon)

		got, err := ValidateCoupon("admin", code, "alice", []string{"any"}, 100, "CNY")
		if err != nil {
			t.Fatalf("ValidateCoupon() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("ValidateCoupon() should pass for empty currency coupon")
		}
	})
}

// ============================================================
// Module C: ApplyCoupon tests (require database)
// ============================================================

// C1: Normal apply - first usage
func TestApplyCoupon_Normal(t *testing.T) {
	InitConfig()

	coupon := newTestCoupon("admin", "test_apply_normal_"+util.GenerateTimeId(), "CODE_APPLY_"+util.GenerateTimeId())
	coupon.StartTime = "2020-01-01T00:00:00Z"
	coupon.ExpireTime = "2099-12-31T23:59:59Z"
	coupon.UsedCount = 0
	coupon.Quantity = 10

	t.Cleanup(func() {
		_, _ = DeleteCoupon(coupon)
		_, _ = ormer.Engine.Where("coupon_owner = ? AND coupon_name = ?", coupon.Owner, coupon.Name).Delete(&CouponUsage{})
	})

	_, err := AddCoupon(coupon)
	if err != nil {
		t.Fatalf("AddCoupon() error: %v", err)
	}

	err = ApplyCoupon(coupon.Owner, coupon.Name, "alice", "order_test_1", 30.0)
	if err != nil {
		t.Fatalf("ApplyCoupon() error: %v", err)
	}

	// Verify usedCount incremented
	updated, err := GetCoupon(coupon.GetId())
	if err != nil {
		t.Fatalf("GetCoupon() error: %v", err)
	}
	if updated.UsedCount != 1 {
		t.Errorf("UsedCount: got %d, want 1", updated.UsedCount)
	}

	// Verify CouponUsage record
	count, err := GetUserCouponUsageCount(coupon.Owner, coupon.Name, "alice")
	if err != nil {
		t.Fatalf("GetUserCouponUsageCount() error: %v", err)
	}
	if count != 1 {
		t.Errorf("Usage count: got %d, want 1", count)
	}
}

// C3: Amount recorded correctly
func TestApplyCoupon_Amount(t *testing.T) {
	InitConfig()

	coupon := newTestCoupon("admin", "test_apply_amount_"+util.GenerateTimeId(), "CODE_AMT_"+util.GenerateTimeId())
	coupon.StartTime = "2020-01-01T00:00:00Z"
	coupon.ExpireTime = "2099-12-31T23:59:59Z"

	t.Cleanup(func() {
		_, _ = DeleteCoupon(coupon)
		_, _ = ormer.Engine.Where("coupon_owner = ? AND coupon_name = ?", coupon.Owner, coupon.Name).Delete(&CouponUsage{})
	})

	_, err := AddCoupon(coupon)
	if err != nil {
		t.Fatalf("AddCoupon() error: %v", err)
	}

	err = ApplyCoupon(coupon.Owner, coupon.Name, "bob", "order_test_amt", 42.5)
	if err != nil {
		t.Fatalf("ApplyCoupon() error: %v", err)
	}

	// Verify amount in CouponUsage
	var usages []CouponUsage
	err = ormer.Engine.Where("coupon_owner = ? AND coupon_name = ?", coupon.Owner, coupon.Name).And("`user` = ?", "bob").Find(&usages)
	if err != nil {
		t.Fatalf("Query CouponUsage error: %v", err)
	}
	if len(usages) != 1 {
		t.Fatalf("Expected 1 usage record, got %d", len(usages))
	}
	if usages[0].Amount != 42.5 {
		t.Errorf("Amount: got %v, want 42.5", usages[0].Amount)
	}
}
