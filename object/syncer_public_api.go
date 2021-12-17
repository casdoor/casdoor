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

import "fmt"

func getEnabledSyncerForOrganization(organization string) *Syncer {
	syncers := GetSyncers("admin")
	for _, syncer := range syncers {
		if syncer.Organization == organization && syncer.IsEnabled {
			return syncer
		}
	}
	return nil
}

func AddUserToOriginalDatabase(user *User) {
	syncer := getEnabledSyncerForOrganization(user.Owner)
	if syncer == nil {
		return
	}

	updatedOUser := syncer.createOriginalUserFromUser(user)
	syncer.addUser(updatedOUser)
	fmt.Printf("Add from user to oUser: %v\n", updatedOUser)
}

func UpdateUserToOriginalDatabase(user *User) {
	syncer := getEnabledSyncerForOrganization(user.Owner)
	if syncer == nil {
		return
	}

	newUser := GetUser(user.GetId())

	updatedOUser := syncer.createOriginalUserFromUser(newUser)
	syncer.updateUser(updatedOUser)
	fmt.Printf("Update from user to oUser: %v\n", updatedOUser)
}
