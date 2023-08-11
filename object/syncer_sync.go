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

func (syncer *Syncer) syncUsers() {
	if len(syncer.TableColumns) == 0 {
		return
	}

	fmt.Printf("Running syncUsers()..\n")

	users, _, _ := syncer.getUserMap()
	oUsers, oUserMap, err := syncer.getOriginalUserMap()
	if err != nil {
		fmt.Printf(err.Error())

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		line := fmt.Sprintf("[%s] %s\n", timestamp, err.Error())
		updateSyncerErrorText(syncer, line)
		return
	}

	fmt.Printf("Users: %d, oUsers: %d\n", len(users), len(oUsers))

	var affiliationMap map[int]string
	if syncer.AffiliationTable != "" {
		_, affiliationMap, err = syncer.getAffiliationMap()
	}

	newUsers := []*User{}

	key := "id"
	hasKey := false
	hasId := false
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.IsKey {
			hasKey = true
			key = tableColumn.Name
		}
		if tableColumn.Name == "id" {
			hasId = true
		}
	}

	if !hasKey && !hasId {
		key = syncer.TableColumns[0].Name
	}

	myUsers := map[string]*User{}
	for _, m := range users {
		myUsers[syncer.getUserValue(m, key)] = m
	}

	for _, oUser := range oUsers {
		primary := syncer.getUserValue(oUser, key)

		if _, ok := myUsers[primary]; !ok {
			newUser := syncer.createUserFromOriginalUser(oUser, affiliationMap)
			fmt.Printf("New user: %v\n", newUser)
			newUsers = append(newUsers, newUser)
		} else {
			user := myUsers[primary]
			oHash := syncer.calculateHash(oUser)
			if user.Hash == user.PreHash {
				if user.Hash != oHash {
					updatedUser := syncer.createUserFromOriginalUser(oUser, affiliationMap)
					updatedUser.Hash = oHash
					updatedUser.PreHash = oHash
					syncer.updateUserForOriginalByFields(updatedUser, key)
					fmt.Printf("Update from oUser to user: %v\n", updatedUser)
				}
			} else {
				if user.PreHash == oHash {
					if !syncer.IsReadOnly {
						updatedOUser := syncer.createOriginalUserFromUser(user)
						syncer.updateUser(updatedOUser)
						fmt.Printf("Update from user to oUser: %v\n", updatedOUser)
					}

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
						syncer.updateUserForOriginalByFields(updatedUser, key)
						fmt.Printf("Update from oUser to user (2nd condition): %v\n", updatedUser)
					}
				}
			}
		}
	}
	_, err = AddUsersInBatch(newUsers)
	if err != nil {
		panic(err)
	}

	if !syncer.IsReadOnly {
		for _, user := range users {
			id := user.Id
			if _, ok := oUserMap[id]; !ok {
				newOUser := syncer.createOriginalUserFromUser(user)
				_, err = syncer.addUser(newOUser)
				if err != nil {
					panic(err)
				}
				fmt.Printf("New oUser: %v\n", newOUser)
			}
		}
	}
}
