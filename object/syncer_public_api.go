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
	"errors"
	"fmt"
)

func getDbSyncerForUser(user *User) (*Syncer, error) {
	syncers, err := GetSyncers("admin")
	if err != nil {
		return nil, err
	}

	for _, syncer := range syncers {
		if syncer.Organization == user.Owner && syncer.IsEnabled && syncer.Type == "Database" {
			return syncer, nil
		}
	}
	return nil, nil
}

func getEnabledSyncerForOrganization(organization string) (*Syncer, error) {
	syncers, err := GetSyncers("admin")
	if err != nil {
		return nil, err
	}

	for _, syncer := range syncers {
		if syncer.Organization == organization && syncer.IsEnabled {
			syncer.initAdapter()
			return syncer, nil
		}
	}
	return nil, errors.New("no enabled syncer found")
}

func AddUserToOriginalDatabase(user *User) error {
	syncer, err := getEnabledSyncerForOrganization(user.Owner)
	if err != nil {
		return err
	}

	updatedOUser := syncer.createOriginalUserFromUser(user)
	_, err = syncer.addUser(updatedOUser)
	if err != nil {
		return err
	}

	fmt.Printf("Add from user to oUser: %v\n", updatedOUser)
	return nil
}

func UpdateUserToOriginalDatabase(user *User) error {
	syncer, err := getEnabledSyncerForOrganization(user.Owner)
	if err != nil {
		return err
	}

	newUser, err := GetUser(user.GetId())
	if err != nil {
		return err
	}

	updatedOUser := syncer.createOriginalUserFromUser(newUser)
	_, err = syncer.updateUser(updatedOUser)
	if err != nil {
		return err
	}

	fmt.Printf("Update from user to oUser: %v\n", updatedOUser)
	return nil
}
