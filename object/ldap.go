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
	"github.com/casdoor/casdoor/util"
)

type Ldap struct {
	Id          string `xorm:"varchar(100) notnull pk" json:"id"`
	Owner       string `xorm:"varchar(100)" json:"owner"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	ServerName   string   `xorm:"varchar(100)" json:"serverName"`
	Host         string   `xorm:"varchar(100)" json:"host"`
	Port         int      `xorm:"int" json:"port"`
	EnableSsl    bool     `xorm:"bool" json:"enableSsl"`
	Username     string   `xorm:"varchar(100)" json:"username"`
	Password     string   `xorm:"varchar(100)" json:"password"`
	BaseDn       string   `xorm:"varchar(100)" json:"baseDn"`
	Filter       string   `xorm:"varchar(200)" json:"filter"`
	FilterFields []string `xorm:"varchar(100)" json:"filterFields"`

	AutoSync int    `json:"autoSync"`
	LastSync string `xorm:"varchar(100)" json:"lastSync"`
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
		Owner:    ldap.Owner,
		Host:     ldap.Host,
		Port:     ldap.Port,
		Username: ldap.Username,
		Password: ldap.Password,
		BaseDn:   ldap.BaseDn,
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
		"port", "enable_ssl", "username", "password", "base_dn", "filter", "filter_fields", "auto_sync").Update(ldap)
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
