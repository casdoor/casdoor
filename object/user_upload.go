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

func getUserMap(owner string) map[string]*User {
	m := map[string]*User{}

	users := GetUsers(owner)
	for _, user := range users {
		m[user.GetId()] = user
	}

	return m
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

func UploadUsers(owner string, fileId string) bool {
	table := xlsx.ReadXlsxFile(fileId)

	oldUserMap := getUserMap(owner)
	newUsers := []*User{}
	for index, line := range table {
		if index == 0 || parseLineItem(&line, 0) == "" {
			continue
		}

		user := &User{
			Owner:             parseLineItem(&line, 0),
			Name:              parseLineItem(&line, 1),
			CreatedTime:       parseLineItem(&line, 2),
			UpdatedTime:       parseLineItem(&line, 3),
			Id:                parseLineItem(&line, 4),
			Type:              parseLineItem(&line, 5),
			Password:          parseLineItem(&line, 6),
			PasswordSalt:      parseLineItem(&line, 7),
			DisplayName:       parseLineItem(&line, 8),
			FirstName:         parseLineItem(&line, 9),
			LastName:          parseLineItem(&line, 10),
			Avatar:            parseLineItem(&line, 11),
			PermanentAvatar:   "",
			Email:             parseLineItem(&line, 12),
			Phone:             parseLineItem(&line, 13),
			Location:          parseLineItem(&line, 14),
			Address:           []string{parseLineItem(&line, 15)},
			Affiliation:       parseLineItem(&line, 16),
			Title:             parseLineItem(&line, 17),
			IdCardType:        parseLineItem(&line, 18),
			IdCard:            parseLineItem(&line, 19),
			Homepage:          parseLineItem(&line, 20),
			Bio:               parseLineItem(&line, 21),
			Tag:               parseLineItem(&line, 22),
			Region:            parseLineItem(&line, 23),
			Language:          parseLineItem(&line, 24),
			Gender:            parseLineItem(&line, 25),
			Birthday:          parseLineItem(&line, 26),
			Education:         parseLineItem(&line, 27),
			Score:             parseLineItemInt(&line, 28),
			Karma:             parseLineItemInt(&line, 29),
			Ranking:           parseLineItemInt(&line, 30),
			IsDefaultAvatar:   false,
			IsOnline:          parseLineItemBool(&line, 31),
			IsAdmin:           parseLineItemBool(&line, 32),
			IsGlobalAdmin:     parseLineItemBool(&line, 33),
			IsForbidden:       parseLineItemBool(&line, 34),
			IsDeleted:         parseLineItemBool(&line, 35),
			SignupApplication: parseLineItem(&line, 36),
			Hash:              "",
			PreHash:           "",
			CreatedIp:         parseLineItem(&line, 37),
			LastSigninTime:    parseLineItem(&line, 38),
			LastSigninIp:      parseLineItem(&line, 39),
			Ldap:              "",
			Properties:        map[string]string{},
		}

		if _, ok := oldUserMap[user.GetId()]; !ok {
			newUsers = append(newUsers, user)
		}
	}

	if len(newUsers) == 0 {
		return false
	}
	return AddUsersInBatch(newUsers)
}
