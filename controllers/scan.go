// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/scan"
)

// Scan
// @Title Scan
// @Tag Scan API
// @Description run scan provider (type=Security Scan), persist result to provider metadata, and return parsed result
// @Param   owner   query  string  true  "The provider owner"
// @Param   name    query  string  true  "The provider name"
// @Param   target  query  string  false "Optional scan target"
// @Success 200 {object} controllers.Response The Response object
// @router /scan [get]
func (c *ApiController) Scan() {
	owner := strings.TrimSpace(c.GetString("owner"))
	name := strings.TrimSpace(c.GetString("name"))
	target := strings.TrimSpace(c.GetString("target"))
	if owner == "" || name == "" {
		c.ResponseError("provider owner and name are required")
		return
	}

	providerId := fmt.Sprintf("%s/%s", owner, name)
	configuredProvider, err := object.GetProvider(providerId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if configuredProvider == nil {
		c.ResponseError("provider does not exist")
		return
	}

	if !c.requireProviderPermission(configuredProvider) {
		return
	}

	if configuredProvider.Category != "Scan" || configuredProvider.Type != "Security Scan" {
		c.ResponseError("provider type Security Scan is required")
		return
	}

	if strings.EqualFold(configuredProvider.SubType, "Url") && target == "" && strings.TrimSpace(configuredProvider.Content) == "" {
		c.ResponseError("target URL is required for Url scan")
		return
	}

	scanProvider, err := scan.GetScanProviderFromProvider(configuredProvider)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	rawResult, err := scanProvider.Scan(target, "")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	parsedResult, err := scanProvider.ParseResult(rawResult)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	configuredProvider.Metadata = parsedResult
	_, err = object.UpdateProvider(configuredProvider.GetId(), configuredProvider)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	var result interface{}
	if err := json.Unmarshal([]byte(parsedResult), &result); err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(result)
}
