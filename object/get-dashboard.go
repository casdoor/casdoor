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

type Dashboard struct {
	OrganizationCounts []int `json:"organizationCounts"`
	UserCounts         []int `json:"userCounts"`
	ProviderCounts     []int `json:"providerCounts"`
	ApplicationCounts  []int `json:"applicationCounts"`
	SubscriptionCounts []int `json:"subscriptionCounts"`
	RoleCounts         []int `json:"roleCounts"`
	GroupCounts        []int `json:"groupCounts"`
	ResourceCounts     []int `json:"resourceCounts"`
	CertCounts         []int `json:"certCounts"`
	PermissionCounts   []int `json:"permissionCounts"`
	TransactionCounts  []int `json:"transactionCounts"`
	ModelCounts        []int `json:"modelCounts"`
	AdapterCounts      []int `json:"adapterCounts"`
	EnforcerCounts     []int `json:"enforcerCounts"`
}

type DashboardDateItem struct {
	CreatedTime string `json:"createTime"`
}

func GetDashboard(owner string) (*Dashboard, error) {
	if owner == "All" {
		owner = ""
	}

	dashboard := &Dashboard{
		OrganizationCounts: make([]int, 31),
		UserCounts:         make([]int, 31),
		ProviderCounts:     make([]int, 31),
		ApplicationCounts:  make([]int, 31),
		SubscriptionCounts: make([]int, 31),
		RoleCounts:         make([]int, 31),
		GroupCounts:        make([]int, 31),
		ResourceCounts:     make([]int, 31),
		CertCounts:         make([]int, 31),
		PermissionCounts:   make([]int, 31),
		TransactionCounts:  make([]int, 31),
		ModelCounts:        make([]int, 31),
		AdapterCounts:      make([]int, 31),
		EnforcerCounts:     make([]int, 31),
	}

	organizations := []DashboardDateItem{}
	users := []DashboardDateItem{}
	providers := []DashboardDateItem{}
	applications := []DashboardDateItem{}
	subscriptions := []DashboardDateItem{}
	roles := []DashboardDateItem{}
	groups := []DashboardDateItem{}
	resources := []DashboardDateItem{}
	certs := []DashboardDateItem{}
	permissions := []DashboardDateItem{}
	transactions := []DashboardDateItem{}
	models := []DashboardDateItem{}
	adapters := []DashboardDateItem{}
	enforcers := []DashboardDateItem{}

	var wg sync.WaitGroup
	wg.Add(14)
	go func() {
		defer wg.Done()
		if err := ormer.Engine.Table(&Organization{}).Select("created_time").Find(&organizations, &Organization{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()

		if err := ormer.Engine.Table(&User{}).Select("created_time").Find(&users, &User{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()

		if err := ormer.Engine.Table(&Provider{}).Select("created_time").Find(&providers, &Provider{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()

		if err := ormer.Engine.Table(&Application{}).Select("created_time").Find(&applications, &Application{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()

		if err := ormer.Engine.Table(&Subscription{}).Select("created_time").Find(&subscriptions, &Subscription{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()

		if err := ormer.Engine.Table(&Role{}).Select("created_time").Find(&roles, &Role{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()

		if err := ormer.Engine.Table(&Group{}).Find(&groups, &Group{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := ormer.Engine.Table(&Resource{}).Select("created_time").Find(&resources, &Resource{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := ormer.Engine.Table(&Cert{}).Select("created_time").Find(&certs, &Cert{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := ormer.Engine.Table(&Permission{}).Select("created_time").Find(&permissions, &Permission{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := ormer.Engine.Table(&Transaction{}).Select("created_time").Find(&transactions, &Transaction{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := ormer.Engine.Table(&Model{}).Select("created_time").Find(&models, &Model{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := ormer.Engine.Table(&Adapter{}).Select("created_time").Find(&adapters, &Adapter{Owner: owner}); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := ormer.Engine.Table(&Enforcer{}).Select("created_time").Find(&enforcers, &Enforcer{Owner: owner}); err != nil {
			panic(err)
		}
	}()
	wg.Wait()

	nowTime := time.Now()
	for i := 30; i >= 0; i-- {
		cutTime := nowTime.AddDate(0, 0, -i)
		dashboard.OrganizationCounts[30-i] = countCreatedBefore(organizations, cutTime)
		dashboard.UserCounts[30-i] = countCreatedBefore(users, cutTime)
		dashboard.ProviderCounts[30-i] = countCreatedBefore(providers, cutTime)
		dashboard.ApplicationCounts[30-i] = countCreatedBefore(applications, cutTime)
		dashboard.SubscriptionCounts[30-i] = countCreatedBefore(subscriptions, cutTime)
		dashboard.RoleCounts[30-i] = countCreatedBefore(roles, cutTime)
		dashboard.GroupCounts[30-i] = countCreatedBefore(groups, cutTime)
		dashboard.ResourceCounts[30-i] = countCreatedBefore(resources, cutTime)
		dashboard.CertCounts[30-i] = countCreatedBefore(certs, cutTime)
		dashboard.PermissionCounts[30-i] = countCreatedBefore(permissions, cutTime)
		dashboard.TransactionCounts[30-i] = countCreatedBefore(transactions, cutTime)
		dashboard.ModelCounts[30-i] = countCreatedBefore(models, cutTime)
		dashboard.AdapterCounts[30-i] = countCreatedBefore(adapters, cutTime)
		dashboard.EnforcerCounts[30-i] = countCreatedBefore(enforcers, cutTime)
	}
	return dashboard, nil
}

func countCreatedBefore(objects interface{}, before time.Time) int {
	count := 0
	switch obj := objects.(type) {
	case []Organization:
		for _, o := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", o.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []User:
		for _, u := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", u.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Provider:
		for _, p := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", p.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Application:
		for _, a := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", a.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Subscription:
		for _, s := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", s.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Role:
		for _, r := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", r.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Group:
		for _, g := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", g.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Resource:
		for _, r := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", r.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Cert:
		for _, c := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", c.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Permission:
		for _, p := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", p.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Transaction:
		for _, t := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", t.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Model:
		for _, m := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", m.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Adapter:
		for _, a := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", a.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []Enforcer:
		for _, e := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", e.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	case []DashboardDateItem:
		for _, e := range obj {
			createdTime, _ := time.Parse("2006-01-02T15:04:05-07:00", e.CreatedTime)
			if createdTime.Before(before) {
				count++
			}
		}
	}
	return count
}
