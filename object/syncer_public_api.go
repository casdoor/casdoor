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
	"time"
)

var (
	originalDatabaseWriteSlots       = make(chan struct{}, 16)
	originalDatabaseWriteRetryDelays = []time.Duration{0, 2 * time.Second, 10 * time.Second}
	originalDatabaseWriteTimeout     = 30 * time.Second
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
			return syncer, nil
		}
	}
	return nil, nil
}

func AddUserToOriginalDatabase(user *User) error {
	return writeUserToOriginalDatabase(user, true)
}

func UpdateUserToOriginalDatabase(user *User) error {
	return writeUserToOriginalDatabase(user, false)
}

func writeUserToOriginalDatabase(user *User, isAdd bool) error {
	syncer, err := getEnabledSyncerForOrganization(user.Owner)
	if err != nil {
		return err
	}
	if syncer == nil || !syncer.writesUsersBack() {
		return nil
	}

	select {
	case originalDatabaseWriteSlots <- struct{}{}:
	default:
		fmt.Printf("writeUserToOriginalDatabase() skipped user [%s] because too many writes are pending, the next sync of syncer [%s] will write it\n", user.GetId(), syncer.GetId())
		return nil
	}

	userCopy := *user
	go func() {
		defer func() { <-originalDatabaseWriteSlots }()

		var err error
		for i, delay := range originalDatabaseWriteRetryDelays {
			time.Sleep(delay)
			err = writeUserToOriginalDatabaseWithTimeout(*syncer, &userCopy, isAdd)
			if err == nil {
				return
			}
			fmt.Printf("writeUserToOriginalDatabase() attempt %d for user [%s] failed: %s\n", i+1, userCopy.GetId(), err.Error())
		}

		recordSyncerError(syncer, fmt.Errorf("failed to write user [%s] back to the original database: %s", userCopy.GetId(), err.Error()))
	}()

	return nil
}

func writeUserToOriginalDatabaseWithTimeout(syncer Syncer, user *User, isAdd bool) error {
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				syncer.Close()
				done <- fmt.Errorf("panic: %v", r)
			}
		}()

		err := syncer.writeUserBack(user, isAdd)
		closeErr := syncer.Close()
		if err == nil {
			err = closeErr
		}
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(originalDatabaseWriteTimeout):
		return fmt.Errorf("timed out after %s", originalDatabaseWriteTimeout)
	}
}

func (syncer *Syncer) writeUserBack(user *User, isAdd bool) error {
	err := syncer.initAdapter()
	if err != nil {
		return err
	}

	if isAdd {
		oUser := syncer.createOriginalUserFromUser(user)
		_, err = syncer.addUser(oUser)
		if err != nil {
			return err
		}

		fmt.Printf("Add from user to oUser: %v\n", oUser)
		return nil
	}

	newUser, err := GetUser(user.GetId())
	if err != nil {
		return err
	}
	if newUser == nil {
		return nil
	}

	oUser := syncer.createOriginalUserFromUser(newUser)
	_, err = syncer.updateUser(oUser)
	if err != nil {
		return err
	}

	fmt.Printf("Update from user to oUser: %v\n", oUser)
	return nil
}
