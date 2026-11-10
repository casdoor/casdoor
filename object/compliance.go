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
	"strings"

	"github.com/casdoor/casdoor/cred"
	"github.com/casdoor/casdoor/gm"
	"github.com/casdoor/casdoor/util"
)

// ComplianceItem is one check of an organization's settings against a requirement of the MLPS 2.0 level 3 (等保 2.0
// 三级, GB/T 22239-2019) or the Chinese cryptography evaluation (密评, GB/T 39786-2021)
type ComplianceItem struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Requirement string `json:"requirement"`
	Status      string `json:"status"`
	Detail      string `json:"detail"`
}

const (
	ComplianceStatusPass = "pass"
	ComplianceStatusWarn = "warn"
	ComplianceStatusFail = "fail"
)

type ComplianceReport struct {
	Organization string            `json:"organization"`
	CreatedTime  string            `json:"createdTime"`
	Pass         int               `json:"pass"`
	Warn         int               `json:"warn"`
	Fail         int               `json:"fail"`
	Items        []*ComplianceItem `json:"items"`
}

type complianceText struct {
	zh string
	en string
}

func (t complianceText) get(lang string) string {
	if strings.HasPrefix(lang, "zh") {
		return t.zh
	}
	return t.en
}

// GetComplianceReport checks the organization, its applications and their certs against the requirements an
// assessor goes through for an identity system. It only reads the settings, the report is what to fix before the
// assessment, not the assessment itself.
func GetComplianceReport(organizationName string, lang string) (*ComplianceReport, error) {
	organization, err := getOrganization("admin", organizationName)
	if err != nil {
		return nil, err
	}
	if organization == nil {
		return nil, fmt.Errorf("the organization: %s does not exist", organizationName)
	}
	applications, err := GetOrganizationApplications("admin", organizationName)
	if err != nil {
		return nil, err
	}

	security := organization.SecurityPolicy
	if security == nil {
		security = &SecurityPolicy{}
	}
	enterprise := organization.EnterprisePolicy
	if enterprise == nil {
		enterprise = &EnterprisePolicy{}
	}

	report := &ComplianceReport{Organization: organizationName, CreatedTime: util.GetCurrentTime(), Items: []*ComplianceItem{}}
	add := func(category complianceText, title complianceText, requirement complianceText, status string, detail string) {
		report.Items = append(report.Items, &ComplianceItem{
			Category:    category.get(lang),
			Title:       title.get(lang),
			Requirement: requirement.get(lang),
			Status:      status,
			Detail:      detail,
		})
		switch status {
		case ComplianceStatusPass:
			report.Pass++
		case ComplianceStatusWarn:
			report.Warn++
		default:
			report.Fail++
		}
	}
	text := func(zh string, en string) complianceText {
		return complianceText{zh: zh, en: en}
	}
	pick := func(zh string, en string) string {
		return text(zh, en).get(lang)
	}

	identification := text("身份鉴别", "Identification and authentication")
	accessControl := text("访问控制", "Access control")
	audit := text("安全审计", "Security audit")
	cryptography := text("数据完整性和保密性（密评）", "Data integrity and confidentiality (cryptography evaluation)")
	privacy := text("个人信息保护", "Personal information protection")
	intrusion := text("入侵防范", "Intrusion prevention")

	// 8.1.4.1 a) passwords are complex and changed regularly
	options := organization.PasswordOptions
	complexity := ComplianceStatusFail
	if util.InSlice(options, "AtLeast8") && util.InSlice(options, "Aa123") && util.InSlice(options, "SpecialChar") {
		complexity = ComplianceStatusPass
	} else if util.InSlice(options, "AtLeast8") && util.InSlice(options, "Aa123") {
		complexity = ComplianceStatusWarn
	}
	add(identification, text("口令复杂度", "Password complexity"),
		text("8.1.4.1 a) 口令具有复杂度要求：至少 8 位，包含大小写字母、数字和特殊字符", "8.1.4.1 a) passwords of at least 8 characters with upper and lower case letters, digits and special characters"),
		complexity, pick("当前口令规则：", "Current password rules: ")+strings.Join(options, ", "))

	expire := ComplianceStatusFail
	if organization.PasswordExpireDays > 0 && organization.PasswordExpireDays <= 90 {
		expire = ComplianceStatusPass
	} else if organization.PasswordExpireDays > 90 {
		expire = ComplianceStatusWarn
	}
	add(identification, text("口令定期更换", "Password expiry"),
		text("8.1.4.1 a) 口令定期更换，建议不超过 90 天", "8.1.4.1 a) passwords are changed regularly, every 90 days at most"),
		expire, fmt.Sprintf(pick("口令有效期：%d 天（0 表示永不过期）", "Password expiry: %d days (0 is never)"), organization.PasswordExpireDays))

	// 8.1.4.1 b) failed sign-ins lock the account, idle sessions end
	weakApplications := []string{}
	for _, application := range applications {
		limit := application.FailedSigninLimit
		if limit == 0 {
			limit = DefaultFailedSigninLimit
		}
		if limit > 5 {
			weakApplications = append(weakApplications, fmt.Sprintf("%s (%d)", application.Name, limit))
		}
	}
	lockout := ComplianceStatusPass
	detail := pick("所有应用的登录失败次数上限都不超过 5 次", "Every application locks the account after 5 failed sign-ins at most")
	if len(weakApplications) > 0 {
		lockout = ComplianceStatusFail
		detail = pick("登录失败次数上限超过 5 次的应用：", "Applications allowing more than 5 failed sign-ins: ") + strings.Join(weakApplications, ", ")
	}
	add(identification, text("登录失败处理", "Failed sign-in handling"),
		text("8.1.4.1 b) 具有登录失败处理功能，连续失败后锁定账号", "8.1.4.1 b) accounts are locked after repeated failed sign-ins"), lockout, detail)

	timeout := ComplianceStatusFail
	if security.InactiveTimeoutMinutes > 0 && security.InactiveTimeoutMinutes <= 30 {
		timeout = ComplianceStatusPass
	} else if security.InactiveTimeoutMinutes > 30 {
		timeout = ComplianceStatusWarn
	}
	add(identification, text("登录连接超时", "Idle session timeout"),
		text("8.1.4.1 b) 登录连接超时自动退出，建议不超过 30 分钟", "8.1.4.1 b) idle sessions end, after 30 minutes at most"),
		timeout, fmt.Sprintf(pick("组织空闲超时：%d 分钟（0 表示不限制）", "Organization idle timeout: %d minutes (0 is none)"), security.InactiveTimeoutMinutes))

	// 8.1.4.1 d) two or more kinds of authentication, one of them cryptographic
	mfa := ComplianceStatusFail
	requiredMfa := []string{}
	for _, item := range organization.MfaItems {
		if item != nil && item.Rule == "Required" {
			requiredMfa = append(requiredMfa, item.Name)
		}
	}
	if len(requiredMfa) > 0 {
		mfa = ComplianceStatusPass
	}
	add(identification, text("双因素认证", "Multi-factor authentication"),
		text("8.1.4.1 d) 采用口令、密码技术、生物技术等两种或两种以上组合的鉴别技术", "8.1.4.1 d) two or more kinds of authentication, such as a password and a one-time code"),
		mfa, pick("强制开启的多因素认证：", "Required multi-factor authentication: ")+strings.Join(requiredMfa, ", "))

	masterPassword := ComplianceStatusPass
	detail = pick("未设置万能密码", "No master password")
	if organization.MasterPassword != "" {
		masterPassword = ComplianceStatusFail
		detail = pick("组织设置了万能密码，可以登录任意账号，测评会判为后门", "The organization has a master password that signs in to any account, an assessor treats it as a backdoor")
	}
	add(identification, text("万能密码", "Master password"),
		text("8.1.4.1 a) 不存在可绕过身份鉴别的通用口令", "8.1.4.1 a) no password signs in to every account"), masterPassword, detail)

	// 8.1.4.2 access control: default accounts, separated admins, inactive accounts
	if organizationName == "built-in" {
		defaultAdmin := ComplianceStatusPass
		detail = pick("内置管理员 admin 已改掉默认口令", "The built-in admin no longer has the default password")
		admin, err := getUser("built-in", "admin")
		if err != nil {
			return nil, err
		}
		if admin != nil {
			credManager := cred.GetCredManager(admin.PasswordType)
			if admin.Password == "123" || (credManager != nil && credManager.IsPasswordCorrect("123", admin.Password, admin.PasswordSalt)) {
				defaultAdmin = ComplianceStatusFail
				detail = pick("内置管理员 admin 仍是默认口令 123", "The built-in admin still has the default password 123")
			}
		}
		add(accessControl, text("默认账户口令", "Default account password"),
			text("8.1.4.2 b) 重命名或删除默认账户，修改默认账户的默认口令", "8.1.4.2 b) default accounts are renamed or removed and their default passwords changed"), defaultAdmin, detail)
	}

	duties := ComplianceStatusFail
	detail = pick("未开启三员分立", "The admins are not separated")
	if enterprise.SeparationOfDuties {
		duties = ComplianceStatusPass
		detail = fmt.Sprintf(pick("系统管理员 %d 人、安全管理员 %d 人、审计管理员 %d 人", "%d system, %d security and %d audit admins"), len(enterprise.SystemAdmins), len(enterprise.SecurityAdmins), len(enterprise.AuditAdmins))
	}
	add(accessControl, text("三员分立", "Separated admins"),
		text("8.1.4.2 d) 授予管理用户所需的最小权限，实现管理用户的权限分离（系统管理员、安全管理员、审计管理员）", "8.1.4.2 d) admins get the least privilege, the system, security and audit admins are separated"), duties, detail)

	inactive := ComplianceStatusFail
	if security.DisableInactiveUserDays > 0 && security.DisableInactiveUserDays <= 90 {
		inactive = ComplianceStatusPass
	} else if security.DisableInactiveUserDays > 90 {
		inactive = ComplianceStatusWarn
	}
	add(accessControl, text("多余和过期账户", "Unused accounts"),
		text("8.1.4.2 c) 及时删除或停用多余的、过期的账户", "8.1.4.2 c) unused and expired accounts are removed or disabled in time"),
		inactive, fmt.Sprintf(pick("超过 %d 天未登录自动禁用（0 表示不禁用）", "Disabled after %d days without signing in (0 is never)"), security.DisableInactiveUserDays))

	sod := ComplianceStatusWarn
	if len(enterprise.SodRules) > 0 {
		sod = ComplianceStatusPass
	}
	add(accessControl, text("职责分离", "Separation of duties"),
		text("8.1.4.2 d) 互斥的角色不授予同一用户", "8.1.4.2 d) one user does not hold conflicting roles"),
		sod, fmt.Sprintf(pick("职责分离规则 %d 条", "%d separation-of-duty rules"), len(enterprise.SodRules)))

	// 8.1.4.3 security audit: protected records kept 6 months, sent out of the system
	tamperProof := ComplianceStatusFail
	if enterprise.TamperProofRecords {
		tamperProof = ComplianceStatusPass
	}
	add(audit, text("审计记录防篡改", "Tamper-proof audit records"),
		text("8.1.4.3 c) 对审计记录进行保护，避免受到未预期的删除、修改或覆盖", "8.1.4.3 c) audit records are protected against deletion, modification and overwriting"),
		tamperProof, pick("开启后审计日志按 SM3 哈希链保存，可以校验完整性", "With it the audit records form an SM3 hash chain that can be verified"))

	retention := ComplianceStatusFail
	if organization.RecordRetentionDays <= 0 || organization.RecordRetentionDays >= MinTamperProofRecordDays || enterprise.TamperProofRecords {
		retention = ComplianceStatusPass
	}
	add(audit, text("审计记录留存", "Audit record retention"),
		text("8.1.4.3 c) 审计记录留存不少于 6 个月（网络安全法第二十一条）", "8.1.4.3 c) audit records are kept for 6 months at least"),
		retention, fmt.Sprintf(pick("组织日志保留：%d 天（0 表示永久）", "Organization record retention: %d days (0 is forever)"), organization.RecordRetentionDays))

	auditProviders, err := GetProvidersByCategory(organizationName, "Audit")
	if err != nil {
		return nil, err
	}
	forward := ComplianceStatusWarn
	if len(auditProviders) > 0 {
		forward = ComplianceStatusPass
	}
	add(audit, text("集中审计", "Central audit"),
		text("8.1.5.4 c) 对分散在各设备上的审计数据进行收集汇总和集中分析", "8.1.5.4 c) audit data is collected and analysed centrally"),
		forward, fmt.Sprintf(pick("审计类身份源（Syslog 外发）%d 个", "%d audit providers (syslog)"), len(auditProviders)))

	// GB/T 39786-2021: passwords stored with SM3, tokens signed with SM2, personal data encrypted with SM4
	passwordType := organization.PasswordType
	storage := ComplianceStatusFail
	if passwordType == "pbkdf2-sm3" {
		storage = ComplianceStatusPass
	} else if util.InSlice([]string{"bcrypt", "argon2id", "pbkdf2-salt", "pbkdf2-django"}, passwordType) {
		storage = ComplianceStatusWarn
	}
	add(cryptography, text("口令存储", "Password storage"),
		text("GB/T 39786 8.3：采用密码技术保证鉴别数据存储的机密性和完整性，宜使用 SM3", "GB/T 39786 8.3: authentication data is stored with cryptography, preferably SM3"),
		storage, pick("口令哈希算法：", "Password hash: ")+passwordType)

	sm2Applications, otherApplications := []string{}, []string{}
	for _, application := range applications {
		if application.TokenSigningMethod == gm.Alg {
			sm2Applications = append(sm2Applications, application.Name)
		} else {
			otherApplications = append(otherApplications, application.Name)
		}
	}
	signing := ComplianceStatusWarn
	detail = pick("未使用 SM2 签名的应用：", "Applications not signing with SM2: ") + strings.Join(otherApplications, ", ")
	if len(applications) > 0 && len(otherApplications) == 0 {
		signing = ComplianceStatusPass
		detail = pick("所有应用的令牌都用 SM2 签名", "Every application signs its tokens with SM2")
	}
	add(cryptography, text("令牌签名", "Token signatures"),
		text("GB/T 39786 8.3：采用密码技术保证重要数据传输的完整性，宜使用 SM2 数字签名", "GB/T 39786 8.3: the integrity of important data in transit is protected with cryptography, preferably SM2 signatures"),
		signing, detail)

	encryption := ComplianceStatusWarn
	if util.InSlice(enterprise.EncryptedFields, "idCard") {
		encryption = ComplianceStatusPass
	}
	add(cryptography, text("敏感数据加密存储", "Encrypted personal data"),
		text("GB/T 39786 8.3：采用密码技术保证重要数据存储的机密性，宜使用 SM4", "GB/T 39786 8.3: important data is stored encrypted, preferably with SM4"),
		encryption, pick("SM4 加密的字段：", "Fields encrypted with SM4: ")+strings.Join(enterprise.EncryptedFields, ", "))

	// the Personal Information Protection Law
	masking := ComplianceStatusWarn
	if security.MaskSensitiveData {
		masking = ComplianceStatusPass
	}
	add(privacy, text("后台脱敏显示", "Masked personal data"),
		text("个人信息保护法第五十一条：采取去标识化等安全技术措施", "PIPL article 51: personal information is de-identified where possible"),
		masking, pick("管理后台显示手机号、邮箱、身份证时是否脱敏", "Whether the console masks phone numbers, emails and ID numbers"))

	consent := ComplianceStatusWarn
	if enterprise.RecordConsents {
		consent = ComplianceStatusPass
	}
	add(privacy, text("授权同意留痕", "Consent log"),
		text("个人信息保护法第十四条：个人同意应当由个人在充分知情的前提下自愿、明确作出，处理者应能证明", "PIPL article 14: consent is given knowingly and explicitly, and can be proven"),
		consent, pick("记录用户每次授权和撤销授权的时间、应用、范围和 IP", "Logs the time, application, scopes and IP of each consent and revocation"))

	stuffing := ComplianceStatusWarn
	if security.StuffingFailureThreshold > 0 {
		stuffing = ComplianceStatusPass
	}
	add(intrusion, text("撞库防护", "Credential stuffing protection"),
		text("8.1.4.5 入侵防范：能够检测到对重要节点进行入侵的行为", "8.1.4.5: attacks on important nodes are detected"),
		stuffing, fmt.Sprintf(pick("同一 IP 登录失败 %d 个账号后封禁（0 表示不封禁）", "An IP is blocked after failing on %d accounts (0 is never)"), security.StuffingFailureThreshold))

	return report, nil
}
