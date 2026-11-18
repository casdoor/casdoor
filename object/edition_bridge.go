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

// The helpers of the core that the features in ee/ use

func GetOrmer() *Ormer {
	return ormer
}

func GetUserEnforcer() *UserGroupEnforcer {
	return userEnforcer
}

func LookupUser(owner string, name string) (*User, error) {
	return getUser(owner, name)
}

func LookupOrganization(owner string, name string) (*Organization, error) {
	return getOrganization(owner, name)
}

func LookupApplication(owner string, name string) (*Application, error) {
	return getApplication(owner, name)
}

func LookupGroup(owner string, name string) (*Group, error) {
	return getGroup(owner, name)
}

func LookupRole(owner string, name string) (*Role, error) {
	return getRole(owner, name)
}

func UpdateUserColumns(id string, user *User, columns []string) (int64, error) {
	return updateUser(id, user, columns)
}

func InsertRecord(record *Record) (int64, error) {
	return addRecord(record)
}
