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
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/conf"
)

type DashboardDateItem struct {
	CreatedTime string `json:"createTime"`
}

type DashboardLoginHeatmap struct {
	XAxis []int     `json:"xAxis"`
	YAxis []string  `json:"yAxis"`
	Data  [][]int64 `json:"data"`
	Max   int64     `json:"max"`
}

type DashboardResourceByProviderItem struct {
	Provider string `json:"provider" xorm:"provider"`
	Count    int64  `json:"count" xorm:"count"`
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
			defer wg.Done()
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

func GetDashboardUsersByProvider(owner string) (*map[string]int64, error) {
	if owner == "All" {
		owner = ""
	}

	allowColumns := getUserProviderColumns()
	var providers []*Provider
	var err error
	if owner == "" {
		providers, err = GetGlobalProviders()
	} else {
		providers, err = GetProviders(owner)
	}
	if err != nil {
		return nil, err
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	userTable := tableNamePrefix + "user"

	// Filter unique provider types
	seenTypes := map[string]struct{}{}
	var uniqueProviders []*Provider
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		if provider.Category != "OAuth" {
			continue
		}
		if provider.Type == "" {
			continue
		}
		if _, ok := seenTypes[provider.Type]; ok {
			continue
		}
		seenTypes[provider.Type] = struct{}{}
		uniqueProviders = append(uniqueProviders, provider)
	}

	// Use goroutines for parallel database queries
	var wg sync.WaitGroup
	var resMap sync.Map
	ch := make(chan error, len(uniqueProviders))

	wg.Add(len(uniqueProviders))
	for _, provider := range uniqueProviders {
		go func(p *Provider, ch chan error) {
			defer wg.Done()

			column := normalizeProviderColumn(p.Type)
			dbQuery := ormer.Engine.Table(userTable)
			if owner != "" {
				dbQuery = dbQuery.And("owner = ?", owner)
			}
			dbQuery = dbQuery.And("is_deleted <> ?", 1)

			if _, ok := allowColumns[column]; ok && isSafeIdentifier(column) {
				dbQuery = dbQuery.And(fmt.Sprintf("%s <> ''", column))
			} else {
				dbQuery = dbQuery.And("properties like ?", fmt.Sprintf("%%oauth_%s_%%", p.Type))
			}

			cnt, err := dbQuery.Count()
			if err != nil {
				ch <- err
				return
			}
			resMap.Store(p.Type, cnt)
		}(provider, ch)
	}

	wg.Wait()
	close(ch)

	// Check for errors
	for err = range ch {
		if err != nil {
			return nil, err
		}
	}

	// Convert sync.Map to regular map
	res := map[string]int64{}
	resMap.Range(func(key, value interface{}) bool {
		res[key.(string)] = value.(int64)
		return true
	})

	return &res, nil
}

func GetDashboardLoginHeatmap(owner string) (*DashboardLoginHeatmap, error) {
	if owner == "All" {
		owner = ""
	}

	type recordCreatedTimeItem struct {
		CreatedTime string `xorm:"created_time"`
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	recordTable := tableNamePrefix + "record"

	time7day := time.Now().AddDate(0, 0, -7).Format(time.RFC3339)

	dbQuery := ormer.Engine.Table(recordTable).Cols("created_time").
		And("action = ?", "login").
		And("created_time >= ?", time7day)
	if owner != "" {
		dbQuery = dbQuery.And("owner = ?", owner)
	}

	nowLocal := time.Now().Local()
	// Past 7 days (oldest -> newest)
	yAxis := make([]string, 7)
	yIndex := map[string]int{}
	for i := 6; i >= 0; i-- {
		dayTime := nowLocal.AddDate(0, 0, -i)
		dateKey := dayTime.Format("2006-01-02")
		dateStr := dayTime.Format("1-2")
		row := 6 - i
		yAxis[row] = dateStr
		yIndex[dateKey] = row
	}

	xAxis := make([]int, 24)
	for i := 0; i < 24; i++ {
		xAxis[i] = i
	}

	counts := make([][]int64, 7)
	for i := 0; i < 7; i++ {
		counts[i] = make([]int64, 24)
	}

	rows, err := dbQuery.Rows(&recordCreatedTimeItem{})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := recordCreatedTimeItem{}
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}

		t, err := time.Parse(time.RFC3339, item.CreatedTime)
		if err != nil {
			continue
		}

		localTime := t.Local()
		hour := localTime.Hour()

		row, ok := yIndex[localTime.Format("2006-01-02")]
		if !ok {
			continue
		}
		counts[row][hour]++
	}

	var max int64
	data := [][]int64{}
	for y := 0; y < 7; y++ {
		for x := 0; x < 24; x++ {
			v := counts[y][x]
			if v > max {
				max = v
			}
			if v == 0 {
				continue
			}
			data = append(data, []int64{int64(x), int64(y), v})
		}
	}

	return &DashboardLoginHeatmap{
		XAxis: xAxis,
		YAxis: yAxis,
		Data:  data,
		Max:   max,
	}, nil
}

func GetDashboardResourcesByProvider(owner string) (*[]DashboardResourceByProviderItem, error) {
	if owner == "All" {
		owner = ""
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	resourceTable := tableNamePrefix + "resource"

	results := []DashboardResourceByProviderItem{}

	dbQuery := ormer.Engine.Table(resourceTable).
		Select("provider, COUNT(*) as count").
		GroupBy("provider")

	if owner != "" {
		dbQuery = dbQuery.Where("owner = ?", owner)
	}

	if err := dbQuery.Find(&results); err != nil {
		return nil, err
	}

	// Sort by count DESC
	sort.Slice(results, func(i, j int) bool {
		return results[i].Count > results[j].Count
	})

	return &results, nil
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

func normalizeProviderColumn(providerType string) string {
	column := strings.ToLower(providerType)
	column = strings.ReplaceAll(column, " ", "")
	column = strings.ReplaceAll(column, "-", "")
	column = strings.ReplaceAll(column, "_", "")
	return column
}

func isSafeIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			continue
		}
		if c >= '0' && c <= '9' {
			continue
		}
		return false
	}
	return true
}

func getUserProviderColumns() map[string]struct{} {
	userType := reflect.TypeOf(User{})
	start := -1
	end := -1
	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)
		if field.Name == "GitHub" {
			start = i
		} else if field.Name == "Custom10" {
			end = i
		}
	}

	if start == -1 || end == -1 || end < start {
		return map[string]struct{}{}
	}

	res := map[string]struct{}{}
	for i := start; i <= end; i++ {
		field := userType.Field(i)
		if field.Type.Kind() != reflect.String {
			continue
		}

		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		res[jsonTag] = struct{}{}
	}
	return res
}
