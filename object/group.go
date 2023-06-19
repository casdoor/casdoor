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
	"errors"
	"fmt"

	"github.com/xorm-io/builder"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Group struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk unique index" json:"name"`
	CreatedTime string `xorm:"varchar(100) created" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100) updated" json:"updatedTime"`

	DisplayName  string  `xorm:"varchar(100)" json:"displayName"`
	Manager      string  `xorm:"varchar(100)" json:"manager"`
	ContactEmail string  `xorm:"varchar(100)" json:"contactEmail"`
	Type         string  `xorm:"varchar(100)" json:"type"`
	ParentId     string  `xorm:"varchar(100)" json:"parentId"`
	IsTopGroup   bool    `xorm:"bool" json:"isTopGroup"`
	Users        []*User `xorm:"-" json:"users"`

	Title    string   `json:"title,omitempty"`
	Key      string   `json:"key,omitempty"`
	Children []*Group `json:"children,omitempty"`

	IsEnabled bool `json:"isEnabled"`
}

type GroupNode struct{}

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

	if name != group.Name {
		err := GroupChangeTrigger(name, group.Name)
		if err != nil {
			return false, err
		}
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(group)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddGroup(group *Group) (bool, error) {
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
	_, err := adapter.Engine.Get(group)
	if err != nil {
		return false, err
	}

	if count, err := adapter.Engine.Where("parent_id = ?", group.Name).Count(&Group{}); err != nil {
		return false, err
	} else if count > 0 {
		return false, errors.New("group has children group")
	}

	if count, err := GetGroupUserCount(group.Name, "", ""); err != nil {
		return false, err
	} else if count > 0 {
		return false, errors.New("group has users")
	}

	affected, err := adapter.Engine.ID(core.PK{group.Owner, group.Name}).Delete(&Group{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (group *Group) GetId() string {
	return fmt.Sprintf("%s/%s", group.Owner, group.Name)
}

func ConvertToTreeData(groups []*Group, parentId string) []*Group {
	treeData := []*Group{}

	for _, group := range groups {
		if group.ParentId == parentId {
			node := &Group{
				Title: group.DisplayName,
				Key:   group.Name,
				Type:  group.Type,
				Owner: group.Owner,
			}
			children := ConvertToTreeData(groups, group.Name)
			if len(children) > 0 {
				node.Children = children
			}
			treeData = append(treeData, node)
		}
	}
	return treeData
}

func RemoveUserFromGroup(owner, name, groupName string) (bool, error) {
	user, err := getUser(owner, name)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("user not exist")
	}

	user.Groups = util.DeleteVal(user.Groups, groupName)
	affected, err := updateUser(user.GetId(), user, []string{"groups"})
	if err != nil {
		return false, err
	}
	return affected != 0, err
}

func GetGroupUserCount(groupName string, field, value string) (int64, error) {
	if field == "" && value == "" {
		return adapter.Engine.Where(builder.Like{"`groups`", groupName}).
			Count(&User{})
	} else {
		return adapter.Engine.Table("user").
			Where(builder.Like{"`groups`", groupName}).
			And(fmt.Sprintf("user.%s LIKE ?", util.CamelToSnakeCase(field)), "%"+value+"%").
			Count()
	}
}

func GetPaginationGroupUsers(groupName string, offset, limit int, field, value, sortField, sortOrder string) ([]*User, error) {
	users := []*User{}
	session := adapter.Engine.Table("user").
		Where(builder.Like{"`groups`", groupName})

	if offset != -1 && limit != -1 {
		session.Limit(limit, offset)
	}

	if field != "" && value != "" {
		session = session.And(fmt.Sprintf("user.%s LIKE ?", util.CamelToSnakeCase(field)), "%"+value+"%")
	}

	if sortField == "" || sortOrder == "" {
		sortField = "created_time"
	}
	if sortOrder == "ascend" {
		session = session.Asc(fmt.Sprintf("user.%s", util.SnakeString(sortField)))
	} else {
		session = session.Desc(fmt.Sprintf("user.%s", util.SnakeString(sortField)))
	}

	err := session.Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetGroupUsers(groupName string) ([]*User, error) {
	users := []*User{}
	err := adapter.Engine.Table("user").
		Where(builder.Like{"`groups`", groupName}).
		Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GroupChangeTrigger(oldName, newName string) error {
	session := adapter.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}

	users := []*User{}
	err = session.Where(builder.Like{"`groups`", oldName}).Find(&users)
	if err != nil {
		return err
	}

	for _, user := range users {
		user.Groups = util.ReplaceVal(user.Groups, oldName, newName)
		_, err := updateUser(user.GetId(), user, []string{"groups"})
		if err != nil {
			return err
		}
	}

	groups := []*Group{}
	err = session.Where("parent_id = ?", oldName).Find(&groups)
	for _, group := range groups {
		group.ParentId = newName
		_, err := session.ID(core.PK{group.Owner, group.Name}).Cols("parent_id").Update(group)
		if err != nil {
			return err
		}
	}

	err = session.Commit()
	if err != nil {
		return err
	}
	return nil
}
