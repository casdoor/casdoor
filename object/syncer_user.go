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
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/casdoor/casdoor/util"
	"github.com/go-sql-driver/mysql"
)

type OriginalUser = User

type Credential struct {
	Value string `json:"value"`
	Salt  string `json:"salt"`
}

func (syncer *Syncer) getOriginalUsers() ([]*OriginalUser, error) {
	var results []map[string]sql.NullString
	err := syncer.Ormer.Engine.Table(syncer.getTable()).Find(&results)
	if err != nil {
		return nil, err
	}

	// Memory leak problem handling
	// https://github.com/casdoor/casdoor/issues/1256
	users := syncer.getOriginalUsersFromMap(results)
	for _, m := range results {
		for k := range m {
			delete(m, k)
		}
	}

	return users, nil
}

func (syncer *Syncer) addUser(user *OriginalUser) (bool, error) {
	m := syncer.getMapFromOriginalUser(user)
	affected, err := syncer.Ormer.Engine.Table(syncer.getTable()).Insert(m)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (syncer *Syncer) getCasdoorColumns() []string {
	res := []string{}
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.CasdoorName != "Id" {
			v := util.CamelToSnakeCase(tableColumn.CasdoorName)
			res = append(res, v)
		}
	}
	return res
}

func (syncer *Syncer) updateUser(user *OriginalUser) (bool, error) {
	key := syncer.getKey()
	m := syncer.getMapFromOriginalUser(user)
	pkValue := m[key]
	delete(m, key)

	affected, err := syncer.Ormer.Engine.Table(syncer.getTable()).Where(fmt.Sprintf("%s = ?", key), pkValue).Update(&m)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (syncer *Syncer) updateUserForOriginalFields(user *User, key string) (bool, error) {
	var err error
	oldUser := User{}

	existed, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getUserValue(user, key), user.Owner).Get(&oldUser)
	if err != nil {
		return false, err
	}
	if !existed {
		return false, nil
	}

	if user.Avatar != oldUser.Avatar && user.Avatar != "" {
		user.PermanentAvatar, err = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar, true)
		if err != nil {
			return false, err
		}
	}

	columns := syncer.getCasdoorColumns()
	columns = append(columns, "affiliation", "hash", "pre_hash")
	affected, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getUserValue(&oldUser, key), oldUser.Owner).Cols(columns...).Update(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (syncer *Syncer) calculateHash(user *OriginalUser) string {
	values := []string{}
	m := syncer.getMapFromOriginalUser(user)
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.IsHashed {
			values = append(values, m[tableColumn.Name])
		}
	}

	s := strings.Join(values, "|")
	return util.GetMd5Hash(s)
}

type dsnConnector struct {
	dsn    string
	driver driver.Driver
}

func (t dsnConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return t.driver.Open(t.dsn)
}

func (t dsnConnector) Driver() driver.Driver {
	return t.driver
}

func (syncer *Syncer) initAdapter() error {
	if syncer.Ormer != nil {
		return nil
	}

	var dataSourceName string
	if syncer.DatabaseType == "mssql" {
		dataSourceName = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", syncer.User, syncer.Password, syncer.Host, syncer.Port, syncer.Database)
	} else if syncer.DatabaseType == "postgres" {
		sslMode := "disable"
		if syncer.SslMode != "" {
			sslMode = syncer.SslMode
		}
		dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=%s dbname=%s", syncer.User, syncer.Password, syncer.Host, syncer.Port, sslMode, syncer.Database)
	} else {
		dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/", syncer.User, syncer.Password, syncer.Host, syncer.Port)
	}

	var db *sql.DB
	var err error

	if syncer.SshType != "" && (syncer.DatabaseType == "mysql" || syncer.DatabaseType == "postgres" || syncer.DatabaseType == "mssql") {
		var dial *ssh.Client
		if syncer.SshType == "password" {
			dial, err = DialWithPassword(syncer.SshUser, syncer.SshPassword, syncer.SshHost, syncer.SshPort)
		} else {
			dial, err = DialWithCert(syncer.SshUser, syncer.Owner+"/"+syncer.Cert, syncer.SshHost, syncer.SshPort)
		}
		if err != nil {
			return err
		}

		if syncer.DatabaseType == "mysql" {
			dataSourceName = fmt.Sprintf("%s:%s@%s(%s:%d)/", syncer.User, syncer.Password, syncer.Owner+syncer.Name, syncer.Host, syncer.Port)
			mysql.RegisterDialContext(syncer.Owner+syncer.Name, (&ViaSSHDialer{Client: dial, Context: nil}).MysqlDial)
		} else if syncer.DatabaseType == "postgres" || syncer.DatabaseType == "mssql" {
			db = sql.OpenDB(dsnConnector{dsn: dataSourceName, driver: &ViaSSHDialer{Client: dial, Context: nil, DatabaseType: syncer.DatabaseType}})
		}
	}

	if !isCloudIntranet {
		dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
	}

	if db != nil {
		syncer.Ormer, err = NewAdapterFromDb(syncer.DatabaseType, dataSourceName, syncer.Database, db)
	} else {
		syncer.Ormer, err = NewAdapter(syncer.DatabaseType, dataSourceName, syncer.Database)
	}

	return err
}

func RunSyncUsersJob() {
	syncers, err := GetSyncers("admin")
	if err != nil {
		panic(err)
	}

	for _, syncer := range syncers {
		err = addSyncerJob(syncer)
		if err != nil {
			panic(err)
		}
	}

	time.Sleep(time.Duration(1<<63 - 1))
}
