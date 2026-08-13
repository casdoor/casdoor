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
	"strconv"
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
	"c": {userField: "region", fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.Region)
	}},
	"co": {userField: "region", fieldMapper: func(user *object.User) message.AttributeValue {
		return message.AttributeValue(user.Region)
	}},
	"userPassword": {
		userField:     "userPassword",
		notSearchable: true,
		fieldMapper: func(user *object.User) message.AttributeValue {
			return message.AttributeValue(getUserPasswordWithType(user))
		},
	},
	"loginShell": {
		userField:     "loginShell",
		notSearchable: true,
		fieldMapper: func(user *object.User) message.AttributeValue {
			// Check user properties first, otherwise return default shell
			if user.Properties != nil {
				if shell, ok := user.Properties["loginShell"]; ok && shell != "" {
					return message.AttributeValue(shell)
				}
			}
			return message.AttributeValue("/bin/bash")
		},
	},
	"gecos": {
		userField:     "gecos",
		notSearchable: true,
		fieldMapper: func(user *object.User) message.AttributeValue {
			// GECOS field typically contains full name and other user info
			// Format: Full Name,Room Number,Work Phone,Home Phone,Other
			gecos := user.DisplayName
			if gecos == "" {
				gecos = user.Name
			}
			return message.AttributeValue(gecos)
		},
	},
	"sshPublicKey": {
		userField:     "sshPublicKey",
		notSearchable: true,
		fieldMapper: func(user *object.User) message.AttributeValue {
			// Return SSH public key from user properties
			if user.Properties != nil {
				if sshKey, ok := user.Properties["sshPublicKey"]; ok && sshKey != "" {
					return message.AttributeValue(sshKey)
				}
			}
			return message.AttributeValue("")
		},
	},
}

const ldapMemberOfAttr = "memberOf"

// syntheticUserAttribute is a POSIX attribute the LDAP server computes, so a
// filter on it is finished in memory. column holds the assigned value, if any.
type syntheticUserAttribute struct {
	column   string
	getValue func(user *object.User) string
}

var syntheticUserAttributes = map[string]syntheticUserAttribute{
	"uidnumber":     {column: "uid_number", getValue: getUidNumber},
	"gidnumber":     {column: "uid_number", getValue: getUidNumber},
	"homedirectory": {getValue: getHomeDirectory},
}

// buildCoarseCondition keeps the rows that can still match: those the value is
// assigned to, plus the unassigned ones whose value is only known in memory.
func (a syntheticUserAttribute) buildCoarseCondition(value string) builder.Cond {
	if a.column == "" {
		return builder.Expr("1 = 1")
	}

	unassigned := builder.Or(builder.Eq{a.column: 0}, builder.IsNull{a.column})
	number, err := strconv.Atoi(value)
	if err != nil {
		return unassigned
	}
	return builder.Or(builder.Eq{a.column: number}, unassigned)
}

// getUidNumber derives the number from the name when none is assigned, so the
// uid of an existing user stays stable across the upgrade.
func getUidNumber(user *object.User) string {
	if user.UidNumber != 0 {
		return strconv.Itoa(user.UidNumber)
	}
	return fmt.Sprintf("%v", hash(user.Name))
}

func getHomeDirectory(user *object.User) string {
	return "/home/" + user.Name
}

func getGidNumber(group *object.Group) string {
	if group.GidNumber != 0 {
		return strconv.Itoa(group.GidNumber)
	}
	return fmt.Sprintf("%v", hash(group.Name))
}

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
		if kv := strings.SplitN(field, "=", 2); len(kv) == 2 {
			// Attribute names are case-insensitive in LDAP; normalize the key
			// so cn=/CN=/Cn= all match. Values stay as-is (org names are
			// case-sensitive in Casdoor).
			key := strings.ToLower(strings.TrimSpace(kv[0]))
			params[key] = strings.TrimSpace(kv[1])
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

// IsLdapAttrAllowed checks whether the given LDAP attribute is allowed for the organization.
// An empty filter or a filter containing "All" means all attributes are allowed.
func IsLdapAttrAllowed(org *object.Organization, attr string) bool {
	if org == nil || len(org.LdapAttributes) == 0 {
		return true
	}
	for _, f := range org.LdapAttributes {
		if strings.EqualFold(f, "All") || strings.EqualFold(f, attr) {
			return true
		}
	}
	return false
}

// userSearchFilter is an LDAP filter translated for the user table: a SQL
// condition, plus the terms that have to be applied to the rows it returned.
type userSearchFilter struct {
	condition  builder.Cond
	predicates []func(user *object.User) bool
}

func (q *userSearchFilter) matches(user *object.User) bool {
	for _, predicate := range q.predicates {
		if !predicate(user) {
			return false
		}
	}
	return true
}

func (q *userSearchFilter) apply(users []*object.User) []*object.User {
	if len(q.predicates) == 0 {
		return users
	}

	res := make([]*object.User, 0, len(users))
	for _, user := range users {
		if q.matches(user) {
			res = append(res, user)
		}
	}
	return res
}

func buildUserFilter(filter interface{}) (*userSearchFilter, error) {
	q := &userSearchFilter{}
	condition, err := q.buildCondition(filter, true)
	if err != nil {
		return nil, err
	}

	q.condition = condition
	return q, nil
}

// buildCondition translates the filter into a SQL condition. conjunctive means
// the term is reached through AND branches only, the only place a synthetic
// attribute can leave the query without changing the result.
func (q *userSearchFilter) buildCondition(filter interface{}, conjunctive bool) (builder.Cond, error) {
	switch f := filter.(type) {
	case message.FilterAnd:
		conditions := make([]builder.Cond, len(f))
		for i, v := range f {
			cond, err := q.buildCondition(v, conjunctive)
			if err != nil {
				return nil, err
			}
			conditions[i] = cond
		}
		return builder.And(conditions...), nil
	case message.FilterOr:
		conditions := make([]builder.Cond, len(f))
		for i, v := range f {
			cond, err := q.buildCondition(v, false)
			if err != nil {
				return nil, err
			}
			conditions[i] = cond
		}
		return builder.Or(conditions...), nil
	case message.FilterNot:
		cond, err := q.buildCondition(f.Filter, false)
		if err != nil {
			return nil, err
		}
		return builder.Not{cond}, nil
	case message.FilterEqualityMatch:
		attr := string(f.AttributeDesc())

		if strings.EqualFold(attr, "objectclass") && strings.EqualFold(string(f.AssertionValue()), "posixAccount") {
			return builder.Expr("1 = 1"), nil
		}

		if attribute, ok := syntheticUserAttributes[strings.ToLower(attr)]; ok {
			if !conjunctive {
				return nil, fmt.Errorf("attribute %s is only supported in an AND filter", attr)
			}

			value := string(f.AssertionValue())
			q.predicates = append(q.predicates, func(user *object.User) bool {
				return attribute.getValue(user) == value
			})
			return attribute.buildCoarseCondition(value), nil
		}

		if attr == ldapMemberOfAttr {
			var names []string
			groupId := string(f.AssertionValue())
			// Accept the DN form (cn=name,ou=owner,...) emitted by memberOf and
			// map it back to the Casdoor group id "owner/name" for the lookup.
			if strings.Contains(strings.ToLower(groupId), "cn=") {
				if name, org, err := getNameAndOrgFromDN(groupId); err == nil {
					groupId = util.GetId(org, name)
				}
			}
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
		if strings.EqualFold(string(f), "objectclass") {
			return builder.Expr("1 = 1"), nil
		}
		// Synthetic attributes are computed for every user, so they always exist.
		if _, ok := syntheticUserAttributes[strings.ToLower(string(f))]; ok {
			return builder.Expr("1 = 1"), nil
		}
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

func buildSafeFilter(filter interface{}) *userSearchFilter {
	q, err := buildUserFilter(filter)
	if err != nil {
		log.Printf("err = %v", err.Error())
		return &userSearchFilter{condition: builder.And(builder.Expr("1 != 1"))}
	}
	return q
}

func GetFilteredUsers(m *ldap.Message) (filteredUsers []*object.User, code int) {
	var err error
	r := m.GetSearchRequest()

	name, org, code := getNameAndOrgFromFilter(string(r.BaseObject()), r.FilterString())
	if code != ldap.LDAPResultSuccess {
		return nil, code
	}

	filter := buildSafeFilter(r.Filter())

	if name == "*" { // get all users from organization 'org'
		if m.Client.IsGlobalAdmin && org == "*" {
			filteredUsers, err = object.GetGlobalUsersWithFilter(filter.condition)
			if err != nil {
				panic(err)
			}
			return filter.apply(filteredUsers), ldap.LDAPResultSuccess
		}
		if m.Client.IsGlobalAdmin || (m.Client.IsOrgAdmin && org == m.Client.OrgName) {
			filteredUsers, err = object.GetUsersWithFilter(org, filter.condition)
			if err != nil {
				panic(err)
			}

			return filter.apply(filteredUsers), ldap.LDAPResultSuccess
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
			if !filter.matches(user) {
				return nil, ldap.LDAPResultSuccess
			}
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

		users, err := object.GetUsersByTagWithFilter(org, name, filter.condition)
		if err != nil {
			panic(err)
		}

		filteredUsers = append(filteredUsers, filter.apply(users)...)
		return filteredUsers, ldap.LDAPResultSuccess
	}
}

// GetFilteredGroups returns the groups of the organization the search is scoped
// to. The filter itself is applied by matchGroupFilter on each candidate entry.
func GetFilteredGroups(m *ldap.Message, baseDN string, filterStr string) ([]*object.Group, int) {
	_, org, code := getNameAndOrgFromFilter(baseDN, filterStr)
	if code != ldap.LDAPResultSuccess {
		return nil, code
	}

	var groups []*object.Group
	var err error

	if m.Client.IsGlobalAdmin && org == "*" {
		groups, err = object.GetGlobalGroups()
		if err != nil {
			panic(err)
		}
	} else if m.Client.IsGlobalAdmin || (m.Client.IsOrgAdmin && org == m.Client.OrgName) {
		groups, err = object.GetGroups(org)
		if err != nil {
			panic(err)
		}
	} else {
		return nil, ldap.LDAPResultInsufficientAccessRights
	}

	return groups, ldap.LDAPResultSuccess
}

// isPosixGroupFilter reports whether the filter asks for posixGroup entries,
// on its own or combined, e.g. "(&(objectClass=posixGroup)(gidNumber=12345))".
func isPosixGroupFilter(filter interface{}) bool {
	switch f := filter.(type) {
	case message.FilterAnd:
		for _, v := range f {
			if isPosixGroupFilter(v) {
				return true
			}
		}
	case message.FilterEqualityMatch:
		return strings.EqualFold(string(f.AttributeDesc()), "objectClass") &&
			strings.EqualFold(string(f.AssertionValue()), "posixGroup")
	}
	return false
}

// getGroupAttributes returns the attributes of the posixGroup entry published
// for a Casdoor group, keyed by lowercased attribute name.
func getGroupAttributes(group *object.Group, memberUids []string) map[string][]string {
	return map[string][]string{
		"cn":          {group.Name},
		"gidnumber":   {getGidNumber(group)},
		"memberuid":   memberUids,
		"objectclass": {"posixGroup"},
	}
}

// matchGroupFilter evaluates an LDAP filter against a group entry in memory,
// since group entries have no table to query.
func matchGroupFilter(filter interface{}, attributes map[string][]string) bool {
	switch f := filter.(type) {
	case message.FilterAnd:
		for _, v := range f {
			if !matchGroupFilter(v, attributes) {
				return false
			}
		}
		return true
	case message.FilterOr:
		for _, v := range f {
			if matchGroupFilter(v, attributes) {
				return true
			}
		}
		return false
	case message.FilterNot:
		return !matchGroupFilter(f.Filter, attributes)
	case message.FilterEqualityMatch:
		for _, value := range attributes[strings.ToLower(string(f.AttributeDesc()))] {
			if strings.EqualFold(value, string(f.AssertionValue())) {
				return true
			}
		}
		return false
	case message.FilterPresent:
		return len(attributes[strings.ToLower(string(f))]) > 0
	case message.FilterSubstrings:
		for _, value := range attributes[strings.ToLower(string(f.Type_()))] {
			if matchSubstrings(value, f.Substrings()) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func matchSubstrings(value string, substrings []message.Substring) bool {
	rest := strings.ToLower(value)
	for _, substring := range substrings {
		switch s := substring.(type) {
		case message.SubstringInitial:
			prefix := strings.ToLower(string(s))
			if !strings.HasPrefix(rest, prefix) {
				return false
			}
			rest = rest[len(prefix):]
		case message.SubstringAny:
			any := strings.ToLower(string(s))
			i := strings.Index(rest, any)
			if i < 0 {
				return false
			}
			rest = rest[i+len(any):]
		case message.SubstringFinal:
			if !strings.HasSuffix(rest, strings.ToLower(string(s))) {
				return false
			}
			rest = ""
		}
	}
	return true
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
