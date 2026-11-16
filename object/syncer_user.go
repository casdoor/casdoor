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
	res := []string{}
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.CasdoorName != "Id" {
			v := util.CamelToSnakeCase(tableColumn.CasdoorName)
			// CasdoorName is org-admin controlled and these values become column names in
			// an UPDATE SET clause, so drop anything that is not a plain SQL identifier.
			if !util.FilterSQLIdentifier(v) {
				continue
			}
			res = append(res, v)
		}
	}
	return res
}

func (syncer *Syncer) updateUser(user *OriginalUser) (bool, error) {
	provider := GetSyncerProvider(syncer)
	return provider.UpdateUser(user)
}

// getUpdateColumns returns the columns that updateUserForOriginalFields() writes
// when it updates a user that already exists in Casdoor.
func (syncer *Syncer) getUpdateColumns(user *OriginalUser) []string {
	columns := syncer.getCasdoorColumns()
	columns = append(columns, "affiliation", "hash", "pre_hash")

	// Skip password-related columns when the incoming user has no password data.
	// API-based syncers (DingTalk, WeCom, Lark, etc.) do not provide passwords,
	// so updating these columns would wipe out locally set passwords.
	if user.Password == "" {
		filtered := make([]string, 0, len(columns))
		for _, col := range columns {
			if col != "password" && col != "password_salt" && col != "password_type" {
				filtered = append(filtered, col)
			}
		}
		columns = filtered
	}

	// Add provider-specific field for API-based syncers to enable login binding
	// This allows synced users to login via their provider accounts.
	//
	// is_forbidden is part of the update for the same providers: they derive it
	// from the account state upstream (resigned, frozen, not activated, exited,
	// suspended, disabled...), so the value has to be written on every sync, not
	// only when the user is created. Otherwise a user that was created while the
	// upstream account was not activated yet stays forbidden forever, and a user
	// that is disabled upstream is never disabled in Casdoor. Both callbacks that
	// react to the transition - triggerUserForbiddenWebhook() and
	// isUserAccessRevoked() (which terminates the user's sessions) - also key off
	// is_forbidden being one of the updated columns.
	switch GetSyncerProvider(syncer).(type) {
	case *WecomSyncerProvider:
		columns = append(columns, "wecom", "is_forbidden")
	case *DingtalkSyncerProvider:
		columns = append(columns, "dingtalk", "is_forbidden")
	case *LarkSyncerProvider:
		columns = append(columns, "lark", "is_forbidden")
	case *OktaSyncerProvider, *SCIMSyncerProvider, *AzureAdSyncerProvider,
		*GoogleWorkspaceSyncerProvider, *ActiveDirectorySyncerProvider,
		*AwsIamSyncerProvider, *PluginSyncerProvider:
		columns = append(columns, "is_forbidden")
	}

	// Database-backed syncers keep is_forbidden under the admin's control: it is
	// only written when the source column is mapped explicitly, because adding it
	// unconditionally would reset every user to "not forbidden".
	return columns
}

func (syncer *Syncer) updateUserForOriginalFields(user *User, key string) (bool, error) {
	var err error
	oldUser := User{}

	// key is the snake_cased CasdoorName of the syncer's key column, which an org admin
	// controls, and it is concatenated into the WHERE clause below, so reject any value
	// that is not a plain SQL identifier to keep it from injecting SQL into Casdoor's DB.
	if !util.FilterSQLIdentifier(key) {
		return false, fmt.Errorf("object.updateUserForOriginalFields() error: invalid key field: %s", key)
	}

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

	columns := syncer.getUpdateColumns(user)

	affected, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getUserValue(&oldUser, key), oldUser.Owner).Cols(columns...).Update(user)
	if err != nil {
		return false, err
	}

	if affected != 0 {
		triggerUserForbiddenWebhook(&oldUser, user, columns)

		// A user disabled in the upstream directory is forced offline like one forbidden by an admin
		if isUserAccessRevoked(&oldUser, user, columns) {
			err = TerminateUserAccess(&oldUser)
			if err != nil {
				return true, err
			}
		}
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

func (syncer *Syncer) initAdapter() error {
	provider := GetSyncerProvider(syncer)
	return provider.InitAdapter()
}

func RunSyncUsersJob() {
	syncers, err := GetSyncers("admin")
	if err != nil {
		fmt.Printf("RunSyncUsersJob() error: %s\n", err.Error())
		return
	}

	// A failing syncer must not stop the others; its error is already in its error text
	for _, syncer := range syncers {
		err = addSyncerJob(syncer)
		if err != nil {
			fmt.Printf("RunSyncUsersJob() failed to start syncer [%s]: %s\n", syncer.GetId(), err.Error())
		}
	}

	time.Sleep(time.Duration(1<<63 - 1))
}
