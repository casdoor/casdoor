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
		if err := object.ValidateDelegatedAdmins(organization.Name, organization.EnterprisePolicy, c.GetAcceptLanguage()); err != nil {
			return err
		}
		return object.ValidateEnterprisePolicy(organization.EnterprisePolicy, nil, c.GetAcceptLanguage())
	}

	err := object.CheckOrganizationDutyChange(oldOrganization, organization, editor)
	if err != nil {
		return err
	}
	if util.StructToJson(oldOrganization.EnterprisePolicy) == util.StructToJson(organization.EnterprisePolicy) {
		return nil
	}
	if err = object.ValidateDelegatedAdmins(organization.Name, organization.EnterprisePolicy, c.GetAcceptLanguage()); err != nil {
		return err
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

// getSecurityPeriod reads the period of a security overview or report from the query: a month such as 2026-10, a
// range "from" and "to" of dates such as 2026-10-01 (both included), or the last "days" (7 by default)
func (c *ApiController) getSecurityPeriod() (time.Time, time.Time, bool, error) {
	if month := c.Ctx.Input.Query("month"); month != "" {
		start, err := time.ParseInLocation("2006-01", month, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, false, fmt.Errorf("the month: %s is not like 2026-10", month)
		}
		return start, start.AddDate(0, 1, 0), true, nil
	}

	if fromDate, toDate := c.Ctx.Input.Query("from"), c.Ctx.Input.Query("to"); fromDate != "" || toDate != "" {
		from, err := time.ParseInLocation("2006-01-02", fromDate, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, false, fmt.Errorf("the date: %s is not like 2026-10-01", fromDate)
		}
		to, err := time.ParseInLocation("2006-01-02", toDate, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, false, fmt.Errorf("the date: %s is not like 2026-10-01", toDate)
		}
		to = to.AddDate(0, 0, 1)
		if !to.After(from) || to.Sub(from) > 366*24*time.Hour {
			return time.Time{}, time.Time{}, false, fmt.Errorf("the period from %s to %s is not between 1 and 366 days", fromDate, toDate)
		}
		return from, to, false, nil
	}

	now := time.Now()
	days := util.ParseInt(c.Ctx.Input.Query("days"))
	if days <= 0 || days > 90 {
		days = 7
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return today.AddDate(0, 0, 1-days), today.AddDate(0, 0, 1), false, nil
}

// GetSecurityOverview
// @Title GetSecurityOverview
// @Tag Enterprise API
// @Description sum up the sign-ins, failures and admin operations of an organization
// @Param   owner     query    string  true        "The organization"
// @Param   days      query    int     false       "The last days, 7 by default"
// @Param   month     query    string  false       "A month such as 2026-10, for the monthly report"
// @Param   from      query    string  false       "The first day of a custom period, such as 2026-10-01"
// @Param   to        query    string  false       "The last day of a custom period, such as 2026-10-31"
// @Success 200 {object} object.SecurityOverview The Response object
// @router /get-security-overview [get]
func (c *ApiController) GetSecurityOverview() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureSecurityOverview)
	if !ok {
		return
	}

	from, to, _, err := c.getSecurityPeriod()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	overview, err := object.GetSecurityOverview(owner, from, to)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(overview)
}

// GetSecurityScreen
// @Title GetSecurityScreen
// @Tag Enterprise API
// @Description get what the security situation screen of an organization shows
// @Param   owner     query    string  true        "The organization"
// @Success 200 {object} object.SecurityScreen The Response object
// @router /get-security-screen [get]
func (c *ApiController) GetSecurityScreen() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureSecurityOverview)
	if !ok {
		return
	}
	screen, err := object.GetSecurityScreen(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(screen)
}

// GetSecurityReportDocx
// @Title GetSecurityReportDocx
// @Tag Enterprise API
// @Description download the security report of an organization for a period as a Word document
// @Param   owner     query    string  true        "The organization"
// @Param   month     query    string  false       "A month such as 2026-10"
// @Param   from      query    string  false       "The first day of a custom period, such as 2026-10-01"
// @Param   to        query    string  false       "The last day of a custom period, such as 2026-10-31"
// @Param   days      query    int     false       "The last days, 7 by default"
// @Success 200 {file} file The .docx file
// @router /get-security-report-docx [get]
func (c *ApiController) GetSecurityReportDocx() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureSecurityOverview)
	if !ok {
		return
	}
	from, to, isMonth, err := c.getSecurityPeriod()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	data, fileName, err := object.GetSecurityReportDocx(owner, from, to, isMonth, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Ctx.Output.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Ctx.Output.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	_ = c.Ctx.Output.Body(data)
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
