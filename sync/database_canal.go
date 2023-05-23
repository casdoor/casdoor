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

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/siddontang/go-log/log"
)

func (db *Database) OnGTID(header *replication.EventHeader, gtid mysql.GTIDSet) error {
	log.Info("OnGTID: ", gtid.String())
	db.Gtid = gtid.String()
	return nil
}

func (db *Database) onDDL(header *replication.EventHeader, nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	log.Info("into DDL event")
	return nil
}

func (db *Database) OnRow(e *canal.RowsEvent) error {
	log.Info("serverId: ", e.Header.ServerID)
	if strings.Contains(db.Gtid, db.serverUuid) {
		return nil
	}

	// Set the next gtid of the target library to the gtid of the current target library to avoid loopbacks
	db.engine.Exec(fmt.Sprintf("SET GTID_NEXT= '%s'", db.Gtid))
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
		db.engine.Exec("BEGIN")
		for i, row := range e.Rows {
			for j, item := range row {
				if i%2 == 0 {
					if isChar[j]{
						oldColumnValue[j] = fmt.Sprintf("%s", item)
					} else {
						oldColumnValue[j] = fmt.Sprintf("%d", item)
					}
				} else {
					if isChar[j] {
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

				res, err := db.engine.DB().Exec(updateSql, args...)
				if err != nil {
					return err
				}
				log.Info(updateSql, args, res)
			}
		}
		db.engine.Exec("COMMIT")
		db.engine.Exec("SET GTID_NEXT='automatic'")
	case canal.DeleteAction:
		db.engine.Exec("BEGIN")
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j]  {
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

			res, err := db.engine.DB().Exec(deleteSql, args...)
			if err != nil {
				return err
			}
			log.Info(deleteSql, args, res)
		}
		db.engine.Exec("COMMIT")
		db.engine.Exec("SET GTID_NEXT='automatic'")
	case canal.InsertAction:
		db.engine.Exec("BEGIN")
		for _, row := range e.Rows {
			for j, item := range row {
				if isChar[j] {
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

			res, err := db.engine.DB().Exec(insertSql, args...)
			if err != nil {
				return err
			}
			log.Info(insertSql, args, res)
		}
		db.engine.Exec("COMMIT")
		db.engine.Exec("SET GTID_NEXT='automatic'")
	default:
		log.Infof("%v", e.String())
	}
	return nil
}

func (db *Database) String() string {
	return "Database"
}
