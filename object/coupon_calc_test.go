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
	"testing"
)

// ============================================================
// Module A: CalculateDiscount pure function tests (no database needed)
// ============================================================

func TestCalculateDiscount(t *testing.T) {
	tests := []struct {
		name        string
		coupon      *Coupon
		orderAmount float64
		want        float64
	}{
		// A1: percentage discount - normal
		{"percentage normal 20%", &Coupon{DiscountType: "percentage", Discount: 20}, 100, 20},
		// A2: percentage discount - with cap
		{"percentage with cap", &Coupon{DiscountType: "percentage", Discount: 50, MaxDiscount: 30}, 100, 30},
		// A3: percentage discount - no cap (MaxDiscount=0 means unlimited)
		{"percentage no cap", &Coupon{DiscountType: "percentage", Discount: 50, MaxDiscount: 0}, 100, 50},
		// A4: fixed discount - normal
		{"fixed normal", &Coupon{DiscountType: "fixed", Discount: 15}, 100, 15},
		// A5: fixed discount - exceeds order amount
		{"fixed exceeds order", &Coupon{DiscountType: "fixed", Discount: 150}, 100, 100},
		// A6: percentage 100% - full discount
		{"percentage 100%", &Coupon{DiscountType: "percentage", Discount: 100}, 50, 50},
		// A7: zero discount
		{"zero discount", &Coupon{DiscountType: "percentage", Discount: 0}, 100, 0},
		// A8: zero order amount
		{"zero order amount", &Coupon{DiscountType: "fixed", Discount: 10}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDiscount(tt.coupon, tt.orderAmount)
			if got != tt.want {
				t.Errorf("CalculateDiscount() = %v, want %v", got, tt.want)
			}
		})
	}
}
