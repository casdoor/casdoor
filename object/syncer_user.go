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
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

type OriginalUser = User

func (syncer *Syncer) getOriginalUsers() ([]*OriginalUser, error) {
	provider := GetSyncerProvider(syncer)
	return provider.GetOriginalUsers()
}

func (syncer *Syncer) addUser(user *OriginalUser) (bool, error) {
	provider := GetSyncerProvider(syncer)
	return provider.AddUser(user)
}

func (syncer *Syncer) getCasdoorColumns() []string {
	// For API-based syncers, return a default set of common columns to update
	if syncer.isApiBasedSyncer() {
		return []string{
			"display_name", "email", "phone", "avatar", "title",
			"is_forbidden", "created_time", "updated_time",
		}
	}

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
	provider := GetSyncerProvider(syncer)
	return provider.UpdateUser(user)
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

	// For API-based syncers, hash a default set of important fields
	if syncer.isApiBasedSyncer() {
		// Hash key user fields to detect changes
		fields := []string{"Name", "DisplayName", "Email", "Phone", "Avatar", "Title", "IsForbidden"}
		for _, field := range fields {
			values = append(values, m[field])
		}
	} else {
		// For database syncers, use the configured columns
		for _, tableColumn := range syncer.TableColumns {
			if tableColumn.IsHashed {
				values = append(values, m[tableColumn.Name])
			}
		}
	}

	s := strings.Join(values, "|")
	return util.GetMd5Hash(s)
}

func (syncer *Syncer) initAdapter() error {
	provider := GetSyncerProvider(syncer)
	return provider.InitAdapter()
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
