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
	"errors"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/thanhpk/randstr"
)

type LdapConn struct {
	Conn *goldap.Conn
	IsAD bool
}

//type ldapGroup struct {
//	GidNumber string
//	Cn        string
//}

type LdapUser struct {
	UidNumber string `json:"uidNumber"`
	Uid       string `json:"uid"`
	Cn        string `json:"cn"`
	GidNumber string `json:"gidNumber"`
	// Gcn                   string
	Uuid                  string `json:"uuid"`
	DisplayName           string `json:"displayName"`
	Mail                  string
	Email                 string `json:"email"`
	EmailAddress          string
	TelephoneNumber       string
	Mobile                string
	MobileTelephoneNumber string
	RegisteredAddress     string
	PostalAddress         string

	GroupId string `json:"groupId"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

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

	err = conn.Bind(ldap.Username, ldap.Password)
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

func (l *LdapConn) GetLdapUsers(ldapServer *Ldap) ([]LdapUser, error) {
	SearchAttributes := []string{
		"uidNumber", "cn", "sn", "gidNumber", "entryUUID", "displayName", "mail", "email",
		"emailAddress", "telephoneNumber", "mobile", "mobileTelephoneNumber", "registeredAddress", "postalAddress",
	}
	if l.IsAD {
		SearchAttributes = append(SearchAttributes, "sAMAccountName")
	} else {
		SearchAttributes = append(SearchAttributes, "uid")
	}

	searchReq := goldap.NewSearchRequest(ldapServer.BaseDn, goldap.ScopeWholeSubtree, goldap.NeverDerefAliases,
		0, 0, false,
		ldapServer.Filter, SearchAttributes, nil)
	searchResult, err := l.Conn.SearchWithPaging(searchReq, 100)
	if err != nil {
		return nil, err
	}

	if len(searchResult.Entries) == 0 {
		return nil, errors.New("no result")
	}

	var ldapUsers []LdapUser
	for _, entry := range searchResult.Entries {
		var user LdapUser
		for _, attribute := range entry.Attributes {
			switch attribute.Name {
			case "uidNumber":
				user.UidNumber = attribute.Values[0]
			case "uid":
				user.Uid = attribute.Values[0]
			case "sAMAccountName":
				user.Uid = attribute.Values[0]
			case "cn":
				user.Cn = attribute.Values[0]
			case "gidNumber":
				user.GidNumber = attribute.Values[0]
			case "entryUUID":
				user.Uuid = attribute.Values[0]
			case "objectGUID":
				user.Uuid = attribute.Values[0]
			case "displayName":
				user.DisplayName = attribute.Values[0]
			case "mail":
				user.Mail = attribute.Values[0]
			case "email":
				user.Email = attribute.Values[0]
			case "emailAddress":
				user.EmailAddress = attribute.Values[0]
			case "telephoneNumber":
				user.TelephoneNumber = attribute.Values[0]
			case "mobile":
				user.Mobile = attribute.Values[0]
			case "mobileTelephoneNumber":
				user.MobileTelephoneNumber = attribute.Values[0]
			case "registeredAddress":
				user.RegisteredAddress = attribute.Values[0]
			case "postalAddress":
				user.PostalAddress = attribute.Values[0]
			}
		}
		ldapUsers = append(ldapUsers, user)
	}

	return ldapUsers, nil
}

// FIXME: The Base DN does not necessarily contain the Group
//
//	func (l *ldapConn) GetLdapGroups(baseDn string) (map[string]ldapGroup, error) {
//		SearchFilter := "(objectClass=posixGroup)"
//		SearchAttributes := []string{"cn", "gidNumber"}
//		groupMap := make(map[string]ldapGroup)
//
//		searchReq := goldap.NewSearchRequest(baseDn,
//			goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
//			SearchFilter, SearchAttributes, nil)
//		searchResult, err := l.Conn.Search(searchReq)
//		if err != nil {
//			return nil, err
//		}
//
//		if len(searchResult.Entries) == 0 {
//			return nil, errors.New("no result")
//		}
//
//		for _, entry := range searchResult.Entries {
//			var ldapGroupItem ldapGroup
//			for _, attribute := range entry.Attributes {
//				switch attribute.Name {
//				case "gidNumber":
//					ldapGroupItem.GidNumber = attribute.Values[0]
//					break
//				case "cn":
//					ldapGroupItem.Cn = attribute.Values[0]
//					break
//				}
//			}
//			groupMap[ldapGroupItem.GidNumber] = ldapGroupItem
//		}
//
//		return groupMap, nil
//	}

func AutoAdjustLdapUser(users []LdapUser) []LdapUser {
	res := make([]LdapUser, len(users))
	for i, user := range users {
		res[i] = LdapUser{
			UidNumber:         user.UidNumber,
			Uid:               user.Uid,
			Cn:                user.Cn,
			GroupId:           user.GidNumber,
			Uuid:              user.GetLdapUuid(),
			DisplayName:       user.DisplayName,
			Email:             util.ReturnAnyNotEmpty(user.Email, user.EmailAddress, user.Mail),
			Mobile:            util.ReturnAnyNotEmpty(user.Mobile, user.MobileTelephoneNumber, user.TelephoneNumber),
			RegisteredAddress: util.ReturnAnyNotEmpty(user.PostalAddress, user.RegisteredAddress),
		}
	}
	return res
}

func SyncLdapUsers(owner string, syncUsers []LdapUser, ldapId string) (existUsers []LdapUser, failedUsers []LdapUser, err error) {
	var uuids []string
	for _, user := range syncUsers {
		uuids = append(uuids, user.Uuid)
	}

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
	for _, admin := range strings.Split(ldap.Username, ",") {
		if strings.Contains(admin, "ou=") {
			ou = append(ou, admin[3:])
		}
	}
	tag := strings.Join(ou, ".")

	for _, syncUser := range syncUsers {
		existUuids := GetExistUuids(owner, uuids)
		found := false
		if len(existUuids) > 0 {
			for _, existUuid := range existUuids {
				if syncUser.Uuid == existUuid {
					existUsers = append(existUsers, syncUser)
					found = true
				}
			}
		}

		if !found {
			score, _ := organization.GetInitScore()
			newUser := &User{
				Owner:       owner,
				Name:        syncUser.buildLdapUserName(),
				CreatedTime: util.GetCurrentTime(),
				DisplayName: syncUser.buildLdapDisplayName(),
				Avatar:      organization.DefaultAvatar,
				Email:       syncUser.Email,
				Phone:       syncUser.Phone,
				Address:     []string{syncUser.Address},
				Affiliation: affiliation,
				Tag:         tag,
				Score:       score,
				Ldap:        syncUser.Uuid,
			}

			affected := AddUser(newUser)
			if !affected {
				failedUsers = append(failedUsers, syncUser)
				continue
			}
		}
	}

	return existUsers, failedUsers, err
}

func GetExistUuids(owner string, uuids []string) []string {
	var existUuids []string

	err := adapter.Engine.Table("user").Where("owner = ?", owner).Cols("ldap").
		In("ldap", uuids).Select("DISTINCT ldap").Find(&existUuids)
	if err != nil {
		panic(err)
	}

	return existUuids
}

func (ldapUser *LdapUser) buildLdapUserName() string {
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

	if ldapUser.Uid != "" {
		return ldapUser.Uid
	}

	return ldapUser.Cn
}

func (ldapUser *LdapUser) buildLdapDisplayName() string {
	if ldapUser.DisplayName != "" {
		return ldapUser.DisplayName
	}

	return ldapUser.Cn
}

func (ldapUser *LdapUser) GetLdapUuid() string {
	if ldapUser.Uuid != "" {
		return ldapUser.Uuid
	}
	if ldapUser.Uid != "" {
		return ldapUser.Uid
	}

	return ldapUser.Cn
}

func (ldap *Ldap) buildAuthFilterString(user *User) string {
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
	case "sAMAccountName":
		return user.Name
	case "mail":
		return user.Email
	case "mobile":
		return user.Phone
	default:
		return ""
	}
}
