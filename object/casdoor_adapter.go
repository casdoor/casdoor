// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/core"
)

type CasdoorAdapter struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Type  string `xorm:"varchar(100)" json:"type"`
	Model string `xorm:"varchar(100)" json:"model"`

	DatabaseType    string `xorm:"varchar(100)" json:"databaseType"`
	Host            string `xorm:"varchar(100)" json:"host"`
	Port            string `xorm:"varchar(20)" json:"port"`
	User            string `xorm:"varchar(100)" json:"user"`
	Password        string `xorm:"varchar(100)" json:"password"`
	Database        string `xorm:"varchar(100)" json:"database"`
	Table           string `xorm:"varchar(100)" json:"table"`
	TableNamePrefix string `xorm:"varchar(100)" json:"tableNamePrefix"`
	File            string `xorm:"varchar(100)" json:"file"`
	DataSourceName  string `xorm:"varchar(200)" json:"dataSourceName"`
	IsEnabled       bool   `json:"isEnabled"`

	Adapter *xormadapter.Adapter `xorm:"-" json:"-"`
}

func GetCasdoorAdapterCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&CasdoorAdapter{})
}

func GetCasdoorAdapters(owner string) ([]*CasdoorAdapter, error) {
	adapters := []*CasdoorAdapter{}
	err := adapter.Engine.Desc("created_time").Find(&adapters, &CasdoorAdapter{Owner: owner})
	if err != nil {
		return adapters, err
	}

	return adapters, nil
}

func GetPaginationCasdoorAdapters(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*CasdoorAdapter, error) {
	adapters := []*CasdoorAdapter{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&adapters)
	if err != nil {
		return adapters, err
	}

	return adapters, nil
}

func getCasdoorAdapter(owner, name string) (*CasdoorAdapter, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	casdoorAdapter := CasdoorAdapter{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&casdoorAdapter)
	if err != nil {
		return nil, err
	}

	if existed {
		return &casdoorAdapter, nil
	} else {
		return nil, nil
	}
}

func GetCasdoorAdapter(id string) (*CasdoorAdapter, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getCasdoorAdapter(owner, name)
}

func UpdateCasdoorAdapter(id string, casdoorAdapter *CasdoorAdapter) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if casdoorAdapter, err := getCasdoorAdapter(owner, name); casdoorAdapter == nil {
		return false, err
	}

	if name != casdoorAdapter.Name {
		err := casbinAdapterChangeTrigger(name, casdoorAdapter.Name)
		if err != nil {
			return false, err
		}
	}

	session := adapter.Engine.ID(core.PK{owner, name}).AllCols()
	if casdoorAdapter.Password == "***" {
		session.Omit("password")
	}
	affected, err := session.Update(casdoorAdapter)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddCasdoorAdapter(casdoorAdapter *CasdoorAdapter) (bool, error) {
	affected, err := adapter.Engine.Insert(casdoorAdapter)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteCasdoorAdapter(casdoorAdapter *CasdoorAdapter) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{casdoorAdapter.Owner, casdoorAdapter.Name}).Delete(&CasdoorAdapter{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (casdoorAdapter *CasdoorAdapter) GetId() string {
	return fmt.Sprintf("%s/%s", casdoorAdapter.Owner, casdoorAdapter.Name)
}

func (casdoorAdapter *CasdoorAdapter) getTable() string {
	if casdoorAdapter.DatabaseType == "mssql" {
		return fmt.Sprintf("[%s]", casdoorAdapter.Table)
	} else {
		return casdoorAdapter.Table
	}
}

func initEnforcer(modelObj *Model, casdoorAdapter *CasdoorAdapter) (*casbin.Enforcer, error) {
	// init Adapter
	if casdoorAdapter.Adapter == nil {
		var dataSourceName string
		if casdoorAdapter.DatabaseType == "mssql" {
			dataSourceName = fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", casdoorAdapter.User, casdoorAdapter.Password, casdoorAdapter.Host, casdoorAdapter.Port, casdoorAdapter.Database)
		} else if casdoorAdapter.DatabaseType == "postgres" {
			dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable dbname=%s", casdoorAdapter.User, casdoorAdapter.Password, casdoorAdapter.Host, casdoorAdapter.Port, casdoorAdapter.Database)
		} else {
			dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%s)/", casdoorAdapter.User, casdoorAdapter.Password, casdoorAdapter.Host, casdoorAdapter.Port)
		}

		if !isCloudIntranet {
			dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
		}

		var err error
		casdoorAdapter.Adapter, err = xormadapter.NewAdapterByEngineWithTableName(NewAdapter(casdoorAdapter.DatabaseType, dataSourceName, casdoorAdapter.Database).Engine, casdoorAdapter.getTable(), "")
		if err != nil {
			return nil, err
		}
	}

	// init Model
	m, err := model.NewModelFromString(modelObj.ModelText)
	if err != nil {
		return nil, err
	}

	// init Enforcer
	enforcer, err := casbin.NewEnforcer(m, casdoorAdapter.Adapter)
	if err != nil {
		return nil, err
	}

	return enforcer, nil
}

func (casdoorAdapter *CasdoorAdapter) initAdapter() (*xormadapter.Adapter, error) {
	// init Adapter
	if casdoorAdapter.Adapter == nil {
		var dataSourceName string
		if casdoorAdapter.DataSourceName != "" {
			dataSourceName = casdoorAdapter.DataSourceName
		} else {
			switch casdoorAdapter.DatabaseType {
			case "mssql":
				dataSourceName = fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", casdoorAdapter.User,
					casdoorAdapter.Password, casdoorAdapter.Host, casdoorAdapter.Port, casdoorAdapter.Database)
			case "mysql":
				dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%s)/", casdoorAdapter.User,
					casdoorAdapter.Password, casdoorAdapter.Host, casdoorAdapter.Port)
			case "postgres":
				dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable dbname=%s", casdoorAdapter.User,
					casdoorAdapter.Password, casdoorAdapter.Host, casdoorAdapter.Port, casdoorAdapter.Database)
			case "CockroachDB":
				dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable dbname=%s serial_normalization=virtual_sequence",
					casdoorAdapter.User, casdoorAdapter.Password, casdoorAdapter.Host, casdoorAdapter.Port, casdoorAdapter.Database)
			case "sqlite3":
				dataSourceName = fmt.Sprintf("file:%s", casdoorAdapter.File)
			default:
				return nil, fmt.Errorf("unsupported database type: %s", casdoorAdapter.DatabaseType)
			}
		}

		if !isCloudIntranet {
			dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
		}

		var err error
		casdoorAdapter.Adapter, err = xormadapter.NewAdapterByEngineWithTableName(NewAdapter(casdoorAdapter.DatabaseType, dataSourceName, casdoorAdapter.Database).Engine, casdoorAdapter.getTable(), casdoorAdapter.TableNamePrefix)
		if err != nil {
			return nil, err
		}
	}
	return casdoorAdapter.Adapter, nil
}

func casbinAdapterChangeTrigger(oldName string, newName string) error {
	session := adapter.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	enforcer := new(Enforcer)
	enforcer.Adapter = newName
	_, err = session.Where("adapter=?", oldName).Update(enforcer)
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit()
}

func safeReturn(policy []string, i int) string {
	if len(policy) > i {
		return policy[i]
	} else {
		return ""
	}
}

func matrixToCasbinRules(Ptype string, policies [][]string) []*xormadapter.CasbinRule {
	res := []*xormadapter.CasbinRule{}

	for _, policy := range policies {
		line := xormadapter.CasbinRule{
			Ptype: Ptype,
			V0:    safeReturn(policy, 0),
			V1:    safeReturn(policy, 1),
			V2:    safeReturn(policy, 2),
			V3:    safeReturn(policy, 3),
			V4:    safeReturn(policy, 4),
			V5:    safeReturn(policy, 5),
		}
		res = append(res, &line)
	}

	return res
}

func SyncPolicies(casdoorAdapter *CasdoorAdapter) ([]*xormadapter.CasbinRule, error) {
	modelObj, err := getModel(casdoorAdapter.Owner, casdoorAdapter.Model)
	if err != nil {
		return nil, err
	}

	if modelObj == nil {
		return nil, fmt.Errorf("The model: %s does not exist", util.GetId(casdoorAdapter.Owner, casdoorAdapter.Model))
	}

	enforcer, err := initEnforcer(modelObj, casdoorAdapter)
	if err != nil {
		return nil, err
	}

	policies := matrixToCasbinRules("p", enforcer.GetPolicy())
	if strings.Contains(modelObj.ModelText, "[role_definition]") {
		policies = append(policies, matrixToCasbinRules("g", enforcer.GetGroupingPolicy())...)
	}

	return policies, nil
}

func UpdatePolicy(oldPolicy, newPolicy []string, casdoorAdapter *CasdoorAdapter) (bool, error) {
	modelObj, err := getModel(casdoorAdapter.Owner, casdoorAdapter.Model)
	if err != nil {
		return false, err
	}

	enforcer, err := initEnforcer(modelObj, casdoorAdapter)
	if err != nil {
		return false, err
	}

	affected, err := enforcer.UpdatePolicy(oldPolicy, newPolicy)
	if err != nil {
		return affected, err
	}
	return affected, nil
}

func AddPolicy(policy []string, casdoorAdapter *CasdoorAdapter) (bool, error) {
	modelObj, err := getModel(casdoorAdapter.Owner, casdoorAdapter.Model)
	if err != nil {
		return false, err
	}

	enforcer, err := initEnforcer(modelObj, casdoorAdapter)
	if err != nil {
		return false, err
	}

	affected, err := enforcer.AddPolicy(policy)
	if err != nil {
		return affected, err
	}
	return affected, nil
}

func RemovePolicy(policy []string, casdoorAdapter *CasdoorAdapter) (bool, error) {
	modelObj, err := getModel(casdoorAdapter.Owner, casdoorAdapter.Model)
	if err != nil {
		return false, err
	}

	enforcer, err := initEnforcer(modelObj, casdoorAdapter)
	if err != nil {
		return false, err
	}

	affected, err := enforcer.RemovePolicy(policy)
	if err != nil {
		return affected, err
	}

	return affected, nil
}
