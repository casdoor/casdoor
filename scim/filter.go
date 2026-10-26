// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package scim

import (
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/elimity-com/scim/errors"
	"github.com/scim2/filter-parser/v2"
	"github.com/xorm-io/builder"
)

// attributeCond returns the database condition of a comparison on a lowercase attribute path, e.g. "emails.value".
type attributeCond func(path string, op filter.CompareOperator, value interface{}) (builder.Cond, error)

var userFilterColumns = map[string]string{
	"externalid":         "external_id",
	"username":           "name",
	"displayname":        "display_name",
	"usertype":           "type",
	"profileurl":         "homepage",
	"name.givenname":     "first_name",
	"name.familyname":    "last_name",
	"emails":             "email",
	"emails.value":       "email",
	"phonenumbers":       "phone",
	"phonenumbers.value": "phone",
	"photos":             "avatar",
	"photos.value":       "avatar",
	"addresses.locality": "location",
	"addresses.region":   "region",
	"addresses.country":  "country_code",
	"organization":       "owner",
}

var groupFilterColumns = map[string]string{
	"displayname":  "display_name",
	"organization": "owner",
}

var (
	condFalse   = builder.Expr("1=0")
	likeEscaper = strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
)

func newFilterError(detail string) error {
	return errors.ScimError{
		ScimType: errors.ScimTypeInvalidFilter,
		Detail:   detail,
		Status:   http.StatusBadRequest,
	}
}

// filterToCond converts a SCIM filter to a database condition, so that the count and the pages of a list
// only contain the matching resources.
func filterToCond(expr filter.Expression, parent *filter.AttributePath, attrCond attributeCond) (builder.Cond, error) {
	switch e := expr.(type) {
	case *filter.LogicalExpression:
		left, err := filterToCond(e.Left, parent, attrCond)
		if err != nil {
			return nil, err
		}
		right, err := filterToCond(e.Right, parent, attrCond)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(string(e.Operator), string(filter.OR)) {
			return builder.Or(left, right), nil
		}
		return builder.And(left, right), nil
	case *filter.NotExpression:
		cond, err := filterToCond(e.Expression, parent, attrCond)
		if err != nil {
			return nil, err
		}
		return builder.Not{cond}, nil
	case *filter.ValuePath:
		// emails[value eq "a@b.com"] compares emails.value
		return filterToCond(e.ValueFilter, &e.AttributePath, attrCond)
	case *filter.AttributeExpression:
		path := strings.ToLower(e.AttributePath.AttributeName)
		if parent != nil {
			path = strings.ToLower(parent.AttributeName) + "." + path
		} else if e.AttributePath.SubAttribute != nil {
			path += "." + strings.ToLower(*e.AttributePath.SubAttribute)
		}
		return attrCond(path, e.Operator, e.CompareValue)
	default:
		return nil, newFilterError("the filter is not supported")
	}
}

func compareColumn(column string, op filter.CompareOperator, value interface{}) (builder.Cond, error) {
	if op == filter.PR {
		return builder.And(builder.NotNull{column}, builder.Neq{column: ""}), nil
	}

	s, ok := value.(string)
	if !ok {
		return nil, newFilterError("the attribute for the column: " + column + " can only be compared with a string")
	}

	switch op {
	case filter.EQ:
		return builder.Eq{column: s}, nil
	case filter.NE:
		return builder.Neq{column: s}, nil
	case filter.CO:
		return builder.Expr(column+" LIKE ? ESCAPE '!'", "%"+likeEscaper.Replace(s)+"%"), nil
	case filter.SW:
		return builder.Expr(column+" LIKE ? ESCAPE '!'", likeEscaper.Replace(s)+"%"), nil
	case filter.EW:
		return builder.Expr(column+" LIKE ? ESCAPE '!'", "%"+likeEscaper.Replace(s)), nil
	case filter.GT:
		return builder.Gt{column: s}, nil
	case filter.GE:
		return builder.Gte{column: s}, nil
	case filter.LT:
		return builder.Lt{column: s}, nil
	case filter.LE:
		return builder.Lte{column: s}, nil
	default:
		return nil, newFilterError("the operator: " + string(op) + " is not supported")
	}
}

func userFilterCond(path string, op filter.CompareOperator, value interface{}) (builder.Cond, error) {
	column, ok := userFilterColumns[path]
	if !ok {
		return nil, newFilterError("filtering users by the attribute: " + path + " is not supported")
	}
	return compareColumn(column, op, value)
}

func groupFilterCond(path string, op filter.CompareOperator, value interface{}) (builder.Cond, error) {
	if path == "members" || path == "members.value" {
		userId, ok := value.(string)
		if !ok || (op != filter.EQ && op != filter.NE) {
			return nil, newFilterError("the attribute: " + path + " only supports eq and ne with a string")
		}
		cond, err := memberCond(userId)
		if err != nil {
			return nil, err
		}
		if op == filter.NE {
			return builder.Not{cond}, nil
		}
		return cond, nil
	}

	column, ok := groupFilterColumns[path]
	if !ok {
		return nil, newFilterError("filtering groups by the attribute: " + path + " is not supported")
	}
	return compareColumn(column, op, value)
}

// memberCond matches the groups of the user with the SCIM id.
func memberCond(userId string) (builder.Cond, error) {
	user, err := object.GetUserByUserIdOnly(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return condFalse, nil
	}

	cond := builder.Cond(condFalse)
	for _, groupId := range user.Groups {
		owner, name, err := util.GetOwnerAndNameFromIdWithError(groupId)
		if err == nil {
			cond = cond.Or(builder.Eq{"owner": owner, "name": name})
		}
	}
	return cond, nil
}
