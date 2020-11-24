// Copyright 2020 The casbin Authors. All Rights Reserved.
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

package store

import (
	"github.com/casdoor/casdoor/internal/object"
	"github.com/casdoor/casdoor/internal/store/shared"
	"github.com/casdoor/casdoor/pkg/util"
	"xorm.io/core"
)

type UserStore struct {
	db *shared.DB
}

func NewUserStore(db *shared.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (u *UserStore) GetUsers(owner string) ([]*object.User, error) {
	var users []*object.User
	err := u.db.GetEngine().Desc("created_time").Find(&users, &object.User{Owner: owner})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (u *UserStore) getUser(owner string, name string) (*object.User, error) {
	user := object.User{Owner: owner, Name: name}
	existed, err := u.db.GetEngine().Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func (u *UserStore) GetUser(id string) (*object.User, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return u.getUser(owner, name)
}

func (u *UserStore) UpdateUser(id string, user *object.User) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)

	data, err := u.getUser(owner, name)
	if err != nil || data == nil {
		return false, err
	}

	_, err = u.db.GetEngine().Id(core.PK{owner, name}).AllCols().Update(user)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (u *UserStore) AddUser(user *object.User) (bool, error) {
	affected, err := u.db.GetEngine().Insert(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (u *UserStore) DeleteUser(user *object.User) (bool, error) {
	affected, err := u.db.GetEngine().Id(core.PK{user.Owner, user.Name}).Delete(&object.User{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
