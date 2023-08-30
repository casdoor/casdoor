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

type Pricing struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(100)" json:"description"`

	Plans         []string `xorm:"mediumtext" json:"plans"`
	IsEnabled     bool     `json:"isEnabled"`
	TrialDuration int      `json:"trialDuration"`
	Application   string   `xorm:"varchar(100)" json:"application"`
}

func (pricing *Pricing) GetId() string {
	return fmt.Sprintf("%s/%s", pricing.Owner, pricing.Name)
}

func (pricing *Pricing) HasPlan(planName string) (bool, error) {
	planId := util.GetId(pricing.Owner, planName)
	plan, err := GetPlan(planId)
	if err != nil {
		return false, err
	}
	if plan == nil {
		return false, fmt.Errorf("plan: %s does not exist", planId)
	}

	if util.InSlice(pricing.Plans, plan.Name) {
		return true, nil
	}
	return false, nil
}

func GetPricingCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Pricing{})
}

func GetPricings(owner string) ([]*Pricing, error) {
	pricings := []*Pricing{}
	err := ormer.Engine.Desc("created_time").Find(&pricings, &Pricing{Owner: owner})
	if err != nil {
		return pricings, err
	}

	return pricings, nil
}

func GetPaginatedPricings(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Pricing, error) {
	pricings := []*Pricing{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&pricings)
	if err != nil {
		return pricings, err
	}
	return pricings, nil
}

func getPricing(owner, name string) (*Pricing, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	pricing := Pricing{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&pricing)
	if err != nil {
		return nil, err
	}
	if existed {
		return &pricing, nil
	} else {
		return nil, nil
	}
}

func GetPricing(id string) (*Pricing, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getPricing(owner, name)
}

func GetApplicationDefaultPricing(owner, appName string) (*Pricing, error) {
	pricings := make([]*Pricing, 0, 1)
	err := ormer.Engine.Asc("created_time").Find(&pricings, &Pricing{Owner: owner, Application: appName})
	if err != nil {
		return nil, err
	}
	for _, pricing := range pricings {
		if pricing.IsEnabled {
			return pricing, nil
		}
	}
	return nil, nil
}

func UpdatePricing(id string, pricing *Pricing) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if p, err := getPricing(owner, name); err != nil {
		return false, err
	} else if p == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(pricing)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddPricing(pricing *Pricing) (bool, error) {
	affected, err := ormer.Engine.Insert(pricing)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeletePricing(pricing *Pricing) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{pricing.Owner, pricing.Name}).Delete(pricing)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func CheckPricingAndPlan(owner, pricingName, planName string) error {
	pricingId := util.GetId(owner, pricingName)
	pricing, err := GetPricing(pricingId)
	if pricing == nil || err != nil {
		if pricing == nil && err == nil {
			err = fmt.Errorf("pricing: %s does not exist", pricingName)
		}
		return err
	}
	ok, err := pricing.HasPlan(planName)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("pricing: %s does not have plan: %s", pricingName, planName)
	}
	return nil
}
