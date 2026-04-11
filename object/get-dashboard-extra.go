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
	"sort"
	"time"

	"github.com/casdoor/casdoor/conf"
)

// ProviderTypeCount holds a provider type label and its count.
type ProviderTypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// MfaCoverageData holds the MFA adoption statistics for users.
type MfaCoverageData struct {
	Enabled  int64 `json:"enabled"`
	Disabled int64 `json:"disabled"`
	Total    int64 `json:"total"`
}

// DayCount holds a date and the number of login events on that day.
type DayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// LoginHeatmapData holds login activity aggregated by calendar day over the past year.
type LoginHeatmapData struct {
	Data      []DayCount `json:"data"`
	MaxCount  int        `json:"maxCount"`
	DateRange [2]string  `json:"dateRange"`
}

// GetDashboardProviderDistribution returns the number of providers grouped by type.
func GetDashboardProviderDistribution(owner string) ([]*ProviderTypeCount, error) {
	if owner == "All" {
		owner = ""
	}
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tableName := tableNamePrefix + "provider"

	var providers []*Provider
	query := ormer.Engine.Table(tableName)
	if owner != "" {
		query = query.Where("owner = ?", owner)
	}
	if err := query.Find(&providers); err != nil {
		return nil, err
	}

	typeCounts := make(map[string]int64)
	for _, p := range providers {
		t := p.Type
		if t == "" {
			t = "Unknown"
		}
		typeCounts[t]++
	}

	result := make([]*ProviderTypeCount, 0, len(typeCounts))
	for t, count := range typeCounts {
		result = append(result, &ProviderTypeCount{Type: t, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result, nil
}

// GetDashboardMfaCoverage returns MFA adoption statistics for users in the given owner.
// A user is considered "enabled" when phone MFA, email MFA, or a preferred MFA type is set.
func GetDashboardMfaCoverage(owner string) (*MfaCoverageData, error) {
	if owner == "All" {
		owner = ""
	}
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tableName := tableNamePrefix + "user"

	totalQuery := ormer.Engine.Table(tableName)
	if owner != "" {
		totalQuery = totalQuery.Where("owner = ?", owner)
	}
	total, err := totalQuery.Count()
	if err != nil {
		return nil, err
	}

	// Group MFA conditions inside parentheses so the optional owner AND is appended correctly.
	mfaCond := "(mfa_phone_enabled = ? OR mfa_email_enabled = ? OR (preferred_mfa_type IS NOT NULL AND preferred_mfa_type != ''))"
	enabledQuery := ormer.Engine.Table(tableName).Where(mfaCond, true, true)
	if owner != "" {
		enabledQuery = enabledQuery.And("owner = ?", owner)
	}
	enabled, err := enabledQuery.Count()
	if err != nil {
		return nil, err
	}

	return &MfaCoverageData{
		Enabled:  enabled,
		Disabled: total - enabled,
		Total:    total,
	}, nil
}

// GetDashboardLoginHeatmap returns daily login counts over the past year from the record table,
// suitable for rendering a GitHub-style calendar heatmap.
// Returns empty data gracefully when the record table is unavailable.
func GetDashboardLoginHeatmap(owner string) (*LoginHeatmapData, error) {
	if owner == "All" {
		owner = ""
	}
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tableName := tableNamePrefix + "record"

	type recordItem struct {
		CreatedTime string `xorm:"created_time"`
	}

	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	dateRange := [2]string{oneYearAgo.Format("2006-01-02"), now.Format("2006-01-02")}

	var records []recordItem
	query := ormer.Engine.Table(tableName).Cols("created_time").
		Where("created_time >= ?", oneYearAgo).
		And("action = ?", "login")
	if owner != "" {
		query = query.And("organization = ?", owner)
	}
	if err := query.Find(&records); err != nil {
		return &LoginHeatmapData{Data: []DayCount{}, MaxCount: 0, DateRange: dateRange}, nil
	}

	dateCounts := make(map[string]int)
	for _, r := range records {
		t, err := time.Parse(time.RFC3339, r.CreatedTime)
		if err != nil {
			continue
		}
		dateCounts[t.Format("2006-01-02")]++
	}

	data := make([]DayCount, 0, len(dateCounts))
	maxCount := 0
	for date, count := range dateCounts {
		data = append(data, DayCount{Date: date, Count: count})
		if count > maxCount {
			maxCount = count
		}
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date < data[j].Date
	})

	return &LoginHeatmapData{Data: data, MaxCount: maxCount, DateRange: dateRange}, nil
}
