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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/plugins/cors"
	_ "github.com/astaxie/beego/session/redis"
	"github.com/casbin/casdoor/authz"
	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/proxy"
	"github.com/casbin/casdoor/routers"
	_ "github.com/casbin/casdoor/routers"
	"github.com/casbin/casdoor/util"
	"github.com/go-ini/ini"
)

func main() {
	createDatabase := flag.Bool("createDatabase", false, "true if you need casdoor to create database")
	flag.Parse()
	object.InitAdapter(*createDatabase)
	object.InitDb()
	object.InitDefaultStorageProvider()
	object.InitLdapAutoSynchronizer()
	proxy.InitHttpClient()
	authz.InitAuthz()

	go object.RunSyncUsersJob()

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	//beego.DelStaticPath("/static")
	beego.SetStaticPath("/static", "web/build/static")
	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.SetStaticPath("/swagger", "swagger")
	beego.SetStaticPath("/files", "files")
	// https://studygolang.com/articles/2303
	beego.InsertFilter("*", beego.BeforeRouter, routers.StaticFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.AutoSigninFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.AuthzFilter)
	beego.InsertFilter("*", beego.BeforeRouter, routers.RecordMessage)

	beego.BConfig.WebConfig.Session.SessionName = "casdoor_session_id"
	if beego.AppConfig.String("redisEndpoint") == "" {
		beego.BConfig.WebConfig.Session.SessionProvider = "file"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = "./tmp"
	} else {
		beego.BConfig.WebConfig.Session.SessionProvider = "redis"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = beego.AppConfig.String("redisEndpoint")
	}
	beego.BConfig.WebConfig.Session.SessionCookieLifeTime = 3600 * 24 * 30
	//beego.BConfig.WebConfig.Session.SessionCookieSameSite = http.SameSiteNoneMode

	err := logs.SetLogger("file", `{"filename":"logs/casdoor.log","maxdays":99999,"perm":"0770"}`)
	if err != nil {
		panic(err)
	}
	//logs.SetLevel(logs.LevelInformational)
	logs.SetLogFuncCall(false)

	if casnodeConf := os.Getenv("CASNODE_CONF"); casnodeConf != "" {
		initCasnodeInformation(casnodeConf)
	}

	beego.Run()
}

func initCasnodeInformation(casnodeConfiguration string) {
	cfg, err := ini.Load(casnodeConfiguration)
	if err != nil {
		logs.Warning("failed to detect casnode configuration at " + casnodeConfiguration)
		return
	}
	casnodeEndpoint := cfg.Section("").Key("casnodeEndpoint").String()
	casdoorOrganization := cfg.Section("").Key("casdoorOrganization").String()
	casdoorApplication := cfg.Section("").Key("casdoorApplication").String()
	clientID := cfg.Section("").Key("clientId").String()
	clientSecret := cfg.Section("").Key("clientSecret").String()
	if casnodeEndpoint == "" || casdoorOrganization == "" || casdoorApplication == "" || clientID == "" || clientSecret == "" {
		logs.Warning("missing required fields in " + casnodeConfiguration)
		return
	}
	organization := &object.Organization{
		Owner:         "admin",
		Name:          casdoorOrganization,
		CreatedTime:   util.GetCurrentTime(),
		DisplayName:   casdoorOrganization,
		WebsiteUrl:    "https://example.com",
		Favicon:       "https://cdn.casbin.com/static/favicon.ico",
		PhonePrefix:   "86",
		DefaultAvatar: "https://casbin.org/img/casbin.svg",
		PasswordType:  "plain",
	}
	if org := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", casdoorOrganization)); org == nil {
		object.AddOrganization(organization)
	} else {
		logs.Warning("organization already exists")
		return
	}

	application := object.Application{
		Owner:                "admin",
		Name:                 casdoorApplication,
		CreatedTime:          util.GetCurrentTime(),
		DisplayName:          casdoorApplication,
		Logo:                 "https://cdn.casbin.com/logo/logo_1024x256.png",
		HomepageUrl:          "https://casdoor.org",
		Organization:         organization.Name,
		EnablePassword:       true,
		EnableSignUp:         true,
		EnableSigninSession:  true,
		ClientId:             clientID,
		ClientSecret:         clientSecret,
		RedirectUris:         []string{casnodeEndpoint},
		TokenFormat:          "JWT",
		ExpireInHours:        168,
		RefreshExpireInHours: 0,
		Providers:            []*object.ProviderItem{},
		SignupItems: []*object.SignupItem{
			{Name: "ID", Visible: false, Required: true, Prompted: false, Rule: "Random"},
			{Name: "Username", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Display name", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Password", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Confirm password", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Email", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Phone", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Agreement", Visible: true, Required: true, Prompted: false, Rule: "None"},
		},
	}
	if app := object.GetApplication(fmt.Sprintf("admin/%s", casdoorApplication)); app == nil {
		object.AddApplication(&application)
	} else {
		logs.Warning("application already exists")
		return
	}

	logs.Info("organization and applications for casnode created")

}
