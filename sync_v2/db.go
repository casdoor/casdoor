// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package sync_v2

import (
	"fmt"
	"log"

	"github.com/xorm-io/xorm"
)

type Database struct {
	host          string
	port          int
	database      string
	username      string
	password      string
	slaveUser     string
	slavePassword string
	engine        *xorm.Engine
}

func (db *Database) exec(format string, args ...interface{}) []map[string]string {
	sql := fmt.Sprintf(format, args...)
	res, err := db.engine.QueryString(sql)
	if err != nil {
		panic(err)
	}
	return res
}

func createEngine(dataSourceName string) (*xorm.Engine, error) {
	engine, err := xorm.NewEngine("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	// ping mysql
	err = engine.Ping()
	if err != nil {
		return nil, err
	}

	engine.ShowSQL(true)
	log.Println("mysql connection success")
	return engine, nil
}

func newDatabase(db *Database) *Database {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", db.username, db.password, db.host, db.port, db.database)
	engine, err := createEngine(dataSourceName)
	if err != nil {
		panic(err)
	}

	db.engine = engine
	return db
}
