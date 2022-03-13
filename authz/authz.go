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

package authz

import (
	"github.com/astaxie/beego"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casbin/xorm-adapter/v2"
	"github.com/casdoor/casdoor/conf"
	stringadapter "github.com/qiangmzsx/string-adapter/v2"
)

var Enforcer *casbin.Enforcer

func InitAuthz() {
	var err error

	tableNamePrefix := beego.AppConfig.String("tableNamePrefix")
	a, err := xormadapter.NewAdapterWithTableName(beego.AppConfig.String("driverName"), conf.GetBeegoConfDataSourceName()+beego.AppConfig.String("dbName"), "casbin_rule", tableNamePrefix, true)
	if err != nil {
		panic(err)
	}

	modelText := `
[request_definition]
r = subOwner, subName, method, urlPath, objOwner, objName

[policy_definition]
p = subOwner, subName, method, urlPath, objOwner, objName

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (r.subOwner == p.subOwner || p.subOwner == "*") && \
    (r.subName == p.subName || p.subName == "*" || r.subName != "anonymous" && p.subName == "!anonymous") && \
    (r.method == p.method || p.method == "*") && \
    (r.urlPath == p.urlPath || p.urlPath == "*") && \
    (r.objOwner == p.objOwner || p.objOwner == "*") && \
    (r.objName == p.objName || p.objName == "*") || \
    (r.urlPath == "/api/update-user" && r.subOwner == r.objOwner && r.subName == r.objName)
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		panic(err)
	}

	Enforcer, err = casbin.NewEnforcer(m, a)
	if err != nil {
		panic(err)
	}

	Enforcer.ClearPolicy()

	//if len(Enforcer.GetPolicy()) == 0 {
	if true {
		ruleText := `
p, built-in, *, *, *, *, *
p, app, *, *, *, *, *
p, *, *, POST, /api/signup, *, *
p, *, *, POST, /api/get-email-and-phone, *, *
p, *, *, POST, /api/login, *, *
p, *, *, GET, /api/get-app-login, *, *
p, *, *, POST, /api/logout, *, *
p, *, *, GET, /api/get-account, *, *
p, *, *, GET, /api/userinfo, *, *
p, *, *, POST, /api/login/oauth/access_token, *, *
p, *, *, POST, /api/login/oauth/refresh_token, *, *
p, *, *, GET, /api/get-application, *, *
p, *, *, GET, /api/get-users, *, *
p, *, *, GET, /api/get-user, *, *
p, *, *, GET, /api/get-organizations, *, *
p, *, *, GET, /api/get-user-application, *, *
p, *, *, GET, /api/get-default-providers, *, *
p, *, *, GET, /api/get-resources, *, *
p, *, *, POST, /api/upload-avatar, *, *
p, *, *, POST, /api/unlink, *, *
p, *, *, POST, /api/set-password, *, *
p, *, *, POST, /api/send-verification-code, *, *
p, *, *, GET, /api/get-human-check, *, *
p, *, *, POST, /api/reset-email-or-phone, *, *
p, *, *, POST, /api/upload-resource, *, *
p, *, *, GET, /.well-known/openid-configuration, *, *
p, *, *, *, /api/certs, *, *
p, *, *, GET, /api/get-saml-login, *, *
p, *, *, POST, /api/acs, *, *
p, *, *, POST, /api/two-factor/setup/totp/init, *, *
p, *, *, POST, /api/two-factor/setup/totp/verity, *, *
p, *, *, POST, /api/two-factor/auth/totp, *, *
`

		sa := stringadapter.NewAdapter(ruleText)
		// load all rules from string adapter to enforcer's memory
		err := sa.LoadPolicy(Enforcer.GetModel())
		if err != nil {
			panic(err)
		}

		// save all rules from enforcer's memory to Xorm adapter (DB)
		// same as:
		// a.SavePolicy(Enforcer.GetModel())
		err = Enforcer.SavePolicy()
		if err != nil {
			panic(err)
		}
	}
}

func IsAllowed(subOwner string, subName string, method string, urlPath string, objOwner string, objName string) bool {
	res, err := Enforcer.Enforce(subOwner, subName, method, urlPath, objOwner, objName)
	if err != nil {
		panic(err)
	}

	return res
}
