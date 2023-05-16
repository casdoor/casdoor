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

type Migrator_1_314_0_PR_1841 struct{}

func (*Migrator_1_314_0_PR_1841) IsMigrationNeeded() bool {
	users := []*User{}

	err := adapter.Engine.Table("user").Find(&users)
	if err != nil {
		return false
	}

	for _, u := range users {
		if u.PasswordType != "" {
			return false
		}
	}

	return true
}

func (*Migrator_1_314_0_PR_1841) DoMigration() *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230515MigrateUser--Create a new field 'passwordType' for table `user`",
		Migrate: func(engine *xorm.Engine) error {
			tx := engine.NewSession()

			defer tx.Close()

			err := tx.Begin()
			if err != nil {
				return err
			}

			users := []*User{}
			organizations := []*Organization{}

			err = tx.Table("user").Find(&users)
			if err != nil {
				return err
			}

			err = tx.Table("organization").Find(&organizations)
			if err != nil {
				return err
			}

			passwordTypes := make(map[string]string)
			for _, org := range organizations {
				passwordTypes[org.Name] = org.PasswordType
			}

			columns := []string{
				"password_type",
			}

			for _, u := range users {
				u.PasswordType = passwordTypes[u.Owner]

				_, err := tx.ID(core.PK{u.Owner, u.Name}).Cols(columns...).Update(u)
				if err != nil {
					return err
				}
			}

			tx.Commit()

			return nil
		},
	}

	return &migration
}
