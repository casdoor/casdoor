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
	"strings"

	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

type Migrator_1_101_0_PR_1083 struct{}

func (*Migrator_1_101_0_PR_1083) IsMigrationNeeded() bool {
	exist1, _ := adapter.Engine.IsTableExist("model")
	exist2, _ := adapter.Engine.IsTableExist("permission")
	exist3, _ := adapter.Engine.IsTableExist("permission_rule")

	if exist1 && exist2 && exist3 {
		return true
	}
	return false
}

func (*Migrator_1_101_0_PR_1083) DoMigration() *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230209MigratePermissionRule--Use V5 instead of V1 to store permissionID",
		Migrate: func(engine *xorm.Engine) error {
			models := []*Model{}
			err := engine.Table("model").Find(&models, &Model{})
			if err != nil {
				panic(err)
			}

			isHit := false
			for _, model := range models {
				if strings.Contains(model.ModelText, "permission") {
					// update model table
					model.ModelText = strings.Replace(model.ModelText, "permission,", "", -1)
					UpdateModel(model.GetId(), model)
					isHit = true
				}
			}

			if isHit {
				// update permission_rule table
				sql := "UPDATE `permission_rule`SET V0 = V1, V1 = V2, V2 = V3, V3 = V4, V4 = V5 WHERE V0 IN (SELECT CONCAT(owner, '/', name) AS permission_id FROM `permission`)"
				_, err = engine.Exec(sql)
				if err != nil {
					return err
				}
			}
			return err
		},
	}

	return &migration
}
