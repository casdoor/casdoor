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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/i18n"
)

// runPasswordChecks runs the organization's password check actions, returning the reason of the first that refuses
func runPasswordChecks(owner string, policy *SecurityPolicy, password string, username string, isSignin bool) string {
	for _, name := range policy.PasswordCheckPlugins {
		m, _, err := getInstalledAction(owner, name, "password-check")
		if err != nil {
			logs.Warn("the password check: %s is skipped: %s", name, err.Error())
			continue
		}
		resp, err := action.CheckPassword(m, &action.Config{}, password, username, isSignin)
		if err != nil {
			logs.Warn("the password check: %s is skipped: %s", name, err.Error())
			continue
		}
		if !resp.Ok {
			return resp.Reason
		}
	}
	return ""
}

// checkPasswordSecurityPolicy refuses the passwords that a check action of the organization refuses
func checkPasswordSecurityPolicy(organization *Organization, password string, username string, lang string) string {
	policy := GetSecurityPolicy(organization)
	if policy == nil || password == "" || len(policy.PasswordCheckPlugins) == 0 {
		return ""
	}

	reason := runPasswordChecks(organization.Name, policy, password, username, false)
	if reason != "" {
		return fmt.Sprintf(i18n.Translate(lang, "check:The password is not allowed: %s"), reason)
	}
	return ""
}

// MarkWeakPasswordOnSignin asks the user to change the password at this sign-in when a check action of the
// organization refuses it, e.g. because it has appeared in data breaches since it was set
func MarkWeakPasswordOnSignin(user *User, password string) {
	if user == nil || user.NeedUpdatePassword || password == "" {
		return
	}
	policy := GetCachedSecurityPolicy(user.Owner)
	if policy == nil || len(policy.PasswordCheckPlugins) == 0 {
		return
	}

	if runPasswordChecks(user.Owner, policy, password, user.Name, true) == "" {
		return
	}

	user.NeedUpdatePassword = true
	_, err := ormer.Engine.Where("owner = ? and name = ?", user.Owner, user.Name).Cols("need_update_password").Update(user)
	if err != nil {
		logs.Warn("the user: %s is not asked to change the password: %s", user.GetId(), err.Error())
	}
}

const stuffingWindow = 10 * time.Minute

type stuffingRecord struct {
	failures     map[string]time.Time
	blockedUntil time.Time
}

var (
	stuffingRecords = map[string]*stuffingRecord{}
	stuffingLock    sync.Mutex
)

// CheckCredentialStuffing refuses the sign-ins from an IP that has been blocked for trying many accounts
func CheckCredentialStuffing(organization string, clientIp string, lang string) error {
	policy := GetCachedSecurityPolicy(organization)
	if policy == nil || policy.StuffingFailureThreshold <= 0 || clientIp == "" {
		return nil
	}

	stuffingLock.Lock()
	defer stuffingLock.Unlock()
	record := stuffingRecords[organization+"/"+clientIp]
	if record != nil && time.Now().Before(record.blockedUntil) {
		minutes := int(time.Until(record.blockedUntil).Minutes()) + 1
		return newSigninError(SigninReasonAccountFrozen, fmt.Sprintf(i18n.Translate(lang, "check:Too many accounts failed to sign in from your IP address, please wait for %d minutes and try again"), minutes))
	}
	return nil
}

// GetCredentialStuffingBlockedIps lists the IPs that are blocked from signing in to the organization right now
func GetCredentialStuffingBlockedIps(organization string) []string {
	stuffingLock.Lock()
	defer stuffingLock.Unlock()

	res := []string{}
	now := time.Now()
	prefix := organization + "/"
	for key, record := range stuffingRecords {
		if strings.HasPrefix(key, prefix) && now.Before(record.blockedUntil) {
			res = append(res, strings.TrimPrefix(key, prefix))
		}
	}
	sort.Strings(res)
	return res
}

// RecordCredentialStuffingFailure counts the distinct accounts that fail to sign in from an IP, and blocks the IP
// when they reach the organization's StuffingFailureThreshold within 10 minutes
func RecordCredentialStuffingFailure(organization string, clientIp string, username string) {
	policy := GetCachedSecurityPolicy(organization)
	if policy == nil || policy.StuffingFailureThreshold <= 0 || clientIp == "" || username == "" {
		return
	}

	stuffingLock.Lock()
	defer stuffingLock.Unlock()

	now := time.Now()
	for key, record := range stuffingRecords {
		for name, t := range record.failures {
			if now.Sub(t) > stuffingWindow {
				delete(record.failures, name)
			}
		}
		if len(record.failures) == 0 && now.After(record.blockedUntil) {
			delete(stuffingRecords, key)
		}
	}

	key := organization + "/" + clientIp
	record := stuffingRecords[key]
	if record == nil {
		record = &stuffingRecord{failures: map[string]time.Time{}}
		stuffingRecords[key] = record
	}
	record.failures[strings.ToLower(username)] = now
	if len(record.failures) < policy.StuffingFailureThreshold || now.Before(record.blockedUntil) {
		return
	}

	blockMinutes := policy.StuffingBlockMinutes
	if blockMinutes <= 0 {
		blockMinutes = 60
	}
	record.blockedUntil = now.Add(time.Duration(blockMinutes) * time.Minute)
	count := len(record.failures)
	record.failures = map[string]time.Time{}

	go func() {
		org, err := getOrganization("admin", organization)
		if err != nil || org == nil {
			return
		}
		logs.Warn("blocked %s for %d minutes: %d accounts of the organization: %s failed to sign in from it within 10 minutes", clientIp, blockMinutes, count, organization)
		SendSecurityAlert(org, fmt.Sprintf("[%s] Credential stuffing blocked from %s", organization, clientIp),
			fmt.Sprintf("%d different accounts failed to sign in from %s within 10 minutes. The IP address is blocked from signing in for %d minutes.", count, clientIp, blockMinutes))
	}()
}
