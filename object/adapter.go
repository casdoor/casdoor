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
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	_ "github.com/denisenkom/go-mssqldb" // db = mssql
	_ "github.com/go-sql-driver/mysql"   // db = mysql
	_ "github.com/lib/pq"                // db = postgres
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"xorm.io/core"
	"xorm.io/xorm"
	//_ "github.com/mattn/go-sqlite3"    // db = sqlite3
)

var adapter *Adapter

func InitConfig() {
	err := beego.LoadAppConfig("ini", "../conf/app.conf")
	if err != nil {
		panic(err)
	}

	beego.BConfig.WebConfig.Session.SessionOn = true

	InitAdapter(true)
}

func InitAdapter(createDatabase bool) {
	adapter = NewAdapter(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
	if createDatabase {
		adapter.CreateDatabase()
	}
	adapter.createTable()
}

type Adapter struct {
	driverName     string
	dataSourceName string
	dbName         string
	Engine         *xorm.Engine
	MongoEngine    *mongo.Database
}

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

func (a *Adapter) isMongo() bool {
	return a.driverName == "mongodb"
}

func (a *Adapter) CreateDatabase() error {
	if a.isMongo() {
		// If a database does not exist, MongoDB creates the database when you first store data for that database.
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

func (a *Adapter) open() {
	if a.isMongo() {
		c, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(a.dataSourceName))
		if err != nil {
			panic(err)
		}
		a.MongoEngine = c.Database(a.dbName)
	} else {
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
}

// finalizer is the destructor for Adapter.
func finalizer(a *Adapter) {
	err := a.close()
	if err != nil {
		panic(err)
	}
}

func (a *Adapter) close() error {
	if a.isMongo() {
		err := a.MongoEngine.Client().Disconnect(context.TODO())
		if err != nil {
			return err
		}
		a.MongoEngine = nil
	} else {
		err := a.Engine.Close()
		if err != nil {
			return err
		}
		a.Engine = nil
	}
	return nil
}

func (a *Adapter) createTable() {
	if a.isMongo() {
		// If a collection does not exist, MongoDB creates the collection when you first store data for that collection.
		return
	}
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
}

func (a *Adapter) GetSession(c *Conditions) *Session {
	s := new(Session)
	s.adapter = a
	s.conditions = c
	return s
}

type Session struct {
	adapter    *Adapter
	conditions *Conditions
}

type Conditions struct {
	Limit     int
	Offset    int
	SortField string
	SortOrder string

	Field string // Field 为数据库字段名，SnakeString
	Value string
	Owner string
}

func (s *Session) getXormConditions() (xormSession *xorm.Session) {
	xormSession = s.adapter.Engine.Prepare()
	if s.conditions == nil {
		return xormSession
	}
	if s.conditions.Offset != -1 && s.conditions.Limit != -1 {
		xormSession.Limit(s.conditions.Limit, s.conditions.Offset)
	}
	if s.conditions.Owner != "" {
		xormSession = xormSession.And("owner=?", s.conditions.Owner)
	}
	if s.conditions.Field != "" && s.conditions.Value != "" {
		if filterField(s.conditions.Field) {
			xormSession = xormSession.And(fmt.Sprintf("%s like ?", util.SnakeString(s.conditions.Field)), fmt.Sprintf("%%%s%%", s.conditions.Value))
		}
	}
	var sortField string
	if s.conditions.SortField == "" || s.conditions.SortOrder == "" {
		sortField = "created_time"
	}
	if s.conditions.SortOrder == "ascend" {
		xormSession = xormSession.Asc(util.SnakeString(sortField))
	} else {
		xormSession = xormSession.Desc(util.SnakeString(sortField))
	}
	return
}

func (s *Session) getMongoFindConditions() (filter bson.D, opts *options.FindOptions) {
	if s.conditions == nil {
		return nil, nil
	}
	opts = options.Find()

	if s.conditions.Owner != "" {
		filter = bson.D{{Key: "owner", Value: s.conditions.Owner}}
	}
	if s.conditions.Field != "" && s.conditions.Value != "" {
		if filterField(s.conditions.Field) {
			filter = append(filter, bson.E{Key: util.SnakeString(s.conditions.Field), Value: s.conditions.Value})
		}
	}

	if s.conditions.Offset != -1 && s.conditions.Limit != -1 {
		opts.SetLimit(int64(s.conditions.Limit))
		opts.SetSkip(int64(s.conditions.Offset))
	}
	if s.conditions.SortField != "" && s.conditions.SortOrder != "" {
		if s.conditions.SortOrder == "asc" {
			opts.SetSort(bson.D{{Key: s.conditions.SortField, Value: 1}})
		} else if s.conditions.SortOrder == "desc" {
			opts.SetSort(bson.D{{Key: s.conditions.SortField, Value: -1}})
		} else {
			panic("invalid sortOrder")
		}
	} else {
		opts.SetSort(bson.D{{Key: "created_time", Value: 1}})
	}

	return
}

func getCollectionName(objectPtr interface{}) (name string) {
	e := reflect.ValueOf(objectPtr)
	if e.Kind() == reflect.Slice {
		e = e.Index(0).Elem()
	} else {
		e = e.Elem()
	}
	return util.SnakeString(e.Type().Name())
}

func (s *Session) Find(ctx context.Context, rowsSlicePtr interface{}, condiBean ...interface{}) error {
	if s.adapter.isMongo() {
		// TODO: Parse condiBean
		collName := getCollectionName(rowsSlicePtr)
		filter, opts := s.getMongoFindConditions()
		cur, err := s.adapter.MongoEngine.Collection(collName).Find(ctx, filter, opts)
		if err != nil {
			return err
		}
		return cur.All(ctx, rowsSlicePtr)
	} else {
		xormSession := s.getXormConditions()
		return xormSession.Find(rowsSlicePtr, condiBean...)
	}
}
