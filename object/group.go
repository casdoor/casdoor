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

type Group struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk unique" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	Id            string    `xorm:"varchar(100) not null index" json:"id"`
	DisplayName   string    `xorm:"varchar(100)" json:"displayName"`
	Manager       string    `xorm:"varchar(100)" json:"manager"`
	ContactEmail  string    `xorm:"varchar(100)" json:"contactEmail"`
	Type          string    `xorm:"varchar(100)" json:"type"`
	ParentGroupId string    `xorm:"varchar(100)" json:"parentGroupId"`
	Users         *[]string `xorm:"-" json:"users"`

	IsEnabled bool `json:"isEnabled"`
}

type GroupNode struct {
	Title    string       `json:"title"`
	Key      string       `json:"key"`
	Children []*GroupNode `json:"children,omitempty"`
}

func GetGroupCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Group{})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetGroups(owner string) ([]*Group, error) {
	groups := []*Group{}
	err := adapter.Engine.Desc("created_time").Find(&groups, &Group{Owner: owner})
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func GetPaginationGroups(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Group, error) {
	groups := []*Group{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func getGroup(owner string, name string) (*Group, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	group := Group{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&group)
	if err != nil {
		return nil, err
	}

	if existed {
		return &group, nil
	} else {
		return nil, nil
	}
}

func getGroupById(id string) (*Group, error) {
	if id == "" {
		return nil, nil
	}

	group := Group{Id: id}
	existed, err := adapter.Engine.Get(&group)
	if err != nil {
		return nil, err
	}

	if existed {
		return &group, nil
	} else {
		return nil, nil
	}
}

func GetGroup(id string) (*Group, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getGroup(owner, name)
}

func UpdateGroup(id string, group *Group) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	oldGroup, err := getGroup(owner, name)
	if oldGroup == nil {
		return false, err
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(group)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddGroup(group *Group) (bool, error) {
	if group.Id == "" {
		group.Id = util.GenerateId()
	}

	affected, err := adapter.Engine.Insert(group)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddGroups(groups []*Group) (bool, error) {
	if len(groups) == 0 {
		return false, nil
	}
	affected, err := adapter.Engine.Insert(groups)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteGroup(group *Group) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{group.Owner, group.Name}).Delete(&Group{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (group *Group) GetId() string {
	return fmt.Sprintf("%s/%s", group.Owner, group.Name)
}

func ConvertToTreeData(groups []*Group, parentGroupId string) []*GroupNode {
	treeData := []*GroupNode{}

	for _, group := range groups {
		if group.ParentGroupId == parentGroupId {
			node := &GroupNode{
				Title: group.DisplayName,
				Key:   group.Id,
			}
			children := ConvertToTreeData(groups, group.Id)
			if len(children) > 0 {
				node.Children = children
			}
			treeData = append(treeData, node)
		}
	}
	return treeData
}
