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
	xormadapter "github.com/casbin/xorm-adapter/v2"
	"github.com/casdoor/casdoor/util"

	"xorm.io/core"
)

type CasbinAdapter struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	Type         string `xorm:"varchar(100)" json:"type"`
	Model        string `xorm:"varchar(100)" json:"model"`

	Host         string `xorm:"varchar(100)" json:"host"`
	Port         int    `json:"port"`
	User         string `xorm:"varchar(100)" json:"user"`
	Password     string `xorm:"varchar(100)" json:"password"`
	DatabaseType string `xorm:"varchar(100)" json:"databaseType"`
	Database     string `xorm:"varchar(100)" json:"database"`
	Table        string `xorm:"varchar(100)" json:"table"`
	IsEnabled    bool   `json:"isEnabled"`

	Adapter *xormadapter.Adapter `xorm:"-" json:"-"`
}

func GetCasbinAdapterCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&CasbinAdapter{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetCasbinAdapters(owner string) []*CasbinAdapter {
	adapters := []*CasbinAdapter{}
	err := adapter.Engine.Where("owner = ?", owner).Find(&adapters)
	if err != nil {
		panic(err)
	}

	return adapters
}

func GetPaginationCasbinAdapters(owner string, page, limit int, field, value, sort, order string) []*CasbinAdapter {
	session := GetSession(owner, page, limit, field, value, sort, order)
	adapters := []*CasbinAdapter{}
	err := session.Find(&adapters)
	if err != nil {
		panic(err)
	}

	return adapters
}

func getCasbinAdapter(owner, name string) *CasbinAdapter {
	if owner == "" || name == "" {
		return nil
	}

	casbinAdapter := CasbinAdapter{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&casbinAdapter)
	if err != nil {
		panic(err)
	}

	if existed {
		return &casbinAdapter
	} else {
		return nil
	}
}

func GetCasbinAdapter(id string) *CasbinAdapter {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getCasbinAdapter(owner, name)
}

func GetMasterCasbinAdapter(casbinAdapter *CasbinAdapter) *CasbinAdapter {
	if casbinAdapter == nil {
		return nil
	}

	if casbinAdapter.Password != "" {
		casbinAdapter.Password = "***"
	}

	return casbinAdapter
}

func UpdateCasbinAdapter(id string, casbinAdapter *CasbinAdapter) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getCasbinAdapter(owner, name) == nil {
		return false
	}

	session := adapter.Engine.ID(core.PK{owner, name}).AllCols()
	if casbinAdapter.Password == "***" {
		session.Omit("password")
	}
	affected, err := session.Update(casbinAdapter)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddCasbinAdapter(casbinAdapter *CasbinAdapter) bool {
	affected, err := adapter.Engine.Insert(casbinAdapter)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteCasbinAdapter(casbinAdapter *CasbinAdapter) bool {
	affected, err := adapter.Engine.ID(core.PK{casbinAdapter.Owner, casbinAdapter.Name}).Delete(&CasbinAdapter{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (casbinAdapter *CasbinAdapter) GetId() string {
	return fmt.Sprintf("%s/%s", casbinAdapter.Owner, casbinAdapter.Name)
}

func (casbinAdapter *CasbinAdapter) getTable() string {
	if casbinAdapter.DatabaseType == "mssql" {
		return fmt.Sprintf("[%s]", casbinAdapter.Table)
	} else {
		return casbinAdapter.Table
	}
}

func safeReturn(policy []string, i int) string {
	if len(policy) > i {
		return policy[i]
	} else {
		return ""
	}
}

func matrixToCasbinRules(pType string, policies [][]string) []*xormadapter.CasbinRule {
	res := []*xormadapter.CasbinRule{}

	for _, policy := range policies {
		line := xormadapter.CasbinRule{
			PType: pType,
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

func SyncPolicies(casbinAdapter *CasbinAdapter) []*xormadapter.CasbinRule {
	// init Adapter
	if casbinAdapter.Adapter == nil {
		var dataSourceName string
		if casbinAdapter.DatabaseType == "mssql" {
			dataSourceName = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", casbinAdapter.User, casbinAdapter.Password, casbinAdapter.Host, casbinAdapter.Port, casbinAdapter.Database)
		} else if casbinAdapter.DatabaseType == "postgres" {
			dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=disable dbname=%s", casbinAdapter.User, casbinAdapter.Password, casbinAdapter.Host, casbinAdapter.Port, casbinAdapter.Database)
		} else {
			dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/", casbinAdapter.User, casbinAdapter.Password, casbinAdapter.Host, casbinAdapter.Port)
		}

		if !isCloudIntranet {
			dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
		}

		casbinAdapter.Adapter, _ = xormadapter.NewAdapterByEngineWithTableName(NewAdapter(casbinAdapter.DatabaseType, dataSourceName, casbinAdapter.Database).Engine, casbinAdapter.getTable(), "")
	}

	// init Model
	modelObj := getModel(casbinAdapter.Owner, casbinAdapter.Model)
	m, err := model.NewModelFromString(modelObj.ModelText)
	if err != nil {
		panic(err)
	}

	// init Enforcer
	enforcer, err := casbin.NewEnforcer(m, casbinAdapter.Adapter)
	if err != nil {
		panic(err)
	}

	return matrixToCasbinRules("p", enforcer.GetPolicy())
}
