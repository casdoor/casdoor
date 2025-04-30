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
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	stringadapter "github.com/qiangmzsx/string-adapter/v2"
)

var Enforcer *casbin.Enforcer

func InitApi() {
	e, err := object.GetInitializedEnforcer(util.GetId("built-in", "api-enforcer-built-in"))
	if err != nil {
		panic(err)
	}

	Enforcer = e.Enforcer
	Enforcer.ClearPolicy()

	// if len(Enforcer.GetPolicy()) == 0 {
	if true {
		ruleText := `
p, built-in, *, *, *, *, *
p, app, *, *, *, *, *
p, *, *, POST, /api/signup, *, *
p, *, *, GET, /api/get-email-and-phone, *, *
p, *, *, POST, /api/login, *, *
p, *, *, GET, /api/get-app-login, *, *
p, *, *, POST, /api/logout, *, *
p, *, *, GET, /api/logout, *, *
p, *, *, POST, /api/callback, *, *
p, *, *, POST, /api/device-auth, *, *
p, *, *, GET, /api/get-account, *, *
p, *, *, GET, /api/userinfo, *, *
p, *, *, GET, /api/user, *, *
p, *, *, GET, /api/health, *, *
p, *, *, *, /api/webhook, *, *
p, *, *, GET, /api/get-qrcode, *, *
p, *, *, GET, /api/get-webhook-event, *, *
p, *, *, GET, /api/get-captcha-status, *, *
p, *, *, *, /api/login/oauth, *, *
p, *, *, GET, /api/get-application, *, *
p, *, *, GET, /api/get-organization-applications, *, *
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
p, *, *, POST, /api/verify-code, *, *
p, *, *, POST, /api/reset-email-or-phone, *, *
p, *, *, POST, /api/upload-resource, *, *
p, *, *, GET, /.well-known/openid-configuration, *, *
p, *, *, GET, /.well-known/webfinger, *, *
p, *, *, *, /.well-known/jwks, *, *
p, *, *, GET, /api/get-saml-login, *, *
p, *, *, POST, /api/acs, *, *
p, *, *, GET, /api/saml/metadata, *, *
p, *, *, *, /api/saml/redirect, *, *
p, *, *, *, /cas, *, *
p, *, *, *, /scim, *, *
p, *, *, *, /api/webauthn, *, *
p, *, *, GET, /api/get-release, *, *
p, *, *, GET, /api/get-default-application, *, *
p, *, *, GET, /api/get-prometheus-info, *, *
p, *, *, *, /api/metrics, *, *
p, *, *, GET, /api/get-pricing, *, *
p, *, *, GET, /api/get-plan, *, *
p, *, *, GET, /api/get-subscription, *, *
p, *, *, GET, /api/get-provider, *, *
p, *, *, GET, /api/get-organization-names, *, *
p, *, *, GET, /api/get-all-objects, *, *
p, *, *, GET, /api/get-all-actions, *, *
p, *, *, GET, /api/get-all-roles, *, *
p, *, *, GET, /api/run-casbin-command, *, *
p, *, *, POST, /api/refresh-engines, *, *
p, *, *, GET, /api/get-invitation-info, *, *
p, *, *, GET, /api/faceid-signin-begin, *, *
`

		sa := stringadapter.NewAdapter(ruleText)
		// load all rules from string adapter to enforcer's memory
		err = sa.LoadPolicy(Enforcer.GetModel())
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

	user, err := object.GetUser(util.GetId(subOwner, subName))
	if err != nil {
		panic(err)
	}

	if subOwner == "app" {
		return true
	}

	if user != nil {
		if user.IsDeleted {
			return false
		}

		if user.IsAdmin && (subOwner == objOwner || (objOwner == "admin")) {
			return true
		}
	}

	res, err := Enforcer.Enforce(subOwner, subName, method, urlPath, objOwner, objName)
	if err != nil {
		panic(err)
	}

	return res
}

func isAllowedInDemoMode(subOwner string, subName string, method string, urlPath string, objOwner string, objName string) bool {
	if method == "POST" {
		if strings.HasPrefix(urlPath, "/api/login") || urlPath == "/api/logout" || urlPath == "/api/signup" || urlPath == "/api/callback" || urlPath == "/api/send-verification-code" || urlPath == "/api/send-email" || urlPath == "/api/verify-captcha" || urlPath == "/api/verify-code" || urlPath == "/api/check-user-password" || strings.HasPrefix(urlPath, "/api/mfa/") || urlPath == "/api/webhook" || urlPath == "/api/get-qrcode" || urlPath == "/api/refresh-engines" {
			return true
		} else if urlPath == "/api/update-user" {
			// Allow ordinary users to update their own information
			if (subOwner == objOwner && subName == objName || subOwner == "app") && !(subOwner == "built-in" && subName == "admin") {
				return true
			}
			return false
		} else if urlPath == "/api/upload-resource" {
			if subOwner == "app" && subName == "app-casibase" {
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
