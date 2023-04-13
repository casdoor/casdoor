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

	"github.com/casdoor/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
)

type Ldap struct {
	Id          string `xorm:"varchar(100) notnull pk" json:"id"`
	Owner       string `xorm:"varchar(100)" json:"owner"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	ServerName   string   `xorm:"varchar(100)" json:"serverName"`
	Host         string   `xorm:"varchar(100)" json:"host"`
	Port         int      `xorm:"int" json:"port"`
	EnableSsl    bool     `xorm:"bool" json:"enableSsl"`
	Admin        string   `xorm:"varchar(100)" json:"admin"`
	Passwd       string   `xorm:"varchar(100)" json:"passwd"`
	BaseDn       string   `xorm:"varchar(100)" json:"baseDn"`
	Filter       string   `xorm:"varchar(200)" json:"filter"`
	FilterFields []string `xorm:"varchar(100)" json:"filterFields"`

	AutoSync int    `json:"autoSync"`
	LastSync string `xorm:"varchar(100)" json:"lastSync"`
}

type LdapConn struct {
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
	DisplayName           string
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
	Uuid        string `json:"uuid"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
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

func (l *LdapConn) GetLdapUsers(ldapServer *Ldap) ([]ldapUser, error) {
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
		"port", "enable_ssl", "admin", "passwd", "base_dn", "filter", "filter_fields", "auto_sync").Update(ldap)
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
