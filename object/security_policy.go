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
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
	"github.com/robfig/cron/v3"
)

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

// GetSecurityPolicy returns the organization's security policy, or nil when there is none or the license
// gives a lower edition than the commercial one
func GetSecurityPolicy(organization *Organization) *SecurityPolicy {
	if organization == nil || organization.SecurityPolicy == nil || !edition.Has(edition.Commercial) {
		return nil
	}
	return organization.SecurityPolicy
}

func GetSecurityPolicyByOwner(owner string) *SecurityPolicy {
	if !edition.Has(edition.Commercial) {
		return nil
	}
	organization, err := getOrganization("admin", owner)
	if err != nil {
		return nil
	}
	return GetSecurityPolicy(organization)
}

type cachedSecurityPolicy struct {
	policy    *SecurityPolicy
	fetchTime time.Time
}

var securityPolicyCache sync.Map

// GetCachedSecurityPolicy is GetSecurityPolicyByOwner cached for a minute, for the checks on every request
func GetCachedSecurityPolicy(owner string) *SecurityPolicy {
	if !edition.Has(edition.Commercial) || owner == "" {
		return nil
	}
	if value, ok := securityPolicyCache.Load(owner); ok {
		cached := value.(*cachedSecurityPolicy)
		if time.Since(cached.fetchTime) < time.Minute {
			return cached.policy
		}
	}
	policy := GetSecurityPolicyByOwner(owner)
	securityPolicyCache.Store(owner, &cachedSecurityPolicy{policy: policy, fetchTime: time.Now()})
	return policy
}

// getOrganizationEmailProvider picks the email provider of the organization's default application, then any
// email provider owned by the organization
func getOrganizationEmailProvider(organization *Organization) (*Provider, error) {
	if organization.DefaultApplication != "" {
		application, err := getApplication("admin", organization.DefaultApplication)
		if err != nil {
			return nil, err
		}
		if application != nil {
			provider, err := application.GetEmailProvider("")
			if err != nil {
				return nil, err
			}
			if provider != nil {
				return provider, nil
			}
		}
	}

	providers := []*Provider{}
	err := ormer.Engine.Where("category = ? and (owner = ? or owner = ?)", "Email", organization.Name, "admin").Find(&providers)
	if err != nil {
		return nil, err
	}
	for _, provider := range providers {
		if provider.Owner == organization.Name {
			return provider, nil
		}
	}
	if len(providers) > 0 {
		return providers[0], nil
	}
	return nil, nil
}

func sendOrganizationEmail(organization *Organization, title string, content string, dest []string) error {
	if len(dest) == 0 {
		return nil
	}

	provider, err := getOrganizationEmailProvider(organization)
	if err != nil {
		return err
	}
	if provider == nil {
		return fmt.Errorf("the organization: %s has no email provider", organization.Name)
	}

	sender := organization.DisplayName
	if sender == "" {
		sender = organization.Name
	}
	return SendEmail(provider, title, content, dest, sender)
}

// SendSecurityAlert emails the organization's alert addresses
func SendSecurityAlert(organization *Organization, title string, content string) {
	policy := GetSecurityPolicy(organization)
	if policy == nil || len(policy.AlertEmails) == 0 {
		return
	}

	err := sendOrganizationEmail(organization, title, content, policy.AlertEmails)
	if err != nil {
		logs.Warn("security alert of the organization: %s is not sent: %s", organization.Name, err.Error())
	}
}

func parseAnyTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05.999999", "2006-01-02T15:04:05", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// DisableInactiveUsers forbids the users who haven't signed in for the organization's DisableInactiveUserDays,
// counting from the creation time for the users who never signed in. Administrators are kept.
func DisableInactiveUsers() error {
	organizations, err := GetOrganizations("admin")
	if err != nil {
		return err
	}

	for _, organization := range organizations {
		policy := GetSecurityPolicy(organization)
		if policy == nil || policy.DisableInactiveUserDays <= 0 {
			continue
		}

		cutoff := time.Now().Add(-time.Duration(policy.DisableInactiveUserDays) * 24 * time.Hour)
		users := []*User{}
		err = ormer.Engine.Cols("owner", "name", "created_time", "last_signin_time", "is_admin").
			Where("owner = ? and is_forbidden = ? and is_deleted = ? and is_admin = ?", organization.Name, false, false, false).Find(&users)
		if err != nil {
			return err
		}

		names := []string{}
		for _, user := range users {
			lastTime, ok := parseAnyTime(user.LastSigninTime)
			if !ok {
				lastTime, ok = parseAnyTime(user.CreatedTime)
			}
			if !ok || lastTime.After(cutoff) {
				continue
			}

			_, err = ormer.Engine.Where("owner = ? and name = ?", user.Owner, user.Name).Cols("is_forbidden", "updated_time").
				Update(&User{IsForbidden: true, UpdatedTime: util.GetCurrentTime()})
			if err != nil {
				return err
			}
			names = append(names, user.Name)
		}

		if len(names) > 0 {
			logs.Info("disabled %d users of the organization: %s who haven't signed in for %d days", len(names), organization.Name, policy.DisableInactiveUserDays)
			shown := names
			if len(shown) > 50 {
				shown = shown[:50]
			}
			SendSecurityAlert(organization, fmt.Sprintf("[%s] %d inactive accounts disabled", organization.Name, len(names)),
				fmt.Sprintf("These accounts haven't signed in for %d days and have been disabled: %s", policy.DisableInactiveUserDays, strings.Join(shown, ", ")))
		}
	}
	return nil
}

func runSecurityJob(name string, job func() error) {
	if !edition.Has(edition.Commercial) {
		return
	}
	if err := job(); err != nil {
		logs.Warn("security job %s failed: %s", name, err.Error())
	}
}

// InitSecurityPolicyJobs runs the scheduled jobs of the security policies
func InitSecurityPolicyJobs() {
	go func() {
		time.Sleep(time.Minute)
		runSecurityJob("disable inactive users", DisableInactiveUsers)
		runSecurityJob("check certs", CheckCertExpiration)
	}()

	cronJob := cron.New()
	_, err := cronJob.AddFunc("30 3 * * *", func() {
		runSecurityJob("disable inactive users", DisableInactiveUsers)
		runSecurityJob("check certs", CheckCertExpiration)
	})
	if err != nil {
		logs.Warn("security jobs are not scheduled: %s", err.Error())
		return
	}
	cronJob.Start()
}
