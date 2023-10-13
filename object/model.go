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

	"github.com/casbin/casbin/v2/config"
	"github.com/casbin/casbin/v2/model"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Model struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(100)" json:"description"`

	ModelText string `xorm:"mediumtext" json:"modelText"`

	model.Model `xorm:"-" json:"-"`
}

func GetModelCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Model{})
}

func GetModels(owner string) ([]*Model, error) {
	models := []*Model{}
	err := ormer.Engine.Desc("created_time").Find(&models, &Model{Owner: owner})
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
	existed, err := ormer.Engine.Get(&m)
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

func GetModelEx(id string) (*Model, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	model, err := getModel(owner, name)
	if err != nil {
		return nil, err
	}
	if model != nil {
		return model, nil
	}

	return getModel("built-in", name)
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

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(modelObj)
	if err != nil {
		return false, err
	}

	return affected != 0, err
}

func AddModel(model *Model) (bool, error) {
	affected, err := ormer.Engine.Insert(model)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteModel(model *Model) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{model.Owner, model.Name}).Delete(&Model{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (m *Model) GetId() string {
	return fmt.Sprintf("%s/%s", m.Owner, m.Name)
}

func modelChangeTrigger(oldName string, newName string) error {
	session := ormer.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	permission := new(Permission)
	permission.Model = newName
	_, err = session.Where("model=?", oldName).Update(permission)
	if err != nil {
		session.Rollback()
		return err
	}

	enforcer := new(Enforcer)
	enforcer.Model = newName
	_, err = session.Where("model=?", oldName).Update(enforcer)
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit()
}

func HasRoleDefinition(m model.Model) bool {
	if m == nil {
		return false
	}
	return m["g"] != nil
}

func (m *Model) initModel() error {
	if m.Model == nil {
		casbinModel, err := model.NewModelFromString(m.ModelText)
		if err != nil {
			return err
		}
		m.Model = casbinModel
	}

	return nil
}

func getModelCfg(m *Model) (map[string]string, error) {
	cfg, err := config.NewConfigFromText(m.ModelText)
	if err != nil {
		return nil, err
	}

	modelCfg := make(map[string]string)
	modelCfg["p"] = cfg.String("policy_definition::p")
	if cfg.String("role_definition::g") != "" {
		modelCfg["g"] = cfg.String("role_definition::g")
	}
	return modelCfg, nil
}
