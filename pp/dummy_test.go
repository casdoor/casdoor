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

package pp

import "testing"

func TestDummyPaymentProvider_PayAndNotify(t *testing.T) {
	provider, err := NewDummyPaymentProvider()
	if err != nil {
		t.Fatalf("Failed to create dummy provider: %v", err)
	}

	// Test the Pay method
	payReq := &PayReq{
		ProviderName:       "test-provider",
		ProductName:        "test-product",
		PayerName:          "Test User",
		PayerId:            "user-123",
		PayerEmail:         "test@example.com",
		PaymentName:        "payment_test_123",
		ProductDisplayName: "Product App1, Product - Casbin Software, Product - Recharge",
		ProductDescription: "Description, , ",
		Price:              340.0,
		Currency:           "USD",
		ReturnUrl:          "https://example.com/return",
		NotifyUrl:          "https://example.com/notify",
	}

	payResp, err := provider.Pay(payReq)
	if err != nil {
		t.Fatalf("Pay method failed: %v", err)
	}

	if payResp.PayUrl != payReq.ReturnUrl {
		t.Errorf("Expected PayUrl to be %s, got %s", payReq.ReturnUrl, payResp.PayUrl)
	}

	if payResp.OrderId == "" {
		t.Error("Expected OrderId to be set, got empty string")
	}

	// Test the Notify method
	notifyResult, err := provider.Notify([]byte{}, payResp.OrderId)
	if err != nil {
		t.Fatalf("Notify method failed: %v", err)
	}

	if notifyResult.PaymentStatus != PaymentStatePaid {
		t.Errorf("Expected PaymentStatus to be %s, got %s", PaymentStatePaid, notifyResult.PaymentStatus)
	}

	if notifyResult.Price != payReq.Price {
		t.Errorf("Expected Price to be %f, got %f", payReq.Price, notifyResult.Price)
	}

	if notifyResult.Currency != payReq.Currency {
		t.Errorf("Expected Currency to be %s, got %s", payReq.Currency, notifyResult.Currency)
	}

	if notifyResult.ProductDisplayName != payReq.ProductDisplayName {
		t.Errorf("Expected ProductDisplayName to be %s, got %s", payReq.ProductDisplayName, notifyResult.ProductDisplayName)
	}
}

func TestDummyPaymentProvider_NotifyWithEmptyOrderId(t *testing.T) {
	provider, err := NewDummyPaymentProvider()
	if err != nil {
		t.Fatalf("Failed to create dummy provider: %v", err)
	}

	// Test Notify with empty orderId (should not fail, but return zero values)
	notifyResult, err := provider.Notify([]byte{}, "")
	if err != nil {
		t.Fatalf("Notify method with empty orderId failed: %v", err)
	}

	if notifyResult.PaymentStatus != PaymentStatePaid {
		t.Errorf("Expected PaymentStatus to be %s, got %s", PaymentStatePaid, notifyResult.PaymentStatus)
	}

	if notifyResult.Price != 0 {
		t.Errorf("Expected Price to be 0, got %f", notifyResult.Price)
	}
}
