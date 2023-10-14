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

	users, err := GetUsers(syncer.Organization)
	if err != nil {
		line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
		_, err2 := updateSyncerErrorText(syncer, line)
		if err2 != nil {
			panic(err2)
		}

		return err
	}

	oUsers, err := syncer.getOriginalUsers()
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
	}

	key := syncer.getKey()

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

	_, err = AddUsersInBatch(newUsers)
	if err != nil {
		return err
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
	}

	return nil
}

func (syncer *Syncer) syncUsersNoError() {
	err := syncer.syncUsers()
	if err != nil {
		fmt.Printf("syncUsersNoError() error: %s\n", err.Error())
	}
}

func (syncer *Syncer) syncGroups() error {
	if syncer.Type != "WeCom" {
		return nil
	}

	fmt.Printf("Running syncUsers()..\n")

	groups, err := GetGroups(syncer.Owner)
	if err != nil {
		line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
		_, err2 := updateSyncerErrorText(syncer, line)
		if err2 != nil {
			panic(err2)
		}

		return err
	}
	oGroups, _, err := syncer.getOriginalGroupMap()
	if err != nil {
		line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
		_, err2 := updateSyncerErrorText(syncer, line)
		if err2 != nil {
			panic(err2)
		}

		return err
	}

	fmt.Printf("Groups: %d, oGroups: %d\n", len(groups), len(oGroups))

	var affiliationMap map[int]string
	if syncer.AffiliationTable != "" {
		_, affiliationMap, err = syncer.getAffiliationMap()
	}

	key := "name"

	myGroups := map[string]*Group{}
	for _, m := range groups {
		myGroups[syncer.getGroupValue(m, key)] = m
	}

	myOGroups := map[string]*Group{}
	for _, m := range oGroups {
		myOGroups[syncer.getGroupValue(m, key)] = m
	}

	newGroups := []*Group{}
	for _, oGroup := range oGroups {
		primary := syncer.getGroupValue(oGroup, key)

		if _, ok := myGroups[primary]; !ok {
			fmt.Printf("New user: %v\n", oGroup)
			newGroups = append(newGroups, oGroup)
		} else {
			group := myGroups[primary]
			oHash := syncer.calculateGroupHash(oGroup)
			if group.Hash == group.PreHash {
				if group.Hash != oHash {
					updatedGroup := syncer.createGroupFromOriginalGroup(oGroup, affiliationMap)
					updatedGroup.Hash = oHash
					updatedGroup.PreHash = oHash

					fmt.Printf("Update from oGroup to group: %v\n", updatedGroup)
					_, err = syncer.updateGroupForOriginalByFields(updatedGroup, key)
					if err != nil {
						return err
					}
				}
			} else {
				if group.PreHash == oHash {
					if !syncer.IsReadOnly {
						updatedOGroup := group

						fmt.Printf("Update from group to oGroup: %v\n", updatedOGroup)
						_, err = syncer.updateGroup(updatedOGroup)
						if err != nil {
							return err
						}
					}

					// update preHash
					group.PreHash = group.Hash
					_, err = SetGroupField(group, "pre_hash", group.PreHash)
					if err != nil {
						return err
					}
				} else {
					if group.Hash == oHash {
						// update preHash
						group.PreHash = group.Hash
						_, err = SetGroupField(group, "pre_hash", group.PreHash)
						if err != nil {
							return err
						}
					} else {
						updatedGroup := syncer.createGroupFromOriginalGroup(oGroup, affiliationMap)
						updatedGroup.Hash = oHash
						updatedGroup.PreHash = oHash

						fmt.Printf("Update from oGroup to group (2nd condition): %v\n", updatedGroup)
						_, err = syncer.updateGroupForOriginalByFields(updatedGroup, key)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	_, err = AddGroupsInBatch(newGroups)
	if err != nil {
		return err
	}

	if !syncer.IsReadOnly {
		for _, group := range groups {
			primary := syncer.getGroupValue(group, key)
			if _, ok := myOGroups[primary]; !ok {
				newOGroup := group

				fmt.Printf("New oGroup: %v\n", newOGroup)
				_, err = syncer.addGroup(newOGroup)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
