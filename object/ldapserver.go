// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

	ldap "github.com/forestmgy/ldapserver"
	"github.com/lor00x/goldap/message"
)

func GetNameAndOrgFromDN(DN string) (name, org, err string) {
	DNvalue := strings.Split(DN, ",")
	if len(DNvalue) == 1 || strings.ToLower(DNvalue[0])[0] != 'c' || strings.ToLower(DNvalue[1])[0] != 'o' {
		return "", "", "please use correct Admin Name format like cn=xxx,ou=xxx,dc=example,dc=com"
	}
	return DNvalue[0][3:], DNvalue[1][3:], ""
}

func GetUserNameAndOrgFromBaseDnAndFilter(BaseDN, Filter string) (name, org string, errCode int) {
	if !strings.Contains(BaseDN, "ou=") || !strings.Contains(Filter, "cn=") {
		name, org = "", ""
		errCode = ldap.LDAPResultInvalidDNSyntax
		return
	}
	name = getUserNameFromFilter(Filter)
	_, org, _ = GetNameAndOrgFromDN(fmt.Sprintf("cn=%s,", name) + BaseDN)
	errCode = ldap.LDAPResultSuccess
	return
}

func getUserNameFromFilter(Filter string) string {
	nameIndex := strings.Index(Filter, "cn=")
	var name string
	for i := nameIndex + 3; Filter[i] != ')'; i++ {
		name = name + string(Filter[i])
	}
	return name
}

func PrintSearchInfo(r message.SearchRequest) {
	log.Printf("Request BaseDn=%s", r.BaseObject())
	log.Printf("Request Filter=%s", r.Filter())
	log.Printf("Request FilterString=%s", r.FilterString())
	log.Printf("Request Attributes=%s", r.Attributes())
	log.Printf("Request TimeLimit=%d", r.TimeLimit().Int())
}

func GetFilteredUsers(m *ldap.Message, name, org string) ([]*User, int) {
	var filteredUsers []*User
	if name == "*" && m.Client.IsOrgAdmin { // get all users from organization 'org'
		if m.Client.OrgName == "built-in" && org == "*" {
			filteredUsers = GetGlobalUsers()
			return filteredUsers, ldap.LDAPResultSuccess
		} else if m.Client.OrgName == "built-in" || org == m.Client.OrgName {
			filteredUsers = GetUsers(org)
			return filteredUsers, ldap.LDAPResultSuccess
		} else {
			return nil, ldap.LDAPResultInsufficientAccessRights
		}
	} else {
		hasPermission, err := CheckUserPermission(fmt.Sprintf("%s/%s", m.Client.OrgName, m.Client.UserName), fmt.Sprintf("%s/%s", org, name), org, true)
		if !hasPermission {
			log.Printf("ErrMsg = %v", err.Error())
			return nil, ldap.LDAPResultInsufficientAccessRights
		}
		user := getUser(org, name)
		filteredUsers = append(filteredUsers, user)
		return filteredUsers, ldap.LDAPResultSuccess
	}
}
