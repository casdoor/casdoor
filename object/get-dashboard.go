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
	"sync"
	"time"
)

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
	wg.Add(len(tableNames))

	for _, tableName := range tableNames {
		dashboard[tableName+"Counts"] = make([]int64, 31)
		tableName := tableName
		go func() {
			defer wg.Done()
			var dashboardDateItems []DashboardDateItem
			var countResult int64
			if owner == "" {
				if countResult, err = ormer.Engine.Table(tableName).Where("created_time < ?", time30day).Count(); err != nil {
					panic(err)
				}
				if err := ormer.Engine.Table(tableName).Select("created_time").Where("created_time >= ?", time30day).Find(&dashboardDateItems); err != nil {
					panic(err)
				}
			} else {
				if countResult, err = ormer.Engine.Table(tableName).Where("created_time < ? and owner = ?", time30day, owner).Count(); err != nil {
					panic(err)
				}
				if err := ormer.Engine.Table(tableName).Select("created_time").Where("created_time >= ? and owner = ?", time30day, owner).Find(&dashboardDateItems); err != nil {
					panic(err)
				}
			}

			dashboardMap.Store(tableName, DashboardMapItem{
				dashboardDateItems: dashboardDateItems,
				itemCount:          countResult,
			})
		}()
	}

	wg.Wait()

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
		createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", e.CreatedTime)
		if createdTime.Before(before) {
			count++
		}
	}
	return count
}
