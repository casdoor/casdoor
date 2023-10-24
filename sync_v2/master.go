package sync_v2

import (
	"fmt"
	"log"
)

func createSlaveUser(masterdb *Database) {
	var res = make([]map[string]string, 0)
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln(err)
		}
	}()

	res = masterdb.exec("show databases")
	dbNames := make([]string, 0, len(res))
	for _, dbInfo := range res {
		dbName := dbInfo["Database"]
		dbNames = append(dbNames, dbName)
	}
	log.Println("dbs in mysql: ", dbNames)
	res = masterdb.exec("show tables")
	tableNames := make([]string, 0, len(res))
	for _, table := range res {
		tableName := table[fmt.Sprintf("Tables_in_%v", masterdb.database)]
		tableNames = append(tableNames, tableName)
	}
	log.Printf("tables in %v: %v", masterdb.database, tableNames)

	// delete user to prevent user already exists
	res = masterdb.exec("delete from mysql.user where user = '%v'", masterdb.slaveUser)
	res = masterdb.exec("flush privileges")

	// create replication user
	res = masterdb.exec("create user '%s'@'%s' identified by '%s'", masterdb.slaveUser, "%", masterdb.slavePassword)
	res = masterdb.exec("select host, user from mysql.user where user = '%v'", masterdb.slaveUser)
	log.Println("user: ", res[0])
	res = masterdb.exec("grant replication slave on *.* to '%s'@'%s'", masterdb.slaveUser, "%")
	res = masterdb.exec("flush privileges")
	res = masterdb.exec("show grants for '%s'@'%s'", masterdb.slaveUser, "%")
	log.Println("grants: ", res[0])

	// check env
	res = masterdb.exec("show variables like 'server_id'")
	log.Println("server_id: ", res[0]["Value"])
	res = masterdb.exec("show variables like 'log_bin'")
	log.Println("log_bin: ", res[0]["Value"])
	res = masterdb.exec("show variables like 'binlog_format'")
	log.Println("binlog_format: ", res[0]["Value"])
	res = masterdb.exec("show variables like 'binlog_row_image'")

}

func masterStatus(masterdb *Database) {
	res := masterdb.exec("show master status")
	if len(res) == 0 {
		log.Printf("no master status for master [%v:%v]\n", masterdb.host, masterdb.port)
		return
	}
	pos := res[0]["Position"]
	file := res[0]["File"]
	log.Printf("\n[master: %v:%v]\nfile: %v\nposition: %v\nmaster status: %v\n", masterdb.host, masterdb.port, file, pos, res)

}
