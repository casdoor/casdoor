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
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Pricing struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(100)" json:"description"`

	Plans         []string `xorm:"mediumtext" json:"plans"`
	IsEnabled     bool     `json:"isEnabled"`
	HasTrial      bool     `json:"hasTrial"`
	TrialDuration int      `json:"trialDuration"`
	Application   string   `xorm:"varchar(100)" json:"application"`

	Submitter   string `xorm:"varchar(100)" json:"submitter"`
	Approver    string `xorm:"varchar(100)" json:"approver"`
	ApproveTime string `xorm:"varchar(100)" json:"approveTime"`

	State string `xorm:"varchar(100)" json:"state"`
}

func GetPricingCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Pricing{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetPricings(owner string) []*Pricing {
	pricings := []*Pricing{}
	err := adapter.Engine.Desc("created_time").Find(&pricings, &Pricing{Owner: owner})
	if err != nil {
		panic(err)
	}
	return pricings
}

func GetPaginatedPricings(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Pricing {
	pricings := []*Pricing{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&pricings)
	if err != nil {
		panic(err)
	}
	return pricings
}

func getPricing(owner, name string) *Pricing {
	if owner == "" || name == "" {
		return nil
	}

	pricing := Pricing{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&pricing)
	if err != nil {
		panic(err)
	}
	if existed {
		return &pricing
	} else {
		return nil
	}
}

func GetPricing(id string) *Pricing {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getPricing(owner, name)
}

func UpdatePricing(id string, pricing *Pricing) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getPricing(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(pricing)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddPricing(pricing *Pricing) bool {
	affected, err := adapter.Engine.Insert(pricing)
	if err != nil {
		panic(err)
	}
	return affected != 0
}

func DeletePricing(pricing *Pricing) bool {
	affected, err := adapter.Engine.ID(core.PK{pricing.Owner, pricing.Name}).Delete(pricing)
	if err != nil {
		panic(err)
	}
	return affected != 0
}

func (pricing *Pricing) GetId() string {
	return fmt.Sprintf("%s/%s", pricing.Owner, pricing.Name)
}

func (pricing *Pricing) HasPlan(owner string, plan string) bool {
	selectedPlan := GetPlan(fmt.Sprintf("%s/%s", owner, plan))

	if selectedPlan == nil {
		return false
	}

	result := false

	for _, pricingPlan := range pricing.Plans {
		if strings.Contains(pricingPlan, selectedPlan.Name) {
			result = true
			break
		}
	}

	return result
}
