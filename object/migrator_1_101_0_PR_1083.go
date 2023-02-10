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
	"strings"

	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

type Migrator_1_101_0_PR_1083 struct{}

type modelV1 struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	ModelText string `xorm:"mediumtext" json:"modelText"`
	IsEnabled bool   `json:"isEnabled"`
}

type permissionV1 struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Users   []string `xorm:"mediumtext" json:"users"`
	Roles   []string `xorm:"mediumtext" json:"roles"`
	Domains []string `xorm:"mediumtext" json:"domains"`

	Model        string   `xorm:"varchar(100)" json:"model"`
	Adapter      string   `xorm:"varchar(100)" json:"adapter"`
	ResourceType string   `xorm:"varchar(100)" json:"resourceType"`
	Resources    []string `xorm:"mediumtext" json:"resources"`
	Actions      []string `xorm:"mediumtext" json:"actions"`
	Effect       string   `xorm:"varchar(100)" json:"effect"`
	IsEnabled    bool     `json:"isEnabled"`

	Submitter   string `xorm:"varchar(100)" json:"submitter"`
	Approver    string `xorm:"varchar(100)" json:"approver"`
	ApproveTime string `xorm:"varchar(100)" json:"approveTime"`
	State       string `xorm:"varchar(100)" json:"state"`
}

func (*Migrator_1_101_0_PR_1083) IsMigrationNeeded(adapter *Adapter) bool {
	exist1, _ := adapter.Engine.IsTableExist("model")
	exist2, _ := adapter.Engine.IsTableExist("permission")
	exist3, _ := adapter.Engine.IsTableExist("permission_rule")

	if exist1 && exist2 && exist3 {
		return true
	}
	return false
}

func (*Migrator_1_101_0_PR_1083) DoMigration(adapter *Adapter) *migrate.Migration {
	migration := migrate.Migration{
		ID: "20230209MigratePermissionRule--Use V5 instead of V1 to store permissionID",
		Migrate: func(engine *xorm.Engine) error {
			models := []*modelV1{}
			err := engine.Table("model").Find(&models, &modelV1{})
			if err != nil {
				panic(err)
			}

			isHit := false
			for _, model := range models {
				if strings.Contains(model.ModelText, "permission") {
					// update model table
					model.ModelText = strings.Replace(model.ModelText, "permission,", "", -1)
					migrateUpdateModel(adapter, model.getId(), model)
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

func (oldModel *modelV1) getId() string {
	return fmt.Sprintf("%s/%s", oldModel.Owner, oldModel.Name)
}

func migrateUpdateModel(adapter *Adapter, id string, modelObj *modelV1) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if migrateGetModel(adapter, owner, name) == nil {
		return false
	}

	if name != modelObj.Name {
		err := migrateModelChangeTrigger(adapter, name, modelObj.Name)
		if err != nil {
			return false
		}
	}
	// check model grammar
	_, err := casbinmodel.NewModelFromString(modelObj.ModelText)
	if err != nil {
		panic(err)
	}

	affected, err := adapter.Engine.Table("model").ID(core.PK{owner, name}).AllCols().Update(modelObj)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func migrateGetModel(adapter *Adapter, owner string, name string) *modelV1 {
	if owner == "" || name == "" {
		return nil
	}

	m := modelV1{Owner: owner, Name: name}
	existed, err := adapter.Engine.Table("model").Get(&m)
	if err != nil {
		panic(err)
	}

	if existed {
		return &m
	} else {
		return nil
	}
}

func migrateModelChangeTrigger(adapter *Adapter, oldName string, newName string) error {
	session := adapter.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	permission := new(permissionV1)
	permission.Model = newName
	_, err = session.Table("permission").Where("model=?", oldName).Update(permission)
	if err != nil {
		return err
	}

	return session.Commit()
}
