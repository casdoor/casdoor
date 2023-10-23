package sync_v2

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xorm-io/xorm"
	"log"
	"testing"
)

type Database struct {
	host     string
	port     int
	database string
	username string
	password string

	engine *xorm.Engine
}

func (db *Database) exec(format string, args ...interface{}) []map[string]string {
	sql := fmt.Sprintf(format, args...)
	res, err := db.engine.QueryString(sql)
	if err != nil {
		panic(err)
	}
	return res
}

func newDatabase(host string, port int, database string, username string, password string) *Database {
	db := &Database{
		host:     host,
		port:     port,
		database: database,
		username: username,
		password: password,
	}

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
	engine, err := createEngine(dataSourceName)
	if err != nil {
		panic(err)
	}

	db.engine = engine

	return db
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
	log.Println("mysql connection success……")
	return engine, nil
}

func TestSlaveChangeToMaster(t *testing.T) {
	var res = make([]map[string]string, 0)
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln(err)
		}
	}()

	slavedb := newDatabase(slaveIp, slavePort, syncDbName, "root", "test_mysql_password")
	masterdb := newDatabase(masterIp, masterPort, syncDbName, "root", "test_mysql_password")
	res = masterdb.exec("show master status")
	pos := res[0]["Position"]
	file := res[0]["File"]
	log.Println("file:", file, ", position:", pos, ", master status:", res)
	res = slavedb.exec("stop slave")
	res = slavedb.exec(
		"change master to master_host='%v', master_port=%v, master_user='%v', master_password='%v', master_log_file='%v', master_log_pos=%v;",
		masterIp, masterPort, slaveUser, slaveUser, file, pos,
	)
	res = slavedb.exec("start slave")
	res = slavedb.exec("show slave status")

	slaveIoState := res[0]["Slave_IO_State"]
	slaveIoRunning := res[0]["Slave_IO_Running"]
	slaveSqlRunning := res[0]["Slave_SQL_Running"]
	slaveSqlRunningState := res[0]["Slave_SQL_Running_State"]
	masterServerId := res[0]["Master_Server_Id"]
	log.Printf("\nslave io state: %v\nslave io running: %v\nslave sql running: %v\nslave sql running state: %v\nmaster server id: %v\nslave status: %v\n",
		slaveIoState, slaveIoRunning, slaveSqlRunning, slaveSqlRunningState, masterServerId, res)
}

func TestMasterCreateSlaveUser(t *testing.T) {
	var res = make([]map[string]string, 0)
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln(err)
		}
	}()

	db := newDatabase(masterIp, masterPort, syncDbName, "root", "test_mysql_password")
	res = db.exec("show databases")
	dbNames := make([]string, 0, len(res))
	for _, dbInfo := range res {
		dbName := dbInfo["Database"]
		dbNames = append(dbNames, dbName)
	}
	log.Println("dbs in mysql: ", dbNames)
	res = db.exec("show tables")
	tableNames := make([]string, 0, len(res))
	for _, table := range res {
		tableName := table[fmt.Sprintf("Tables_in_%v", syncDbName)]
		tableNames = append(tableNames, tableName)
	}
	log.Printf("tables in %v: %v", syncDbName, tableNames)

	// delete user to prevent user already exists
	res = db.exec("delete from mysql.user where user = '%v'", slaveUser)
	res = db.exec("flush privileges")

	// create replication user
	res = db.exec("create user '%s'@'%s' identified by '%s'", slaveUser, "%", slaveUser)
	res = db.exec("select host, user from mysql.user where user = '%v'", slaveUser)
	log.Println("user: ", res[0])
	res = db.exec("grant replication slave on *.* to '%s'@'%s'", slaveUser, "%")
	res = db.exec("flush privileges")
	res = db.exec("show grants for '%s'@'%s'", slaveUser, "%")
	log.Println("grants: ", res[0])

	// check env
	res = db.exec("show variables like 'server_id'")
	log.Println("server_id: ", res[0]["Value"])
	res = db.exec("show variables like 'log_bin'")
	log.Println("log_bin: ", res[0]["Value"])
	res = db.exec("show variables like 'binlog_format'")
	log.Println("binlog_format: ", res[0]["Value"])

}
