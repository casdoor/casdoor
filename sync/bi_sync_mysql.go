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

type Database struct {
	dataSourceName string
	engine         *xorm.Engine
	serverId       uint32
	serverUuid     string
	Gtid           string
	canal.DummyEventHandler
}

func StartCanal(cfg *canal.Config, username string, password string, host string, port int, database string) error {
	c, err := canal.NewCanal(cfg)
	if err != nil {
		return err
	}

	gtidSet, err := c.GetMasterGTIDSet()
	if err != nil {
		return err
	}

	db := createDatabase(username, password, host, port, database)
	// Register a handler to handle RowsEvent
	c.SetEventHandler(&db)

	// Start replication
	err = c.StartFromGTID(gtidSet)
	if err != nil {
		return err
	}
	return nil
}

func StartBinlogSync() error {
	var wg sync.WaitGroup
	// init config
	cfg1 := getCanalConfig(username1, password1, host1, port1, database1)
	cfg2 := getCanalConfig(username2, password2, host2, port2, database2)

	// start canal1 replication
	go StartCanal(cfg1, username2, password2, host2, port2, database2)
	wg.Add(1)

	// start canal2 replication
	go StartCanal(cfg2, username1, password1, host1, port1, database1)
	wg.Add(1)

	wg.Wait()
	return nil
}

func (h *Database) OnGTID(header *replication.EventHeader, gtid mysql.GTIDSet) error {
	log.Info("OnGTID: ", gtid.String())
	h.Gtid = gtid.String()
	return nil
}

func (h *Database) onDDL(header *replication.EventHeader, nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	log.Info("into DDL event")
	return nil
}

func (h *Database) OnRow(e *canal.RowsEvent) error {
	log.Info("serverId: ", e.Header.ServerID)
	if strings.Contains(h.Gtid, h.serverUuid) {
		return nil
	}

	// Set the next gtid of the target library to the gtid of the current target library to avoid loopbacks
	h.engine.Exec(fmt.Sprintf("SET GTID_NEXT= '%s'", h.Gtid))
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
	pkColumnNames := getPkColumnNames(columnNames, e.Table.PKColumns)

	switch e.Action {
	case canal.UpdateAction:
		h.engine.Exec("BEGIN")
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
				pkColumnValue := getPkColumnValues(oldColumnValue, e.Table.PKColumns)
				updateSql, args, err := getUpdateSql(e.Table.Schema, e.Table.Name, columnNames, newColumnValue, pkColumnNames, pkColumnValue)
				if err != nil {
					return err
				}

				res, err := h.engine.DB().Exec(updateSql, args...)
				if err != nil {
					return err
				}
				log.Info(updateSql, args, res)
			}
		}
		h.engine.Exec("COMMIT")
		h.engine.Exec("SET GTID_NEXT='automatic'")
	case canal.DeleteAction:
		h.engine.Exec("BEGIN")
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j] == true {
					oldColumnValue[j] = fmt.Sprintf("%s", item)
				} else {
					oldColumnValue[j] = fmt.Sprintf("%d", item)
				}
			}

			pkColumnValue := getPkColumnValues(oldColumnValue, e.Table.PKColumns)
			deleteSql, args, err := getDeleteSql(e.Table.Schema, e.Table.Name, pkColumnNames, pkColumnValue)
			if err != nil {
				return err
			}

			res, err := h.engine.DB().Exec(deleteSql, args...)
			if err != nil {
				return err
			}
			log.Info(deleteSql, args, res)
		}
		h.engine.Exec("COMMIT")
		h.engine.Exec("SET GTID_NEXT='automatic'")
	case canal.InsertAction:
		h.engine.Exec("BEGIN")
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

			insertSql, args, err := getInsertSql(e.Table.Schema, e.Table.Name, columnNames, newColumnValue)
			if err != nil {
				return err
			}

			res, err := h.engine.DB().Exec(insertSql, args...)
			if err != nil {
				return err
			}
			log.Info(insertSql, args, res)
		}
		h.engine.Exec("COMMIT")
		h.engine.Exec("SET GTID_NEXT='automatic'")
	default:
		log.Infof("%v", e.String())
	}
	return nil
}

func (h *Database) String() string {
	return "Database"
}
