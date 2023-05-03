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
	"encoding/gob"
	"flag"
	"fmt"

	"github.com/google/uuid"

	"github.com/beego/beego"
	"github.com/beego/beego/logs"
	_ "github.com/beego/beego/session/redis"
	"github.com/casdoor/casdoor/authz"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/ldap"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/casdoor/routers"
	_ "github.com/casdoor/casdoor/routers"
	"github.com/casdoor/casdoor/util"
)

func main() {
	createDatabase := flag.Bool("createDatabase", false, "true if you need Casdoor to create database")
	flag.Parse()

	object.InitAdapter()
	object.DoMigration()
	object.CreateTables(*createDatabase)

	object.InitDb()
	object.InitFromFile()
	object.InitDefaultStorageProvider()
	object.InitLdapAutoSynchronizer()
	proxy.InitHttpClient()
	authz.InitAuthz()

	util.SafeGoroutine(func() { object.RunSyncUsersJob() })

	// beego.DelStaticPath("/static")
	// beego.SetStaticPath("/static", "web/build/static")

	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.SetStaticPath("/swagger", "swagger")
	beego.SetStaticPath("/files", "files")
	// https://studygolang.com/articles/2303
	beego.InsertFilter("*", beego.BeforeRouter, routers.StaticFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.AutoSigninFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.CorsFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.AuthzFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.RecordMessage)
	beego.InsertFilter("*", beego.BeforeRouter, routers.PrometheusFilter)

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
	// beego.BConfig.WebConfig.Session.SessionCookieSameSite = http.SameSiteNoneMode

	err := logs.SetLogger("file", `{"filename":"logs/casdoor.log","maxdays":99999,"perm":"0770"}`)
	if err != nil {
		panic(err)
	}
	port := beego.AppConfig.DefaultInt("httpport", 8000)
	// logs.SetLevel(logs.LevelInformational)
	logs.SetLogFuncCall(false)

	go ldap.StartLdapServer()
	go object.ClearThroughputPerSecond()

	gob.Register(uuid.UUID{})
	gob.Register([]uuid.UUID{})

	beego.Run(fmt.Sprintf(":%v", port))
}
