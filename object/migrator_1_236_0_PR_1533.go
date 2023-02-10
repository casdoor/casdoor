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

	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

type Migrator_1_236_0_PR_1533 struct{}

type sessionV2 struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	Application string `xorm:"varchar(100) notnull pk" json:"application"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	SessionId []string `json:"sessionId"`
}

type sessionV1 struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	SessionId []string `json:"sessionId"`
}

func (*Migrator_1_236_0_PR_1533) IsMigrationNeeded(adapter *Adapter) bool {
	exist, _ := adapter.Engine.IsTableExist("session")
	err := adapter.Engine.Table("session").Find(&[]*sessionV2{})

	if exist && err != nil {
		return true
	}
	return false
}

func (*Migrator_1_236_0_PR_1533) DoMigration(adapter *Adapter) *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230210MigrateSession--Create a new field called 'application' and add it to the primary key for table `session`",
		Migrate: func(engine *xorm.Engine) error {
			var err error
			tx := adapter.Engine.NewSession()

			if alreadyCreated, _ := adapter.Engine.IsTableExist("session_tmp"); alreadyCreated {
				return errors.New("there is already a table called 'session_tmp', please rename or delete it for casdoor version migration and restart")
			}

			tx.Table("session_tmp").CreateTable(&sessionV2{})

			oldSessions := []*sessionV1{}
			newSessions := []*sessionV2{}

			tx.Table("session").Find(&oldSessions)

			for _, oldSession := range oldSessions {
				newApplication := "null"
				if oldSession.Owner == "built-in" {
					newApplication = "app-built-in"
				}
				newSessions = append(newSessions, &sessionV2{
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

			delete := &sessionV2{
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
			err = tx.Table("session").CreateTable(&sessionV2{})
			if err != nil {
				return errors.New("there is something wrong with data migration for table `session`, please restart")
			}

			sessions := []*sessionV2{}
			tx.Table("session_tmp").Find(&sessions)
			_, err = tx.Table("session").Insert(sessions)
			if err != nil {
				return errors.New("there is something wrong with data migration for table `session`, please drop table `session` and rename table `session_tmp` to `session` and restart")
			}

			err = tx.DropTable("session_tmp")
			if err != nil {
				return errors.New("fail to drop table `session_tmp` for casdoor, please drop it manually and restart")
			}

			return nil
		},
	}

	return &migration
}
