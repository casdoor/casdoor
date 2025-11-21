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
	"math"
	"testing"
)

func TestGetExchangeRate(t *testing.T) {
	tests := []struct {
		name         string
		fromCurrency string
		toCurrency   string
		expected     float64
	}{
		{"Same currency", "USD", "USD", 1.0},
		{"USD to EUR", "USD", "EUR", 0.92},
		{"EUR to USD", "EUR", "USD", 1.0 / 0.92},
		{"USD to JPY", "USD", "JPY", 149.50},
		{"EUR to JPY", "EUR", "JPY", 149.50 / 0.92},
		{"Unknown to USD", "XYZ", "USD", 1.0},
		{"USD to Unknown", "USD", "ABC", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExchangeRate(tt.fromCurrency, tt.toCurrency)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("GetExchangeRate(%s, %s) = %v, want %v", tt.fromCurrency, tt.toCurrency, result, tt.expected)
			}
		})
	}
}

func TestConvertCurrency(t *testing.T) {
	tests := []struct {
		name         string
		amount       float64
		fromCurrency string
		toCurrency   string
		expected     float64
	}{
		{"Same currency", 100.0, "USD", "USD", 100.0},
		{"USD to EUR", 100.0, "USD", "EUR", 92.0},
		{"EUR to USD", 92.0, "EUR", "USD", 100.0},
		{"USD to JPY", 100.0, "USD", "JPY", 14950.0},
		{"JPY to USD", 14950.0, "JPY", "USD", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertCurrency(tt.amount, tt.fromCurrency, tt.toCurrency)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("ConvertCurrency(%v, %s, %s) = %v, want %v", tt.amount, tt.fromCurrency, tt.toCurrency, result, tt.expected)
			}
		})
	}
}

func TestAddPrices(t *testing.T) {
	tests := []struct {
		name     string
		price1   float64
		price2   float64
		expected float64
	}{
		{"Simple addition", 10.5, 20.3, 30.8},
		{"Negative addition", 100.0, -50.0, 50.0},
		{"Precision test", 0.12345678, 0.87654322, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddPrices(tt.price1, tt.price2)
			if math.Abs(result-tt.expected) > 1e-8 {
				t.Errorf("AddPrices(%v, %v) = %v, want %v", tt.price1, tt.price2, result, tt.expected)
			}
		})
	}
}
