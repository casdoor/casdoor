package routers

import (
	"fmt"
	"strings"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/object"
)

var mfaExcludedPrefixes = []string{
	"/api/get-",
	"/api/login",
	"/api/signup",
	"/api/send-verification-code",
	"/api/verify-code",
	"/api/set-password",
	"/api/get-captcha",
	"/api/verify-captcha",
	"/api/reset-email-or-phone",
	"/api/login/oauth",
	"/api/get-app-login",
	"/api/get-saml-login",
	"/api/saml/metadata",
	"/api/saml/redirect",
	"/api/acs",
	"/cas",
	"/scim",
	"/.well-known",
	"/api/health",
	"/api/get-system-info",
	"/api/get-version-info",
	"/api/get-prometheus-info",
	"/api/metrics",
}

func shouldSkipMfa(urlPath string) bool {
	for _, prefix := range mfaExcludedPrefixes {
		if strings.HasPrefix(urlPath, prefix) {
			return true
		}
	}
	return false
}

func MfaFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path
	if shouldSkipMfa(urlPath) {
		return
	}

	userId := getSessionUser(ctx)

	if userId == "" {
		responseError(ctx, T(ctx, fmt.Sprintf("general:The user: %s doesn't exist", userId)))
		return
	}

	user, err := object.GetUser(userId)
	if err != nil {
		responseError(ctx, err.Error())
		return
	}

	if user == nil {
		responseError(ctx, T(ctx, fmt.Sprintf("general:The user: %s doesn't exist", userId)))
		return
	}

	if user.IsMfaEnabled() {
		mfaVerified := ctx.Input.CruSession.Get(object.MfaCompleted)
		if mfaVerified == nil || mfaVerified.(bool) == false {
			denyRequest(ctx)
		}
	}
}
