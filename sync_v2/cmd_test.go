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

	gtid_mode=on
	enforce_gtid_consistency=on
	binlog-format=ROW
	server-id = 1 # this should be different for each mysql instance (1,2)
	auto_increment_offset = 1 # this is same as server-id
	auto_increment_increment = 2 # this is same as the number of mysql instances (2)
	log-bin = mysql-bin
	replicate-do-db = casdoor # this is the database name
	binlog-do-db = casdoor # this is the database name
*/

var Configs = []Database{
	{
		host:     "test-db.v2tl.com",
		port:     3306,
		username: "root",
		password: "password",
		database: "casdoor",
		// the following two fields are used to create replication user, you don't need to change them
		slaveUser:     "repl_user",
		slavePassword: "repl_user",
	},
	{
		host:     "localhost",
		port:     3306,
		username: "root",
		password: "password",
		database: "casdoor",
		// the following two fields are used to create replication user, you don't need to change them
		slaveUser:     "repl_user",
		slavePassword: "repl_user",
	},
}

func TestStartMasterSlaveSync(t *testing.T) {
	// for example, this is aliyun rds
	db0 := newDatabase(&Configs[0])
	// for example, this is local mysql instance
	db1 := newDatabase(&Configs[1])

	createSlaveUser(db0)
	// db0 is master, db1 is slave
	startSlave(db0, db1)
}

func TestStopMasterSlaveSync(t *testing.T) {
	// for example, this is aliyun rds
	db0 := newDatabase(&Configs[0])
	// for example, this is local mysql instance
	db1 := newDatabase(&Configs[1])

	stopSlave(db1)
	deleteSlaveUser(db0)
}

func TestStartMasterMasterSync(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	createSlaveUser(db0)
	createSlaveUser(db1)
	// db0 is master, db1 is slave
	startSlave(db0, db1)
	// db1 is master, db0 is slave
	startSlave(db1, db0)
}

func TestStopMasterMasterSync(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	stopSlave(db0)
	stopSlave(db1)
	deleteSlaveUser(db0)
	deleteSlaveUser(db1)
}

func TestShowSlaveStatus(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	slaveStatus(db0)
	slaveStatus(db1)
}

func TestShowMasterStatus(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	masterStatus(db0)
	masterStatus(db1)
}
