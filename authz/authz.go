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
	stringadapter "github.com/qiangmzsx/string-adapter/v2"
)

var Enforcer *casbin.Enforcer

func InitAuthz() {
	var err error

	a, err := xormadapter.NewAdapter("mysql", beego.AppConfig.String("dataSourceName")+beego.AppConfig.String("dbName"), true)
	if err != nil {
		panic(err)
	}

	modelText := `
[request_definition]
r = userId, method, urlPath, objOwner, objName

[policy_definition]
p = userId, method, urlPath, objOwner

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.method == "GET" || r.userId == r.objOwner
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		panic(err)
	}

	Enforcer, err = casbin.NewEnforcer(m, a)
	if err != nil {
		panic(err)
	}

	if len(Enforcer.GetPolicy()) == 0 {
		ruleText := `
p, 1, 2, 3, 4
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

func IsAllowed(userId string, method string, urlPath string, objOwner string, objName string) bool {
	res, err := Enforcer.Enforce(userId, method, urlPath, objOwner, objName)
	if err != nil {
		panic(err)
	}

	return res
}
