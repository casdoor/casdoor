// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/object"
)

const (
	InvalidRequest           string = "INVALID_REQUEST"
	InvalidTicketSpec        string = "INVALID_TICKET_SPEC"
	UnauthorizedServiceProxy string = "UNAUTHORIZED_SERVICE_PROXY"
	InvalidProxyCallback     string = "INVALID_PROXY_CALLBACK"
	InvalidTicket            string = "INVALID_TICKET"
	InvalidService           string = "INVALID_SERVICE"
	InternalError            string = "INTERNAL_ERROR"
	UnauthorizedService      string = "UNAUTHORIZED_SERVICE"
)

func (c *RootController) CasValidate() {
	ticket := c.Input().Get("ticket")
	service := c.Input().Get("service")
	c.Ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	if service == "" || ticket == "" {
		c.Ctx.Output.Body([]byte("no\n"))
		return
	}
	if ok, response, issuedService, _ := object.GetCasTokenByTicket(ticket); ok {
		// check whether service is the one for which we previously issued token
		if issuedService == service {
			c.Ctx.Output.Body([]byte(fmt.Sprintf("yes\n%s\n", response.User)))
			return
		}
	}
	// token not found
	c.Ctx.Output.Body([]byte("no\n"))
}

func (c *RootController) CasServiceValidate() {
	ticket := c.Input().Get("ticket")
	format := c.Input().Get("format")
	if !strings.HasPrefix(ticket, "ST") {
		c.sendCasAuthenticationResponseErr(InvalidTicket, fmt.Sprintf("Ticket %s not recognized", ticket), format)
	}
	c.CasP3ServiceAndProxyValidate()
}

func (c *RootController) CasProxyValidate() {
	ticket := c.Input().Get("ticket")
	format := c.Input().Get("format")
	if !strings.HasPrefix(ticket, "PT") {
		c.sendCasAuthenticationResponseErr(InvalidTicket, fmt.Sprintf("Ticket %s not recognized", ticket), format)
	}
	c.CasP3ServiceAndProxyValidate()
}

func (c *RootController) CasP3ServiceAndProxyValidate() {
	ticket := c.Input().Get("ticket")
	format := c.Input().Get("format")
	service := c.Input().Get("service")
	pgtUrl := c.Input().Get("pgtUrl")

	serviceResponse := object.CasServiceResponse{
		Xmlns: "http://www.yale.edu/tp/cas",
	}

	// check whether all required parameters are met
	if service == "" || ticket == "" {
		c.sendCasAuthenticationResponseErr(InvalidRequest, "service and ticket must exist", format)
		return
	}
	ok, response, issuedService, userId := object.GetCasTokenByTicket(ticket)
	// find the token
	if ok {
		// check whether service is the one for which we previously issued token
		if strings.HasPrefix(service, issuedService) {
			serviceResponse.Success = response
		} else {
			// service not match
			c.sendCasAuthenticationResponseErr(InvalidService, fmt.Sprintf("service %s and %s does not match", service, issuedService), format)
			return
		}
	} else {
		// token not found
		c.sendCasAuthenticationResponseErr(InvalidTicket, fmt.Sprintf("Ticket %s not recognized", ticket), format)
		return
	}

	if pgtUrl != "" && serviceResponse.Failure == nil {
		// that means we are in proxy web flow
		pgt := object.StoreCasTokenForPgt(serviceResponse.Success, service, userId)
		pgtiou := serviceResponse.Success.ProxyGrantingTicket
		// todo: check whether it is https
		pgtUrlObj, err := url.Parse(pgtUrl)
		if pgtUrlObj.Scheme != "https" {
			c.sendCasAuthenticationResponseErr(InvalidProxyCallback, "callback is not https", format)
			return
		}
		// make a request to pgturl passing pgt and pgtiou
		if err != nil {
			c.sendCasAuthenticationResponseErr(InternalError, err.Error(), format)
			return
		}
		param := pgtUrlObj.Query()
		param.Add("pgtId", pgt)
		param.Add("pgtIou", pgtiou)
		pgtUrlObj.RawQuery = param.Encode()

		request, err := http.NewRequest("GET", pgtUrlObj.String(), nil)
		if err != nil {
			c.sendCasAuthenticationResponseErr(InternalError, err.Error(), format)
			return
		}

		resp, err := http.DefaultClient.Do(request)
		if err != nil || !(resp.StatusCode >= 200 && resp.StatusCode < 400) {
			// failed to send request
			c.sendCasAuthenticationResponseErr(InvalidProxyCallback, err.Error(), format)
			return
		}
	}
	// everything is ok, send the response
	if format == "json" {
		c.ResponseOk(serviceResponse)
	} else {
		c.Data["xml"] = serviceResponse
		c.ServeXML()
	}
}

func (c *RootController) CasProxy() {
	pgt := c.Input().Get("pgt")
	targetService := c.Input().Get("targetService")
	format := c.Input().Get("format")
	if pgt == "" || targetService == "" {
		c.sendCasProxyResponseErr(InvalidRequest, "pgt and targetService must exist", format)
		return
	}

	ok, authenticationSuccess, issuedService, userId := object.GetCasTokenByPgt(pgt)
	if !ok {
		c.sendCasProxyResponseErr(UnauthorizedService, "service not authorized", format)
		return
	}

	newAuthenticationSuccess := authenticationSuccess.DeepCopy()
	if newAuthenticationSuccess.Proxies == nil {
		newAuthenticationSuccess.Proxies = &object.CasProxies{}
	}
	newAuthenticationSuccess.Proxies.Proxies = append(newAuthenticationSuccess.Proxies.Proxies, issuedService)
	proxyTicket := object.StoreCasTokenForProxyTicket(&newAuthenticationSuccess, targetService, userId)

	serviceResponse := object.CasServiceResponse{
		Xmlns: "http://www.yale.edu/tp/cas",
		ProxySuccess: &object.CasProxySuccess{
			ProxyTicket: proxyTicket,
		},
	}

	if format == "json" {
		c.ResponseOk(serviceResponse)
	} else {
		c.Data["xml"] = serviceResponse
		c.ServeXML()
	}
}

func (c *RootController) SamlValidate() {
	c.Ctx.Output.Header("Content-Type", "text/xml; charset=utf-8")
	target := c.Input().Get("TARGET")
	body := c.Ctx.Input.RequestBody
	envelopRequest := struct {
		XMLName xml.Name `xml:"Envelope"`
		Body    struct {
			XMLName xml.Name `xml:"Body"`
			Content string   `xml:",innerxml"`
		}
	}{}

	err := xml.Unmarshal(body, &envelopRequest)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	response, service, err := object.GetValidationBySaml(envelopRequest.Body.Content, c.Ctx.Request.Host)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !strings.HasPrefix(target, service) {
		c.ResponseError(fmt.Sprintf(c.T("cas:Service %s and %s do not match"), target, service))
		return
	}

	envelopResponse := struct {
		XMLName xml.Name `xml:"SOAP-ENV:Envelope"`
		Xmlns   string   `xml:"xmlns:SOAP-ENV"`
		Body    struct {
			XMLName xml.Name `xml:"SOAP-ENV:Body"`
			Content string   `xml:",innerxml"`
		}
	}{}
	envelopResponse.Xmlns = "http://schemas.xmlsoap.org/soap/envelope/"
	envelopResponse.Body.Content = response

	data, err := xml.Marshal(envelopResponse)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Ctx.Output.Body(data)
}

func (c *RootController) sendCasProxyResponseErr(code, msg, format string) {
	serviceResponse := object.CasServiceResponse{
		Xmlns: "http://www.yale.edu/tp/cas",
		ProxyFailure: &object.CasProxyFailure{
			Code:    code,
			Message: msg,
		},
	}
	if format == "json" {
		c.ResponseOk(serviceResponse)
	} else {
		c.Data["xml"] = serviceResponse
		c.ServeXML()
	}
}

func (c *RootController) sendCasAuthenticationResponseErr(code, msg, format string) {
	serviceResponse := object.CasServiceResponse{
		Xmlns: "http://www.yale.edu/tp/cas",
		Failure: &object.CasAuthenticationFailure{
			Code:    code,
			Message: msg,
		},
	}

	if format == "json" {
		c.ResponseOk(serviceResponse)
	} else {
		c.Data["xml"] = serviceResponse
		c.ServeXML()
	}
}
