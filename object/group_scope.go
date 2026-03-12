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
	"reflect"
	"sort"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
)

func normalizeGroupAdminUsers(adminUsers []string) []string {
	seen := map[string]struct{}{}
	res := make([]string, 0, len(adminUsers))
	for _, adminUser := range adminUsers {
		adminUser = strings.TrimSpace(adminUser)
		if adminUser == "" {
			continue
		}
		if _, ok := seen[adminUser]; ok {
			continue
		}
		seen[adminUser] = struct{}{}
		res = append(res, adminUser)
	}
	sort.Strings(res)
	return res
}

func validateGroupAdminUsers(group *Group) error {
	if group == nil {
		return nil
	}

	group.AdminUsers = normalizeGroupAdminUsers(group.AdminUsers)
	if group.Owner == "" {
		return nil
	}

	for _, username := range group.AdminUsers {
		user, err := GetUser(util.GetId(group.Owner, username))
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("the admin user: %s doesn't exist", util.GetId(group.Owner, username))
		}
	}

	return nil
}

func GetManagedGroupsByUser(owner string, username string) ([]*Group, error) {
	if owner == "" || username == "" {
		return []*Group{}, nil
	}

	groups, err := GetGroups(owner)
	if err != nil {
		return nil, err
	}

	childrenMap := map[string][]*Group{}
	for _, group := range groups {
		childrenMap[group.ParentId] = append(childrenMap[group.ParentId], group)
	}

	managed := map[string]*Group{}
	queue := make([]*Group, 0)
	for _, group := range groups {
		if util.InSlice(group.AdminUsers, username) {
			managed[group.GetId()] = group
			queue = append(queue, group)
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, child := range childrenMap[current.Name] {
			if _, ok := managed[child.GetId()]; ok {
				continue
			}
			managed[child.GetId()] = child
			queue = append(queue, child)
		}
	}

	res := make([]*Group, 0, len(managed))
	for _, group := range groups {
		if _, ok := managed[group.GetId()]; ok {
			res = append(res, group)
		}
	}
	return res, nil
}

func GetManagedGroupIdsByUser(owner string, username string) ([]string, error) {
	groups, err := GetManagedGroupsByUser(owner, username)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0, len(groups))
	for _, group := range groups {
		res = append(res, group.GetId())
	}
	return res, nil
}

func CanUserManageGroup(owner string, username string, groupName string) (bool, error) {
	groups, err := GetManagedGroupsByUser(owner, username)
	if err != nil {
		return false, err
	}

	for _, group := range groups {
		if group.Name == groupName {
			return true, nil
		}
	}

	return false, nil
}

func GetScopedTreeData(groups []*Group, owner string) []*Group {
	if len(groups) == 0 {
		return []*Group{}
	}

	groupMap := map[string]*Group{}
	for _, group := range groups {
		groupMap[group.Name] = group
	}

	var build func(parentId string) []*Group
	build = func(parentId string) []*Group {
		res := []*Group{}
		for _, group := range groups {
			if group.ParentId != parentId {
				continue
			}
			node := &Group{
				Title: group.DisplayName,
				Key:   group.Name,
				Type:  group.Type,
				Owner: group.Owner,
			}
			children := build(group.Name)
			if len(children) > 0 {
				node.Children = children
			}
			res = append(res, node)
		}
		return res
	}

	roots := []*Group{}
	for _, group := range groups {
		if group.ParentId == owner {
			roots = append(roots, group)
			continue
		}
		if _, ok := groupMap[group.ParentId]; !ok {
			roots = append(roots, group)
		}
	}

	tree := []*Group{}
	for _, group := range roots {
		node := &Group{
			Title: group.DisplayName,
			Key:   group.Name,
			Type:  group.Type,
			Owner: group.Owner,
		}
		children := build(group.Name)
		if len(children) > 0 {
			node.Children = children
		}
		tree = append(tree, node)
	}
	return tree
}

func GetManagedUsers(owner string, username string, groupName string) ([]*User, error) {
	userNames, err := getManagedUserNames(owner, username, groupName)
	if err != nil {
		return nil, err
	}
	if len(userNames) == 0 {
		return []*User{}, nil
	}

	users := []*User{}
	err = ormer.Engine.Where("owner = ?", owner).In("name", userNames).Desc("created_time").Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetManagedUserCount(owner, username, field, value, groupName string) (int64, error) {
	userNames, err := getManagedUserNames(owner, username, groupName)
	if err != nil {
		return 0, err
	}
	if len(userNames) == 0 {
		return 0, nil
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	prefixedUserTable := tableNamePrefix + "user"
	session := ormer.Engine.Table(prefixedUserTable).Where("owner = ?", owner).In("name", userNames)
	if field != "" && value != "" {
		session = session.And(fmt.Sprintf("%s.%s like ?", prefixedUserTable, util.CamelToSnakeCase(field)), "%"+value+"%")
	}
	return session.Count(&User{})
}

func GetManagedOnlineUserCount(owner, username string, isOnline int) (int64, error) {
	userNames, err := getManagedUserNames(owner, username, "")
	if err != nil {
		return 0, err
	}
	if len(userNames) == 0 {
		return 0, nil
	}

	return ormer.Engine.Where("owner = ?", owner).And("is_online = ?", isOnline).In("name", userNames).Count(&User{})
}

func GetPaginationManagedUsers(owner, username string, offset, limit int, field, value, sortField, sortOrder, groupName string) ([]*User, error) {
	userNames, err := getManagedUserNames(owner, username, groupName)
	if err != nil {
		return nil, err
	}
	if len(userNames) == 0 {
		return []*User{}, nil
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	prefixedUserTable := tableNamePrefix + "user"
	session := ormer.Engine.Table(prefixedUserTable).Where("owner = ?", owner).In("name", userNames)

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

	users := []*User{}
	err = session.Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func CanUserManageTargetUser(owner string, username string, targetUser *User) (bool, error) {
	if targetUser == nil || owner == "" || username == "" {
		return false, nil
	}
	if targetUser.Owner != owner {
		return false, nil
	}

	managedGroups, err := GetManagedGroupIdsByUser(owner, username)
	if err != nil {
		return false, err
	}
	if len(managedGroups) == 0 {
		return false, nil
	}
	return util.HaveIntersection(targetUser.Groups, managedGroups), nil
}

func CanUserManageAllGroups(owner string, username string, groupIds []string) (bool, error) {
	if len(groupIds) == 0 {
		return false, nil
	}

	managedGroups, err := GetManagedGroupIdsByUser(owner, username)
	if err != nil {
		return false, err
	}
	if len(managedGroups) == 0 {
		return false, nil
	}

	managedSet := map[string]struct{}{}
	for _, groupId := range managedGroups {
		managedSet[groupId] = struct{}{}
	}
	for _, groupId := range groupIds {
		if _, ok := managedSet[groupId]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func GetManagedGroupsForAccount(owner string, username string) ([]string, error) {
	groupIds, err := GetManagedGroupIdsByUser(owner, username)
	if err != nil {
		return nil, err
	}
	sort.Strings(groupIds)
	return groupIds, nil
}

func FilterManagedGroups(groups []*Group, field, value, sortField, sortOrder string, offset, limit int) []*Group {
	filtered := make([]*Group, 0, len(groups))
	for _, group := range groups {
		if field != "" && value != "" {
			matched := false
			groupValue := reflect.Indirect(reflect.ValueOf(group)).FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, field)
			})
			if groupValue.IsValid() {
				matched = strings.Contains(strings.ToLower(fmt.Sprintf("%v", groupValue.Interface())), strings.ToLower(value))
			}
			if !matched {
				continue
			}
		}
		filtered = append(filtered, group)
	}

	if sortField != "" {
		sort.SliceStable(filtered, func(i, j int) bool {
			left := reflect.Indirect(reflect.ValueOf(filtered[i])).FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, sortField)
			})
			right := reflect.Indirect(reflect.ValueOf(filtered[j])).FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, sortField)
			})
			leftValue := fmt.Sprintf("%v", filtered[i].CreatedTime)
			rightValue := fmt.Sprintf("%v", filtered[j].CreatedTime)
			if left.IsValid() {
				leftValue = fmt.Sprintf("%v", left.Interface())
			}
			if right.IsValid() {
				rightValue = fmt.Sprintf("%v", right.Interface())
			}
			if sortOrder == "ascend" {
				return leftValue < rightValue
			}
			return leftValue > rightValue
		})
	}

	if offset == -1 || limit == -1 || offset >= len(filtered) {
		if offset >= len(filtered) && offset != -1 {
			return []*Group{}
		}
		return filtered
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end]
}

func getManagedUserNames(owner string, username string, groupName string) ([]string, error) {
	var groupIds []string
	if groupName != "" {
		allowed, err := CanUserManageGroup(owner, username, groupName)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return []string{}, nil
		}
		groupIds = []string{util.GetId(owner, groupName)}
	} else {
		managedGroups, err := GetManagedGroupsByUser(owner, username)
		if err != nil {
			return nil, err
		}
		if len(managedGroups) == 0 {
			return []string{}, nil
		}
		groupIds = make([]string, 0, len(managedGroups))
		for _, group := range managedGroups {
			groupIds = append(groupIds, group.GetId())
		}
	}

	userNameSet := map[string]struct{}{}
	for _, groupId := range groupIds {
		userNames, err := userEnforcer.GetUserNamesByGroupName(groupId)
		if err != nil {
			return nil, err
		}
		for _, userName := range userNames {
			userNameSet[userName] = struct{}{}
		}
	}

	res := make([]string, 0, len(userNameSet))
	for userName := range userNameSet {
		res = append(res, userName)
	}
	sort.Strings(res)
	return res, nil
}
