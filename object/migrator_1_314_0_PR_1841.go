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

type Migrator_1_314_0_PR_1841 struct{}

func (*Migrator_1_314_0_PR_1841) IsMigrationNeeded() bool {
	count, err := adapter.Engine.Where("password_type=?", "").Count(&User{})
	if err != nil {
		panic(err)
	}

	return count > 100
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

			organizations := []*Organization{}
			err = tx.Table("organization").Find(&organizations)
			if err != nil {
				return err
			}

			for _, organization := range organizations {
				user := &User{PasswordType: organization.PasswordType}
				_, err = tx.Where("owner = ?", organization.Name).Cols("password_type").Update(user)
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
