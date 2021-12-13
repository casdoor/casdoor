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

	"github.com/casbin/casdoor/object"
)

func isEnabled() bool {
	if adapter == nil {
		InitAdapter()
		if adapter == nil {
			return false
		}
	}
	return true
}

func AddUserToOriginalDatabase(user *object.User) {
	if user.Owner != orgName {
		return
	}

	if !isEnabled() {
		return
	}

	updatedOUser := createOriginalUserFromUser(user)
	addUser(updatedOUser)
	fmt.Printf("Add from user to oUser: %v\n", updatedOUser)
}

func UpdateUserToOriginalDatabase(user *object.User) {
	if user.Owner != orgName {
		return
	}

	if !isEnabled() {
		return
	}

	newUser := object.GetUser(user.GetId())

	updatedOUser := createOriginalUserFromUser(newUser)
	updateUser(updatedOUser)
	fmt.Printf("Update from user to oUser: %v\n", updatedOUser)
}
