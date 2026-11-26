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

	"github.com/casdoor/casdoor/conf"
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
	logo := fmt.Sprintf("%s/img/casdoor-logo_1185x256.png", conf.GetConfigString("staticBaseUrl"))
	logoDark := fmt.Sprintf("%s/img/casdoor-logo_1185x256_dark.png", conf.GetConfigString("staticBaseUrl"))
	if organization, err := object.GetOrganization("admin/built-in"); err == nil && organization != nil && organization.Logo != "" {
		logo = organization.Logo
		logoDark = organization.LogoDark
	}
	logoHtml := fmt.Sprintf(`<img class="logo" src="%s" alt="Casdoor">`, html.EscapeString(logo))
	if logoDark != "" {
		logoHtml = fmt.Sprintf(`<picture><source srcset="%s" media="(prefers-color-scheme: dark)">%s</picture>`, html.EscapeString(logoDark), logoHtml)
	}

	var title, content string
	if msg == "" {
		c.Data["json"] = &controllers.Response{Status: "ok"}
		title = "Signing in"
		content = `<div class="spinner"></div><h1>Signing you in</h1><p>Taking you to the Casdoor console&hellip;</p>` +
			`<p class="later">Not redirected? <a href="/">Open the console</a></p>` +
			`<script>location.replace("/")</script>`
	} else {
		c.Data["json"] = &controllers.Response{Status: "error", Msg: msg}
		title = "Sign-in failed"
		content = fmt.Sprintf(`<div class="icon">!</div><h1>Sign-in failed</h1><p class="msg">%s</p>`+
			`<a class="button" href="/login">Sign in with a password</a>`, html.EscapeString(msg))
	}

	refresh := ""
	if msg == "" {
		refresh = `<noscript><meta http-equiv="refresh" content="0;url=/"></noscript>`
	}
	body := fmt.Sprintf(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">%s<title>%s</title><style>%s</style></head>`+
		`<body><main>%s%s</main></body></html>`, refresh, title, cloudSigninPageStyle, logoHtml, content)

	c.Ctx.Output.SetStatus(status)
	c.Ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	_ = c.Ctx.Output.Body([]byte(body))
}

const cloudSigninPageStyle = `:root{--bg:#f5f7fa;--card:#fff;--text:#1f2329;--muted:#646a73;--border:#e5e6eb;--primary:#5734d3;--danger:#d93026}` +
	`@media (prefers-color-scheme:dark){:root{--bg:#141414;--card:#1f1f1f;--text:#e8e8e8;--muted:#a0a0a0;--border:#303030;--primary:#8b6cf0;--danger:#ff6b5e}}` +
	`*{box-sizing:border-box}` +
	`body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;padding:16px;background:var(--bg);color:var(--text);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif}` +
	`main{width:100%;max-width:400px;padding:40px 32px;background:var(--card);border:1px solid var(--border);border-radius:12px;box-shadow:0 4px 24px rgba(0,0,0,.06);text-align:center}` +
	`.logo{display:block;max-width:180px;max-height:48px;margin:0 auto 32px}` +
	`h1{margin:0 0 8px;font-size:20px;font-weight:600}` +
	`p{margin:0;font-size:14px;line-height:1.6;color:var(--muted)}` +
	`a{color:var(--primary);text-decoration:none}a:hover{text-decoration:underline}` +
	`.spinner{width:36px;height:36px;margin:0 auto 20px;border:3px solid var(--border);border-top-color:var(--primary);border-radius:50%;animation:spin .8s linear infinite}` +
	`@keyframes spin{to{transform:rotate(360deg)}}` +
	`.later{margin-top:24px;font-size:13px;opacity:0;animation:show 0s linear 3s forwards}` +
	`@keyframes show{to{opacity:1}}` +
	`.icon{width:40px;height:40px;margin:0 auto 20px;border-radius:50%;background:var(--danger);color:#fff;font-size:22px;font-weight:700;line-height:40px}` +
	`.msg{word-break:break-word}` +
	`.button{display:inline-block;margin-top:24px;padding:8px 20px;border-radius:6px;background:var(--primary);color:#fff;font-size:14px}.button:hover{text-decoration:none;opacity:.9}`
