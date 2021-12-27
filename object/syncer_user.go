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
	"fmt"
	"strings"
	"time"

	"github.com/casbin/casdoor/util"
	"xorm.io/core"
)

type OriginalUser = User

func (syncer *Syncer) getOriginalUsers() []*OriginalUser {
	sql := fmt.Sprintf("select * from %s", syncer.getTable())
	results, err := syncer.Adapter.Engine.QueryString(sql)
	if err != nil {
		panic(err)
	}

	return syncer.getOriginalUsersFromMap(results)
}

func (syncer *Syncer) getOriginalUserMap() ([]*OriginalUser, map[string]*OriginalUser) {
	users := syncer.getOriginalUsers()

	m := map[string]*OriginalUser{}
	for _, user := range users {
		m[user.Id] = user
	}
	return users, m
}

func (syncer *Syncer) addUser(user *OriginalUser) bool {
	m := syncer.getMapFromOriginalUser(user)
	keyString, valueString := syncer.getSqlKeyValueStringFromMap(m)

	sql := fmt.Sprintf("insert into %s (%s) values (%s)", syncer.getTable(), keyString, valueString)
	res, err := syncer.Adapter.Engine.Exec(sql)
	if err != nil {
		panic(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	return affected != 0
}

/*func (syncer *Syncer) getOriginalColumns() []string {
	res := []string{}
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.CasdoorName != "Id" {
			res = append(res, tableColumn.Name)
		}
	}
	return res
}*/

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

func (syncer *Syncer) updateUser(user *OriginalUser) bool {
	m := syncer.getMapFromOriginalUser(user)
	pkValue := m[syncer.TablePrimaryKey]
	delete(m, syncer.TablePrimaryKey)
	setString := syncer.getSqlSetStringFromMap(m)

	sql := fmt.Sprintf("update %s set %s where %s = %s", syncer.getTable(), setString, syncer.TablePrimaryKey, pkValue)
	res, err := syncer.Adapter.Engine.Exec(sql)
	if err != nil {
		panic(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (syncer *Syncer) updateUserForOriginalFields(user *User) bool {
	owner, name := util.GetOwnerAndNameFromId(user.GetId())
	oldUser := getUserById(owner, name)
	if oldUser == nil {
		return false
	}

	if user.Avatar != oldUser.Avatar && user.Avatar != "" {
		user.PermanentAvatar = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar)
	}

	columns := syncer.getCasdoorColumns()
	columns = append(columns, "affiliation", "hash", "pre_hash")
	affected, err := adapter.Engine.ID(core.PK{oldUser.Owner, oldUser.Name}).Cols(columns...).Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
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

func (syncer *Syncer) initAdapter() {
	if syncer.Adapter == nil {
		var dataSourceName string
		if syncer.DatabaseType == "mssql" {
			dataSourceName = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", syncer.User, syncer.Password, syncer.Host, syncer.Port, syncer.Database)
		} else {
			dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/", syncer.User, syncer.Password, syncer.Host, syncer.Port)
		}

		if !isCloudIntranet {
			dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
		}

		syncer.Adapter = NewAdapter(syncer.DatabaseType, dataSourceName, syncer.Database)
	}
}

func RunSyncUsersJob() {
	syncers := GetSyncers("admin")
	for _, syncer := range syncers {
		if !syncer.IsEnabled {
			continue
		}

		syncer.initAdapter()

		syncer.syncUsers()

		// run at every minute
		//schedule := fmt.Sprintf("* * * * %d", syncer.SyncInterval)
		schedule := "* * * * *"
		ctab := getCrontab(syncer.Name)
		err := ctab.AddJob(schedule, syncer.syncUsers)
		if err != nil {
			panic(err)
		}
	}

	time.Sleep(time.Duration(1<<63 - 1))
}
