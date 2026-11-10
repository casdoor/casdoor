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
	"sort"
	"strings"
	"time"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
)

// SodConflict is a user holding two or more roles that a separation-of-duty rule keeps apart
type SodConflict struct {
	User        string   `json:"user"`
	DisplayName string   `json:"displayName"`
	Rule        string   `json:"rule"`
	Roles       []string `json:"roles"`
}

type governanceUser struct {
	Id          string
	DisplayName string
	Groups      []string
}

// getUserRoleIds lists the enabled roles of each user of the organization, given directly or through a group.
// changedRole replaces the saved role of the same id, changedUser the saved groups of that user, so that a change can
// be checked before it is saved.
func getUserRoleIds(owner string, changedRole *Role, changedUser *User) (map[string][]string, map[string]*governanceUser, error) {
	roles, err := GetRoles(owner)
	if err != nil {
		return nil, nil, err
	}
	if changedRole != nil {
		replaced := false
		for i, role := range roles {
			if role.GetId() == changedRole.GetId() {
				roles[i] = changedRole
				replaced = true
			}
		}
		if !replaced {
			roles = append(roles, changedRole)
		}
	}

	users := []*User{}
	err = ormer.Engine.Cols("owner", "name", "display_name", "groups").Where("owner = ?", owner).Find(&users)
	if err != nil {
		return nil, nil, err
	}
	userMap := map[string]*governanceUser{}
	for _, user := range users {
		groups := user.Groups
		if changedUser != nil && changedUser.GetId() == user.GetId() {
			groups = changedUser.Groups
		}
		userMap[user.GetId()] = &governanceUser{Id: user.GetId(), DisplayName: user.DisplayName, Groups: groups}
	}

	res := map[string][]string{}
	for _, role := range roles {
		if !role.IsEnabled {
			continue
		}
		holders := map[string]bool{}
		for _, userId := range role.Users {
			holders[userId] = true
		}
		if len(role.Groups) > 0 {
			for _, user := range userMap {
				for _, group := range user.Groups {
					if util.InSlice(role.Groups, group) {
						holders[user.Id] = true
					}
				}
			}
		}
		for userId := range holders {
			res[userId] = append(res[userId], role.GetId())
		}
	}
	return res, userMap, nil
}

func findSodConflicts(policy *EnterprisePolicy, userRoleIds map[string][]string, userMap map[string]*governanceUser, onlyUsers map[string]bool) []*SodConflict {
	res := []*SodConflict{}
	for userId, roleIds := range userRoleIds {
		if onlyUsers != nil && !onlyUsers[userId] {
			continue
		}
		for _, rule := range policy.SodRules {
			held := []string{}
			for _, roleId := range rule.Roles {
				if util.InSlice(roleIds, roleId) {
					held = append(held, roleId)
				}
			}
			if len(held) >= 2 {
				conflict := &SodConflict{User: userId, Rule: rule.Name, Roles: held}
				if user, ok := userMap[userId]; ok {
					conflict.DisplayName = user.DisplayName
				}
				res = append(res, conflict)
			}
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].User != res[j].User {
			return res[i].User < res[j].User
		}
		return res[i].Rule < res[j].Rule
	})
	return res
}

// GetSodConflicts lists the users of the organization who already hold roles that a rule keeps apart, such as the
// ones given before the rule was added
func GetSodConflicts(owner string) ([]*SodConflict, error) {
	policy := GetCachedEnterprisePolicy(owner)
	if policy == nil || len(policy.SodRules) == 0 {
		return []*SodConflict{}, nil
	}
	userRoleIds, userMap, err := getUserRoleIds(owner, nil, nil)
	if err != nil {
		return nil, err
	}
	return findSodConflicts(policy, userRoleIds, userMap, nil), nil
}

func formatSodConflict(conflict *SodConflict, lang string) error {
	return fmt.Errorf(i18n.Translate(lang, "enterprise:Separation of duties: the user %s would hold the roles %s together, which the rule %s keeps apart"), conflict.User, strings.Join(conflict.Roles, ", "), conflict.Rule)
}

// CheckRoleSod refuses a role change that would give a user conflicting roles. Only the users the change touches are
// checked, so conflicts from before the rule don't block unrelated changes.
func CheckRoleSod(role *Role, lang string) error {
	if role == nil {
		return nil
	}
	policy := GetCachedEnterprisePolicy(role.Owner)
	if policy == nil || len(policy.SodRules) == 0 {
		return nil
	}

	oldRole, err := getRole(role.Owner, role.Name)
	if err != nil {
		return err
	}
	userRoleIds, userMap, err := getUserRoleIds(role.Owner, role, nil)
	if err != nil {
		return err
	}

	touched := map[string]bool{}
	for _, userId := range role.Users {
		if oldRole == nil || !util.InSlice(oldRole.Users, userId) || !oldRole.IsEnabled {
			touched[userId] = true
		}
	}
	for _, user := range userMap {
		for _, group := range user.Groups {
			if util.InSlice(role.Groups, group) && (oldRole == nil || !util.InSlice(oldRole.Groups, group) || !oldRole.IsEnabled) {
				touched[user.Id] = true
			}
		}
	}
	if len(touched) == 0 {
		return nil
	}

	conflicts := findSodConflicts(policy, userRoleIds, userMap, touched)
	if len(conflicts) > 0 {
		return formatSodConflict(conflicts[0], lang)
	}
	return nil
}

// CheckUserSod refuses moving a user into groups whose roles conflict
func CheckUserSod(oldUser *User, user *User, lang string) error {
	if oldUser == nil || user == nil || util.StructToJson(oldUser.Groups) == util.StructToJson(user.Groups) {
		return nil
	}
	policy := GetCachedEnterprisePolicy(user.Owner)
	if policy == nil || len(policy.SodRules) == 0 {
		return nil
	}

	userRoleIds, userMap, err := getUserRoleIds(user.Owner, nil, user)
	if err != nil {
		return err
	}
	conflicts := findSodConflicts(policy, userRoleIds, userMap, map[string]bool{user.GetId(): true})
	if len(conflicts) > 0 {
		return formatSodConflict(conflicts[0], lang)
	}
	return nil
}

// OrphanAccount is an account that nobody seems to use or that should not have access any more
type OrphanAccount struct {
	User           string   `json:"user"`
	DisplayName    string   `json:"displayName"`
	IsAdmin        bool     `json:"isAdmin"`
	IsForbidden    bool     `json:"isForbidden"`
	CreatedTime    string   `json:"createdTime"`
	LastSigninTime string   `json:"lastSigninTime"`
	Reasons        []string `json:"reasons"`
	Roles          []string `json:"roles"`
}

// GetOrphanAccounts finds the accounts of the organization that never signed in or didn't sign in for the given days
// ("never-signed-in", "inactive"), and the disabled accounts that still hold roles ("disabled-with-roles")
func GetOrphanAccounts(owner string, days int) ([]*OrphanAccount, error) {
	if days <= 0 {
		days = 90
	}
	cutoff := time.Now().AddDate(0, 0, -days)

	users := []*User{}
	err := ormer.Engine.Cols("owner", "name", "display_name", "is_admin", "is_forbidden", "is_deleted", "created_time", "last_signin_time", "groups").Where("owner = ?", owner).Find(&users)
	if err != nil {
		return nil, err
	}
	userRoleIds, _, err := getUserRoleIds(owner, nil, nil)
	if err != nil {
		return nil, err
	}

	isBefore := func(value string) bool {
		t, err := time.Parse(time.RFC3339, value)
		return err == nil && t.Before(cutoff)
	}

	res := []*OrphanAccount{}
	for _, user := range users {
		if user.IsDeleted {
			continue
		}
		roles := userRoleIds[user.GetId()]
		reasons := []string{}
		if user.IsForbidden {
			if len(roles) > 0 {
				reasons = append(reasons, "disabled-with-roles")
			}
		} else if user.LastSigninTime == "" {
			if isBefore(user.CreatedTime) {
				reasons = append(reasons, "never-signed-in")
			}
		} else if isBefore(user.LastSigninTime) {
			reasons = append(reasons, "inactive")
		}
		if len(reasons) == 0 {
			continue
		}
		res = append(res, &OrphanAccount{
			User:           user.GetId(),
			DisplayName:    user.DisplayName,
			IsAdmin:        user.IsAdmin,
			IsForbidden:    user.IsForbidden,
			CreatedTime:    user.CreatedTime,
			LastSigninTime: user.LastSigninTime,
			Reasons:        reasons,
			Roles:          roles,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].LastSigninTime < res[j].LastSigninTime
	})
	return res, nil
}
