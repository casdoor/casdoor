// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

// Package object. dashboard_analytics provides analytics aggregation logic
// for the admin and user dashboard. It queries Casdoor's record (audit log)
// table as the source of truth for login events, parsing the JSON-encoded
// object field to extract application, organization, and username.
package object

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/casdoor/casdoor/conf"
)

// -------------------------------------------- Types --------------------------------------------

// Period represents a time window for filtering analytics data.
type Period string

const (
	PeriodDay   Period = "day"
	PeriodWeek  Period = "week"
	PeriodMonth Period = "month"
)

// LoginTrendItem represents the login count for a single calendar day.
type LoginTrendItem struct {
	// Date is formatted as "YYYY-MM-DD".
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// RealTimeActivity holds aggregated login counts over a short rolling window.
type RealTimeActivity struct{}

// AdminDashboardAnalytics is the full analytics payload for the admin dashboard.
type AdminDashboardAnalytics struct {
	TotalUsers       int64            `json:"totalUsers"`
	WeeklyLoginTrend []LoginTrendItem `json:"weeklyLoginTrend"`
	TopApps          []AppUsage       `json:"topApps"`
	RealTimeActivity RealTimeActivity `json:"realTimeActivity"`
}

// UserDashboardAnalytics is the full analytics payload for a specific user's dashboard.
type UserDashboardAnalytics struct {
	TopApps []AppUsage `json:"topApps"`
	// ActivityHeatmap has 24 slots; index = hour of day (0–23).
	ActivityHeatmap []int64 `json:"activityHeatmap"`
}

// recordEntry is an internal struct for scanning raw rows from the record table.
type recordEntry struct {
	CreatedTime string `xorm:"created_time"`
	// Object is a raw JSON string containing application, organization, username, etc.
	Object    string `xorm:"object"`
	IsSuccess bool   `xorm:"is_success"`
}

// loginObjectPayload mirrors the JSON structure stored in the record.object column.
// Example:
//
//	{"application":"app-built-in","organization":"built-in","username":"admin",...}
type loginObjectPayload struct {
	Application  string `json:"application"`
	Organization string `json:"organization"`
	Username     string `json:"username"`
}

// parsedRecord is a fully decoded login event, ready for aggregation.
type parsedRecord struct {
	CreatedTime  time.Time
	Application  string
	Organization string
	Username     string
}

// -------------------------------------------- Public Functions --------------------------------------------

// GetAdminDashboardAnalytics returns aggregated analytics for the admin dashboard.
// owner filters results to a specific organization; pass "" or "All" for global scope.
// topAppsLimit controls how many top apps to return (defaults to 5 if <= 0).
// topAppsPeriod controls the time window for top apps ("day", "week", "month").
func GetAdminDashboardAnalytics(owner string, topAppsLimit int, topAppsPeriod Period) (*AdminDashboardAnalytics, error) {
	if owner == "All" {
		owner = ""
	}
	if topAppsLimit <= 0 {
		topAppsLimit = 5
	}

	totalUsers, err := countTotalUsers(owner)
	if err != nil {
		return nil, err
	}

	// Fetch enough records to cover the widest window we need (30 days for monthly top apps).
	since := periodToTime(PeriodMonth)
	rawRecords, err := fetchRawLoginRecords(since)
	if err != nil {
		return nil, err
	}

	// Parse and decode every raw record once; filter by owner here too.
	parsed := parseAndFilterRecords(rawRecords, owner)

	weeklyTrend := buildWeeklyLoginTrend(parsed)
	topApps := buildTopApps(filterByPeriod(parsed, topAppsPeriod), topAppsLimit)
	realTime := buildRealTimeActivity(parsed)

	return &AdminDashboardAnalytics{
		TotalUsers:       totalUsers,
		WeeklyLoginTrend: weeklyTrend,
		TopApps:          topApps,
		RealTimeActivity: realTime,
	}, nil
}

// GetUserDashboardAnalytics returns analytics scoped to a single user.
// owner is the organization name; username is the user's login name (from the JSON payload).
// topAppsPeriod controls the time window ("day", "week", "month").
func GetUserDashboardAnalytics(owner, username string, topAppsPeriod Period) (*UserDashboardAnalytics, error) {
	if owner == "All" {
		owner = ""
	}

	// 30 days is enough for both heatmap and any supported period.
	since := periodToTime(PeriodMonth)
	rawRecords, err := fetchRawLoginRecords(since)
	if err != nil {
		return nil, err
	}

	parsed := parseAndFilterRecords(rawRecords, owner)
	userRecords := filterByUsername(parsed, username)

	topApps := buildTopApps(filterByPeriod(userRecords, topAppsPeriod), 5)
	heatmap := buildActivityHeatmap(userRecords)

	return &UserDashboardAnalytics{
		TopApps:         topApps,
		ActivityHeatmap: heatmap,
	}, nil
}

// -------------------------------------------- Private Helper Functions --------------------------------------------

// countTotalUsers returns the total number of users for the given owner organization.
func countTotalUsers(owner string) (int64, error) {
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	userTable := tableNamePrefix + "user"

	session := ormer.Engine.Table(userTable)
	if owner != "" {
		session = session.Where("owner = ?", owner)
	}
	return session.Count()
}

// fetchRawLoginRecords retrieves all raw login records from the record table
// created at or after `since`. No owner filtering is done at DB level because
// the owner/organization lives inside the JSON object column — we filter in Go.
func fetchRawLoginRecords(since time.Time) ([]recordEntry, error) {
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	recordTable := tableNamePrefix + "record"

	var entries []recordEntry
	err := ormer.Engine.Table(recordTable).
		Cols("created_time", "object").
		Where("action = ?", "login").
		And("created_time >= ?", since).
		Find(&entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// parseAndFilterRecords decodes the raw record rows into parsedRecord values.
// If owner is non-empty, only records whose JSON organization matches are kept.
func parseAndFilterRecords(raw []recordEntry, owner string) []parsedRecord {
	result := make([]parsedRecord, 0, len(raw))

	for _, r := range raw {
		t, err := parseRecordTime(r.CreatedTime)
		if err != nil {
			continue
		}

		payload, err := parseObjectJSON(r.Object)
		if err != nil {
			continue
		}

		// Owner / organization filter: the record table's own `owner` column is
		// empty for login events; the real org lives inside the JSON payload.
		if owner != "" && payload.Organization != owner {
			continue
		}

		result = append(result, parsedRecord{
			CreatedTime:  t,
			Application:  payload.Application,
			Organization: payload.Organization,
			Username:     payload.Username,
		})
	}
	return result
}

// filterByPeriod returns only the records that fall within the given period window.
func filterByPeriod(records []parsedRecord, period Period) []parsedRecord {
	since := periodToTime(period)
	filtered := make([]parsedRecord, 0, len(records))
	for _, r := range records {
		if !r.CreatedTime.Before(since) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// filterByUsername returns only the records belonging to the given username.
func filterByUsername(records []parsedRecord, username string) []parsedRecord {
	filtered := make([]parsedRecord, 0)
	for _, r := range records {
		if r.Username == username {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// buildWeeklyLoginTrend builds a 7-slot trend from the past 7 days of records.
// Index 0 = 6 days ago, index 6 = today.
func buildWeeklyLoginTrend(records []parsedRecord) []LoginTrendItem {
	weekAgo := periodToTime(PeriodWeek)

	dayCountMap := make(map[string]int64)
	for _, r := range records {
		if r.CreatedTime.Before(weekAgo) {
			continue
		}
		key := r.CreatedTime.Format("2006-01-02")
		dayCountMap[key]++
	}

	now := time.Now()
	trend := make([]LoginTrendItem, 7)
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -(6 - i))
		key := day.Format("2006-01-02")
		trend[i] = LoginTrendItem{
			Date:  key,
			Count: dayCountMap[key],
		}
	}
	return trend
}

// buildTopApps counts logins per application, sorts descending, and returns top N.
func buildTopApps(records []parsedRecord, limit int) []AppUsage {
	appCountMap := make(map[string]int64)
	for _, r := range records {
		if r.Application != "" {
			appCountMap[r.Application]++
		}
	}

	topApps := make([]AppUsage, 0, len(appCountMap))
	for appName, count := range appCountMap {
		topApps = append(topApps, AppUsage{AppName: appName, Count: count})
	}
	sort.Slice(topApps, func(i, j int) bool {
		return topApps[i].Count > topApps[j].Count
	})

	if len(topApps) > limit {
		topApps = topApps[:limit]
	}
	return topApps
}

// buildActivityHeatmap returns a 24-slot slice where index = hour of day,
// value = number of logins during that hour.
func buildActivityHeatmap(records []parsedRecord) []int64 {
	heatmap := make([]int64, 24)
	for _, r := range records {
		heatmap[r.CreatedTime.Hour()]++
	}
	return heatmap
}

// buildRealTimeActivity ...
func buildRealTimeActivity(_ []parsedRecord) RealTimeActivity {
	//TODO: maybe use websocket here?
	return RealTimeActivity{}
}

// parseObjectJSON decodes the JSON string stored in the record's object column.
func parseObjectJSON(raw string) (loginObjectPayload, error) {
	var payload loginObjectPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return loginObjectPayload{}, err
	}
	return payload, nil
}

// periodToTime converts a Period to an absolute time.Time boundary in the past.
func periodToTime(period Period) time.Time {
	now := time.Now()
	switch period {
	case PeriodDay:
		return now.AddDate(0, 0, -1)
	case PeriodMonth:
		return now.AddDate(0, -1, 0)
	case PeriodWeek:
		fallthrough
	default:
		return now.AddDate(0, 0, -7)
	}
}

// parseRecordTime handles both RFC3339 ("2006-01-02T15:04:05Z07:00") and
// MySQL DATETIME ("2006-01-02 15:04:05") formats.
func parseRecordTime(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02 15:04:05", raw, time.Local)
}
