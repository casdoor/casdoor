// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casbin/casbin/v2/model"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Model struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	ModelText string `xorm:"mediumtext" json:"modelText"`
	IsEnabled bool   `json:"isEnabled"`
}

func GetModelCount(owner, field, value string) (int, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Model{})
	if err != nil {
		return int(count), err
	}

	return int(count), nil
}

func GetModels(owner string) ([]*Model, error) {
	models := []*Model{}
	err := adapter.Engine.Desc("created_time").Find(&models, &Model{Owner: owner})
	if err != nil {
		return models, err
	}

	return models, nil
}

func GetPaginationModels(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Model, error) {
	models := []*Model{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&models)
	if err != nil {
		return models, err
	}

	return models, nil
}

func getModel(owner string, name string) (*Model, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	m := Model{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&m)
	if err != nil {
		return &m, err
	}

	if existed {
		return &m, nil
	} else {
		return nil, nil
	}
}

func GetModel(id string) (*Model, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getModel(owner, name)
}

func UpdateModelWithCheck(id string, modelObj *Model) error {
	// check model grammar
	_, err := model.NewModelFromString(modelObj.ModelText)
	if err != nil {
		return err
	}
	_, err = UpdateModel(id, modelObj)
	if err != nil {
		return err
	}

	return nil
}

func UpdateModel(id string, modelObj *Model) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if m, err := getModel(owner, name); err != nil {
		return false, err
	} else if m == nil {
		return false, nil
	}

	if name != modelObj.Name {
		err := modelChangeTrigger(name, modelObj.Name)
		if err != nil {
			return false, err
		}
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(modelObj)
	if err != nil {
		return false, err
	}

	return affected != 0, err
}

func AddModel(model *Model) (bool, error) {
	affected, err := adapter.Engine.Insert(model)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteModel(model *Model) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{model.Owner, model.Name}).Delete(&Model{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (model *Model) GetId() string {
	return fmt.Sprintf("%s/%s", model.Owner, model.Name)
}

func modelChangeTrigger(oldName string, newName string) error {
	session := adapter.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	permission := new(Permission)
	permission.Model = newName
	_, err = session.Where("model=?", oldName).Update(permission)
	if err != nil {
		return err
	}

	return session.Commit()
}

func HasRoleDefinition(m model.Model) bool {
	return m["g"] != nil
}
