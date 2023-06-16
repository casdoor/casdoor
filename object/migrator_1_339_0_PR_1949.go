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
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
	"github.com/xorm-io/xorm/migrate"
)

type Migrator_1_339_0_PR_1949 struct{}

func (*Migrator_1_339_0_PR_1949) IsMigrationNeeded() bool {
	count, err := adapter.Engine.Where("password_complex_options=?", "").Count(&Organization{})
	if err != nil {
		// table doesn't exist
		return false
	}

	return count > 100
}

func (*Migrator_1_339_0_PR_1949) DoMigration() *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230616MigrateUser--Create a new field 'passwordComplexOptions' for table `organization`",
		Migrate: func(engine *xorm.Engine) error {
			tx := engine.NewSession()

			defer tx.Close()

			err := tx.Begin()
			if err != nil {
				return err
			}

			pageSize := 100
			page := 1
			const defaultOption = "AtLeast6"
			const batchSize = 100

			for {
				// Paginate the query
				organizations := []*Organization{}
				err = engine.Limit(pageSize, (page-1)*pageSize).Find(&organizations)
				if err != nil {
					return err
				}

				if len(organizations) == 0 {
					// All data has been processed
					break
				}

				// Update the password_complex_options field for the current page
				tx := engine.NewSession()
				defer tx.Close()
				err = tx.Begin()
				if err != nil {
					return err
				}

				for _, organization := range organizations {
					if organization.PasswordComplexOptions == nil {
						organization.PasswordComplexOptions = []string{defaultOption}
						// Accumulate the changes in the tx
						_, err = tx.ID(core.PK{organization.Owner, organization.Name}).Cols("password_complex_options").Update(organization)
						if err != nil {
							tx.Rollback()
							return err
						}
					}

					// Commit the changes in batches
					if len(organizations) >= batchSize {
						err = tx.Commit()
						if err != nil {
							return err
						}
						// Start a new batch
						tx = engine.NewSession()
						defer tx.Close()
						err = tx.Begin()
						if err != nil {
							return err
						}
					}
				}

				// Commit any remaining changes

				err = tx.Commit()
				if err != nil {
					return err
				}

				page++
			}
			return nil
		},
	}

	return &migration
}
