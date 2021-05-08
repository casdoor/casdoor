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

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func getFullAvatarUrl(avatar string) string {
	return fmt.Sprintf("%s%s", avatarBaseUrl, avatar)
}

func createUserFromOriginalUser(originalUser *User) *object.User {
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
		PhonePrefix:   "86",
		Phone:         originalUser.Cellphone,
		Affiliation:   "",
		IsAdmin:       false,
		IsGlobalAdmin: false,
		IsForbidden:   originalUser.Deleted != 0,
	}
	return user
}

func syncUsers() {
	fmt.Printf("Running syncUsers()..\n")

	users, userMap := getUserMap()
	oUsers, _ := getUserMapOriginal()
	fmt.Printf("Users: %d, oUsers: %d\n", len(users), len(oUsers))

	newUsers := []*object.User{}
	for _, oUser := range oUsers {
		id := strconv.Itoa(oUser.Id)
		if _, ok := userMap[id]; !ok {
			user := createUserFromOriginalUser(oUser)
			fmt.Printf("New user: %v\n", user)
			newUsers = append(newUsers, user)
		} else {
			user := userMap[id]
			hash := calculateHash(oUser)
			if user.Hash != hash {
				user := createUserFromOriginalUser(oUser)
				object.UpdateUser(user.GetId(), user)
				fmt.Printf("Update user: %v\n", user)
			}
		}
	}
	object.AddUsersSafe(newUsers)
}
