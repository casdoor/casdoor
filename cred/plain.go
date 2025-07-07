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

type PlainCredManager struct{}

func NewPlainCredManager() *PlainCredManager {
	cm := &PlainCredManager{}
	return cm
}

func (cm *PlainCredManager) GetHashedPassword(password string, salt string) string {
	return password
}

func (cm *PlainCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, salt string) bool {
	return hashedPwd == plainPwd
}
