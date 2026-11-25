// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package controllers

import (
	"fmt"
	"html"
	"net/http"

	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/ee"
	"github.com/casdoor/casdoor/object"
)

// CloudSignin
// @Title CloudSignin
// @Tag Login API
// @Description signs the owner of a Casdoor Cloud hosted instance in from the Casdoor Cloud console, with a one-time ticket the console signed, posted as the form field "ticket"
// @Param   ticket     formData    string  true        "The sign-in ticket"
// @Success 200 goes on to the console
// @router /cloud-signin [post]
func (c *ApiController) CloudSignin() {
	c.Ctx.Output.Header("Cache-Control", "no-store")
	if err := edition.Require(edition.Commercial, edition.FeatureCloudSignin); err != nil {
		c.writeCloudSigninPage(http.StatusForbidden, err.Error())
		return
	}

	ticket, err := ee.ParseCloudSigninTicket(c.Ctx.Request.FormValue("ticket"))
	if err != nil {
		c.writeCloudSigninPage(http.StatusBadRequest, err.Error())
		return
	}

	user, err := object.GetUser(ticket.User)
	if err != nil {
		c.writeCloudSigninPage(http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil || user.IsDeleted || user.IsForbidden || !user.IsGlobalAdmin() {
		c.writeCloudSigninPage(http.StatusBadRequest, fmt.Sprintf("the user %s can't sign in, use \"Reset password and unlock\" on the instance page of the console", ticket.User))
		return
	}

	err = c.StartBrowserSession(user, "app-built-in", 24)
	if err != nil {
		c.writeCloudSigninPage(http.StatusInternalServerError, err.Error())
		return
	}

	detail := "Casdoor Cloud console: " + ticket.Operator
	if len(detail) > 100 {
		detail = detail[:100]
	}
	c.Ctx.Input.SetParam("recordUserId", user.GetId())
	c.Ctx.Input.SetParam("recordFailureReason", detail)
	c.writeCloudSigninPage(http.StatusOK, "")
}

// writeCloudSigninPage answers the form post with a page: on success it goes on to the console from this site, so
// the new session cookie is sent with it
func (c *ApiController) writeCloudSigninPage(status int, msg string) {
	body := `<!doctype html><html><head><meta charset="utf-8"><meta http-equiv="refresh" content="0;url=/"><title>Signing in</title></head><body><a href="/">Continue</a></body></html>`
	if msg == "" {
		c.Data["json"] = &controllers.Response{Status: "ok"}
	} else {
		c.Data["json"] = &controllers.Response{Status: "error", Msg: msg}
		body = fmt.Sprintf(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Sign-in failed</title></head>`+
			`<body style="font-family:system-ui,sans-serif;max-width:560px;margin:15vh auto;padding:0 16px;color:#333">`+
			`<h2>Sign-in failed</h2><p>%s</p><p><a href="/login">Sign in with a password</a></p></body></html>`, html.EscapeString(msg))
	}
	c.Ctx.Output.SetStatus(status)
	c.Ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	_ = c.Ctx.Output.Body([]byte(body))
}
