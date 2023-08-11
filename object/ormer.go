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
	"fmt"
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

var ormer *Ormer

type DatabaseConfig struct {
	driverName      string
	host            string
	port            int
	user            string
	password        string
	database        string
	table           string
	tableNamePrefix string
}

// Ormer represents the MySQL adapter for policy storage.
type Ormer struct {
	*DatabaseConfig
	dataSourceName string

	Engine *xorm.Engine `xorm:"-" json:"-"`
}

func InitConfig() {
	err := beego.LoadAppConfig("ini", "../conf/app.conf")
	if err != nil {
		panic(err)
	}

	beego.BConfig.WebConfig.Session.SessionOn = true

	InitOrmer(true)
	CreateTables(true)
	DoMigration()
}

func InitOrmer(createDatabase bool) {
	if createDatabase {
		err := createDatabaseForPostgres(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
		if err != nil {
			panic(err)
		}
	}

	ormer = NewOrmer(nil, true)

	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, ormer.tableNamePrefix)
	ormer.Engine.SetTableMapper(tbMapper)
}

func CreateTables(createDatabase bool) {
	if createDatabase {
		err := ormer.CreateDatabase()
		if err != nil {
			panic(err)
		}
	}

	ormer.createTable()
}

// NewOrmer is the constructor for Ormer.
func NewOrmer(config *DatabaseConfig, isCasdoorSelf ...bool) *Ormer {
	if config == nil && len(isCasdoorSelf) == 0 {
		panic(fmt.Errorf("database config is nil, please check your database config"))
	}

	o := &Ormer{
		DatabaseConfig: config,
	}

	if len(isCasdoorSelf) != 0 && isCasdoorSelf[0] == true {
		// Get the dataSourceName from the app.conf.
		o.dataSourceName = conf.GetConfigDataSourceName()

		o.DatabaseConfig = &DatabaseConfig{
			driverName:      conf.GetConfigString("driverName"),
			database:        conf.GetConfigString("dbName"),
			tableNamePrefix: conf.GetConfigString("tableNamePrefix"),
		}
	} else {
		// Concat the dataSourceName from config of adapter.
		var dataSourceName string
		switch o.driverName {
		case "mssql":
			dataSourceName = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", o.user,
				o.password, o.host, o.port, o.database)
		case "mysql":
			dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/", o.user,
				o.password, o.host, o.port)
		case "postgres":
			dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=disable dbname=%s", o.user,
				o.password, o.host, o.port, o.database)
		case "CockroachDB":
			dataSourceName = fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=disable dbname=%s serial_normalization=virtual_sequence",
				o.user, o.password, o.host, o.port, o.database)
		case "sqlite3":
			dataSourceName = fmt.Sprintf("file:%s", o.host)
		default:
			panic(fmt.Errorf("unsupport driver name: %s", o.driverName))
		}

		if !isCloudIntranet {
			dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
		}
		o.dataSourceName = dataSourceName
	}

	// Open the DB, create it if not existed.
	o.open()

	return o
}

func createDatabaseForPostgres(databaseType string, dataSourceName string, dbName string) error {
	if databaseType == "postgres" {
		db, err := sql.Open(databaseType, dataSourceName)
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

	_, err = engine.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s default charset utf8mb4 COLLATE utf8mb4_general_ci", a.database))
	return err
}

func (a *Ormer) open() {
	// We create database of mysql for docker scenario, it needs to connect to the server without database name.
	// So split the database name and dataSourceName of mysql. We need to concat the dataSourceName and database name when we really connect to the database.
	dataSourceName := a.dataSourceName
	if a.driverName == "mysql" {
		dataSourceName = dataSourceName + a.database
	}

	engine, err := xorm.NewEngine(a.driverName, dataSourceName)
	if err != nil {
		panic(err)
	}

	a.Engine = engine
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

func GetSession(owner string, offset, limit int, field, value, sortField, sortOrder string) *xorm.Session {
	session := ormer.Engine.Prepare()
	if offset != -1 && limit != -1 {
		session.Limit(limit, offset)
	}
	if owner != "" {
		session = session.And("owner=?", owner)
	}
	if field != "" && value != "" {
		if util.FilterField(field) {
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

func GetSessionForUser(owner string, offset, limit int, field, value, sortField, sortOrder string) *xorm.Session {
	session := ormer.Engine.Prepare()
	if offset != -1 && limit != -1 {
		session.Limit(limit, offset)
	}
	if owner != "" {
		if offset == -1 {
			session = session.And("owner=?", owner)
		} else {
			session = session.And("a.owner=?", owner)
		}
	}
	if field != "" && value != "" {
		if util.FilterField(field) {
			if offset != -1 {
				field = fmt.Sprintf("a.%s", field)
			}
			session = session.And(fmt.Sprintf("%s like ?", util.SnakeString(field)), fmt.Sprintf("%%%s%%", value))
		}
	}
	if sortField == "" || sortOrder == "" {
		sortField = "created_time"
	}

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	tableName := tableNamePrefix + "user"
	if offset == -1 {
		if sortOrder == "ascend" {
			session = session.Asc(util.SnakeString(sortField))
		} else {
			session = session.Desc(util.SnakeString(sortField))
		}
	} else {
		if sortOrder == "ascend" {
			session = session.Alias("a").
				Join("INNER", []string{tableName, "b"}, "a.owner = b.owner and a.name = b.name").
				Select("b.*").
				Asc("a." + util.SnakeString(sortField))
		} else {
			session = session.Alias("a").
				Join("INNER", []string{tableName, "b"}, "a.owner = b.owner and a.name = b.name").
				Select("b.*").
				Desc("a." + util.SnakeString(sortField))
		}
	}

	return session
}
