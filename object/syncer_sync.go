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

	"github.com/casdoor/casdoor/util"
)

func (syncer *Syncer) syncUsers() error {
	if len(syncer.TableColumns) == 0 {
		return fmt.Errorf("The syncer table columns should not be empty")
	}

	fmt.Printf("Running syncUsers()..\n")

	// Determine if incremental sync is possible:
	// - LastSyncTime must be set (from a previous successful sync)
	// - External table must have an UpdatedTime column mapped
	// When both conditions are met, only fetch users modified since last sync
	useIncrementalSync := syncer.LastSyncTime != "" && syncer.getUpdatedTimeColumn() != ""
	if useIncrementalSync {
		fmt.Printf("Using incremental sync (last sync: %s)\n", syncer.LastSyncTime)
	} else {
		fmt.Printf("Using full sync\n")
	}

	users, err := GetUsers(syncer.Organization)
	if err != nil {
		line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
		_, err2 := updateSyncerErrorText(syncer, line)
		if err2 != nil {
			panic(err2)
		}

		return err
	}

	var oUsers []*OriginalUser
	if useIncrementalSync {
		oUsers, err = syncer.getOriginalUsersWithFilter(syncer.LastSyncTime)
	} else {
		oUsers, err = syncer.getOriginalUsers()
	}
	if err != nil {
		line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
		_, err2 := updateSyncerErrorText(syncer, line)
		if err2 != nil {
			panic(err2)
		}

		return err
	}

	fmt.Printf("Users: %d, oUsers: %d\n", len(users), len(oUsers))

	var affiliationMap map[int]string
	if syncer.AffiliationTable != "" {
		_, affiliationMap, err = syncer.getAffiliationMap()
		if err != nil {
			line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
			_, err2 := updateSyncerErrorText(syncer, line)
			if err2 != nil {
				panic(err2)
			}

			return err
		}
	}

	key := syncer.getLocalPrimaryKey()

	myUsers := map[string]*User{}
	for _, m := range users {
		myUsers[syncer.getUserValue(m, key)] = m
	}

	myOUsers := map[string]*User{}
	for _, m := range oUsers {
		myOUsers[syncer.getUserValue(m, key)] = m
	}

	newUsers := []*User{}
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

					fmt.Printf("Update from oUser to user: %v\n", updatedUser)
					_, err = syncer.updateUserForOriginalFields(updatedUser, key)
					if err != nil {
						return err
					}
				}
			} else {
				if user.PreHash == oHash {
					if !syncer.IsReadOnly {
						updatedOUser := syncer.createOriginalUserFromUser(user)

						fmt.Printf("Update from user to oUser: %v\n", updatedOUser)
						_, err = syncer.updateUser(updatedOUser)
						if err != nil {
							return err
						}
					}

					// update preHash
					user.PreHash = user.Hash
					_, err = SetUserField(user, "pre_hash", user.PreHash)
					if err != nil {
						return err
					}
				} else {
					if user.Hash == oHash {
						// update preHash
						user.PreHash = user.Hash
						_, err = SetUserField(user, "pre_hash", user.PreHash)
						if err != nil {
							return err
						}
					} else {
						updatedUser := syncer.createUserFromOriginalUser(oUser, affiliationMap)
						updatedUser.Hash = oHash
						updatedUser.PreHash = oHash

						fmt.Printf("Update from oUser to user (2nd condition): %v\n", updatedUser)
						_, err = syncer.updateUserForOriginalFields(updatedUser, key)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	if len(newUsers) != 0 {
		_, err = AddUsersInBatch(newUsers)
		if err != nil {
			return err
		}

		// Trigger webhooks for syncer user additions
		for _, newUser := range newUsers {
			TriggerWebhookForUser("new-user-syncer", newUser)
		}
	}

	// Only sync new local users to external database during full sync
	// In incremental sync, myOUsers doesn't contain all external users, so we can't determine if a user is truly new
	if !syncer.IsReadOnly && !useIncrementalSync {
		for _, user := range users {
			primary := syncer.getUserValue(user, key)
			if _, ok := myOUsers[primary]; !ok {
				newOUser := syncer.createOriginalUserFromUser(user)

				fmt.Printf("New oUser: %v\n", newOUser)
				_, err = syncer.addUser(newOUser)
				if err != nil {
					return err
				}
			}
		}
	}

	// Update LastSyncTime after successful sync
	err = updateSyncerLastSyncTime(syncer)
	if err != nil {
		fmt.Printf("Warning: Failed to update LastSyncTime: %s\n", err.Error())
		// Don't fail the sync if we can't update the timestamp
	}

	return nil
}

func (syncer *Syncer) syncUsersNoError() {
	err := syncer.syncUsers()
	if err != nil {
		fmt.Printf("syncUsersNoError() error: %s\n", err.Error())
	}
}
