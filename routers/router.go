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
// @APIVersion 1.503.0
// @Title Casdoor RESTful API
// @Description Swagger Docs of Casdoor Backend API
// @Contact casbin@googlegroups.com
// @SecurityDefinition AccessToken apiKey Authorization header
// @Schemes https,http
// @ExternalDocs Find out more about Casdoor
// @ExternalDocsUrl https://casdoor.org/
package routers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/casdoor/casdoor/controllers"
)

func InitAPI() {
	ns := web.NewNamespace("/",
		web.NSNamespace("/api",
			web.NSInclude(
				&controllers.ApiController{},
			),
		),
		web.NSNamespace("",
			web.NSInclude(
				&controllers.RootController{},
			),
		),
	)
	web.AddNamespace(ns)

	web.Router("/api/signup", &controllers.ApiController{}, "POST:Signup")
	web.Router("/api/login", &controllers.ApiController{}, "POST:Login")
	web.Router("/api/get-app-login", &controllers.ApiController{}, "GET:GetApplicationLogin")
	web.Router("/api/get-dashboard", &controllers.ApiController{}, "GET:GetDashboard")
	web.Router("/api/logout", &controllers.ApiController{}, "GET,POST:Logout")
	web.Router("/api/sso-logout", &controllers.ApiController{}, "GET,POST:SsoLogout")
	web.Router("/api/get-account", &controllers.ApiController{}, "GET:GetAccount")
	web.Router("/api/userinfo", &controllers.ApiController{}, "GET:GetUserinfo")
	web.Router("/api/user", &controllers.ApiController{}, "GET:GetUserinfo2")
	web.Router("/api/unlink", &controllers.ApiController{}, "POST:Unlink")
	web.Router("/api/get-saml-login", &controllers.ApiController{}, "GET:GetSamlLogin")
	web.Router("/api/acs", &controllers.ApiController{}, "POST:HandleSamlLogin")
	web.Router("/api/saml/metadata", &controllers.ApiController{}, "GET:GetSamlMeta")
	web.Router("/api/saml/redirect/:owner/:application", &controllers.ApiController{}, "*:HandleSamlRedirect")
	web.Router("/api/webhook", &controllers.ApiController{}, "*:HandleOfficialAccountEvent")
	web.Router("/api/get-qrcode", &controllers.ApiController{}, "GET:GetQRCode")
	web.Router("/api/get-webhook-event", &controllers.ApiController{}, "GET:GetWebhookEventType")
	web.Router("/api/get-captcha-status", &controllers.ApiController{}, "GET:GetCaptchaStatus")
	web.Router("/api/callback", &controllers.ApiController{}, "POST:Callback")
	web.Router("/api/device-auth", &controllers.ApiController{}, "POST:DeviceAuth")

	web.Router("/api/get-organizations", &controllers.ApiController{}, "GET:GetOrganizations")
	web.Router("/api/get-organization", &controllers.ApiController{}, "GET:GetOrganization")
	web.Router("/api/update-organization", &controllers.ApiController{}, "POST:UpdateOrganization")
	web.Router("/api/add-organization", &controllers.ApiController{}, "POST:AddOrganization")
	web.Router("/api/delete-organization", &controllers.ApiController{}, "POST:DeleteOrganization")
	web.Router("/api/get-default-application", &controllers.ApiController{}, "GET:GetDefaultApplication")
	web.Router("/api/get-organization-names", &controllers.ApiController{}, "GET:GetOrganizationNames")

	web.Router("/api/get-groups", &controllers.ApiController{}, "GET:GetGroups")
	web.Router("/api/get-group", &controllers.ApiController{}, "GET:GetGroup")
	web.Router("/api/update-group", &controllers.ApiController{}, "POST:UpdateGroup")
	web.Router("/api/add-group", &controllers.ApiController{}, "POST:AddGroup")
	web.Router("/api/delete-group", &controllers.ApiController{}, "POST:DeleteGroup")
	web.Router("/api/upload-groups", &controllers.ApiController{}, "POST:UploadGroups")

	web.Router("/api/get-global-users", &controllers.ApiController{}, "GET:GetGlobalUsers")
	web.Router("/api/get-users", &controllers.ApiController{}, "GET:GetUsers")
	web.Router("/api/get-sorted-users", &controllers.ApiController{}, "GET:GetSortedUsers")
	web.Router("/api/get-user-count", &controllers.ApiController{}, "GET:GetUserCount")
	web.Router("/api/get-user", &controllers.ApiController{}, "GET:GetUser")
	web.Router("/api/update-user", &controllers.ApiController{}, "POST:UpdateUser")
	web.Router("/api/add-user-keys", &controllers.ApiController{}, "POST:AddUserKeys")
	web.Router("/api/add-user", &controllers.ApiController{}, "POST:AddUser")
	web.Router("/api/delete-user", &controllers.ApiController{}, "POST:DeleteUser")
	web.Router("/api/upload-users", &controllers.ApiController{}, "POST:UploadUsers")
	web.Router("/api/remove-user-from-group", &controllers.ApiController{}, "POST:RemoveUserFromGroup")
	web.Router("/api/verify-identification", &controllers.ApiController{}, "POST:VerifyIdentification")

	web.Router("/api/get-invitations", &controllers.ApiController{}, "GET:GetInvitations")
	web.Router("/api/get-invitation", &controllers.ApiController{}, "GET:GetInvitation")
	web.Router("/api/get-invitation-info", &controllers.ApiController{}, "GET:GetInvitationCodeInfo")
	web.Router("/api/update-invitation", &controllers.ApiController{}, "POST:UpdateInvitation")
	web.Router("/api/add-invitation", &controllers.ApiController{}, "POST:AddInvitation")
	web.Router("/api/delete-invitation", &controllers.ApiController{}, "POST:DeleteInvitation")
	web.Router("/api/verify-invitation", &controllers.ApiController{}, "GET:VerifyInvitation")
	web.Router("/api/send-invitation", &controllers.ApiController{}, "POST:SendInvitation")

	web.Router("/api/get-applications", &controllers.ApiController{}, "GET:GetApplications")
	web.Router("/api/get-application", &controllers.ApiController{}, "GET:GetApplication")
	web.Router("/api/get-user-application", &controllers.ApiController{}, "GET:GetUserApplication")
	web.Router("/api/get-organization-applications", &controllers.ApiController{}, "GET:GetOrganizationApplications")
	web.Router("/api/update-application", &controllers.ApiController{}, "POST:UpdateApplication")
	web.Router("/api/add-application", &controllers.ApiController{}, "POST:AddApplication")
	web.Router("/api/delete-application", &controllers.ApiController{}, "POST:DeleteApplication")

	web.Router("/api/get-providers", &controllers.ApiController{}, "GET:GetProviders")
	web.Router("/api/get-provider", &controllers.ApiController{}, "GET:GetProvider")
	web.Router("/api/get-global-providers", &controllers.ApiController{}, "GET:GetGlobalProviders")
	web.Router("/api/update-provider", &controllers.ApiController{}, "POST:UpdateProvider")
	web.Router("/api/add-provider", &controllers.ApiController{}, "POST:AddProvider")
	web.Router("/api/delete-provider", &controllers.ApiController{}, "POST:DeleteProvider")

	web.Router("/api/get-resources", &controllers.ApiController{}, "GET:GetResources")
	web.Router("/api/get-resource", &controllers.ApiController{}, "GET:GetResource")
	web.Router("/api/update-resource", &controllers.ApiController{}, "POST:UpdateResource")
	web.Router("/api/add-resource", &controllers.ApiController{}, "POST:AddResource")
	web.Router("/api/delete-resource", &controllers.ApiController{}, "POST:DeleteResource")
	web.Router("/api/upload-resource", &controllers.ApiController{}, "POST:UploadResource")

	web.Router("/api/get-certs", &controllers.ApiController{}, "GET:GetCerts")
	web.Router("/api/get-global-certs", &controllers.ApiController{}, "GET:GetGlobalCerts")
	web.Router("/api/get-cert", &controllers.ApiController{}, "GET:GetCert")
	web.Router("/api/update-cert", &controllers.ApiController{}, "POST:UpdateCert")
	web.Router("/api/add-cert", &controllers.ApiController{}, "POST:AddCert")
	web.Router("/api/delete-cert", &controllers.ApiController{}, "POST:DeleteCert")

	web.Router("/api/get-roles", &controllers.ApiController{}, "GET:GetRoles")
	web.Router("/api/get-role", &controllers.ApiController{}, "GET:GetRole")
	web.Router("/api/update-role", &controllers.ApiController{}, "POST:UpdateRole")
	web.Router("/api/add-role", &controllers.ApiController{}, "POST:AddRole")
	web.Router("/api/delete-role", &controllers.ApiController{}, "POST:DeleteRole")
	web.Router("/api/upload-roles", &controllers.ApiController{}, "POST:UploadRoles")

	web.Router("/api/get-permissions", &controllers.ApiController{}, "GET:GetPermissions")
	web.Router("/api/get-permissions-by-submitter", &controllers.ApiController{}, "GET:GetPermissionsBySubmitter")
	web.Router("/api/get-permissions-by-role", &controllers.ApiController{}, "GET:GetPermissionsByRole")
	web.Router("/api/get-permission", &controllers.ApiController{}, "GET:GetPermission")
	web.Router("/api/update-permission", &controllers.ApiController{}, "POST:UpdatePermission")
	web.Router("/api/add-permission", &controllers.ApiController{}, "POST:AddPermission")
	web.Router("/api/delete-permission", &controllers.ApiController{}, "POST:DeletePermission")
	web.Router("/api/upload-permissions", &controllers.ApiController{}, "POST:UploadPermissions")

	web.Router("/api/get-models", &controllers.ApiController{}, "GET:GetModels")
	web.Router("/api/get-model", &controllers.ApiController{}, "GET:GetModel")
	web.Router("/api/update-model", &controllers.ApiController{}, "POST:UpdateModel")
	web.Router("/api/add-model", &controllers.ApiController{}, "POST:AddModel")
	web.Router("/api/delete-model", &controllers.ApiController{}, "POST:DeleteModel")

	web.Router("/api/get-adapters", &controllers.ApiController{}, "GET:GetAdapters")
	web.Router("/api/get-adapter", &controllers.ApiController{}, "GET:GetAdapter")
	web.Router("/api/update-adapter", &controllers.ApiController{}, "POST:UpdateAdapter")
	web.Router("/api/add-adapter", &controllers.ApiController{}, "POST:AddAdapter")
	web.Router("/api/delete-adapter", &controllers.ApiController{}, "POST:DeleteAdapter")
	web.Router("/api/get-policies", &controllers.ApiController{}, "GET:GetPolicies")
	web.Router("/api/get-filtered-policies", &controllers.ApiController{}, "POST:GetFilteredPolicies")
	web.Router("/api/update-policy", &controllers.ApiController{}, "POST:UpdatePolicy")
	web.Router("/api/add-policy", &controllers.ApiController{}, "POST:AddPolicy")
	web.Router("/api/remove-policy", &controllers.ApiController{}, "POST:RemovePolicy")

	web.Router("/api/get-enforcers", &controllers.ApiController{}, "GET:GetEnforcers")
	web.Router("/api/get-enforcer", &controllers.ApiController{}, "GET:GetEnforcer")
	web.Router("/api/update-enforcer", &controllers.ApiController{}, "POST:UpdateEnforcer")
	web.Router("/api/add-enforcer", &controllers.ApiController{}, "POST:AddEnforcer")
	web.Router("/api/delete-enforcer", &controllers.ApiController{}, "POST:DeleteEnforcer")

	web.Router("/api/enforce", &controllers.ApiController{}, "POST:Enforce")
	web.Router("/api/batch-enforce", &controllers.ApiController{}, "POST:BatchEnforce")
	web.Router("/api/get-all-objects", &controllers.ApiController{}, "GET:GetAllObjects")
	web.Router("/api/get-all-actions", &controllers.ApiController{}, "GET:GetAllActions")
	web.Router("/api/get-all-roles", &controllers.ApiController{}, "GET:GetAllRoles")

	web.Router("/api/run-casbin-command", &controllers.ApiController{}, "GET:RunCasbinCommand")
	web.Router("/api/refresh-engines", &controllers.ApiController{}, "POST:RefreshEngines")

	web.Router("/api/get-sessions", &controllers.ApiController{}, "GET:GetSessions")
	web.Router("/api/get-session", &controllers.ApiController{}, "GET:GetSingleSession")
	web.Router("/api/update-session", &controllers.ApiController{}, "POST:UpdateSession")
	web.Router("/api/add-session", &controllers.ApiController{}, "POST:AddSession")
	web.Router("/api/delete-session", &controllers.ApiController{}, "POST:DeleteSession")
	web.Router("/api/is-session-duplicated", &controllers.ApiController{}, "GET:IsSessionDuplicated")

	web.Router("/api/get-tokens", &controllers.ApiController{}, "GET:GetTokens")
	web.Router("/api/get-token", &controllers.ApiController{}, "GET:GetToken")
	web.Router("/api/update-token", &controllers.ApiController{}, "POST:UpdateToken")
	web.Router("/api/add-token", &controllers.ApiController{}, "POST:AddToken")
	web.Router("/api/delete-token", &controllers.ApiController{}, "POST:DeleteToken")

	web.Router("/api/get-products", &controllers.ApiController{}, "GET:GetProducts")
	web.Router("/api/get-product", &controllers.ApiController{}, "GET:GetProduct")
	web.Router("/api/update-product", &controllers.ApiController{}, "POST:UpdateProduct")
	web.Router("/api/add-product", &controllers.ApiController{}, "POST:AddProduct")
	web.Router("/api/delete-product", &controllers.ApiController{}, "POST:DeleteProduct")

	web.Router("/api/get-orders", &controllers.ApiController{}, "GET:GetOrders")
	web.Router("/api/get-user-orders", &controllers.ApiController{}, "GET:GetUserOrders")
	web.Router("/api/get-order", &controllers.ApiController{}, "GET:GetOrder")
	web.Router("/api/update-order", &controllers.ApiController{}, "POST:UpdateOrder")
	web.Router("/api/add-order", &controllers.ApiController{}, "POST:AddOrder")
	web.Router("/api/delete-order", &controllers.ApiController{}, "POST:DeleteOrder")
	web.Router("/api/place-order", &controllers.ApiController{}, "POST:PlaceOrder")
	web.Router("/api/cancel-order", &controllers.ApiController{}, "POST:CancelOrder")
	web.Router("/api/pay-order", &controllers.ApiController{}, "POST:PayOrder")

	web.Router("/api/get-payments", &controllers.ApiController{}, "GET:GetPayments")
	web.Router("/api/get-user-payments", &controllers.ApiController{}, "GET:GetUserPayments")
	web.Router("/api/get-payment", &controllers.ApiController{}, "GET:GetPayment")
	web.Router("/api/update-payment", &controllers.ApiController{}, "POST:UpdatePayment")
	web.Router("/api/add-payment", &controllers.ApiController{}, "POST:AddPayment")
	web.Router("/api/delete-payment", &controllers.ApiController{}, "POST:DeletePayment")
	web.Router("/api/notify-payment/?:owner/?:payment", &controllers.ApiController{}, "POST:NotifyPayment")
	web.Router("/api/invoice-payment", &controllers.ApiController{}, "POST:InvoicePayment")

	web.Router("/api/get-plans", &controllers.ApiController{}, "GET:GetPlans")
	web.Router("/api/get-plan", &controllers.ApiController{}, "GET:GetPlan")
	web.Router("/api/update-plan", &controllers.ApiController{}, "POST:UpdatePlan")
	web.Router("/api/add-plan", &controllers.ApiController{}, "POST:AddPlan")
	web.Router("/api/delete-plan", &controllers.ApiController{}, "POST:DeletePlan")

	web.Router("/api/get-pricings", &controllers.ApiController{}, "GET:GetPricings")
	web.Router("/api/get-pricing", &controllers.ApiController{}, "GET:GetPricing")
	web.Router("/api/update-pricing", &controllers.ApiController{}, "POST:UpdatePricing")
	web.Router("/api/add-pricing", &controllers.ApiController{}, "POST:AddPricing")
	web.Router("/api/delete-pricing", &controllers.ApiController{}, "POST:DeletePricing")

	web.Router("/api/get-subscriptions", &controllers.ApiController{}, "GET:GetSubscriptions")
	web.Router("/api/get-subscription", &controllers.ApiController{}, "GET:GetSubscription")
	web.Router("/api/update-subscription", &controllers.ApiController{}, "POST:UpdateSubscription")
	web.Router("/api/add-subscription", &controllers.ApiController{}, "POST:AddSubscription")
	web.Router("/api/delete-subscription", &controllers.ApiController{}, "POST:DeleteSubscription")

	web.Router("/api/get-transactions", &controllers.ApiController{}, "GET:GetTransactions")
	web.Router("/api/get-transaction", &controllers.ApiController{}, "GET:GetTransaction")
	web.Router("/api/update-transaction", &controllers.ApiController{}, "POST:UpdateTransaction")
	web.Router("/api/add-transaction", &controllers.ApiController{}, "POST:AddTransaction")
	web.Router("/api/delete-transaction", &controllers.ApiController{}, "POST:DeleteTransaction")

	web.Router("/api/get-system-info", &controllers.ApiController{}, "GET:GetSystemInfo")
	web.Router("/api/get-version-info", &controllers.ApiController{}, "GET:GetVersionInfo")
	web.Router("/api/health", &controllers.ApiController{}, "GET:Health")
	web.Router("/api/get-prometheus-info", &controllers.ApiController{}, "GET:GetPrometheusInfo")
	web.Router("/api/metrics", &controllers.ApiController{}, "GET:GetMetrics")

	web.Router("/api/get-global-forms", &controllers.ApiController{}, "GET:GetGlobalForms")
	web.Router("/api/get-forms", &controllers.ApiController{}, "GET:GetForms")
	web.Router("/api/get-form", &controllers.ApiController{}, "GET:GetForm")
	web.Router("/api/update-form", &controllers.ApiController{}, "POST:UpdateForm")
	web.Router("/api/add-form", &controllers.ApiController{}, "POST:AddForm")
	web.Router("/api/delete-form", &controllers.ApiController{}, "POST:DeleteForm")

	web.Router("/api/get-syncers", &controllers.ApiController{}, "GET:GetSyncers")
	web.Router("/api/get-syncer", &controllers.ApiController{}, "GET:GetSyncer")
	web.Router("/api/update-syncer", &controllers.ApiController{}, "POST:UpdateSyncer")
	web.Router("/api/add-syncer", &controllers.ApiController{}, "POST:AddSyncer")
	web.Router("/api/delete-syncer", &controllers.ApiController{}, "POST:DeleteSyncer")
	web.Router("/api/run-syncer", &controllers.ApiController{}, "GET:RunSyncer")
	web.Router("/api/test-syncer-db", &controllers.ApiController{}, "POST:TestSyncerDb")

	web.Router("/api/get-webhooks", &controllers.ApiController{}, "GET:GetWebhooks")
	web.Router("/api/get-webhook", &controllers.ApiController{}, "GET:GetWebhook")
	web.Router("/api/update-webhook", &controllers.ApiController{}, "POST:UpdateWebhook")
	web.Router("/api/add-webhook", &controllers.ApiController{}, "POST:AddWebhook")
	web.Router("/api/delete-webhook", &controllers.ApiController{}, "POST:DeleteWebhook")

	web.Router("/api/get-tickets", &controllers.ApiController{}, "GET:GetTickets")
	web.Router("/api/get-ticket", &controllers.ApiController{}, "GET:GetTicket")
	web.Router("/api/update-ticket", &controllers.ApiController{}, "POST:UpdateTicket")
	web.Router("/api/add-ticket", &controllers.ApiController{}, "POST:AddTicket")
	web.Router("/api/delete-ticket", &controllers.ApiController{}, "POST:DeleteTicket")
	web.Router("/api/add-ticket-message", &controllers.ApiController{}, "POST:AddTicketMessage")

	web.Router("/api/set-password", &controllers.ApiController{}, "POST:SetPassword")
	web.Router("/api/check-user-password", &controllers.ApiController{}, "POST:CheckUserPassword")
	web.Router("/api/get-email-and-phone", &controllers.ApiController{}, "GET:GetEmailAndPhone")
	web.Router("/api/send-verification-code", &controllers.ApiController{}, "POST:SendVerificationCode")
	web.Router("/api/verify-code", &controllers.ApiController{}, "POST:VerifyCode")
	web.Router("/api/verify-captcha", &controllers.ApiController{}, "POST:VerifyCaptcha")
	web.Router("/api/reset-email-or-phone", &controllers.ApiController{}, "POST:ResetEmailOrPhone")
	web.Router("/api/get-captcha", &controllers.ApiController{}, "GET:GetCaptcha")
	web.Router("/api/get-verifications", &controllers.ApiController{}, "GET:GetVerifications")

	web.Router("/api/get-ldap-users", &controllers.ApiController{}, "GET:GetLdapUsers")
	web.Router("/api/get-ldaps", &controllers.ApiController{}, "GET:GetLdaps")
	web.Router("/api/get-ldap", &controllers.ApiController{}, "GET:GetLdap")
	web.Router("/api/add-ldap", &controllers.ApiController{}, "POST:AddLdap")
	web.Router("/api/update-ldap", &controllers.ApiController{}, "POST:UpdateLdap")
	web.Router("/api/delete-ldap", &controllers.ApiController{}, "POST:DeleteLdap")
	web.Router("/api/sync-ldap-users", &controllers.ApiController{}, "POST:SyncLdapUsers")

	web.Router("/api/login/oauth/access_token", &controllers.ApiController{}, "POST:GetOAuthToken")
	web.Router("/api/login/oauth/refresh_token", &controllers.ApiController{}, "POST:RefreshToken")
	web.Router("/api/login/oauth/introspect", &controllers.ApiController{}, "POST:IntrospectToken")

	web.Router("/api/get-records", &controllers.ApiController{}, "GET:GetRecords")
	web.Router("/api/get-records-filter", &controllers.ApiController{}, "POST:GetRecordsByFilter")
	web.Router("/api/add-record", &controllers.ApiController{}, "POST:AddRecord")

	web.Router("/api/send-email", &controllers.ApiController{}, "POST:SendEmail")
	web.Router("/api/send-sms", &controllers.ApiController{}, "POST:SendSms")
	web.Router("/api/send-notification", &controllers.ApiController{}, "POST:SendNotification")

	web.Router("/api/webauthn/signup/begin", &controllers.ApiController{}, "GET:WebAuthnSignupBegin")
	web.Router("/api/webauthn/signup/finish", &controllers.ApiController{}, "POST:WebAuthnSignupFinish")
	web.Router("/api/webauthn/signin/begin", &controllers.ApiController{}, "GET:WebAuthnSigninBegin")
	web.Router("/api/webauthn/signin/finish", &controllers.ApiController{}, "POST:WebAuthnSigninFinish")

	web.Router("/api/mfa/setup/initiate", &controllers.ApiController{}, "POST:MfaSetupInitiate")
	web.Router("/api/mfa/setup/verify", &controllers.ApiController{}, "POST:MfaSetupVerify")
	web.Router("/api/mfa/setup/enable", &controllers.ApiController{}, "POST:MfaSetupEnable")
	web.Router("/api/delete-mfa", &controllers.ApiController{}, "POST:DeleteMfa")
	web.Router("/api/set-preferred-mfa", &controllers.ApiController{}, "POST:SetPreferredMfa")

	web.Router("/.well-known/openid-configuration", &controllers.RootController{}, "GET:GetOidcDiscovery")
	web.Router("/.well-known/:application/openid-configuration", &controllers.RootController{}, "GET:GetOidcDiscoveryByApplication")
	web.Router("/.well-known/jwks", &controllers.RootController{}, "*:GetJwks")
	web.Router("/.well-known/:application/jwks", &controllers.RootController{}, "*:GetJwksByApplication")
	web.Router("/.well-known/webfinger", &controllers.RootController{}, "GET:GetWebFinger")
	web.Router("/.well-known/:application/webfinger", &controllers.RootController{}, "GET:GetWebFingerByApplication")

	web.Router("/cas/:organization/:application/serviceValidate", &controllers.RootController{}, "GET:CasServiceValidate")
	web.Router("/cas/:organization/:application/proxyValidate", &controllers.RootController{}, "GET:CasProxyValidate")
	web.Router("/cas/:organization/:application/proxy", &controllers.RootController{}, "GET:CasProxy")
	web.Router("/cas/:organization/:application/validate", &controllers.RootController{}, "GET:CasValidate")

	web.Router("/cas/:organization/:application/p3/serviceValidate", &controllers.RootController{}, "GET:CasP3ServiceValidate")
	web.Router("/cas/:organization/:application/p3/proxyValidate", &controllers.RootController{}, "GET:CasP3ProxyValidate")
	web.Router("/cas/:organization/:application/samlValidate", &controllers.RootController{}, "POST:SamlValidate")

	web.Router("/scim/*", &controllers.RootController{}, "*:HandleScim")

	web.Router("/api/faceid-signin-begin", &controllers.ApiController{}, "GET:FaceIDSigninBegin")
}
