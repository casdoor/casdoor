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

import "math"

// Fixed exchange rates (temporary implementation as per requirements)
// All rates represent how many units of the currency equal 1 USD
// Example: EUR: 0.92 means 1 USD = 0.92 EUR
var exchangeRates = map[string]float64{
	"USD": 1.0,
	"EUR": 0.92,
	"GBP": 0.79,
	"JPY": 149.50,
	"CNY": 7.24,
	"AUD": 1.52,
	"CAD": 1.39,
	"CHF": 0.88,
	"HKD": 7.82,
	"SGD": 1.34,
	"INR": 83.12,
	"KRW": 1319.50,
	"BRL": 4.97,
	"MXN": 17.09,
	"ZAR": 18.15,
	"RUB": 92.50,
	"TRY": 32.15,
	"NZD": 1.67,
	"SEK": 10.35,
	"NOK": 10.72,
	"DKK": 6.87,
	"PLN": 3.91,
	"THB": 34.50,
	"MYR": 4.47,
	"IDR": 15750.00,
	"PHP": 55.50,
	"VND": 24500.00,
}

// GetExchangeRate returns the exchange rate from fromCurrency to toCurrency
func GetExchangeRate(fromCurrency, toCurrency string) float64 {
	if fromCurrency == toCurrency {
		return 1.0
	}

	// Default to USD if currency not found
	fromRate, fromExists := exchangeRates[fromCurrency]
	if !fromExists {
		fromRate = 1.0
	}

	toRate, toExists := exchangeRates[toCurrency]
	if !toExists {
		toRate = 1.0
	}

	// Convert from source currency to USD, then from USD to target currency
	// Example: EUR to JPY = (1/0.92) * 149.50 = USD/EUR * JPY/USD
	return toRate / fromRate
}

// ConvertCurrency converts an amount from one currency to another using exchange rates
func ConvertCurrency(amount float64, fromCurrency, toCurrency string) float64 {
	if fromCurrency == toCurrency {
		return amount
	}

	rate := GetExchangeRate(fromCurrency, toCurrency)
	converted := amount * rate
	return math.Round(converted*1e8) / 1e8
}

func AddPrices(price1 float64, price2 float64) float64 {
	res := price1 + price2
	return math.Round(res*1e8) / 1e8
}
