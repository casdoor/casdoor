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

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
)

type Adapter struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Table        string `xorm:"varchar(100)" json:"table"`
	UseSameDb    bool   `json:"useSameDb"`
	Type         string `xorm:"varchar(100)" json:"type"`
	DatabaseType string `xorm:"varchar(100)" json:"databaseType"`
	Host         string `xorm:"varchar(100)" json:"host"`
	Port         int    `json:"port"`
	User         string `xorm:"varchar(100)" json:"user"`
	Password     string `xorm:"varchar(100)" json:"password"`
	Database     string `xorm:"varchar(100)" json:"database"`

	*xormadapter.Adapter `xorm:"-" json:"-"`
}

func GetAdapterCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Adapter{})
}

func GetAdapters(owner string) ([]*Adapter, error) {
	adapters := []*Adapter{}
	err := ormer.Engine.Desc("created_time").Find(&adapters, &Adapter{Owner: owner})
	if err != nil {
		return adapters, err
	}

	return adapters, nil
}

func GetPaginationAdapters(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Adapter, error) {
	adapters := []*Adapter{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&adapters)
	if err != nil {
		return adapters, err
	}

	return adapters, nil
}

func getAdapter(owner, name string) (*Adapter, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	adapter := Adapter{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&adapter)
	if err != nil {
		return nil, err
	}

	if existed {
		return &adapter, nil
	} else {
		return nil, nil
	}
}

func GetAdapter(id string) (*Adapter, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getAdapter(owner, name)
}

func UpdateAdapter(id string, adapter *Adapter) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if adapter, err := getAdapter(owner, name); adapter == nil {
		return false, err
	}

	if name != adapter.Name {
		err := adapterChangeTrigger(name, adapter.Name)
		if err != nil {
			return false, err
		}
	}

	session := ormer.Engine.ID(core.PK{owner, name}).AllCols()
	if adapter.Password == "***" {
		session.Omit("password")
	}
	affected, err := session.Update(adapter)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddAdapter(adapter *Adapter) (bool, error) {
	affected, err := ormer.Engine.Insert(adapter)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteAdapter(adapter *Adapter) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{adapter.Owner, adapter.Name}).Delete(&Adapter{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (adapter *Adapter) GetId() string {
	return fmt.Sprintf("%s/%s", adapter.Owner, adapter.Name)
}

func (adapter *Adapter) InitAdapter() error {
	if adapter.Adapter != nil {
		return nil
	}

	var driverName string
	var dataSourceName string
	if adapter.UseSameDb || adapter.isBuiltIn() {
		driverName = conf.GetConfigString("driverName")
		dataSourceName = conf.GetConfigString("dataSourceName")
		if conf.GetConfigString("driverName") == "mysql" {
			dataSourceName = dataSourceName + conf.GetConfigString("dbName")
		}
	} else {
		driverName = adapter.DatabaseType
		switch driverName {
		case "mssql":
			dataSourceName = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", adapter.User,
				adapter.Password, adapter.Host, adapter.Port, adapter.Database)
		case "mysql":
			dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", adapter.User,
				adapter.Password, adapter.Host, adapter.Port, adapter.Database)
		case "postgres":
			dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=disable dbname=%s", adapter.User,
				adapter.Password, adapter.Host, adapter.Port, adapter.Database)
		case "CockroachDB":
			dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=disable dbname=%s serial_normalization=virtual_sequence",
				adapter.User, adapter.Password, adapter.Host, adapter.Port, adapter.Database)
		case "sqlite3":
			dataSourceName = fmt.Sprintf("file:%s", adapter.Host)
		default:
			return fmt.Errorf("unsupported database type: %s", adapter.DatabaseType)
		}
	}

	if !isCloudIntranet {
		dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
	}

	engine, err := xorm.NewEngine(driverName, dataSourceName)
	if err != nil {
		return err
	}

	if (adapter.UseSameDb || adapter.isBuiltIn()) && driverName == "postgres" {
		schema := util.GetValueFromDataSourceName("search_path", dataSourceName)
		if schema != "" {
			engine.SetSchema(schema)
		}
	}

	var tableName string
	if driverName == "mssql" {
		tableName = fmt.Sprintf("[%s]", adapter.Table)
	} else {
		tableName = adapter.Table
	}

	adapter.Adapter, err = xormadapter.NewAdapterByEngineWithTableName(engine, tableName, "")
	if err != nil {
		return err
	}

	return nil
}

func adapterChangeTrigger(oldName string, newName string) error {
	session := ormer.Engine.NewSession()
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

func (adapter *Adapter) isBuiltIn() bool {
	if adapter.Owner != "built-in" {
		return false
	}

	return adapter.Name == "user-adapter-built-in" || adapter.Name == "api-adapter-built-in"
}
