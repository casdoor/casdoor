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
// for the admin and user dashboard, querying from the record (audit log) table
// as the source of truth for login events.
package object

import (
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

// LoginTrendItem represents the login count for a single day.
type LoginTrendItem struct {
	Date  string `json:"date"` // Date is formatted as "YYYY-MM-DD".
	Count int64  `json:"count"`
}

// RealTimeActivity holds aggregated counts for a short rolling window.
type RealTimeActivity struct {
	SuccessCount int64 `json:"successCount"`
	FailureCount int64 `json:"failureCount"`
}

// AdminDashboardAnalytics is the full analytics payload for the admin dashboard.
type AdminDashboardAnalytics struct {
	TotalUsers       int64            `json:"totalUsers"`
	WeeklyLoginTrend []LoginTrendItem `json:"weeklyLoginTrend"`
	TopApps          []AppUsage       `json:"topApps"`
	RealTimeActivity RealTimeActivity `json:"realTimeActivity"`
}

// UserDashboardAnalytics is the full analytics payload for a specific user's dashboard.
type UserDashboardAnalytics struct {
	TopApps         []AppUsage `json:"topApps"`
	ActivityHeatmap []int64    `json:"activityHeatmap"` // length 24, index = hour of day
}

// recordEntry is an internal struct for scanning rows from the record table.
type recordEntry struct {
	CreatedTime string `xorm:"created_time"`
	// Object in the record table stores the application name for login actions.
	Object    string `xorm:"object"`
	User      string `xorm:"user"`
	IsSuccess bool   `xorm:"is_success"`
}

// -------------------------------------------- Public Functions --------------------------------------------

// GetAdminDashboardAnalytics returns aggregated analytics for the admin dashboard.
// owner filters results to a specific organization; pass "" or "All" for global.
// topAppsLimit controls how many top apps to return (defaults to 5).
// topAppsPeriod controls the time window for top apps ("week" or "month").
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

	weeklyTrend, err := getWeeklyLoginTrend(owner)
	if err != nil {
		return nil, err
	}

	topApps, err := getTopApps(owner, topAppsLimit, topAppsPeriod)
	if err != nil {
		return nil, err
	}

	realTime, err := getRealTimeActivity(owner)
	if err != nil {
		return nil, err
	}

	return &AdminDashboardAnalytics{
		TotalUsers:       totalUsers,
		WeeklyLoginTrend: weeklyTrend,
		TopApps:          topApps,
		RealTimeActivity: *realTime,
	}, nil
}

// GetUserDashboardAnalytics returns analytics scoped to a single user.
// owner is the organization name; userId is the user's name field.
// topAppsPeriod controls the time window ("week" or "month").
func GetUserDashboardAnalytics(owner, userId string, topAppsPeriod Period) (*UserDashboardAnalytics, error) {
	if owner == "All" {
		owner = ""
	}

	topApps, err := getUserTopApps(owner, userId, 5, topAppsPeriod)
	if err != nil {
		return nil, err
	}

	heatmap, err := getUserActivityHeatmap(owner, userId)
	if err != nil {
		return nil, err
	}

	return &UserDashboardAnalytics{
		TopApps:         topApps,
		ActivityHeatmap: heatmap,
	}, nil
}

// -------------------------------------------- Private Helper Functions --------------------------------------------

// countTotalUsers returns the total number of users for the given owner.
func countTotalUsers(owner string) (int64, error) {
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	userTable := tableNamePrefix + "user"

	session := ormer.Engine.Table(userTable)
	if owner != "" {
		session = session.Where("owner = ?", owner)
	}

	count, err := session.Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// getWeeklyLoginTrend queries the record table and returns daily login counts
// for the past 7 days. Index 0 = 6 days ago, index 6 = today.
func getWeeklyLoginTrend(owner string) ([]LoginTrendItem, error) {
	records, err := fetchLoginRecords(owner, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, err
	}

	now := time.Now()
	// Build a map keyed by "YYYY-MM-DD" → count for quick lookup.
	dayCountMap := make(map[string]int64)
	for _, r := range records {
		t, parseErr := parseRecordTime(r.CreatedTime)
		if parseErr != nil {
			continue
		}
		key := t.Format("2006-01-02")
		dayCountMap[key]++
	}

	trend := make([]LoginTrendItem, 7)
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -(6 - i))
		key := day.Format("2006-01-02")
		trend[i] = LoginTrendItem{
			Date:  key,
			Count: dayCountMap[key],
		}
	}
	return trend, nil
}

// getTopApps returns the top N applications by login count within the given period.
func getTopApps(owner string, limit int, period Period) ([]AppUsage, error) {
	since := periodToTime(period)
	records, err := fetchLoginRecords(owner, since)
	if err != nil {
		return nil, err
	}
	return buildTopApps(records, limit), nil
}

// getUserTopApps returns top N apps for a specific user within the given period.
func getUserTopApps(owner, userId string, limit int, period Period) ([]AppUsage, error) {
	since := periodToTime(period)
	records, err := fetchUserLoginRecords(owner, userId, since)
	if err != nil {
		return nil, err
	}
	return buildTopApps(records, limit), nil
}

// getUserActivityHeatmap returns a 24-slot slice where each index is the login
// count for that hour of the day, aggregated over the past 30 days.
func getUserActivityHeatmap(owner, userId string) ([]int64, error) {
	since := time.Now().AddDate(0, 0, -30)
	records, err := fetchUserLoginRecords(owner, userId, since)
	if err != nil {
		return nil, err
	}

	heatmap := make([]int64, 24)
	for _, r := range records {
		t, err := parseRecordTime(r.CreatedTime)
		if err != nil {
			continue
		}
		heatmap[t.Hour()]++
	}
	return heatmap, nil
}

// getRealTimeActivity returns successful and failed login counts in the last 5 minutes.
func getRealTimeActivity(owner string) (*RealTimeActivity, error) {
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	recordTable := tableNamePrefix + "record"

	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	session := ormer.Engine.Table(recordTable).
		Where("action = ?", "login").
		And("created_time >= ?", fiveMinutesAgo)
	if owner != "" {
		session = session.And("owner = ?", owner)
	}

	var entries []recordEntry
	if err := session.Cols("is_success").Find(&entries); err != nil {
		return nil, err
	}

	var result RealTimeActivity
	for _, e := range entries {
		if e.IsSuccess {
			result.SuccessCount++
		} else {
			result.FailureCount++
		}
	}
	return &result, nil
}

// fetchLoginRecords retrieves all login records for an owner since a given time.
// It queries the record table where action = "login".
func fetchLoginRecords(owner string, since time.Time) ([]recordEntry, error) {
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	recordTable := tableNamePrefix + "record"

	session := ormer.Engine.Table(recordTable).
		Cols("created_time", "object", "user", "is_success").
		Where("action = ?", "login").
		And("created_time >= ?", since)
	if owner != "" {
		session = session.And("owner = ?", owner)
	}

	var entries []recordEntry
	if err := session.Find(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// fetchUserLoginRecords is like fetchLoginRecords but scoped to a single user.
func fetchUserLoginRecords(owner, userId string, since time.Time) ([]recordEntry, error) {
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	recordTable := tableNamePrefix + "record"

	session := ormer.Engine.Table(recordTable).
		Cols("created_time", "object", "user", "is_success").
		Where("action = ?", "login").
		And("created_time >= ?", since).
		And("user = ?", userId)
	if owner != "" {
		session = session.And("owner = ?", owner)
	}

	var entries []recordEntry
	if err := session.Find(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// buildTopApps takes a slice of records, counts per application, sorts descending,
// and returns the top N entries.
func buildTopApps(records []recordEntry, limit int) []AppUsage {
	appCountMap := make(map[string]int64)
	for _, r := range records {
		if r.Object != "" {
			appCountMap[r.Object]++
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

// periodToTime converts a Period constant to an absolute time.Time in the past.
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

// parseRecordTime parses timestamps stored by Casdoor in the record table.
// Casdoor stores times in "2006-01-02T15:04:05Z07:00" (RFC3339) or
// "2006-01-02 15:04:05" (MySQL DATETIME) format.
func parseRecordTime(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02 15:04:05", raw, time.Local)
}
