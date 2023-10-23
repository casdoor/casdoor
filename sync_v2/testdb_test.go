package sync_v2

import (
	"fmt"
	"github.com/casdoor/casdoor/util"
	"log"
	"math/rand"
	"testing"
)

type TestUser struct {
	Id       int64  `xorm:"pk autoincr"`
	Username string `xorm:"varchar(50)"`
	Age      int
}

var (
	masterIp   = "62.234.28.252"
	masterPort = 8001
	masterAddr = fmt.Sprintf("%v:%v", masterIp, masterPort)
	slaveIp    = "62.234.28.252"
	slavePort  = 8002
	slaveAddr  = fmt.Sprintf("%v:%v", slaveIp, slavePort)
	syncDbName = "testdb"
	slaveUser  = "repl_user"
)

func TestCreateUserTable(t *testing.T) {
	masterdb := newDatabase(masterIp, masterPort, syncDbName, "root", "test_mysql_password")
	err := masterdb.engine.Sync2(new(TestUser))
	if err != nil {
		log.Fatalln(err)
	}
}

func TestInsertUser(t *testing.T) {
	masterdb := newDatabase(masterIp, masterPort, syncDbName, "root", "test_mysql_password")
	// random generate user
	user := &TestUser{
		Username: util.GetRandomName(),
		Age:      rand.Intn(100) + 10,
	}
	_, err := masterdb.engine.Insert(user)
	if err != nil {
		log.Fatalln(err)
	}
}
