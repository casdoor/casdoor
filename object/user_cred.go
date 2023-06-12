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

import "github.com/casdoor/casdoor/cred"

func calculateHash(user *User) (string, error) {
	syncer, err := getDbSyncerForUser(user)
	if err != nil {
		return "", err
	}

	if syncer == nil {
		return "", nil
	}

	return syncer.calculateHash(user), nil
}

func (user *User) UpdateUserHash() error {
	hash, err := calculateHash(user)
	if err != nil {
		return err
	}

	user.Hash = hash
	return nil
}

func (user *User) UpdateUserPassword(organization *Organization) {
	credManager := cred.GetCredManager(organization.PasswordType)
	if credManager != nil {
		hashedPassword := credManager.GetHashedPassword(user.Password, user.PasswordSalt, organization.PasswordSalt)
		user.Password = hashedPassword
		user.PasswordType = organization.PasswordType
	}
}
