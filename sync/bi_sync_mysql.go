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
	"log"

	"github.com/2tvenom/myreplication"
	"github.com/xorm-io/xorm"
	"github.com/xorm-io/xorm/schemas"
)

var (
	dbTables        = make(map[string]*schemas.Table)
	dataSourceName1 string
	dataSourceName2 string
	engin1          *xorm.Engine
	engin2          *xorm.Engine
)

func InitConfig() {
	// init dataSource
	dataSourceName1 = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username1, password1, host1, port1, database1)
	dataSourceName2 = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username2, password2, host2, port2, database2)

	// create engine
	engin1, _ = CreateEngine(dataSourceName1)
	engin2, _ = CreateEngine(dataSourceName2)

	// create connection
	log.Println("init sync success……")
}

func StartSync() {

	// init config
	InitConfig()

	// start dump mysql1 binlog and reloading to mysql2
	StartDumpBinlog(host1, port1, username1, password1, uint32(serverId1), engin2)
}

func StartDumpBinlog(host string, port int, username string, password string, serverId uint32, engin *xorm.Engine) {
	newConnection := myreplication.NewConnection()
	err := newConnection.ConnectAndAuth(host, port, username, password)

	if err != nil {
		panic("Client not connected and not autentificate to master server with error:" + err.Error())
	}
	//Get position and file name
	pos, filename, err := newConnection.GetMasterStatus()

	if err != nil {
		panic("Master status fail: " + err.Error())
	}

	el, err := newConnection.StartBinlogDump(pos, filename, serverId)

	if err != nil {
		panic("Cant start bin log: " + err.Error())
	}

	if err != nil {
		panic("cah: " + err.Error())
	}

	// get event chan
	events := el.GetEventChan()

	go ListenAndRelay(events, engin)

	err = el.Start()
	println(err.Error())
}

func ListenAndRelay(events <-chan interface{}, engin *xorm.Engine) {
	for {
		event := <-events
		//fmt.Println(event)
		switch e := event.(type) {
		case *myreplication.QueryEvent:
			// Output query event
			// BEGIN is to start the transaction, which is closed here
			if e.GetQuery() != "BEGIN" {
				_, err := engin.Exec(e.GetQuery())
				if err != nil {
					panic("exec sql error " + err.Error())
				}
				err = updateTable(engin)
				if err != nil {
					panic(err.Error())
				}
				log.Println("Query:", e.GetQuery())
			}
		case *myreplication.IntVarEvent:
			//Output last insert_id  if statement based replication
			println(e.GetValue())
		case *myreplication.WriteEvent:
			//Output Write (insert) event
			log.Println("Write", e.GetTable())
			//Rows loop
			columnNames := GetColumns(dbTables[e.GetTable()].Columns())
			for _, row := range e.GetRows() {
				//Columns loop
				columnVals := make([]string, len(row))
				for j, col := range row {
					//Output row number, column number, column type and column value
					if IsChar(col.GetType()) {
						columnVals[j] = fmt.Sprintf("'%v'", col.GetValue())
					} else {
						columnVals[j] = fmt.Sprintf("%v", col.GetValue())
					}
				}
				strSql := GetInsertSql(e.GetSchema(), e.GetTable(), columnNames, columnVals)
				_, err := engin.Exec(strSql)
				if err != nil {
					panic("exec sql error " + err.Error())
				}
				log.Println(strSql)
			}
		case *myreplication.DeleteEvent:
			//Output delete event
			log.Println("Delete", e.GetTable())
			columnNames := GetColumns(dbTables[e.GetTable()].Columns())
			for _, row := range e.GetRows() {
				//Columns loop
				columnVals := make([]string, len(row))
				for j, col := range row {
					if IsChar(col.GetType()) {
						columnVals[j] = fmt.Sprintf("'%v'", col.GetValue())
					} else {
						columnVals[j] = fmt.Sprintf("%v", col.GetValue())
					}
				}
				strSql := GetdeleteSql(e.GetSchema(), e.GetTable(), columnNames, columnVals)
				_, err := engin.Exec(strSql)
				if err != nil {
					panic("exec sql error " + err.Error())
				}
				log.Println(strSql)
			}

		case *myreplication.UpdateEvent:
			//Output update event
			println("Update", e.GetTable())
			columnNames := GetColumns(dbTables[e.GetTable()].Columns())
			// Output old
			oldColumnValList := make([][]string, len(e.GetRows()))
			for i, row := range e.GetRows() {
				//Columns loop
				oldColumnValList[i] = make([]string, len(row))
				for j, col := range row {
					if IsChar(col.GetType()) {
						oldColumnValList[i][j] = fmt.Sprintf("'%v'", col.GetValue())
					} else {
						oldColumnValList[i][j] = fmt.Sprintf("%v", col.GetValue())
					}
				}
			}
			// Output new
			newColumnValList := make([][]string, len(e.GetNewRows()))
			for i, row := range e.GetNewRows() {
				//Columns loop
				newColumnValList[i] = make([]string, len(row))
				for j, col := range row {
					if IsChar(col.GetType()) {
						newColumnValList[i][j] = fmt.Sprintf("'%v'", col.GetValue())
					} else {
						newColumnValList[i][j] = fmt.Sprintf("%v", col.GetValue())
					}
				}
			}
			strSql := GetUpdateSql(e.GetSchema(), e.GetTable(), columnNames, newColumnValList)
			_, err := engin.Exec(strSql)
			if err != nil {
				panic("exec sql error " + err.Error())
			}
			log.Println(strSql)
		case *myreplication.XidEvent:
			fmt.Println("serverID : ", e.ServerId)
		default:
		}
	}
}
