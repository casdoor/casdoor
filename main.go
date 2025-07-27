// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/beego/beego"
	"github.com/beego/beego/logs"
	_ "github.com/beego/beego/session/redis"
	"github.com/casdoor/casdoor/authz"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/casdoor/radius"
	"github.com/casdoor/casdoor/routers"
	"github.com/casdoor/casdoor/util"
)

func main() {
	// Load environment variables from .env if present
	_ = godotenv.Load()
	if len(os.Args) > 1 && os.Args[1] == "init" {
		object.InitInitFlag()
	} else {
		object.InitFlag()
	}

	object.InitAdapter()
	object.CreateTables()

	// 初始化内置组织对象
	object.InitDb()
	object.InitDefaultStorageProvider()
	object.InitLdapAutoSynchronizer()
	proxy.InitHttpClient()
	// 初始化接口权限
	authz.InitApi()
	// 初始化enforcer
	object.InitUserManager()
	// object.InitCasvisorConfig() // Disabled: Uncomment if Casvisor integration is required
	object.InitCleanupTokens()

	if len(os.Args) > 1 && os.Args[1] == "init" {
		object.InitFromFile()
		return
	}

	// util.SafeGoroutine(func() { object.RunSyncUsersJob() })
	util.SafeGoroutine(func() { controllers.InitCLIDownloader() })

	// beego.DelStaticPath("/static")
	// beego.SetStaticPath("/static", "web/build/static")

	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.SetStaticPath("/swagger", "swagger")
	beego.SetStaticPath("/files", "files")
	// https://studygolang.com/articles/2303
	beego.InsertFilter("*", beego.BeforeRouter, routers.StaticFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.AutoSigninFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.CorsFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.TimeoutFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.ApiFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.PrometheusFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.RecordMessage)
	beego.InsertFilter("*", beego.BeforeRouter, routers.FieldValidationFilter)
	beego.InsertFilter("*", beego.AfterExec, routers.AfterRecordMessage, false)

	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.BConfig.WebConfig.Session.SessionName = "casdoor_session_id"
	if conf.GetConfigString("redisEndpoint") == "" {
		beego.BConfig.WebConfig.Session.SessionProvider = "file"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = "./tmp"
	} else {
		beego.BConfig.WebConfig.Session.SessionProvider = "redis"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = conf.GetConfigString("redisEndpoint")
	}
	beego.BConfig.WebConfig.Session.SessionCookieLifeTime = 3600 * 24 * 30
	beego.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600 * 24 * 30
	// beego.BConfig.WebConfig.Session.SessionCookieSameSite = http.SameSiteNoneMode

	var logAdapter string
	logConfigMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(conf.GetConfigString("logConfig")), &logConfigMap)
	if err != nil {
		panic(err)
	}
	_, ok := logConfigMap["adapter"]
	if !ok {
		logAdapter = "file"
	} else {
		logAdapter = logConfigMap["adapter"].(string)
	}
	if logAdapter == "console" {
		logs.Reset()
	}
	err = logs.SetLogger(logAdapter, conf.GetConfigString("logConfig"))
	if err != nil {
		panic(err)
	}

	port := beego.AppConfig.DefaultInt("httpport", 8000)
	// logs.SetLevel(logs.LevelInformational)
	logs.SetLogFuncCall(false)
	err = util.StopOldInstance(port)
	if err != nil {
		panic(err)
	}

	// go ldap.StartLdapServer()
	go radius.StartRadiusServer()
	go object.ClearThroughputPerSecond()

	beego.Info("starting...")
	logs.Info("Casdoor is starting...")
	beego.Run(fmt.Sprintf(":%v", port))
}
