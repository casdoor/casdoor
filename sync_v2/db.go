package sync_v2

import (
	"fmt"
	"log"

	"github.com/xorm-io/xorm"
)

type Database struct {
	host          string
	port          int
	database      string
	username      string
	password      string
	slaveUser     string
	slavePassword string
	engine        *xorm.Engine
}

func (db *Database) exec(format string, args ...interface{}) []map[string]string {
	sql := fmt.Sprintf(format, args...)
	res, err := db.engine.QueryString(sql)
	if err != nil {
		panic(err)
	}
	return res
}

func createEngine(dataSourceName string) (*xorm.Engine, error) {
	engine, err := xorm.NewEngine("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	// ping mysql
	err = engine.Ping()
	if err != nil {
		return nil, err
	}

	engine.ShowSQL(true)
	log.Println("mysql connection success")
	return engine, nil
}

func newDatabase(db *Database) *Database {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", db.username, db.password, db.host, db.port, db.database)
	engine, err := createEngine(dataSourceName)
	if err != nil {
		panic(err)
	}

	db.engine = engine
	return db
}
