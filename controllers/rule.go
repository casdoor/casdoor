// Copyright 2023 The casbin Authors. All Rights Reserved.
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
	"errors"
	"net"
	"strings"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/hsluoyz/modsecurity-go/seclang/parser"
)

// GetRules
// @Title GetRules
// @Tag Rule API
// @Description get rules
// @Param   owner     query    string  true        "The owner of rules"
// @Success 200 {array} object.Rule The Response object
// @router /get-rules [get]
func (c *ApiController) GetRules() {
	owner := c.Ctx.Input.Query("owner")
	if owner == "admin" {
		owner = ""
	}
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		rules, err := object.GetRules(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(rules)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetRuleCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		rules, err := object.GetPaginationRules(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(rules, paginator.Nums())
	}
}

// GetRule
// @Title GetRule
// @Tag Rule API
// @Description get rule
// @Param   id     query    string  true        "The id ( owner/name ) of the rule"
// @Success 200 {object} object.Rule The Response object
// @router /get-rule [get]
func (c *ApiController) GetRule() {
	id := c.Ctx.Input.Query("id")
	rule, err := object.GetRule(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(rule)
}

// AddRule
// @Title AddRule
// @Tag Rule API
// @Description add rule
// @Param   body    body   object.Rule  true        "The details of the rule"
// @Success 200 {object} controllers.Response The Response object
// @router /add-rule [post]
func (c *ApiController) AddRule() {
	currentTime := util.GetCurrentTime()
	rule := object.Rule{
		CreatedTime: currentTime,
		UpdatedTime: currentTime,
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &rule)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	err = checkExpressions(rule.Expressions, rule.Type)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = wrapActionResponse(object.AddRule(&rule))
	c.ServeJSON()
}

// UpdateRule
// @Title UpdateRule
// @Tag Rule API
// @Description update rule
// @Param   id     query    string  true        "The id ( owner/name ) of the rule"
// @Param   body    body   object.Rule  true        "The details of the rule"
// @Success 200 {object} controllers.Response The Response object
// @router /update-rule [post]
func (c *ApiController) UpdateRule() {
	var rule object.Rule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &rule)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	err = checkExpressions(rule.Expressions, rule.Type)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	id := c.Ctx.Input.Query("id")
	c.Data["json"] = wrapActionResponse(object.UpdateRule(id, &rule))
	c.ServeJSON()
}

// DeleteRule
// @Title DeleteRule
// @Tag Rule API
// @Description delete rule
// @Param   body    body   object.Rule  true        "The details of the rule"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-rule [post]
func (c *ApiController) DeleteRule() {
	var rule object.Rule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &rule)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteRule(&rule))
	c.ServeJSON()
}

func checkExpressions(expressions []*object.Expression, ruleType string) error {
	values := make([]string, len(expressions))
	for i, expression := range expressions {
		values[i] = expression.Value
	}
	switch ruleType {
	case "WAF":
		return checkWafRule(values)
	case "IP":
		return checkIpRule(values)
	case "IP Rate Limiting":
		return checkIpRateRule(expressions)
	case "Compound":
		return checkCompoundRules(values)
	}
	return nil
}

func checkWafRule(rules []string) error {
	for _, rule := range rules {
		scanner := parser.NewSecLangScannerFromString(rule)
		_, err := scanner.AllDirective()
		if err != nil {
			return err
		}
	}
	return nil
}

func checkIpRule(ipLists []string) error {
	for _, ipList := range ipLists {
		for _, ip := range strings.Split(ipList, ",") {
			_, _, err := net.ParseCIDR(ip)
			if net.ParseIP(ip) == nil && err != nil {
				return errors.New("Invalid IP address: " + ip)
			}
		}
	}
	return nil
}

func checkIpRateRule(expressions []*object.Expression) error {
	if len(expressions) != 1 {
		return errors.New("IP Rate Limiting rule must have exactly one expression")
	}
	expression := expressions[0]
	_, err := util.ParseIntWithError(expression.Operator)
	if err != nil {
		return err
	}
	_, err = util.ParseIntWithError(expression.Value)
	if err != nil {
		return err
	}
	return nil
}

func checkCompoundRules(rules []string) error {
	_, err := object.GetRulesByRuleIds(rules)
	if err != nil {
		return err
	}
	return nil
}
