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
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
)

type SecurityHour struct {
	Hour    string `json:"hour"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
}

type SecurityRegion struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Count   int    `json:"count"`
	Failure int    `json:"failure"`
}

type SecurityEvent struct {
	Time   string `json:"time"`
	Kind   string `json:"kind"`
	User   string `json:"user"`
	Ip     string `json:"ip"`
	Detail string `json:"detail"`
}

type SecurityAlert struct {
	Kind   string `json:"kind"`
	Level  string `json:"level"`
	Count  int    `json:"count"`
	Detail string `json:"detail"`
}

// SecurityScreen is what the security situation screen of an organization shows: today's numbers, the last 24 hours
// and 30 days, where the sign-ins come from, the latest events and what needs attention
type SecurityScreen struct {
	Organization  string `json:"organization"`
	DisplayName   string `json:"displayName"`
	GeneratedTime string `json:"generatedTime"`

	SigninSuccessToday int     `json:"signinSuccessToday"`
	SigninFailureToday int     `json:"signinFailureToday"`
	FailureRateToday   float64 `json:"failureRateToday"`
	BlockedToday       int     `json:"blockedToday"`
	ActiveUsersToday   int     `json:"activeUsersToday"`
	ActiveUsers7Days   int     `json:"activeUsers7Days"`
	ActiveSessions     int     `json:"activeSessions"`
	Users              int64   `json:"users"`
	MfaUsers           int64   `json:"mfaUsers"`
	McpBlockedToday    int     `json:"mcpBlockedToday"`
	Truncated          bool    `json:"truncated"`

	Hours []*SecurityHour `json:"hours"`
	Days  []*SecurityDay  `json:"days"`

	IpDbLoaded bool              `json:"ipDbLoaded"`
	Provinces  []*SecurityRegion `json:"provinces"`
	Cities     []*SecurityRegion `json:"cities"`

	TopFailedIps    []*SecurityCount `json:"topFailedIps"`
	TopFailedUsers  []*SecurityCount `json:"topFailedUsers"`
	TopApplications []*SecurityCount `json:"topApplications"`

	Events []*SecurityEvent `json:"events"`
	Alerts []*SecurityAlert `json:"alerts"`

	Security *DashboardSecurity `json:"security"`
}

const (
	securityScreenDays     = 30
	securityScreenEvents   = 40
	securityScreenCacheTTL = 20 * time.Second
)

var (
	securityScreenCache    = map[string]*SecurityScreen{}
	securityScreenCachedAt = map[string]time.Time{}
	securityScreenLock     sync.Mutex
	recordApplicationRegex = regexp.MustCompile(`"application"\s*:\s*"([^"]+)"`)
)

// GetSecurityScreen builds the security situation screen of an organization from its records, sign-in traces,
// sessions, users and certs. A wall display refreshes it often, so the result is kept for a few seconds.
func GetSecurityScreen(owner string) (*SecurityScreen, error) {
	securityScreenLock.Lock()
	defer securityScreenLock.Unlock()
	if res := securityScreenCache[owner]; res != nil && time.Since(securityScreenCachedAt[owner]) < securityScreenCacheTTL {
		return res, nil
	}

	res, err := buildSecurityScreen(owner)
	if err != nil {
		return nil, err
	}
	securityScreenCache[owner] = res
	securityScreenCachedAt[owner] = time.Now()
	return res, nil
}

// newIpLocator names the places of IPs, first after the sign-in traces, which were looked up when the user signed
// in, then after the IP database or the network
func newIpLocator(traces []*SigninTrace) func(string) string {
	ipLocations := map[string]string{}
	for _, trace := range traces {
		if trace.Kind == "location" && trace.Ip != "" {
			ipLocations[trace.Ip] = trace.TraceKey
		}
	}
	return func(clientIp string) string {
		location, ok := ipLocations[clientIp]
		if !ok {
			location = getIpLocation(clientIp)
			ipLocations[clientIp] = location
		}
		return location
	}
}

// splitLocation turns a location of getIpLocation into its province (or country abroad) and its city
func splitLocation(location string) (string, string, string) {
	if location == "" {
		return "unknown", "unknown", "unknown"
	}
	if location == "intranet" {
		return "intranet", "intranet", "intranet"
	}
	if strings.Contains(location, "/") {
		return location, location, "network"
	}

	parts := strings.Fields(location)
	country := parts[0]
	province, city := country, country
	isChina := country == "中国" || country == "China"
	if len(parts) >= 2 && isChina {
		province = parts[1]
		city = parts[1]
	} else if len(parts) >= 2 {
		province = country
		city = country + " " + parts[1]
	}
	if len(parts) >= 3 && parts[2] != parts[1] {
		city = city + " " + parts[2]
		if !isChina {
			city = country + " " + parts[2]
		}
	}
	return province, city, "place"
}

func rankRegions(regions map[string]*SecurityRegion, n int) []*SecurityRegion {
	res := []*SecurityRegion{}
	for _, region := range regions {
		res = append(res, region)
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Count != res[j].Count {
			return res[i].Count > res[j].Count
		}
		return res[i].Name < res[j].Name
	})
	if len(res) > n {
		res = res[:n]
	}
	return res
}

func buildSecurityScreen(owner string) (*SecurityScreen, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	from := today.AddDate(0, 0, 1-securityScreenDays)
	weekStart := today.AddDate(0, 0, -6)
	hourStart := now.Truncate(time.Hour).Add(-23 * time.Hour)

	res := &SecurityScreen{
		Organization:    owner,
		DisplayName:     owner,
		GeneratedTime:   util.GetCurrentTime(),
		Hours:           []*SecurityHour{},
		Days:            []*SecurityDay{},
		Provinces:       []*SecurityRegion{},
		Cities:          []*SecurityRegion{},
		TopApplications: []*SecurityCount{},
		Events:          []*SecurityEvent{},
		Alerts:          []*SecurityAlert{},
	}
	organization, err := getOrganization("admin", owner)
	if err != nil {
		return nil, err
	}
	if organization != nil && organization.DisplayName != "" {
		res.DisplayName = organization.DisplayName
	}

	dayMap := map[string]*SecurityDay{}
	for day := from; !day.After(today); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		dayMap[date] = &SecurityDay{Date: date}
		res.Days = append(res.Days, dayMap[date])
	}
	for i := 0; i < 24; i++ {
		res.Hours = append(res.Hours, &SecurityHour{Hour: hourStart.Add(time.Duration(i) * time.Hour).Format("15:00")})
	}

	// the places that the sign-in traces already named, they were looked up when the user signed in
	traces := []*SigninTrace{}
	err = ormer.Engine.Where("owner = ?", owner).Asc("created_time").Find(&traces)
	if err != nil {
		return nil, err
	}
	locate := newIpLocator(traces)

	provinces, cities := map[string]*SecurityRegion{}, map[string]*SecurityRegion{}
	failedIps, failedUsers := map[string]int{}, map[string]int{}
	usersToday, usersWeek := map[string]bool{}, map[string]bool{}
	events := []*SecurityEvent{}
	eventSince := now.AddDate(0, 0, -7).Format(time.RFC3339)

	requests, lastId := 0, 0
	for requests < securityOverviewMaxRecords {
		records := []*Record{}
		err = ormer.Engine.Cols("id", "created_time", "client_ip", "user", "method", "action", "response", "status_code", "detail").
			Where("owner = ? and created_time >= ? and id > ?", owner, from.Format(time.RFC3339), lastId).
			Asc("id").Limit(5000).Find(&records)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			lastId = record.Id
			requests++
			t, err := time.Parse(time.RFC3339, record.CreatedTime)
			if err != nil {
				continue
			}
			t = t.In(now.Location())
			isToday := !t.Before(today)
			isError := record.StatusCode >= 400 || strings.HasPrefix(record.Response, `{status:"error"`)
			isRecent := record.CreatedTime >= eventSince
			addEvent := func(kind string, detail string) {
				if isRecent {
					events = append(events, &SecurityEvent{Time: record.CreatedTime, Kind: kind, User: record.User, Ip: record.ClientIp, Detail: detail})
				}
			}

			if signinActions[record.Action] {
				day := dayMap[t.Format("2006-01-02")]
				var hour *SecurityHour
				if !t.Before(hourStart) {
					if i := int(t.Sub(hourStart) / time.Hour); i >= 0 && i < 24 {
						hour = res.Hours[i]
					}
				}

				province, city, kind := splitLocation(locate(record.ClientIp))
				for _, item := range []struct {
					regions map[string]*SecurityRegion
					name    string
				}{{provinces, province}, {cities, city}} {
					region := item.regions[item.name]
					if region == nil {
						region = &SecurityRegion{Name: item.name, Kind: kind}
						item.regions[item.name] = region
					}
					region.Count++
					if isError {
						region.Failure++
					}
				}

				if isError {
					if day != nil {
						day.Failure++
					}
					if hour != nil {
						hour.Failure++
					}
					if !t.Before(weekStart) {
						failedIps[record.ClientIp]++
						failedUsers[record.User]++
					}
					if isToday {
						res.SigninFailureToday++
						if record.Detail == SigninReasonAccountFrozen || record.Detail == SigninReasonAccountDisabled {
							res.BlockedToday++
						}
					}
					if record.Detail == SigninReasonAccountFrozen {
						addEvent("blocked-signin", record.Detail)
					} else {
						addEvent("failed-signin", record.Detail)
					}
				} else {
					if day != nil {
						day.Success++
					}
					if hour != nil {
						hour.Success++
					}
					if isToday {
						res.SigninSuccessToday++
						usersToday[record.User] = true
					}
					if !t.Before(weekStart) {
						usersWeek[record.User] = true
					}
				}
				continue
			}

			if record.StatusCode == 403 || strings.Contains(record.Response, "Unauthorized operation") {
				addEvent("denied", record.Action)
				continue
			}
			if record.Action == "signup" && !isError {
				if day := dayMap[t.Format("2006-01-02")]; day != nil {
					day.Signups++
				}
				continue
			}
			if record.Method != "GET" && !nonAdminActions[record.Action] && !isError && record.User != "" {
				addEvent("admin-operation", record.Action)
			}
		}
	}
	res.Truncated = requests >= securityOverviewMaxRecords
	delete(usersToday, "")
	delete(usersWeek, "")
	res.ActiveUsersToday = len(usersToday)
	res.ActiveUsers7Days = len(usersWeek)
	if total := res.SigninSuccessToday + res.SigninFailureToday; total > 0 {
		res.FailureRateToday = float64(res.SigninFailureToday) / float64(total)
	}
	res.IpDbLoaded = ipDbLoaded
	res.Provinces = rankRegions(provinces, 12)
	res.Cities = rankRegions(cities, 12)
	res.TopFailedIps = topCounts(failedIps, 8)
	res.TopFailedUsers = topCounts(failedUsers, 8)

	res.TopApplications, err = getTopSigninApplications(owner, weekStart)
	if err != nil {
		return nil, err
	}

	// a user's first place and first device are where they always sign in, every later one is new
	seen := map[string]bool{}
	for _, trace := range traces {
		key := trace.UserName + "/" + trace.Kind
		isFirst := !seen[key]
		seen[key] = true
		if isFirst || trace.CreatedTime < eventSince {
			continue
		}
		if trace.Kind == "location" {
			events = append(events, &SecurityEvent{Time: trace.CreatedTime, Kind: "new-location", User: trace.UserName, Ip: trace.Ip, Detail: trace.TraceKey})
		} else {
			events = append(events, &SecurityEvent{Time: trace.CreatedTime, Kind: "new-device", User: trace.UserName, Ip: trace.Ip, Detail: trace.UserAgent})
		}
	}

	mcpRecords := []*McpRecord{}
	err = ormer.Engine.Cols("created_time", "user_name", "client_ip", "server", "tool").
		Where("owner = ? and result = ? and created_time >= ?", owner, "Blocked", eventSince).
		Desc("id").Limit(200).Find(&mcpRecords)
	if err != nil {
		return nil, err
	}
	for _, record := range mcpRecords {
		events = append(events, &SecurityEvent{Time: record.CreatedTime, Kind: "mcp-blocked", User: record.UserName, Ip: record.ClientIp, Detail: record.Server + " / " + record.Tool})
		if record.CreatedTime >= today.Format(time.RFC3339) {
			res.McpBlockedToday++
		}
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Time > events[j].Time
	})
	if len(events) > securityScreenEvents {
		events = events[:securityScreenEvents]
	}
	res.Events = events

	res.ActiveSessions, err = countActiveSessions(owner, now)
	if err != nil {
		return nil, err
	}
	mfa, err := GetDashboardMfaCoverage(owner)
	if err != nil {
		return nil, err
	}
	res.Users, res.MfaUsers = mfa.Total, mfa.Enabled

	res.Security, err = GetDashboardSecurity(owner)
	if err != nil {
		return nil, err
	}
	res.Alerts, err = getSecurityAlerts(owner, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// getTopSigninApplications counts the successful sign-ins to each application since the given time
func getTopSigninApplications(owner string, since time.Time) ([]*SecurityCount, error) {
	counts := map[string]int{}
	lastId := 0
	for read := 0; read < 50000; {
		records := []*Record{}
		err := ormer.Engine.Cols("id", "object", "response", "status_code").
			Where("owner = ? and action = ? and created_time >= ? and id > ?", owner, "login", since.Format(time.RFC3339), lastId).
			Asc("id").Limit(2000).Find(&records)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			break
		}
		for _, record := range records {
			lastId = record.Id
			read++
			if record.StatusCode >= 400 || strings.HasPrefix(record.Response, `{status:"error"`) {
				continue
			}
			if match := recordApplicationRegex.FindStringSubmatch(record.Object); match != nil {
				counts[match[1]]++
			}
		}
	}
	return topCounts(counts, 8), nil
}

func countActiveSessions(owner string, now time.Time) (int, error) {
	sessions := []*Session{}
	err := ormer.Engine.Where("owner = ?", owner).Find(&sessions)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, session := range sessions {
		if len(session.SessionInfos) == 0 {
			count += len(session.SessionId)
			continue
		}
		for _, info := range session.SessionInfos {
			expireTime, err := time.Parse(time.RFC3339, info.ExpireTime)
			if info.ExpireTime == "" || (err == nil && expireTime.After(now)) {
				count++
			}
		}
	}
	return count, nil
}

func getSecurityAlerts(owner string, screen *SecurityScreen) ([]*SecurityAlert, error) {
	alerts := []*SecurityAlert{}
	add := func(kind string, level string, count int, detail string) {
		if count > 0 {
			alerts = append(alerts, &SecurityAlert{Kind: kind, Level: level, Count: count, Detail: detail})
		}
	}

	blockedIps := GetCredentialStuffingBlockedIps(owner)
	add("ip-blocked", "critical", len(blockedIps), strings.Join(blockedIps, ", "))
	add("accounts-frozen", "warning", screen.BlockedToday, "")
	if screen.SigninFailureToday >= 10 && screen.FailureRateToday >= 0.3 {
		add("high-failure-rate", "critical", screen.SigninFailureToday, "")
	}

	security := screen.Security
	add("admins-without-mfa", "critical", int(security.AdminsWithoutMfa), "")
	certs, err := getExpiringCertNames(owner, time.Now(), dashboardCertAlertDays)
	if err != nil {
		return nil, err
	}
	add("cert-expiring", "warning", len(certs), strings.Join(certs, ", "))

	orphans, err := GetOrphanAccounts(owner, dashboardInactiveDays)
	if err != nil {
		return nil, err
	}
	add("orphan-accounts", "warning", len(orphans), "")

	conflicts, err := GetSodConflicts(owner)
	if err != nil {
		return nil, err
	}
	add("sod-conflicts", "warning", len(conflicts), "")
	add("mcp-blocked", "warning", screen.McpBlockedToday, "")
	sort.SliceStable(alerts, func(i, j int) bool {
		return alerts[i].Level == "critical" && alerts[j].Level != "critical"
	})
	return alerts, nil
}

// getExpiringCertNames lists the certs of the organization that expire within the given days
func getExpiringCertNames(owner string, now time.Time, days int) ([]string, error) {
	certs := []*Cert{}
	err := ormer.Engine.Cols("owner", "name", "expire_time", "certificate").Where("owner = ?", owner).Find(&certs)
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, cert := range certs {
		expire := cert.ExpireTime
		if expire == "" && cert.Certificate != "" {
			expire, _ = util.GetCertExpireTime(cert.Certificate)
		}
		expireTime, err := time.Parse(time.RFC3339, expire)
		if err != nil {
			continue
		}
		if expireTime.Sub(now) < time.Duration(days)*24*time.Hour {
			res = append(res, cert.Name)
		}
	}
	return res, nil
}
