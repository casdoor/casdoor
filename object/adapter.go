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
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/beego/beego"
	xormadapter "github.com/casbin/xorm-adapter/v3"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	_ "github.com/denisenkom/go-mssqldb" // db = mssql
	_ "github.com/go-sql-driver/mysql"   // db = mysql
	_ "github.com/lib/pq"                // db = postgres
	"xorm.io/xorm/migrate"

	//_ "github.com/mattn/go-sqlite3"    // db = sqlite3
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"xorm.io/core"
	"xorm.io/xorm"
)

var adapter *Adapter

func InitConfig() {
	err := beego.LoadAppConfig("ini", "../conf/app.conf")
	if err != nil {
		panic(err)
	}

	beego.BConfig.WebConfig.Session.SessionOn = true

	InitAdapter(true)
	if adapter.isMongo() {
		// TODO: implement MongoDB migrations
		return
	} else {
		initMigrations()
	}
}

func InitAdapter(createDatabase bool) {
	adapter = NewAdapter(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName(), conf.GetConfigString("dbName"))
	if createDatabase {
		adapter.CreateDatabase()
	}
	adapter.createTable()
}

// Adapter represents the database adapter for policy storage.
type Adapter struct {
	driverName     string
	dataSourceName string
	dbName         string
	Engine         *xorm.Engine
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
		err := mgm.SetDefaultConfig(nil, a.dbName, options.Client().ApplyURI(a.dataSourceName))
		if err != nil {
			panic(err)
		}
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
	if !a.isMongo() {
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

	err = a.Engine.Sync2(new(xormadapter.CasbinRule))
	if err != nil {
		panic(err)
	}
}

type Conditions struct {
	limit     int
	offset    int
	sortField string
	sortOrder string

	equalFields EqualFields
	likeFields  LikeFields
	owner       string
}

func ConditionsBuilder() *Conditions {
	return &Conditions{
		limit:  -1,
		offset: -1,
	}
}

func (c *Conditions) SetLimit(limit int) *Conditions {
	c.limit = limit
	return c
}

func (c *Conditions) SetOffset(offset int) *Conditions {
	c.offset = offset
	return c
}

func (c *Conditions) SetSortField(sortField string) *Conditions {
	c.sortField = sortField
	return c
}

func (c *Conditions) SetSortOrder(sortOrder string) *Conditions {
	c.sortOrder = sortOrder
	return c
}

func (c *Conditions) SetSortOrderASC() *Conditions {
	c.sortOrder = "ascend"
	return c
}

func (c *Conditions) SetSortOrderDESC() *Conditions {
	c.sortOrder = "descend"
	return c
}

func (c *Conditions) SetOwner(owner string) *Conditions {
	c.owner = owner
	return c
}

type EqualFields map[string]interface{}

func (c *Conditions) SetEqualFields(equalFields EqualFields) *Conditions {
	c.equalFields = equalFields
	return c
}

type LikeFields map[string]string

func (c *Conditions) SetLikeFields(likeFields LikeFields) *Conditions {
	c.likeFields = likeFields
	return c
}

type Session struct {
	adapter    *Adapter
	conditions *Conditions
	ctx        context.Context
}

// open a db session
func (a *Adapter) CreateSession(ctx context.Context, c *Conditions) *Session {
	s := new(Session)
	s.adapter = a
	s.conditions = c
	s.ctx = ctx
	return s
}

// TODO: will remove
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

func (s *Session) useXormConditions() (xormSession *xorm.Session) {
	xormSession = s.adapter.Engine.Context(s.ctx)
	if s.conditions == nil {
		return xormSession
	}
	if s.conditions.offset != -1 && s.conditions.limit != -1 {
		xormSession.Limit(s.conditions.limit, s.conditions.offset)
	}
	if s.conditions.owner != "" {
		xormSession = xormSession.And("owner=?", s.conditions.owner)
	}
	if s.conditions.equalFields != nil {
		for k, v := range s.conditions.equalFields {
			if k == "" {
				continue
			}
			xormSession = xormSession.And(fmt.Sprintf("%s=?", util.SnakeString(k)), v)
		}
	}
	if s.conditions.likeFields != nil {
		for k, v := range s.conditions.likeFields {
			if k == "" {
				continue
			}
			xormSession = xormSession.And(fmt.Sprintf("%s like ?", util.SnakeString(k)), fmt.Sprintf("%%%s%%", v))
		}
	}
	if s.conditions.sortField != "" && s.conditions.sortOrder != "" {
		if s.conditions.sortOrder == "ascend" {
			xormSession = xormSession.Asc(util.SnakeString(s.conditions.sortField))
		} else if s.conditions.sortOrder == "descend" {
			xormSession = xormSession.Desc(util.SnakeString(s.conditions.sortField))
		} else {
			panic("invalid sortOrder")
		}
	} else {
		xormSession = xormSession.Asc("created_time")
	}

	return
}

func (s *Session) useMongoConditions() (bson.M, *options.FindOptions) {
	if s.conditions == nil {
		return nil, nil
	}
	opts := options.Find()

	filter := bson.D{}
	if s.conditions.owner != "" {
		filter = append(filter, bson.E{Key: "owner", Value: s.conditions.owner})
	}

	if s.conditions.likeFields != nil {
		for k, v := range s.conditions.likeFields {
			if filterField(k) {
				filter = append(filter, bson.E{Key: util.SnakeString(k), Value: bson.D{{Key: "$regex", Value: fmt.Sprintf("/%s/", v)}}})
			}
		}
	}

	if s.conditions.equalFields != nil {
		for k, v := range s.conditions.equalFields {
			if filterField(k) {
				filter = append(filter, bson.E{Key: util.SnakeString(k), Value: v})
			}
		}
	}

	if s.conditions.offset != -1 && s.conditions.limit != -1 {
		opts.SetLimit(int64(s.conditions.limit))
		opts.SetSkip(int64(s.conditions.offset))
	}
	if s.conditions.sortField != "" && s.conditions.sortOrder != "" {
		if s.conditions.sortOrder == "ascend" {
			opts.SetSort(bson.D{{Key: s.conditions.sortField, Value: 1}})
		} else if s.conditions.sortOrder == "descend" {
			opts.SetSort(bson.D{{Key: s.conditions.sortField, Value: -1}})
		} else {
			panic("invalid sortOrder")
		}
	} else {
		opts.SetSort(bson.D{{Key: "created_time", Value: 1}})
	}

	return filter.Map(), opts
}

func getCollectionName(objectPtr interface{}) (name string) {
	v := reflect.ValueOf(objectPtr).Elem()
	t := v.Type()
	if t.Kind() == reflect.Slice {
		t = t.Elem() // slice element's type is ptr
		t = t.Elem() // ptr element's type is obj
	}
	return util.SnakeString(t.Name())
}

// Insert one or more objects
func (s *Session) Insert(objectPtr interface{}) (int64, error) {
	if s.adapter.isMongo() {
		coll := mgm.CollectionByName(getCollectionName(objectPtr))
		value := reflect.ValueOf(objectPtr).Elem()
		if value.Kind() == reflect.Slice {
			// insert object slice
			list := make([]interface{}, 0, value.Len())
			for i := 0; i < value.Len(); i++ {
				list = append(list, value.Index(i).Interface())
			}
			res, err := coll.InsertMany(s.ctx, list)
			if err != nil {
				return 0, err
			}
			return int64(len(res.InsertedIDs)), nil
		} else {
			// insert one object
			_, err := coll.InsertOne(s.ctx, objectPtr)
			if err != nil {
				return 0, err
			}
			return 1, nil
		}
	} else {
		return s.adapter.Engine.Context(s.ctx).Insert(objectPtr)
	}
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

func (s *Session) Get(objectPtr interface{}) (bool, error) {
	if s.adapter.isMongo() {
		filter, _ := s.useMongoConditions()
		coll := mgm.CollectionByName(getCollectionName(objectPtr))
		err := coll.FindOne(s.ctx, filter).Decode(objectPtr)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return false, nil
			} else {
				return false, err
			}
		}
		return true, nil
	} else {
		xormSession := s.useXormConditions()
		return xormSession.Get(objectPtr)
	}
}

func (s *Session) Find(objectSlicePtr interface{}) error {
	if s.adapter.isMongo() {
		filter, _ := s.useMongoConditions()
		name := getCollectionName(objectSlicePtr)
		coll := mgm.CollectionByName(name)
		return coll.SimpleFindWithCtx(s.ctx, objectSlicePtr, filter)
		// TODO: Parse condiBean
	} else {
		xormSession := s.useXormConditions()
		return xormSession.Find(objectSlicePtr)
	}
}

func (s *Session) Count(objectPtr interface{}) (int64, error) {
	if s.adapter.isMongo() {
		filter, _ := s.useMongoConditions()
		coll := mgm.CollectionByName(getCollectionName(objectPtr))
		return coll.CountDocuments(s.ctx, filter)
	} else {
		session := s.useXormConditions()
		return session.Count(getElemOfPointer(objectPtr))
	}
}

func (s *Session) DeleteByID(objectPtr interface{}, owner, name string) (int64, error) {
	if s.adapter.isMongo() {
		filter := bson.M{}
		filter["owner"] = owner
		filter["name"] = name
		coll := mgm.CollectionByName(getCollectionName(objectPtr))
		result, err := coll.DeleteMany(s.ctx, filter)
		return result.DeletedCount, err
	} else {
		return s.adapter.Engine.Context(s.ctx).ID(core.PK{owner, name}).Delete(getElemOfPointer(objectPtr))
	}
}

// objectPtr's non-empty fields are updated contents
func (s *Session) UpdateByID(objectPtr interface{}, owner string, name string, omitFields ...string) (int64, error) {
	if s.adapter.isMongo() {
		filter := bson.D{{Key: "owner", Value: owner}, {Key: "name", Value: name}}.Map()
		coll := mgm.CollectionByName(getCollectionName(objectPtr))
		updateBson := getUpdateBson(objectPtr)
		for _, f := range omitFields {
			delete(updateBson, f)
		}
		res, err := coll.UpdateMany(s.ctx, filter, updateBson)
		return res.ModifiedCount, err
	} else {
		return s.adapter.Engine.Context(s.ctx).ID(core.PK{owner, name}).AllCols().Omit(omitFields...).Update(getElemOfPointer(objectPtr))
	}
}

func getUpdateBson(objectPtr interface{}) bson.M {
	objValue := reflect.ValueOf(objectPtr).Elem()
	objType := objValue.Type()
	updateFields := bson.D{}
	for i := 0; i < objType.NumField(); i++ {
		k := util.SnakeString(objType.Field(i).Name)
		v := objValue.Field(i).Interface()
		updateFields = append(updateFields, bson.E{Key: k, Value: v})
	}
	return bson.D{{Key: "$set", Value: updateFields}}.Map()
}

func getElemOfPointer(objectPtr interface{}) interface{} {
	return reflect.ValueOf(objectPtr).Elem().Interface()
}
