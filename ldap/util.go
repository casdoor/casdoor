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

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/lor00x/goldap/message"

	ldap "github.com/casdoor/ldapserver"

	"github.com/xorm-io/builder"
)

type AttributeMapper func(user *object.User) message.AttributeValue

type FieldRelation struct {
	userField     string
	notSearchable bool
	hideOnStarOp  bool
	fieldMapper   AttributeMapper
}

func (rel FieldRelation) GetField() (string, error) {
	if rel.notSearchable {
		return "", fmt.Errorf("attribute %s not supported", rel.userField)
	}
	return rel.userField, nil
}

func (rel FieldRelation) GetAttributeValue(user *object.User) message.AttributeValue {
	return rel.fieldMapper(user)
}

var ldapAttributesMapping = map[string]FieldRelation{
	"cn": {userField: "name", hideOnStarOp: true, fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.Name)
	}},
	"uid": {userField: "name", hideOnStarOp: true, fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.Name)
	}},
	"displayname": {userField: "displayName", fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.DisplayName)
	}},
	"email": {userField: "email", fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.Email)
	}},
	"mail": {userField: "email", fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.Email)
	}},
	"mobile": {userField: "phone", fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.Phone)
	}},
	"title": {userField: "tag", fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.Tag)
	}},
	"userPassword": {
		userField:     "userPassword",
		notSearchable: true,
		fieldMapper: func(user *object.User) message.AttributeValue {
			return message.AttributeValue(getUserPasswordWithType(user))
		},
	},
}

const ldapMemberOfAttr = "memberOf"

var AdditionalLdapAttributes []message.LDAPString

func init() {
	for k, v := range ldapAttributesMapping {
		if v.hideOnStarOp {
			continue
		}
		AdditionalLdapAttributes = append(AdditionalLdapAttributes, message.LDAPString(k))
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
		attr := string(f.AttributeDesc())

		if attr == ldapMemberOfAttr {
			var names []string
			groupId := string(f.AssertionValue())
			users := object.GetGroupUsersWithoutError(groupId)
			for _, user := range users {
				names = append(names, user.Name)
			}
			return builder.In("name", names), nil
		}

		field, err := getUserFieldFromAttribute(attr)
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
		return builder.And(builder.Expr("1 != 1"))
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

	if name == "*" { // get all users from organization 'org'
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

func getAttribute(attributeName string, user *object.User) message.AttributeValue {
	v, ok := ldapAttributesMapping[attributeName]
	if !ok {
		return ""
	}
	return v.GetAttributeValue(user)
}

func getUserFieldFromAttribute(attributeName string) (string, error) {
	v, ok := ldapAttributesMapping[attributeName]
	if !ok {
		return "", fmt.Errorf("attribute %s not supported", attributeName)
	}
	return v.GetField()
}
