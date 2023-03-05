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
	"strings"
	"sync"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/siddontang/go-log/log"
	"github.com/xorm-io/xorm"
)

var (
	GTID            string
	dataSourceName1 string
	dataSourceName2 string
	engin           *xorm.Engine
	engin1          *xorm.Engine
	engin2          *xorm.Engine
	serverId1       uint32
	serverId2       uint32
)

func InitConfig() (*canal.Config, *canal.Config) {
	// init dataSource
	dataSourceName1 = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username1, password1, host1, port1, database1)
	dataSourceName2 = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username2, password2, host2, port2, database2)

	// create engine
	engin1, _ = CreateEngine(dataSourceName1)
	engin2, _ = CreateEngine(dataSourceName2)

	// get serverId
	serverId1, _ = GetServerId(engin1)
	serverId2, _ = GetServerId(engin2)

	// config canal1
	cfg1 := canal.NewDefaultConfig()
	cfg1.Addr = fmt.Sprintf("%s:%d", host1, port1)
	cfg1.Password = password1
	cfg1.User = username1
	// We only care table in database1
	cfg1.Dump.TableDB = database1

	// config canal2
	cfg2 := canal.NewDefaultConfig()
	cfg2.Addr = fmt.Sprintf("%s:%d", host2, port2)
	cfg2.Password = password2
	cfg2.User = username2
	// We only care table in database2
	cfg2.Dump.TableDB = database2
	// cfg.Dump.Tables = []string{"user"}
	log.Info("config canal success…")
	return cfg1, cfg2
}

func StartBinlogSync() error {
	var wg sync.WaitGroup

	// init config
	config1, config2 := InitConfig()

	c1, err := canal.NewCanal(config1)
	GTIDSet1, err := c1.GetMasterGTIDSet()

	if err != nil {
		return err
	}

	// Register a handler to handle RowsEvent
	c1.SetEventHandler(&MyEventHandler{})

	// Start canal2
	go func() {
		err := c1.StartFromGTID(GTIDSet1)
		if err != nil {
			panic(err)
		}
	}()
	wg.Add(1)

	c2, err := canal.NewCanal(config2)
	GTIDSet2, _ := c2.GetMasterGTIDSet()

	// Register a handler to handle RowsEvent
	c2.SetEventHandler(&MyEventHandler{})

	// Start canal2
	go func() {
		err := c2.StartFromGTID(GTIDSet2)
		if err != nil {
			panic(err)
		}
	}()
	wg.Add(1)

	wg.Wait()
	return nil
}

type MyEventHandler struct {
	canal.DummyEventHandler
}

func (h *MyEventHandler) OnGTID(header *replication.EventHeader, gtid mysql.GTIDSet) error {
	log.Info("OnGTID: ", gtid.String())
	GTID = gtid.String()
	return nil
}

func (h *MyEventHandler) onDDL(header *replication.EventHeader, nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	log.Info("into DDL event")
	return nil
}

func (h *MyEventHandler) OnRow(e *canal.RowsEvent) error {
	log.Info("serverId: ", e.Header.ServerID)

	if e.Header.ServerID == serverId1 {
		engin = engin2
		if strings.Contains(GTID, "92fcbc2d-aaa2-11ed-a1c2-00163e08698e") {
			return nil
		}
	} else if e.Header.ServerID == serverId2 {
		if strings.Contains(GTID, "a95cbaa8-aaac-11ed-b110-00163e1d3d4e") {
			return nil
		}
		engin = engin1
	}
	// Set the next gtid of the target library to the gtid of the current target library to avoid loopbacks
	engin.Exec(fmt.Sprintf("SET GTID_NEXT= '%s'", GTID))
  
	length := len(e.Table.Columns)
	columnNames := make([]string, length)
	oldColumnValue := make([]interface{}, length)
	newColumnValue := make([]interface{}, length)
	isChar := make([]bool, len(e.Table.Columns))

	for i, col := range e.Table.Columns {
		columnNames[i] = col.Name
		if col.Type <= 2 {
			isChar[i] = false
		} else {
			isChar[i] = true
		}
	}

	// get pk column name
	pkColumnNames := GetPKColumnNames(columnNames, e.Table.PKColumns)

	switch e.Action {
	case canal.UpdateAction:
		engin.Exec("BEGIN")
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
						if item == nil {
							newColumnValue[j] = nil
						} else {
							newColumnValue[j] = fmt.Sprintf("%s", item)
						}
					} else {
						newColumnValue[j] = fmt.Sprintf("%d", item)
					}
				}
			}
			if i%2 == 1 {
				pkColumnValue := GetPKColumnValues(oldColumnValue, e.Table.PKColumns)
				updateSql, args, err := GetUpdateSql(e.Table.Schema, e.Table.Name, columnNames, newColumnValue, pkColumnNames, pkColumnValue)

				if err != nil {
					return err
				}

				res, err := engin.DB().Exec(updateSql, args...)

				if err != nil {
					return err
				}
				log.Info(updateSql, args, res)
			}
		}
		engin.Exec("COMMIT")
		engin.Exec("SET GTID_NEXT='automatic'")
	case canal.DeleteAction:
		engin.Exec("BEGIN")
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j] == true {
					oldColumnValue[j] = fmt.Sprintf("%s", item)
				} else {
					oldColumnValue[j] = fmt.Sprintf("%d", item)
				}
			}

			pkColumnValue := GetPKColumnValues(oldColumnValue, e.Table.PKColumns)
			deleteSql, args, err := GetDeleteSql(e.Table.Schema, e.Table.Name, pkColumnNames, pkColumnValue)

			if err != nil {
				return err
			}

			res, err := engin.DB().Exec(deleteSql, args...)
      
			if err != nil {
				return err
			}
			log.Info(deleteSql, args, res)
		}
		engin.Exec("COMMIT")
		engin.Exec("SET GTID_NEXT='automatic'")
	case canal.InsertAction:
		engin.Exec("BEGIN")
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j] == true {
					if item == nil {
						newColumnValue[j] = nil
					} else {
						newColumnValue[j] = fmt.Sprintf("%s", item)
					}
				} else {
					newColumnValue[j] = fmt.Sprintf("%d", item)
				}
			}

			insertSql, args, err := GetInsertSql(e.Table.Schema, e.Table.Name, columnNames, newColumnValue)
			if err != nil {
				return err
			}

			res, err := engin.DB().Exec(insertSql, args...)
      
			if err != nil {
				return err
			}
			log.Info(insertSql, args, res)
		}
		engin.Exec("COMMIT")
		engin.Exec("SET GTID_NEXT='automatic'")
	default:
		log.Infof("%v", e.String())
	}
	return nil
}

func (h *MyEventHandler) String() string {
	return "MyEventHandler"
}
