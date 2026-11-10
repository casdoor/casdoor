// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package controllers

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/tlsserver"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) getCertLoginApplication(name string) (*object.Application, error) {
	application, err := object.GetApplication(util.GetId("admin", name))
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, fmt.Errorf(c.T("auth:The application: %s does not exist"), name)
	}
	if !application.IsCertLoginEnabled() {
		return nil, errors.New(c.T("enterprise:The login method: login with certificate is not enabled for the application"))
	}
	if err = edition.Require(edition.Enterprise, edition.FeatureCertLogin); err != nil {
		return nil, err
	}
	return application, nil
}

// getCertLoginOrganization is the organization whose trusted CAs check the certificate: the application's own, or
// the one the user picked for a shared application
func getCertLoginOrganization(application *object.Application, organization string) string {
	if application.IsShared && organization != "" {
		return organization
	}
	return application.Organization
}

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

func (c *ApiController) getClientCertificateChain() ([]*x509.Certificate, error) {
	ders, isTlsPort := tlsserver.GetPeerCertificates(c.Ctx.Request)
	if isTlsPort {
		if len(ders) == 0 {
			return nil, errors.New(c.T("enterprise:No client certificate was presented"))
		}
		return object.ParseCertificateDers(ders)
	}

	certificate := tlsserver.GetProxiedCertificate(c.Ctx.Request)
	if certificate == "" {
		return nil, errors.New(c.T("enterprise:No client certificate was presented"))
	}
	return object.ParseCertificateChain(certificate)
}

func (c *ApiController) addCertLoginTicket() (string, string, error) {
	application, err := c.getCertLoginApplication(c.Ctx.Input.Query("application"))
	if err != nil {
		return "", "", err
	}
	chain, err := c.getClientCertificateChain()
	if err != nil {
		return "", "", err
	}
	user, err := object.VerifyCertLogin(getCertLoginOrganization(application, c.Ctx.Input.Query("organization")), chain, c.GetAcceptLanguage())
	if err != nil {
		return "", "", err
	}
	ticket, err := object.AddCertTicket(application, user, object.GetCertFingerprint(chain[0]))
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
		resp := &Response{Status: "error", Msg: err.Error()}
		c.localizeEditionError(resp)
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
	application, err := c.getCertLoginApplication(c.Ctx.Input.Query("application"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	nonce, err := object.AddCertChallenge(application)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(nonce, int(object.CertChallengeTimeout.Seconds()))
}

// checkCertSignin signs in with the ticket of /api/login/cert, or a certificate and its signature of a challenge
func (c *ApiController) checkCertSignin(authForm *form.AuthForm) (*object.User, error) {
	application, err := c.getCertLoginApplication(authForm.Application)
	if err != nil {
		return nil, err
	}
	lang := c.GetAcceptLanguage()

	var user *object.User
	if authForm.Code != "" {
		ticket, err := object.ConsumeCertLoginTicket(object.CertLoginKindTicket, authForm.Code, application, lang)
		if err != nil {
			return nil, err
		}
		user, err = object.GetUser(ticket.User)
		if err != nil {
			return nil, err
		}
		if user == nil || user.IsDeleted || user.IsForbidden {
			return nil, errors.New(c.T("check:The user is forbidden to sign in, please contact the administrator"))
		}
	} else {
		chain, err := object.ParseCertificateChain(authForm.Certificate)
		if err != nil {
			return nil, fmt.Errorf(c.T("enterprise:The certificate is invalid: %s"), err.Error())
		}
		signature, err := object.DecodeSignature(authForm.Signature)
		if err != nil {
			return nil, errors.New(c.T("enterprise:The signature of the challenge is invalid"))
		}
		// the challenge is used up by any attempt, a wrong signature included
		_, err = object.ConsumeCertLoginTicket(object.CertLoginKindChallenge, authForm.Nonce, application, lang)
		if err != nil {
			return nil, err
		}
		if !object.VerifyCertSignature(chain[0], []byte(authForm.Nonce), signature) {
			return nil, errors.New(c.T("enterprise:The signature of the challenge is invalid"))
		}
		user, err = object.VerifyCertLogin(getCertLoginOrganization(application, authForm.Organization), chain, lang)
		if err != nil {
			return nil, err
		}
	}

	authForm.Organization = user.Owner
	authForm.Username = user.Name
	return user, nil
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
		resp := &Response{Status: "error", Msg: status.Error}
		c.localizeEditionError(resp)
		status.Error = resp.Msg
		status.Code = resp.Code
	}
	c.ResponseOk(statuses)
}
