// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

func getGroupMap(groups []*Group) (map[string]*Group, error) {
	m := map[string]*Group{}

	owners := map[string]bool{}
	for _, group := range groups {
		if owners[group.Owner] {
			continue
		}
		owners[group.Owner] = true

		oldGroups, err := GetGroups(group.Owner)
		if err != nil {
			return nil, err
		}

		for _, oldGroup := range oldGroups {
			m[oldGroup.GetId()] = oldGroup
		}
	}

	return m, nil
}

func UploadGroups(path string) (bool, error) {
	table, err := xlsx.ReadXlsxFile(path)
	if err != nil {
		return false, err
	}

	transGroups, err := StringArrayToStruct[Group](table)
	if err != nil {
		return false, err
	}

	oldGroupMap, err := getGroupMap(transGroups)
	if err != nil {
		return false, err
	}

	newGroups := []*Group{}
	for _, group := range transGroups {
		if _, ok := oldGroupMap[group.GetId()]; !ok {
			newGroups = append(newGroups, group)
		}
	}

	if len(newGroups) == 0 {
		return false, nil
	}

	return AddGroupsInBatch(newGroups)
}
