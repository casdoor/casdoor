// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"github.com/casbin/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/thanhpk/randstr"
	"strings"
)

type Ldap struct {
	Id          string `xorm:"varchar(100) notnull pk" json:"id"`
	Owner       string `xorm:"varchar(100)" json:"owner"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	ServerName string `xorm:"varchar(100)" json:"serverName"`
	Host       string `xorm:"varchar(100)" json:"host"`
	Port       int    `json:"port"`
	Admin      string `xorm:"varchar(100)" json:"admin"`
	Passwd     string `xorm:"varchar(100)" json:"passwd"`
	BaseDn     string `xorm:"varchar(100)" json:"baseDn"`

	AutoSync int    `json:"autoSync"`
	LastSync string `xorm:"varchar(100)" json:"lastSync"`
}

type ldapConn struct {
	Conn *goldap.Conn
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
	//Gcn                   string
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
	//GroupName string `json:"groupName"`
	Uuid    string `json:"uuid"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

func GetLdapConn(host string, port int, adminUser string, adminPasswd string) (*ldapConn, error) {
	conn, err := goldap.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	err = conn.Bind(adminUser, adminPasswd)
	if err != nil {
		return nil, fmt.Errorf("fail to login Ldap server with [%s]", adminUser)
	}

	return &ldapConn{Conn: conn}, nil
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

func (l *ldapConn) GetLdapUsers(baseDn string) ([]ldapUser, error) {
	SearchFilter := "(objectClass=posixAccount)"
	SearchAttributes := []string{"uidNumber", "uid", "cn", "gidNumber", "entryUUID", "mail", "email",
		"emailAddress", "telephoneNumber", "mobile", "mobileTelephoneNumber", "registeredAddress", "postalAddress"}

	searchReq := goldap.NewSearchRequest(baseDn,
		goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
		SearchFilter, SearchAttributes, nil)
	searchResult, err := l.Conn.Search(searchReq)
	if err != nil {
		return nil, err
	}

	if len(searchResult.Entries) == 0 {
		return nil, errors.New("no result")
	}

	var ldapUsers []ldapUser

	for _, entry := range searchResult.Entries {
		var ldapUserItem ldapUser
		for _, attribute := range entry.Attributes {
			switch attribute.Name {
			case "uidNumber":
				ldapUserItem.UidNumber = attribute.Values[0]
			case "uid":
				ldapUserItem.Uid = attribute.Values[0]
			case "cn":
				ldapUserItem.Cn = attribute.Values[0]
			case "gidNumber":
				ldapUserItem.GidNumber = attribute.Values[0]
			case "entryUUID":
				ldapUserItem.Uuid = attribute.Values[0]
			case "mail":
				ldapUserItem.Mail = attribute.Values[0]
			case "email":
				ldapUserItem.Email = attribute.Values[0]
			case "emailAddress":
				ldapUserItem.EmailAddress = attribute.Values[0]
			case "telephoneNumber":
				ldapUserItem.TelephoneNumber = attribute.Values[0]
			case "mobile":
				ldapUserItem.Mobile = attribute.Values[0]
			case "mobileTelephoneNumber":
				ldapUserItem.MobileTelephoneNumber = attribute.Values[0]
			case "registeredAddress":
				ldapUserItem.RegisteredAddress = attribute.Values[0]
			case "postalAddress":
				ldapUserItem.PostalAddress = attribute.Values[0]
			}
		}
		ldapUsers = append(ldapUsers, ldapUserItem)
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
	if util.IsStrsEmpty(id) {
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
		"port", "admin", "passwd", "base_dn", "auto_sync").Update(ldap)
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

func SyncLdapUsers(owner string, users []LdapRespUser) (*[]LdapRespUser, *[]LdapRespUser) {
	var existUsers []LdapRespUser
	var failedUsers []LdapRespUser
	var uuids []string

	for _, user := range users {
		uuids = append(uuids, user.Uuid)
	}

	existUuids := CheckLdapUuidExist(owner, uuids)

	for _, user := range users {
		if len(existUuids) > 0 {
			for index, existUuid := range existUuids {
				if user.Uuid == existUuid {
					existUsers = append(existUsers, user)
					existUuids = append(existUuids[:index], existUuids[index+1:]...)
				}
			}
		}
		if !AddUser(&User{
			Owner:       owner,
			Name:        buildLdapUserName(user.Uid, user.UidNumber),
			CreatedTime: util.GetCurrentTime(),
			Password:    "123",
			DisplayName: user.Cn,
			Avatar:      "https://casbin.org/img/casbin.svg",
			Email:       user.Email,
			Phone:       user.Phone,
			Address:     []string{user.Address},
			Affiliation: "Example Inc.",
			Tag:         "staff",
			Score:       2000,
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
			existUuids = append(existUuids, result.Ldap)
		}
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
