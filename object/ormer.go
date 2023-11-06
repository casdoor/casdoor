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
	"database/sql"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	_ "github.com/denisenkom/go-mssqldb" // db = mssql
	_ "github.com/go-sql-driver/mysql"   // db = mysql
	_ "github.com/lib/pq"                // db = postgres
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
	_ "modernc.org/sqlite" // db = sqlite
)

var (
	ormer                   *Ormer = nil
	isCreateDatabaseDefined        = false
	createDatabase                 = true
)

func InitFlag() {
	if !isCreateDatabaseDefined {
		isCreateDatabaseDefined = true
		createDatabase = getCreateDatabaseFlag()
	}
}

func getCreateDatabaseFlag() bool {
	res := flag.Bool("createDatabase", false, "true if you need to create database")
	flag.Parse()
	return *res
}

func InitConfig() {
	err := beego.LoadAppConfig("ini", "../conf/app.conf")
	if err != nil {
		panic(err)
	}

	beego.BConfig.WebConfig.Session.SessionOn = true

	InitAdapter()
	CreateTables()
}

func InitAdapter() {
	if conf.GetConfigString("driverName") == "" {
		if !util.FileExist("conf/app.conf") {
			dir, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			dir = strings.ReplaceAll(dir, "\\", "/")
			panic(fmt.Sprintf("The Casdoor config file: \"app.conf\" was not found, it should be placed at: \"%s/conf/app.conf\"", dir))
		}
	}

	if createDatabase {
		err := createDatabaseForPostgres(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
		if err != nil {
			panic(err)
		}
	}

	var err error
	ormer, err = NewAdapter(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
	if err != nil {
		panic(err)
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, tableNamePrefix)
	ormer.Engine.SetTableMapper(tbMapper)
}

func CreateTables() {
	if createDatabase {
		err := ormer.CreateDatabase()
		if err != nil {
			panic(err)
		}
	}

	ormer.createTable()
}

// Ormer represents the MySQL adapter for policy storage.
type Ormer struct {
	driverName     string
	dataSourceName string
	dbName         string
	Engine         *xorm.Engine
}

// finalizer is the destructor for Ormer.
func finalizer(a *Ormer) {
	err := a.Engine.Close()
	if err != nil {
		panic(err)
	}
}

// NewAdapter is the constructor for Ormer.
func NewAdapter(driverName string, dataSourceName string, dbName string) (*Ormer, error) {
	a := &Ormer{}
	a.driverName = driverName
	a.dataSourceName = dataSourceName
	a.dbName = dbName

	// Open the DB, create it if not existed.
	err := a.open()
	if err != nil {
		return nil, err
	}

	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, finalizer)

	return a, nil
}

func refineDataSourceNameForPostgres(dataSourceName string) string {
	reg := regexp.MustCompile(`dbname=[^ ]+\s*`)
	return reg.ReplaceAllString(dataSourceName, "")
}

func createDatabaseForPostgres(driverName string, dataSourceName string, dbName string) error {
	if driverName == "postgres" {
		db, err := sql.Open(driverName, refineDataSourceNameForPostgres(dataSourceName))
		if err != nil {
			return err
		}
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
		if err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
		schema := util.GetValueFromDataSourceName("search_path", dataSourceName)
		if schema != "" {
			db, err = sql.Open(driverName, dataSourceName)
			if err != nil {
				return err
			}
			defer db.Close()

			_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %s;", schema))
			if err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					return err
				}
			}
		}

		return nil
	} else {
		return nil
	}
}

func (a *Ormer) CreateDatabase() error {
	if a.driverName == "postgres" {
		return nil
	}

	engine, err := xorm.NewEngine(a.driverName, a.dataSourceName)
	if err != nil {
		return err
	}
	defer engine.Close()

	_, err = engine.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s default charset utf8mb4 COLLATE utf8mb4_general_ci", a.dbName))
	return err
}

func (a *Ormer) open() error {
	dataSourceName := a.dataSourceName + a.dbName
	if a.driverName != "mysql" {
		dataSourceName = a.dataSourceName
	}

	engine, err := xorm.NewEngine(a.driverName, dataSourceName)
	if err != nil {
		return err
	}

	if a.driverName == "postgres" {
		schema := util.GetValueFromDataSourceName("search_path", dataSourceName)
		if schema != "" {
			engine.SetSchema(schema)
		}
	}

	a.Engine = engine
	return nil
}

func (a *Ormer) close() {
	_ = a.Engine.Close()
	a.Engine = nil
}

func (a *Ormer) createTable() {
	showSql := conf.GetConfigBool("showSql")
	a.Engine.ShowSQL(showSql)

	err := a.Engine.Sync2(new(Organization))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(User))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Group))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Role))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Permission))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Model))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Adapter))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Enforcer))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Provider))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Application))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Resource))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Token))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(VerificationRecord))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Webhook))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Syncer))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Cert))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Product))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Payment))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Ldap))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(RadiusAccounting))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(xormadapter.CasbinRule))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Session))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Subscription))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Plan))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(Pricing))
	if err != nil {
		panic(err)
	}
}
