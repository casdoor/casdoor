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
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/casdoor/casdoor/i18n"
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

func isEmptyLine(line []string) bool {
	for _, cell := range line {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func UploadUsers(owner string, path string, userObj *User, lang string) (bool, error) {
	table, err := xlsx.ReadXlsxFile(path)
	if err != nil {
		return false, err
	}

	if len(table) == 0 {
		return false, errors.New(i18n.Translate(lang, "general:The uploaded file is empty"))
	}

	for idx, row := range table[0] {
		splitRow := strings.Split(row, "#")
		if len(splitRow) > 1 {
			table[0][idx] = splitRow[1]
		}
	}

	parsedUsers, err := StringArrayToStruct[User](table)
	if err != nil {
		return false, err
	}

	// Excel keeps blank rows in the sheet, they would otherwise be imported as users with an empty name
	uploadedUsers := []*User{}
	lines := []int{}
	for idx, user := range parsedUsers {
		if isEmptyLine(table[idx+1]) {
			continue
		}

		uploadedUsers = append(uploadedUsers, user)
		lines = append(lines, idx+2)
	}
	if len(uploadedUsers) == 0 {
		return false, errors.New(i18n.Translate(lang, "general:The uploaded file contains no user"))
	}

	organizationName := uploadedUsers[0].Owner
	if organizationName == "" || !userObj.IsGlobalAdmin() {
		organizationName = owner
	}

	organization, err := getOrganization("admin", organizationName)
	if err != nil {
		return false, err
	}
	if organization == nil {
		return false, fmt.Errorf(i18n.Translate(lang, "auth:The organization: %s does not exist"), organizationName)
	}

	oldUserMap, err := getUserMap(organizationName)
	if err != nil {
		return false, err
	}

	newUsers := []*User{}
	existingNames := []string{}
	uploadedLineMap := map[string]int{}
	for idx, user := range uploadedUsers {
		line := lines[idx]

		user.Owner = organizationName
		user.Name = strings.TrimSpace(user.Name)
		if user.Name == "" {
			return false, fmt.Errorf(i18n.Translate(lang, "general:The user name in line %d is empty"), line)
		}
		if oldLine, ok := uploadedLineMap[user.Name]; ok {
			return false, fmt.Errorf(i18n.Translate(lang, "general:The user: %s in line %d is duplicated with line %d"), user.Name, line, oldLine)
		}
		uploadedLineMap[user.Name] = line

		if _, ok := oldUserMap[user.GetId()]; ok {
			existingNames = append(existingNames, user.Name)
			continue
		}

		if user.CreatedTime == "" {
			user.CreatedTime = util.GetCurrentTime()
		}
		if user.Id == "" {
			user.Id = util.GenerateId()
		}
		if user.Type == "" {
			user.Type = "normal-user"
		}
		user.PasswordType = "plain"
		if user.DisplayName == "" {
			user.DisplayName = user.Name
		}
		user.Avatar = organization.DefaultAvatar
		if user.Region == "" {
			user.Region = userObj.Region
		}
		if user.Address == nil {
			user.Address = []string{}
		}
		if user.CountryCode == "" {
			user.CountryCode = userObj.CountryCode
		}
		if user.SignupApplication == "" {
			user.SignupApplication = organization.DefaultApplication
		}
		if user.RegisterType == "" {
			user.RegisterType = "Upload Users"
		}
		if user.RegisterSource == "" {
			user.RegisterSource = userObj.GetId()
		}

		newUsers = append(newUsers, user)
	}

	if len(newUsers) == 0 {
		if len(existingNames) > 10 {
			existingNames = append(existingNames[:10], "...")
		}
		return false, fmt.Errorf(i18n.Translate(lang, "general:The users already exist: %s"), strings.Join(existingNames, ", "))
	}

	return AddUsersInBatch(newUsers)
}
