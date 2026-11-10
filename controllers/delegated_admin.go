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

package controllers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// DelegatedScopeKey is where the authorization filter leaves the scope of a delegated admin's request
const DelegatedScopeKey = "delegatedScope"

// getDelegatedScope is the scope of the current request when it is made by a delegated admin, nil otherwise
func (c *ApiController) getDelegatedScope() *object.DelegatedScope {
	scope, ok := c.Ctx.Input.GetData(DelegatedScopeKey).(*object.DelegatedScope)
	if !ok || scope == nil || scope.User == "" || c.IsAdmin() {
		return nil
	}
	if c.GetSessionUsername() != util.GetId(scope.Organization, scope.User) {
		return nil
	}
	return scope
}

func (c *ApiController) isSessionUser(user *object.User) bool {
	return user != nil && c.GetSessionUsername() == user.GetId()
}

func (c *ApiController) unauthorizedError() error {
	return errors.New(c.T("auth:Unauthorized operation"))
}

func (c *ApiController) getDelegatedUsers(scope *object.DelegatedScope) {
	groupName := c.Ctx.Input.Query("groupName")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		users, err := object.GetMaskedUsers(scope.GetScopedUsers(-1, -1, "", "", "", "", groupName))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.maskPersonalInfo(users...)
		c.ResponseOk(users)
		return
	}

	pageSize := util.ParseInt(limit)
	count, err := scope.GetScopedUserCount(field, value, groupName, nil)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.NewPaginator(c.Ctx.Request, pageSize, count)
	users, err := object.GetMaskedUsers(scope.GetScopedUsers(paginator.Offset(), pageSize, field, value, sortField, sortOrder, groupName))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.maskPersonalInfo(users...)
	c.ResponseOk(users, paginator.Nums())
}

// checkDelegatedUserUpdate checks a user a delegated admin changes: an administered user who keeps their
// organization, name and ID, doesn't become an admin and stays in the administered groups
func (c *ApiController) checkDelegatedUserUpdate(scope *object.DelegatedScope, oldUser *object.User, user *object.User, columns string) error {
	if !scope.CanManageUser(oldUser) {
		return c.unauthorizedError()
	}
	if user.Owner != oldUser.Owner || user.Name != oldUser.Name || (user.Id != "" && user.Id != oldUser.Id) || user.IsAdmin {
		return c.unauthorizedError()
	}
	if !scope.CheckGroupChange(oldUser.Groups, user.Groups) {
		return errors.New(c.T("enterprise:A delegated admin can only put users in the groups they administer"))
	}
	isPasswordChanged := user.Password != "" && user.Password != "***" && user.Password != oldUser.Password
	if (isPasswordChanged || strings.Contains(strings.ToLower(columns), "password")) && !scope.Can(object.DelegatedActionPasswords) {
		return c.unauthorizedError()
	}
	return nil
}

// checkDelegatedUserAdd checks a user a delegated admin adds, with no group it goes to the delegated admin's first group
func (c *ApiController) checkDelegatedUserAdd(scope *object.DelegatedScope, user *object.User) error {
	if user.Owner != scope.Organization || user.IsAdmin || user.Name == scope.User || scope.IsProtectedName(user.Name) {
		return c.unauthorizedError()
	}
	// only a merge lists the accounts merged into a user, the applications hand their data over by it
	delete(user.Properties, object.UserPropertiesMergedUsers)
	if len(user.Groups) == 0 {
		if len(scope.RootGroups) == 0 {
			return errors.New(c.T("enterprise:A delegated admin can only put users in the groups they administer"))
		}
		user.Groups = []string{util.GetId(scope.Organization, scope.RootGroups[0])}
	}
	for _, group := range user.Groups {
		if !scope.HasGroup(group) {
			return errors.New(c.T("enterprise:A delegated admin can only put users in the groups they administer"))
		}
	}
	return nil
}

func (c *ApiController) getDelegatedGroups(scope *object.DelegatedScope) {
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		groups, err := object.GetGroups(scope.Organization)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if c.Ctx.Input.Query("withTree") == "true" {
			c.ResponseOk(scope.GetGroupTree(groups))
			return
		}
		groups = scope.FilterGroups(groups)
		err = object.ExtendGroupsWithUsers(groups)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(groups)
		return
	}

	pageSize := util.ParseInt(limit)
	count, err := scope.GetScopedGroupCount(field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.NewPaginator(c.Ctx.Request, pageSize, count)
	groups, err := scope.GetPaginationScopedGroups(paginator.Offset(), pageSize, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	groupsParentMap, err := object.GetGroupsParentMap(groups)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	groupIdsHaveChildren, err := object.GetGroupIdsHaveChildren(groups)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	for _, group := range groups {
		group.HaveChildren = groupIdsHaveChildren[group.GetId()]
		if parent, ok := groupsParentMap[fmt.Sprintf("%s/%s", group.Owner, group.ParentId)]; ok {
			group.ParentName = parent.DisplayName
		}
	}
	err = object.ExtendGroupsWithUsers(groups)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(groups, paginator.Nums())
}

func (c *ApiController) getDelegatedApplications(scope *object.DelegatedScope) {
	applications := []*object.Application{}
	if scope.Can(object.DelegatedActionApplications) {
		all, err := object.GetOrganizationApplications("admin", scope.Organization)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		for _, application := range all {
			if scope.HasApplication(application) {
				applications = append(applications, application)
			}
		}
	}

	limit := c.Ctx.Input.Query("pageSize")
	if limit == "" || c.Ctx.Input.Query("p") == "" {
		c.ResponseOk(applications)
		return
	}
	pageSize := util.ParseInt(limit)
	paginator := pagination.NewPaginator(c.Ctx.Request, pageSize, int64(len(applications)))
	start := paginator.Offset()
	if start > len(applications) {
		start = len(applications)
	}
	end := start + pageSize
	if pageSize <= 0 || end > len(applications) {
		end = len(applications)
	}
	c.ResponseOk(applications[start:end], paginator.Nums())
}

// checkDelegatedApplicationUpdate checks an application a delegated admin changes: one of their applications that
// keeps its organization and name
func (c *ApiController) checkDelegatedApplicationUpdate(scope *object.DelegatedScope, id string, application *object.Application) error {
	oldApplication, err := object.GetApplication(id)
	if err != nil {
		return err
	}
	if oldApplication == nil || !scope.HasApplication(oldApplication) {
		return c.unauthorizedError()
	}
	if application.Organization != oldApplication.Organization || application.Name != oldApplication.Name || (application.Owner != "" && application.Owner != oldApplication.Owner) {
		return c.unauthorizedError()
	}
	return nil
}
