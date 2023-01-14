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

package object

import (
	"fmt"
	"log"
	"strings"

	"github.com/forestmgy/ldapserver"
)

func GetNameAndOrgFromDN(DN string) (string, string, string) {
	DNValue := strings.Split(DN, ",")
	if len(DNValue) == 1 || strings.ToLower(DNValue[0])[0] != 'c' || strings.ToLower(DNValue[1])[0] != 'o' {
		return "", "", "please use correct Admin Name format like cn=xxx,ou=xxx,dc=example,dc=com"
	}
	return DNValue[0][3:], DNValue[1][3:], ""
}

func GetUserNameAndOrgFromBaseDnAndFilter(baseDN, filter string) (string, string, int) {
	if !strings.Contains(baseDN, "ou=") || !strings.Contains(filter, "cn=") {
		return "", "", ldapserver.LDAPResultInvalidDNSyntax
	}
	name := getUserNameFromFilter(filter)
	_, org, _ := GetNameAndOrgFromDN(fmt.Sprintf("cn=%s,", name) + baseDN)
	errCode := ldapserver.LDAPResultSuccess
	return name, org, errCode
}

func getUserNameFromFilter(filter string) string {
	nameIndex := strings.Index(filter, "cn=")
	var name string
	for i := nameIndex + 3; filter[i] != ')'; i++ {
		name = name + string(filter[i])
	}
	return name
}

func GetFilteredUsers(m *ldapserver.Message, name, org string) ([]*User, int) {
	var filteredUsers []*User
	if name == "*" && m.Client.IsOrgAdmin { // get all users from organization 'org'
		if m.Client.OrgName == "built-in" && org == "*" {
			filteredUsers = GetGlobalUsers()
			return filteredUsers, ldapserver.LDAPResultSuccess
		} else if m.Client.OrgName == "built-in" || org == m.Client.OrgName {
			filteredUsers = GetUsers(org)
			return filteredUsers, ldapserver.LDAPResultSuccess
		} else {
			return nil, ldapserver.LDAPResultInsufficientAccessRights
		}
	} else {
		hasPermission, err := CheckUserPermission(fmt.Sprintf("%s/%s", m.Client.OrgName, m.Client.UserName), fmt.Sprintf("%s/%s", org, name), org, true, "en")
		if !hasPermission {
			log.Printf("ErrMsg = %v", err.Error())
			return nil, ldapserver.LDAPResultInsufficientAccessRights
		}
		user := getUser(org, name)
		filteredUsers = append(filteredUsers, user)
		return filteredUsers, ldapserver.LDAPResultSuccess
	}
}
