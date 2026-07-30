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
	"fmt"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/xorm"
)

func GetSession(owner string, offset, limit int, field, value, sortField, sortOrder string) *xorm.Session {
	session := ormer.Engine.Prepare()
	if offset != -1 && limit != -1 {
		session.Limit(limit, offset)
	}
	if owner != "" {
		session = session.And("owner=?", owner)
	}
	if field != "" && value != "" {
		if util.FilterField(field) {
			session = session.And(fmt.Sprintf("%s like ?", util.CamelToSnakeCase(field)), fmt.Sprintf("%%%s%%", value))
		}
	}
	if sortField == "" || sortOrder == "" || !util.FilterField(sortField) {
		sortField = "created_time"
	}
	if sortOrder == "ascend" {
		session = session.Asc(util.CamelToSnakeCase(sortField))
	} else {
		session = session.Desc(util.CamelToSnakeCase(sortField))
	}
	return session
}

func GetSessionForUser(owner string, offset, limit int, field, value, sortField, sortOrder string) *xorm.Session {
	session := ormer.Engine.Prepare()
	if offset != -1 && limit != -1 {
		session.Limit(limit, offset)
	}
	if owner != "" {
		if offset == -1 {
			session = session.And("owner=?", owner)
		} else {
			session = session.And("a.owner=?", owner)
		}
	}
	if field != "" && value != "" {
		if util.FilterField(field) {
			if offset != -1 {
				field = fmt.Sprintf("a.%s", field)
			}
			session = session.And(fmt.Sprintf("%s like ?", util.CamelToSnakeCase(field)), fmt.Sprintf("%%%s%%", value))
		}
	}
	if sortField == "" || sortOrder == "" || !util.FilterField(sortField) {
		sortField = "created_time"
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tableName := tableNamePrefix + "user"
	if offset == -1 {
		if sortOrder == "ascend" {
			session = session.Asc(util.CamelToSnakeCase(sortField))
		} else {
			session = session.Desc(util.CamelToSnakeCase(sortField))
		}
	} else {
		if sortOrder == "ascend" {
			session = session.Alias("a").
				Join("INNER", []string{tableName, "b"}, "a.owner = b.owner and a.name = b.name").
				Select("b.*").
				Asc("a." + util.CamelToSnakeCase(sortField))
		} else {
			session = session.Alias("a").
				Join("INNER", []string{tableName, "b"}, "a.owner = b.owner and a.name = b.name").
				Select("b.*").
				Desc("a." + util.CamelToSnakeCase(sortField))
		}
	}

	return session
}

// UserSearchFilters holds optional substring filters applied with AND semantics.
type UserSearchFilters struct {
	Name        string
	DisplayName string
	Email       string
	Phone       string
	Affiliation string
}

var userSearchFilterFields = []struct {
	field string
	value func(UserSearchFilters) string
}{
	{"name", func(f UserSearchFilters) string { return f.Name }},
	{"displayName", func(f UserSearchFilters) string { return f.DisplayName }},
	{"email", func(f UserSearchFilters) string { return f.Email }},
	{"phone", func(f UserSearchFilters) string { return f.Phone }},
	{"affiliation", func(f UserSearchFilters) string { return f.Affiliation }},
}

func applyUserSearchFilters(session *xorm.Session, filters UserSearchFilters, useAlias bool) *xorm.Session {
	for _, item := range userSearchFilterFields {
		value := item.value(filters)
		if value == "" || !util.FilterField(item.field) {
			continue
		}
		field := item.field
		if useAlias {
			field = fmt.Sprintf("a.%s", field)
		}
		session = session.And(fmt.Sprintf("%s like ?", util.CamelToSnakeCase(field)), fmt.Sprintf("%%%s%%", value))
	}
	return session
}

func GetSessionForUserSearch(owner string, offset, limit int, filters UserSearchFilters, sortField, sortOrder string) *xorm.Session {
	session := ormer.Engine.Prepare()
	paginated := offset != -1 && limit != -1
	if paginated {
		session.Limit(limit, offset)
	}
	if owner != "" {
		if paginated {
			session = session.And("a.owner=?", owner)
		} else {
			session = session.And("owner=?", owner)
		}
	}
	session = applyUserSearchFilters(session, filters, paginated)

	if sortField == "" || sortOrder == "" || !util.FilterField(sortField) {
		sortField = "created_time"
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tableName := tableNamePrefix + "user"
	if !paginated {
		if sortOrder == "ascend" {
			session = session.Asc(util.CamelToSnakeCase(sortField))
		} else {
			session = session.Desc(util.CamelToSnakeCase(sortField))
		}
	} else {
		if sortOrder == "ascend" {
			session = session.Alias("a").
				Join("INNER", []string{tableName, "b"}, "a.owner = b.owner and a.name = b.name").
				Select("b.*").
				Asc("a." + util.CamelToSnakeCase(sortField))
		} else {
			session = session.Alias("a").
				Join("INNER", []string{tableName, "b"}, "a.owner = b.owner and a.name = b.name").
				Select("b.*").
				Desc("a." + util.CamelToSnakeCase(sortField))
		}
	}

	return session
}
