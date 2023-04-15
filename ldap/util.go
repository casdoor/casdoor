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

	ldap "github.com/forestmgy/ldapserver"
)

func getNameAndOrgFromDN(DN string) (string, string, string) {
	DNFields := strings.Split(DN, ",")
	params := make(map[string]string, len(DNFields))
	for _, field := range DNFields {
		if strings.Contains(field, "=") {
			k := strings.Split(field, "=")
			params[k[0]] = k[1]
		}
	}

	if params["cn"] == "" {
		return "", "", "please use Admin Name format like cn=xxx,ou=xxx,dc=example,dc=com"
	}
	if params["ou"] == "" {
		return params["cn"], object.CasdoorOrganization, ""
	}
	return params["cn"], params["ou"], ""
}

func getNameAndOrgFromFilter(baseDN, filter string) (string, string, int) {
	if !strings.Contains(baseDN, "ou=") {
		return "", "", ldap.LDAPResultInvalidDNSyntax
	}

	name, org, _ := getNameAndOrgFromDN(fmt.Sprintf("cn=%s,", getUsername(filter)) + baseDN)
	return name, org, ldap.LDAPResultSuccess
}

func getUsername(filter string) string {
	nameIndex := strings.Index(filter, "cn=")
	if nameIndex == -1 {
		return "*"
	}

	var name string
	for i := nameIndex + 3; filter[i] != ')'; i++ {
		name = name + string(filter[i])
	}
	return name
}

func GetFilteredUsers(m *ldap.Message) (filteredUsers []*object.User, code int) {
	r := m.GetSearchRequest()

	name, org, code := getNameAndOrgFromFilter(string(r.BaseObject()), r.FilterString())
	if code != ldap.LDAPResultSuccess {
		return nil, code
	}

	if name == "*" && m.Client.IsOrgAdmin { // get all users from organization 'org'
		if m.Client.IsGlobalAdmin && org == "*" {
			filteredUsers = object.GetGlobalUsers()
			return filteredUsers, ldap.LDAPResultSuccess
		}
		if m.Client.IsGlobalAdmin || org == m.Client.OrgName {
			filteredUsers = object.GetUsers(org)
			return filteredUsers, ldap.LDAPResultSuccess
		} else {
			return nil, ldap.LDAPResultInsufficientAccessRights
		}
	} else {
		hasPermission, err := object.CheckUserPermission(fmt.Sprintf("%s/%s", m.Client.OrgName, m.Client.UserName), fmt.Sprintf("%s/%s", org, name), true, "en")
		if !hasPermission {
			log.Printf("ErrMsg = %v", err.Error())
			return nil, ldap.LDAPResultInsufficientAccessRights
		}
		user := object.GetUser(util.GetId(org, name))
		filteredUsers = append(filteredUsers, user)
		return filteredUsers, ldap.LDAPResultSuccess
	}
}

// get user password with hash type prefix
// TODO not handle salt yet
// @return {md5}5f4dcc3b5aa765d61d8327deb882cf99
func getUserPasswordWithType(user *object.User) string {
	org := object.GetOrganizationByUser(user)
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
	switch attributeName {
	case "cn":
		return message.AttributeValue(user.Name)
	case "uid":
		return message.AttributeValue(user.Name)
	case "email":
		return message.AttributeValue(user.Email)
	case "mobile":
		return message.AttributeValue(user.Phone)
	case "userPassword":
		return message.AttributeValue(getUserPasswordWithType(user))
	default:
		return ""
	}
}
