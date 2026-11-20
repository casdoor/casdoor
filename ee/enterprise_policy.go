// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetEnterprisePolicy returns the organization's enterprise policy, or nil when there is none or the license gives
// a lower edition than the enterprise one
func GetEnterprisePolicy(organization *object.Organization) *object.EnterprisePolicy {
	if organization == nil || organization.EnterprisePolicy == nil || !edition.Has(edition.Enterprise) {
		return nil
	}
	return organization.EnterprisePolicy
}

type cachedEnterprisePolicy struct {
	policy    *object.EnterprisePolicy
	fetchTime time.Time
}

var enterprisePolicyCache sync.Map

// GetCachedEnterprisePolicy is the enterprise policy of an organization cached for 10 seconds, for the checks on
// every request and every loaded user
func GetCachedEnterprisePolicy(owner string) *object.EnterprisePolicy {
	if owner == "" || !edition.Has(edition.Enterprise) {
		return nil
	}
	if value, ok := enterprisePolicyCache.Load(owner); ok {
		cached := value.(*cachedEnterprisePolicy)
		if time.Since(cached.fetchTime) < 10*time.Second {
			return cached.policy
		}
	}

	var policy *object.EnterprisePolicy
	organization := &object.Organization{}
	existed, err := object.GetOrmer().Engine.Cols("enterprise_policy").Where("owner = ? and name = ?", "admin", owner).Get(organization)
	if err == nil && existed {
		policy = GetEnterprisePolicy(organization)
	}
	enterprisePolicyCache.Store(owner, &cachedEnterprisePolicy{policy: policy, fetchTime: time.Now()})
	return policy
}

func clearEnterprisePolicyCache(owner string) {
	enterprisePolicyCache.Delete(owner)
}

// ValidateEnterprisePolicy checks the policy an admin saves: the three admin lists don't overlap, and whoever turns
// on the separation stays able to change it, as only the security admin may
func ValidateEnterprisePolicy(policy *object.EnterprisePolicy, editor *object.User, lang string) error {
	if policy == nil {
		return nil
	}
	for _, field := range policy.EncryptedFields {
		if !util.InSlice(EncryptableUserFields, field) {
			return fmt.Errorf(i18n.Translate(lang, "enterprise:The field: %s cannot be encrypted"), field)
		}
	}
	if len(policy.EncryptedFields) > 0 && !IsFieldEncryptionKeyConfigured() {
		return errors.New(i18n.Translate(lang, "enterprise:Set the configuration fieldEncryptionKey before encrypting personal fields"))
	}
	if err := validateCertLoginPolicy(policy.CertLogin, lang); err != nil {
		return err
	}
	if !util.InSlice([]string{"", "Delete", "Write"}, policy.McpApproval) {
		return fmt.Errorf(i18n.Translate(lang, "enterprise:Unknown approval of agent changes: %s"), policy.McpApproval)
	}
	for _, rule := range policy.SodRules {
		if rule == nil || rule.Name == "" || (len(rule.Roles) < 2 && len(rule.Tools) < 2) {
			return errors.New(i18n.Translate(lang, "enterprise:A separation-of-duty rule needs a name and at least two roles or two MCP tools"))
		}
	}
	if !policy.SeparationOfDuties {
		return nil
	}

	seen := map[string]string{}
	for duty, users := range map[string][]string{object.AdminDutySystem: policy.SystemAdmins, object.AdminDutySecurity: policy.SecurityAdmins, object.AdminDutyAudit: policy.AuditAdmins} {
		for _, user := range users {
			if other, ok := seen[user]; ok && other != duty {
				return fmt.Errorf(i18n.Translate(lang, "enterprise:The user: %s cannot be more than one of the system, security and audit admins"), user)
			}
			seen[user] = duty
		}
	}
	if len(policy.SecurityAdmins) == 0 || len(policy.AuditAdmins) == 0 {
		return errors.New(i18n.Translate(lang, "enterprise:The separation of admins needs at least one security admin and one audit admin"))
	}
	if editor != nil && !util.InSlice(policy.SecurityAdmins, editor.Name) {
		return errors.New(i18n.Translate(lang, "enterprise:Add yourself to the security admins before separating the admins"))
	}
	return nil
}

// AdminDutyError is a request refused by the separation of the three admins, Key is its message in the "enterprise"
// section of the translations
type AdminDutyError struct {
	Key string
}

func (e *AdminDutyError) Error() string {
	return e.Key
}

var auditApiKeywords = []string{"record", "consent-log", "security-overview"}

var securityApiObjects = []string{"permission", "permissions", "role", "roles", "model", "models", "enforcer", "enforcers", "adapter", "adapters", "cert", "certs", "privileged-account", "privileged-accounts"}

// selfServiceApis are the APIs every admin uses on their own account, whatever their duty
var selfServiceApis = []string{
	"/api/login", "/api/logout", "/api/signin", "/api/signout", "/api/sso-logout", "/api/set-password", "/api/mfa/",
	"/api/send-verification-code", "/api/verify-code", "/api/set-preferred-mfa", "/api/delete-mfa", "/api/webauthn/",
	"/api/upload-resource", "/api/get-account", "/api/userinfo", "/api/get-my-access-review-items", "/api/decide-access-review-items",
	"/api/get-privileged-access", "/api/request-privileged-access", "/api/approve-privileged-request", "/api/reject-privileged-request",
	"/api/cancel-privileged-request", "/api/return-privileged-request", "/api/reveal-privileged-password",
}

func isAuditApi(urlPath string) bool {
	for _, keyword := range auditApiKeywords {
		if strings.Contains(urlPath, keyword) {
			return true
		}
	}
	return false
}

func isSecurityApi(urlPath string) bool {
	if strings.Contains(urlPath, "-polic") || urlPath == "/api/rotate-cert" || urlPath == "/api/import-otp-devices" || urlPath == "/api/rotate-privileged-account" || strings.HasSuffix(urlPath, "-access-review") {
		return true
	}
	for _, prefix := range []string{"/api/add-", "/api/update-", "/api/delete-"} {
		if strings.HasPrefix(urlPath, prefix) && util.InSlice(securityApiObjects, strings.TrimPrefix(urlPath, prefix)) {
			return true
		}
	}
	return false
}

func isSelfServiceApi(urlPath string, subOwner string, subName string, objOwner string, objName string) bool {
	for _, api := range selfServiceApis {
		if urlPath == api || (strings.HasSuffix(api, "/") && strings.HasPrefix(urlPath, api)) || strings.HasPrefix(urlPath, api+"/") {
			return true
		}
	}
	return urlPath == "/api/update-user" && subOwner == objOwner && subName == objName
}

// CheckAdminDuty refuses what an admin's duty doesn't cover when the organization separates the system, security and
// audit admins: the audit admin only reads, the security admin only manages permissions, roles, models, certs and
// the organization's security settings, the system admin manages everything else but can't read the audit records
func CheckAdminDuty(subOwner string, subName string, method string, urlPath string, objOwner string, objName string) error {
	if subOwner == "" || subOwner == "anonymous" || subOwner == "app" {
		return nil
	}
	policy := GetCachedEnterprisePolicy(subOwner)
	if policy == nil || !policy.SeparationOfDuties {
		return nil
	}

	user, err := object.LookupUser(subOwner, subName)
	if err != nil || user == nil || !(user.IsAdmin || user.IsGlobalAdmin()) {
		return nil
	}

	isRead := method == "GET" || method == "HEAD" || method == "OPTIONS"
	switch policy.GetAdminDuty(subName) {
	case object.AdminDutyAudit:
		if isRead || urlPath == "/api/verify-records" || isSelfServiceApi(urlPath, subOwner, subName, objOwner, objName) {
			return nil
		}
		return &AdminDutyError{Key: "The audit admin can only view, not change anything"}
	case object.AdminDutySecurity:
		if isAuditApi(urlPath) {
			return &AdminDutyError{Key: "Only the audit admin can view the audit records"}
		}
		if isRead || isSecurityApi(urlPath) || urlPath == "/api/update-organization" || isSelfServiceApi(urlPath, subOwner, subName, objOwner, objName) {
			return nil
		}
		return &AdminDutyError{Key: "The security admin can only manage permissions, roles, certs and security settings"}
	default:
		if isAuditApi(urlPath) {
			return &AdminDutyError{Key: "Only the audit admin can view the audit records"}
		}
		if !isRead && isSecurityApi(urlPath) {
			return &AdminDutyError{Key: "Only the security admin can manage permissions, roles, certs and security settings"}
		}
		return nil
	}
}

// AfterEnterprisePolicyChange drops the cached policy and rewrites the users' personal fields when the encrypted
// fields changed
func AfterEnterprisePolicyChange(oldOrganization *object.Organization, newOrganization *object.Organization) {
	clearEnterprisePolicyCache(newOrganization.Name)
	if oldOrganization != nil {
		clearEnterprisePolicyCache(oldOrganization.Name)
	}

	oldFields, newFields := []string{}, []string{}
	if oldOrganization != nil && oldOrganization.EnterprisePolicy != nil {
		oldFields = oldOrganization.EnterprisePolicy.EncryptedFields
	}
	if newOrganization.EnterprisePolicy != nil {
		newFields = newOrganization.EnterprisePolicy.EncryptedFields
	}
	if util.StructToJson(oldFields) == util.StructToJson(newFields) {
		return
	}
	util.SafeGoroutine(func() {
		count, err := ReencryptOrganizationUsers(newOrganization.Name)
		if err != nil {
			logs.Error("failed to rewrite the encrypted fields of the users of the organization: %s: %v", newOrganization.Name, err)
			return
		}
		logs.Info("rewrote the encrypted fields of %d users of the organization: %s", count, newOrganization.Name)
	})
}

// CheckOrganizationDutyChange stops a system or audit admin from changing the security settings of an organization
// that separates the admins
func CheckOrganizationDutyChange(oldOrganization *object.Organization, newOrganization *object.Organization, editor *object.User) error {
	if oldOrganization == nil || editor == nil {
		return nil
	}
	policy := GetCachedEnterprisePolicy(editor.Owner)
	if policy == nil || !policy.SeparationOfDuties || policy.GetAdminDuty(editor.Name) == object.AdminDutySecurity {
		return nil
	}
	if util.StructToJson(oldOrganization.EnterprisePolicy) != util.StructToJson(newOrganization.EnterprisePolicy) ||
		util.StructToJson(oldOrganization.SecurityPolicy) != util.StructToJson(newOrganization.SecurityPolicy) ||
		oldOrganization.PasswordType != newOrganization.PasswordType || util.StructToJson(oldOrganization.PasswordOptions) != util.StructToJson(newOrganization.PasswordOptions) {
		return &AdminDutyError{Key: "Only the security admin can manage permissions, roles, certs and security settings"}
	}
	return nil
}
