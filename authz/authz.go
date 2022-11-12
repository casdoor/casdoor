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

package authz

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casbin/xorm-adapter/v3"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	stringadapter "github.com/qiangmzsx/string-adapter/v2"
)

var Enforcer *casbin.Enforcer

func InitAuthz() {
	var err error

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	a, err := xormadapter.NewAdapterWithTableName(conf.GetConfigString("driverName"), conf.GetConfigDataSourceName()+conf.GetConfigString("dbName"), "casbin_rule", tableNamePrefix, true)
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
    (r.subOwner == r.objOwner && r.subName == r.objName)
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

	// if len(Enforcer.GetPolicy()) == 0 {
	if true {
		ruleText := `
p, built-in, *, *, *, *, *
p, app, *, *, *, *, *
p, *, *, POST, /api/signup, *, *
p, *, *, POST, /api/get-email-and-phone, *, *
p, *, *, POST, /api/login, *, *
p, *, *, GET, /api/get-app-login, *, *
p, *, *, POST, /api/logout, *, *
p, *, *, GET, /api/logout, *, *
p, *, *, GET, /api/get-account, *, *
p, *, *, GET, /api/userinfo, *, *
p, *, *, *, /api/login/oauth, *, *
p, *, *, GET, /api/get-application, *, *
p, *, *, GET, /api/get-user, *, *
p, *, *, GET, /api/get-user-application, *, *
p, *, *, GET, /api/get-resources, *, *
p, *, *, GET, /api/get-records, *, *
p, *, *, GET, /api/get-product, *, *
p, *, *, POST, /api/buy-product, *, *
p, *, *, GET, /api/get-payment, *, *
p, *, *, POST, /api/update-payment, *, *
p, *, *, POST, /api/invoice-payment, *, *
p, *, *, POST, /api/notify-payment, *, *
p, *, *, POST, /api/unlink, *, *
p, *, *, POST, /api/set-password, *, *
p, *, *, POST, /api/send-verification-code, *, *
p, *, *, GET, /api/get-captcha, *, *
p, *, *, POST, /api/verify-captcha, *, *
p, *, *, POST, /api/reset-email-or-phone, *, *
p, *, *, POST, /api/upload-resource, *, *
p, *, *, GET, /.well-known/openid-configuration, *, *
p, *, *, *, /.well-known/jwks, *, *
p, *, *, GET, /api/get-saml-login, *, *
p, *, *, POST, /api/acs, *, *
p, *, *, GET, /api/saml/metadata, *, *
p, *, *, *, /cas, *, *
p, *, *, *, /api/webauthn, *, *
p, *, *, GET, /api/get-release, *, *
p, *, *, GET, /api/get-default-application, *, *
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
	if conf.IsDemoMode() {
		if !isAllowedInDemoMode(subOwner, subName, method, urlPath, objOwner, objName) {
			return false
		}
	}

	userId := fmt.Sprintf("%s/%s", subOwner, subName)
	user := object.GetUser(userId)
	if user != nil && user.IsAdmin && (subOwner == objOwner || (objOwner == "admin")) {
		return true
	}

	res, err := Enforcer.Enforce(subOwner, subName, method, urlPath, objOwner, objName)
	if err != nil {
		panic(err)
	}

	return res
}

func isAllowedInDemoMode(subOwner string, subName string, method string, urlPath string, objOwner string, objName string) bool {
	if method == "POST" {
		if strings.HasPrefix(urlPath, "/api/login") || urlPath == "/api/logout" || urlPath == "/api/signup" || urlPath == "/api/send-verification-code" {
			return true
		} else if urlPath == "/api/update-user" {
			// Allow ordinary users to update their own information
			if subOwner == objOwner && subName == objName && !(subOwner == "built-in" && subName == "admin") {
				return true
			}
			return false
		} else {
			return false
		}
	}

	// If method equals GET
	return true
}
