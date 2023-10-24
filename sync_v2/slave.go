package sync_v2

import "log"

// slaveStatus shows slave status
func slaveStatus(slavedb *Database) {
	res := slavedb.exec("show slave status")
	if len(res) == 0 {
		log.Printf("no slave status for slave [%v:%v]\n", slavedb.host, slavedb.port)
		return
	}
	slaveIoState := res[0]["Slave_IO_State"]
	slaveIoRunning := res[0]["Slave_IO_Running"]
	slaveSqlRunning := res[0]["Slave_SQL_Running"]
	slaveSqlRunningState := res[0]["Slave_SQL_Running_State"]
	masterServerId := res[0]["Master_Server_Id"]
	slaveSecondsBehindMaster := res[0]["Seconds_Behind_Master"]
	lastIoError := res[0]["Last_IO_Error"]
	log.Printf("\n[slave: %v:%v]\nlast io err: %v\nslave io state: %v\nslave io running: %v\nslave sql running: %v\nslave sql running state: %v\nmaster server id: %v\nseconds behind master: %v\nslave status: %v\n",
		slavedb.host, slavedb.port, lastIoError, slaveIoState, slaveIoRunning, slaveSqlRunning, slaveSqlRunningState, masterServerId, slaveSecondsBehindMaster, res)
}

// stopSlave stops slave
func stopSlave(slavedb *Database) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln(err)
		}
	}()
	slavedb.exec("stop slave")
	slaveStatus(slavedb)
}

// startSlave starts slave
func startSlave(masterdb *Database, slavedb *Database) {
	var res = make([]map[string]string, 0)
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln(err)
		}
	}()
	stopSlave(slavedb)
	res = masterdb.exec("show master status")
	pos := res[0]["Position"]
	file := res[0]["File"]
	log.Println("file:", file, ", position:", pos, ", master status:", res)
	res = slavedb.exec("stop slave")
	res = slavedb.exec(
		"change master to master_host='%v', master_port=%v, master_user='%v', master_password='%v', master_log_file='%v', master_log_pos=%v;",
		masterdb.host, masterdb.port, masterdb.slaveUser, masterdb.slavePassword, file, pos,
	)
	res = slavedb.exec("start slave")
	slaveStatus(slavedb)
}
