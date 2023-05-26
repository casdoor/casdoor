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

func GetPlanCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Plan{})
}

func GetPlans(owner string) ([]*Plan, error) {
	plans := []*Plan{}
	err := adapter.Engine.Desc("created_time").Find(&plans, &Plan{Owner: owner})
	if err != nil {
		return plans, err
	}
	return plans, nil
}

func GetPaginatedPlans(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Plan, error) {
	plans := []*Plan{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&plans)
	if err != nil {
		return plans, err
	}
	return plans, nil
}

func getPlan(owner, name string) (*Plan, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	plan := Plan{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&plan)
	if err != nil {
		return &plan, err
	}
	if existed {
		return &plan, nil
	} else {
		return nil, nil
	}
}

func GetPlan(id string) (*Plan, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getPlan(owner, name)
}

func UpdatePlan(id string, plan *Plan) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if p, err := getPlan(owner, name); err != nil {
		return false, err
	} else if p == nil {
		return false, nil
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(plan)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddPlan(plan *Plan) (bool, error) {
	affected, err := adapter.Engine.Insert(plan)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeletePlan(plan *Plan) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{plan.Owner, plan.Name}).Delete(plan)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (plan *Plan) GetId() string {
	return fmt.Sprintf("%s/%s", plan.Owner, plan.Name)
}

func Subscribe(owner string, user string, plan string, pricing string) (*Subscription, error) {
	selectedPricing, err := GetPricing(fmt.Sprintf("%s/%s", owner, pricing))
	if err != nil {
		return nil, err
	}

	valid := selectedPricing != nil && selectedPricing.IsEnabled

	if !valid {
		return nil, nil
	}

	planBelongToPricing, err := selectedPricing.HasPlan(owner, plan)
	if err != nil {
		return nil, err
	}

	if planBelongToPricing {
		newSubscription := NewSubscription(owner, user, plan, selectedPricing.TrialDuration)
		affected, err := AddSubscription(newSubscription)
		if err != nil {
			return nil, err
		}

		if affected {
			return newSubscription, nil
		}
	}
	return nil, nil
}
