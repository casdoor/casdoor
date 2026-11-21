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

import "github.com/casdoor/casdoor/util"

// IsCodeSigninAllowedForUser checks the groups that an application limits the sign-in with a verification
// code or a magic link to, no group means everyone
func (application *Application) IsCodeSigninAllowedForUser(user *User) bool {
	if len(application.CodeSigninGroups) == 0 || user == nil {
		return true
	}

	for _, group := range user.Groups {
		if util.InSlice(application.CodeSigninGroups, group) || util.InSlice(application.CodeSigninGroups, util.GetId(user.Owner, group)) {
			return true
		}
	}
	return false
}
