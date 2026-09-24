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

package object

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/casdoor/casdoor/util"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
)

// SAML Single Logout, see the SAML 2.0 core spec section 3.7 and the profiles spec section 4.4
const (
	samlRedirectBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
	samlPostBinding     = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
	samlStatusSuccess   = "urn:oasis:names:tc:SAML:2.0:status:Success"
)

type SamlLogoutRequest struct {
	XMLName        xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol LogoutRequest"`
	ID             string   `xml:"ID,attr"`
	Issuer         string   `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	NameId         string   `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	SessionIndexes []string `xml:"urn:oasis:names:tc:SAML:2.0:protocol SessionIndex"`
}

func getSamlSingleLogoutLocation(application *Application, host string) string {
	_, originBackend := getOriginFromHost(host)
	return fmt.Sprintf("%s/api/saml/logout/%s/%s", originBackend, application.Owner, application.Name)
}

// ParseSamlLogoutRequest parses a LogoutRequest an SP sent to the application's SingleLogoutService
func ParseSamlLogoutRequest(application *Application, samlRequest string) (*SamlLogoutRequest, error) {
	data, err := decodeSamlMessage(samlRequest)
	if err != nil {
		return nil, err
	}

	var request SamlLogoutRequest
	err = xml.Unmarshal(data, &request)
	if err != nil {
		return nil, fmt.Errorf("err: Failed to unmarshal LogoutRequest, please check the SAML request, %s", err.Error())
	}

	if !application.IsRedirectUriValid(request.Issuer) {
		return nil, fmt.Errorf("err: Issuer URI: %s doesn't exist in the allowed Redirect URI list", request.Issuer)
	}
	if request.NameId == "" {
		return nil, fmt.Errorf("err: SAML request doesn't have <saml:NameID> in <samlp:LogoutRequest>")
	}

	return &request, nil
}

// CheckSamlLogoutResponse validates the LogoutResponse an SP returns for a LogoutRequest of Casdoor
func CheckSamlLogoutResponse(application *Application, samlResponse string) error {
	data, err := decodeSamlMessage(samlResponse)
	if err != nil {
		return err
	}

	var response types.LogoutResponse
	err = xml.Unmarshal(data, &response)
	if err != nil {
		return fmt.Errorf("err: Failed to unmarshal LogoutResponse, %s", err.Error())
	}

	if response.Issuer == nil || !application.IsRedirectUriValid(response.Issuer.Value) {
		return fmt.Errorf("err: the issuer of the LogoutResponse doesn't exist in the allowed Redirect URI list")
	}
	if response.Status == nil || response.Status.StatusCode == nil || response.Status.StatusCode.Value != samlStatusSuccess {
		return fmt.Errorf("err: the SP failed to log out")
	}

	return nil
}

// LogoutBySamlRequest ends the Casdoor sessions an SP-initiated LogoutRequest points to, and tells
// whether the browser session the request came with is one of them
func LogoutBySamlRequest(application *Application, request *SamlLogoutRequest, currentUser string, currentSessionId string, host string) (bool, error) {
	targets := []*SamlSession{}
	if len(request.SessionIndexes) > 0 {
		err := ormer.Engine.In("session_index", request.SessionIndexes).Where("application = ? and name_id = ?", application.GetId(), request.NameId).Find(&targets)
		if err != nil {
			return false, err
		}
	} else if currentUser != "" {
		// The request is not signed, so without a SessionIndex only the browser's own session is
		// ended and not every session of the principal
		user, err := GetUser(currentUser)
		if err != nil {
			return false, err
		}
		if user != nil {
			if nameId, _ := getSamlNameId(application, user); nameId == request.NameId {
				targets = append(targets, &SamlSession{Owner: user.Owner, Name: user.Name, SessionId: currentSessionId})
			}
		}
	}

	isCurrentSession := false
	for _, target := range targets {
		// The requesting SP has already logged out, it must not get a LogoutRequest of its own
		_, err := ormer.Engine.Where("owner = ? and name = ? and application = ? and session_id = ?", target.Owner, target.Name, application.GetId(), target.SessionId).Delete(&SamlSession{})
		if err != nil {
			return false, err
		}

		user, err := GetUser(util.GetId(target.Owner, target.Name))
		if err != nil {
			return false, err
		}
		if user == nil {
			continue
		}

		err = logoutUserSession(user, target.SessionId, host)
		if err != nil {
			return false, err
		}

		if util.GetId(target.Owner, target.Name) == currentUser && target.SessionId == currentSessionId {
			isCurrentSession = true
		}
	}

	return isCurrentSession, nil
}

// logoutUserSession ends one Casdoor session and tells the OIDC and SAML applications signed in with it
func logoutUserSession(user *User, sessionId string, host string) error {
	tokens, err := GetActiveTokensByUser(user.Owner, user.Name)
	if err != nil {
		return err
	}

	sessionTokens := []*Token{}
	for _, token := range tokens {
		if token.SessionId == sessionId {
			sessionTokens = append(sessionTokens, token)
		}
	}

	sendBackchannelLogoutForTokens(user, sessionTokens, sessionId, host)
	SendSamlLogout(user.Owner, user.Name, []string{sessionId}, host)

	err = DeleteUserSessionId(user.Owner, user.Name, sessionId)
	if err != nil {
		return err
	}
	DeleteBeegoSession([]string{sessionId})

	go func() {
		_ = SendSsoLogoutNotifications(user, []string{sessionId}, sessionTokens)
	}()

	return nil
}

// SendSamlLogout sends a signed LogoutRequest to the SingleLogoutService of every SAML SP the user
// signed in to, only for the given Casdoor sessions unless sessionIds is nil. Like the OIDC
// Back-Channel Logout, it is posted from the server and does not go through the browser.
func SendSamlLogout(owner string, name string, sessionIds []string, host string) {
	if sessionIds != nil && len(sessionIds) == 0 {
		return
	}

	samlSessions, err := getUserSamlSessions(owner, name, sessionIds)
	if err != nil {
		return
	}

	for _, samlSession := range samlSessions {
		err = deleteSamlSession(samlSession)
		if err != nil {
			continue
		}

		application, err := GetApplication(samlSession.Application)
		if err != nil || application == nil || application.SamlSingleLogoutUrl == "" {
			continue
		}

		logoutRequest, err := newSamlLogoutRequest(application, samlSession, host)
		if err != nil {
			continue
		}

		go postSamlLogoutRequest(application.SamlSingleLogoutUrl, logoutRequest)
	}
}

func newSamlLogoutRequest(application *Application, samlSession *SamlSession, host string) (string, error) {
	_, originBackend := getOriginFromHost(host)

	request := &etree.Element{Space: "samlp", Tag: "LogoutRequest"}
	request.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	request.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	request.CreateAttr("ID", fmt.Sprintf("_%s", util.GenerateUUID()))
	request.CreateAttr("Version", "2.0")
	request.CreateAttr("IssueInstant", time.Now().UTC().Format(time.RFC3339))
	request.CreateAttr("Destination", application.SamlSingleLogoutUrl)
	request.CreateElement("saml:Issuer").SetText(originBackend)
	nameId := request.CreateElement("saml:NameID")
	nameId.CreateAttr("Format", samlSession.NameIdFormat)
	nameId.SetText(samlSession.NameId)
	request.CreateElement("samlp:SessionIndex").SetText(samlSession.SessionIndex)

	ctx, err := getSamlLogoutSigningContext(application)
	if err != nil {
		return "", err
	}

	return signSamlPostMessage(ctx, request)
}

func postSamlLogoutRequest(logoutUrl string, logoutRequest string) {
	resp, err := http.PostForm(logoutUrl, url.Values{"SAMLRequest": {logoutRequest}})
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// GetSamlLogoutResponse builds the LogoutResponse for an SP-initiated LogoutRequest. Over the HTTP-POST
// binding it returns the base64 message to post to the SP, over the HTTP-Redirect binding the URL to redirect to.
func GetSamlLogoutResponse(application *Application, request *SamlLogoutRequest, relayState string, host string, usePost bool) (string, error) {
	_, originBackend := getOriginFromHost(host)

	response := &etree.Element{Space: "samlp", Tag: "LogoutResponse"}
	response.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	response.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	response.CreateAttr("ID", fmt.Sprintf("_%s", util.GenerateUUID()))
	response.CreateAttr("Version", "2.0")
	response.CreateAttr("IssueInstant", time.Now().UTC().Format(time.RFC3339))
	response.CreateAttr("Destination", application.SamlSingleLogoutUrl)
	if request.ID != "" {
		response.CreateAttr("InResponseTo", request.ID)
	}
	response.CreateElement("saml:Issuer").SetText(originBackend)
	response.CreateElement("samlp:Status").CreateElement("samlp:StatusCode").CreateAttr("Value", samlStatusSuccess)

	ctx, err := getSamlLogoutSigningContext(application)
	if err != nil {
		return "", err
	}

	if usePost {
		return signSamlPostMessage(ctx, response)
	}
	return getSamlRedirectUrl(ctx, application.SamlSingleLogoutUrl, "SAMLResponse", response, relayState)
}

func getSamlLogoutSigningContext(application *Application) (*dsig.SigningContext, error) {
	cert, err := getCertByApplication(application)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, fmt.Errorf("please set a cert for the application first")
	}

	block, _ := pem.Decode([]byte(cert.Certificate))
	if block == nil {
		return nil, fmt.Errorf("the certificate field should not be empty for the cert: %s", cert.GetId())
	}
	certificate := base64.StdEncoding.EncodeToString(block.Bytes)

	return newSamlSigningContext(application, cert.PrivateKey, certificate), nil
}

// signSamlPostMessage embeds the signature in the message, as the HTTP-POST binding requires
func signSamlPostMessage(ctx *dsig.SigningContext, element *etree.Element) (string, error) {
	signature, err := ctx.ConstructSignature(element, true)
	if err != nil {
		return "", err
	}
	// the signature goes right after <saml:Issuer>
	element.InsertChildAt(1, signature)

	doc := etree.NewDocument()
	doc.SetRoot(element)
	xmlBytes, err := doc.WriteToBytes()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(xmlBytes), nil
}

// getSamlRedirectUrl deflates the message and signs the query string instead of the XML, as the HTTP-Redirect binding requires
func getSamlRedirectUrl(ctx *dsig.SigningContext, destination string, param string, element *etree.Element, relayState string) (string, error) {
	doc := etree.NewDocument()
	doc.SetRoot(element)
	xmlBytes, err := doc.WriteToBytes()
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	writer, err := flate.NewWriter(&buffer, flate.DefaultCompression)
	if err != nil {
		return "", err
	}
	_, err = writer.Write(xmlBytes)
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf("%s=%s", param, url.QueryEscape(base64.StdEncoding.EncodeToString(buffer.Bytes())))
	if relayState != "" {
		query += fmt.Sprintf("&RelayState=%s", url.QueryEscape(relayState))
	}
	query += fmt.Sprintf("&SigAlg=%s", url.QueryEscape(ctx.GetSignatureMethodIdentifier()))

	signature, err := ctx.SignString(query)
	if err != nil {
		return "", err
	}
	query += fmt.Sprintf("&Signature=%s", url.QueryEscape(base64.StdEncoding.EncodeToString(signature)))

	if strings.Contains(destination, "?") {
		return fmt.Sprintf("%s&%s", destination, query), nil
	}
	return fmt.Sprintf("%s?%s", destination, query), nil
}
