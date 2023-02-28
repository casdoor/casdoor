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
	// We only care table in database1
	cfg.Dump.TableDB = database1
	//cfg.Dump.Tables = []string{"user"}
	log.Info("config canal success…")
	return cfg
}

func StartBinlogSync() error {
	// init config
	config := InitConfig()

	c, err := canal.NewCanal(config)
	pos, err := c.GetMasterPos()

	if err != nil {
		return err
	}

	// Register a handler to handle RowsEvent
	c.SetEventHandler(&MyEventHandler{})

	// Start canal
	c.RunFrom(pos)

	return nil
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
	var oldColumnValue = make([]interface{}, length)
	var newColumnValue = make([]interface{}, length)
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
						oldColumnValue[j] = fmt.Sprintf("%s", item)
					} else {
						oldColumnValue[j] = fmt.Sprintf("%d", item)
					}
				} else {
					if isChar[j] == true {
						newColumnValue[j] = fmt.Sprintf("%s", item)
					} else {
						newColumnValue[j] = fmt.Sprintf("%d", item)
					}
				}
			}

			if i%2 == 1 {
				updateSql, args, err := GetUpdateSql(e.Table.Schema, e.Table.Name, columnNames, newColumnValue, oldColumnValue)
				if err != nil {
					return err
				}

				res, err := engin2.DB().Exec(updateSql, args...)
				if err != nil {
					return err
				}
				log.Info(updateSql, args, res)
			}
		}
	case canal.DeleteAction:
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j] == true {
					oldColumnValue[j] = fmt.Sprintf("%s", item)
				} else {
					oldColumnValue[j] = fmt.Sprintf("%d", item)
				}
			}

			deleteSql, args, err := GetDeleteSql(e.Table.Schema, e.Table.Name, columnNames, oldColumnValue)
			if err != nil {
				return err
			}

			res, err := engin2.DB().Exec(deleteSql, args...)
			if err != nil {
				return err
			}
			log.Info(deleteSql, args, res)
		}
	case canal.InsertAction:
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j] == true {
					newColumnValue[j] = fmt.Sprintf("%s", item)
				} else {
					newColumnValue[j] = fmt.Sprintf("%d", item)
				}
			}

			insertSql, args, err := GetInsertSql(e.Table.Schema, e.Table.Name, columnNames, newColumnValue)

			if err != nil {
				return err
			}

			res, err := engin2.DB().Exec(insertSql, args...)
			if err != nil {
				return err
			}
			log.Info(insertSql, args, res)
		}
	default:
		log.Infof("%v", e.String())
	}
	return nil
}

func (h *MyEventHandler) String() string {
	return "MyEventHandler"
}
