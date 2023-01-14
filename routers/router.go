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

// Package routers
// @APIVersion 1.0.0
// @Title Casdoor API
// @Description Documentation of Casdoor API
// @Contact admin@casbin.org
package routers

import (
	"github.com/beego/beego"

	"github.com/casdoor/casdoor/controllers"
)

func init() {
	initAPI()
}

func initAPI() {
	ns := beego.NewNamespace("/",
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
	beego.Router("/api/logout", &controllers.ApiController{}, "GET,POST:Logout")
	beego.Router("/api/get-account", &controllers.ApiController{}, "GET:GetAccount")
	beego.Router("/api/userinfo", &controllers.ApiController{}, "GET:GetUserinfo")
	beego.Router("/api/unlink", &controllers.ApiController{}, "POST:Unlink")
	beego.Router("/api/get-saml-login", &controllers.ApiController{}, "GET:GetSamlLogin")
	beego.Router("/api/acs", &controllers.ApiController{}, "POST:HandleSamlLogin")
	beego.Router("/api/saml/metadata", &controllers.ApiController{}, "GET:GetSamlMeta")
	beego.Router("/api/webhook", &controllers.ApiController{}, "POST:HandleOfficialAccountEvent")
	beego.Router("/api/get-webhook-event", &controllers.ApiController{}, "GET:GetWebhookEventType")

	beego.Router("/api/get-organizations", &controllers.ApiController{}, "GET:GetOrganizations")
	beego.Router("/api/get-organization", &controllers.ApiController{}, "GET:GetOrganization")
	beego.Router("/api/update-organization", &controllers.ApiController{}, "POST:UpdateOrganization")
	beego.Router("/api/add-organization", &controllers.ApiController{}, "POST:AddOrganization")
	beego.Router("/api/delete-organization", &controllers.ApiController{}, "POST:DeleteOrganization")
	beego.Router("/api/get-default-application", &controllers.ApiController{}, "GET:GetDefaultApplication")

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
	beego.Router("/api/get-permissions-by-submitter", &controllers.ApiController{}, "GET:GetPermissionsBySubmitter")
	beego.Router("/api/get-permissions-by-role", &controllers.ApiController{}, "GET:GetPermissionsByRole")
	beego.Router("/api/get-permission", &controllers.ApiController{}, "GET:GetPermission")
	beego.Router("/api/update-permission", &controllers.ApiController{}, "POST:UpdatePermission")
	beego.Router("/api/add-permission", &controllers.ApiController{}, "POST:AddPermission")
	beego.Router("/api/delete-permission", &controllers.ApiController{}, "POST:DeletePermission")

	beego.Router("/api/enforce", &controllers.ApiController{}, "POST:Enforce")
	beego.Router("/api/batch-enforce", &controllers.ApiController{}, "POST:BatchEnforce")
	beego.Router("/api/get-all-objects", &controllers.ApiController{}, "GET:GetAllObjects")
	beego.Router("/api/get-all-actions", &controllers.ApiController{}, "GET:GetAllActions")
	beego.Router("/api/get-all-roles", &controllers.ApiController{}, "GET:GetAllRoles")

	beego.Router("/api/get-models", &controllers.ApiController{}, "GET:GetModels")
	beego.Router("/api/get-model", &controllers.ApiController{}, "GET:GetModel")
	beego.Router("/api/update-model", &controllers.ApiController{}, "POST:UpdateModel")
	beego.Router("/api/add-model", &controllers.ApiController{}, "POST:AddModel")
	beego.Router("/api/delete-model", &controllers.ApiController{}, "POST:DeleteModel")

	beego.Router("/api/get-adapters", &controllers.ApiController{}, "GET:GetCasbinAdapters")
	beego.Router("/api/get-adapter", &controllers.ApiController{}, "GET:GetCasbinAdapter")
	beego.Router("/api/update-adapter", &controllers.ApiController{}, "POST:UpdateCasbinAdapter")
	beego.Router("/api/add-adapter", &controllers.ApiController{}, "POST:AddCasbinAdapter")
	beego.Router("/api/delete-adapter", &controllers.ApiController{}, "POST:DeleteCasbinAdapter")
	beego.Router("/api/sync-policies", &controllers.ApiController{}, "GET:SyncPolicies")
	beego.Router("/api/update-policy", &controllers.ApiController{}, "POST:UpdatePolicy")
	beego.Router("/api/add-policy", &controllers.ApiController{}, "POST:AddPolicy")
	beego.Router("/api/remove-policy", &controllers.ApiController{}, "POST:RemovePolicy")

	beego.Router("/api/set-password", &controllers.ApiController{}, "POST:SetPassword")
	beego.Router("/api/check-user-password", &controllers.ApiController{}, "POST:CheckUserPassword")
	beego.Router("/api/get-email-and-phone", &controllers.ApiController{}, "POST:GetEmailAndPhone")
	beego.Router("/api/send-verification-code", &controllers.ApiController{}, "POST:SendVerificationCode")
	beego.Router("/api/verify-captcha", &controllers.ApiController{}, "POST:VerifyCaptcha")
	beego.Router("/api/reset-email-or-phone", &controllers.ApiController{}, "POST:ResetEmailOrPhone")
	beego.Router("/api/get-captcha", &controllers.ApiController{}, "GET:GetCaptcha")

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
	beego.Router("/api/get-global-providers", &controllers.ApiController{}, "GET:GetGlobalProviders")
	beego.Router("/api/update-provider", &controllers.ApiController{}, "POST:UpdateProvider")
	beego.Router("/api/add-provider", &controllers.ApiController{}, "POST:AddProvider")
	beego.Router("/api/delete-provider", &controllers.ApiController{}, "POST:DeleteProvider")

	beego.Router("/api/get-applications", &controllers.ApiController{}, "GET:GetApplications")
	beego.Router("/api/get-application", &controllers.ApiController{}, "GET:GetApplication")
	beego.Router("/api/get-user-application", &controllers.ApiController{}, "GET:GetUserApplication")
	beego.Router("/api/get-organization-applications", &controllers.ApiController{}, "GET:GetOrganizationApplications")
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
	beego.Router("/api/login/oauth/introspect", &controllers.ApiController{}, "POST:IntrospectToken")
	beego.Router("/api/login/oauth/logout", &controllers.ApiController{}, "GET:TokenLogout")

	beego.Router("/api/get-records", &controllers.ApiController{}, "GET:GetRecords")
	beego.Router("/api/get-records-filter", &controllers.ApiController{}, "POST:GetRecordsByFilter")
	beego.Router("/api/add-record", &controllers.ApiController{}, "POST:AddRecord")

	beego.Router("/api/get-sessions", &controllers.ApiController{}, "GET:GetSessions")
	beego.Router("/api/delete-session", &controllers.ApiController{}, "POST:DeleteSession")

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
	beego.Router("/api/run-syncer", &controllers.ApiController{}, "GET:RunSyncer")

	beego.Router("/api/get-certs", &controllers.ApiController{}, "GET:GetCerts")
	beego.Router("/api/get-cert", &controllers.ApiController{}, "GET:GetCert")
	beego.Router("/api/update-cert", &controllers.ApiController{}, "POST:UpdateCert")
	beego.Router("/api/add-cert", &controllers.ApiController{}, "POST:AddCert")
	beego.Router("/api/delete-cert", &controllers.ApiController{}, "POST:DeleteCert")

	beego.Router("/api/get-products", &controllers.ApiController{}, "GET:GetProducts")
	beego.Router("/api/get-product", &controllers.ApiController{}, "GET:GetProduct")
	beego.Router("/api/update-product", &controllers.ApiController{}, "POST:UpdateProduct")
	beego.Router("/api/add-product", &controllers.ApiController{}, "POST:AddProduct")
	beego.Router("/api/delete-product", &controllers.ApiController{}, "POST:DeleteProduct")
	beego.Router("/api/buy-product", &controllers.ApiController{}, "POST:BuyProduct")

	beego.Router("/api/get-payments", &controllers.ApiController{}, "GET:GetPayments")
	beego.Router("/api/get-user-payments", &controllers.ApiController{}, "GET:GetUserPayments")
	beego.Router("/api/get-payment", &controllers.ApiController{}, "GET:GetPayment")
	beego.Router("/api/update-payment", &controllers.ApiController{}, "POST:UpdatePayment")
	beego.Router("/api/add-payment", &controllers.ApiController{}, "POST:AddPayment")
	beego.Router("/api/delete-payment", &controllers.ApiController{}, "POST:DeletePayment")
	beego.Router("/api/notify-payment/?:owner/?:provider/?:product/?:payment", &controllers.ApiController{}, "POST:NotifyPayment")
	beego.Router("/api/invoice-payment", &controllers.ApiController{}, "POST:InvoicePayment")

	beego.Router("/api/send-email", &controllers.ApiController{}, "POST:SendEmail")
	beego.Router("/api/send-sms", &controllers.ApiController{}, "POST:SendSms")

	beego.Router("/.well-known/openid-configuration", &controllers.RootController{}, "GET:GetOidcDiscovery")
	beego.Router("/.well-known/jwks", &controllers.RootController{}, "*:GetJwks")

	beego.Router("/cas/:organization/:application/serviceValidate", &controllers.RootController{}, "GET:CasServiceValidate")
	beego.Router("/cas/:organization/:application/proxyValidate", &controllers.RootController{}, "GET:CasProxyValidate")
	beego.Router("/cas/:organization/:application/proxy", &controllers.RootController{}, "GET:CasProxy")
	beego.Router("/cas/:organization/:application/validate", &controllers.RootController{}, "GET:CasValidate")

	beego.Router("/cas/:organization/:application/p3/serviceValidate", &controllers.RootController{}, "GET:CasP3ServiceAndProxyValidate")
	beego.Router("/cas/:organization/:application/p3/proxyValidate", &controllers.RootController{}, "GET:CasP3ServiceAndProxyValidate")
	beego.Router("/cas/:organization/:application/samlValidate", &controllers.RootController{}, "POST:SamlValidate")

	beego.Router("/api/webauthn/signup/begin", &controllers.ApiController{}, "Get:WebAuthnSignupBegin")
	beego.Router("/api/webauthn/signup/finish", &controllers.ApiController{}, "Post:WebAuthnSignupFinish")
	beego.Router("/api/webauthn/signin/begin", &controllers.ApiController{}, "Get:WebAuthnSigninBegin")
	beego.Router("/api/webauthn/signin/finish", &controllers.ApiController{}, "Post:WebAuthnSigninFinish")

	beego.Router("/api/get-system-info", &controllers.ApiController{}, "GET:GetSystemInfo")
	beego.Router("/api/get-release", &controllers.ApiController{}, "GET:GitRepoVersion")

}
