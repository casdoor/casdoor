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

//go:build !skipCi
// +build !skipCi

package sync_v2

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

/*
	The following config should be added to my.cnf:

	binlog-format=ROW
	server-id = 1 # this should be different for each mysql instance (1,2)
	auto_increment_offset = 1 # this is same as server-id
	auto_increment_increment = 2 # this is same as the number of mysql instances (2)
	log-bin = mysql-bin
	replicate-do-db = testdb # this is the database name
	replicate-ignore-db = mysql,information_schema,performance_schema,sys
	binlog-do-db = testdb # this is the database name
	binlog-ignore-db = mysql,information_schema,performance_schema,sys
*/

var Configs = []Database{
	{
		host:          "127.0.0.1",
		port:          3306,
		username:      "root",
		password:      "test_mysql_password",
		slaveUser:     "repl_user",
		slavePassword: "repl_user",
		database:      "casdoor",
	},
	{
		host:          "127.0.0.1",
		port:          3307,
		username:      "root",
		password:      "test_mysql_password",
		slaveUser:     "repl_user",
		slavePassword: "repl_user",
		database:      "casdoor",
	},
}

func TestStartSlave(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	createSlaveUser(db0)
	createSlaveUser(db1)
	startSlave(db0, db1)
	startSlave(db1, db0)
}

func TestStopSlave(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	stopSlave(db0)
	stopSlave(db1)
}

func TestCheckSlaveStatus(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	slaveStatus(db0)
	slaveStatus(db1)
}

func TestCheckMasterStatus(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	masterStatus(db0)
	masterStatus(db1)
}
