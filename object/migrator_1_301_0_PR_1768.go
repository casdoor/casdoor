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

type Migrator_1_301_0_PR_1768 struct{}

func (*Migrator_1_301_0_PR_1768) IsMigrationNeeded() bool {
	exist, _ := adapter.Engine.IsTableExist("organization")
	return exist
}

func (*Migrator_1_301_0_PR_1768) DoMigration() *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230420MigrateOrganization--Create a new field 'properties' for table `organization`",
		Migrate: func(engine *xorm.Engine) error {
			return engine.Sync2(&Organization{})
		},
	}

	return &migration
}
