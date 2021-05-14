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

// @APIVersion 1.0.0
// @Title Casdoor API
// @Description Documentation of Casdoor API
// @Contact admin@casbin.org
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
			beego.NSNamespace("/api",
				beego.NSInclude(
					&controllers.ApiController{},
				),
			),
		)
	beego.AddNamespace(ns)

	beego.Router("/api/signup", &controllers.ApiController{}, "POST:Signup")
	beego.Router("/api/login", &controllers.ApiController{}, "POST:Login")
	beego.Router("/api/get-app-login", &controllers.ApiController{}, "GET:GetApplicationLogin")
	beego.Router("/api/logout", &controllers.ApiController{}, "POST:Logout")
	beego.Router("/api/get-account", &controllers.ApiController{}, "GET:GetAccount")
	beego.Router("/api/unlink", &controllers.ApiController{}, "POST:Unlink")

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
	beego.Router("/api/upload-avatar", &controllers.ApiController{}, "POST:UploadAvatar")
	beego.Router("/api/set-password", &controllers.ApiController{}, "POST:SetPassword")
	beego.Router("/api/send-verification-code", &controllers.ApiController{}, "POST:SendVerificationCode")
	beego.Router("/api/reset-email-or-phone", &controllers.ApiController{}, "POST:ResetEmailOrPhone")

	beego.Router("/api/get-providers", &controllers.ApiController{}, "GET:GetProviders")
	beego.Router("/api/get-provider", &controllers.ApiController{}, "GET:GetProvider")
	beego.Router("/api/update-provider", &controllers.ApiController{}, "POST:UpdateProvider")
	beego.Router("/api/add-provider", &controllers.ApiController{}, "POST:AddProvider")
	beego.Router("/api/delete-provider", &controllers.ApiController{}, "POST:DeleteProvider")

	beego.Router("/api/get-applications", &controllers.ApiController{}, "GET:GetApplications")
	beego.Router("/api/get-application", &controllers.ApiController{}, "GET:GetApplication")
	beego.Router("/api/get-default-application", &controllers.ApiController{}, "GET:GetDefaultApplication")
	beego.Router("/api/update-application", &controllers.ApiController{}, "POST:UpdateApplication")
	beego.Router("/api/add-application", &controllers.ApiController{}, "POST:AddApplication")
	beego.Router("/api/delete-application", &controllers.ApiController{}, "POST:DeleteApplication")

	beego.Router("/api/get-tokens", &controllers.ApiController{}, "GET:GetTokens")
	beego.Router("/api/get-token", &controllers.ApiController{}, "GET:GetToken")
	beego.Router("/api/update-token", &controllers.ApiController{}, "POST:UpdateToken")
	beego.Router("/api/add-token", &controllers.ApiController{}, "POST:AddToken")
	beego.Router("/api/delete-token", &controllers.ApiController{}, "POST:DeleteToken")
	beego.Router("/api/login/oauth/access_token", &controllers.ApiController{}, "POST:GetOAuthToken")
}
