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
	"fmt"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
)

func (p *Cert) IsHsmKey() bool {
	return p.KeyStorage == CertKeyStorageHsm
}

// ProvisioningConfig pushes the accounts of an application's users to the system behind it: created when a user can
// use the application, updated when the user changes, and disabled or deleted when the user leaves, is forbidden or no
// longer belongs to the chosen groups. A downstream system with a SCIM 2.0 API is the target itself, a legacy one is
// reached through a database syncer whose connection and column mapping write its user table.
type ProvisioningConfig struct {
	IsEnabled bool   `json:"isEnabled"`
	Type      string `json:"type"`

	// SCIM: the base URL ending before "/Users" and the bearer token of the SCIM API
	Url   string `json:"url"`
	Token string `json:"token"`
	// SCIM: the user field that becomes "userName": "Name", "Email", "Phone" or "Id"
	UserNameField string `json:"userNameField"`

	// Database: the database syncer ("<organization>/<name>") whose connection and table columns are used
	Syncer string `json:"syncer"`

	// only the users in these groups (or the groups under them), every user of the organization when empty
	Groups []string `json:"groups"`
	// "Disable" or "Delete" the account of a user who leaves
	Deprovision string `json:"deprovision"`
	// reconcile every user every so many minutes, 0 for only on changes and on demand
	SyncIntervalMinutes int    `json:"syncIntervalMinutes"`
	LastSyncTime        string `json:"lastSyncTime"`
}

func (config *ProvisioningConfig) IsActive() bool {
	return config != nil && config.IsEnabled
}

func (config *ProvisioningConfig) GetFeature() (string, string) {
	if config.Type == ProvisioningTypeDatabase {
		return edition.Enterprise, edition.FeatureDbProvisioning
	}
	return edition.Commercial, edition.FeatureScimProvisioning
}

const PermissionResourceTypeMcpTool = "MCP tool"

// PermissionCondition limits an "MCP tool" permission to the calls whose argument passes it, e.g. the "factory" argument
// must be one of the user's groups. The value is a comma-separated list that may refer to the calling user, such as
// ${user.name}, ${user.owner}, ${user.groups}, ${user.groupNames}, ${user.affiliation} and ${user.properties.<key>}.
type PermissionCondition struct {
	Argument string `json:"argument"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// GetKeyId is the "kid" of the tokens the cert signs: the cert name until the key is rotated
func (p *Cert) GetKeyId() string {
	if p.KeyId != "" {
		return p.KeyId
	}
	return p.Name
}

// getCertificateByKeyId picks the public key that verifies a token: the key before the last rotation for the
// tokens it signed, otherwise the current one
func (p *Cert) getCertificateByKeyId(kid string) string {
	if kid != "" && p.PreviousCertificate != "" && kid == p.PreviousKeyId && kid != p.GetKeyId() {
		return p.PreviousCertificate
	}
	return p.Certificate
}

// FormFill is how the browser extension signs users into a legacy system that has no SSO, by filling its sign-in
// form with the user's managed account of the application
type FormFill struct {
	Enabled bool `json:"enabled"`
	// the sign-in pages, e.g. "https://oa.example.com/login*", "*.example.com/login" or "http://10.0.0.8:8080/"
	Urls []string `json:"urls"`
	// CSS selectors of the form fields, found automatically when empty
	UsernameSelector string `json:"usernameSelector"`
	PasswordSelector string `json:"passwordSelector"`
	SubmitSelector   string `json:"submitSelector"`
	AutoSubmit       bool   `json:"autoSubmit"`
	// whether users may save the account they sign in with by hand, so that it is filled the next time
	AllowSave bool `json:"allowSave"`
	// the desktop (C/S) programs whose sign-in windows the desktop autofill client fills
	Programs []*FormFillProgram `json:"programs"`
	// whether the browser extension reads the image captcha of the sign-in page (on the user's computer), the CSS
	// selectors of the captcha input and picture (found automatically when empty), and its characters: "" (digits
	// when the page says so, else letters and digits), "Digits", "Alphanumeric" or "Any"
	Captcha              bool   `json:"captcha"`
	CaptchaSelector      string `json:"captchaSelector"`
	CaptchaImageSelector string `json:"captchaImageSelector"`
	CaptchaCharset       string `json:"captchaCharset"`
}

func (formFill *FormFill) IsOn() bool {
	return formFill != nil && formFill.Enabled
}

// FormFillTarget is where a client is about to fill a sign-in form: a web page, or a window of a desktop program
type FormFillTarget struct {
	Url string
	// the full path of the program's executable, and the sign-in window's title
	Program   string
	Title     string
	Publisher string
	FileHash  string
	// the client is starting the program with the account in its command line, there is no window yet
	Launch bool
}

func (target *FormFillTarget) String() string {
	if target.Program == "" {
		return target.Url
	}
	if target.Title == "" {
		return target.Program
	}
	return fmt.Sprintf("%s (%s)", target.Program, target.Title)
}

func (user *User) HasManagedAccountOf(applicationName string) bool {
	for _, account := range user.ManagedAccounts {
		if account.Application == applicationName && account.Username != "" {
			return true
		}
	}
	return false
}

// DelegatedAdmin is a user who is not an organization admin but administers some groups (with their subgroups)
// and applications of the organization, 分级管理
type DelegatedAdmin struct {
	User         string   `json:"user"`
	Groups       []string `json:"groups"`
	Applications []string `json:"applications"`
	Actions      []string `json:"actions"`
}

func (policy *EnterprisePolicy) GetDelegatedAdmin(username string) *DelegatedAdmin {
	if policy == nil {
		return nil
	}
	for _, delegatedAdmin := range policy.DelegatedAdmins {
		if delegatedAdmin != nil && delegatedAdmin.User == username {
			return delegatedAdmin
		}
	}
	return nil
}

// IsDelegatedAdmin tells whether the policy lists the user as a delegated admin
func (policy *EnterprisePolicy) IsDelegatedAdmin(username string) bool {
	return policy.GetDelegatedAdmin(username) != nil
}

// getProtectedUsers are the users of the organization a delegated admin never administers: the other delegated
// admins and the system, security and audit admins
func (policy *EnterprisePolicy) GetProtectedUsers() map[string]bool {
	protected := map[string]bool{}
	if policy == nil {
		return protected
	}
	for _, delegatedAdmin := range policy.DelegatedAdmins {
		if delegatedAdmin != nil {
			protected[delegatedAdmin.User] = true
		}
	}
	for _, users := range [][]string{policy.SystemAdmins, policy.SecurityAdmins, policy.AuditAdmins} {
		for _, user := range users {
			protected[user] = true
		}
	}
	return protected
}

// CertLoginPolicy lets the users of an organization sign in with a digital certificate, usually kept in a UKey: the
// certificate must be issued by one of the trusted CAs, and one of its fields is matched against a user attribute
type CertLoginPolicy struct {
	Enabled bool `json:"enabled"`
	// the PEM certificates of the trusted CAs, RSA or ECDSA
	TrustedCas string `json:"trustedCas"`
	// the certificate field that identifies the user, one of CertLoginFields
	CertField string `json:"certField"`
	// the user attribute it is matched against, one of CertLoginUserFields
	UserField string `json:"userField"`
	// optional, the CRLs that the certificate is checked against, separated by commas
	CrlUrl string `json:"crlUrl"`
	// the full URL of /api/login/cert on the port that asks for client certificates, the login page's button goes
	// there, the same origin's /api/login/cert when empty (a reverse proxy passing the client certificate)
	LoginUrl string `json:"loginUrl"`
	// optional, an action of kind "cert" installed from the Marketplace that checks the certificates instead of
	// Casdoor, for the algorithms Casdoor doesn't verify itself
	Verifier string `json:"verifier"`
}

func (application *Application) IsCertLoginEnabled() bool {
	return application.HasSigninMethod(SigninMethodCertificate)
}

func (policy *CertLoginPolicy) IsEnabled() bool {
	return policy != nil && policy.Enabled
}

// FormFillProgram is a desktop (C/S) program whose sign-in window the desktop autofill client fills
type FormFillProgram struct {
	Name string `json:"name"`
	// the executable: a file name like "u8client.exe", or a full path where * matches anything, like
	// "C:\Program Files\UFIDA\*\u8client.exe"
	Path string `json:"path"`
	// the subject CN of the Authenticode certificate that the executable must be signed with, and its SHA-256, both
	// optional, so that another program with the same name doesn't get the password
	Publisher string `json:"publisher"`
	FileHash  string `json:"fileHash"`
	// regular expressions of the sign-in window's title and its window class, any window of the program when empty
	WindowTitle string `json:"windowTitle"`
	WindowClass string `json:"windowClass"`
	// the controls, like "id=txtUser", "name=Password;type=Edit" or "type=Edit;index=1", found automatically when empty
	UsernameControl string `json:"usernameControl"`
	PasswordControl string `json:"passwordControl"`
	SubmitControl   string `json:"submitControl"`
	// "" fills the controls (UI Automation, then Win32 messages), "Keyboard" types the key sequence into the window
	// for the programs whose controls can't be seen, like the self-drawn ones
	FillMethod  string `json:"fillMethod"`
	KeySequence string `json:"keySequence"`
	// how the client starts the program from its menu or the portal, the arguments may have {username} and {password}
	// for the programs that take the account in their command line
	LaunchPath string `json:"launchPath"`
	LaunchArgs string `json:"launchArgs"`
}

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

	DelegatedAdmins []*DelegatedAdmin `json:"delegatedAdmins"`
	CertLogin       *CertLoginPolicy  `json:"certLogin"`

	// which changes that agents make through Casdoor's own MCP server wait for an admin's approval: "" none,
	// "Delete" the deletions, "Write" all of them
	McpApproval string `json:"mcpApproval"`
}

// SodRule says that no user may hold two or more of the roles at once, such as the requester and the approver, and
// that no agent may be allowed two or more of the MCP tools ("<server>/<tool>") at once, such as paying and approving
type SodRule struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
	Tools []string `json:"tools"`
}

// MinTamperProofRecordDays is how long the records of an organization with tamper-proof records are kept at least,
// the 6 months of the MLPS 2.0 (等保 2.0) level 3
const MinTamperProofRecordDays = 180

func (policy *EnterprisePolicy) isEmpty() bool {
	if policy == nil {
		return true
	}
	return !policy.SeparationOfDuties && len(policy.SystemAdmins) == 0 && len(policy.SecurityAdmins) == 0 && len(policy.AuditAdmins) == 0 &&
		!policy.TamperProofRecords && len(policy.EncryptedFields) == 0 && len(policy.SodRules) == 0 && !policy.RecordConsents &&
		len(policy.DelegatedAdmins) == 0 && !policy.CertLogin.IsEnabled() && policy.McpApproval == ""
}

// getPublic is the part of the policy that the sign-in pages need, the button of the certificate login, so that the
// admin lists, separation-of-duty rules, delegated admins and trusted CAs aren't handed to anonymous visitors
func (policy *EnterprisePolicy) getPublic() *EnterprisePolicy {
	if policy == nil || policy.CertLogin == nil {
		return nil
	}
	return &EnterprisePolicy{CertLogin: &CertLoginPolicy{Enabled: policy.CertLogin.Enabled, LoginUrl: policy.CertLogin.LoginUrl}}
}

// getAdminDuty tells which of the three admins the user is when the organization separates them: a listed security
// or audit admin, otherwise the system admin
func (policy *EnterprisePolicy) GetAdminDuty(username string) string {
	if util.InSlice(policy.AuditAdmins, username) {
		return AdminDutyAudit
	}
	if util.InSlice(policy.SecurityAdmins, username) {
		return AdminDutySecurity
	}
	return AdminDutySystem
}

// SecurityPolicy holds the organization's security and risk control settings, which need a commercial edition license
type SecurityPolicy struct {
	InactiveTimeoutMinutes  int `json:"inactiveTimeoutMinutes"`
	DisableInactiveUserDays int `json:"disableInactiveUserDays"`

	// the installed actions of kind "password-check", such as the checks against leaked or common passwords
	PasswordCheckPlugins     []string `json:"passwordCheckPlugins"`
	StuffingFailureThreshold int      `json:"stuffingFailureThreshold"`
	StuffingBlockMinutes     int      `json:"stuffingBlockMinutes"`

	AlertOnNewLocation bool `json:"alertOnNewLocation"`
	AlertOnNewDevice   bool `json:"alertOnNewDevice"`

	// adaptive risk: the signals of a sign-in add up to a score, which asks for MFA or refuses the sign-in
	AdaptiveRisk          bool           `json:"adaptiveRisk"`
	RiskMfaScore          int            `json:"riskMfaScore"`
	RiskDenyScore         int            `json:"riskDenyScore"`
	RiskScores            map[string]int `json:"riskScores"`
	RiskTrustedNetworks   []string       `json:"riskTrustedNetworks"`
	RiskBlockedCountries  []string       `json:"riskBlockedCountries"`
	RiskCountryHeader     string         `json:"riskCountryHeader"`
	ImpossibleTravelHours int            `json:"impossibleTravelHours"`

	MaskSensitiveData    bool `json:"maskSensitiveData"`
	AllowAccountDeletion bool `json:"allowAccountDeletion"`
	AllowDataExport      bool `json:"allowDataExport"`

	MergeByWechatUnionId bool `json:"mergeByWechatUnionId"`

	SmsPerPhoneHourly  int  `json:"smsPerPhoneHourly"`
	SmsPerPhoneDaily   int  `json:"smsPerPhoneDaily"`
	SmsPerPrefixHourly int  `json:"smsPerPrefixHourly"`
	SmsDomesticOnly    bool `json:"smsDomesticOnly"`
	SmsDailyAlertCount int  `json:"smsDailyAlertCount"`
	SmsDailyLimit      int  `json:"smsDailyLimit"`

	// the installed action of kind "password-migration", with the address and secret of the old system it asks
	PasswordMigrationPlugin string `json:"passwordMigrationPlugin"`
	PasswordMigrationUrl    string `json:"passwordMigrationUrl"`
	PasswordMigrationSecret string `json:"passwordMigrationSecret"`

	CertExpireAlertDays int  `json:"certExpireAlertDays"`
	AutoRotateCerts     bool `json:"autoRotateCerts"`

	RecordMcpToolCalls bool `json:"recordMcpToolCalls"`

	AlertEmails []string `json:"alertEmails"`
}

// getPublic keeps only what the pages of a signed-in user need
func (policy *SecurityPolicy) getPublic() *SecurityPolicy {
	if policy == nil {
		return nil
	}
	return &SecurityPolicy{
		AllowAccountDeletion: policy.AllowAccountDeletion,
		AllowDataExport:      policy.AllowDataExport,
		MaskSensitiveData:    policy.MaskSensitiveData,
	}
}

const (
	ProvisioningTypeScim     = "SCIM"
	ProvisioningTypeDatabase = "Database"

	ProvisioningDeprovisionDisable = "Disable"
	ProvisioningDeprovisionDelete  = "Delete"

	ProvisionedAccountActive   = "Active"
	ProvisionedAccountDisabled = "Disabled"
	ProvisionedAccountDeleted  = "Deleted"

	ProvisioningTaskPending  = "Pending"
	ProvisioningTaskRetrying = "Retrying"
	ProvisioningTaskDone     = "Done"
	ProvisioningTaskFailed   = "Failed"
)

const (
	AdminDutySystem   = "system"
	AdminDutySecurity = "security"
	AdminDutyAudit    = "audit"
)

const SigninMethodCertificate = "Certificate"

// CertKeyStorageHsm is the KeyStorage of a cert whose private key stays in the HSM
const CertKeyStorageHsm = "HSM"
