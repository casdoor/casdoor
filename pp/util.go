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
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"text/template"
)

type PaymentDescription struct {
	ProductName        string `json:"productName"`
	ProductDisplayName string `json:"productDisplayName"`
	ProviderName       string `json:"providerName"`
}

func (p *PaymentDescription) JsonString() string {
	bytes, _ := json.Marshal(p)
	return string(bytes)
}

func (p *PaymentDescription) FromJsonString(str string) error {
	return json.Unmarshal([]byte(str), p)
}

const DefaultDescriptionTemplate = `
Product Name : {{.ProductName}} / Product Display Name : {{.ProductDisplayName}} / Provider Name : {{.ProviderName}}.
`

func (p *PaymentDescription) TemplateString() string {
	tmpl, err := template.New("paymentDescriptionTemplate").Parse(DefaultDescriptionTemplate)
	if err != nil {
		panic(err)
	}
	var output strings.Builder
	err = tmpl.Execute(&output, p)
	if err != nil {
		panic(err)
	}
	return output.String()
}

func getPriceString(price float64) string {
	priceString := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", price), "0"), ".")
	return priceString
}

func joinAttachString(tokens []string) string {
	return strings.Join(tokens, "|")
}

func parseAttachString(s string) (string, string, string, error) {
	tokens := strings.Split(s, "|")
	if len(tokens) != 3 {
		return "", "", "", fmt.Errorf("parseAttachString() error: len(tokens) expected 3, got: %d", len(tokens))
	}
	return tokens[0], tokens[1], tokens[2], nil
}

func int64ToFloat64Price(price int64) float64 {
	return float64(price) / 100
}

func float64ToInt64Price(price float64) int64 {
	return int64(math.Round(price * 100))
}
