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
	Cert string `xorm:"varchar(100)" json:"cert"`
}

func AddLdap(ldap *Ldap) (bool, error) {
	if len(ldap.Id) == 0 {
		ldap.Id = util.GenerateId()
	}

	if len(ldap.CreatedTime) == 0 {
		ldap.CreatedTime = util.GetCurrentTime()
	}

	affected, err := adapter.Engine.Insert(ldap)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func CheckLdapExist(ldap *Ldap) (bool, error) {
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
		return false, err
	}

	if len(result) > 0 {
		return true, nil
	}

	return false, nil
}

func GetLdaps(owner string) ([]*Ldap, error) {
	var ldaps []*Ldap
	err := adapter.Engine.Desc("created_time").Find(&ldaps, &Ldap{Owner: owner})
	if err != nil {
		return ldaps, err
	}

	return ldaps, nil
}

func GetLdap(id string) (*Ldap, error) {
	if util.IsStringsEmpty(id) {
		return nil, nil
	}

	ldap := Ldap{Id: id}
	existed, err := adapter.Engine.Get(&ldap)
	if err != nil {
		return &ldap, nil
	}

	if existed {
		return &ldap, nil
	} else {
		return nil, nil
	}
}

func GetMaskedLdap(ldap *Ldap, errs ...error) (*Ldap, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	if ldap == nil {
		return nil, nil
	}

	if ldap.Password != "" {
		ldap.Password = "***"
	}

	return ldap, nil
}

func GetMaskedLdaps(ldaps []*Ldap, errs ...error) ([]*Ldap, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	var err error
	for _, ldap := range ldaps {
		ldap, err = GetMaskedLdap(ldap)
		if err != nil {
			return nil, err
		}
	}
	return ldaps, nil
}

func UpdateLdap(ldap *Ldap) (bool, error) {
	if l, err := GetLdap(ldap.Id); err != nil {
		return false, nil
	} else if l == nil {
		return false, nil
	}

	affected, err := adapter.Engine.ID(ldap.Id).Cols("owner", "server_name", "host",
		"port", "enable_ssl", "cert", "username", "password", "base_dn", "filter", "filter_fields", "auto_sync").Update(ldap)
	if err != nil {
		return false, nil
	}

	return affected != 0, nil
}

func DeleteLdap(ldap *Ldap) (bool, error) {
	affected, err := adapter.Engine.ID(ldap.Id).Delete(&Ldap{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
