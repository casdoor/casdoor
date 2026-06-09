package object

import (
	errors2 "errors"
	"fmt"
	"sync"

	"github.com/casbin/casbin/v2/errors"

	"github.com/casbin/casbin/v2"
	"github.com/casdoor/casdoor/util"
)

type UserGroupEnforcer struct {
	// use rbac model implement use group, the enforcer can also implement user role
	enforcer *casbin.Enforcer
	// The wrapped casbin.Enforcer is a plain (non-synced) enforcer and is NOT
	// safe for concurrent use. Casdoor shares one global UserGroupEnforcer
	// across all goroutines (many concurrent LDAP auto-sync routines plus web
	// requests). Every access must be serialized; otherwise concurrent
	// read/write of casbin's internal maps triggers Go's
	// "fatal error: concurrent map read and map write", which recover() cannot
	// catch and which crashes/restarts the whole process.
	mu sync.Mutex
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
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.checkModel()
	if err != nil {
		return false, err
	}

	return e.enforcer.AddRoleForUser(user, GetGroupWithPrefix(group))
}

func (e *UserGroupEnforcer) AddGroupsForUser(user string, groups []string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.addGroupsForUser(user, groups)
}

// addGroupsForUser mutates the enforcer without locking; callers must already
// hold e.mu.
func (e *UserGroupEnforcer) addGroupsForUser(user string, groups []string) (bool, error) {
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
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.checkModel()
	if err != nil {
		return false, err
	}

	return e.enforcer.DeleteRoleForUser(user, GetGroupWithPrefix(group))
}

func (e *UserGroupEnforcer) DeleteGroupsForUser(user string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.deleteGroupsForUser(user)
}

// deleteGroupsForUser mutates the enforcer without locking; callers must
// already hold e.mu.
func (e *UserGroupEnforcer) deleteGroupsForUser(user string) (bool, error) {
	err := e.checkModel()
	if err != nil {
		return false, err
	}

	return e.enforcer.DeleteRolesForUser(user)
}

func (e *UserGroupEnforcer) GetGroupsForUser(user string) ([]string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.checkModel()
	if err != nil {
		return nil, err
	}

	if err = e.enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	groups, err := e.enforcer.GetRolesForUser(user)
	for i, group := range groups {
		groups[i] = GetGroupWithoutPrefix(group)
	}
	return groups, err
}

func (e *UserGroupEnforcer) GetAllUsersByGroup(group string) ([]string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.getAllUsersByGroup(group)
}

// getAllUsersByGroup reads the enforcer without locking; callers must already
// hold e.mu.
func (e *UserGroupEnforcer) getAllUsersByGroup(group string) ([]string, error) {
	err := e.checkModel()
	if err != nil {
		return nil, err
	}

	if err = e.enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	users, err := e.enforcer.GetUsersForRole(GetGroupWithPrefix(group))
	if err != nil {
		if errors2.Is(err, errors.ErrNameNotFound) {
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
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.checkModel()
	if err != nil {
		return nil, err
	}

	userIds, err := e.getAllUsersByGroup(groupName)
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
	e.mu.Lock()
	defer e.mu.Unlock()

	err := e.checkModel()
	if err != nil {
		return false, err
	}

	_, err = e.deleteGroupsForUser(user)
	if err != nil {
		return false, err
	}

	affected, err := e.addGroupsForUser(user, groups)
	if err != nil {
		return false, err
	}

	return affected, nil
}
