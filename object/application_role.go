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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
)

// AppRole is an application-scoped entitlement. It changes nothing in Casdoor: a user does not
// join a group or role through it and gets no Casdoor privilege. It only decides what the tokens
// and the userinfo of its own application say about the users an application grant assigns it to.
//
// Groups and Roles name groups and roles of the application's organization: the role adds the
// value each of them has in the application's "groups" and "roles" claims, exactly as for a user
// of the organization who belongs to them, so a relying party authorizes the user the same way.
// Claims are literal custom claims for relying parties that authorize on their own claim names.
type AppRole struct {
	Name         string     `json:"name"`
	DisplayName  string     `json:"displayName"`
	Description  string     `json:"description"`
	IsPrivileged bool       `json:"isPrivileged"`
	Groups       []string   `json:"groups"`
	Roles        []string   `json:"roles"`
	Claims       []*JwtItem `json:"claims"`
}

// ApplicationEntitlements is what the application grants matching a user add to the user's
// claims for one application, with the grants and roles they come from.
type ApplicationEntitlements struct {
	Grants []*ApplicationGrant `json:"grants"`
	Roles  []*AppRole          `json:"roles"`
	Groups []string            `json:"groups"`
	// RoleNames are the roles of the application's organization, as the "roles" claim names them
	RoleNames []string               `json:"roleNames"`
	Claims    map[string]interface{} `json:"claims"`
	Warnings  []string               `json:"warnings"`
}

func (e *ApplicationEntitlements) isPrivileged() bool {
	for _, role := range e.Roles {
		if role.IsPrivileged {
			return true
		}
	}
	return false
}

func (e *ApplicationEntitlements) isEmpty() bool {
	return e == nil || (len(e.Groups) == 0 && len(e.RoleNames) == 0 && len(e.Claims) == 0)
}

var reAppRoleName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

var reClaimName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.:-]{0,99}$`)

var protocolClaims = []string{
	"iss", "sub", "aud", "exp", "nbf", "iat", "jti", "azp", "nonce", "auth_time", "at_hash", "c_hash", "sid",
	"acr", "amr", "scope", "scp", "client_id", "cnf", "act", "may_act", "tokenType", "signinMethod", "provider",
	"preferred_username", "name", "given_name", "family_name", "email", "email_verified", "phone_number",
	"phone_number_verified", "picture", "address", "groups", "roles", "permissions", "permissionNames",
}

// reservedClaimNames are the claims an application role must not set as custom claims: protocol
// and identity claims, every field of the user that a token may carry, and the groups and roles
// claims, which a role only fills through validated references.
var reservedClaimNames = func() map[string]bool {
	res := map[string]bool{}
	for _, name := range protocolClaims {
		res[strings.ToLower(name)] = true
	}
	for _, t := range []reflect.Type{reflect.TypeOf(User{}), reflect.TypeOf(Userinfo{})} {
		for i := 0; i < t.NumField(); i++ {
			tag := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
			if tag != "" && tag != "-" {
				res[strings.ToLower(tag)] = true
			}
		}
	}
	return res
}()

func isReservedClaimName(name string) bool {
	return reservedClaimNames[strings.ToLower(name)]
}

func (application *Application) getAppRole(name string) *AppRole {
	for _, role := range application.AppRoles {
		if role != nil && role.Name == name {
			return role
		}
	}
	return nil
}

// CheckApplicationAppRoles validates the application roles of an application before it is saved:
// the groups and roles they project must belong to the application's own organization, and their
// custom claims must not set protocol, identity or application-wide claims.
//
// stored is the application as saved before, if any: the custom claims must not collide with its
// token attributes either, so a partial update of one side can't sneak past the other.
func CheckApplicationAppRoles(application *Application, lang string, stored ...*Application) error {
	if len(application.AppRoles) == 0 {
		return nil
	}
	if application.IsShared {
		return errors.New(i18n.Translate(lang, "application:A shared application is available to every organization and does not take grants"))
	}

	appAttributes := map[string]bool{}
	for _, app := range append([]*Application{application}, stored...) {
		if app == nil {
			continue
		}
		for _, item := range app.TokenAttributes {
			if item != nil {
				appAttributes[strings.ToLower(item.Name)] = true
			}
		}
		// "Properties.<name>" token fields become the claim <name>
		for _, field := range app.TokenFields {
			if strings.HasPrefix(field, "Properties.") {
				appAttributes[strings.ToLower(strings.TrimPrefix(field, "Properties."))] = true
			}
		}
	}

	names := map[string]bool{}
	for _, role := range application.AppRoles {
		if role == nil || !reAppRoleName.MatchString(role.Name) {
			return errors.New(i18n.Translate(lang, "application:Application role names may only contain letters, digits, dots, underscores and hyphens"))
		}
		if names[role.Name] {
			return fmt.Errorf(i18n.Translate(lang, "application:The application role: %s is defined twice"), role.Name)
		}
		names[role.Name] = true

		for _, groupId := range role.Groups {
			owner, name, err := util.GetOwnerAndNameFromIdWithError(groupId)
			if err != nil || owner != application.Organization {
				return fmt.Errorf(i18n.Translate(lang, "application:The application role: %s may only use groups of the organization: %s"), role.Name, application.Organization)
			}
			group, err := getGroup(owner, name)
			if err != nil {
				return err
			}
			if group == nil {
				return fmt.Errorf(i18n.Translate(lang, "auth:The group: %s does not exist"), groupId)
			}
		}
		for _, roleId := range role.Roles {
			owner, name, err := util.GetOwnerAndNameFromIdWithError(roleId)
			if err != nil || owner != application.Organization {
				return fmt.Errorf(i18n.Translate(lang, "application:The application role: %s may only use roles of the organization: %s"), role.Name, application.Organization)
			}
			casdoorRole, err := getRole(owner, name)
			if err != nil {
				return err
			}
			if casdoorRole == nil {
				return fmt.Errorf(i18n.Translate(lang, "application:The role: %s does not exist"), roleId)
			}
		}
		for _, claim := range role.Claims {
			if claim == nil || !reClaimName.MatchString(claim.Name) || isReservedClaimName(claim.Name) || appAttributes[strings.ToLower(claim.Name)] {
				name := ""
				if claim != nil {
					name = claim.Name
				}
				return fmt.Errorf(i18n.Translate(lang, "application:The claim: %s can't be set by an application role"), name)
			}
			if claim.Value == "" || strings.Contains(claim.Value, "$user.") {
				return fmt.Errorf(i18n.Translate(lang, "application:The claim: %s needs a literal value"), claim.Name)
			}
			if claim.Type != "String" && claim.Type != "Array" {
				return fmt.Errorf(i18n.Translate(lang, "application:The claim: %s must be of type String or Array"), claim.Name)
			}
		}
	}
	return nil
}

// isUserInGrantedGroup reports whether the user is a member of the group or of one of its
// descendants. Every group from the user's group up to the granted one must be enabled.
func isUserInGrantedGroup(user *User, groupName string, groups map[string]*Group) bool {
	return isUserInGroupTree(user, groupName, groups)
}

// getMatchingApplicationGrants returns the valid grants of the application that apply to the user,
// through the user or one of the user's groups, sorted by name for a deterministic result.
func getMatchingApplicationGrants(user *User, application *Application) ([]*ApplicationGrant, error) {
	res := []*ApplicationGrant{}
	if user == nil || application == nil || application.IsShared {
		return res, nil
	}

	grants, err := GetApplicationGrants(application.Organization, application.Name)
	if err != nil || len(grants) == 0 {
		return res, err
	}

	now := time.Now()
	var groups map[string]*Group
	for _, grant := range grants {
		if grant.isExpired(now) {
			continue
		}
		owner, name := util.GetOwnerAndNameFromIdNoCheck(grant.Principal)
		if owner != user.Owner {
			continue
		}

		switch grant.PrincipalType {
		case ApplicationGrantPrincipalUser:
			if name == user.Name && grant.PrincipalId != "" && grant.PrincipalId == user.Id {
				res = append(res, grant)
			}
		case ApplicationGrantPrincipalGroup:
			if groups == nil {
				orgGroups, err := GetGroups(user.Owner)
				if err != nil {
					return nil, err
				}
				groups = make(map[string]*Group, len(orgGroups))
				for _, group := range orgGroups {
					groups[group.Name] = group
				}
			}
			if isUserInGrantedGroup(user, name, groups) {
				res = append(res, grant)
			}
		}
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res, nil
}

// formatGroupForApplication is the value a group of the application's organization has in the
// application's groups claim, in its token group format, as GetGroupsForToken gives it to members.
func formatGroupForApplication(group *Group, groups map[string]*Group, application *Application) (string, bool) {
	switch application.TokenGroupFormat {
	case "Name":
		return group.Name, true
	case "Path":
		path := group.Name
		current := group
		visited := map[string]bool{group.Name: true}
		for !current.IsTopGroup {
			parent, ok := groups[current.ParentId]
			if !ok || visited[parent.Name] {
				return "", false
			}
			visited[parent.Name] = true
			path = util.GetId(parent.Name, path)
			current = parent
		}
		return util.GetId(current.Owner, path), true
	default:
		return group.GetId(), true
	}
}

// GetApplicationEntitlements resolves the application roles that the grants matching the user
// assign, plus extraRoles (used to preview a grant before it is saved), into the values they add to
// the user's claims for this application only. References that are no longer valid, e.g. a deleted
// or disabled group, or a group of another organization written into the database directly, add
// nothing.
func GetApplicationEntitlements(user *User, application *Application, extraRoles []string) (*ApplicationEntitlements, error) {
	res := &ApplicationEntitlements{Grants: []*ApplicationGrant{}, Roles: []*AppRole{}, Groups: []string{}, RoleNames: []string{}, Claims: map[string]interface{}{}, Warnings: []string{}}
	if application == nil || application.IsShared || len(application.AppRoles) == 0 {
		return res, nil
	}

	grants, err := getMatchingApplicationGrants(user, application)
	if err != nil {
		return nil, err
	}
	res.Grants = grants

	roleNames := map[string]bool{}
	for _, grant := range grants {
		for _, name := range grant.Roles {
			roleNames[name] = true
		}
	}
	for _, name := range extraRoles {
		roleNames[name] = true
	}
	if len(roleNames) == 0 {
		return res, nil
	}

	for _, role := range application.AppRoles {
		if role != nil && roleNames[role.Name] {
			res.Roles = append(res.Roles, role)
		}
	}

	var orgGroups map[string]*Group
	groupValues := map[string]bool{}
	casdoorRoles := map[string]bool{}
	stringClaims := map[string]map[string]bool{}
	arrayClaims := map[string]map[string]bool{}
	for _, role := range res.Roles {
		for _, groupId := range role.Groups {
			owner, name := util.GetOwnerAndNameFromIdNoCheck(groupId)
			if owner != application.Organization {
				res.Warnings = append(res.Warnings, fmt.Sprintf("group %s ignored: not of the application's organization", groupId))
				continue
			}
			if orgGroups == nil {
				groups, err := GetGroups(application.Organization)
				if err != nil {
					return nil, err
				}
				orgGroups = make(map[string]*Group, len(groups))
				for _, group := range groups {
					orgGroups[group.Name] = group
				}
			}
			group, ok := orgGroups[name]
			if !ok || !group.IsEnabled {
				res.Warnings = append(res.Warnings, fmt.Sprintf("group %s ignored: missing or disabled", groupId))
				continue
			}
			value, ok := formatGroupForApplication(group, orgGroups, application)
			if !ok {
				res.Warnings = append(res.Warnings, fmt.Sprintf("group %s ignored: its parent group is missing", groupId))
				continue
			}
			groupValues[value] = true
		}
		for _, roleId := range role.Roles {
			owner, name := util.GetOwnerAndNameFromIdNoCheck(roleId)
			if owner != application.Organization {
				res.Warnings = append(res.Warnings, fmt.Sprintf("role %s ignored: not of the application's organization", roleId))
				continue
			}
			casdoorRole, err := getRole(owner, name)
			if err != nil {
				return nil, err
			}
			if casdoorRole == nil || !casdoorRole.IsEnabled {
				res.Warnings = append(res.Warnings, fmt.Sprintf("role %s ignored: missing or disabled", roleId))
				continue
			}
			casdoorRoles[casdoorRole.Name] = true
		}
		for _, claim := range role.Claims {
			if claim == nil || isReservedClaimName(claim.Name) || claim.Value == "" {
				continue
			}
			target := arrayClaims
			if claim.Type == "String" {
				target = stringClaims
			}
			if target[claim.Name] == nil {
				target[claim.Name] = map[string]bool{}
			}
			target[claim.Name][claim.Value] = true
		}
	}

	res.Groups = sortedKeys(groupValues)
	res.RoleNames = sortedKeys(casdoorRoles)
	for name, values := range arrayClaims {
		if _, ok := stringClaims[name]; ok {
			res.Warnings = append(res.Warnings, fmt.Sprintf("claim %s dropped: set as both String and Array", name))
			continue
		}
		res.Claims[name] = sortedKeys(values)
	}
	for name, values := range stringClaims {
		if _, ok := arrayClaims[name]; ok {
			continue
		}
		// two roles giving one single-valued claim different values can't be resolved safely
		if len(values) != 1 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("claim %s dropped: conflicting values %s", name, strings.Join(sortedKeys(values), ", ")))
			continue
		}
		res.Claims[name] = sortedKeys(values)[0]
	}
	return res, nil
}

func sortedKeys(values map[string]bool) []string {
	res := make([]string, 0, len(values))
	for value := range values {
		res = append(res, value)
	}
	sort.Strings(res)
	return res
}

// applyEntitlements adds the entitlements to the user the application's claims are made from.
// The user's own groups and roles stay as they are; the projected ones are appended.
func applyEntitlements(user *User, application *Application, entitlements *ApplicationEntitlements) {
	if entitlements.isEmpty() {
		return
	}
	for _, value := range entitlements.Groups {
		if !util.InSlice(user.Groups, value) {
			user.Groups = append(user.Groups, value)
		}
	}
	if len(entitlements.RoleNames) > 0 {
		roles := make([]*Role, 0, len(user.Roles)+len(entitlements.RoleNames))
		existing := map[string]bool{}
		for _, role := range user.Roles {
			roles = append(roles, role)
			existing[role.Name] = true
		}
		for _, name := range entitlements.RoleNames {
			if !existing[name] {
				roles = append(roles, &Role{Owner: application.Organization, Name: name, IsEnabled: true})
			}
		}
		user.Roles = roles
	}
}

// addClaims adds the custom claims of application roles to a token's claims without replacing
// any claim the token already has.
func addClaims(claims interface{}, extra map[string]interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	for name, value := range extra {
		if _, ok := res[name]; !ok {
			res[name] = value
		}
	}
	return res, nil
}

// ApplicationAccessPreview explains a user's access to an application: how the user may sign in,
// which grants and roles apply and what the application's claims will say.
type ApplicationAccessPreview struct {
	User             string                   `json:"user"`
	HomeOrganization string                   `json:"homeOrganization"`
	Access           string                   `json:"access"`
	Reason           string                   `json:"reason"`
	PreferredName    string                   `json:"preferredName"`
	OwnGroups        []string                 `json:"ownGroups"`
	OwnRoles         []string                 `json:"ownRoles"`
	Entitlements     *ApplicationEntitlements `json:"entitlements"`
	Groups           []string                 `json:"groups"`
	Roles            []string                 `json:"roles"`
	EmitsGroups      bool                     `json:"emitsGroups"`
	EmitsRoles       bool                     `json:"emitsRoles"`
	Privileged       bool                     `json:"privileged"`
}

func EmitsTokenField(application *Application, field string) bool {
	if application.TokenFormat != "JWT-Custom" {
		return true
	}
	return util.InSlice(application.TokenFields, field)
}

// GetApplicationAccessPreview explains the access of a user, with extraRoles added as if granted.
// The user's own groups and roles are only shown when the user may already sign in to the
// application, so a preview can't be used to read the memberships of another organization.
//
// With draft set, the preview is for a grant about to be made for the user: when the missing grant
// is the only reason the user can't sign in, the access is "Pending".
func GetApplicationAccessPreview(user *User, application *Application, extraRoles []string, draft bool, lang string) (*ApplicationAccessPreview, error) {
	res := &ApplicationAccessPreview{
		User: user.GetId(), HomeOrganization: user.Owner, OwnGroups: []string{}, OwnRoles: []string{},
		EmitsGroups: EmitsTokenField(application, "Groups"), EmitsRoles: EmitsTokenField(application, "Roles"),
	}

	isUserOfApplication, err := IsUserOfApplication(user, application)
	if err != nil {
		return nil, err
	}
	msg := ""
	if isUserOfApplication {
		res.Access = "Native"
	} else {
		msg, err = CheckApplicationAccess(user, application, lang)
		if err != nil {
			return nil, err
		}
		if msg == "" {
			res.Access = "Grant"
		} else {
			res.Access = "Denied"
			res.Reason = msg
			granted, err := HasApplicationGrant(user, application)
			if err != nil {
				return nil, err
			}
			// the MFA set-up and the identity of a user of another organization are only checked
			// once the user holds a grant, a draft can't be used to probe them
			if draft && !granted && user.Owner != application.Organization {
				res.Access = "Pending"
				res.Reason = ""
			}
		}
	}

	entitlements, err := GetApplicationEntitlements(user, application, extraRoles)
	if err != nil {
		return nil, err
	}
	res.Entitlements = entitlements
	res.Privileged = entitlements.isPrivileged()

	if res.Access == "Denied" || res.Access == "Pending" {
		res.PreferredName = GetApplicationGrantUserName(user)
		res.Groups = entitlements.Groups
		res.Roles = entitlements.RoleNames
		return res, nil
	}

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, err
	}
	tokenUser, _, err := getTokenUser(application, user)
	if err != nil {
		return nil, err
	}
	res.PreferredName = tokenUser.Name

	ownUser := *user
	if res.Access == "Grant" {
		qualified, err := qualifyApplicationGrantUser(user, application)
		if err != nil {
			return nil, err
		}
		ownUser = *qualified
	} else {
		ownUser.Groups, err = user.GetGroupsForToken(application)
		if err != nil {
			return nil, err
		}
	}
	// the memberships of a user of another organization are only shown when a grant gives the
	// application that user, not for e.g. a global admin who signs in natively
	if user.Owner != application.Organization && res.Access != "Grant" {
		ownUser.Groups = []string{}
		ownUser.Roles = []*Role{}
	}
	res.OwnGroups = append(res.OwnGroups, ownUser.Groups...)
	res.OwnRoles = getUserRoleNames(&ownUser)

	// what the application's claims finally hold: the user's own values plus the projected ones
	applyEntitlements(&ownUser, application, entitlements)
	res.Groups = ownUser.Groups
	res.Roles = getUserRoleNames(&ownUser)
	return res, nil
}

// renameAppRoleReferences follows the rename of a group or role of an organization in the
// application roles of its applications.
func renameAppRoleReferences(organization string, isGroup bool, oldId string, newId string) error {
	applications := []*Application{}
	err := ormer.Engine.Where("organization = ?", organization).Find(&applications)
	if err != nil {
		return err
	}
	for _, application := range applications {
		changed := false
		for _, role := range application.AppRoles {
			if role == nil {
				continue
			}
			references := role.Roles
			if isGroup {
				references = role.Groups
			}
			for i, id := range references {
				if id == oldId {
					references[i] = newId
					changed = true
				}
			}
		}
		if changed {
			_, err = ormer.Engine.ID([]interface{}{application.Owner, application.Name}).Cols("app_roles").Update(application)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateApplicationAppRoles replaces the application roles of an application. A role that a grant
// still assigns can't be removed: the grant would silently lose what it gives.
func UpdateApplicationAppRoles(application *Application, roles []*AppRole, lang string) (bool, error) {
	updated := *application
	updated.AppRoles = roles
	err := CheckApplicationAppRoles(&updated, lang)
	if err != nil {
		return false, err
	}

	grants, err := GetApplicationGrants(application.Organization, application.Name)
	if err != nil {
		return false, err
	}
	for _, grant := range grants {
		for _, name := range grant.Roles {
			if updated.getAppRole(name) == nil {
				return false, fmt.Errorf(i18n.Translate(lang, "application:The application role: %s is still assigned by the grant for: %s"), name, grant.Principal)
			}
		}
		// a role can't become privileged while a group of another organization holds it
		edited := *grant
		if err = checkApplicationGrantRoles(&edited, &updated, lang); err != nil {
			return false, err
		}
	}

	_, err = ormer.Engine.ID([]interface{}{application.Owner, application.Name}).Cols("app_roles").Update(&Application{AppRoles: roles})
	if err != nil {
		return false, err
	}
	return true, nil
}

// removeAppRoleReferences drops a deleted group or role of an organization from the application
// roles of its applications, so one created later under the same name is not projected.
func removeAppRoleReferences(organization string, isGroup bool, id string) error {
	applications := []*Application{}
	err := ormer.Engine.Where("organization = ?", organization).Find(&applications)
	if err != nil {
		return err
	}
	for _, application := range applications {
		changed := false
		for _, role := range application.AppRoles {
			if role == nil {
				continue
			}
			if isGroup && util.InSlice(role.Groups, id) {
				role.Groups = util.DeleteVal(role.Groups, id)
				changed = true
			}
			if !isGroup && util.InSlice(role.Roles, id) {
				role.Roles = util.DeleteVal(role.Roles, id)
				changed = true
			}
		}
		if changed {
			_, err = ormer.Engine.ID([]interface{}{application.Owner, application.Name}).Cols("app_roles").Update(application)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// IsApplicationGrantPrincipal reports whether a valid grant of the application covers the user.
func IsApplicationGrantPrincipal(user *User, application *Application) (bool, error) {
	grants, err := getMatchingApplicationGrants(user, application)
	if err != nil {
		return false, err
	}
	return len(grants) > 0, nil
}
