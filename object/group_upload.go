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

func getGroupMap(owner string) (map[string]*Group, error) {
	m := map[string]*Group{}

	groups, err := GetGroups(owner)
	if err != nil {
		return m, err
	}

	for _, group := range groups {
		m[group.GetId()] = group
	}

	return m, nil
}

func UploadGroups(owner string, path string) (bool, error) {
	table := xlsx.ReadXlsxFile(path)

	oldGroupMap, err := getGroupMap(owner)
	if err != nil {
		return false, err
	}

	transGroups, err := StringArrayToStruct[Group](table)
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
