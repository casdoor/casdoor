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

package object

import (
	"sort"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/casdoor/casdoor/xlsx"
)

func getUserMap(owner string) (map[string]*User, error) {
	m := map[string]*User{}

	users, err := GetUsers(owner)
	if err != nil {
		return m, err
	}
	for _, user := range users {
		m[user.GetId()] = user
	}

	return m, nil
}

func parseLineItem(line *[]string, i int) string {
	if i >= len(*line) {
		return ""
	} else {
		return (*line)[i]
	}
}

func parseLineItemInt(line *[]string, i int) int {
	s := parseLineItem(line, i)
	return util.ParseInt(s)
}

func parseLineItemBool(line *[]string, i int) bool {
	return parseLineItemInt(line, i) != 0
}

func parseListItem(lines *[]string, i int) []string {
	if i >= len(*lines) {
		return nil
	}
	line := (*lines)[i]
	items := strings.Split(line, ";")
	trimmedItems := make([]string, 0, len(items))

	for _, item := range items {
		trimmedItem := strings.TrimSpace(item)
		if trimmedItem != "" {
			trimmedItems = append(trimmedItems, trimmedItem)
		}
	}

	sort.Strings(trimmedItems)

	return trimmedItems
}

func UploadUsers(owner string, path string) (bool, error) {
	table := xlsx.ReadXlsxFile(path)

	oldUserMap, err := getUserMap(owner)
	if err != nil {
		return false, err
	}

	transUsers, err := StringArrayToStruct[User](table)
	if err != nil {
		return false, err
	}
	newUsers := []*User{}
	for _, user := range transUsers {
		if _, ok := oldUserMap[user.GetId()]; !ok {
			newUsers = append(newUsers, user)
		}
	}

	if len(newUsers) == 0 {
		return false, nil
	}

	return AddUsersInBatch(newUsers)
}
