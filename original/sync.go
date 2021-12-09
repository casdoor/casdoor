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

package original

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

func getFullAvatarUrl(avatar string) string {
	if !strings.HasPrefix(avatar, "https://") {
		return fmt.Sprintf("%s%s", avatarBaseUrl, avatar)
	}
	return avatar
}

func getPartialAvatarUrl(avatar string) string {
	if strings.HasPrefix(avatar, avatarBaseUrl) {
		return avatar[len(avatarBaseUrl):]
	}
	return avatar
}

func createUserFromOriginalUser(originalUser *User, affiliationMap map[int]string) *object.User {
	affiliation := ""
	if originalUser.SchoolId != 0 {
		var ok bool
		affiliation, ok = affiliationMap[originalUser.SchoolId]
		if !ok {
			panic(fmt.Sprintf("SchoolId not found: %d", originalUser.SchoolId))
		}
	}

	user := &object.User{
		Owner:         orgName,
		Name:          strconv.Itoa(originalUser.Id),
		CreatedTime:   util.GetCurrentTime(),
		Id:            strconv.Itoa(originalUser.Id),
		Type:          "normal-user",
		Password:      originalUser.Password,
		DisplayName:   originalUser.Name,
		Avatar:        getFullAvatarUrl(originalUser.Avatar),
		Email:         "",
		Phone:         originalUser.Cellphone,
		Address:       []string{},
		Affiliation:   affiliation,
		Score:         originalUser.SchoolId,
		IsAdmin:       false,
		IsGlobalAdmin: false,
		IsForbidden:   originalUser.Deleted != 0,
		IsDeleted:     false,
		Properties:    map[string]string{},
	}
	return user
}

func createOriginalUserFromUser(user *object.User) *User {
	deleted := 0
	if user.IsForbidden {
		deleted = 1
	}

	originalUser := &User{
		Id:        util.ParseInt(user.Id),
		Name:      user.DisplayName,
		Password:  user.Password,
		Cellphone: user.Phone,
		SchoolId:  user.Score,
		Avatar:    getPartialAvatarUrl(user.Avatar),
		Deleted:   deleted,
	}
	return originalUser
}

func syncUsers() {
	fmt.Printf("Running syncUsers()..\n")

	users, userMap := getUserMap()
	oUsers, oUserMap := getUserMapOriginal()
	fmt.Printf("Users: %d, oUsers: %d\n", len(users), len(oUsers))

	_, affiliationMap := getAffiliationMap()

	newUsers := []*object.User{}
	for _, oUser := range oUsers {
		id := strconv.Itoa(oUser.Id)
		if _, ok := userMap[id]; !ok {
			newUser := createUserFromOriginalUser(oUser, affiliationMap)
			fmt.Printf("New user: %v\n", newUser)
			newUsers = append(newUsers, newUser)
		} else {
			user := userMap[id]
			oHash := calculateHash(oUser)

			if user.Hash == user.PreHash {
				if user.Hash != oHash {
					updatedUser := createUserFromOriginalUser(oUser, affiliationMap)
					updatedUser.Hash = oHash
					updatedUser.PreHash = oHash
					object.UpdateUserForOriginalFields(updatedUser)
					fmt.Printf("Update from oUser to user: %v\n", updatedUser)
				}
			} else {
				if user.PreHash == oHash {
					updatedOUser := createOriginalUserFromUser(user)
					updateUser(updatedOUser)
					fmt.Printf("Update from user to oUser: %v\n", updatedOUser)

					// update preHash
					user.PreHash = user.Hash
					object.SetUserField(user, "pre_hash", user.PreHash)
				} else {
					if user.Hash == oHash {
						// update preHash
						user.PreHash = user.Hash
						object.SetUserField(user, "pre_hash", user.PreHash)
					} else {
						updatedUser := createUserFromOriginalUser(oUser, affiliationMap)
						updatedUser.Hash = oHash
						updatedUser.PreHash = oHash
						object.UpdateUserForOriginalFields(updatedUser)
						fmt.Printf("Update from oUser to user (2nd condition): %v\n", updatedUser)
					}
				}
			}
		}
	}
	object.AddUsersInBatch(newUsers)

	for _, user := range users {
		id := user.Id
		if _, ok := oUserMap[id]; !ok {
			panic(fmt.Sprintf("New original user: cannot create now, user = %v", user))
		}
	}
}
