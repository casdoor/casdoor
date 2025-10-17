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
	"github.com/casdoor/casdoor/v2/xlsx"
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

	oldUserMap, err := getRoleMap(owner)
	if err != nil {
		return false, err
	}

	newRoles := []*Role{}
	for index, line := range table {
		line := line
		if index == 0 || parseLineItem(&line, 0) == "" {
			continue
		}

		role := &Role{
			Owner:       parseLineItem(&line, 0),
			Name:        parseLineItem(&line, 1),
			CreatedTime: parseLineItem(&line, 2),
			DisplayName: parseLineItem(&line, 3),

			Users:     parseListItem(&line, 4),
			Roles:     parseListItem(&line, 5),
			Domains:   parseListItem(&line, 6),
			IsEnabled: parseLineItemBool(&line, 7),
		}

		if _, ok := oldUserMap[role.GetId()]; !ok {
			newRoles = append(newRoles, role)
		}
	}

	if len(newRoles) == 0 {
		return false, nil
	}

	return AddRolesInBatch(newRoles), nil
}
