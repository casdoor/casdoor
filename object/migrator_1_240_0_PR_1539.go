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
	"errors"

	"github.com/xorm-io/xorm"
	"github.com/xorm-io/xorm/migrate"
)

type Migrator_1_240_0_PR_1539 struct{}

func (*Migrator_1_240_0_PR_1539) IsMigrationNeeded() bool {
	exist, _ := adapter.Engine.IsTableExist("session")
	err := adapter.Engine.Table("session").Find(&[]*Session{})

	if exist && err != nil {
		return true
	}
	return false
}

func (*Migrator_1_240_0_PR_1539) DoMigration() *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230211MigrateSession--Create a new field 'application' for table `session`",
		Migrate: func(engine *xorm.Engine) error {
			if alreadyCreated, _ := engine.IsTableExist("session_tmp"); alreadyCreated {
				return errors.New("there is already a table called 'session_tmp', please rename or delete it for casdoor version migration and restart")
			}

			type oldSession struct {
				Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
				Name        string `xorm:"varchar(100) notnull pk" json:"name"`
				CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

				SessionId []string `json:"sessionId"`
			}

			tx := engine.NewSession()

			defer tx.Close()

			err := tx.Begin()
			if err != nil {
				return err
			}

			err = tx.Table("session_tmp").CreateTable(&Session{})
			if err != nil {
				return err
			}

			oldSessions := []*oldSession{}
			newSessions := []*Session{}

			err = tx.Table("session").Find(&oldSessions)
			if err != nil {
				return err
			}

			for _, oldSession := range oldSessions {
				newApplication := "null"
				if oldSession.Owner == "built-in" {
					newApplication = "app-built-in"
				}
				newSessions = append(newSessions, &Session{
					Owner:       oldSession.Owner,
					Name:        oldSession.Name,
					Application: newApplication,
					CreatedTime: oldSession.CreatedTime,
					SessionId:   oldSession.SessionId,
				})
			}

			rollbackFlag := false
			_, err = tx.Table("session_tmp").Insert(newSessions)
			count1, _ := tx.Table("session_tmp").Count()
			count2, _ := tx.Table("session").Count()

			if err != nil || count1 != count2 {
				rollbackFlag = true
			}

			delete := &Session{
				Application: "null",
			}
			_, err = tx.Table("session_tmp").Delete(*delete)
			if err != nil {
				rollbackFlag = true
			}

			if rollbackFlag {
				tx.DropTable("session_tmp")
				return errors.New("there is something wrong with data migration for table `session`, if there is a table called `session_tmp` not created by you in casdoor, please drop it, then restart anyhow")
			}

			err = tx.DropTable("session")
			if err != nil {
				return errors.New("fail to drop table `session` for casdoor, please drop it and rename the table `session_tmp` to `session` manually and restart")
			}

			// Already drop table `session`
			// Can't find an api from xorm for altering table name
			err = tx.Table("session").CreateTable(&Session{})
			if err != nil {
				return errors.New("there is something wrong with data migration for table `session`, please restart")
			}

			sessions := []*Session{}
			tx.Table("session_tmp").Find(&sessions)
			_, err = tx.Table("session").Insert(sessions)
			if err != nil {
				return errors.New("there is something wrong with data migration for table `session`, please drop table `session` and rename table `session_tmp` to `session` and restart")
			}

			err = tx.DropTable("session_tmp")
			if err != nil {
				return errors.New("fail to drop table `session_tmp` for casdoor, please drop it manually and restart")
			}

			tx.Commit()

			return nil
		},
	}

	return &migration
}
