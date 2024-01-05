// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

package ldap

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/lor00x/goldap/message"

	ldap "github.com/forestmgy/ldapserver"

	"github.com/xorm-io/builder"
)

type V = message.AttributeValue

type UserAttributeMapper func(user *object.User) []V

type UserFieldRelation struct {
	userField     string
	ldapField     string
	notSearchable bool
	hideOnStarOp  bool
	fieldMapper   UserAttributeMapper
	constantValue []V
}

func (rel UserFieldRelation) GetField() (string, error) {
	if rel.notSearchable {
		return "", fmt.Errorf("attribute %s not supported", rel.userField)
	}
	return rel.userField, nil
}

func (rel UserFieldRelation) GetAttributeValues(user *object.User) []V {
	if rel.constantValue != nil && rel.fieldMapper == nil {
		return rel.constantValue
	}
	return rel.fieldMapper(user)
}

type UserFieldRelationMap map[string]UserFieldRelation

func (m UserFieldRelationMap) CaseInsensitiveGet(key string) (UserFieldRelation, bool) {
	lowerKey := strings.ToLower(key)
	ret, ok := m[lowerKey]
	return ret, ok
}

type GroupAttributeMapper func(group *object.Group) []V

type GroupFieldRelation struct {
	groupField    string
	ldapField     string
	notSearchable bool
	hideOnStarOp  bool
	fieldMapper   GroupAttributeMapper
	constantValue []V
}

func (rel GroupFieldRelation) GetField() (string, error) {
	if rel.notSearchable {
		return "", fmt.Errorf("attribute %s not supported", rel.groupField)
	}
	return rel.groupField, nil
}

func (rel GroupFieldRelation) GetAttributeValues(group *object.Group) []V {
	if rel.constantValue != nil && rel.fieldMapper == nil {
		return rel.constantValue
	}
	return rel.fieldMapper(group)
}

type GroupFieldRelationMap map[string]GroupFieldRelation

func (m GroupFieldRelationMap) CaseInsensitiveGet(key string) (GroupFieldRelation, bool) {
	lowerKey := strings.ToLower(key)
	ret, ok := m[lowerKey]
	return ret, ok
}

var ldapUserAttributesMapping = UserFieldRelationMap{
	"cn": {ldapField: "cn", userField: "name", hideOnStarOp: true, fieldMapper: func(user *object.User) []V {
		return []V{V(user.Name)}
	}},
	"uid": {ldapField: "uid", userField: "name", hideOnStarOp: true, fieldMapper: func(user *object.User) []V {
		return []V{V(user.Name)}
	}},
	"displayname": {ldapField: "displayName", userField: "displayName", fieldMapper: func(user *object.User) []V {
		return []V{V(user.DisplayName)}
	}},
	"email": {ldapField: "email", userField: "email", fieldMapper: func(user *object.User) []V {
		return []V{V(user.Email)}
	}},
	"mail": {ldapField: "mail", userField: "email", fieldMapper: func(user *object.User) []V {
		return []V{V(user.Email)}
	}},
	"mobile": {ldapField: "mobile", userField: "phone", fieldMapper: func(user *object.User) []V {
		return []V{V(user.Phone)}
	}},
	"telephonenumber": {ldapField: "telephoneNumber", userField: "phone", fieldMapper: func(user *object.User) []V {
		return []V{V(user.Phone)}
	}},
	"postaladdress": {ldapField: "postalAddress", userField: "address", fieldMapper: func(user *object.User) []V {
		return []V{V(strings.Join(user.Address, " "))}
	}},
	"title": {ldapField: "title", userField: "title", fieldMapper: func(user *object.User) []V {
		return []V{V(user.Title)}
	}},
	"gecos": {ldapField: "gecos", userField: "displayName", fieldMapper: func(user *object.User) []V {
		return []V{V(user.DisplayName)}
	}},
	"description": {ldapField: "description", userField: "displayName", fieldMapper: func(user *object.User) []V {
		return []V{V(user.DisplayName)}
	}},
	"logindisabled": {ldapField: "loginDisabled", userField: "isForbidden", fieldMapper: func(user *object.User) []V {
		if user.IsForbidden {
			return []V{V("1")}
		} else {
			return []V{V("0")}
		}
	}},
	"userpassword": {
		ldapField:     "userPassword",
		userField:     "userPassword",
		notSearchable: true,
		fieldMapper: func(user *object.User) []V {
			return []V{V(getUserPasswordWithType(user))}
		},
	},
	"uidnumber": {ldapField: "uidNumber", notSearchable: true, fieldMapper: func(user *object.User) []V {
		return []V{V(fmt.Sprintf("%v", hash(user.Name)))}
	}},
	"gidnumber": {ldapField: "gidNumber", notSearchable: true, fieldMapper: func(user *object.User) []V {
		if len(user.Groups) == 0 {
			return []V{V("")}
		}
		group, err := object.GetGroup(user.Groups[0])
		if err != nil {
			log.Printf("gidnumber object.GetGroup error: %s", err)
			return []V{V("")}
		}
		return []V{V(fmt.Sprintf("%v", hash(group.Name)))}
	}},
	"homedirectory": {ldapField: "homeDirectory", notSearchable: true, fieldMapper: func(user *object.User) []V {
		return []V{V("/home/" + user.Name)}
	}},
	"loginshell": {ldapField: "loginShell", notSearchable: true, fieldMapper: func(user *object.User) []V {
		if user.IsForbidden || user.IsDeleted {
			return []V{V("/sbin/nologin")}
		} else {
			return []V{V("/bin/bash")}
		}
	}},
	"shadowlastchange": {ldapField: "shadowLastChange", notSearchable: true, fieldMapper: func(user *object.User) []V {
		// "this attribute specifies number of days between January 1, 1970, and the date that the password was last modified"
		updatedTime, err := time.Parse(time.RFC3339, user.UpdatedTime)
		if err != nil {
			log.Printf("shadowlastchange time.Parse error: %s", err)
			updatedTime = time.Now()
		}
		return []V{V(fmt.Sprint(updatedTime.Unix() / 86400))}
	}},
	"pwdchangedtime": {ldapField: "pwdChangedTime", notSearchable: true, fieldMapper: func(user *object.User) []V {
		updatedTime, err := time.Parse(time.RFC3339, user.UpdatedTime)
		if err != nil {
			log.Printf("pwdchangedtime time.Parse error: %s", err)
			updatedTime = time.Now()
		}
		return []V{V(updatedTime.UTC().Format("20060102030405Z"))}
	}},
	"shadowmin":     {ldapField: "shadowMin", notSearchable: true, constantValue: []V{V("0")}},
	"shadowmax":     {ldapField: "shadowMax", notSearchable: true, constantValue: []V{V("99999")}},
	"shadowwarning": {ldapField: "shadowWarning", notSearchable: true, constantValue: []V{V("7")}},
	"shadowexpire": {ldapField: "shadowExpire", notSearchable: true, fieldMapper: func(user *object.User) []V {
		if user.IsForbidden {
			return []V{V("1")}
		} else {
			return []V{V("-1")}
		}
	}},
	"shadowinactive": {ldapField: "shadowInactive", notSearchable: true, constantValue: []V{V("0")}},
	"shadowflag":     {ldapField: "shadowFlag", notSearchable: true, constantValue: []V{V("0")}},
	"memberof": {ldapField: "memberOf", notSearchable: true, fieldMapper: func(user *object.User) []V {
		var groupdn []V
		for _, groupId := range user.Groups {
			group, err := object.GetGroup(groupId)
			if err != nil {
				log.Printf("memberOf object.GetGroup error: %s", err)
				continue
			}
			groupdn = append(groupdn, V(fmt.Sprintf("cn=%s,cn=groups,ou=%s", group.Name, group.Owner)))
		}
		return groupdn
	}},
	"objectclass": {ldapField: "objectClass", notSearchable: true, constantValue: []V{
		V("top"),
		V("posixAccount"),
		V("shadowAccount"),
		V("person"),
		V("organizationalPerson"),
		V("inetOrgPerson"),
		V("apple-user"),
		V("sambaSamAccount"),
		V("sambaIdmapEntry"),
		V("extensibleObject"),
	}},
}

var ldapGroupAttributesMapping = GroupFieldRelationMap{
	"cn": {ldapField: "cn", hideOnStarOp: true, fieldMapper: func(group *object.Group) []V {
		return []V{V(group.Name)}
	}},
	"gidnumber": {ldapField: "gidNumber", hideOnStarOp: true, fieldMapper: func(group *object.Group) []V {
		return []V{V(fmt.Sprintf("%v", hash(group.Name)))}
	}},
	"member": {ldapField: "member", fieldMapper: func(group *object.Group) []V {
		users, err := object.GetGroupUsers(group.GetId())
		if err != nil {
			log.Printf("member object.GetGroupUsers error: %s", err)
			return []V{V("")}
		}
		var members []V
		for _, user := range users {
			members = append(members, V(fmt.Sprintf("uid=%s,cn=users,ou=%s", user.Name, user.Owner)))
		}
		return members
	}},
	"memberuid": {ldapField: "memberUid", fieldMapper: func(group *object.Group) []V {
		users, err := object.GetGroupUsers(group.GetId())
		if err != nil {
			log.Printf("member object.GetGroupUsers error: %s", err)
			return []V{V("")}
		}
		var members []message.AttributeValue
		for _, user := range users {
			members = append(members, message.AttributeValue(user.Name))
		}
		return members
	}},
	"description": {ldapField: "description", hideOnStarOp: true, fieldMapper: func(group *object.Group) []V {
		return []V{V(group.DisplayName)}
	}},
	"objectclass": {ldapField: "objectClass", hideOnStarOp: true, constantValue: []V{
		V("top"),
		V("posixGroup"),
	}},
}

var (
	AdditionalLdapUserAttributes  []message.LDAPString
	AdditionalLdapGroupAttributes []message.LDAPString
)

func init() {
	for _, v := range ldapUserAttributesMapping {
		if v.hideOnStarOp {
			continue
		}
		AdditionalLdapUserAttributes = append(AdditionalLdapUserAttributes, message.LDAPString(v.ldapField))
	}
	for _, v := range ldapGroupAttributesMapping {
		if v.hideOnStarOp {
			continue
		}
		AdditionalLdapGroupAttributes = append(AdditionalLdapGroupAttributes, message.LDAPString(v.ldapField))
	}
}

func getNameAndOrgFromDN(DN string) (string, string, error) {
	DNFields := strings.Split(DN, ",")
	params := make(map[string]string, len(DNFields))
	for _, field := range DNFields {
		if strings.Contains(field, "=") {
			k := strings.Split(field, "=")
			params[k[0]] = k[1]
		}
	}

	if params["cn"] == "" {
		return "", "", fmt.Errorf("please use Admin Name format like cn=xxx,ou=xxx,dc=example,dc=com")
	}
	if params["ou"] == "" {
		return params["cn"], object.CasdoorOrganization, nil
	}
	return params["cn"], params["ou"], nil
}

func getNameAndOrgFromFilter(baseDN, filter string) (string, string, int) {
	if !strings.Contains(baseDN, "ou=") {
		return "", "", ldap.LDAPResultInvalidDNSyntax
	}

	name, org, err := getNameAndOrgFromDN(fmt.Sprintf("cn=%s,", getUsername(filter)) + baseDN)
	if err != nil {
		panic(err)
	}

	return name, org, ldap.LDAPResultSuccess
}

func getUsername(filter string) string {
	nameIndex := strings.Index(filter, "cn=")
	if nameIndex == -1 {
		nameIndex = strings.Index(filter, "uid=")
		if nameIndex == -1 {
			return "*"
		} else {
			nameIndex += 4
		}
	} else {
		nameIndex += 3
	}

	var name string
	for i := nameIndex; filter[i] != ')'; i++ {
		name = name + string(filter[i])
	}
	return name
}

func stringInSlice(value string, list []string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func buildUserFilterCondition(filter interface{}) (builder.Cond, error) {
	switch f := filter.(type) {
	case message.FilterAnd:
		conditions := make([]builder.Cond, len(f))
		for i, v := range f {
			cond, err := buildUserFilterCondition(v)
			if err != nil {
				return nil, err
			}
			conditions[i] = cond
		}
		return builder.And(conditions...), nil
	case message.FilterOr:
		conditions := make([]builder.Cond, len(f))
		for i, v := range f {
			cond, err := buildUserFilterCondition(v)
			if err != nil {
				return nil, err
			}
			conditions[i] = cond
		}
		return builder.Or(conditions...), nil
	case message.FilterNot:
		cond, err := buildUserFilterCondition(f.Filter)
		if err != nil {
			return nil, err
		}
		return builder.Not{cond}, nil
	case message.FilterEqualityMatch:
		field, err := getUserFieldFromAttribute(string(f.AttributeDesc()))
		if err != nil {
			return nil, err
		}
		return builder.Eq{field: string(f.AssertionValue())}, nil
	case message.FilterPresent:
		field, err := getUserFieldFromAttribute(string(f))
		if err != nil {
			return nil, err
		}
		return builder.NotNull{field}, nil
	case message.FilterGreaterOrEqual:
		field, err := getUserFieldFromAttribute(string(f.AttributeDesc()))
		if err != nil {
			return nil, err
		}
		return builder.Gte{field: string(f.AssertionValue())}, nil
	case message.FilterLessOrEqual:
		field, err := getUserFieldFromAttribute(string(f.AttributeDesc()))
		if err != nil {
			return nil, err
		}
		return builder.Lte{field: string(f.AssertionValue())}, nil
	case message.FilterSubstrings:
		field, err := getUserFieldFromAttribute(string(f.Type_()))
		if err != nil {
			return nil, err
		}
		var expr string
		for _, substring := range f.Substrings() {
			switch s := substring.(type) {
			case message.SubstringInitial:
				expr += string(s) + "%"
				continue
			case message.SubstringAny:
				expr += string(s) + "%"
				continue
			case message.SubstringFinal:
				expr += string(s)
				continue
			}
		}
		return builder.Expr(field+" LIKE ?", expr), nil
	default:
		return nil, fmt.Errorf("LDAP filter operation %#v not supported", f)
	}
}

func buildSafeCondition(filter interface{}) builder.Cond {
	condition, err := buildUserFilterCondition(filter)
	if err != nil {
		log.Printf("err = %v", err.Error())
		return nil
	}
	return condition
}

func GetFilteredUsers(m *ldap.Message) (filteredUsers []*object.User, code int) {
	var err error
	r := m.GetSearchRequest()

	name, org, code := getNameAndOrgFromFilter(string(r.BaseObject()), r.FilterString())
	if code != ldap.LDAPResultSuccess {
		return nil, code
	}

	if name == "*" && m.Client.IsOrgAdmin { // get all users from organization 'org'
		if m.Client.IsGlobalAdmin && org == "*" {
			filteredUsers, err = object.GetGlobalUsersWithFilter(buildSafeCondition(r.Filter()))
			if err != nil {
				panic(err)
			}
			return filteredUsers, ldap.LDAPResultSuccess
		}
		if m.Client.IsGlobalAdmin || org == m.Client.OrgName {
			filteredUsers, err = object.GetUsersWithFilter(org, buildSafeCondition(r.Filter()))
			if err != nil {
				panic(err)
			}

			return filteredUsers, ldap.LDAPResultSuccess
		} else {
			return nil, ldap.LDAPResultInsufficientAccessRights
		}
	} else {
		requestUserId := util.GetId(m.Client.OrgName, m.Client.UserName)
		userId := util.GetId(org, name)

		hasPermission, err := object.CheckUserPermission(requestUserId, userId, true, "en")
		if !hasPermission {
			log.Printf("err = %v", err.Error())
			return nil, ldap.LDAPResultInsufficientAccessRights
		}

		user, err := object.GetUser(userId)
		if err != nil {
			panic(err)
		}

		if user != nil {
			filteredUsers = append(filteredUsers, user)
			return filteredUsers, ldap.LDAPResultSuccess
		}

		organization, err := object.GetOrganization(util.GetId("admin", org))
		if err != nil {
			panic(err)
		}

		if organization == nil {
			return nil, ldap.LDAPResultNoSuchObject
		}

		if !stringInSlice(name, organization.Tags) {
			return nil, ldap.LDAPResultNoSuchObject
		}

		users, err := object.GetUsersByTagWithFilter(org, name, buildSafeCondition(r.Filter()))
		if err != nil {
			panic(err)
		}

		filteredUsers = append(filteredUsers, users...)
		return filteredUsers, ldap.LDAPResultSuccess
	}
}

func GetFilteredOrganizations(m *ldap.Message) ([]*object.Organization, int) {
	if m.Client.IsGlobalAdmin {
		organizations, err := object.GetOrganizations("")
		if err != nil {
			panic(err)
		}
		return organizations, ldap.LDAPResultSuccess
	} else if m.Client.IsOrgAdmin {
		requestUserId := util.GetId(m.Client.OrgName, m.Client.UserName)
		user, err := object.GetUser(requestUserId)
		if err != nil {
			panic(err)
		}
		organization, err := object.GetOrganizationByUser(user)
		if err != nil {
			panic(err)
		}
		return []*object.Organization{organization}, ldap.LDAPResultSuccess
	} else {
		return nil, ldap.LDAPResultInsufficientAccessRights
	}
}

func GetFilteredGroups(m *ldap.Message) ([]*object.Group, int) {
	if m.Client.IsGlobalAdmin {
		groups, err := object.GetGroups("")
		if err != nil {
			panic(err)
		}
		return groups, ldap.LDAPResultSuccess
	} else if m.Client.IsOrgAdmin {
		requestUserId := util.GetId(m.Client.OrgName, m.Client.UserName)
		user, err := object.GetUser(requestUserId)
		if err != nil {
			panic(err)
		}
		groups, err := object.GetGroups(user.Owner)
		if err != nil {
			panic(err)
		}
		return groups, ldap.LDAPResultSuccess
	} else {
		return nil, ldap.LDAPResultInsufficientAccessRights
	}
}

// get user password with hash type prefix
// TODO not handle salt yet
// @return {md5}5f4dcc3b5aa765d61d8327deb882cf99
func getUserPasswordWithType(user *object.User) string {
	org, err := object.GetOrganizationByUser(user)
	if err != nil {
		panic(err)
	}

	if org.PasswordType == "" || org.PasswordType == "plain" {
		return user.Password
	}
	prefix := org.PasswordType
	if prefix == "salt" {
		prefix = "sha256"
	} else if prefix == "md5-salt" {
		prefix = "md5"
	} else if prefix == "pbkdf2-salt" {
		prefix = "pbkdf2"
	}
	return fmt.Sprintf("{%s}%s", prefix, user.Password)
}

func getUserFieldFromAttribute(attributeName string) (string, error) {
	v, ok := ldapUserAttributesMapping.CaseInsensitiveGet(attributeName)
	if !ok {
		return "", fmt.Errorf("attribute %s not supported", attributeName)
	}
	return v.GetField()
}

func searchFilterForEquality(filter message.Filter, desc string, values ...string) string {
	switch f := filter.(type) {
	case message.FilterAnd:
		for _, child := range f {
			if val := searchFilterForEquality(child, desc, values...); val != "" {
				return val
			}
		}
	case message.FilterOr:
		for _, child := range f {
			if val := searchFilterForEquality(child, desc, values...); val != "" {
				return val
			}
		}
	case message.FilterNot:
		return searchFilterForEquality(f.Filter, desc, values...)
	case message.FilterSubstrings:
		// Handle FilterSubstrings case if needed
	case message.FilterEqualityMatch:
		if strings.EqualFold(string(f.AttributeDesc()), desc) {
			for _, value := range values {
				if val := string(f.AssertionValue()); val == value {
					return val
				}
			}
		}
	case message.FilterGreaterOrEqual:
		// Handle FilterGreaterOrEqual case if needed
	case message.FilterLessOrEqual:
		// Handle FilterLessOrEqual case if needed
	case message.FilterPresent:
		// Handle FilterPresent case if needed
	case message.FilterApproxMatch:
		// Handle FilterApproxMatch case if needed
	}

	return ""
}
