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

type Migrator_1_336_0_PR_1970 struct{}

func (*Migrator_1_336_0_PR_1970) IsMigrationNeeded() bool {
	count, err := adapter.Engine.Where("password_type=? and password_salt=?", "salt", "").Count(&User{})
	if err != nil {
		// table doesn't exist
		return false
	}

	return count > 0
}

func (*Migrator_1_336_0_PR_1970) DoMigration() *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230615MigrateUser--Update salt if necessary",
		Migrate: func(engine *xorm.Engine) error {
			tx := engine.NewSession()

			defer tx.Close()

			err := tx.Begin()
			if err != nil {
				return err
			}

			users := []*User{}
			err = tx.Table("user").Where("password_type=? and password_salt=?", "salt", "").Find(&users)
			if err != nil {
				return err
			}

			keys := make(map[string]struct{})
			userOwners := []string{}
			for _, user := range users {
				if _, value := keys[user.Owner]; !value {
					keys[user.Owner] = struct{}{}
					userOwners = append(userOwners, user.Owner)
				}
			}

			organizations := []*Organization{}
			err = tx.Where("owner = ?", "admin").In("name", userOwners).Find(&organizations)
			if err != nil {
				return err
			}
			organizationSalts := make(map[string]string)
			for _, organization := range organizations {
				organizationSalts[organization.Name] = organization.PasswordSalt
			}

			for _, user := range users {
				user.PasswordSalt = organizationSalts[user.Owner]
				_, err = tx.Where("owner = ?", user.Owner).And("name = ?", user.Name).Cols("password_salt").Update(user)
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
