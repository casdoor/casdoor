package sync_v2

import (
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

func TestCreateUserTable(t *testing.T) {
	db := newDatabase(&Configs[0])
	err := db.engine.Sync2(new(TestUser))
	if err != nil {
		log.Fatalln(err)
	}
}

func TestInsertUser(t *testing.T) {
	db := newDatabase(&Configs[0])
	// random generate user
	user := &TestUser{
		Username: util.GetRandomName(),
		Age:      rand.Intn(100) + 10,
	}
	_, err := db.engine.Insert(user)
	if err != nil {
		log.Fatalln(err)
	}
}
