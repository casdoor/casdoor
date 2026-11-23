// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"errors"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/builder"
)

const (
	DelegatedActionUsers        = "users"
	DelegatedActionGroups       = "groups"
	DelegatedActionApplications = "applications"
	DelegatedActionPasswords    = "passwords"
)

var DelegatedAdminActions = []string{DelegatedActionUsers, DelegatedActionGroups, DelegatedActionApplications, DelegatedActionPasswords}

// DelegatedScope is what a delegated admin administers, resolved when a request comes in
type DelegatedScope struct {
	Organization string   `json:"organization"`
	User         string   `json:"user"`
	RootGroups   []string `json:"rootGroups"`
	Groups       []string `json:"groups"`
	Applications []string `json:"applications"`
	Actions      []string `json:"actions"`
	groupSet     map[string]bool
	rootSet      map[string]bool
	protected    map[string]bool
}

// DelegatedAdminInfo tells the console what the signed-in user administers, Active is false when the license no
// longer covers the delegated administration
type DelegatedAdminInfo struct {
	Active       bool     `json:"active"`
	RootGroups   []string `json:"rootGroups"`
	Groups       []string `json:"groups"`
	Applications []string `json:"applications"`
	Actions      []string `json:"actions"`
	Message      string   `json:"message,omitempty"`
}

// GetDelegatedScope returns what the user administers as a delegated admin, or nil when the user isn't one, is an
// organization admin anyway, or the license doesn't cover the delegated administration
func GetDelegatedScope(user *object.User) (*DelegatedScope, error) {
	if user == nil || user.IsAdmin || user.IsGlobalAdmin() || user.IsForbidden || user.IsDeleted {
		return nil, nil
	}
	policy := GetCachedEnterprisePolicy(user.Owner)
	delegatedAdmin := policy.GetDelegatedAdmin(user.Name)
	if delegatedAdmin == nil || !edition.Has(edition.Enterprise) {
		return nil, nil
	}

	groups, err := object.GetGroups(user.Owner)
	if err != nil {
		return nil, err
	}
	children := map[string][]string{}
	existing := map[string]bool{}
	for _, group := range groups {
		existing[group.Name] = true
		if group.ParentId != "" && group.ParentId != group.Name {
			children[group.ParentId] = append(children[group.ParentId], group.Name)
		}
	}

	scope := &DelegatedScope{
		Organization: user.Owner,
		User:         user.Name,
		RootGroups:   []string{},
		Groups:       []string{},
		Applications: []string{},
		Actions:      []string{},
		groupSet:     map[string]bool{},
		rootSet:      map[string]bool{},
		protected:    policy.GetProtectedUsers(),
	}
	for _, action := range delegatedAdmin.Actions {
		if util.InSlice(DelegatedAdminActions, action) && !util.InSlice(scope.Actions, action) {
			scope.Actions = append(scope.Actions, action)
		}
	}
	scope.Applications = append(scope.Applications, delegatedAdmin.Applications...)

	queue := []string{}
	for _, name := range delegatedAdmin.Groups {
		if existing[name] && !scope.rootSet[name] {
			scope.rootSet[name] = true
			scope.RootGroups = append(scope.RootGroups, name)
			queue = append(queue, name)
		}
	}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if scope.groupSet[name] {
			continue
		}
		scope.groupSet[name] = true
		scope.Groups = append(scope.Groups, name)
		queue = append(queue, children[name]...)
	}
	return scope, nil
}

// GetDelegatedAdminInfo is the delegated administration of the user for the console, also when the license has
// lapsed so that the console can say why the user can't administer anything
func GetDelegatedAdminInfo(user *object.User) (*DelegatedAdminInfo, error) {
	if user == nil || user.IsAdmin || user.IsGlobalAdmin() {
		return nil, nil
	}
	organization, err := object.LookupOrganization("admin", user.Owner)
	if err != nil || organization == nil {
		return nil, err
	}
	delegatedAdmin := organization.EnterprisePolicy.GetDelegatedAdmin(user.Name)
	if delegatedAdmin == nil {
		return nil, nil
	}

	info := &DelegatedAdminInfo{RootGroups: []string{}, Groups: []string{}, Applications: []string{}, Actions: []string{}}
	if err = edition.Require(edition.Enterprise, edition.FeatureDelegatedAdmin); err != nil {
		info.Message = err.Error()
		return info, nil
	}
	scope, err := GetDelegatedScope(user)
	if err != nil || scope == nil {
		return info, err
	}
	info.Active = true
	info.RootGroups = scope.RootGroups
	info.Groups = scope.Groups
	info.Applications = scope.Applications
	info.Actions = scope.Actions
	return info, nil
}

func (scope *DelegatedScope) Can(action string) bool {
	return scope != nil && util.InSlice(scope.Actions, action)
}

// HasGroup tells whether the group, a name or an "owner/name" ID, is administered
func (scope *DelegatedScope) HasGroup(group string) bool {
	if scope == nil {
		return false
	}
	if strings.Contains(group, "/") {
		owner, name := util.GetOwnerAndNameFromIdNoCheck(group)
		if owner != scope.Organization {
			return false
		}
		group = name
	}
	return scope.groupSet[group]
}

func (scope *DelegatedScope) IsRootGroup(group string) bool {
	return scope != nil && scope.rootSet[group]
}

func (scope *DelegatedScope) HasApplication(application *object.Application) bool {
	return scope != nil && application != nil && application.Organization == scope.Organization && util.InSlice(scope.Applications, application.Name)
}

// CanManageUser tells whether the delegated admin administers the user: a user of the organization in at least one
// administered group, who is not an admin, a delegated admin or one of the three admins, and not themselves
func (scope *DelegatedScope) CanManageUser(user *object.User) bool {
	if scope == nil || user == nil || user.Owner != scope.Organization || user.Name == scope.User {
		return false
	}
	if user.IsAdmin || user.IsGlobalAdmin() || scope.protected[user.Name] {
		return false
	}
	return scope.hasAnyGroup(user.Groups)
}

func (scope *DelegatedScope) hasAnyGroup(groups []string) bool {
	for _, group := range groups {
		if scope.HasGroup(group) {
			return true
		}
	}
	return false
}

// IsProtectedName tells whether a new user must not get the name, as it is configured as an admin of a kind
func (scope *DelegatedScope) IsProtectedName(name string) bool {
	return scope != nil && scope.protected[name]
}

// CheckGroupChange checks the groups the delegated admin gives a user: the user stays in at least one administered
// group, joins no other group and keeps the groups the delegated admin doesn't administer
func (scope *DelegatedScope) CheckGroupChange(oldGroups []string, newGroups []string) bool {
	if !scope.hasAnyGroup(newGroups) {
		return false
	}
	for _, group := range newGroups {
		if !scope.HasGroup(group) && !util.InSlice(oldGroups, group) {
			return false
		}
	}
	for _, group := range oldGroups {
		if !scope.HasGroup(group) && !util.InSlice(newGroups, group) {
			return false
		}
	}
	return true
}

// GetScopedUserNames are the names of the users the delegated admin administers, optionally only in one group
func (scope *DelegatedScope) GetScopedUserNames(groupName string) ([]string, error) {
	groups := scope.Groups
	if groupName != "" {
		if !scope.HasGroup(groupName) {
			return []string{}, nil
		}
		groups = []string{groupName}
	}

	candidates := map[string]bool{}
	for _, group := range groups {
		names, err := object.GetUserEnforcer().GetUserNamesByGroupName(util.GetId(scope.Organization, group))
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			candidates[name] = true
		}
	}
	if len(candidates) == 0 {
		return []string{}, nil
	}

	names := make([]string, 0, len(candidates))
	for name := range candidates {
		names = append(names, name)
	}
	users := []*object.User{}
	for start := 0; start < len(names); start += 500 {
		end := start + 500
		if end > len(names) {
			end = len(names)
		}
		batch := []*object.User{}
		err := object.GetOrmer().Engine.Cols("owner", "name", "groups", "is_admin").Where("owner = ?", scope.Organization).In("name", names[start:end]).Find(&batch)
		if err != nil {
			return nil, err
		}
		users = append(users, batch...)
	}

	res := []string{}
	for _, user := range users {
		if !scope.CanManageUser(user) {
			continue
		}
		if groupName != "" && !util.InSlice(user.Groups, util.GetId(scope.Organization, groupName)) {
			continue
		}
		res = append(res, user.Name)
	}
	return res, nil
}

// GetScopedUserCount counts the administered users, filtered like the user list, cond may further limit them
func (scope *DelegatedScope) GetScopedUserCount(field string, value string, groupName string, cond builder.Cond) (int64, error) {
	names, err := scope.GetScopedUserNames(groupName)
	if err != nil || len(names) == 0 {
		return 0, err
	}
	session := object.GetOrmer().Engine.Where("owner = ?", scope.Organization).In("name", names)
	if field != "" && value != "" && util.FilterField(field) {
		session = session.And(fmt.Sprintf("%s like ?", util.CamelToSnakeCase(field)), "%"+value+"%")
	}
	if cond != nil {
		session = session.And(cond)
	}
	return session.Count(&object.User{})
}

// GetScopedUsers lists the administered users, a page of them when limit is not -1, cond may further limit them
func (scope *DelegatedScope) GetScopedUsers(offset int, limit int, field string, value string, sortField string, sortOrder string, groupName string, cond builder.Cond) ([]*object.User, error) {
	users := []*object.User{}
	names, err := scope.GetScopedUserNames(groupName)
	if err != nil || len(names) == 0 {
		return users, err
	}
	session := object.GetOrmer().Engine.Where("owner = ?", scope.Organization).In("name", names)
	if offset != -1 && limit != -1 {
		session = session.Limit(limit, offset)
	}
	if field != "" && value != "" && util.FilterField(field) {
		session = session.And(fmt.Sprintf("%s like ?", util.CamelToSnakeCase(field)), "%"+value+"%")
	}
	if cond != nil {
		session = session.And(cond)
	}
	if sortField == "" || sortOrder == "" || !util.FilterField(sortField) {
		sortField = "created_time"
	}
	if sortOrder == "ascend" {
		session = session.Asc(util.CamelToSnakeCase(sortField))
	} else {
		session = session.Desc(util.CamelToSnakeCase(sortField))
	}
	err = session.Find(&users)
	return users, err
}

// FilterGroups keeps the administered groups
func (scope *DelegatedScope) FilterGroups(groups []*object.Group) []*object.Group {
	res := []*object.Group{}
	for _, group := range groups {
		if group.Owner == scope.Organization && scope.HasGroup(group.Name) {
			res = append(res, group)
		}
	}
	return res
}

// GetGroupTree is the tree of the administered groups with the delegated admin's groups at the top
func (scope *DelegatedScope) GetGroupTree(groups []*object.Group) []*object.Group {
	groups = scope.FilterGroups(groups)
	tree := []*object.Group{}
	for _, group := range groups {
		if !scope.IsRootGroup(group.Name) {
			continue
		}
		node := &object.Group{
			Title:       group.DisplayName,
			Key:         group.Name,
			Type:        group.Type,
			Owner:       group.Owner,
			Name:        group.Name,
			DisplayName: group.DisplayName,
		}
		children := object.ConvertToTreeData(groups, group.Name)
		if len(children) > 0 {
			node.Children = children
		}
		tree = append(tree, node)
	}
	return tree
}

// GetScopedGroupCount counts the administered groups, filtered like the group list
func (scope *DelegatedScope) GetScopedGroupCount(field string, value string) (int64, error) {
	if len(scope.Groups) == 0 {
		return 0, nil
	}
	session := object.GetSession(scope.Organization, -1, -1, field, value, "", "")
	return session.In("name", scope.Groups).Count(&object.Group{})
}

// GetPaginationScopedGroups lists a page of the administered groups
func (scope *DelegatedScope) GetPaginationScopedGroups(offset int, limit int, field string, value string, sortField string, sortOrder string) ([]*object.Group, error) {
	groups := []*object.Group{}
	if len(scope.Groups) == 0 {
		return groups, nil
	}
	session := object.GetSession(scope.Organization, offset, limit, field, value, sortField, sortOrder)
	err := session.In("name", scope.Groups).Find(&groups)
	return groups, err
}

func (scope *DelegatedScope) isDescendantOrSelf(group string, ancestor string, groups map[string]*object.Group) bool {
	for i := 0; i < 1000 && group != ""; i++ {
		if group == ancestor {
			return true
		}
		parent, ok := groups[group]
		if !ok || parent.ParentId == group {
			return false
		}
		group = parent.ParentId
	}
	return false
}

// CheckGroupUpdate checks a group the delegated admin adds (oldGroup is nil) or changes: it stays in the
// administered subtree, a top group of the delegated admin keeps its place and name, and no group moves under itself
func (scope *DelegatedScope) CheckGroupUpdate(oldGroup *object.Group, newGroup *object.Group, lang string) error {
	unauthorized := errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
	if newGroup == nil || newGroup.Owner != scope.Organization {
		return unauthorized
	}
	if oldGroup == nil {
		if newGroup.IsTopGroup || !scope.HasGroup(newGroup.ParentId) {
			return errors.New(i18n.Translate(lang, "enterprise:A delegated admin can only add groups under the groups they administer"))
		}
		return nil
	}

	if oldGroup.Owner != scope.Organization || !scope.HasGroup(oldGroup.Name) || newGroup.Name != oldGroup.Name {
		return unauthorized
	}
	if scope.IsRootGroup(oldGroup.Name) {
		if newGroup.ParentId != oldGroup.ParentId || newGroup.IsTopGroup != oldGroup.IsTopGroup {
			return unauthorized
		}
		return nil
	}
	if newGroup.IsTopGroup || !scope.HasGroup(newGroup.ParentId) {
		return errors.New(i18n.Translate(lang, "enterprise:A delegated admin can only add groups under the groups they administer"))
	}
	groups, err := object.GetGroups(scope.Organization)
	if err != nil {
		return err
	}
	groupMap := map[string]*object.Group{}
	for _, group := range groups {
		groupMap[group.Name] = group
	}
	if scope.isDescendantOrSelf(newGroup.ParentId, oldGroup.Name, groupMap) {
		return unauthorized
	}
	return nil
}

// ValidateDelegatedAdmins checks the delegated admins an organization admin saves
func ValidateDelegatedAdmins(organization string, policy *object.EnterprisePolicy, lang string) error {
	if policy == nil || len(policy.DelegatedAdmins) == 0 {
		return nil
	}
	seen := map[string]bool{}
	for _, delegatedAdmin := range policy.DelegatedAdmins {
		if delegatedAdmin == nil || delegatedAdmin.User == "" {
			return errors.New(i18n.Translate(lang, "enterprise:A delegated admin needs a user"))
		}
		if seen[delegatedAdmin.User] {
			return fmt.Errorf(i18n.Translate(lang, "enterprise:The user: %s is listed twice as a delegated admin"), delegatedAdmin.User)
		}
		seen[delegatedAdmin.User] = true
		if organization == "" {
			continue
		}

		user, err := object.LookupUser(organization, delegatedAdmin.User)
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf(i18n.Translate(lang, "general:The user: %s doesn't exist"), util.GetId(organization, delegatedAdmin.User))
		}
		if user.IsAdmin {
			return fmt.Errorf(i18n.Translate(lang, "enterprise:The user: %s is already an admin of the organization"), delegatedAdmin.User)
		}
		if len(delegatedAdmin.Groups) == 0 && len(delegatedAdmin.Applications) == 0 {
			return fmt.Errorf(i18n.Translate(lang, "enterprise:The delegated admin: %s needs at least one group or application"), delegatedAdmin.User)
		}
		for _, action := range delegatedAdmin.Actions {
			if !util.InSlice(DelegatedAdminActions, action) {
				return fmt.Errorf("unknown delegated admin action: %s", action)
			}
		}
		for _, name := range delegatedAdmin.Groups {
			group, err := object.LookupGroup(organization, name)
			if err != nil {
				return err
			}
			if group == nil {
				return fmt.Errorf(i18n.Translate(lang, "enterprise:The group: %s doesn't exist"), name)
			}
		}
		for _, name := range delegatedAdmin.Applications {
			application, err := object.LookupApplication("admin", name)
			if err != nil {
				return err
			}
			if application == nil || application.Organization != organization {
				return fmt.Errorf(i18n.Translate(lang, "enterprise:The application: %s doesn't belong to the organization"), name)
			}
		}
	}
	return nil
}
