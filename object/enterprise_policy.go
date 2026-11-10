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

package object

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
)

// EnterprisePolicy holds the organization's compliance and governance settings, which need an enterprise edition
// license: the separation of the system, security and audit admins (三员分立), tamper-proof audit records, encrypted
// personal fields, separation-of-duty rules between roles and the consent log
type EnterprisePolicy struct {
	SeparationOfDuties bool     `json:"separationOfDuties"`
	SystemAdmins       []string `json:"systemAdmins"`
	SecurityAdmins     []string `json:"securityAdmins"`
	AuditAdmins        []string `json:"auditAdmins"`

	TamperProofRecords bool `json:"tamperProofRecords"`

	EncryptedFields []string `json:"encryptedFields"`

	SodRules []*SodRule `json:"sodRules"`

	RecordConsents bool `json:"recordConsents"`
}

// SodRule says that no user may hold two or more of the roles at once, such as the requester and the approver
type SodRule struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

const (
	AdminDutySystem   = "system"
	AdminDutySecurity = "security"
	AdminDutyAudit    = "audit"
)

// MinTamperProofRecordDays is how long the records of an organization with tamper-proof records are kept at least,
// the 6 months of the MLPS 2.0 (等保 2.0) level 3
const MinTamperProofRecordDays = 180

func (policy *EnterprisePolicy) isEmpty() bool {
	if policy == nil {
		return true
	}
	return !policy.SeparationOfDuties && len(policy.SystemAdmins) == 0 && len(policy.SecurityAdmins) == 0 && len(policy.AuditAdmins) == 0 &&
		!policy.TamperProofRecords && len(policy.EncryptedFields) == 0 && len(policy.SodRules) == 0 && !policy.RecordConsents
}

// GetEnterprisePolicy returns the organization's enterprise policy, or nil when there is none or the license gives
// a lower edition than the enterprise one
func GetEnterprisePolicy(organization *Organization) *EnterprisePolicy {
	if organization == nil || organization.EnterprisePolicy == nil || !edition.Has(edition.Enterprise) {
		return nil
	}
	return organization.EnterprisePolicy
}

type cachedEnterprisePolicy struct {
	policy    *EnterprisePolicy
	fetchTime time.Time
}

var enterprisePolicyCache sync.Map

// GetCachedEnterprisePolicy is the enterprise policy of an organization cached for 10 seconds, for the checks on
// every request and every loaded user
func GetCachedEnterprisePolicy(owner string) *EnterprisePolicy {
	if owner == "" || !edition.Has(edition.Enterprise) {
		return nil
	}
	if value, ok := enterprisePolicyCache.Load(owner); ok {
		cached := value.(*cachedEnterprisePolicy)
		if time.Since(cached.fetchTime) < 10*time.Second {
			return cached.policy
		}
	}

	var policy *EnterprisePolicy
	organization := &Organization{}
	existed, err := ormer.Engine.Cols("enterprise_policy").Where("owner = ? and name = ?", "admin", owner).Get(organization)
	if err == nil && existed {
		policy = GetEnterprisePolicy(organization)
	}
	enterprisePolicyCache.Store(owner, &cachedEnterprisePolicy{policy: policy, fetchTime: time.Now()})
	return policy
}

func clearEnterprisePolicyCache(owner string) {
	enterprisePolicyCache.Delete(owner)
}

// getAdminDuty tells which of the three admins the user is when the organization separates them: a listed security
// or audit admin, otherwise the system admin
func (policy *EnterprisePolicy) getAdminDuty(username string) string {
	if util.InSlice(policy.AuditAdmins, username) {
		return AdminDutyAudit
	}
	if util.InSlice(policy.SecurityAdmins, username) {
		return AdminDutySecurity
	}
	return AdminDutySystem
}

// ValidateEnterprisePolicy checks the policy an admin saves: the three admin lists don't overlap, and whoever turns
// on the separation stays able to change it, as only the security admin may
func ValidateEnterprisePolicy(policy *EnterprisePolicy, editor *User, lang string) error {
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
	for _, rule := range policy.SodRules {
		if rule == nil || rule.Name == "" || len(rule.Roles) < 2 {
			return errors.New(i18n.Translate(lang, "enterprise:A separation-of-duty rule needs a name and at least two roles"))
		}
	}
	if !policy.SeparationOfDuties {
		return nil
	}

	seen := map[string]string{}
	for duty, users := range map[string][]string{AdminDutySystem: policy.SystemAdmins, AdminDutySecurity: policy.SecurityAdmins, AdminDutyAudit: policy.AuditAdmins} {
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
	"/api/upload-resource", "/api/get-account", "/api/userinfo",
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
	if strings.Contains(urlPath, "-polic") || urlPath == "/api/rotate-cert" || urlPath == "/api/import-otp-devices" || urlPath == "/api/rotate-privileged-account" {
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

	user, err := getUser(subOwner, subName)
	if err != nil || user == nil || !(user.IsAdmin || user.IsGlobalAdmin()) {
		return nil
	}

	isRead := method == "GET" || method == "HEAD" || method == "OPTIONS"
	switch policy.getAdminDuty(subName) {
	case AdminDutyAudit:
		if isRead || urlPath == "/api/verify-records" || isSelfServiceApi(urlPath, subOwner, subName, objOwner, objName) {
			return nil
		}
		return &AdminDutyError{Key: "The audit admin can only view, not change anything"}
	case AdminDutySecurity:
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
func AfterEnterprisePolicyChange(oldOrganization *Organization, newOrganization *Organization) {
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
func CheckOrganizationDutyChange(oldOrganization *Organization, newOrganization *Organization, editor *User) error {
	if oldOrganization == nil || editor == nil {
		return nil
	}
	policy := GetCachedEnterprisePolicy(editor.Owner)
	if policy == nil || !policy.SeparationOfDuties || policy.getAdminDuty(editor.Name) == AdminDutySecurity {
		return nil
	}
	if util.StructToJson(oldOrganization.EnterprisePolicy) != util.StructToJson(newOrganization.EnterprisePolicy) ||
		util.StructToJson(oldOrganization.SecurityPolicy) != util.StructToJson(newOrganization.SecurityPolicy) ||
		oldOrganization.PasswordType != newOrganization.PasswordType || util.StructToJson(oldOrganization.PasswordOptions) != util.StructToJson(newOrganization.PasswordOptions) {
		return &AdminDutyError{Key: "Only the security admin can manage permissions, roles, certs and security settings"}
	}
	return nil
}
