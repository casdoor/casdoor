// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func getPriceString(price float64) string {
	priceString := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", price), "0"), ".")
	return priceString
}

// The attach string packs product and provider info into the one opaque field a
// gateway echoes back. Both ends go through AttachInfo, so the token order is
// fixed here and cannot be swapped at a call site.
const attachStringSeparator = "|"

var (
	attachTokenEscaper   = strings.NewReplacer("%", "%25", attachStringSeparator, "%7C")
	attachTokenUnescaper = strings.NewReplacer("%7C", attachStringSeparator, "%25", "%")
)

type AttachInfo struct {
	ProductName        string
	ProductDisplayName string
	ProviderName       string
}

func joinAttachString(r *PayReq) string {
	tokens := []string{r.ProductName, r.ProductDisplayName, r.ProviderName}
	for i, token := range tokens {
		tokens[i] = attachTokenEscaper.Replace(token)
	}
	return strings.Join(tokens, attachStringSeparator)
}

// An empty input means the gateway echoed nothing back, which is not an error.
func parseAttachString(s string) (*AttachInfo, error) {
	if s == "" {
		return &AttachInfo{}, nil
	}

	tokens := strings.Split(s, attachStringSeparator)
	if len(tokens) != 3 {
		return nil, fmt.Errorf("parseAttachString() error: len(tokens) expected 3, got: %d", len(tokens))
	}
	for i, token := range tokens {
		tokens[i] = attachTokenUnescaper.Replace(token)
	}
	return &AttachInfo{
		ProductName:        tokens[0],
		ProductDisplayName: tokens[1],
		ProviderName:       tokens[2],
	}, nil
}

func priceInt64ToFloat64(price int64) float64 {
	return float64(price) / 100
}

func priceFloat64ToInt64(price float64) int64 {
	return int64(math.Round(price * 100))
}

func priceFloat64ToString(price float64) string {
	return strconv.FormatFloat(price, 'f', 2, 64)
}

func priceStringToFloat64(price string) float64 {
	f, err := strconv.ParseFloat(price, 64)
	if err != nil {
		panic(err)
	}
	return f
}
