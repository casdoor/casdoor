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

package object

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/pp"
)

func TestNewRecord_NotifyPayment_PreservesPaymentData(t *testing.T) {
	// Create a mock payment response that would be returned by NotifyPayment API
	mockPayment := &Payment{
		Owner:       "test-org",
		Name:        "test-payment",
		CreatedTime: "2024-01-28T00:00:00Z",
		State:       pp.PaymentStatePaid,
		Price:       100.0,
		Currency:    "USD",
		User:        "test-user",
		Message:     "Payment successful",
	}

	// Create a mock response that would be set by the controller
	mockResponse := Response{
		Status: "ok",
		Msg:    "",
		Data:   mockPayment,
	}

	// Create a mock context with notify-payment action
	req, _ := http.NewRequest("POST", "/api/notify-payment/test-org/test-payment", strings.NewReader("{}"))
	req.Header.Set("Accept-Language", "en")
	
	ctx := &context.Context{
		Request: req,
		Input:   context.NewInput(),
	}
	ctx.Input.Context = ctx
	ctx.Input.RequestBody = []byte("{}")
	
	// Simulate what the controller does - sets the response data
	respJson, _ := json.Marshal(mockResponse)
	var responseMap map[string]interface{}
	json.Unmarshal(respJson, &responseMap)
	ctx.Input.SetData("json", responseMap)

	// Call NewRecord
	record, err := NewRecord(ctx)
	if err != nil {
		t.Fatalf("NewRecord failed: %v", err)
	}

	// Verify that the action is "notify-payment"
	if record.Action != "notify-payment" {
		t.Errorf("Expected action 'notify-payment', got '%s'", record.Action)
	}

	// Print the actual response for debugging
	t.Logf("Actual response: %s", record.Response)

	// Verify that the Response field contains the payment data
	// The response should be in the format: {status:"ok", msg:"", data:<payment_json>}
	if !strings.Contains(record.Response, `status:"ok"`) {
		t.Error("Response should contain status ok")
	}

	// Most importantly, verify that the payment data is included in the response
	if !strings.Contains(record.Response, "data:") {
		t.Error("Response should contain 'data:' field with payment information")
	}

	// Verify that the payment state is included
	if !strings.Contains(record.Response, string(pp.PaymentStatePaid)) {
		t.Errorf("Response should contain payment state '%s'", pp.PaymentStatePaid)
	}
}

func TestNewRecord_BuyProduct_PreservesProductData(t *testing.T) {
	// Verify that buy-product action also preserves data (existing behavior)
	mockData := map[string]string{"product": "test-product", "price": "100"}
	mockResponse := Response{
		Status: "ok",
		Msg:    "",
		Data:   mockData,
	}

	req, _ := http.NewRequest("POST", "/api/buy-product", strings.NewReader("{}"))
	req.Header.Set("Accept-Language", "en")
	
	ctx := &context.Context{
		Request: req,
		Input:   context.NewInput(),
	}
	ctx.Input.Context = ctx
	ctx.Input.RequestBody = []byte("{}")
	
	respJson, _ := json.Marshal(mockResponse)
	ctx.Input.SetData("json", json.RawMessage(respJson))

	record, err := NewRecord(ctx)
	if err != nil {
		t.Fatalf("NewRecord failed: %v", err)
	}

	if !strings.Contains(record.Response, "data:") {
		t.Error("buy-product should preserve data field")
	}
}

func TestNewRecord_OtherActions_DoNotPreserveData(t *testing.T) {
	// Verify that other actions do NOT preserve data (existing behavior)
	mockData := map[string]string{"some": "data"}
	mockResponse := Response{
		Status: "ok",
		Msg:    "",
		Data:   mockData,
	}

	req, _ := http.NewRequest("POST", "/api/some-other-action", strings.NewReader("{}"))
	req.Header.Set("Accept-Language", "en")
	
	ctx := &context.Context{
		Request: req,
		Input:   context.NewInput(),
	}
	ctx.Input.Context = ctx
	ctx.Input.RequestBody = []byte("{}")
	
	respJson, _ := json.Marshal(mockResponse)
	ctx.Input.SetData("json", json.RawMessage(respJson))

	record, err := NewRecord(ctx)
	if err != nil {
		t.Fatalf("NewRecord failed: %v", err)
	}

	// Other actions should NOT have data in the response
	if strings.Contains(record.Response, "data:") {
		t.Error("Other actions should not preserve data field")
	}
}
