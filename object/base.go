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
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/xorm"
)

type QueryParams struct {
	Owner        string
	Limit        int
	Page         int
	Query        string
	SortField    string
	SortOrder    string
	Organization string
}

type Condition map[string]any
type Or Condition
type And Condition
type Operator string

const (
	Like Operator = "like"
	Eq   Operator = "="
)

type Cond interface {
	String() string
	ToSQL(opt Operator) (string, []any)
}

func (cdt Or) String() string {
	return "OR"
}

func (a And) String() string {
	return "AND"
}

func (o Or) ToSQL(opt Operator) (string, []any) {
	return Condition(o).ToSQL(opt)
}

func (a And) ToSQL(opt Operator) (string, []any) {
	return Condition(a).ToSQL(opt)
}

func (cdt Condition) ToSQL(opt Operator) (string, []any) {
	if len(cdt) == 0 {
		return "", nil
	}
	conditions := make([]string, 0)
	params := make([]interface{}, 0)
	for field, value := range cdt {
		if util.FilterField(field) {
			snakeField := util.SnakeString(field)
			conditions = append(conditions, fmt.Sprintf("%s %s ?", snakeField, opt))
			if opt == Like {
				params = append(params, fmt.Sprintf("%%%v%%", value))
			} else {
				params = append(params, fmt.Sprintf("%v", value))

			}
		}
	}
	return fmt.Sprintf(" (%s) ", strings.Join(conditions, fmt.Sprintf(" %s ", cdt))), params
}

func GetResourcesCount[T any](owner string, params *QueryParams, query, match Cond) (int64, error) {
	session := GetFilterSession(owner, params, query, match)
	var obj T
	return session.Count(&obj)
}

func QueryResources[T any](owner string, params *QueryParams, query, match Cond) ([]*T, error) {
	session := GetFilterSession(owner, params, query, match)
	var obj []*T
	err := session.Find(&obj)
	return obj, err
}

func GetFilterSession(owner string, params *QueryParams, query, match Cond) *xorm.Session {
	session := ormer.Engine.Prepare()
	if params.Page > 0 && params.Limit > 0 {
		session.Limit(params.Limit, (params.Page-1)*params.Limit)
	}
	if owner != "" {
		session = session.And("owner=?", owner)
	}
	if query != nil {
		conditions, sqlParams := query.ToSQL(Like)
		if len(conditions) > 0 {
			session = session.And(conditions, sqlParams...)
		}
	}
	if match != nil {
		matchConditions, matchSQLParams := match.ToSQL(Eq)
		if len(matchConditions) > 0 {
			session = session.And(matchConditions, matchSQLParams...)
		}
	}

	sortField, sortOrder := params.SortField, params.SortOrder
	if sortField == "" || sortOrder == "" {
		sortField = "created_time"
	}
	if sortOrder == "asc" {
		session = session.Asc(util.SnakeString(sortField))
	} else {
		session = session.Desc(util.SnakeString(sortField))
	}
	return session
}
