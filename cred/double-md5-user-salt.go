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

package cred

type DoubleMd5UserSaltCredManager struct{}

func NewDoubleMd5UserSaltCredManager() *DoubleMd5UserSaltCredManager {
	cm := &DoubleMd5UserSaltCredManager{}
	return cm
}

func (cm *DoubleMd5UserSaltCredManager) GetHashedPassword(password string, userSalt string, organizationSalt string) string {
	res := getMd5HexDigest(password)
	if userSalt != "" {
		res = getMd5HexDigest(res + userSalt)
	}
	res = getMd5HexDigest(res)
	if userSalt != "" {
		res = getMd5HexDigest(res + userSalt)
	}
	return res
}

func (cm *DoubleMd5UserSaltCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, userSalt string, organizationSalt string) bool {
	return hashedPwd == cm.GetHashedPassword(plainPwd, userSalt, organizationSalt)
}
