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
)

func (syncer *Syncer) syncUsers() {
	fmt.Printf("Running syncUsers()..\n")

	users, userMap := syncer.getUserMap()
	oUsers, oUserMap := syncer.getUserMapOriginal()
	fmt.Printf("Users: %d, oUsers: %d\n", len(users), len(oUsers))

	_, affiliationMap := syncer.getAffiliationMap()

	newUsers := []*User{}
	for _, oUser := range oUsers {
		id := strconv.Itoa(oUser.Id)
		if _, ok := userMap[id]; !ok {
			newUser := syncer.createUserFromOriginalUser(oUser, affiliationMap)
			fmt.Printf("New user: %v\n", newUser)
			newUsers = append(newUsers, newUser)
		} else {
			user := userMap[id]
			oHash := syncer.calculateHash(oUser)

			if user.Hash == user.PreHash {
				if user.Hash != oHash {
					updatedUser := syncer.createUserFromOriginalUser(oUser, affiliationMap)
					updatedUser.Hash = oHash
					updatedUser.PreHash = oHash
					UpdateUserForOriginalFields(updatedUser)
					fmt.Printf("Update from oUser to user: %v\n", updatedUser)
				}
			} else {
				if user.PreHash == oHash {
					updatedOUser := syncer.createOriginalUserFromUser(user)
					syncer.updateUser(updatedOUser)
					fmt.Printf("Update from user to oUser: %v\n", updatedOUser)

					// update preHash
					user.PreHash = user.Hash
					SetUserField(user, "pre_hash", user.PreHash)
				} else {
					if user.Hash == oHash {
						// update preHash
						user.PreHash = user.Hash
						SetUserField(user, "pre_hash", user.PreHash)
					} else {
						updatedUser := syncer.createUserFromOriginalUser(oUser, affiliationMap)
						updatedUser.Hash = oHash
						updatedUser.PreHash = oHash
						UpdateUserForOriginalFields(updatedUser)
						fmt.Printf("Update from oUser to user (2nd condition): %v\n", updatedUser)
					}
				}
			}
		}
	}
	AddUsersInBatch(newUsers)

	for _, user := range users {
		id := user.Id
		if _, ok := oUserMap[id]; !ok {
			panic(fmt.Sprintf("New original user: cannot create now, user = %v", user))
		}
	}
}
