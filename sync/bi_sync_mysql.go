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
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/siddontang/go-log/log"
	"github.com/xorm-io/xorm"
)

var (
	dataSourceName1 string
	dataSourceName2 string
	engin1          *xorm.Engine
	engin2          *xorm.Engine
)

func InitConfig() *canal.Config {
	// init dataSource
	dataSourceName1 = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username1, password1, host1, port1, database1)
	dataSourceName2 = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username2, password2, host2, port2, database2)

	// create engine
	engin1, _ = CreateEngine(dataSourceName1)
	engin2, _ = CreateEngine(dataSourceName2)
	log.Info("init engine success…")

	// config canal
	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", host1, port1)
	cfg.Password = password1
	cfg.User = username1
	// We only care table canal_test in test db
	cfg.Dump.TableDB = database1
	//cfg.Dump.Tables = []string{"user"}
	log.Info("config canal success…")
	return cfg
}

func StartBinlogSync() {
	// init config
	config := InitConfig()

	c, err := canal.NewCanal(config)
	pos, err := c.GetMasterPos()

	if err != nil {
		log.Fatal(err)
	}

	// Register a handler to handle RowsEvent
	c.SetEventHandler(&MyEventHandler{})

	// Start canal
	c.RunFrom(pos)
}

type MyEventHandler struct {
	canal.DummyEventHandler
}

func OnTableChanged(header *replication.EventHeader, schema string, table string) error {
	log.Info("table changed event")
	return nil
}

func (h *MyEventHandler) onDDL(header *replication.EventHeader, nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	log.Info("into DDL event")
	return nil
}

func (h *MyEventHandler) OnRow(e *canal.RowsEvent) error {
	length := len(e.Table.Columns)
	var columnNames = make([]string, length)
	var oldColumnValue = make([]string, length)
	var newColumnValue = make([]string, length)
	var isChar = make([]bool, len(e.Table.Columns))
	for i, col := range e.Table.Columns {
		columnNames[i] = col.Name
		if col.Type <= 2 {
			isChar[i] = false
		} else {
			isChar[i] = true
		}
	}
	switch e.Action {
	case canal.UpdateAction:
		for i, row := range e.Rows {
			for j, item := range row {
				if i%2 == 0 {
					if isChar[j] == true {
						oldColumnValue[j] = fmt.Sprintf("'%s'", item)
					} else {
						oldColumnValue[j] = fmt.Sprintf("%d", item)
					}
				} else {
					if isChar[j] == true {
						newColumnValue[j] = fmt.Sprintf("'%s'", item)
					} else {
						newColumnValue[j] = fmt.Sprintf("%d", item)
					}
				}
			}
			if i%2 == 1 {
				updateSql := GetUpdateSql(e.Table.Schema, e.Table.Name, columnNames, newColumnValue, oldColumnValue)
				engin2.Exec(updateSql)
				log.Info(updateSql)
			}
		}
	case canal.DeleteAction:
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j] == true {
					oldColumnValue[j] = fmt.Sprintf("'%s'", item)
				} else {
					oldColumnValue[j] = fmt.Sprintf("%d", item)
				}
			}
			deleteSql := GetdeleteSql(e.Table.Schema, e.Table.Name, columnNames, oldColumnValue)
			engin2.Exec(deleteSql)
			log.Info(deleteSql)
		}
		log.Infof("%s %v\n", e.Table.Name, e.Rows)
	case canal.InsertAction:
		fmt.Println("Insert")
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j] == true {
					newColumnValue[j] = fmt.Sprintf("'%s'", item)
				} else {
					newColumnValue[j] = fmt.Sprintf("%d", item)
				}
			}
			insertSql := GetInsertSql(e.Table.Schema, e.Table.Name, columnNames, newColumnValue)
			engin2.Exec(insertSql)
			log.Info(insertSql)
		}
	default:
		log.Infof("%v", e.String())
	}
	return nil
}

func (h *MyEventHandler) String() string {
	return "MyEventHandler"
}
