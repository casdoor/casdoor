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
	"github.com/xorm-io/xorm"
	"github.com/xorm-io/xorm/migrate"
)

type Migrator_1_333_0_PR_1950 struct{}

const changePasswordName = "Change password"

func (*Migrator_1_333_0_PR_1950) IsMigrationNeeded() bool {
	orgranizations := []*Organization{}

	err := adapter.Engine.Table("organization").Find(&orgranizations)
	if err != nil {
		return false
	}

	for _, o := range orgranizations {
		missingPasswordChangeItem := true
		for _, ai := range o.AccountItems {
			if ai.Name == changePasswordName {
				missingPasswordChangeItem = false
				break
			}
		}
		if missingPasswordChangeItem {
			return missingPasswordChangeItem
		}
	}

	return false
}

func (*Migrator_1_333_0_PR_1950) DoMigration() *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230608MigrateOrganization--Update AccountItems",
		Migrate: func(engine *xorm.Engine) error {
			tx := engine.NewSession()

			defer tx.Close()

			err := tx.Begin()
			if err != nil {
				return err
			}

			organizations := []*Organization{}
			err = tx.Table("organization").Find(&organizations)
			if err != nil {
				return err
			}

			for _, organization := range organizations {
				missingPasswordChangeItem := true
				for _, ai := range organization.AccountItems {
					if ai.Name == changePasswordName {
						missingPasswordChangeItem = false
						break
					}
				}
				if missingPasswordChangeItem {
					item := AccountItem{Name: changePasswordName, Visible: true, ViewRule: "Admin", ModifyRule: "Admin"}
					organization.AccountItems = append(organization.AccountItems, &item)
					_, err = tx.Cols("account_items").Update(organization)
					if err != nil {
						return err
					}
				}
			}

			tx.Commit()

			return nil
		},
	}

	return &migration
}
