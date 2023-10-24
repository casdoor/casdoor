package sync_v2

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

var Configs = []Database{
	{
		host:          "62.234.28.252",
		port:          8001,
		username:      "root",
		password:      "test_mysql_password",
		slaveUser:     "repl_user",
		slavePassword: "repl_user",
		database:      "testdb",
	},
	{
		host:          "62.234.28.252",
		port:          8002,
		username:      "root",
		password:      "test_mysql_password",
		slaveUser:     "repl_user",
		slavePassword: "repl_user",
		database:      "testdb",
	},
}

func TestMasterCreateSlaveUser(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	createSlaveUser(db0)
	createSlaveUser(db1)
}

func TestStartSlave(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	startSlave(db0, db1)
	startSlave(db1, db0)
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

func TestMaster2MasterSync(t *testing.T) {
	db0 := newDatabase(&Configs[0])
	db1 := newDatabase(&Configs[1])
	createSlaveUser(db0)
	createSlaveUser(db1)
	startSlave(db0, db1)
	startSlave(db1, db0)
	slaveStatus(db0)
	slaveStatus(db1)
}
