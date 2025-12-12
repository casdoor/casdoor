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
	"encoding/json"
	"testing"
)

func TestProductInfo(t *testing.T) {
	productInfo := ProductInfo{
		Name:        "product_test",
		DisplayName: "Test Product",
		Image:       "https://example.com/image.png",
		Description: "Test product description",
		Tag:         "test",
		Price:       10.5,
		Currency:    "USD",
		IsRecharge:  false,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(productInfo)
	if err != nil {
		t.Errorf("Failed to marshal ProductInfo: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledProductInfo ProductInfo
	err = json.Unmarshal(jsonData, &unmarshaledProductInfo)
	if err != nil {
		t.Errorf("Failed to unmarshal ProductInfo: %v", err)
	}

	// Verify fields
	if unmarshaledProductInfo.Name != productInfo.Name {
		t.Errorf("Name mismatch: expected %s, got %s", productInfo.Name, unmarshaledProductInfo.Name)
	}
	if unmarshaledProductInfo.Price != productInfo.Price {
		t.Errorf("Price mismatch: expected %f, got %f", productInfo.Price, unmarshaledProductInfo.Price)
	}
}

func TestOrderWithProductInfo(t *testing.T) {
	productInfo := ProductInfo{
		Name:        "product_recharge",
		DisplayName: "Recharge Product",
		Image:       "https://example.com/recharge.png",
		Description: "Recharge product with custom price",
		Tag:         "recharge",
		Price:       50.0, // Custom recharge price
		Currency:    "USD",
		IsRecharge:  true,
	}

	order := Order{
		Owner:       "admin",
		Name:        "order_test",
		CreatedTime: "2025-01-01T00:00:00Z",
		DisplayName: "Test Order",
		Products:    []ProductInfo{productInfo},
		User:        "test_user",
		Price:       50.0,
		Currency:    "USD",
		State:       "Created",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(order)
	if err != nil {
		t.Errorf("Failed to marshal Order: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledOrder Order
	err = json.Unmarshal(jsonData, &unmarshaledOrder)
	if err != nil {
		t.Errorf("Failed to unmarshal Order: %v", err)
	}

	// Verify Products array
	if len(unmarshaledOrder.Products) != 1 {
		t.Errorf("Expected 1 product, got %d", len(unmarshaledOrder.Products))
	}

	// Verify product info is preserved including custom price
	if len(unmarshaledOrder.Products) > 0 {
		product := unmarshaledOrder.Products[0]
		if product.Name != productInfo.Name {
			t.Errorf("Product name mismatch: expected %s, got %s", productInfo.Name, product.Name)
		}
		if product.Price != productInfo.Price {
			t.Errorf("Product price mismatch: expected %f, got %f", productInfo.Price, product.Price)
		}
		if product.IsRecharge != productInfo.IsRecharge {
			t.Errorf("Product isRecharge mismatch: expected %v, got %v", productInfo.IsRecharge, product.IsRecharge)
		}
	}
}

func TestOrderMultipleProducts(t *testing.T) {
	// Test that Order can support multiple products (for future enhancement)
	products := []ProductInfo{
		{
			Name:        "product1",
			DisplayName: "Product 1",
			Price:       10.0,
			Currency:    "USD",
			IsRecharge:  false,
		},
		{
			Name:        "product2",
			DisplayName: "Product 2",
			Price:       20.0,
			Currency:    "USD",
			IsRecharge:  false,
		},
	}

	order := Order{
		Owner:       "admin",
		Name:        "order_multi",
		CreatedTime: "2025-01-01T00:00:00Z",
		DisplayName: "Multi-Product Order",
		Products:    products,
		User:        "test_user",
		Price:       30.0,
		Currency:    "USD",
		State:       "Created",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(order)
	if err != nil {
		t.Errorf("Failed to marshal Order: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledOrder Order
	err = json.Unmarshal(jsonData, &unmarshaledOrder)
	if err != nil {
		t.Errorf("Failed to unmarshal Order: %v", err)
	}

	// Verify multiple products are preserved
	if len(unmarshaledOrder.Products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(unmarshaledOrder.Products))
	}
}
