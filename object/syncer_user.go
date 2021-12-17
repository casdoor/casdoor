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
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/casbin/casdoor/util"
)

type DbUser struct {
	Id        int    `xorm:"int notnull pk autoincr" json:"id"`
	Name      string `xorm:"varchar(128)" json:"name"`
	Password  string `xorm:"varchar(128)" json:"password"`
	Cellphone string `xorm:"varchar(128)" json:"cellphone"`
	SchoolId  int    `json:"schoolId"`
	Avatar    string `xorm:"varchar(128)" json:"avatar"`
	Deleted   int    `xorm:"tinyint(1)" json:"deleted"`
}

func (syncer *Syncer) getUsersOriginal() []*DbUser {
	users := []*DbUser{}
	err := syncer.Adapter.Engine.Table(syncer.Table).Asc("id").Find(&users)
	if err != nil {
		panic(err)
	}

	return users
}

func (syncer *Syncer) getUserMapOriginal() ([]*DbUser, map[string]*DbUser) {
	users := syncer.getUsersOriginal()

	m := map[string]*DbUser{}
	for _, user := range users {
		m[strconv.Itoa(user.Id)] = user
	}
	return users, m
}

func (syncer *Syncer) addUser(user *DbUser) bool {
	affected, err := syncer.Adapter.Engine.Table(syncer.Table).Insert(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (syncer *Syncer) updateUser(user *DbUser) bool {
	affected, err := syncer.Adapter.Engine.Table(syncer.Table).ID(user.Id).Cols("name", "password", "cellphone", "school_id", "avatar", "deleted").Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (syncer *Syncer) calculateHash(user *DbUser) string {
	s := strings.Join([]string{strconv.Itoa(user.Id), user.Password, user.Name, syncer.getFullAvatarUrl(user.Avatar), user.Cellphone, strconv.Itoa(user.SchoolId)}, "|")
	return util.GetMd5Hash(s)
}

func (syncer *Syncer) initAdapter() {
	if syncer.Adapter == nil {
		dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/", syncer.User, syncer.Password, syncer.Host, syncer.Port)
		syncer.Adapter = NewAdapter(beego.AppConfig.String("driverName"), dataSourceName, syncer.Database)
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
