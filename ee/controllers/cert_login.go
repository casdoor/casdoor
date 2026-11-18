// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package controllers

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/ee"
	"github.com/casdoor/casdoor/ee/tlsserver"
	"github.com/casdoor/casdoor/util"
)

func getHostname(host string) string {
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		return strings.ToLower(strings.Trim(hostname, "[]"))
	}
	return strings.ToLower(strings.Trim(host, "[]"))
}

// getCertLoginReturnUrl checks the page that the certificate sign-in goes back to: a sign-in page of this Casdoor,
// on the host of the request or of the configured origin, as the URL gets the one-time ticket
func (c *ApiController) getCertLoginReturnUrl(value string) (*url.URL, bool) {
	if value == "" {
		return nil, true
	}
	returnUrl, err := url.Parse(value)
	if err != nil || (returnUrl.Scheme != "https" && returnUrl.Scheme != "http") || returnUrl.Host == "" || returnUrl.User != nil {
		return nil, false
	}

	hostnames := []string{getHostname(c.Ctx.Request.Host)}
	for _, key := range []string{"origin", "originFrontend"} {
		if origin, err := url.Parse(conf.GetConfigString(key)); err == nil && origin.Host != "" {
			hostnames = append(hostnames, getHostname(origin.Host))
		}
	}
	if !util.InSlice(hostnames, getHostname(returnUrl.Host)) {
		return nil, false
	}

	path := returnUrl.Path
	if basePath := conf.GetBasePath(); basePath != "" {
		path = strings.TrimPrefix(path, basePath)
	}
	if path != "/login" && !strings.HasPrefix(path, "/login/") && !strings.HasPrefix(path, "/cas/") {
		return nil, false
	}
	return returnUrl, true
}

func (c *ApiController) getClientCertificateDers() ([][]byte, error) {
	ders, isTlsPort := tlsserver.GetPeerCertificates(c.Ctx.Request)
	if isTlsPort {
		if len(ders) == 0 {
			return nil, errors.New(c.T("enterprise:No client certificate was presented"))
		}
		return ders, nil
	}

	certificate := tlsserver.GetProxiedCertificate(c.Ctx.Request)
	if certificate == "" {
		return nil, errors.New(c.T("enterprise:No client certificate was presented"))
	}
	return ee.ReadCertificateDers(certificate)
}

func (c *ApiController) addCertLoginTicket() (string, string, error) {
	application, err := c.GetCertLoginApplication(c.Ctx.Input.Query("application"))
	if err != nil {
		return "", "", err
	}
	ders, err := c.getClientCertificateDers()
	if err != nil {
		return "", "", err
	}
	user, fingerprint, err := ee.VerifyCertLogin(ee.GetCertLoginOrganization(application, c.Ctx.Input.Query("organization")), &ee.CertLoginInput{Ders: ders}, c.GetAcceptLanguage())
	if err != nil {
		return "", "", err
	}
	ticket, err := ee.AddCertTicket(application, user, fingerprint)
	if err != nil {
		return "", "", err
	}
	return ticket, user.GetId(), nil
}

// LoginByCert
// @Title LoginByCert
// @Tag Login API
// @Description sign in with the client certificate of the TLS connection (the port certLoginPort, or a reverse proxy in certLoginTrustedProxies passing it in the header certLoginHeader). With "return", the browser goes back to that sign-in page with "certLoginToken", which the login page posts to /api/login with signinMethod "Certificate" as "code"; without it the ticket is returned
// @Param   application    query    string  true     "The application"
// @Param   organization   query    string  false    "The organization, for a shared application"
// @Param   return         query    string  false    "The sign-in page to go back to"
// @Success 200 {object} controllers.Response The Response object
// @router /login/cert [get]
func (c *ApiController) LoginByCert() {
	returnUrl, ok := c.getCertLoginReturnUrl(c.Ctx.Input.Query("return"))
	if !ok {
		c.ResponseError(c.T("enterprise:The return URL is not a sign-in page of this site"))
		return
	}

	ticket, userId, err := c.addCertLoginTicket()
	if returnUrl == nil {
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(ticket, userId)
		return
	}

	query := returnUrl.Query()
	query.Del("certLoginToken")
	query.Del("certLoginError")
	if err != nil {
		resp := &controllers.Response{Status: "error", Msg: err.Error()}
		c.LocalizeEditionError(resp)
		query.Set("certLoginError", resp.Msg)
	} else {
		query.Set("certLoginToken", ticket)
	}
	returnUrl.RawQuery = query.Encode()
	c.Ctx.Output.Header("Cache-Control", "no-store")
	c.Redirect(returnUrl.String(), http.StatusFound)
}

// GetCertChallenge
// @Title GetCertChallenge
// @Tag Login API
// @Description get a one-time challenge for a UKey to sign, valid for 5 minutes. Post {application, organization, signinMethod: "Certificate", certificate (PEM or base64 DER), nonce, signature (base64)} to /api/login, the signature being of the UTF-8 bytes of the nonce: RSA PKCS #1 v1.5 with SHA-256 or ECDSA with SHA-256
// @Param   application    query    string  true     "The application"
// @Success 200 {object} controllers.Response The Response object
// @router /get-cert-challenge [get]
func (c *ApiController) GetCertChallenge() {
	application, err := c.GetCertLoginApplication(c.Ctx.Input.Query("application"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	nonce, err := ee.AddCertChallenge(application)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(nonce, int(ee.CertChallengeTimeout.Seconds()))
}

// GetTlsListeners
// @Title GetTlsListeners
// @Tag System API
// @Description get the state of the certificate sign-in port configured in app.conf
// @Success 200 {array} tlsserver.ListenerStatus The Response object
// @router /get-tls-listeners [get]
func (c *ApiController) GetTlsListeners() {
	if !c.RequireGlobalAdmin() {
		return
	}

	statuses := tlsserver.GetStatuses()
	for _, status := range statuses {
		if status.Error == "" {
			continue
		}
		resp := &controllers.Response{Status: "error", Msg: status.Error}
		c.LocalizeEditionError(resp)
		status.Error = resp.Msg
		status.Code = resp.Code
	}
	c.ResponseOk(statuses)
}
