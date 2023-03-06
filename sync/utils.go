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
	"strconv"

	"github.com/go-mysql-org/go-mysql/canal"

	"github.com/Masterminds/squirrel"
	"github.com/xorm-io/xorm"
)

func GetUpdateSql(schemaName string, tableName string, columnNames []string, newColumnVal []interface{}, pkColumnNames []string, pkColumnValue []interface{}) (string, []interface{}, error) {
	updateSql := squirrel.Update(schemaName + "." + tableName)
	for i, columnName := range columnNames {
		updateSql = updateSql.Set(columnName, newColumnVal[i])
	}

	for i, pkColumnName := range pkColumnNames {
		updateSql = updateSql.Where(squirrel.Eq{pkColumnName: pkColumnValue[i]})
	}

	sql, args, err := updateSql.ToSql()
	if err != nil {
		return "", nil, err
	}

	return sql, args, nil
}

func GetInsertSql(schemaName string, tableName string, columnNames []string, columnValue []interface{}) (string, []interface{}, error) {
	insertSql := squirrel.Insert(schemaName + "." + tableName).Columns(columnNames...).Values(columnValue...)

	return insertSql.ToSql()
}

func GetDeleteSql(schemaName string, tableName string, pkColumnNames []string, pkColumnValue []interface{}) (string, []interface{}, error) {
	deleteSql := squirrel.Delete(schemaName + "." + tableName)

	for i, columnName := range pkColumnNames {
		deleteSql = deleteSql.Where(squirrel.Eq{columnName: pkColumnValue[i]})
	}

	return deleteSql.ToSql()
}

func CreateEngine(dataSourceName string) (*xorm.Engine, error) {
	engine, err := xorm.NewEngine("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	// ping mysql
	err = engine.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("mysql connection success……")
	return engine, nil
}

func GetServerId(engin *xorm.Engine) (uint32, error) {
	res, err := engin.QueryInterface("SELECT @@server_id")
	if err != nil {
		return 0, err
	}
	serverId, _ := strconv.ParseUint(fmt.Sprintf("%s", res[0]["@@server_id"]), 10, 32)
	return uint32(serverId), nil
}

func GetServerUUID(engin *xorm.Engine) (string, error) {
	res, err := engin.QueryString("show variables like 'server_uuid'")
	if err != nil {
		return "", err
	}
	serverUUID := fmt.Sprintf("%s", res[0]["Value"])
	return serverUUID, err
}

func GetPKColumnNames(columnNames []string, PKColumns []int) []string {
	pkColumnNames := make([]string, len(PKColumns))
	for i, index := range PKColumns {
		pkColumnNames[i] = columnNames[index]
	}
	return pkColumnNames
}

func GetPKColumnValues(columnValues []interface{}, PKColumns []int) []interface{} {
	pkColumnNames := make([]interface{}, len(PKColumns))
	for i, index := range PKColumns {
		pkColumnNames[i] = columnValues[index]
	}
	return pkColumnNames
}

func GetCanalConfig(username string, password string, host string, port int, database string) *canal.Config {
	// config canal
	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)
	cfg.Password = password
	cfg.User = username
	// We only care table in database1
	cfg.Dump.TableDB = database
	return cfg
}

func GetMyEventHandler(username string, password string, host string, port int, database string) MyEventHandler {
	var eventHandler MyEventHandler
	eventHandler.dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
	eventHandler.engine, _ = CreateEngine(eventHandler.dataSourceName)
	eventHandler.serverId, _ = GetServerId(eventHandler.engine)
	eventHandler.serverUUID, _ = GetServerUUID(eventHandler.engine)
	return eventHandler
}
