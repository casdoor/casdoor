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

package original

import (
	"strconv"
	"strings"
	"time"

	"github.com/casbin/casdoor/util"
)

type User struct {
	Id        int    `xorm:"int notnull pk autoincr" json:"id"`
	Name      string `xorm:"varchar(128)" json:"name"`
	Password  string `xorm:"varchar(128)" json:"password"`
	Cellphone string `xorm:"varchar(128)" json:"cellphone"`
	SchoolId  int    `json:"schoolId"`
	Avatar    string `xorm:"varchar(128)" json:"avatar"`
	Deleted   int    `xorm:"tinyint(1)" json:"deleted"`
}

func (User) TableName() string {
	return userTableName
}

func getUsersOriginal() []*User {
	users := []*User{}
	err := adapter.Engine.Asc("id").Find(&users)
	if err != nil {
		panic(err)
	}

	return users
}

func getUserMapOriginal() ([]*User, map[string]*User) {
	users := getUsersOriginal()

	m := map[string]*User{}
	for _, user := range users {
		m[strconv.Itoa(user.Id)] = user
	}
	return users, m
}

func addUser(user *User) bool {
	affected, err := adapter.Engine.Insert(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func updateUser(user *User) bool {
	affected, err := adapter.Engine.ID(user.Id).Cols("name", "password", "cellphone", "school_id", "avatar", "deleted").Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func calculateHash(user *User) string {
	s := strings.Join([]string{strconv.Itoa(user.Id), user.Password, user.Name, getFullAvatarUrl(user.Avatar), user.Cellphone, strconv.Itoa(user.SchoolId)}, "|")
	return util.GetMd5Hash(s)
}

func RunSyncUsersJob() {
	syncUsers()

	// run at every minute
	schedule := "* * * * *"
	err := ctab.AddJob(schedule, syncUsers)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Duration(1<<63 - 1))
}
