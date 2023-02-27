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
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/xorm-io/xorm"
)

func GetUpdateSql(schemaName string, tableName string, columnNames []string, newColumnVal []interface{}, oldColumnVal []interface{}) (string, []interface{}, error) {
	updateSql := squirrel.Update(schemaName + "." + tableName)
	for i, columnName := range columnNames {
		updateSql = updateSql.Set(columnName, newColumnVal[i])
	}

	for i, columnName := range columnNames {
		updateSql = updateSql.Where(squirrel.Eq{columnName: oldColumnVal[i]})
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

func GetDeleteSql(schemaName string, tableName string, columnNames []string, columnValue []interface{}) (string, []interface{}, error) {
	deleteSql := squirrel.Delete(schemaName + "." + tableName)

	for i, columnName := range columnNames {
		deleteSql = deleteSql.Where(squirrel.Eq{columnName: columnValue[i]})
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
