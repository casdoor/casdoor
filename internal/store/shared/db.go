// Copyright 2020 The casbin Authors. All Rights Reserved.
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

package shared

import (
	"fmt"

	"github.com/casdoor/casdoor/internal/config"
	"github.com/casdoor/casdoor/internal/object"
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

type DB struct {
	engine         *xorm.Engine
	driverName     string
	dataSourceName string
}

func NewDB(cfg *config.Config) (*DB, error) {
	db := &DB{
		dataSourceName: cfg.DBDataSource,
		driverName:     "mysql",
	}
	engine, err := xorm.NewEngine(db.driverName, db.dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("xorm.NewEngine: %v", err)
	}
	err = engine.Ping()
	if err != nil {
		return nil, fmt.Errorf("engine.Ping(): %v", err)
	}

	db.engine = engine

	err = db.createTable()
	if err != nil {
		return nil, fmt.Errorf("db.createTable(): %v", err)
	}

	return db, nil
}

func (db *DB) GetEngine() *xorm.Engine {
	return db.engine
}

func (db *DB) createTable() error {
	err := db.engine.Sync2(new(object.User), new(object.Application))
	return err
}
