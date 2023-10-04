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
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

type DatabaseSyncer struct {
	Ormer            *Ormer
	Table            string
	TableColumns     []*TableColumn
	Type             string
	DatabaseType     string
	AvatarBaseUrl    string
	AffiliationTable string
}

func NewDatabaseSyncer(typ string, databaseType string, sslMode string, user string, password string, host string, port int, database string, table string, isCloudIntranet bool, tableColumns []*TableColumn, avatarBaseUrl string, affiliationTable string) (*DatabaseSyncer, error) {
	var dataSourceName string
	if databaseType == "mssql" {
		dataSourceName = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", user, password, host, port, database)
	} else if databaseType == "postgres" {
		if sslMode == "" {
			sslMode = "disable"
		}
		dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=%s dbname=%s", user, password, host, port, sslMode, database)
	} else {
		dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)
	}

	if !isCloudIntranet {
		dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
	}

	dbSyncer := &DatabaseSyncer{
		Table:            table,
		TableColumns:     tableColumns,
		Type:             typ,
		DatabaseType:     databaseType,
		AvatarBaseUrl:    avatarBaseUrl,
		AffiliationTable: affiliationTable,
	}
	dbSyncer.Table = dbSyncer.getTable()

	ormer, err := NewAdapter(databaseType, dataSourceName, database)
	if err != nil {
		return nil, err
	}
	dbSyncer.Ormer = ormer

	return dbSyncer, nil
}

func (dbSyncer *DatabaseSyncer) getOriginalUsers() ([]*OriginalUser, error) {
	var results []map[string]sql.NullString
	err := dbSyncer.Ormer.Engine.Table(dbSyncer.getTable()).Find(&results)
	if err != nil {
		return nil, err
	}

	// Memory leak problem handling
	// https://github.com/casdoor/casdoor/issues/1256
	users := dbSyncer.getOriginalUsersFromMap(results)
	for _, m := range results {
		for k := range m {
			delete(m, k)
		}
	}

	return users, nil
}

func (dbSyncer *DatabaseSyncer) GetOriginalUserMap() ([]*OriginalUser, map[string]*OriginalUser, error) {
	users, err := dbSyncer.getOriginalUsers()
	if err != nil {
		return users, nil, err
	}

	m := map[string]*OriginalUser{}
	for _, user := range users {
		m[user.Id] = user
	}
	return users, m, nil
}

func (dbSyncer *DatabaseSyncer) UpdateUser(oUser *OriginalUser) (bool, error) {
	key := dbSyncer.getKey()
	m := dbSyncer.getMapFromOriginalUser(oUser)
	pkValue := m[key]
	delete(m, key)

	affected, err := dbSyncer.Ormer.Engine.Table(dbSyncer.getTable()).ID(pkValue).Update(&m)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (dbSyncer *DatabaseSyncer) AddUser(oUser *OriginalUser) (bool, error) {
	m := dbSyncer.getMapFromOriginalUser(oUser)
	affected, err := dbSyncer.Ormer.Engine.Table(dbSyncer.getTable()).Insert(m)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (dbSyncer *DatabaseSyncer) GetOriginalGroupMap() ([]*OriginalGroup, map[string]*OriginalGroup, error) {
	return nil, nil, nil
}

func (dbSyncer *DatabaseSyncer) UpdateGroup(oGroup *OriginalGroup) (bool, error) {
	return true, nil
}

func (dbSyncer *DatabaseSyncer) AddGroup(oGroup *OriginalGroup) (bool, error) {
	return true, nil
}

func (dbSyncer *DatabaseSyncer) GetAffiliationMap() ([]*Affiliation, map[int]string, error) {
	affiliations, err := dbSyncer.getAffiliations()
	if err != nil {
		return nil, nil, err
	}

	m := map[int]string{}
	for _, affiliation := range affiliations {
		m[affiliation.Id] = affiliation.Name
	}
	return affiliations, m, nil
}

func (dbSyncer *DatabaseSyncer) getAffiliations() ([]*Affiliation, error) {
	affiliations := []*Affiliation{}
	err := dbSyncer.Ormer.Engine.Table(dbSyncer.AffiliationTable).Asc("id").Find(&affiliations)
	if err != nil {
		return nil, err
	}

	return affiliations, nil
}

func (dbSyncer *DatabaseSyncer) getKey() string {
	key := "id"
	hasKey := false
	hasId := false
	for _, tableColumn := range dbSyncer.TableColumns {
		if tableColumn.IsKey {
			hasKey = true
			key = tableColumn.Name
		}
		if tableColumn.Name == "id" {
			hasId = true
		}
	}

	if !hasKey && !hasId {
		key = dbSyncer.TableColumns[0].Name
	}

	return key
}

func (dbSyncer *DatabaseSyncer) getMapFromOriginalUser(user *OriginalUser) map[string]string {
	m := map[string]string{}
	m["Name"] = user.Name
	m["CreatedTime"] = user.CreatedTime
	m["UpdatedTime"] = user.UpdatedTime
	m["Id"] = user.Id
	m["Type"] = user.Type
	m["Password"] = user.Password
	m["PasswordSalt"] = user.PasswordSalt
	m["DisplayName"] = user.DisplayName
	m["Avatar"] = dbSyncer.getFullAvatarUrl(user.Avatar)
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
	for _, tableColumn := range dbSyncer.TableColumns {
		m2[tableColumn.Name] = m[tableColumn.CasdoorName]
	}

	return m2
}

func (dbSyncer *DatabaseSyncer) getFullAvatarUrl(avatar string) string {
	if dbSyncer.AvatarBaseUrl == "" {
		return avatar
	}

	if !strings.HasPrefix(avatar, "http") {
		return fmt.Sprintf("%s%s", dbSyncer.AvatarBaseUrl, avatar)
	}
	return avatar
}

func (dbSyncer *DatabaseSyncer) getTable() string {
	if dbSyncer.DatabaseType == "mssql" {
		return fmt.Sprintf("[%s]", dbSyncer.Table)
	} else {
		return dbSyncer.Table
	}
}

func (dbSyncer *DatabaseSyncer) getPartialAvatarUrl(avatar string) string {
	if strings.HasPrefix(avatar, dbSyncer.AvatarBaseUrl) {
		return avatar[len(dbSyncer.AvatarBaseUrl):]
	}
	return avatar
}

func (dbSyncer *DatabaseSyncer) setUserByKeyValue(user *User, key string, value string) {
	switch key {
	case "Name":
		user.Name = value
	case "CreatedTime":
		user.CreatedTime = value
	case "UpdatedTime":
		user.UpdatedTime = value
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
		user.Avatar = dbSyncer.getPartialAvatarUrl(value)
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
	}
}

func (dbSyncer *DatabaseSyncer) getOriginalUsersFromMap(results []map[string]sql.NullString) []*OriginalUser {
	users := []*OriginalUser{}
	for _, result := range results {
		originalUser := &OriginalUser{
			Address:    []string{},
			Properties: map[string]string{},
			Groups:     []string{},
		}

		for _, tableColumn := range dbSyncer.TableColumns {
			tableColumnName := tableColumn.Name
			if dbSyncer.Type == "Keycloak" && dbSyncer.DatabaseType == "postgres" {
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
			dbSyncer.setUserByKeyValue(originalUser, tableColumn.CasdoorName, value)
		}

		if dbSyncer.Type == "Keycloak" {
			// query and set password and password salt from credential table
			sql := fmt.Sprintf("select * from credential where type = 'password' and user_id = '%s'", originalUser.Id)
			credentialResult, _ := dbSyncer.Ormer.Engine.QueryString(sql)
			if len(credentialResult) > 0 {
				credential := Credential{}
				_ = json.Unmarshal([]byte(credentialResult[0]["SECRET_DATA"]), &credential)
				originalUser.Password = credential.Value
				originalUser.PasswordSalt = credential.Salt
			}
			// query and set signup application from user group table
			sql = fmt.Sprintf("select name from keycloak_group where id = "+
				"(select group_id as gid from user_group_membership where user_id = '%s')", originalUser.Id)
			groupResult, _ := dbSyncer.Ormer.Engine.QueryString(sql)
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
