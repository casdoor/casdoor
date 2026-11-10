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
	"time"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// requireEnterpriseAdmin checks that the enterprise feature is licensed and that the current user administers the
// organization in the query parameter "owner", then returns that organization
func (c *ApiController) requireEnterpriseAdmin(feature string) (string, bool) {
	if err := edition.Require(edition.Enterprise, feature); err != nil {
		c.ResponseError(err.Error())
		return "", false
	}
	owner := c.Ctx.Input.Query("owner")
	user := c.getCurrentUser()
	if owner == "" || user == nil || !(c.IsGlobalAdmin() || (user.IsAdmin && user.Owner == owner)) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return "", false
	}
	return owner, true
}

// checkEnterprisePolicyChange validates a changed enterprise policy and stops the admins whose duty doesn't cover it
func (c *ApiController) checkEnterprisePolicyChange(oldOrganization *object.Organization, organization *object.Organization) error {
	editor := c.getCurrentUser()
	if oldOrganization == nil {
		return object.ValidateEnterprisePolicy(organization.EnterprisePolicy, nil, c.GetAcceptLanguage())
	}

	err := object.CheckOrganizationDutyChange(oldOrganization, organization, editor)
	if err != nil {
		return err
	}
	if util.StructToJson(oldOrganization.EnterprisePolicy) == util.StructToJson(organization.EnterprisePolicy) {
		return nil
	}
	// whoever separates the admins of their own organization has to stay a security admin
	if editor != nil && editor.Owner != organization.Name {
		editor = nil
	}
	return object.ValidateEnterprisePolicy(organization.EnterprisePolicy, editor, c.GetAcceptLanguage())
}

// VerifyRecords
// @Title VerifyRecords
// @Tag Enterprise API
// @Description verify the hash chain of the tamper-proof records of an organization
// @Param   owner     query    string  true        "The organization"
// @Success 200 {object} object.RecordChainResult The Response object
// @router /verify-records [get]
func (c *ApiController) VerifyRecords() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureTamperProofLogs)
	if !ok {
		return
	}
	result, err := object.VerifyRecordChain(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}

// GetComplianceReport
// @Title GetComplianceReport
// @Tag Enterprise API
// @Description check an organization's settings against the MLPS 2.0 level 3 and the cryptography evaluation
// @Param   owner     query    string  true        "The organization"
// @Success 200 {object} object.ComplianceReport The Response object
// @router /get-compliance-report [get]
func (c *ApiController) GetComplianceReport() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureComplianceReport)
	if !ok {
		return
	}
	report, err := object.GetComplianceReport(owner, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(report)
}

// GetSodConflicts
// @Title GetSodConflicts
// @Tag Enterprise API
// @Description list the users holding roles that a separation-of-duty rule keeps apart
// @Param   owner     query    string  true        "The organization"
// @Success 200 {array} object.SodConflict The Response object
// @router /get-sod-conflicts [get]
func (c *ApiController) GetSodConflicts() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureSod)
	if !ok {
		return
	}
	conflicts, err := object.GetSodConflicts(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(conflicts)
}

// GetOrphanAccounts
// @Title GetOrphanAccounts
// @Tag Enterprise API
// @Description list the accounts that nobody uses or that are disabled but still hold roles
// @Param   owner     query    string  true        "The organization"
// @Param   days      query    int     false       "Days without signing in, 90 by default"
// @Success 200 {array} object.OrphanAccount The Response object
// @router /get-orphan-accounts [get]
func (c *ApiController) GetOrphanAccounts() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureGovernance)
	if !ok {
		return
	}
	accounts, err := object.GetOrphanAccounts(owner, util.ParseInt(c.Ctx.Input.Query("days")))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(accounts)
}

// GetSecurityOverview
// @Title GetSecurityOverview
// @Tag Enterprise API
// @Description sum up the sign-ins, failures and admin operations of an organization
// @Param   owner     query    string  true        "The organization"
// @Param   days      query    int     false       "The last days, 7 by default"
// @Param   month     query    string  false       "A month such as 2026-10, for the monthly report"
// @Success 200 {object} object.SecurityOverview The Response object
// @router /get-security-overview [get]
func (c *ApiController) GetSecurityOverview() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureSecurityOverview)
	if !ok {
		return
	}

	now := time.Now()
	var from, to time.Time
	if month := c.Ctx.Input.Query("month"); month != "" {
		start, err := time.ParseInLocation("2006-01", month, time.Local)
		if err != nil {
			c.ResponseError(fmt.Sprintf("the month: %s is not like 2026-10", month))
			return
		}
		from, to = start, start.AddDate(0, 1, 0)
	} else {
		days := util.ParseInt(c.Ctx.Input.Query("days"))
		if days <= 0 || days > 90 {
			days = 7
		}
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		from, to = today.AddDate(0, 0, 1-days), today.AddDate(0, 0, 1)
	}

	overview, err := object.GetSecurityOverview(owner, from, to)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(overview)
}

// GetConsentLogs
// @Title GetConsentLogs
// @Tag Enterprise API
// @Description list the consents and revocations of the users of an organization
// @Param   owner     query    string  true        "The organization"
// @Success 200 {array} object.ConsentLog The Response object
// @router /get-consent-logs [get]
func (c *ApiController) GetConsentLogs() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureConsentLog)
	if !ok {
		return
	}
	limit := util.ParseInt(c.Ctx.Input.Query("pageSize"))
	if limit <= 0 || limit > 500 {
		limit = 20
	}
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	count, err := object.GetConsentLogCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	consentLogs, err := object.GetPaginationConsentLogs(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(consentLogs, paginator.Nums())
}

// ImportOtpTokens
// @Title ImportOtpTokens
// @Tag Enterprise API
// @Description bind hardware OTP tokens to users from a PSKC file or a CSV of seeds
// @Param   owner     query    string  true        "The organization"
// @Success 200 {object} object.OtpTokenImportResult The Response object
// @router /import-otp-devices [post]
func (c *ApiController) ImportOtpTokens() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureOtpTokenImport)
	if !ok {
		return
	}
	var request struct {
		Tokens      string `json:"tokens"`
		Assignments string `json:"assignments"`
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	result, err := object.ImportOtpTokens(owner, request.Tokens, request.Assignments)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}
