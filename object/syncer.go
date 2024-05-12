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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type TableColumn struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	CasdoorName string   `json:"casdoorName"`
	IsKey       bool     `json:"isKey"`
	IsHashed    bool     `json:"isHashed"`
	Values      []string `json:"values"`
}

type Syncer struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	Type         string `xorm:"varchar(100)" json:"type"`
	DatabaseType string `xorm:"varchar(100)" json:"databaseType"`
	SslMode      string `xorm:"varchar(100)" json:"sslMode"`
	SshType      string `xorm:"varchar(100)" json:"sshType"`

	Host             string         `xorm:"varchar(100)" json:"host"`
	Port             int            `json:"port"`
	User             string         `xorm:"varchar(100)" json:"user"`
	Password         string         `xorm:"varchar(150)" json:"password"`
	SshHost          string         `xorm:"varchar(100)" json:"sshHost"`
	SshPort          int            `json:"sshPort"`
	SshUser          string         `xorm:"varchar(100)" json:"sshUser"`
	SshPassword      string         `xorm:"varchar(150)" json:"sshPassword"`
	Cert             string         `xorm:"varchar(100)" json:"cert"`
	Database         string         `xorm:"varchar(100)" json:"database"`
	Table            string         `xorm:"varchar(100)" json:"table"`
	TableColumns     []*TableColumn `xorm:"mediumtext" json:"tableColumns"`
	AffiliationTable string         `xorm:"varchar(100)" json:"affiliationTable"`
	AvatarBaseUrl    string         `xorm:"varchar(100)" json:"avatarBaseUrl"`
	ErrorText        string         `xorm:"mediumtext" json:"errorText"`
	SyncInterval     int            `json:"syncInterval"`
	IsReadOnly       bool           `json:"isReadOnly"`
	IsEnabled        bool           `json:"isEnabled"`

	Ormer *Ormer `xorm:"-" json:"-"`
}

func GetSyncerCount(owner, organization, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Syncer{Organization: organization})
}

func GetSyncers(owner string) ([]*Syncer, error) {
	syncers := []*Syncer{}
	err := ormer.Engine.Desc("created_time").Find(&syncers, &Syncer{Owner: owner})
	if err != nil {
		return syncers, err
	}

	return syncers, nil
}

func GetOrganizationSyncers(owner, organization string) ([]*Syncer, error) {
	syncers := []*Syncer{}
	err := ormer.Engine.Desc("created_time").Find(&syncers, &Syncer{Owner: owner, Organization: organization})
	if err != nil {
		return syncers, err
	}

	return syncers, nil
}

func GetPaginationSyncers(owner, organization string, offset, limit int, field, value, sortField, sortOrder string) ([]*Syncer, error) {
	syncers := []*Syncer{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&syncers, &Syncer{Organization: organization})
	if err != nil {
		return syncers, err
	}

	return syncers, nil
}

func getSyncer(owner string, name string) (*Syncer, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	syncer := Syncer{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&syncer)
	if err != nil {
		return &syncer, err
	}

	if existed {
		return &syncer, nil
	} else {
		return nil, nil
	}
}

func GetSyncer(id string) (*Syncer, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getSyncer(owner, name)
}

func GetMaskedSyncer(syncer *Syncer, errs ...error) (*Syncer, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	if syncer == nil {
		return nil, nil
	}

	if syncer.Password != "" {
		syncer.Password = "***"
	}
	return syncer, nil
}

func GetMaskedSyncers(syncers []*Syncer, errs ...error) ([]*Syncer, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	var err error
	for _, syncer := range syncers {
		syncer, err = GetMaskedSyncer(syncer)
		if err != nil {
			return nil, err
		}
	}

	return syncers, nil
}

func UpdateSyncer(id string, syncer *Syncer) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	s, err := getSyncer(owner, name)
	if err != nil {
		return false, err
	} else if s == nil {
		return false, nil
	}

	session := ormer.Engine.ID(core.PK{owner, name}).AllCols()
	if syncer.Password == "***" {
		syncer.Password = s.Password
	}
	affected, err := session.Update(syncer)
	if err != nil {
		return false, err
	}

	if affected == 1 {
		err = addSyncerJob(syncer)
		if err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func updateSyncerErrorText(syncer *Syncer, line string) (bool, error) {
	s, err := getSyncer(syncer.Owner, syncer.Name)
	if err != nil {
		return false, err
	}

	if s == nil {
		return false, nil
	}

	s.ErrorText = s.ErrorText + line

	affected, err := ormer.Engine.ID(core.PK{s.Owner, s.Name}).Cols("error_text").Update(s)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddSyncer(syncer *Syncer) (bool, error) {
	affected, err := ormer.Engine.Insert(syncer)
	if err != nil {
		return false, err
	}

	if affected == 1 {
		err = addSyncerJob(syncer)
		if err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func DeleteSyncer(syncer *Syncer) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{syncer.Owner, syncer.Name}).Delete(&Syncer{})
	if err != nil {
		return false, err
	}

	if affected == 1 {
		deleteSyncerJob(syncer)
	}

	return affected != 0, nil
}

func (syncer *Syncer) GetId() string {
	return fmt.Sprintf("%s/%s", syncer.Owner, syncer.Name)
}

func (syncer *Syncer) getTableColumnsTypeMap() map[string]string {
	m := map[string]string{}
	for _, tableColumn := range syncer.TableColumns {
		m[tableColumn.Name] = tableColumn.Type
	}
	return m
}

func (syncer *Syncer) getTable() string {
	if syncer.DatabaseType == "mssql" {
		return fmt.Sprintf("[%s]", syncer.Table)
	} else {
		return syncer.Table
	}
}

func (syncer *Syncer) getKeyColumn() *TableColumn {
	var column *TableColumn
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.IsKey {
			column = tableColumn
		}
	}

	if column == nil {
		for _, tableColumn := range syncer.TableColumns {
			if tableColumn.Name == "id" {
				column = tableColumn
			}
		}
	}

	if column == nil {
		column = syncer.TableColumns[0]
	}

	return column
}

func (syncer *Syncer) getKey() string {
	column := syncer.getKeyColumn()
	return util.CamelToSnakeCase(column.CasdoorName)
}

func RunSyncer(syncer *Syncer) error {
	err := syncer.initAdapter()
	if err != nil {
		return err
	}

	return syncer.syncUsers()
}

func TestSyncerDb(syncer Syncer) error {
	oldSyncer, err := getSyncer(syncer.Owner, syncer.Name)
	if err != nil {
		return err
	}

	if syncer.Password == "***" {
		syncer.Password = oldSyncer.Password
	}

	err = syncer.initAdapter()
	if err != nil {
		return err
	}

	err = syncer.Ormer.Engine.Ping()
	if err != nil {
		return err
	}
	return nil
}
