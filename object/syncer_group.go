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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/util"
	wework "github.com/go-laoji/wecom-go-sdk/v2"
)

type OriginalGroup = Group

func (syncer *Syncer) getOriginalGroupFromWecom(groupList []wework.Department) []*OriginalGroup {
	groups := make([]*OriginalGroup, 0)
	m := make(map[int32]string)

	for _, wecomGroup := range groupList {
		m[wecomGroup.Id] = wecomGroup.Name
		oGroup := &OriginalGroup{
			Key:         string(wecomGroup.Id),
			Name:        wecomGroup.Name,
			DisplayName: wecomGroup.Name,
			ParentKey:   string(wecomGroup.ParentId),
		}

		if parentId, ok := m[wecomGroup.ParentId]; ok {
			oGroup.ParentId = syncer.Organization + "/" + parentId
		}

		groups = append(groups, oGroup)
	}
	return groups
}

func (syncer *Syncer) getOriginalGroups() ([]*OriginalGroup, error) {
	groupList := make([]wework.Department, 0)

	listRsp := syncer.WeComClient.DepartmentList(0, 0)
	if listRsp.ErrCode != 0 {
		return nil, fmt.Errorf(listRsp.ErrorMsg)
	}
	groupList = append(groupList, listRsp.Department...)

	return syncer.getOriginalGroupFromWecom(groupList), nil
}

func (syncer *Syncer) getOriginalGroupMap() ([]*OriginalGroup, map[string]*OriginalGroup, error) {
	groups, err := syncer.getOriginalGroups()
	if err != nil {
		return nil, nil, err
	}

	m := map[string]*OriginalGroup{}
	for _, group := range groups {
		m[group.Key] = group
	}
	return groups, m, nil
}

func (syncer *Syncer) getGroupValue(group *Group, key string) string {
	jsonData, _ := json.Marshal(group)
	var mapData map[string]interface{}
	if err := json.Unmarshal(jsonData, &mapData); err != nil {
		fmt.Println("conversion failed:", err)
		return group.Name
	}
	value := mapData[util.SnakeToCamel(key)]
	if str, ok := value.(string); ok {
		return str
	} else {
		if value != nil {
			valType := reflect.TypeOf(value)
			typeName := valType.Name()
			switch typeName {
			case "bool":
				return strconv.FormatBool(value.(bool))
			case "int":
				return strconv.Itoa(value.(int))
			}
		}
		return group.Name
	}
}

func (syncer *Syncer) getMapFromOriginalGroup(group *OriginalGroup) map[string]string {
	m := map[string]string{}
	m["Name"] = group.Name
	m["DisplayName"] = group.DisplayName
	m["Manager"] = group.Manager
	m["ContactEmail"] = group.ContactEmail
	m["Type"] = group.Type
	m["ParentId"] = group.ParentId
	m["IsTopGroup"] = util.BoolToString(group.IsTopGroup)
	m["Title"] = group.Title
	m["Key"] = group.Key
	m["IsEnabled"] = util.BoolToString(group.IsEnabled)

	return m
}

func (syncer *Syncer) calculateGroupHash(group *OriginalGroup) string {
	values := []string{}
	m := syncer.getMapFromOriginalGroup(group)
	for _, value := range m {
		values = append(values, value)
	}

	s := strings.Join(values, "|")
	return util.GetMd5Hash(s)
}

func (syncer *Syncer) createGroupFromOriginalGroup(originalGroup *OriginalGroup, affiliationMap map[int]string) *Group {
	group := *originalGroup
	group.Owner = syncer.Organization

	if group.CreatedTime == "" {
		group.CreatedTime = util.GetCurrentTime()
	}

	if group.Type == "" {
		group.Type = "Virtual"
	}

	return &group
}

func (syncer *Syncer) updateGroupForOriginalByFields(group *Group, key string) (bool, error) {
	var err error
	oldGroup := Group{}

	existed, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getGroupValue(group, key), group.Owner).Get(&oldGroup)
	if err != nil {
		return false, err
	}
	if !existed {
		return false, nil
	}

	affected, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getGroupValue(&oldGroup, key), oldGroup.Owner).Update(group)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (syncer *Syncer) getWeComGroupFromOriginalGroup(oGroup *OriginalGroup) (wework.Department, error) {
	wParentId, err := strconv.Atoi(oGroup.ParentKey)
	if err != nil {
		return wework.Department{}, fmt.Errorf("Parse parentKey error")
	}

	return wework.Department{
		Name:     oGroup.Name,
		ParentId: int32(wParentId),
	}, nil
}

func (syncer *Syncer) updateGroup(oGroup *OriginalGroup) (bool, error) {
	wecomGroup, err := syncer.getWeComGroupFromOriginalGroup(oGroup)
	if err != nil {
		return false, err
	}

	updateRsp := syncer.WeComClient.DepartmentUpdate(0, wecomGroup)
	if updateRsp.ErrCode != 0 {
		return false, fmt.Errorf(updateRsp.ErrorMsg)
	}

	return true, nil
}

func (syncer *Syncer) addGroup(oGroup *OriginalGroup) (bool, error) {
	wecomGroup, err := syncer.getWeComGroupFromOriginalGroup(oGroup)
	if err != nil {
		return false, err
	}

	createRsp := syncer.WeComClient.DepartmentCreate(0, wecomGroup)
	if createRsp.ErrCode != 0 {
		return false, fmt.Errorf(createRsp.ErrorMsg)
	}

	return true, nil
}
