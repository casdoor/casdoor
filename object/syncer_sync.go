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

	"github.com/xorm-io/core"
)

func (syncer *Syncer) syncUsers() error {
	if len(syncer.TableColumns) == 0 {
		return fmt.Errorf("The syncer table columns should not be empty")
	}

	fmt.Printf("Running syncUsers()..\n")

	users, err := GetUsers(syncer.Organization)
	if err != nil {
		return err
	}

	oUsers, err := syncer.getOriginalUsers()
	if err != nil {
		return err
	}

	fmt.Printf("Users: %d, oUsers: %d\n", len(users), len(oUsers))

	var affiliationMap map[int]string
	if syncer.AffiliationTable != "" {
		_, affiliationMap, err = syncer.getAffiliationMap()
		if err != nil {
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

	if !syncer.IsReadOnly {
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
	} else if len(oUsers) != 0 {
		err = syncer.forbidRemovedUsers(users, myOUsers, key)
		if err != nil {
			return err
		}
	}

	return nil
}

// forbidRemovedUsers forbids the users created by this syncer that no longer exist upstream
func (syncer *Syncer) forbidRemovedUsers(users []*User, myOUsers map[string]*User, key string) error {
	registerSource := fmt.Sprintf("%s/%s", syncer.Organization, syncer.Name)
	for _, user := range users {
		if user.IsForbidden || user.IsAdmin || user.RegisterType != "Add Users" || user.RegisterSource != registerSource {
			continue
		}

		if _, ok := myOUsers[syncer.getUserValue(user, key)]; ok {
			continue
		}

		fmt.Printf("Forbid user removed from oUsers: %s\n", user.GetId())
		user.IsForbidden = true
		user.Hash = ""
		user.PreHash = ""
		_, err := ormer.Engine.ID(core.PK{user.Owner, user.Name}).Cols("is_forbidden", "hash", "pre_hash").Update(user)
		if err != nil {
			return err
		}

		TriggerWebhookForUser("forbid-user", user)

		err = terminateUserAccess(user)
		if err != nil {
			return err
		}
	}

	return nil
}
