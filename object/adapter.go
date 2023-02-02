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
	"errors"
	"fmt"
	"runtime"

	"github.com/beego/beego"
	xormadapter "github.com/casbin/xorm-adapter/v3"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	_ "github.com/denisenkom/go-mssqldb" // db = mssql
	_ "github.com/go-sql-driver/mysql"   // db = mysql
	_ "github.com/lib/pq"                // db = postgres
	_ "modernc.org/sqlite"               // db = sqlite
	"xorm.io/core"
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
)

var adapter *Adapter

func InitConfig() {
	err := beego.LoadAppConfig("ini", "../conf/app.conf")
	if err != nil {
		panic(err)
	}

	beego.BConfig.WebConfig.Session.SessionOn = true

	InitAdapter(true)
	initMigrations()
}

func InitAdapter(createDatabase bool) {
	adapter = NewAdapter(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
	if createDatabase {
		adapter.CreateDatabase()
	}
	adapter.createTable()
}

// Adapter represents the MySQL adapter for policy storage.
type Adapter struct {
	driverName     string
	dataSourceName string
	dbName         string
	Engine         *xorm.Engine
}

// finalizer is the destructor for Adapter.
func finalizer(a *Adapter) {
	err := a.Engine.Close()
	if err != nil {
		panic(err)
	}
}

// NewAdapter is the constructor for Adapter.
func NewAdapter(driverName string, dataSourceName string, dbName string) *Adapter {
	a := &Adapter{}
	a.driverName = driverName
	a.dataSourceName = dataSourceName
	a.dbName = dbName

	// Open the DB, create it if not existed.
	a.open()

	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, finalizer)

	return a
}

func (a *Adapter) CreateDatabase() error {
	engine, err := xorm.NewEngine(a.driverName, a.dataSourceName)
	if err != nil {
		return err
	}
	defer engine.Close()

	_, err = engine.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s default charset utf8mb4 COLLATE utf8mb4_general_ci", a.dbName))
	return err
}

func (a *Adapter) open() {
	dataSourceName := a.dataSourceName + a.dbName
	if a.driverName != "mysql" {
		dataSourceName = a.dataSourceName
	}

	engine, err := xorm.NewEngine(a.driverName, dataSourceName)
	if err != nil {
		panic(err)
	}

	a.Engine = engine
}

func (a *Adapter) close() {
	_ = a.Engine.Close()
	a.Engine = nil
}

func (a *Adapter) createTable() {
	showSql, _ := conf.GetConfigBool("showSql")
	a.Engine.ShowSQL(showSql)

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, tableNamePrefix)
	a.Engine.SetTableMapper(tbMapper)

	err := a.Engine.Sync2(new(Organization))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(User))
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

	err = a.Engine.Sync2(new(CasbinAdapter))
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

	err = a.Engine.Sync2(new(Record))
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

	err = a.Engine.Sync2(new(PermissionRule))
	if err != nil {
		panic(err)
	}

	err = a.Engine.Sync2(new(xormadapter.CasbinRule))
	if err != nil {
		panic(err)
	}

	err = a.syncSession()
	if err != nil {
		panic(err)
	}
}

func (a *Adapter) syncSession() error {
	if exist, _ := a.Engine.IsTableExist("session"); exist {
		if colErr := a.Engine.Find(&[]*Session{}); colErr != nil {
			// Create a new field called 'application' and add it to the primary key for table `session`
			var err error
			tx := a.Engine.NewSession()

			if alreadyCreated, _ := a.Engine.IsTableExist("session_tmp"); alreadyCreated {
				panic(errors.New("there is already a table called 'session_tmp', please rename or delete it for casdoor version migration and restart"))
			}

			tx.Table("session_tmp").CreateTable(&Session{})

			type oldSession struct {
				Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
				Name        string `xorm:"varchar(100) notnull pk" json:"name"`
				CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

				SessionId []string `json:"sessionId"`
			}

			oldSessions := []*oldSession{}
			newSessions := []*Session{}

			tx.Table("session").Find(&oldSessions)

			for _, oldSession := range oldSessions {
				newApplication := "null"
				if oldSession.Owner == "built-in" {
					newApplication = "app-built-in"
				}
				newSessions = append(newSessions, &Session{
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

			delete := &Session{
				Application: "null",
			}
			_, err = tx.Table("session_tmp").Delete(*delete)
			if err != nil {
				rollbackFlag = true
			}

			if rollbackFlag {
				tx.DropTable("session_tmp")
				panic(errors.New("there is something wrong with data migration for table `session`, if there is a table called `session_tmp` not created by you in casdoor, please drop it, then restart anyhow"))
			}

			err = tx.DropTable("session")
			if err != nil {
				panic(errors.New("fail to drop table `session` for casdoor, please drop it and rename the table `session_tmp` to `session` manually and restart"))
			}

			// Already drop table `session`
			// Can't find an api from xorm for altering table name
			err = tx.Table("session").CreateTable(&Session{})
			if err != nil {
				panic(errors.New("there is something wrong with data migration for table `session`, please restart"))
			}

			sessions := []*Session{}
			tx.Table("session_tmp").Find(&sessions)
			_, err = tx.Table("session").Insert(sessions)
			if err != nil {
				panic(errors.New("there is something wrong with data migration for table `session`, please drop table `session` and rename table `session_tmp` to `session` and restart"))
			}

			err = tx.DropTable("session_tmp")
			if err != nil {
				panic(errors.New("fail to drop table `session_tmp` for casdoor, please drop it manually and restart"))
			}

			tx.Close()
		}
	}
	return a.Engine.Sync2(new(Session))
}

func GetSession(owner string, offset, limit int, field, value, sortField, sortOrder string) *xorm.Session {
	session := adapter.Engine.Prepare()
	if offset != -1 && limit != -1 {
		session.Limit(limit, offset)
	}
	if owner != "" {
		session = session.And("owner=?", owner)
	}
	if field != "" && value != "" {
		if filterField(field) {
			session = session.And(fmt.Sprintf("%s like ?", util.SnakeString(field)), fmt.Sprintf("%%%s%%", value))
		}
	}
	if sortField == "" || sortOrder == "" {
		sortField = "created_time"
	}
	if sortOrder == "ascend" {
		session = session.Asc(util.SnakeString(sortField))
	} else {
		session = session.Desc(util.SnakeString(sortField))
	}
	return session
}

func initMigrations() {
	migrations := []*migrate.Migration{
		{
			ID: "20221015CasbinRule--fill ptype field with p",
			Migrate: func(tx *xorm.Engine) error {
				_, err := tx.Cols("ptype").Update(&xormadapter.CasbinRule{
					Ptype: "p",
				})
				return err
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(&xormadapter.CasbinRule{})
			},
		},
	}
	m := migrate.New(adapter.Engine, migrate.DefaultOptions, migrations)
	m.Migrate()
}
