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

func UploadUsers(owner string, fileId string) bool {
	table := xlsx.ReadXlsxFile(fileId)

	oldUserMap := getUserMap(owner)
	newUsers := []*User{}
	for _, line := range table {
		if parseLineItem(&line, 0) == "" {
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
			Avatar:            parseLineItem(&line, 9),
			PermanentAvatar:   "",
			Email:             parseLineItem(&line, 10),
			Phone:             parseLineItem(&line, 11),
			Location:          parseLineItem(&line, 12),
			Address:           []string{parseLineItem(&line, 13)},
			Affiliation:       parseLineItem(&line, 14),
			Title:             parseLineItem(&line, 15),
			IdCardType:        parseLineItem(&line, 16),
			IdCard:            parseLineItem(&line, 17),
			Homepage:          parseLineItem(&line, 18),
			Bio:               parseLineItem(&line, 19),
			Tag:               parseLineItem(&line, 20),
			Region:            parseLineItem(&line, 21),
			Language:          parseLineItem(&line, 22),
			Gender:            parseLineItem(&line, 23),
			Birthday:          parseLineItem(&line, 24),
			Education:         parseLineItem(&line, 25),
			Score:             parseLineItemInt(&line, 26),
			Ranking:           parseLineItemInt(&line, 27),
			IsDefaultAvatar:   false,
			IsOnline:          parseLineItemBool(&line, 28),
			IsAdmin:           parseLineItemBool(&line, 29),
			IsGlobalAdmin:     parseLineItemBool(&line, 30),
			IsForbidden:       parseLineItemBool(&line, 31),
			IsDeleted:         parseLineItemBool(&line, 32),
			SignupApplication: parseLineItem(&line, 33),
			Hash:              "",
			PreHash:           "",
			CreatedIp:         parseLineItem(&line, 34),
			LastSigninTime:    parseLineItem(&line, 35),
			LastSigninIp:      parseLineItem(&line, 36),
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
