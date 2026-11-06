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
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	ApplicationGrantPrincipalUser  = "User"
	ApplicationGrantPrincipalGroup = "Group"
)

// ApplicationGrant lets an existing user or group of another organization sign in to an
// application through its OAuth/OIDC client. The application keeps its organization and its
// client; the user keeps their own organization, credentials and MFA. Owner is the
// organization of the application, the only one whose admins may manage its grants.
type ApplicationGrant struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	CreatedBy   string `xorm:"varchar(255)" json:"createdBy"`

	Application   string `xorm:"varchar(100) index" json:"application"`
	PrincipalType string `xorm:"varchar(100)" json:"principalType"`
	// Principal is the "organization/name" ID of the granted user or group
	Principal string `xorm:"varchar(255) index" json:"principal"`
	// PrincipalId is the immutable ID of a granted user: a user deleted and created again
	// under the same name, or renamed, does not inherit the grant
	PrincipalId string `xorm:"varchar(100)" json:"principalId"`
	ExpireTime  string `xorm:"varchar(100)" json:"expireTime"`
	Description string `xorm:"varchar(255)" json:"description"`

	// Roles are application roles (Application.AppRoles) the grant assigns. Empty means the grant
	// only allows the sign-in.
	Roles       []string `xorm:"mediumtext" json:"roles"`
	UpdatedTime string   `xorm:"varchar(100)" json:"updatedTime"`
	UpdatedBy   string   `xorm:"varchar(255)" json:"updatedBy"`
}

func (grant *ApplicationGrant) GetId() string {
	return util.GetId(grant.Owner, grant.Name)
}

func (grant *ApplicationGrant) isExpired(now time.Time) bool {
	if grant.ExpireTime == "" {
		return false
	}
	expireTime, err := time.Parse(time.RFC3339, grant.ExpireTime)
	if err != nil {
		// an unreadable expiry fails closed
		return true
	}
	return !now.Before(expireTime)
}

func GetApplicationGrants(organization string, application string) ([]*ApplicationGrant, error) {
	grants := []*ApplicationGrant{}
	if organization == "" || application == "" {
		return grants, nil
	}
	err := ormer.Engine.Where("owner = ? and application = ?", organization, application).Desc("created_time").Find(&grants)
	if err != nil {
		return nil, err
	}
	return grants, nil
}

func getApplicationGrant(owner string, name string) (*ApplicationGrant, error) {
	if owner == "" || name == "" {
		return nil, nil
	}
	grant := ApplicationGrant{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&grant)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	return &grant, nil
}

func GetApplicationGrant(id string) (*ApplicationGrant, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getApplicationGrant(owner, name)
}

// GetApplicationGrantUserName is the name a user signed in through a grant is known by in
// the tokens and the userinfo of the application. The organization is appended, so that the
// user can't take the place of the application's own user of the same name in a relying
// party that keys its accounts on the name.
func GetApplicationGrantUserName(user *User) string {
	return fmt.Sprintf("%s@%s", user.Name, user.Owner)
}

// CheckApplicationGrant validates a new grant against the application it is added to and
// fills in what the server decides: the name, the creation time and the principal's ID.
func CheckApplicationGrant(grant *ApplicationGrant, application *Application, lang string) error {
	if application == nil {
		return errors.New(i18n.Translate(lang, "general:The application does not exist"))
	}
	if application.Organization != grant.Owner || application.Name != grant.Application {
		return errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
	}
	if application.IsShared {
		return errors.New(i18n.Translate(lang, "application:A shared application is available to every organization and does not take grants"))
	}

	principalOwner, principalName, err := util.GetOwnerAndNameFromIdWithError(grant.Principal)
	if err != nil || principalOwner == "" || principalName == "" {
		return fmt.Errorf(i18n.Translate(lang, "application:Invalid principal: %s"), grant.Principal)
	}
	isInternal := principalOwner == application.Organization
	if isInternal && len(grant.Roles) == 0 {
		return errors.New(i18n.Translate(lang, "application:The principal belongs to the organization of the application, which already has access"))
	}
	err = checkApplicationGrantRoles(grant, application, lang)
	if err != nil {
		return err
	}
	// one answer for a missing organization, user or group, so that a grant can't be used to
	// probe another organization
	notFound := fmt.Errorf(i18n.Translate(lang, "application:The principal: %s does not exist"), grant.Principal)
	organization, err := getOrganization("admin", principalOwner)
	if err != nil {
		return err
	}
	if organization == nil {
		return notFound
	}
	if strings.Contains(principalOwner, "@") {
		return fmt.Errorf(i18n.Translate(lang, "application:Invalid principal: %s"), grant.Principal)
	}

	switch grant.PrincipalType {
	case ApplicationGrantPrincipalUser:
		user, err := getUser(principalOwner, principalName)
		if err != nil {
			return err
		}
		if user == nil || user.IsDeleted {
			return notFound
		}
		if user.Id == "" {
			return notFound
		}
		if !isInternal {
			collision, err := getApplicationGrantNameCollision(user, application)
			if err != nil {
				return err
			}
			if collision != "" {
				return fmt.Errorf(i18n.Translate(lang, "application:The name %s is already used by a user of the application's organization"), collision)
			}
			identityOk, err := checkApplicationGrantIdentity(user, application)
			if err != nil {
				return err
			}
			if !identityOk {
				return fmt.Errorf(i18n.Translate(lang, "application:The ID of the user: %s is not unique and can't identify the user to the application"), grant.Principal)
			}
		}
		grant.PrincipalId = user.Id
	case ApplicationGrantPrincipalGroup:
		group, err := getGroup(principalOwner, principalName)
		if err != nil {
			return err
		}
		if group == nil {
			return notFound
		}
		grant.PrincipalId = ""
	default:
		return fmt.Errorf(i18n.Translate(lang, "application:Invalid principal type: %s"), grant.PrincipalType)
	}

	if grant.ExpireTime != "" {
		expireTime, err := time.Parse(time.RFC3339, grant.ExpireTime)
		if err != nil {
			return fmt.Errorf(i18n.Translate(lang, "application:Invalid expire time: %s"), grant.ExpireTime)
		}
		if !time.Now().Before(expireTime) {
			return errors.New(i18n.Translate(lang, "application:The expire time must be in the future"))
		}
	}
	if len(grant.Description) > 255 {
		grant.Description = grant.Description[:255]
	}

	existing, err := GetApplicationGrants(grant.Owner, grant.Application)
	if err != nil {
		return err
	}
	for _, item := range existing {
		if item.PrincipalType == grant.PrincipalType && item.Principal == grant.Principal {
			return fmt.Errorf(i18n.Translate(lang, "application:The principal: %s already has a grant for this application"), grant.Principal)
		}
	}

	grant.Name = util.GenerateId()
	grant.CreatedTime = util.GetCurrentTime()
	grant.UpdatedTime = ""
	grant.UpdatedBy = ""
	return nil
}

// checkApplicationGrantRoles requires every role of the grant to be an application role of the
// application, so a grant can only assign what the application's organization defined.
//
// A privileged role can't be assigned to a group of another organization: whoever administers
// that group, outside the application's organization, could add anyone to it, including themselves.
func checkApplicationGrantRoles(grant *ApplicationGrant, application *Application, lang string) error {
	principalOwner, _ := util.GetOwnerAndNameFromIdNoCheck(grant.Principal)
	isForeignGroup := grant.PrincipalType == ApplicationGrantPrincipalGroup && principalOwner != application.Organization
	seen := map[string]bool{}
	roles := []string{}
	for _, name := range grant.Roles {
		role := application.getAppRole(name)
		if role == nil {
			return fmt.Errorf(i18n.Translate(lang, "application:The application role: %s does not exist"), name)
		}
		if role.IsPrivileged && isForeignGroup {
			return fmt.Errorf(i18n.Translate(lang, "application:The privileged application role: %s can only be granted to users of other organizations, not to their groups"), name)
		}
		if !seen[name] {
			seen[name] = true
			roles = append(roles, name)
		}
	}
	sort.Strings(roles)
	grant.Roles = roles
	return nil
}

// CheckApplicationGrantUpdate applies the editable fields of an update to the stored grant: its
// roles, expiry and description. Who and which application it is for can't change; that takes a
// new grant.
func CheckApplicationGrantUpdate(stored *ApplicationGrant, update *ApplicationGrant, application *Application, lang string) error {
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(stored.Principal)
	if owner == application.Organization && len(update.Roles) == 0 {
		return errors.New(i18n.Translate(lang, "application:The principal belongs to the organization of the application, which already has access"))
	}
	edited := *stored
	edited.Roles = update.Roles
	err := checkApplicationGrantRoles(&edited, application, lang)
	if err != nil {
		return err
	}
	if update.ExpireTime != "" && update.ExpireTime != stored.ExpireTime {
		expireTime, err := time.Parse(time.RFC3339, update.ExpireTime)
		if err != nil {
			return fmt.Errorf(i18n.Translate(lang, "application:Invalid expire time: %s"), update.ExpireTime)
		}
		if !time.Now().Before(expireTime) {
			return errors.New(i18n.Translate(lang, "application:The expire time must be in the future"))
		}
	}
	stored.Roles = edited.Roles
	stored.ExpireTime = update.ExpireTime
	stored.Description = update.Description
	if len(stored.Description) > 255 {
		stored.Description = stored.Description[:255]
	}
	stored.UpdatedTime = util.GetCurrentTime()
	return nil
}

func UpdateApplicationGrant(grant *ApplicationGrant) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{grant.Owner, grant.Name}).Cols("roles", "expire_time", "description", "updated_time", "updated_by").Update(grant)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func AddApplicationGrant(grant *ApplicationGrant) (bool, error) {
	affected, err := ormer.Engine.Insert(grant)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteApplicationGrant(grant *ApplicationGrant) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{grant.Owner, grant.Name}).Delete(&ApplicationGrant{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func deleteApplicationGrantsByApplication(organization string, application string) error {
	_, err := ormer.Engine.Where("owner = ? and application = ?", organization, application).Delete(&ApplicationGrant{})
	return err
}

func deleteApplicationGrantsByPrincipal(principalType string, principal string) error {
	_, err := ormer.Engine.Where("principal_type = ? and principal = ?", principalType, principal).Delete(&ApplicationGrant{})
	return err
}

func renameApplicationGrantGroup(oldId string, newId string) error {
	_, err := ormer.Engine.Where("principal_type = ? and principal = ?", ApplicationGrantPrincipalGroup, oldId).
		Cols("principal").Update(&ApplicationGrant{Principal: newId})
	return err
}

func deleteApplicationGrantsByOrganization(organization string) error {
	_, err := ormer.Engine.Where("owner = ?", organization).Delete(&ApplicationGrant{})
	if err != nil {
		return err
	}

	grants := []*ApplicationGrant{}
	err = ormer.Engine.Find(&grants)
	if err != nil {
		return err
	}
	for _, grant := range grants {
		owner, _ := util.GetOwnerAndNameFromIdNoCheck(grant.Principal)
		if owner == organization {
			if _, err = DeleteApplicationGrant(grant); err != nil {
				return err
			}
		}
	}
	return nil
}

// isUserInGroupTree reports whether the user is a member of the group or of one of its
// descendants. Every group from the user's group up to the granted one must be enabled.
func isUserInGroupTree(user *User, groupName string, groups map[string]*Group) bool {
	for _, groupId := range user.Groups {
		owner, name := util.GetOwnerAndNameFromIdNoCheck(groupId)
		if owner != user.Owner {
			continue
		}

		visited := map[string]bool{}
		for name != "" && !visited[name] {
			visited[name] = true
			group, ok := groups[name]
			if !ok || !group.IsEnabled {
				break
			}
			if name == groupName {
				return true
			}
			if group.IsTopGroup {
				break
			}
			name = group.ParentId
		}
	}
	return false
}

// HasApplicationGrant reports whether a user of another organization holds a valid grant for
// the application, directly or through one of their groups.
func HasApplicationGrant(user *User, application *Application) (bool, error) {
	if user == nil || application == nil || application.IsShared || user.Owner == application.Organization {
		return false, nil
	}

	grants, err := getMatchingApplicationGrants(user, application)
	if err != nil {
		return false, err
	}
	return len(grants) > 0, nil
}

// isApplicationGrantUser reports whether a user who obtains tokens of the application does
// so through a grant rather than as a user of the application, and so gets their claims
// qualified with their organization.
func isApplicationGrantUser(user *User, application *Application) (bool, error) {
	if user == nil || application == nil || application.IsShared || user.Owner == application.Organization {
		return false, nil
	}
	isUserOfApplication, err := IsUserOfApplication(user, application)
	if err != nil {
		return false, err
	}
	return !isUserOfApplication, nil
}

// getApplicationGrantNameCollision returns the qualified name of the user when a user of the
// application's organization already has it, compared without case as relying parties often
// fold the case of usernames.
func getApplicationGrantNameCollision(user *User, application *Application) (string, error) {
	qualifiedName := GetApplicationGrantUserName(user)
	count, err := ormer.Engine.Where("owner = ? and lower(name) = ?", application.Organization, strings.ToLower(qualifiedName)).Count(&User{})
	if err != nil {
		return "", err
	}
	if count > 0 {
		return qualifiedName, nil
	}
	return "", nil
}

// checkApplicationGrantIdentity refuses a user whose claims could be taken for another user's
// at the relying party: "sub" is the user's ID, which an admin of the user's organization may
// set, and "name@organization" is only unambiguous while the organization's name has no "@".
func checkApplicationGrantIdentity(user *User, application *Application) (bool, error) {
	if user.Id == "" || strings.Contains(user.Owner, "@") {
		return false, nil
	}
	count, err := ormer.Engine.Where("id = ?", user.Id).Count(&User{})
	if err != nil {
		return false, err
	}
	if count != 1 {
		return false, nil
	}
	collision, err := getApplicationGrantNameCollision(user, application)
	if err != nil {
		return false, err
	}
	return collision == "", nil
}

// CheckApplicationAccess returns why the user may not get tokens of the application through
// OAuth/OIDC, or "" when they may. A user of the application is allowed as before. A user of
// another organization needs a valid grant, has to meet the MFA requirement of the
// application's organization and must not share the qualified name of one of its users.
func CheckApplicationAccess(user *User, application *Application, lang string) (string, error) {
	if user == nil || application == nil {
		return i18n.Translate(lang, "auth:Unauthorized operation"), nil
	}

	isUserOfApplication, err := IsUserOfApplication(user, application)
	if err != nil {
		return "", err
	}
	if isUserOfApplication {
		return "", nil
	}

	granted, err := HasApplicationGrant(user, application)
	if err != nil {
		return "", err
	}
	if !granted {
		return i18n.Translate(lang, "auth:Unauthorized operation"), nil
	}

	return checkApplicationGrantPreconditions(user, application, lang)
}

// checkApplicationGrantPreconditions returns why a user who holds a grant may still not sign in:
// an identity that could be taken for another user's, or MFA the application's organization
// requires and the user has not set up.
func checkApplicationGrantPreconditions(user *User, application *Application, lang string) (string, error) {
	var err error
	identityOk, err := checkApplicationGrantIdentity(user, application)
	if err != nil {
		return "", err
	}
	if !identityOk {
		return i18n.Translate(lang, "auth:Unauthorized operation"), nil
	}

	organization := application.OrganizationObj
	if organization == nil || organization.Name != application.Organization {
		organization, err = getOrganization("admin", application.Organization)
		if err != nil {
			return "", err
		}
	}
	// the grant must not lower the MFA bar of the application's organization: the user's
	// own MFA overrides do not apply to it
	userWithoutOverride := *user
	userWithoutOverride.MfaItems = nil
	if IsNeedPromptMfa(organization, &userWithoutOverride) {
		return i18n.Translate(lang, "auth:The application requires multi-factor authentication, please set it up in your account first"), nil
	}

	return "", nil
}

// qualifyApplicationGrantUser shapes the claims of a user signed in through a grant: the name
// gets the organization appended, groups are always given with their organization and roles
// and permissions are named with theirs, so that none of them can be mistaken for one of the
// application's organization.
func qualifyApplicationGrantUser(user *User, application *Application) (*User, error) {
	userCopy := *user
	userCopy.Name = GetApplicationGrantUserName(user)
	// being an admin of their own organization means nothing for the application
	userCopy.IsAdmin = false

	if application.TokenGroupFormat == "Path" {
		groups, err := user.GetUserFullGroupPath()
		if err != nil {
			return nil, err
		}
		userCopy.Groups = groups
	} else {
		groups := []string{}
		for _, groupId := range user.Groups {
			owner, name := util.GetOwnerAndNameFromIdNoCheck(groupId)
			if owner == user.Owner && name != "" {
				groups = append(groups, util.GetId(owner, name))
			}
		}
		userCopy.Groups = groups
	}

	if user.Roles != nil {
		userCopy.Roles = make([]*Role, 0, len(user.Roles))
		for _, role := range user.Roles {
			roleCopy := *role
			roleCopy.Name = util.GetId(role.Owner, role.Name)
			userCopy.Roles = append(userCopy.Roles, &roleCopy)
		}
	}
	if user.Permissions != nil {
		userCopy.Permissions = make([]*Permission, 0, len(user.Permissions))
		for _, permission := range user.Permissions {
			permissionCopy := *permission
			permissionCopy.Name = util.GetId(permission.Owner, permission.Name)
			userCopy.Permissions = append(userCopy.Permissions, &permissionCopy)
		}
	}
	return &userCopy, nil
}

// ApplicationGrantPrincipal is the part of a user or group shown to the admin choosing
// whom to grant access: only its ID and display name.
type ApplicationGrantPrincipal struct {
	Type        string `json:"type"`
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// SearchApplicationGrantPrincipals finds the users or groups of an organization that may be
// granted access to the application. A global admin may search by part of the name; an admin
// of the application's organization only finds the principal they name exactly, so that the
// members of another organization can't be listed from it.
func SearchApplicationGrantPrincipals(application *Application, principalType string, organization string, query string, exactOnly bool) ([]*ApplicationGrantPrincipal, error) {
	res := []*ApplicationGrantPrincipal{}
	query = strings.TrimSpace(query)
	if application == nil || organization == "" || query == "" {
		return res, nil
	}
	// the users and groups of the application's own organization may be searched by its admins
	if organization == application.Organization {
		exactOnly = false
	}
	const limit = 20

	switch principalType {
	case ApplicationGrantPrincipalUser:
		users := []*User{}
		session := ormer.Engine.Where("owner = ? and is_deleted = ?", organization, false)
		if exactOnly {
			session = session.And("name = ?", query)
		} else {
			like := "%" + escapeLike(query) + "%"
			session = session.And("(name like ? or display_name like ?)", like, like)
		}
		err := session.Cols("owner", "name", "display_name").Limit(limit).Asc("name").Find(&users)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			res = append(res, &ApplicationGrantPrincipal{Type: principalType, Id: user.GetId(), DisplayName: user.DisplayName})
		}
	case ApplicationGrantPrincipalGroup:
		groups := []*Group{}
		session := ormer.Engine.Where("owner = ?", organization)
		if exactOnly {
			session = session.And("name = ?", query)
		} else {
			like := "%" + escapeLike(query) + "%"
			session = session.And("(name like ? or display_name like ?)", like, like)
		}
		err := session.Cols("owner", "name", "display_name").Limit(limit).Asc("name").Find(&groups)
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			res = append(res, &ApplicationGrantPrincipal{Type: principalType, Id: group.GetId(), DisplayName: group.DisplayName})
		}
	}
	return res, nil
}

func escapeLike(value string) string {
	return strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(value)
}

// GetApplicationGrantUserForLogin finds the user of another organization who signs in on the
// application's own sign-in page, when no user of the application's organization matches.
// "organization/username" names the organization explicitly; a plain username, email or
// phone is looked up only among the users that a grant of the application covers. An
// ambiguous match is refused instead of guessed.
func GetApplicationGrantUserForLogin(application *Application, field string, lang string) (*User, error) {
	if application == nil || application.IsShared || field == "" {
		return nil, nil
	}

	if strings.Contains(field, "/") {
		owner, name, err := util.GetOwnerAndNameFromIdWithError(field)
		if err != nil || owner == application.Organization {
			return nil, nil
		}
		user, err := GetUserByFields(owner, name)
		if err != nil || user == nil {
			return nil, err
		}
		granted, err := HasApplicationGrant(user, application)
		if err != nil || !granted {
			return nil, err
		}
		return user, nil
	}

	grants, err := GetApplicationGrants(application.Organization, application.Name)
	if err != nil || len(grants) == 0 {
		return nil, err
	}
	organizations := []string{}
	for _, grant := range grants {
		owner, _ := util.GetOwnerAndNameFromIdNoCheck(grant.Principal)
		if owner != "" && owner != application.Organization && !util.InSlice(organizations, owner) {
			organizations = append(organizations, owner)
		}
	}

	var match *User
	for _, organization := range organizations {
		user, err := GetUserByFields(organization, field)
		if err != nil {
			return nil, err
		}
		if user == nil {
			continue
		}
		granted, err := HasApplicationGrant(user, application)
		if err != nil {
			return nil, err
		}
		if !granted {
			continue
		}
		if match != nil {
			return nil, errors.New(i18n.Translate(lang, "auth:The account matches users of several organizations, please sign in with organization/username"))
		}
		match = user
	}
	return match, nil
}
