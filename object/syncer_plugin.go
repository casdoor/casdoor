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

	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/util"
)

// PluginSyncerProvider imports users through an action of kind "syncer" installed from the Marketplace, such as the
// Authing import. The syncer's host, user and password reach the action as Host, ClientId and ClientSecret.
type PluginSyncerProvider struct {
	Syncer *Syncer
}

func (p *PluginSyncerProvider) getAction() (*action.Module, *action.Config, error) {
	if p.Syncer.Plugin == "" {
		return nil, nil, fmt.Errorf("choose the syncer plugin of the syncer: %s", p.Syncer.Name)
	}
	m, _, err := getInstalledAction(p.Syncer.Organization, p.Syncer.Plugin, "syncer")
	if err != nil {
		return nil, nil, err
	}
	config := &action.Config{Host: p.Syncer.Host, ClientId: p.Syncer.User, ClientSecret: p.Syncer.Password, Metadata: p.Syncer.Metadata}
	return m, config, nil
}

func (p *PluginSyncerProvider) InitAdapter() error {
	_, _, err := p.getAction()
	return err
}

func (p *PluginSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	m, config, err := p.getAction()
	if err != nil {
		return nil, err
	}
	users, err := action.ListSyncerUsers(m, config)
	if err != nil {
		return nil, err
	}

	res := []*OriginalUser{}
	for _, u := range users {
		createdTime := u.CreatedTime
		if createdTime == "" {
			createdTime = util.GetCurrentTime()
		}
		properties := u.Properties
		if properties == nil {
			properties = map[string]string{}
		}
		res = append(res, &OriginalUser{
			Id:            u.Id,
			Name:          u.Name,
			DisplayName:   u.DisplayName,
			Email:         u.Email,
			EmailVerified: u.EmailVerified,
			Phone:         u.Phone,
			CountryCode:   u.CountryCode,
			Avatar:        u.Avatar,
			Gender:        u.Gender,
			Affiliation:   u.Affiliation,
			Title:         u.Title,
			IsForbidden:   u.IsForbidden,
			CreatedTime:   createdTime,
			Address:       []string{},
			Properties:    properties,
			Groups:        p.getGroupIds(u.Groups),
		})
	}
	return res, nil
}

func (p *PluginSyncerProvider) getGroupIds(departmentIds []string) []string {
	res := []string{}
	for _, departmentId := range departmentIds {
		if departmentId != "" {
			res = append(res, util.GetId(p.Syncer.Organization, departmentId))
		}
	}
	return res
}

// GetOriginalGroups reads the departments of a plugin that has "syncer.listGroups", named by their IDs in the
// external system like the DingTalk and WeCom syncers do
func (p *PluginSyncerProvider) GetOriginalGroups() ([]*OriginalGroup, error) {
	m, config, err := p.getAction()
	if err != nil {
		return nil, err
	}
	groups, err := action.ListSyncerGroups(m, config)
	if action.IsNotImplemented(err) {
		return []*OriginalGroup{}, nil
	}
	if err != nil {
		return nil, err
	}

	ids := map[string]bool{}
	for _, group := range groups {
		ids[group.Id] = true
	}
	res := []*OriginalGroup{}
	for _, group := range groups {
		if group.Id == "" {
			continue
		}
		displayName := group.DisplayName
		if displayName == "" {
			displayName = group.Name
		}
		parentId := ""
		if ids[group.ParentId] && group.ParentId != group.Id {
			parentId = group.ParentId
		}
		res = append(res, &OriginalGroup{
			Id:          util.GetId(p.Syncer.Organization, group.Id),
			Name:        group.Id,
			DisplayName: displayName,
			Type:        "department",
			ParentId:    parentId,
		})
	}
	return res, nil
}

func (p *PluginSyncerProvider) GetOriginalUserGroups(userId string) ([]string, error) {
	return []string{}, nil
}

func (p *PluginSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	return false, fmt.Errorf("adding users through the syncer plugin is not supported")
}

func (p *PluginSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	return false, fmt.Errorf("updating users through the syncer plugin is not supported")
}

func (p *PluginSyncerProvider) IsReadOnly() bool {
	return true
}

func (p *PluginSyncerProvider) TestConnection() error {
	m, config, err := p.getAction()
	if err != nil {
		return err
	}
	return action.TestSyncer(m, config)
}

func (p *PluginSyncerProvider) Close() error {
	return nil
}
