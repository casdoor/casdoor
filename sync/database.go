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

package sync

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/xorm-io/xorm"
)

type Database struct {
	host     string
	port     int
	database string
	username string
	password string

	engine     *xorm.Engine
	serverId   uint32
	serverUuid string
	Gtid       string
	canal.DummyEventHandler
}

func newDatabase(host string, port int, database string, username string, password string) *Database {
	db := &Database{
		host:     host,
		port:     port,
		database: database,
		username: username,
		password: password,
	}

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
	engine, err := createEngine(dataSourceName)
	if err != nil {
		panic(err)
	}

	db.engine = engine

	db.serverId, err = getServerId(engine)
	if err != nil {
		panic(err)
	}

	db.serverUuid, err = getServerUuid(engine)
	if err != nil {
		panic(err)
	}

	return db
}

func (db *Database) getCanalConfig() *canal.Config {
	// config canal
	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", db.host, db.port)
	cfg.Password = db.password
	cfg.User = db.username
	// We only care table in database1
	cfg.Dump.TableDB = db.database
	return cfg
}

func (db *Database) startCanal(targetDb *Database) error {
	canalConfig := db.getCanalConfig()
	c, err := canal.NewCanal(canalConfig)
	if err != nil {
		return err
	}

	gtidSet, err := c.GetMasterGTIDSet()
	if err != nil {
		return err
	}

	// Register a handler to handle RowsEvent
	c.SetEventHandler(targetDb)

	// Start replication
	err = c.StartFromGTID(gtidSet)
	if err != nil {
		return err
	}
	return nil
}
