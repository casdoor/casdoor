// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Provider struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName  string `xorm:"varchar(100)" json:"displayName"`
	Category     string `xorm:"varchar(100)" json:"category"`
	Type         string `xorm:"varchar(100)" json:"type"`
	ClientId     string `xorm:"varchar(100)" json:"clientId"`
	ClientSecret string `xorm:"varchar(100)" json:"clientSecret"`
	Host         string `xorm:"varchar(100)" json:"host"`
	Port         int    `json:"port"`
	ProviderUrl  string `xorm:"varchar(200)" json:"providerUrl"`
}

func GetProviders(owner string) []*Provider {
	providers := []*Provider{}
	err := adapter.Engine.Desc("created_time").Find(&providers, &Provider{Owner: owner})
	if err != nil {
		panic(err)
	}

	return providers
}

func getProvider(owner string, name string) *Provider {
	provider := Provider{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&provider)
	if err != nil {
		panic(err)
	}

	if existed {
		return &provider
	} else {
		return nil
	}
}

func GetProvider(id string) *Provider {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getProvider(owner, name)
}

func getDefaultEmailProvider() *Provider {
	provider := Provider{Owner: "admin", Category: "Email"}
	existed, err := adapter.Engine.Get(&provider)
	if err != nil {
		panic(err)
	}

	if !existed {
		return nil
	}
	return &provider
}

func UpdateProvider(id string, provider *Provider) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getProvider(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(provider)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddProvider(provider *Provider) bool {
	affected, err := adapter.Engine.Insert(provider)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteProvider(provider *Provider) bool {
	affected, err := adapter.Engine.ID(core.PK{provider.Owner, provider.Name}).Delete(&Provider{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}
