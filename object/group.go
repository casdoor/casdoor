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
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/builder"
	"github.com/xorm-io/core"
)

type Group struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk unique index" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	DisplayName  string   `xorm:"varchar(100)" json:"displayName"`
	Manager      string   `xorm:"varchar(100)" json:"manager"`
	ContactEmail string   `xorm:"varchar(100)" json:"contactEmail"`
	Type         string   `xorm:"varchar(100)" json:"type"`
	ParentId     string   `xorm:"varchar(100)" json:"parentId"`
	ParentName   string   `xorm:"-" json:"parentName"`
	IsTopGroup   bool     `xorm:"bool" json:"isTopGroup"`
	Users        []string `xorm:"-" json:"users"`

	Title        string   `json:"title,omitempty"`
	Key          string   `json:"key,omitempty"`
	HaveChildren bool     `xorm:"-" json:"haveChildren"`
	Children     []*Group `json:"children,omitempty"`

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
	err := ormer.Engine.Desc("created_time").Find(&groups, &Group{Owner: owner})
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

func GetGroupsHaveChildrenMap(groups []*Group) (map[string]*Group, error) {
	groupsHaveChildren := []*Group{}
	resultMap := make(map[string]*Group)
	groupMap := map[string]*Group{}

	groupIds := []string{}
	for _, group := range groups {
		groupMap[group.Name] = group
		groupIds = append(groupIds, group.Name)
		if !group.IsTopGroup {
			groupIds = append(groupIds, group.ParentId)
		}
	}

	err := ormer.Engine.Cols("owner", "name", "parent_id", "display_name").Distinct("name").In("name", groupIds).Find(&groupsHaveChildren)
	if err != nil {
		return nil, err
	}

	for _, group := range groupsHaveChildren {
		resultMap[group.GetId()] = group
	}
	return resultMap, nil
}

func getGroup(owner string, name string) (*Group, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	group := Group{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&group)
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

	err = checkGroupName(group.Name)
	if err != nil {
		return false, err
	}

	if name != group.Name {
		err := GroupChangeTrigger(name, group.Name)
		if err != nil {
			return false, err
		}
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(group)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddGroup(group *Group) (bool, error) {
	err := checkGroupName(group.Name)
	if err != nil {
		return false, err
	}

	affected, err := ormer.Engine.Insert(group)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddGroups(groups []*Group) (bool, error) {
	if len(groups) == 0 {
		return false, nil
	}
	affected, err := ormer.Engine.Insert(groups)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func deleteGroup(group *Group) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{group.Owner, group.Name}).Delete(&Group{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteGroup(group *Group) (bool, error) {
	_, err := ormer.Engine.Get(group)
	if err != nil {
		return false, err
	}

	if count, err := ormer.Engine.Where("parent_id = ?", group.Name).Count(&Group{}); err != nil {
		return false, err
	} else if count > 0 {
		return false, errors.New("group has children group")
	}

	if count, err := GetGroupUserCount(group.GetId(), "", ""); err != nil {
		return false, err
	} else if count > 0 {
		return false, errors.New("group has users")
	}

	return deleteGroup(group)
}

func checkGroupName(name string) error {
	if name == "" {
		return errors.New("group name can't be empty")
	}
	if strings.Contains(name, "/") {
		return errors.New("group name can't contain \"/\"")
	}
	exist, err := ormer.Engine.Exist(&Organization{Owner: "admin", Name: name})
	if err != nil {
		return err
	}
	if exist {
		return errors.New("group name can't be same as the organization name")
	}
	return nil
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

func GetGroupUserCount(groupId string, field, value string) (int64, error) {
	owner, _ := util.GetOwnerAndNameFromId(groupId)
	names, err := userEnforcer.GetUserNamesByGroupName(groupId)
	if err != nil {
		return 0, err
	}

	if field == "" && value == "" {
		return int64(len(names)), nil
	} else {
		tableNamePrefix := conf.GetConfigString("tableNamePrefix")
		return ormer.Engine.Table(tableNamePrefix+"user").
			Where("owner = ?", owner).In("name", names).
			And(fmt.Sprintf("user.%s like ?", util.CamelToSnakeCase(field)), "%"+value+"%").
			Count()
	}
}

func GetPaginationGroupUsers(groupId string, offset, limit int, field, value, sortField, sortOrder string) ([]*User, error) {
	users := []*User{}
	owner, _ := util.GetOwnerAndNameFromId(groupId)
	names, err := userEnforcer.GetUserNamesByGroupName(groupId)
	if err != nil {
		return nil, err
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	prefixedUserTable := tableNamePrefix + "user"
	session := ormer.Engine.Table(prefixedUserTable).
		Where("owner = ?", owner).In("name", names)

	if offset != -1 && limit != -1 {
		session.Limit(limit, offset)
	}

	if field != "" && value != "" {
		session = session.And(fmt.Sprintf("%s.%s like ?", prefixedUserTable, util.CamelToSnakeCase(field)), "%"+value+"%")
	}

	if sortField == "" || sortOrder == "" {
		sortField = "created_time"
	}

	orderQuery := fmt.Sprintf("%s.%s", prefixedUserTable, util.SnakeString(sortField))

	if sortOrder == "ascend" {
		session = session.Asc(orderQuery)
	} else {
		session = session.Desc(orderQuery)
	}

	err = session.Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetGroupUsers(groupId string) ([]*User, error) {
	users := []*User{}
	owner, _, err := util.GetOwnerAndNameFromIdWithError(groupId)
	if err != nil {
		return nil, err
	}
	names, err := userEnforcer.GetUserNamesByGroupName(groupId)
	if err != nil {
		return nil, err
	}
	err = ormer.Engine.Where("owner = ?", owner).In("name", names).Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetGroupUsersWithoutError(groupId string) []*User {
	users, _ := GetGroupUsers(groupId)
	return users
}

func ExtendGroupWithUsers(group *Group) error {
	if group == nil {
		return nil
	}

	groupId := group.GetId()
	userIds := []string{}
	userIds, err := userEnforcer.GetAllUsersByGroup(groupId)
	if err != nil {
		return err
	}

	group.Users = userIds
	return nil
}

func ExtendGroupsWithUsers(groups []*Group) error {
	for _, group := range groups {
		users, err := userEnforcer.GetAllUsersByGroup(group.GetId())
		if err != nil {
			return err
		}

		group.Users = users
	}
	return nil
}

func GroupChangeTrigger(oldName, newName string) error {
	session := ormer.Engine.NewSession()
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
	if err != nil {
		return err
	}
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
