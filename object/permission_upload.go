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
	"github.com/casdoor/casdoor/xlsx"
)

func getPermissionMap(owner string) map[string]*Permission {
	m := map[string]*Permission{}

	permissions := GetPermissions(owner)
	for _, permission := range permissions {
		m[permission.GetId()] = permission
	}

	return m
}

func UploadPermissions(owner string, fileId string) bool {
	table := xlsx.ReadXlsxFile(fileId)

	oldUserMap := getPermissionMap(owner)
	newPermissions := []*Permission{}
	for index, line := range table {
		if index == 0 || parseLineItem(&line, 0) == "" {
			continue
		}

		permission := &Permission{
			Owner:       parseLineItem(&line, 0),
			Name:        parseLineItem(&line, 1),
			CreatedTime: parseLineItem(&line, 2),
			DisplayName: parseLineItem(&line, 3),

			Users:   parseListItem(&line, 4),
			Roles:   parseListItem(&line, 5),
			Domains: parseListItem(&line, 6),

			Model:        parseLineItem(&line, 7),
			Adapter:      parseLineItem(&line, 8),
			ResourceType: parseLineItem(&line, 9),

			Resources: parseListItem(&line, 10),
			Actions:   parseListItem(&line, 11),

			Effect:    parseLineItem(&line, 12),
			IsEnabled: parseLineItemBool(&line, 13),

			Submitter:   parseLineItem(&line, 14),
			Approver:    parseLineItem(&line, 15),
			ApproveTime: parseLineItem(&line, 16),
			State:       parseLineItem(&line, 17),
		}

		if _, ok := oldUserMap[permission.GetId()]; !ok {
			newPermissions = append(newPermissions, permission)
		}
	}

	if len(newPermissions) == 0 {
		return false
	}
	return AddPermissionsInBatch(newPermissions)
}
