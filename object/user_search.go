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

package object

import (
	"fmt"

	"github.com/xorm-io/builder"
	"github.com/xorm-io/xorm"
)

// BuildUserSearchCond builds OR condition for global user search by name, display_name and email.
// Empty search returns nil (no filter). columnPrefix is used for table aliases (e.g. "a.").
func BuildUserSearchCond(search string, columnPrefix string) builder.Cond {
	if search == "" {
		return nil
	}

	pattern := fmt.Sprintf("%%%s%%", search)
	nameCol := columnPrefix + "name"
	displayNameCol := columnPrefix + "display_name"
	emailCol := columnPrefix + "email"

	return builder.Or(
		builder.Expr(fmt.Sprintf("%s like ?", nameCol), pattern),
		builder.Expr(fmt.Sprintf("%s like ?", displayNameCol), pattern),
		builder.Expr(fmt.Sprintf("%s like ?", emailCol), pattern),
	)
}

// applyUserSearchFilter applies global user search to the session. No-op when search is empty.
func applyUserSearchFilter(session *xorm.Session, search string, columnPrefix string) *xorm.Session {
	cond := BuildUserSearchCond(search, columnPrefix)
	if cond == nil {
		return session
	}
	return session.And(cond)
}
