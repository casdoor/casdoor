// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package controllers

import (
	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/ee"
	"github.com/casdoor/casdoor/util"
)

// requireLicensedAdmin checks the license of the feature and that the current user administers the organization in
// the query parameter "owner", then returns that organization
func (c *ApiController) requireLicensedAdmin(editionName string, feature string) (string, bool) {
	if err := edition.Require(editionName, feature); err != nil {
		c.ResponseError(err.Error())
		return "", false
	}
	owner := c.Ctx.Input.Query("owner")
	if owner == "" || !c.IsAdminOfOrganization(owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return "", false
	}
	return owner, true
}

func (c *ApiController) getPageParams() (int, string, string, string, string) {
	limit := util.ParseInt(c.Ctx.Input.Query("pageSize"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return limit, c.Ctx.Input.Query("field"), c.Ctx.Input.Query("value"), c.Ctx.Input.Query("sortField"), c.Ctx.Input.Query("sortOrder")
}

// GetSigninRiskEvents
// @Title GetSigninRiskEvents
// @Tag Adaptive Risk API
// @Description list the sign-ins of an organization that adaptive risk found a signal in, the newest first
// @Param   owner     query    string  true        "The organization"
// @Success 200 {array} ee.SigninRiskEvent The Response object
// @router /get-signin-risk-events [get]
func (c *ApiController) GetSigninRiskEvents() {
	owner, ok := c.requireLicensedAdmin(edition.Commercial, edition.FeatureAdaptiveRisk)
	if !ok {
		return
	}
	limit, field, value, sortField, sortOrder := c.getPageParams()
	count, err := ee.GetSigninRiskEventCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	events, err := ee.GetPaginationSigninRiskEvents(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(events, paginator.Nums())
}

// GetSigninRiskSummary
// @Title GetSigninRiskSummary
// @Tag Adaptive Risk API
// @Description count the risky sign-ins of an organization in the last days by decision, signal, user and IP
// @Param   owner     query    string  true        "The organization"
// @Param   days      query    int     false       "The days, 7 by default"
// @Success 200 {object} ee.SigninRiskSummary The Response object
// @router /get-signin-risk-summary [get]
func (c *ApiController) GetSigninRiskSummary() {
	owner, ok := c.requireLicensedAdmin(edition.Commercial, edition.FeatureAdaptiveRisk)
	if !ok {
		return
	}
	summary, err := ee.GetSigninRiskSummary(owner, util.ParseInt(c.Ctx.Input.Query("days")))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(summary, ee.DefaultRiskScores)
}
