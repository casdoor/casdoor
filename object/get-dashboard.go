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

package object

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/casdoor/casdoor/conf"
)

type AppUsage struct {
	AppName string `json:"appName"`
	Count   int64  `json:"count"`
}

type DashboardAnalytics struct {
	WeeklyLoginCounts  []int64    `json:"weeklyLoginCounts"`
	TopApps            []AppUsage `json:"topApps"`
	UserTopApps        []AppUsage `json:"userTopApps"`
	UserActivityHeatmap []int64   `json:"userActivityHeatmap"`
}

type DashboardDateItem struct {
	CreatedTime string `json:"createTime"`
}

type DashboardMapItem struct {
	dashboardDateItems []DashboardDateItem
	itemCount          int64
}

func GetDashboard(owner string) (*map[string][]int64, error) {
	if owner == "All" {
		owner = ""
	}

	dashboard := make(map[string][]int64)
	dashboardMap := sync.Map{}
	tableNames := []string{"organization", "user", "provider", "application", "subscription", "role", "group", "resource", "cert", "permission", "transaction", "model", "adapter", "enforcer"}

	time30day := time.Now().AddDate(0, 0, -30)
	var wg sync.WaitGroup
	var err error
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	wg.Add(len(tableNames))
	ch := make(chan error, len(tableNames))
	for _, tableName := range tableNames {
		dashboard[tableName+"Counts"] = make([]int64, 31)
		tableFullName := tableNamePrefix + tableName
		go func(ch chan error) {
			defer func() {
				if r := recover(); r != nil {
					ch <- fmt.Errorf("panic in dashboard goroutine: %v", r)
				}
				wg.Done()
			}()
			dashboardDateItems := []DashboardDateItem{}
			var countResult int64

			dbQueryBefore := ormer.Engine.Cols("created_time")
			dbQueryAfter := ormer.Engine.Cols("created_time")

			if owner != "" {
				dbQueryAfter = dbQueryAfter.And("owner = ?", owner)
				dbQueryBefore = dbQueryBefore.And("owner = ?", owner)
			}

			if countResult, err = dbQueryBefore.And("created_time < ?", time30day).Table(tableFullName).Count(); err != nil {
				ch <- err
				return
			}
			if err = dbQueryAfter.And("created_time >= ?", time30day).Table(tableFullName).Find(&dashboardDateItems); err != nil {
				ch <- err
				return
			}

			dashboardMap.Store(tableFullName, DashboardMapItem{
				dashboardDateItems: dashboardDateItems,
				itemCount:          countResult,
			})
		}(ch)
	}

	wg.Wait()
	close(ch)

	for err = range ch {
		if err != nil {
			return nil, err
		}
	}

	nowTime := time.Now()
	for i := 30; i >= 0; i-- {
		cutTime := nowTime.AddDate(0, 0, -i)
		for _, tableName := range tableNames {
			item, exist := dashboardMap.Load(tableName)
			if !exist {
				continue
			}
			dashboard[tableName+"Counts"][30-i] = countCreatedBefore(item.(DashboardMapItem), cutTime)
		}
	}
	return &dashboard, nil
}

func countCreatedBefore(dashboardMapItem DashboardMapItem, before time.Time) int64 {
	count := dashboardMapItem.itemCount
	for _, e := range dashboardMapItem.dashboardDateItems {
		createdTime, _ := time.Parse(time.RFC3339, e.CreatedTime)
		if createdTime.Before(before) {
			count++
		}
	}
	return count
}

type tokenCreatedTime struct {
	CreatedTime string `xorm:"created_time"`
	Application string `xorm:"application"`
	User        string `xorm:"user"`
}

func GetDashboardAnalytics(owner, userId string) (*DashboardAnalytics, error) {
	if owner == "All" {
		owner = ""
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tokenTable := tableNamePrefix + "token"

	now := time.Now()
	time7day := now.AddDate(0, 0, -7)
	time30day := now.AddDate(0, 0, -30)

	// Query tokens from last 30 days for app usage and last 7 for login trend
	var tokens []tokenCreatedTime
	session := ormer.Engine.Table(tokenTable).
		Cols("created_time", "application", "user").
		Where("created_time >= ?", time30day)
	if owner != "" {
		session = session.And("owner = ?", owner)
	}
	if err := session.Find(&tokens); err != nil {
		return nil, err
	}

	// Weekly login counts: index 0 = 6 days ago, index 6 = today
	weeklyLoginCounts := make([]int64, 7)
	// Top apps: count per application for last 30 days
	appCountMap := make(map[string]int64)
	// User-specific stats
	userAppCountMap := make(map[string]int64)
	userActivityHeatmap := make([]int64, 24)

	for _, t := range tokens {
		createdTime, err := time.Parse(time.RFC3339, t.CreatedTime)
		if err != nil {
			continue
		}

		// Weekly login counts (last 7 days)
		if !createdTime.Before(time7day) {
			daysDiff := int(now.Truncate(24*time.Hour).Sub(createdTime.Truncate(24*time.Hour)).Hours() / 24)
			if daysDiff >= 0 && daysDiff < 7 {
				weeklyLoginCounts[6-daysDiff]++
			}
		}

		// Top apps (last 30 days)
		if t.Application != "" {
			appCountMap[t.Application]++
		}

		// User-specific stats
		if userId != "" && t.User == userId {
			if t.Application != "" {
				userAppCountMap[t.Application]++
			}
			userActivityHeatmap[createdTime.Hour()]++
		}
	}

	// Build top 5 apps slice sorted by count
	topApps := make([]AppUsage, 0, len(appCountMap))
	for appName, count := range appCountMap {
		topApps = append(topApps, AppUsage{AppName: appName, Count: count})
	}
	sort.Slice(topApps, func(i, j int) bool {
		return topApps[i].Count > topApps[j].Count
	})
	if len(topApps) > 5 {
		topApps = topApps[:5]
	}

	// Build user top apps slice
	userTopApps := make([]AppUsage, 0, len(userAppCountMap))
	for appName, count := range userAppCountMap {
		userTopApps = append(userTopApps, AppUsage{AppName: appName, Count: count})
	}
	sort.Slice(userTopApps, func(i, j int) bool {
		return userTopApps[i].Count > userTopApps[j].Count
	})
	if len(userTopApps) > 5 {
		userTopApps = userTopApps[:5]
	}

	return &DashboardAnalytics{
		WeeklyLoginCounts:   weeklyLoginCounts,
		TopApps:             topApps,
		UserTopApps:         userTopApps,
		UserActivityHeatmap: userActivityHeatmap,
	}, nil
}
