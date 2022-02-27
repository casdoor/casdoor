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

// Package routers
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
		beego.NewNamespace("/",
			beego.NSNamespace("/api",
				beego.NSInclude(
					&controllers.ApiController{},
				),
			),
			beego.NSNamespace("",
				beego.NSInclude(
					&controllers.RootController{},
				),
			),
		)
	beego.AddNamespace(ns)

	beego.Router("/api/signup", &controllers.ApiController{}, "POST:Signup")
	beego.Router("/api/login", &controllers.ApiController{}, "POST:Login")
	beego.Router("/api/get-app-login", &controllers.ApiController{}, "GET:GetApplicationLogin")
	beego.Router("/api/logout", &controllers.ApiController{}, "POST:Logout")
	beego.Router("/api/get-account", &controllers.ApiController{}, "GET:GetAccount")
	beego.Router("/api/userinfo", &controllers.ApiController{}, "GET:GetUserinfo")
	beego.Router("/api/unlink", &controllers.ApiController{}, "POST:Unlink")
	beego.Router("/api/get-saml-login", &controllers.ApiController{}, "GET:GetSamlLogin")
	beego.Router("/api/acs", &controllers.ApiController{}, "POST:HandleSamlLogin")

	beego.Router("/api/get-organizations", &controllers.ApiController{}, "GET:GetOrganizations")
	beego.Router("/api/get-organization", &controllers.ApiController{}, "GET:GetOrganization")
	beego.Router("/api/update-organization", &controllers.ApiController{}, "POST:UpdateOrganization")
	beego.Router("/api/add-organization", &controllers.ApiController{}, "POST:AddOrganization")
	beego.Router("/api/delete-organization", &controllers.ApiController{}, "POST:DeleteOrganization")

	beego.Router("/api/get-global-users", &controllers.ApiController{}, "GET:GetGlobalUsers")
	beego.Router("/api/get-users", &controllers.ApiController{}, "GET:GetUsers")
	beego.Router("/api/get-sorted-users", &controllers.ApiController{}, "GET:GetSortedUsers")
	beego.Router("/api/get-user-count", &controllers.ApiController{}, "GET:GetUserCount")
	beego.Router("/api/get-user", &controllers.ApiController{}, "GET:GetUser")
	beego.Router("/api/update-user", &controllers.ApiController{}, "POST:UpdateUser")
	beego.Router("/api/add-user", &controllers.ApiController{}, "POST:AddUser")
	beego.Router("/api/delete-user", &controllers.ApiController{}, "POST:DeleteUser")
	beego.Router("/api/upload-users", &controllers.ApiController{}, "POST:UploadUsers")

	beego.Router("/api/get-roles", &controllers.ApiController{}, "GET:GetRoles")
	beego.Router("/api/get-role", &controllers.ApiController{}, "GET:GetRole")
	beego.Router("/api/update-role", &controllers.ApiController{}, "POST:UpdateRole")
	beego.Router("/api/add-role", &controllers.ApiController{}, "POST:AddRole")
	beego.Router("/api/delete-role", &controllers.ApiController{}, "POST:DeleteRole")

	beego.Router("/api/get-permissions", &controllers.ApiController{}, "GET:GetPermissions")
	beego.Router("/api/get-permission", &controllers.ApiController{}, "GET:GetPermission")
	beego.Router("/api/update-permission", &controllers.ApiController{}, "POST:UpdatePermission")
	beego.Router("/api/add-permission", &controllers.ApiController{}, "POST:AddPermission")
	beego.Router("/api/delete-permission", &controllers.ApiController{}, "POST:DeletePermission")

	beego.Router("/api/set-password", &controllers.ApiController{}, "POST:SetPassword")
	beego.Router("/api/check-user-password", &controllers.ApiController{}, "POST:CheckUserPassword")
	beego.Router("/api/get-email-and-phone", &controllers.ApiController{}, "POST:GetEmailAndPhone")
	beego.Router("/api/send-verification-code", &controllers.ApiController{}, "POST:SendVerificationCode")
	beego.Router("/api/reset-email-or-phone", &controllers.ApiController{}, "POST:ResetEmailOrPhone")
	beego.Router("/api/get-human-check", &controllers.ApiController{}, "GET:GetHumanCheck")

	beego.Router("/api/get-ldap-user", &controllers.ApiController{}, "POST:GetLdapUser")
	beego.Router("/api/get-ldaps", &controllers.ApiController{}, "POST:GetLdaps")
	beego.Router("/api/get-ldap", &controllers.ApiController{}, "POST:GetLdap")
	beego.Router("/api/add-ldap", &controllers.ApiController{}, "POST:AddLdap")
	beego.Router("/api/update-ldap", &controllers.ApiController{}, "POST:UpdateLdap")
	beego.Router("/api/delete-ldap", &controllers.ApiController{}, "POST:DeleteLdap")
	beego.Router("/api/check-ldap-users-exist", &controllers.ApiController{}, "POST:CheckLdapUsersExist")
	beego.Router("/api/sync-ldap-users", &controllers.ApiController{}, "POST:SyncLdapUsers")

	beego.Router("/api/get-providers", &controllers.ApiController{}, "GET:GetProviders")
	beego.Router("/api/get-provider", &controllers.ApiController{}, "GET:GetProvider")
	beego.Router("/api/update-provider", &controllers.ApiController{}, "POST:UpdateProvider")
	beego.Router("/api/add-provider", &controllers.ApiController{}, "POST:AddProvider")
	beego.Router("/api/delete-provider", &controllers.ApiController{}, "POST:DeleteProvider")

	beego.Router("/api/get-applications", &controllers.ApiController{}, "GET:GetApplications")
	beego.Router("/api/get-application", &controllers.ApiController{}, "GET:GetApplication")
	beego.Router("/api/get-user-application", &controllers.ApiController{}, "GET:GetUserApplication")
	beego.Router("/api/update-application", &controllers.ApiController{}, "POST:UpdateApplication")
	beego.Router("/api/add-application", &controllers.ApiController{}, "POST:AddApplication")
	beego.Router("/api/delete-application", &controllers.ApiController{}, "POST:DeleteApplication")

	beego.Router("/api/get-resources", &controllers.ApiController{}, "GET:GetResources")
	beego.Router("/api/get-resource", &controllers.ApiController{}, "GET:GetResource")
	beego.Router("/api/update-resource", &controllers.ApiController{}, "POST:UpdateResource")
	beego.Router("/api/add-resource", &controllers.ApiController{}, "POST:AddResource")
	beego.Router("/api/delete-resource", &controllers.ApiController{}, "POST:DeleteResource")
	beego.Router("/api/upload-resource", &controllers.ApiController{}, "POST:UploadResource")

	beego.Router("/api/get-tokens", &controllers.ApiController{}, "GET:GetTokens")
	beego.Router("/api/get-token", &controllers.ApiController{}, "GET:GetToken")
	beego.Router("/api/update-token", &controllers.ApiController{}, "POST:UpdateToken")
	beego.Router("/api/add-token", &controllers.ApiController{}, "POST:AddToken")
	beego.Router("/api/delete-token", &controllers.ApiController{}, "POST:DeleteToken")
	beego.Router("/api/login/oauth/code", &controllers.ApiController{}, "POST:GetOAuthCode")
	beego.Router("/api/login/oauth/access_token", &controllers.ApiController{}, "POST:GetOAuthToken")
	beego.Router("/api/login/oauth/refresh_token", &controllers.ApiController{}, "POST:RefreshToken")

	beego.Router("/api/get-records", &controllers.ApiController{}, "GET:GetRecords")
	beego.Router("/api/get-records-filter", &controllers.ApiController{}, "POST:GetRecordsByFilter")

	beego.Router("/api/get-webhooks", &controllers.ApiController{}, "GET:GetWebhooks")
	beego.Router("/api/get-webhook", &controllers.ApiController{}, "GET:GetWebhook")
	beego.Router("/api/update-webhook", &controllers.ApiController{}, "POST:UpdateWebhook")
	beego.Router("/api/add-webhook", &controllers.ApiController{}, "POST:AddWebhook")
	beego.Router("/api/delete-webhook", &controllers.ApiController{}, "POST:DeleteWebhook")

	beego.Router("/api/get-syncers", &controllers.ApiController{}, "GET:GetSyncers")
	beego.Router("/api/get-syncer", &controllers.ApiController{}, "GET:GetSyncer")
	beego.Router("/api/update-syncer", &controllers.ApiController{}, "POST:UpdateSyncer")
	beego.Router("/api/add-syncer", &controllers.ApiController{}, "POST:AddSyncer")
	beego.Router("/api/delete-syncer", &controllers.ApiController{}, "POST:DeleteSyncer")

	beego.Router("/api/get-certs", &controllers.ApiController{}, "GET:GetCerts")
	beego.Router("/api/get-cert", &controllers.ApiController{}, "GET:GetCert")
	beego.Router("/api/update-cert", &controllers.ApiController{}, "POST:UpdateCert")
	beego.Router("/api/add-cert", &controllers.ApiController{}, "POST:AddCert")
	beego.Router("/api/delete-cert", &controllers.ApiController{}, "POST:DeleteCert")

	beego.Router("/api/get-payments", &controllers.ApiController{}, "GET:GetPayments")
	beego.Router("/api/get-payment", &controllers.ApiController{}, "GET:GetPayment")
	beego.Router("/api/update-payment", &controllers.ApiController{}, "POST:UpdatePayment")
	beego.Router("/api/add-payment", &controllers.ApiController{}, "POST:AddPayment")
	beego.Router("/api/delete-payment", &controllers.ApiController{}, "POST:DeletePayment")

	beego.Router("/api/send-email", &controllers.ApiController{}, "POST:SendEmail")
	beego.Router("/api/send-sms", &controllers.ApiController{}, "POST:SendSms")

	beego.Router("api/init-totp", &controllers.ApiController{}, "GET:InitTOTP")
	beego.Router("/.well-known/openid-configuration", &controllers.RootController{}, "GET:GetOidcDiscovery")
	beego.Router("/api/certs", &controllers.RootController{}, "*:GetOidcCert")
}
