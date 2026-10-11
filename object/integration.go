// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

// Integration is an app integration, provider template or theme installed from the Marketplace.
type Integration struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	IntegrationId string            `xorm:"varchar(100)" json:"integrationId"`
	Type          string            `xorm:"varchar(100)" json:"type"`
	Version       string            `xorm:"varchar(100)" json:"version"`
	Logo          string            `xorm:"varchar(500)" json:"logo"`
	Variables     map[string]string `xorm:"mediumtext" json:"variables"`

	// the application it created (app) or changed (theme), and the provider it created (provider)
	Application string `xorm:"varchar(100)" json:"application"`
	Provider    string `xorm:"varchar(100)" json:"provider"`

	// the theme fields of the application before a theme was installed, restored on uninstall
	Backup string `xorm:"mediumtext" json:"backup"`
	Bundle string `xorm:"mediumtext" json:"bundle"`
}

func (integration *Integration) GetId() string {
	return fmt.Sprintf("%s/%s", integration.Owner, integration.Name)
}

func GetIntegrationCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Integration{Owner: owner})
}

func GetIntegrations(owner string) ([]*Integration, error) {
	integrations := []*Integration{}
	err := ormer.Engine.Desc("created_time").Find(&integrations, &Integration{Owner: owner})
	if err != nil {
		return nil, err
	}
	return integrations, nil
}

func GetPaginationIntegrations(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Integration, error) {
	integrations := []*Integration{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&integrations, &Integration{Owner: owner})
	if err != nil {
		return nil, err
	}
	return integrations, nil
}

func getIntegration(owner string, name string) (*Integration, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	integration := Integration{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&integration)
	if err != nil {
		return nil, err
	}

	if existed {
		return &integration, nil
	} else {
		return nil, nil
	}
}

func GetIntegration(id string) (*Integration, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getIntegration(owner, name)
}

func addIntegration(integration *Integration) (bool, error) {
	affected, err := ormer.Engine.Insert(integration)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func deleteIntegration(integration *Integration) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{integration.Owner, integration.Name}).Delete(&Integration{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
