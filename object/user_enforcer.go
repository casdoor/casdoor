package object

import (
	"fmt"

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

func (e *UserGroupEnforcer) checkModel() error {
	if _, ok := e.enforcer.GetModel()["g"]; !ok {
		return fmt.Errorf("The Casbin model used by enforcer doesn't support RBAC (\"[role_definition]\" section not found), please use a RBAC enabled Casbin model for the enforcer")
	}
	return nil
}

func (e *UserGroupEnforcer) AddGroupForUser(user string, group string) (bool, error) {
	err := e.checkModel()
	if err != nil {
		return false, err
	}

	return e.enforcer.AddRoleForUser(user, GetGroupWithPrefix(group))
}

func (e *UserGroupEnforcer) AddGroupsForUser(user string, groups []string) (bool, error) {
	err := e.checkModel()
	if err != nil {
		return false, err
	}

	g := make([]string, len(groups))
	for i, group := range groups {
		g[i] = GetGroupWithPrefix(group)
	}
	return e.enforcer.AddRolesForUser(user, g)
}

func (e *UserGroupEnforcer) DeleteGroupForUser(user string, group string) (bool, error) {
	err := e.checkModel()
	if err != nil {
		return false, err
	}

	return e.enforcer.DeleteRoleForUser(user, GetGroupWithPrefix(group))
}

func (e *UserGroupEnforcer) DeleteGroupsForUser(user string) (bool, error) {
	err := e.checkModel()
	if err != nil {
		return false, err
	}

	return e.enforcer.DeleteRolesForUser(user)
}

func (e *UserGroupEnforcer) GetGroupsForUser(user string) ([]string, error) {
	err := e.checkModel()
	if err != nil {
		return nil, err
	}

	groups, err := e.enforcer.GetRolesForUser(user)
	for i, group := range groups {
		groups[i] = GetGroupWithoutPrefix(group)
	}
	return groups, err
}

func (e *UserGroupEnforcer) GetAllUsersByGroup(group string) ([]string, error) {
	err := e.checkModel()
	if err != nil {
		return nil, err
	}

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
	err := e.checkModel()
	if err != nil {
		return nil, err
	}

	userIds, err := e.GetAllUsersByGroup(groupName)
	if err != nil {
		return nil, err
	}

	names := []string{}
	for _, userId := range userIds {
		_, name := util.GetOwnerAndNameFromIdNoCheck(userId)
		names = append(names, name)
	}

	return names, nil
}

func (e *UserGroupEnforcer) UpdateGroupsForUser(user string, groups []string) (bool, error) {
	err := e.checkModel()
	if err != nil {
		return false, err
	}

	_, err = e.DeleteGroupsForUser(user)
	if err != nil {
		return false, err
	}

	affected, err := e.AddGroupsForUser(user, groups)
	if err != nil {
		return false, err
	}

	return affected, nil
}
