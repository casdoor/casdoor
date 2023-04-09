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
	"errors"
	"fmt"
	"strings"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/thanhpk/randstr"
)

type Ldap struct {
	Id          string `xorm:"varchar(100) notnull pk" json:"id"`
	Owner       string `xorm:"varchar(100)" json:"owner"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	ServerName string `xorm:"varchar(100)" json:"serverName"`
	Host       string `xorm:"varchar(100)" json:"host"`
	Port       int    `xorm:"int" json:"port"`
	EnableSsl  bool   `xorm:"bool" json:"enableSsl"`
	Admin      string `xorm:"varchar(100)" json:"admin"`
	Passwd     string `xorm:"varchar(100)" json:"passwd"`
	BaseDn     string `xorm:"varchar(100)" json:"baseDn"`
	Filter     string `xorm:"varchar(200)" json:"filter"`

	AutoSync int    `json:"autoSync"`
	LastSync string `xorm:"varchar(100)" json:"lastSync"`
}

type ldapConn struct {
	Conn *goldap.Conn
	IsAD bool
}

//type ldapGroup struct {
//	GidNumber string
//	Cn        string
//}

type ldapUser struct {
	UidNumber string
	Uid       string
	Cn        string
	GidNumber string
	// Gcn                   string
	Uuid                  string
	Mail                  string
	Email                 string
	EmailAddress          string
	TelephoneNumber       string
	Mobile                string
	MobileTelephoneNumber string
	RegisteredAddress     string
	PostalAddress         string
}

type LdapRespUser struct {
	UidNumber string `json:"uidNumber"`
	Uid       string `json:"uid"`
	Cn        string `json:"cn"`
	GroupId   string `json:"groupId"`
	// GroupName string `json:"groupName"`
	Uuid    string `json:"uuid"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type ldapServerType struct {
	Vendorname           string
	Vendorversion        string
	IsGlobalCatalogReady string
	ForestFunctionality  string
}

func LdapUsersToLdapRespUsers(users []ldapUser) []LdapRespUser {
	returnAnyNotEmpty := func(strs ...string) string {
		for _, str := range strs {
			if str != "" {
				return str
			}
		}
		return ""
	}

	res := make([]LdapRespUser, 0)
	for _, user := range users {
		res = append(res, LdapRespUser{
			UidNumber: user.UidNumber,
			Uid:       user.Uid,
			Cn:        user.Cn,
			GroupId:   user.GidNumber,
			Uuid:      user.Uuid,
			Email:     returnAnyNotEmpty(user.Email, user.EmailAddress, user.Mail),
			Phone:     returnAnyNotEmpty(user.Mobile, user.MobileTelephoneNumber, user.TelephoneNumber),
			Address:   returnAnyNotEmpty(user.PostalAddress, user.RegisteredAddress),
		})
	}
	return res
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
	var ldapServerType ldapServerType
	for _, entry := range searchResult.Entries {
		for _, attribute := range entry.Attributes {
			switch attribute.Name {
			case "vendorname":
				ldapServerType.Vendorname = attribute.Values[0]
			case "vendorversion":
				ldapServerType.Vendorversion = attribute.Values[0]
			case "isGlobalCatalogReady":
				ldapServerType.IsGlobalCatalogReady = attribute.Values[0]
			case "forestFunctionality":
				ldapServerType.ForestFunctionality = attribute.Values[0]
			}
		}
	}
	if ldapServerType.Vendorname == "" &&
		ldapServerType.Vendorversion == "" &&
		ldapServerType.IsGlobalCatalogReady == "TRUE" &&
		ldapServerType.ForestFunctionality != "" {
		isMicrosoft = true
	}
	return isMicrosoft, err
}

func (ldap *Ldap) GetLdapConn() (c *ldapConn, err error) {
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
	return &ldapConn{Conn: conn, IsAD: isAD}, nil
}

//FIXME: The Base DN does not necessarily contain the Group
//func (l *ldapConn) GetLdapGroups(baseDn string) (map[string]ldapGroup, error) {
//	SearchFilter := "(objectClass=posixGroup)"
//	SearchAttributes := []string{"cn", "gidNumber"}
//	groupMap := make(map[string]ldapGroup)
//
//	searchReq := goldap.NewSearchRequest(baseDn,
//		goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
//		SearchFilter, SearchAttributes, nil)
//	searchResult, err := l.Conn.Search(searchReq)
//	if err != nil {
//		return nil, err
//	}
//
//	if len(searchResult.Entries) == 0 {
//		return nil, errors.New("no result")
//	}
//
//	for _, entry := range searchResult.Entries {
//		var ldapGroupItem ldapGroup
//		for _, attribute := range entry.Attributes {
//			switch attribute.Name {
//			case "gidNumber":
//				ldapGroupItem.GidNumber = attribute.Values[0]
//				break
//			case "cn":
//				ldapGroupItem.Cn = attribute.Values[0]
//				break
//			}
//		}
//		groupMap[ldapGroupItem.GidNumber] = ldapGroupItem
//	}
//
//	return groupMap, nil
//}

func (l *ldapConn) GetLdapUsers(ldapServer *Ldap) ([]ldapUser, error) {
	var SearchAttributes []string
	if l.IsAD {
		SearchAttributes = []string{
			"uidNumber", "sAMAccountName", "cn", "gidNumber", "entryUUID", "mail", "email",
			"emailAddress", "telephoneNumber", "mobile", "mobileTelephoneNumber", "registeredAddress", "postalAddress",
		}
	} else {
		SearchAttributes = []string{
			"uidNumber", "uid", "cn", "gidNumber", "entryUUID", "mail", "email",
			"emailAddress", "telephoneNumber", "mobile", "mobileTelephoneNumber", "registeredAddress", "postalAddress",
		}
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

	var ldapUsers []ldapUser
	for _, entry := range searchResult.Entries {
		var user ldapUser
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

func AddLdap(ldap *Ldap) bool {
	if len(ldap.Id) == 0 {
		ldap.Id = util.GenerateId()
	}

	if len(ldap.CreatedTime) == 0 {
		ldap.CreatedTime = util.GetCurrentTime()
	}

	affected, err := adapter.Engine.Insert(ldap)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func CheckLdapExist(ldap *Ldap) bool {
	var result []*Ldap
	err := adapter.Engine.Find(&result, &Ldap{
		Owner:  ldap.Owner,
		Host:   ldap.Host,
		Port:   ldap.Port,
		Admin:  ldap.Admin,
		Passwd: ldap.Passwd,
		BaseDn: ldap.BaseDn,
	})
	if err != nil {
		panic(err)
	}

	if len(result) > 0 {
		return true
	}

	return false
}

func GetLdaps(owner string) []*Ldap {
	var ldaps []*Ldap
	err := adapter.Engine.Desc("created_time").Find(&ldaps, &Ldap{Owner: owner})
	if err != nil {
		panic(err)
	}

	return ldaps
}

func GetLdap(id string) *Ldap {
	if util.IsStringsEmpty(id) {
		return nil
	}

	ldap := Ldap{Id: id}
	existed, err := adapter.Engine.Get(&ldap)
	if err != nil {
		panic(err)
	}

	if existed {
		return &ldap
	} else {
		return nil
	}
}

func UpdateLdap(ldap *Ldap) bool {
	if GetLdap(ldap.Id) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(ldap.Id).Cols("owner", "server_name", "host",
		"port", "enable_ssl", "admin", "passwd", "base_dn", "filter", "auto_sync").Update(ldap)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteLdap(ldap *Ldap) bool {
	affected, err := adapter.Engine.ID(ldap.Id).Delete(&Ldap{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func SyncLdapUsers(owner string, users []LdapRespUser, ldapId string) (*[]LdapRespUser, *[]LdapRespUser) {
	var existUsers []LdapRespUser
	var failedUsers []LdapRespUser
	var uuids []string

	for _, user := range users {
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

	for _, user := range users {
		found := false
		if len(existUuids) > 0 {
			for _, existUuid := range existUuids {
				if user.Uuid == existUuid {
					existUsers = append(existUsers, user)
					found = true
				}
			}
		}

		if !found && !AddUser(&User{
			Owner:       owner,
			Name:        buildLdapUserName(user.Uid, user.UidNumber),
			CreatedTime: util.GetCurrentTime(),
			DisplayName: user.Cn,
			Avatar:      organization.DefaultAvatar,
			Email:       user.Email,
			Phone:       user.Phone,
			Address:     []string{user.Address},
			Affiliation: affiliation,
			Tag:         tag,
			Score:       beego.AppConfig.DefaultInt("initScore", 2000),
			Ldap:        user.Uuid,
		}) {
			failedUsers = append(failedUsers, user)
			continue
		}
	}

	return &existUsers, &failedUsers
}

func UpdateLdapSyncTime(ldapId string) {
	_, err := adapter.Engine.ID(ldapId).Update(&Ldap{LastSync: util.GetCurrentTime()})
	if err != nil {
		panic(err)
	}
}

func CheckLdapUuidExist(owner string, uuids []string) []string {
	var results []User
	var existUuids []string
	existUuidSet := make(map[string]struct{})

	//whereStr := ""
	//for i, uuid := range uuids {
	//	if i == 0 {
	//		whereStr = fmt.Sprintf("'%s'", uuid)
	//	} else {
	//		whereStr = fmt.Sprintf(",'%s'", uuid)
	//	}
	//}

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

func buildLdapUserName(uid, uidNum string) string {
	var result User
	uidWithNumber := fmt.Sprintf("%s_%s", uid, uidNum)

	has, err := adapter.Engine.Where("name = ? or name = ?", uid, uidWithNumber).Get(&result)
	if err != nil {
		panic(err)
	}

	if has {
		if result.Name == uid {
			return uidWithNumber
		}
		return fmt.Sprintf("%s_%s", uidWithNumber, randstr.Hex(6))
	}

	return uid
}
