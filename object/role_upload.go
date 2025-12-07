// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import (
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/xlsx"
)

func getRoleMap(owner string) (map[string]*Role, error) {
	m := map[string]*Role{}

	roles, err := GetRoles(owner)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		m[role.GetId()] = role
	}

	return m, nil
}

func UploadRoles(owner string, path string) (bool, error) {
	table := xlsx.ReadXlsxFile(path)

	if len(table) == 0 {
		return false, fmt.Errorf("empty table")
	}

	for idx, row := range table[0] {
		splitRow := strings.Split(row, "#")
		if len(splitRow) > 1 {
			table[0][idx] = splitRow[1]
		}
	}

	uploadedRoles, err := StringArrayToStruct[Role](table)
	if err != nil {
		return false, err
	}

	oldRoleMap, err := getRoleMap(owner)
	if err != nil {
		return false, err
	}

	newRoles := []*Role{}
	for _, role := range uploadedRoles {
		if _, ok := oldRoleMap[role.GetId()]; !ok {
			newRoles = append(newRoles, role)
		}
	}

	if len(newRoles) == 0 {
		return false, nil
	}

	return AddRolesInBatch(newRoles), nil
}
