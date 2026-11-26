// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/casdoor/casdoor/ee"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// getPageOrigin is the origin of the sign-in page that called the API, which a wallet, for one, compares with the
// domain of what it is asked to sign
func (c *ApiController) getPageOrigin() string {
	if origin := c.Ctx.Request.Header.Get("Origin"); origin != "" && origin != "null" {
		return origin
	}
	if referer, err := url.Parse(c.Ctx.Request.Referer()); err == nil && referer.Host != "" {
		return fmt.Sprintf("%s://%s", referer.Scheme, referer.Host)
	}
	scheme := "http"
	if c.Ctx.Request.TLS != nil || c.Ctx.Request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Ctx.Request.Host)
}

// getSigninPluginProvider returns the identity provider plugin of the application named by the query
func (c *ApiController) getSigninPluginProvider() (*object.Application, *object.Provider, bool) {
	applicationId := c.Ctx.Input.Query("application")
	application, err := object.GetApplication(applicationId)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, nil, false
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), applicationId))
		return nil, nil, false
	}

	providerName := c.Ctx.Input.Query("provider")
	provider, err := object.GetProvider(util.GetId("admin", providerName))
	if err != nil {
		c.ResponseError(err.Error())
		return nil, nil, false
	}
	if provider == nil || provider.Type != object.ProviderTypePlugin || provider.Category != "OAuth" || !application.GetProviderItem(provider.Name).IsProviderVisible() {
		c.ResponseError(fmt.Sprintf(c.T("auth:The provider: %s is not enabled for the application"), providerName))
		return nil, nil, false
	}
	return application, provider, true
}

// GetPluginChallenge
// @Title GetPluginChallenge
// @Tag Login API
// @Description get the challenge an identity provider plugin with a sign-in script has the user answer, such as the message for a wallet to sign, valid for 5 minutes in this session
// @Param   application     query    string  true        "The id of the application, like admin/app-built-in"
// @Param   provider     query    string  true        "The name of the provider"
// @Param   body    body   map[string]string  true        "What the sign-in script sends, such as the wallet address"
// @Success 200 {object} controllers.Response The Response object
// @router /get-plugin-challenge [post]
func (c *ApiController) GetPluginChallenge() {
	application, provider, ok := c.getSigninPluginProvider()
	if !ok {
		return
	}

	input := map[string]string{}
	if len(c.Ctx.Input.RequestBody) > 0 {
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	challenge, session, err := ee.GetPluginChallenge(provider, application, c.getPageOrigin(), input)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if err = c.SetSession(ee.PluginChallengeSessionKey, session); err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(challenge)
}

// GetPluginScript
// @Title GetPluginScript
// @Tag Login API
// @Description get the JavaScript module that an identity provider plugin runs on the sign-in page
// @Param   application     query    string  true        "The id of the application, like admin/app-built-in"
// @Param   provider     query    string  true        "The name of the provider"
// @Success 200 {string} string The JavaScript module
// @router /get-plugin-script [get]
func (c *ApiController) GetPluginScript() {
	_, provider, ok := c.getSigninPluginProvider()
	if !ok {
		return
	}

	script, err := ee.GetPluginScript(provider)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if script == "" {
		c.ResponseError(fmt.Sprintf("the plugin of the provider: %s has no sign-in script", provider.Name))
		return
	}

	header := c.Ctx.ResponseWriter.Header()
	header.Set("Content-Type", "text/javascript; charset=utf-8")
	header.Set("Cache-Control", "no-cache")
	header.Set("X-Content-Type-Options", "nosniff")
	c.Ctx.ResponseWriter.WriteHeader(http.StatusOK)
	_, _ = c.Ctx.ResponseWriter.Write([]byte(script))
}
