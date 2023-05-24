// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Plan struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(100)" json:"description"`

	PricePerMonth float64 `json:"pricePerMonth"`
	PricePerYear  float64 `json:"pricePerYear"`
	Currency      string  `xorm:"varchar(100)" json:"currency"`
	IsEnabled     bool    `json:"isEnabled"`

	Role    string   `xorm:"varchar(100)" json:"role"`
	Options []string `xorm:"-" json:"options"`
}

func GetPlanCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Plan{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetPlans(owner string) []*Plan {
	plans := []*Plan{}
	err := adapter.Engine.Desc("created_time").Find(&plans, &Plan{Owner: owner})
	if err != nil {
		panic(err)
	}
	return plans
}

func GetPaginatedPlans(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Plan {
	plans := []*Plan{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&plans)
	if err != nil {
		panic(err)
	}
	return plans
}

func getPlan(owner, name string) *Plan {
	if owner == "" || name == "" {
		return nil
	}

	plan := Plan{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&plan)
	if err != nil {
		panic(err)
	}
	if existed {
		return &plan
	} else {
		return nil
	}
}

func GetPlan(id string) *Plan {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getPlan(owner, name)
}

func UpdatePlan(id string, plan *Plan) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getPlan(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(plan)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddPlan(plan *Plan) bool {
	affected, err := adapter.Engine.Insert(plan)
	if err != nil {
		panic(err)
	}
	return affected != 0
}

func DeletePlan(plan *Plan) bool {
	affected, err := adapter.Engine.ID(core.PK{plan.Owner, plan.Name}).Delete(plan)
	if err != nil {
		panic(err)
	}
	return affected != 0
}

func (plan *Plan) GetId() string {
	return fmt.Sprintf("%s/%s", plan.Owner, plan.Name)
}

func Subscribe(owner string, user string, plan string, pricing string) *Subscription {
	selectedPricing := GetPricing(fmt.Sprintf("%s/%s", owner, pricing))

	valid := selectedPricing != nil && selectedPricing.IsEnabled

	if !valid {
		return nil
	}

	planBelongToPricing := selectedPricing.HasPlan(owner, plan)

	if planBelongToPricing {
		newSubscription := NewSubscription(owner, user, plan, selectedPricing.TrialDuration)
		affected := AddSubscription(newSubscription)

		if affected {
			return newSubscription
		}
	}
	return nil
}
