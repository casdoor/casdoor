// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

import "github.com/casdoor/casdoor/i18n"

// ApplySignupApproval keeps a user who signs up through an application that needs approval forbidden until an
// administrator approves the user by turning off "Is forbidden"
func (application *Application) ApplySignupApproval(user *User) {
	if application.SignupApproval {
		user.IsForbidden = true
		user.PendingApproval = true
	}
}

func GetUserForbiddenMessage(user *User, lang string) string {
	if user.PendingApproval {
		return i18n.Translate(lang, "check:Your account is waiting for the administrator's approval")
	}
	return i18n.Translate(lang, "check:The user is forbidden to sign in, please contact the administrator")
}
