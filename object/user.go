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

	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type User struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Id            string `xorm:"varchar(100)" json:"id"`
	Password      string `xorm:"varchar(100)" json:"password"`
	PasswordType  string `xorm:"varchar(100)" json:"passwordType"`
	DisplayName   string `xorm:"varchar(100)" json:"displayName"`
	Avatar        string `xorm:"varchar(255)" json:"avatar"`
	Email         string `xorm:"varchar(100)" json:"email"`
	Phone         string `xorm:"varchar(100)" json:"phone"`
	Affiliation   string `xorm:"varchar(100)" json:"affiliation"`
	Tag           string `xorm:"varchar(100)" json:"tag"`
	IsAdmin       bool   `json:"isAdmin"`
	IsGlobalAdmin bool   `json:"isGlobalAdmin"`

	Github string `xorm:"varchar(100)" json:"github"`
	Google string `xorm:"varchar(100)" json:"google"`
}

func GetGlobalUsers() []*User {
	users := []*User{}
	err := adapter.engine.Desc("created_time").Find(&users)
	if err != nil {
		panic(err)
	}

	return users
}

func GetUsers(owner string) []*User {
	users := []*User{}
	err := adapter.engine.Desc("created_time").Find(&users, &User{Owner: owner})
	if err != nil {
		panic(err)
	}

	return users
}

func getUser(owner string, name string) *User {
	user := User{Owner: owner, Name: name}
	existed, err := adapter.engine.Get(&user)
	if err != nil {
		panic(err)
	}

	if existed {
		return &user
	} else {
		return nil
	}
}

func GetUser(id string) *User {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getUser(owner, name)
}

func HasUser(id string) bool {
	return GetUser(id) != nil
}

func IsPasswordCorrect(userId string, password string) bool {
	user := GetUser(userId)
	return user.Password == password
}

func UpdateUser(id string, user *User) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getUser(owner, name) == nil {
		return false
	}

	_, err := adapter.engine.ID(core.PK{owner, name}).AllCols().Update(user)
	if err != nil {
		panic(err)
	}

	//return affected != 0
	return true
}

func AddUser(user *User) bool {
	user.Id = util.GenerateId()
	affected, err := adapter.engine.Insert(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteUser(user *User) bool {
	affected, err := adapter.engine.ID(core.PK{user.Owner, user.Name}).Delete(&User{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetUserByField(organizationName string, field string, value string) *User {
	user := User{Owner: organizationName}
	existed, err := adapter.engine.Where(fmt.Sprintf("%s=?", field), value).Get(&user)
	if err != nil {
		panic(err)
	}

	if existed {
		return &user
	} else {
		return nil
	}
}

func LinkUserAccount(user, field, value string) bool {
	affected, err := adapter.engine.Table(new(User)).ID(user).Update(map[string]interface{}{field: value})
	if err != nil {
		panic(err)
	}

	return affected != 0
}
