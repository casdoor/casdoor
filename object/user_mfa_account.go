// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

func AddMfaAccount(user *User, mfaAccount *MfaAccount) (bool, error) {
	user.MfaAccounts = append(user.MfaAccounts, *mfaAccount)
	affected, err := updateUser(user.GetId(), user, []string{"mfaAccounts"})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteMfaAccount(user *User, mfaAccount *MfaAccount) (bool, error) {
	for i, v := range user.MfaAccounts {
		if v.SecretKey == mfaAccount.SecretKey && v.AccountName == mfaAccount.AccountName {
			user.MfaAccounts = append(user.MfaAccounts[:i], user.MfaAccounts[i+1:]...)
			break
		}
	}
	affected, err := updateUser(user.GetId(), user, []string{"mfaAccounts"})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func UpdateMfaAccount(user *User, updatedAccount *MfaAccount) (bool, error) {
	for i, v := range user.MfaAccounts {
		if v.SecretKey == updatedAccount.SecretKey && v.Issuer == updatedAccount.Issuer {
			// Update the fields of the Mfa account
			user.MfaAccounts[i].AccountName = updatedAccount.AccountName

			// Update the user in the database
			affected, err := updateUser(user.GetId(), user, []string{"mfaAccounts"})
			if err != nil {
				return false, err
			}
			return affected != 0, err
		}
	}
	return false, nil
}
