// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package routers

import (
	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type delegatedApi struct {
	method string
	// the action the delegated admin needs, empty for the reads every delegated admin may do
	action string
	// the API takes the application's owner "admin" in place of the organization
	application bool
}

// delegatedApis are the only APIs a delegated admin may call beyond what every user may. The filter pins every
// object of the request to the delegated admin's organization, the controllers then check the objects they really
// read or change against the administered groups and applications
var delegatedApis = map[string]delegatedApi{
	"/api/get-users":              {method: "GET"},
	"/api/get-sorted-users":       {method: "GET"},
	"/api/get-user-count":         {method: "GET"},
	"/api/get-user":               {method: "GET"},
	"/api/add-user":               {method: "POST", action: object.DelegatedActionUsers},
	"/api/update-user":            {method: "POST", action: object.DelegatedActionUsers},
	"/api/delete-user":            {method: "POST", action: object.DelegatedActionUsers},
	"/api/remove-user-from-group": {method: "POST", action: object.DelegatedActionUsers},
	"/api/set-password":           {method: "POST", action: object.DelegatedActionPasswords},
	"/api/get-groups":             {method: "GET"},
	"/api/get-group":              {method: "GET"},
	"/api/add-group":              {method: "POST", action: object.DelegatedActionGroups},
	"/api/update-group":           {method: "POST", action: object.DelegatedActionGroups},
	"/api/delete-group":           {method: "POST", action: object.DelegatedActionGroups},
	"/api/get-applications":       {method: "GET", application: true},
	"/api/get-application":        {method: "GET", application: true},
	"/api/update-application":     {method: "POST", action: object.DelegatedActionApplications, application: true},
	"/api/get-organization":       {method: "GET"},
}

// checkDelegatedAdmin returns the scope of a delegated admin's request to one of the delegated APIs, nil when the
// request isn't one or is outside the delegated admin's organization
func checkDelegatedAdmin(ctx *context.Context, subOwner string, subName string, method string, urlPath string, objects []Object) (*object.DelegatedScope, error) {
	api, ok := delegatedApis[urlPath]
	if !ok || api.method != method || subOwner == "" || subOwner == "anonymous" || subOwner == "app" || subOwner == "built-in" {
		return nil, nil
	}
	if conf.IsDemoMode() && method != "GET" {
		return nil, nil
	}
	if impersonating, _ := ctx.Input.GetData("impersonating").(bool); impersonating {
		return nil, nil
	}
	if !object.GetCachedEnterprisePolicy(subOwner).IsDelegatedAdmin(subName) {
		return nil, nil
	}

	user, err := object.GetUser(util.GetId(subOwner, subName))
	if err != nil || user == nil {
		return nil, err
	}
	scope, err := object.GetDelegatedScope(user)
	if err != nil || scope == nil {
		return nil, err
	}
	if api.action != "" && !scope.Can(api.action) {
		return nil, nil
	}

	for _, obj := range objects {
		if obj.Owner == "" || obj.Owner == scope.Organization || (api.application && obj.Owner == "admin") {
			continue
		}
		return nil, nil
	}
	for _, param := range []string{"owner", "organization"} {
		value := ctx.Input.Query(param)
		if value == "" || value == scope.Organization || (api.application && param == "owner" && value == "admin") {
			continue
		}
		return nil, nil
	}
	if id := ctx.Input.Query("id"); id != "" {
		owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
		if err != nil {
			return nil, nil
		}
		if urlPath == "/api/get-organization" {
			if owner != "admin" || name != scope.Organization {
				return nil, nil
			}
		} else if owner != scope.Organization && !(api.application && owner == "admin") {
			return nil, nil
		}
	}

	ctx.Input.SetData(controllers.DelegatedScopeKey, scope)
	return scope, nil
}
