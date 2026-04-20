// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"github.com/casdoor/casdoor/object"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GetPrometheusInfo
// @Title GetPrometheusInfo
// @Tag System API
// @Description get Prometheus Info
// @Success 200 {object} object.PrometheusInfo The Response object
// @router /get-prometheus-info [get]
func (c *ApiController) GetPrometheusInfo() {
	_, ok := c.RequireAdmin()
	if !ok {
		return
	}
	prometheusInfo, err := object.GetPrometheusInfo()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(prometheusInfo)
}

// GetMetrics
// @Title GetMetrics
// @Tag System API
// @Description get Prometheus metrics. Accessible either by an admin session or by
// a valid Key (created at /keys) supplied via ?accessKey=...&accessSecret=... query params.
// @Success 200 {string} Prometheus metrics in text format
// @router /metrics [get]
func (c *ApiController) GetMetrics() {
	accessKey := c.Ctx.Input.Query("accessKey")
	accessSecret := c.Ctx.Input.Query("accessSecret")

	if accessKey != "" || accessSecret != "" {
		_, err := object.ValidateKeyByType(accessKey, accessSecret, "Prometheus")
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else {
		_, ok := c.RequireAdmin()
		if !ok {
			return
		}
	}

	promhttp.Handler().ServeHTTP(c.Ctx.ResponseWriter, c.Ctx.Request)
}
