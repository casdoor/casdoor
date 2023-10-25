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

package sync_v2

import "log"

// slaveStatus shows slave status
func slaveStatus(slavedb *Database) {
	res := slavedb.exec("show slave status")
	if len(res) == 0 {
		log.Printf("no slave status for slave [%v:%v]\n", slavedb.host, slavedb.port)
		return
	}
	log.Println("*****check slave status*****")
	log.Println("slave:", slavedb.host, ":", slavedb.port)
	masterServerId := res[0]["Master_Server_Id"]
	log.Println("master server id:", masterServerId)
	lastError := res[0]["Last_Error"]
	log.Println("last error:", lastError) // this should be empty
	lastIoError := res[0]["Last_IO_Error"]
	log.Println("last io error:", lastIoError) // this should be empty
	slaveIoState := res[0]["Slave_IO_State"]
	log.Println("slave io state:", slaveIoState)
	slaveIoRunning := res[0]["Slave_IO_Running"]
	log.Println("slave io running:", slaveIoRunning) // this should be Yes
	slaveSqlRunning := res[0]["Slave_SQL_Running"]
	log.Println("slave sql running:", slaveSqlRunning) // this should be Yes
	slaveSqlRunningState := res[0]["Slave_SQL_Running_State"]
	log.Println("slave sql running state:", slaveSqlRunningState)
	slaveSecondsBehindMaster := res[0]["Seconds_Behind_Master"]
	log.Println("seconds behind master:", slaveSecondsBehindMaster) // this should be 0, if not, it means the slave is behind the master
	log.Println("slave status:", res)
	log.Println("****************************")
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
	res := make([]map[string]string, 0)
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln(err)
		}
	}()
	stopSlave(slavedb)
	// get the info about master
	res = masterdb.exec("show master status")
	if len(res) == 0 {
		log.Println("no master status")
		return
	}
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
