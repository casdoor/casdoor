// Copyright 2021 The casbin Authors. All Rights Reserved.
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

package original

import (
	"github.com/astaxie/beego"
	"github.com/casbin/casdoor/object"
	_ "github.com/go-sql-driver/mysql" // db = mysql
	//_ "github.com/lib/pq"              // db = postgres
)

var adapter *object.Adapter

func initConfig() {
	err := beego.LoadAppConfig("ini", "../conf/app.conf")
	if err != nil {
		panic(err)
	}
}

func InitAdapter() bool {
	if dbName == "dbName" {
		adapter = nil
		return false
	}

	adapter = object.NewAdapter(beego.AppConfig.String("driverName"), beego.AppConfig.String("dataSourceName"), dbName)
	createTable(adapter)
	return true
}

func createTable(a *object.Adapter) {
	err := a.Engine.Sync2(new(User))
	if err != nil {
		panic(err)
	}
}
