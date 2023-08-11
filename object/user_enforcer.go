package object

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/errors"
	"github.com/casdoor/casdoor/util"
)

type UserGroupEnforcer struct {
	// use rbac model implement use group, the enforcer can also implement user role
	enforcer *casbin.Enforcer
}

func NewUserGroupEnforcer(enforcer *casbin.Enforcer) *UserGroupEnforcer {
	return &UserGroupEnforcer{
		enforcer: enforcer,
	}
}

func (e *UserGroupEnforcer) AddGroupForUser(user string, group string) (bool, error) {
	return e.enforcer.AddRoleForUser(user, GetGroupWithPrefix(group))
}

func (e *UserGroupEnforcer) AddGroupsForUser(user string, groups []string) (bool, error) {
	g := make([]string, len(groups))
	for i, group := range groups {
		g[i] = GetGroupWithPrefix(group)
	}
	return e.enforcer.AddRolesForUser(user, g)
}

func (e *UserGroupEnforcer) DeleteGroupForUser(user string, group string) (bool, error) {
	return e.enforcer.DeleteRoleForUser(user, GetGroupWithPrefix(group))
}

func (e *UserGroupEnforcer) DeleteGroupsForUser(user string) (bool, error) {
	return e.enforcer.DeleteRolesForUser(user)
}

func (e *UserGroupEnforcer) GetGroupsForUser(user string) ([]string, error) {
	groups, err := e.enforcer.GetRolesForUser(user)
	for i, group := range groups {
		groups[i] = GetGroupWithoutPrefix(group)
	}
	return groups, err
}

func (e *UserGroupEnforcer) GetAllUsersByGroup(group string) ([]string, error) {
	users, err := e.enforcer.GetUsersForRole(GetGroupWithPrefix(group))
	if err != nil {
		if err == errors.ERR_NAME_NOT_FOUND {
			return []string{}, nil
		}
		return nil, err
	}
	return users, nil
}

func GetGroupWithPrefix(group string) string {
	return "group:" + group
}

func GetGroupWithoutPrefix(group string) string {
	return group[len("group:"):]
}

func (e *UserGroupEnforcer) GetUserNamesByGroupName(groupName string) ([]string, error) {
	var names []string

	userIds, err := e.GetAllUsersByGroup(groupName)
	if err != nil {
		return nil, err
	}

	for _, userId := range userIds {
		_, name := util.GetOwnerAndNameFromIdNoCheck(userId)
		names = append(names, name)
	}

	return names, nil
}

func (e *UserGroupEnforcer) UpdateGroupsForUser(user string, groups []string) (bool, error) {
	_, err := e.DeleteGroupsForUser(user)
	if err != nil {
		return false, err
	}

	affected, err := e.AddGroupsForUser(user, groups)
	if err != nil {
		return false, err
	}

	return affected, nil
}
