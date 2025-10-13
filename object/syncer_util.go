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
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

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

	if user.RegisterType == "" {
		user.RegisterType = "Add Users"
		user.RegisterSource = fmt.Sprintf("%s/%s", syncer.Organization, syncer.Name)
	}

	return &user
}

func (syncer *Syncer) createOriginalUserFromUser(user *User) *OriginalUser {
	originalUser := *user
	originalUser.Avatar = syncer.getPartialAvatarUrl(user.Avatar)
	return &originalUser
}

func (syncer *Syncer) setUserByKeyValue(user *User, key string, value string) {
	switch key {
	case "Name":
		user.Name = value
	case "CreatedTime":
		user.CreatedTime = value
	case "UpdatedTime":
		user.UpdatedTime = value
	case "DeletedTime":
		user.DeletedTime = value
	case "Id":
		user.Id = value
	case "Type":
		user.Type = value
	case "Password":
		user.Password = value
	case "PasswordSalt":
		user.PasswordSalt = value
	case "DisplayName":
		user.DisplayName = value
	case "FirstName":
		user.FirstName = value
	case "LastName":
		user.LastName = value
	case "Avatar":
		user.Avatar = syncer.getPartialAvatarUrl(value)
	case "PermanentAvatar":
		user.PermanentAvatar = value
	case "Email":
		user.Email = value
	case "EmailVerified":
		user.EmailVerified = util.ParseBool(value)
	case "Phone":
		user.Phone = value
	case "Location":
		user.Location = value
	case "Address":
		user.Address = []string{value}
	case "Affiliation":
		user.Affiliation = value
	case "Title":
		user.Title = value
	case "IdCardType":
		user.IdCardType = value
	case "IdCard":
		user.IdCard = value
	case "Homepage":
		user.Homepage = value
	case "Bio":
		user.Bio = value
	case "Tag":
		user.Tag = value
	case "Region":
		user.Region = value
	case "Language":
		user.Language = value
	case "Gender":
		user.Gender = value
	case "Birthday":
		user.Birthday = value
	case "Education":
		user.Education = value
	case "Score":
		user.Score = util.ParseInt(value)
	case "Ranking":
		user.Ranking = util.ParseInt(value)
	case "IsDefaultAvatar":
		user.IsDefaultAvatar = util.ParseBool(value)
	case "IsOnline":
		user.IsOnline = util.ParseBool(value)
	case "IsAdmin":
		user.IsAdmin = util.ParseBool(value)
	case "IsForbidden":
		user.IsForbidden = util.ParseBool(value)
	case "IsDeleted":
		user.IsDeleted = util.ParseBool(value)
	case "CreatedIp":
		user.CreatedIp = value
	case "PreferredMfaType":
		user.PreferredMfaType = value
	case "TotpSecret":
		user.TotpSecret = value
	case "SignupApplication":
		user.SignupApplication = value
	case "MfaPhoneEnabled":
		user.MfaPhoneEnabled = util.ParseBool(value)
	case "MfaEmailEnabled":
		user.MfaEmailEnabled = util.ParseBool(value)
	case "RecoveryCodes":
		user.RecoveryCodes = strings.Split(value, ",")
	}
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

func (syncer *Syncer) getOriginalUsersFromMap(results []map[string]sql.NullString) []*OriginalUser {
	users := []*OriginalUser{}
	for _, result := range results {
		originalUser := &OriginalUser{
			Address:    []string{},
			Properties: map[string]string{},
			Groups:     []string{},
		}

		for _, tableColumn := range syncer.TableColumns {
			tableColumnName := tableColumn.Name
			if syncer.Type == "Keycloak" && syncer.DatabaseType == "postgres" {
				tableColumnName = strings.ToLower(tableColumnName)
			}

			value := ""
			if strings.Contains(tableColumnName, "+") {
				names := strings.Split(tableColumnName, "+")
				var values []string
				for _, name := range names {
					values = append(values, result[strings.Trim(name, " ")].String)
				}
				value = strings.Join(values, " ")
			} else {
				value = result[tableColumnName].String
			}
			syncer.setUserByKeyValue(originalUser, tableColumn.CasdoorName, value)
		}

		if syncer.Type == "Keycloak" {
			// query and set password and password salt from credential table
			sql := fmt.Sprintf("select * from credential where type = 'password' and user_id = '%s'", originalUser.Id)
			credentialResult, _ := syncer.Ormer.Engine.QueryString(sql)
			if len(credentialResult) > 0 {
				credential := Credential{}
				_ = json.Unmarshal([]byte(credentialResult[0]["SECRET_DATA"]), &credential)
				originalUser.Password = credential.Value
				originalUser.PasswordSalt = credential.Salt
			}
			// query and set signup application from user group table
			sql = fmt.Sprintf("select name from keycloak_group where id = "+
				"(select group_id as gid from user_group_membership where user_id = '%s')", originalUser.Id)
			groupResult, _ := syncer.Ormer.Engine.QueryString(sql)
			if len(groupResult) > 0 {
				originalUser.SignupApplication = groupResult[0]["name"]
			}
			// create time
			i, _ := strconv.ParseInt(originalUser.CreatedTime, 10, 64)
			tm := time.Unix(i/int64(1000), 0)
			originalUser.CreatedTime = tm.Format("2006-01-02T15:04:05+08:00")
			// enable
			value, ok := result["ENABLED"]
			if ok {
				originalUser.IsForbidden = !util.ParseBool(value.String)
			} else {
				originalUser.IsForbidden = !util.ParseBool(result["enabled"].String)
			}
		}

		users = append(users, originalUser)
	}
	return users
}

func (syncer *Syncer) getMapFromOriginalUser(user *OriginalUser) map[string]string {
	m := map[string]string{}
	m["Name"] = user.Name
	m["CreatedTime"] = user.CreatedTime
	m["UpdatedTime"] = user.UpdatedTime
	m["DeletedTime"] = user.DeletedTime
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
	m["MfaPhoneEnabled"] = util.BoolToString(user.MfaPhoneEnabled)
	m["MfaEmailEnabled"] = util.BoolToString(user.MfaEmailEnabled)
	m["RecoveryCodes"] = strings.Join(user.RecoveryCodes, ",")

	m2 := map[string]string{}
	for _, tableColumn := range syncer.TableColumns {
		m2[tableColumn.Name] = m[tableColumn.CasdoorName]
	}

	return m2
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
