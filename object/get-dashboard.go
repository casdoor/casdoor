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

func GetDashboard() (map[string]interface{}, error) {
	Dashboard := make(map[string]interface{})
	usersCount := make([]int, 7)
	organizationsCount := make([]int, 7)
	applicationsCount := make([]int, 7)
	providersCount := make([]int, 7)
	subscriptionsCount := make([]int, 7)

	loc, _ := time.LoadLocation("Asia/Shanghai")
	for i := 6; i >= 0; i-- {
		endTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-i, 23, 59, 59, 0, loc).Format("2006-01-02T15:04:05-07:00")

		if count, err := ormer.Engine.Where("created_time <= ?", endTime).Count(new(User)); err != nil {
			return Dashboard, err
		} else {
			usersCount[6-i] = int(count)
		}

		if count, err := ormer.Engine.Where("created_time <= ?", endTime).Count(new(Organization)); err != nil {
			return Dashboard, err
		} else {
			organizationsCount[6-i] = int(count)
		}

		if count, err := ormer.Engine.Where("created_time <= ?", endTime).Count(new(Application)); err != nil {
			return Dashboard, err
		} else {
			applicationsCount[6-i] = int(count)
		}

		if count, err := ormer.Engine.Where("created_time <= ?", endTime).Count(new(Provider)); err != nil {
			return Dashboard, err
		} else {
			providersCount[6-i] = int(count)
		}

		if count, err := ormer.Engine.Where("created_time <= ?", endTime).Count(new(Subscription)); err != nil {
			return Dashboard, err
		} else {
			subscriptionsCount[6-i] = int(count)
		}
	}

	Dashboard["usersCount"] = usersCount
	Dashboard["TodayNewUsersCount"] = usersCount[6] - usersCount[5]
	Dashboard["TotalUsersCount"] = usersCount[6]
	Dashboard["PastSevenDaysNewUsersCount"] = usersCount[6] - usersCount[0]
	Dashboard["organizationsCount"] = organizationsCount
	Dashboard["applicationsCount"] = applicationsCount
	Dashboard["providersCount"] = providersCount
	Dashboard["subscriptionsCount"] = subscriptionsCount

	return Dashboard, nil
}
