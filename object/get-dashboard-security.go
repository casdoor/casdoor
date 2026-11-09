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
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
)

const (
	dashboardInactiveDays  = 90
	dashboardCertAlertDays = 30
	dashboardSigninDays    = 14
	dashboardRecentDays    = 7
)

// DashboardSecurity is what the dashboard checks the security of an organization (or of all of them) against.
type DashboardSecurity struct {
	Users         int64 `json:"users"`
	InactiveUsers int64 `json:"inactiveUsers"`
	InactiveDays  int   `json:"inactiveDays"`

	Admins           int64 `json:"admins"`
	AdminsWithoutMfa int64 `json:"adminsWithoutMfa"`

	Organizations int `json:"organizations"`
	// how many organizations turn each protection of the security policy on
	Policies map[string]int `json:"policies"`

	Applications                   int `json:"applications"`
	ApplicationsWithSessionControl int `json:"applicationsWithSessionControl"`

	Certs             int `json:"certs"`
	CertsExpiring     int `json:"certsExpiring"`
	CertsRotated      int `json:"certsRotated"`
	CertAlertDays     int `json:"certAlertDays"`
	NearestCertExpiry int `json:"nearestCertExpiry"`

	SigninDates   []string `json:"signinDates"`
	Signins       []int    `json:"signins"`
	FailedSignins []int    `json:"failedSignins"`

	RecentDays          int   `json:"recentDays"`
	NewLocationSignins  int64 `json:"newLocationSignins"`
	NewDeviceSignins    int64 `json:"newDeviceSignins"`
	McpCalls            int64 `json:"mcpCalls"`
	McpBlockedCalls     int64 `json:"mcpBlockedCalls"`
	PasswordCheckedOrgs int   `json:"passwordCheckedOrgs"`
}

func getDashboardTable(name string) string {
	return conf.GetConfigString("tableNamePrefix") + name
}

func GetDashboardSecurity(owner string) (*DashboardSecurity, error) {
	if owner == "All" {
		owner = ""
	}
	now := time.Now()
	res := &DashboardSecurity{
		InactiveDays:      dashboardInactiveDays,
		CertAlertDays:     dashboardCertAlertDays,
		RecentDays:        dashboardRecentDays,
		NearestCertExpiry: -1,
		Policies:          map[string]int{},
	}

	err := countDashboardUsers(res, owner, now)
	if err != nil {
		return nil, err
	}
	err = countDashboardPolicies(res, owner)
	if err != nil {
		return nil, err
	}
	err = countDashboardApplications(res, owner)
	if err != nil {
		return nil, err
	}
	err = countDashboardCerts(res, owner, now)
	if err != nil {
		return nil, err
	}
	countDashboardSignins(res, owner, now)
	countDashboardRecent(res, owner, now)
	return res, nil
}

func countDashboardUsers(res *DashboardSecurity, owner string, now time.Time) error {
	table := getDashboardTable("user")
	session := ormer.Engine.Table(table).Where("is_deleted = ? AND is_forbidden = ?", false, false)
	if owner != "" {
		session = session.And("owner = ?", owner)
	}
	var err error
	res.Users, err = session.Count()
	if err != nil {
		return err
	}

	cutoff := now.AddDate(0, 0, -dashboardInactiveDays).Format(time.RFC3339)
	session = ormer.Engine.Table(table).Where("is_deleted = ? AND is_forbidden = ?", false, false).
		And("((last_signin_time != '' AND last_signin_time < ?) OR ((last_signin_time = '' OR last_signin_time IS NULL) AND created_time < ?))", cutoff, cutoff)
	if owner != "" {
		session = session.And("owner = ?", owner)
	}
	res.InactiveUsers, err = session.Count()
	if err != nil {
		return err
	}

	session = ormer.Engine.Table(table).Where("is_deleted = ? AND is_forbidden = ? AND is_admin = ?", false, false, true)
	if owner != "" {
		session = session.And("owner = ?", owner)
	}
	res.Admins, err = session.Count()
	if err != nil {
		return err
	}

	session = ormer.Engine.Table(table).Where("is_deleted = ? AND is_forbidden = ? AND is_admin = ?", false, false, true).
		And("mfa_phone_enabled = ? AND mfa_email_enabled = ? AND (preferred_mfa_type IS NULL OR preferred_mfa_type = '')", false, false)
	if owner != "" {
		session = session.And("owner = ?", owner)
	}
	res.AdminsWithoutMfa, err = session.Count()
	return err
}

func countDashboardPolicies(res *DashboardSecurity, owner string) error {
	organizations := []*Organization{}
	session := ormer.Engine.Table(getDashboardTable("organization")).Cols("owner", "name", "security_policy").Where("owner = ?", "admin")
	if owner != "" {
		session = session.And("name = ?", owner)
	}
	err := session.Find(&organizations)
	if err != nil {
		return err
	}

	res.Organizations = len(organizations)
	for _, organization := range organizations {
		policy := organization.SecurityPolicy
		if policy == nil {
			continue
		}
		on := map[string]bool{
			"inactiveTimeout": policy.InactiveTimeoutMinutes > 0,
			"disableInactive": policy.DisableInactiveUserDays > 0,
			"passwordCheck":   len(policy.PasswordCheckPlugins) > 0,
			"stuffing":        policy.StuffingFailureThreshold > 0,
			"signinAlert":     policy.AlertOnNewLocation || policy.AlertOnNewDevice,
			"masking":         policy.MaskSensitiveData,
			"smsLimit":        policy.SmsPerPhoneHourly > 0 || policy.SmsPerPhoneDaily > 0 || policy.SmsDailyLimit > 0,
			"certAlert":       policy.CertExpireAlertDays > 0,
		}
		for key, value := range on {
			if value {
				res.Policies[key]++
			}
		}
		if on["passwordCheck"] {
			res.PasswordCheckedOrgs++
		}
	}
	return nil
}

func countDashboardApplications(res *DashboardSecurity, owner string) error {
	applications := []*Application{}
	session := ormer.Engine.Table(getDashboardTable("application")).Cols("owner", "name", "organization", "enable_exclusive_signin", "inactive_timeout_minutes")
	if owner != "" {
		session = session.Where("organization = ?", owner)
	}
	err := session.Find(&applications)
	if err != nil {
		return err
	}

	res.Applications = len(applications)
	for _, application := range applications {
		if application.EnableExclusiveSignin || application.InactiveTimeoutMinutes > 0 {
			res.ApplicationsWithSessionControl++
		}
	}
	return nil
}

func countDashboardCerts(res *DashboardSecurity, owner string, now time.Time) error {
	certs := []*Cert{}
	session := ormer.Engine.Table(getDashboardTable("cert")).Cols("owner", "name", "expire_time", "rotated_time", "certificate")
	if owner != "" {
		session = session.Where("owner = ?", owner)
	}
	err := session.Find(&certs)
	if err != nil {
		return err
	}

	for _, cert := range certs {
		if cert.RotatedTime != "" {
			res.CertsRotated++
		}
		// only an SSL cert stores its expiry, a signing cert has it in its certificate
		expire := cert.ExpireTime
		if expire == "" && cert.Certificate != "" {
			expire, _ = util.GetCertExpireTime(cert.Certificate)
		}
		expireTime, err := time.Parse(time.RFC3339, expire)
		if err != nil {
			continue
		}
		res.Certs++
		days := int(expireTime.Sub(now).Hours() / 24)
		if days < dashboardCertAlertDays {
			res.CertsExpiring++
		}
		if res.NearestCertExpiry < 0 || days < res.NearestCertExpiry {
			res.NearestCertExpiry = max(days, 0)
		}
	}
	return nil
}

// countDashboardSignins counts the sign-ins of the last days from the record table, the failed ones are those
// whose response is an error. It leaves zeros when the record table cannot be read.
func countDashboardSignins(res *DashboardSecurity, owner string, now time.Time) {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(dashboardSigninDays - 1))
	index := map[string]int{}
	for i := 0; i < dashboardSigninDays; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		index[date] = i
		res.SigninDates = append(res.SigninDates, date)
	}
	res.Signins = make([]int, dashboardSigninDays)
	res.FailedSignins = make([]int, dashboardSigninDays)

	type recordItem struct {
		CreatedTime string `xorm:"created_time"`
	}
	count := func(failed bool, counts []int) {
		records := []recordItem{}
		session := ormer.Engine.Table(getDashboardTable("record")).Cols("created_time").
			Where("created_time >= ?", start.Format(time.RFC3339)).And("action = ?", "login")
		if failed {
			session = session.And("response LIKE ?", `%"status":"error"%`)
		}
		if owner != "" {
			session = session.And("organization = ?", owner)
		}
		if session.Find(&records) != nil {
			return
		}
		for _, record := range records {
			t, err := time.Parse(time.RFC3339, record.CreatedTime)
			if err != nil {
				continue
			}
			if i, ok := index[t.In(now.Location()).Format("2006-01-02")]; ok {
				counts[i]++
			}
		}
	}
	count(false, res.Signins)
	count(true, res.FailedSignins)
	for i := range res.Signins {
		res.Signins[i] -= res.FailedSignins[i]
		if res.Signins[i] < 0 {
			res.Signins[i] = 0
		}
	}
}

// countDashboardRecent counts what the commercial features recorded lately: sign-ins from a new place or device
// (only kept while the sign-in alerts are on) and the tool calls through the MCP gateway.
func countDashboardRecent(res *DashboardSecurity, owner string, now time.Time) {
	since := now.AddDate(0, 0, -dashboardRecentDays).Format(time.RFC3339)
	count := func(table string, where string, args ...interface{}) int64 {
		session := ormer.Engine.Table(getDashboardTable(table)).Where("created_time >= ?", since)
		if where != "" {
			session = session.And(where, args...)
		}
		if owner != "" {
			session = session.And("owner = ?", owner)
		}
		n, err := session.Count()
		if err != nil {
			return 0
		}
		return n
	}
	res.NewLocationSignins = count("signin_trace", "kind = ?", "location")
	res.NewDeviceSignins = count("signin_trace", "kind = ?", "device")
	res.McpCalls = count("mcp_record", "")
	res.McpBlockedCalls = count("mcp_record", "result = ?", "Blocked")
}
