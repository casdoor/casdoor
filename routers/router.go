// Copyright 2020 The casbin Authors. All Rights Reserved.
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
	"github.com/astaxie/beego"

	"github.com/casdoor/casdoor/controllers"
)

func init() {
	initAPI()
}

func initAPI() {
	ns :=
		beego.NewNamespace("/api",
			beego.NSInclude(
				&controllers.ApiController{},
			),
		)
	beego.AddNamespace(ns)

	beego.Router("/api/register", &controllers.ApiController{}, "POST:Register")
	beego.Router("/api/login", &controllers.ApiController{}, "POST:Login")
	beego.Router("/api/logout", &controllers.ApiController{}, "POST:Logout")
	beego.Router("/api/get-account", &controllers.ApiController{}, "GET:GetAccount")
	beego.Router("/api/auth/github", &controllers.ApiController{}, "GET:AuthGithub")

	beego.Router("/api/get-organizations", &controllers.ApiController{}, "GET:GetOrganizations")
	beego.Router("/api/get-organization", &controllers.ApiController{}, "GET:GetOrganization")
	beego.Router("/api/update-organization", &controllers.ApiController{}, "POST:UpdateOrganization")
	beego.Router("/api/add-organization", &controllers.ApiController{}, "POST:AddOrganization")
	beego.Router("/api/delete-organization", &controllers.ApiController{}, "POST:DeleteOrganization")

	beego.Router("/api/get-global-users", &controllers.ApiController{}, "GET:GetGlobalUsers")
	beego.Router("/api/get-users", &controllers.ApiController{}, "GET:GetUsers")
	beego.Router("/api/get-user", &controllers.ApiController{}, "GET:GetUser")
	beego.Router("/api/update-user", &controllers.ApiController{}, "POST:UpdateUser")
	beego.Router("/api/add-user", &controllers.ApiController{}, "POST:AddUser")
	beego.Router("/api/delete-user", &controllers.ApiController{}, "POST:DeleteUser")

	beego.Router("/api/get-providers", &controllers.ApiController{}, "GET:GetProviders")
	beego.Router("/api/get-provider", &controllers.ApiController{}, "GET:GetProvider")
	beego.Router("/api/update-provider", &controllers.ApiController{}, "POST:UpdateProvider")
	beego.Router("/api/add-provider", &controllers.ApiController{}, "POST:AddProvider")
	beego.Router("/api/delete-provider", &controllers.ApiController{}, "POST:DeleteProvider")

	beego.Router("/api/get-applications", &controllers.ApiController{}, "GET:GetApplications")
	beego.Router("/api/get-application", &controllers.ApiController{}, "GET:GetApplication")
	beego.Router("/api/update-application", &controllers.ApiController{}, "POST:UpdateApplication")
	beego.Router("/api/add-application", &controllers.ApiController{}, "POST:AddApplication")
	beego.Router("/api/delete-application", &controllers.ApiController{}, "POST:DeleteApplication")
}
