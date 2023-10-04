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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

type OriginalUser = User
type OriginalGroup = Group

type Affiliation struct {
	Id   int    `xorm:"int notnull pk autoincr" json:"id"`
	Name string `xorm:"varchar(128)" json:"name"`
}

type Credential struct {
	Value string `json:"value"`
	Salt  string `json:"salt"`
}

func (syncer *Syncer) getUsers() []*User {
	users, err := GetUsers(syncer.Organization)
	if err != nil {
		panic(err)
	}

	return users
}

func (syncer *Syncer) getUserMap() ([]*User, map[string]*User, map[string]*User) {
	users := syncer.getUsers()

	m1 := map[string]*User{}
	m2 := map[string]*User{}
	for _, user := range users {
		m1[user.Id] = user
		m2[user.Name] = user
	}

	return users, m1, m2
}

func (syncer *Syncer) getCasdoorColumns() []string {
	res := []string{}
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.CasdoorName != "Id" {
			v := util.CamelToSnakeCase(tableColumn.CasdoorName)
			res = append(res, v)
		}
	}
	return res
}

func (syncer *Syncer) updateUserForOriginalByFields(user *User, key string) (bool, error) {
	var err error
	oldUser := User{}

	existed, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getUserValue(user, key), user.Owner).Get(&oldUser)
	if err != nil {
		return false, err
	}
	if !existed {
		return false, nil
	}

	if user.Avatar != oldUser.Avatar && user.Avatar != "" {
		user.PermanentAvatar, err = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar, true)
		if err != nil {
			return false, err
		}
	}

	columns := syncer.getCasdoorColumns()
	columns = append(columns, "affiliation", "hash", "pre_hash")
	affected, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getUserValue(&oldUser, key), oldUser.Owner).Cols(columns...).Update(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (syncer *Syncer) updateGroupForOriginalByFields(group *Group, key string) (bool, error) {
	var err error
	oldGroup := Group{}

	existed, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getGroupValue(group, key), group.Owner).Get(&oldGroup)
	if err != nil {
		return false, err
	}
	if !existed {
		return false, nil
	}

	affected, err := ormer.Engine.Where(key+" = ? and owner = ?", syncer.getGroupValue(&oldGroup, key), oldGroup.Owner).Update(group)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (syncer *Syncer) calculateHash(user *OriginalUser) string {
	values := []string{}
	m := syncer.getMapFromOriginalUser(user)
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.IsHashed {
			values = append(values, m[tableColumn.Name])
		}
	}

	s := strings.Join(values, "|")
	return util.GetMd5Hash(s)
}

func (syncer *Syncer) calculateGroupHash(group *OriginalGroup) string {
	values := []string{}
	m := syncer.getMapFromOriginalGroup(group)
	for _, value := range m {
		values = append(values, value)
	}

	s := strings.Join(values, "|")
	return util.GetMd5Hash(s)
}

func RunSyncUsersJob() {
	syncers, err := GetSyncers("admin")
	if err != nil {
		panic(err)
	}

	for _, syncer := range syncers {
		addSyncerJob(syncer)
	}

	time.Sleep(time.Duration(1<<63 - 1))
}

func (syncer *Syncer) getFullAvatarUrl(avatar string) string {
	if syncer.AvatarBaseUrl == "" {
		return avatar
	}

	if !strings.HasPrefix(avatar, "http") {
		return fmt.Sprintf("%s%s", syncer.AvatarBaseUrl, avatar)
	}
	return avatar
}

func (syncer *Syncer) getPartialAvatarUrl(avatar string) string {
	if strings.HasPrefix(avatar, syncer.AvatarBaseUrl) {
		return avatar[len(syncer.AvatarBaseUrl):]
	}
	return avatar
}

func (syncer *Syncer) createUserFromOriginalUser(originalUser *OriginalUser, affiliationMap map[int]string) *User {
	user := *originalUser
	user.Owner = syncer.Organization

	if user.Name == "" {
		user.Name = originalUser.Id
	}

	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	if user.Type == "" {
		user.Type = "normal-user"
	}

	user.Avatar = syncer.getFullAvatarUrl(user.Avatar)

	if affiliationMap != nil {
		if originalUser.Score != 0 {
			affiliation, ok := affiliationMap[originalUser.Score]
			if !ok {
				panic(fmt.Sprintf("Affiliation not found: %d", originalUser.Score))
			}
			user.Affiliation = affiliation
		}
	}

	if user.Properties == nil {
		user.Properties = map[string]string{}
	}

	return &user
}

func (syncer *Syncer) createGroupFromOriginalGroup(originalGroup *OriginalGroup, affiliationMap map[int]string) *Group {
	group := *originalGroup
	group.Owner = syncer.Organization

	if group.CreatedTime == "" {
		group.CreatedTime = util.GetCurrentTime()
	}

	if group.Type == "" {
		group.Type = "Virtual"
	}

	return &group
}

func (syncer *Syncer) createOriginalUserFromUser(user *User) *OriginalUser {
	originalUser := *user
	originalUser.Avatar = syncer.getPartialAvatarUrl(user.Avatar)
	return &originalUser
}

func (syncer *Syncer) getUserValue(user *User, key string) string {
	jsonData, _ := json.Marshal(user)
	var mapData map[string]interface{}
	if err := json.Unmarshal(jsonData, &mapData); err != nil {
		fmt.Println("conversion failed:", err)
		return user.Id
	}
	value := mapData[util.SnakeToCamel(key)]

	if str, ok := value.(string); ok {
		return str
	} else {
		if value != nil {
			valType := reflect.TypeOf(value)

			typeName := valType.Name()
			switch typeName {
			case "bool":
				return strconv.FormatBool(value.(bool))
			case "int":
				return strconv.Itoa(value.(int))
			}
		}
		return user.Id
	}
}

func (syncer *Syncer) getGroupValue(group *Group, key string) string {
	jsonData, _ := json.Marshal(group)
	var mapData map[string]interface{}
	if err := json.Unmarshal(jsonData, &mapData); err != nil {
		fmt.Println("conversion failed:", err)
		return group.Name
	}
	value := mapData[util.SnakeToCamel(key)]

	if str, ok := value.(string); ok {
		return str
	} else {
		if value != nil {
			valType := reflect.TypeOf(value)

			typeName := valType.Name()
			switch typeName {
			case "bool":
				return strconv.FormatBool(value.(bool))
			case "int":
				return strconv.Itoa(value.(int))
			}
		}
		return group.Name
	}
}

func (syncer *Syncer) getMapFromOriginalUser(user *OriginalUser) map[string]string {
	m := map[string]string{}
	m["Name"] = user.Name
	m["CreatedTime"] = user.CreatedTime
	m["UpdatedTime"] = user.UpdatedTime
	m["Id"] = user.Id
	m["Type"] = user.Type
	m["Password"] = user.Password
	m["PasswordSalt"] = user.PasswordSalt
	m["DisplayName"] = user.DisplayName
	m["Avatar"] = syncer.getFullAvatarUrl(user.Avatar)
	m["PermanentAvatar"] = user.PermanentAvatar
	m["Email"] = user.Email
	m["Phone"] = user.Phone
	m["Location"] = user.Location
	m["Address"] = strings.Join(user.Address, "|")
	m["Affiliation"] = user.Affiliation
	m["Title"] = user.Title
	m["IdCardType"] = user.IdCardType
	m["IdCard"] = user.IdCard
	m["Homepage"] = user.Homepage
	m["Bio"] = user.Bio
	m["Tag"] = user.Tag
	m["Region"] = user.Region
	m["Language"] = user.Language
	m["Gender"] = user.Gender
	m["Birthday"] = user.Birthday
	m["Education"] = user.Education
	m["Score"] = strconv.Itoa(user.Score)
	m["Ranking"] = strconv.Itoa(user.Ranking)
	m["IsDefaultAvatar"] = util.BoolToString(user.IsDefaultAvatar)
	m["IsOnline"] = util.BoolToString(user.IsOnline)
	m["IsAdmin"] = util.BoolToString(user.IsAdmin)
	m["IsForbidden"] = util.BoolToString(user.IsForbidden)
	m["IsDeleted"] = util.BoolToString(user.IsDeleted)
	m["CreatedIp"] = user.CreatedIp
	m["PreferredMfaType"] = user.PreferredMfaType
	m["TotpSecret"] = user.TotpSecret
	m["SignupApplication"] = user.SignupApplication

	m2 := map[string]string{}
	for _, tableColumn := range syncer.TableColumns {
		m2[tableColumn.Name] = m[tableColumn.CasdoorName]
	}

	return m2
}

func (syncer *Syncer) getMapFromOriginalGroup(group *OriginalGroup) map[string]string {
	m := map[string]string{}
	m["Name"] = group.Name
	m["DisplayName"] = group.DisplayName
	m["Manager"] = group.Manager
	m["ContactEmail"] = group.ContactEmail
	m["Type"] = group.Type
	m["ParentId"] = group.ParentId
	m["IsTopGroup"] = util.BoolToString(group.IsTopGroup)
	m["Title"] = group.Title
	m["Key"] = group.Key
	m["IsEnabled"] = util.BoolToString(group.IsEnabled)

	return m
}

func (syncer *Syncer) getSqlSetStringFromMap(m map[string]string) string {
	typeMap := syncer.getTableColumnsTypeMap()

	tokens := []string{}
	for k, v := range m {
		token := fmt.Sprintf("%s = %s", k, v)
		if typeMap[k] == "string" {
			token = fmt.Sprintf("%s = '%s'", k, v)
		}

		tokens = append(tokens, token)
	}
	return strings.Join(tokens, ", ")
}
