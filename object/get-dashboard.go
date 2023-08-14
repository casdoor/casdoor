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

import "time"

type Dashboard struct {
	OrganizationCounts []int `json:"organizationCounts"`
	UserCounts         []int `json:"userCounts"`
	ProviderCounts     []int `json:"providerCounts"`
	ApplicationCounts  []int `json:"applicationCounts"`
	SubscriptionCounts []int `json:"subscriptionCounts"`
}

func GetDashboard() (*Dashboard, error) {
	dashboard := &Dashboard{
		OrganizationCounts: make([]int, 31),
		UserCounts:         make([]int, 31),
		ProviderCounts:     make([]int, 31),
		ApplicationCounts:  make([]int, 31),
		SubscriptionCounts: make([]int, 31),
	}

	nowTime := time.Now()

	organizations := make([]Organization, 0)
	if err := ormer.Engine.Find(&organizations); err != nil {
		return dashboard, err
	}
	users := make([]User, 0)
	if err := ormer.Engine.Find(&users); err != nil {
		return dashboard, err
	}
	providers := make([]Provider, 0)
	if err := ormer.Engine.Find(&providers); err != nil {
		return dashboard, err
	}
	applications := make([]Application, 0)
	if err := ormer.Engine.Find(&applications); err != nil {
		return dashboard, err
	}
	subscriptions := make([]Subscription, 0)
	if err := ormer.Engine.Find(&subscriptions); err != nil {
		return dashboard, err
	}

	for i := 30; i >= 0; i-- {
		cutTime := nowTime.AddDate(0, 0, -i)
		dashboard.OrganizationCounts[30-i] = countCreatedBefore(organizations, cutTime)
		dashboard.UserCounts[30-i] = countCreatedBefore(users, cutTime)
		dashboard.ProviderCounts[30-i] = countCreatedBefore(providers, cutTime)
		dashboard.ApplicationCounts[30-i] = countCreatedBefore(applications, cutTime)
		dashboard.SubscriptionCounts[30-i] = countCreatedBefore(subscriptions, cutTime)
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
	}
	return count
}
