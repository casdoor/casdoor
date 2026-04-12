// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"github.com/xorm-io/core"
)

type ThirdPartyLink struct {
	Owner        string `xorm:"varchar(100) notnull pk unique(link_unique)" json:"owner"`
	UserName     string `xorm:"varchar(100) notnull pk" json:"userName"`
	ProviderName string `xorm:"varchar(100) notnull pk unique(link_unique)" json:"providerName"`
	ProviderId   string `xorm:"varchar(100) notnull unique(link_unique)" json:"providerId"`
	CreatedTime  string `xorm:"varchar(100)" json:"createdTime"`
}

func IsFlexibleCustomProvider(providerType string) bool {
	return providerType == "Custom Flexible"
}

func GetThirdPartyLinksByUser(owner string, userName string) ([]*ThirdPartyLink, error) {
	links := []*ThirdPartyLink{}
	err := ormer.Engine.Where("owner = ? AND user_name = ?", owner, userName).Find(&links)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func GetThirdPartyLink(owner string, userName string, providerName string) (*ThirdPartyLink, error) {
	link := ThirdPartyLink{Owner: owner, UserName: userName, ProviderName: providerName}
	existed, err := ormer.Engine.Get(&link)
	if err != nil {
		return nil, err
	}
	if existed {
		return &link, nil
	}
	return nil, nil
}

func GetUserByThirdPartyLink(owner string, providerName string, providerId string) (*User, error) {
	if owner == "" || providerName == "" || providerId == "" {
		return nil, nil
	}

	link := ThirdPartyLink{}
	existed, err := ormer.Engine.Where("owner = ? AND provider_name = ? AND provider_id = ?", owner, providerName, providerId).Get(&link)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, nil
	}

	return getUser(link.Owner, link.UserName)
}

func AddThirdPartyLink(link *ThirdPartyLink) (bool, error) {
	existingLink, err := GetThirdPartyLink(link.Owner, link.UserName, link.ProviderName)
	if err != nil {
		return false, err
	}

	if existingLink != nil {
		existingLink.ProviderId = link.ProviderId
		affected, err := ormer.Engine.ID(core.PK{link.Owner, link.UserName, link.ProviderName}).Cols("provider_id").Update(existingLink)
		if err != nil {
			return false, err
		}
		return affected != 0, nil
	}

	link.CreatedTime = util.GetCurrentTime()
	affected, err := ormer.Engine.Insert(link)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteThirdPartyLink(owner string, userName string, providerName string) (bool, error) {
	affected, err := ormer.Engine.Delete(&ThirdPartyLink{Owner: owner, UserName: userName, ProviderName: providerName})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteThirdPartyLinksByUser(owner string, userName string) (bool, error) {
	affected, err := ormer.Engine.Where("owner = ? AND user_name = ?", owner, userName).Delete(&ThirdPartyLink{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
