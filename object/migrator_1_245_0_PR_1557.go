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

	"github.com/casdoor/casdoor/util"
	"github.com/nyaruka/phonenumbers"
	"github.com/xorm-io/xorm"
	"github.com/xorm-io/xorm/migrate"
)

type Migrator_1_245_0_PR_1557 struct{}

func (*Migrator_1_245_0_PR_1557) IsMigrationNeeded() bool {
	exist, _ := adapter.Engine.IsTableExist("organization")

	if exist {
		return true
	}
	return false
}

func (*Migrator_1_245_0_PR_1557) DoMigration() *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230215organization--transfer phonePrefix to defaultCountryCode, countryCodes",
		Migrate: func(engine *xorm.Engine) error {
			err := adapter.Engine.Sync2(new(Organization))
			if err != nil {
				panic(err)
			}

			organizations := []*Organization{}
			err = engine.Table("organization").Find(&organizations, &Organization{})
			if err != nil {
				panic(err)
			}

			for _, organization := range organizations {
				organization.AccountItems = []*AccountItem{
					{Name: "Organization", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
					{Name: "ID", Visible: true, ViewRule: "Public", ModifyRule: "Immutable"},
					{Name: "Name", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
					{Name: "Display name", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Avatar", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "User type", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
					{Name: "Password", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
					{Name: "Email", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Phone", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Country code", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
					{Name: "Country/Region", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Location", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Affiliation", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Title", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Homepage", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Bio", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
					{Name: "Tag", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
					{Name: "Signup application", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
					{Name: "Roles", Visible: true, ViewRule: "Public", ModifyRule: "Immutable"},
					{Name: "Permissions", Visible: true, ViewRule: "Public", ModifyRule: "Immutable"},
					{Name: "3rd-party logins", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
					{Name: "Properties", Visible: false, ViewRule: "Admin", ModifyRule: "Admin"},
					{Name: "Is admin", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
					{Name: "Is global admin", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
					{Name: "Is forbidden", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
					{Name: "Is deleted", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
					{Name: "WebAuthn credentials", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
					{Name: "Managed accounts", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
				}
				sql := fmt.Sprintf("select phone_prefix from organization where owner='%s' and name='%s'", organization.Owner, organization.Name)
				results, _ := engine.Query(sql)

				phonePrefix := util.ParseInt(string(results[0]["phone_prefix"]))
				organization.CountryCodes = []string{phonenumbers.GetRegionCodeForCountryCode(phonePrefix)}

				UpdateOrganization(util.GetId(organization.Owner, organization.Name), organization)
			}
			return err
		},
	}

	return &migration
}
