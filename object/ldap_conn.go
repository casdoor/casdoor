// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"strings"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/thanhpk/randstr"
)

func (ldap *Ldap) GetLdapConn() (c *LdapConn, err error) {
	var conn *goldap.Conn
	if ldap.EnableSsl {
		conn, err = goldap.DialTLS("tcp", fmt.Sprintf("%s:%d", ldap.Host, ldap.Port), nil)
	} else {
		conn, err = goldap.Dial("tcp", fmt.Sprintf("%s:%d", ldap.Host, ldap.Port))
	}

	if err != nil {
		return nil, err
	}

	err = conn.Bind(ldap.Admin, ldap.Passwd)
	if err != nil {
		return nil, err
	}

	isAD, err := isMicrosoftAD(conn)
	if err != nil {
		return nil, err
	}
	return &LdapConn{Conn: conn, IsAD: isAD}, nil
}

func isMicrosoftAD(Conn *goldap.Conn) (bool, error) {
	SearchFilter := "(objectClass=*)"
	SearchAttributes := []string{"vendorname", "vendorversion", "isGlobalCatalogReady", "forestFunctionality"}

	searchReq := goldap.NewSearchRequest("",
		goldap.ScopeBaseObject, goldap.NeverDerefAliases, 0, 0, false,
		SearchFilter, SearchAttributes, nil)
	searchResult, err := Conn.Search(searchReq)
	if err != nil {
		return false, err
	}
	if len(searchResult.Entries) == 0 {
		return false, nil
	}
	isMicrosoft := false

	type ldapServerType struct {
		Vendorname           string
		Vendorversion        string
		IsGlobalCatalogReady string
		ForestFunctionality  string
	}
	var ldapServerTypes ldapServerType
	for _, entry := range searchResult.Entries {
		for _, attribute := range entry.Attributes {
			switch attribute.Name {
			case "vendorname":
				ldapServerTypes.Vendorname = attribute.Values[0]
			case "vendorversion":
				ldapServerTypes.Vendorversion = attribute.Values[0]
			case "isGlobalCatalogReady":
				ldapServerTypes.IsGlobalCatalogReady = attribute.Values[0]
			case "forestFunctionality":
				ldapServerTypes.ForestFunctionality = attribute.Values[0]
			}
		}
	}
	if ldapServerTypes.Vendorname == "" &&
		ldapServerTypes.Vendorversion == "" &&
		ldapServerTypes.IsGlobalCatalogReady == "TRUE" &&
		ldapServerTypes.ForestFunctionality != "" {
		isMicrosoft = true
	}
	return isMicrosoft, err
}

func LdapUsersToLdapRespUsers(users []ldapUser) []LdapRespUser {
	res := make([]LdapRespUser, 0)
	for _, user := range users {
		res = append(res, LdapRespUser{
			UidNumber:   user.UidNumber,
			Uid:         user.Uid,
			Cn:          user.Cn,
			GroupId:     user.GidNumber,
			Uuid:        user.Uuid,
			DisplayName: user.DisplayName,
			Email:       util.ReturnAnyNotEmpty(user.Email, user.EmailAddress, user.Mail),
			Phone:       util.ReturnAnyNotEmpty(user.Mobile, user.MobileTelephoneNumber, user.TelephoneNumber),
			Address:     util.ReturnAnyNotEmpty(user.PostalAddress, user.RegisteredAddress),
		})
	}
	return res
}

func SyncLdapUsers(owner string, respUsers []LdapRespUser, ldapId string) (*[]LdapRespUser, *[]LdapRespUser) {
	var existUsers []LdapRespUser
	var failedUsers []LdapRespUser
	var uuids []string

	for _, user := range respUsers {
		uuids = append(uuids, user.Uuid)
	}

	existUuids := CheckLdapUuidExist(owner, uuids)

	organization := getOrganization("admin", owner)
	ldap := GetLdap(ldapId)

	var dc []string
	for _, basedn := range strings.Split(ldap.BaseDn, ",") {
		if strings.Contains(basedn, "dc=") {
			dc = append(dc, basedn[3:])
		}
	}
	affiliation := strings.Join(dc, ".")

	var ou []string
	for _, admin := range strings.Split(ldap.Admin, ",") {
		if strings.Contains(admin, "ou=") {
			ou = append(ou, admin[3:])
		}
	}
	tag := strings.Join(ou, ".")

	for _, respUser := range respUsers {
		found := false
		if len(existUuids) > 0 {
			for _, existUuid := range existUuids {
				if respUser.Uuid == existUuid {
					existUsers = append(existUsers, respUser)
					found = true
				}
			}
		}

		if !found {
			newUser := &User{
				Owner:       owner,
				Name:        buildLdapUserName(&respUser),
				CreatedTime: util.GetCurrentTime(),
				DisplayName: respUser.DisplayName,
				Avatar:      organization.DefaultAvatar,
				Email:       respUser.Email,
				Phone:       respUser.Phone,
				Address:     []string{respUser.Address},
				Affiliation: affiliation,
				Tag:         tag,
				Score:       beego.AppConfig.DefaultInt("initScore", 2000),
				Ldap:        respUser.Uuid,
			}

			affected := AddUser(newUser)
			if !affected {
				failedUsers = append(failedUsers, respUser)
				continue
			}
		}
	}

	return &existUsers, &failedUsers
}

func CheckLdapUuidExist(owner string, uuids []string) []string {
	var results []User
	var existUuids []string
	existUuidSet := make(map[string]struct{})

	err := adapter.Engine.Where(fmt.Sprintf("ldap IN (%s) AND owner = ?", "'"+strings.Join(uuids, "','")+"'"), owner).Find(&results)
	if err != nil {
		panic(err)
	}

	if len(results) > 0 {
		for _, result := range results {
			existUuidSet[result.Ldap] = struct{}{}
		}
	}

	for uuid := range existUuidSet {
		existUuids = append(existUuids, uuid)
	}
	return existUuids
}

func buildLdapUserName(ldapUser *LdapRespUser) string {
	user := User{}
	uidWithNumber := fmt.Sprintf("%s_%s", ldapUser.Uid, ldapUser.UidNumber)
	has, err := adapter.Engine.Where("name = ? or name = ?", ldapUser.Uid, uidWithNumber).Get(&user)
	if err != nil {
		panic(err)
	}

	if has {
		if user.Name == ldapUser.Uid {
			return uidWithNumber
		}
		return fmt.Sprintf("%s_%s", uidWithNumber, randstr.Hex(6))
	}

	return ldapUser.Uid
}

func (ldap *Ldap) buildFilterString(user *User) string {
	if len(ldap.FilterFields) == 0 {
		return fmt.Sprintf("(&%s(uid=%s))", ldap.Filter, user.Name)
	}

	filter := fmt.Sprintf("(&%s(|", ldap.Filter)
	for _, field := range ldap.FilterFields {
		filter = fmt.Sprintf("%s(%s=%s)", filter, field, user.getFieldFromLdapAttribute(field))
	}
	filter = fmt.Sprintf("%s))", filter)

	return filter
}

func (user *User) getFieldFromLdapAttribute(attribute string) string {
	switch attribute {
	case "uid":
		return user.Name
	case "mail":
		return user.Email
	case "mobile":
		return user.Phone
	default:
		return ""
	}
}
