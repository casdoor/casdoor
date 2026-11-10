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
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
)

// LogoutTokenClaims represents the claims in an OIDC Back-Channel Logout token.
// See https://openid.net/specs/openid-connect-backchannel-1_0.html
type LogoutTokenClaims struct {
	Events map[string]interface{} `json:"events"`
	Sid    string                 `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

func generateLogoutToken(application *Application, user *User, signinToken *Token, sessionId string, host string) (string, error) {
	nowTime := time.Now()
	// The application checks "iss" against the ID token it got, which may have been issued on another host
	// than the one of this request, or without any request at all when the user is forced offline
	issuer := getTokenIssuer(signinToken)
	if issuer == "" {
		var err error
		issuer, err = GetOidcIssuerByApplication(host, application)
		if err != nil {
			return "", err
		}
	}

	events := map[string]interface{}{
		"http://schemas.openid.net/event/backchannel-logout": map[string]interface{}{},
	}

	claims := LogoutTokenClaims{
		Events: events,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   user.Id,
			Audience:  []string{application.ClientId},
			IssuedAt:  jwt.NewNumericDate(nowTime),
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(2 * time.Minute)),
			ID:        util.GetId(application.Owner, util.GenerateId()),
		},
	}

	// The raw id is the Casdoor session cookie, so only its hash leaves Casdoor, the one in the ID token's "sid"
	if sessionId != "" {
		claims.Sid = GetSessionIdHash(sessionId)
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return "", err
	}
	if cert == nil {
		return "", fmt.Errorf("no cert found for application: %s", application.GetId())
	}

	var key interface{}
	token := jwt.NewWithClaims(getJwtSigningMethod(application), claims)
	key, err = getJwtSigningKey(token.Method, cert)
	if err != nil {
		return "", err
	}

	token.Header["kid"] = cert.GetKeyId()
	return signJwtToken(token, key)
}

// SendBackchannelLogout sends OIDC Back-Channel Logout tokens to all registered
// backchannel_logout_uri endpoints for applications that have active tokens for the user.
// With a sessionId, only the applications signed in with that Casdoor session are told and the
// logout token carries its "sid", without one every session of the user at the applications ends.
// See https://openid.net/specs/openid-connect-backchannel-1_0.html
func SendBackchannelLogout(organization, username, sessionId, host string) {
	tokens, err := GetActiveTokensByUser(organization, username)
	if err != nil {
		return
	}
	if sessionId != "" {
		tokens = slices.DeleteFunc(tokens, func(token *Token) bool { return token.SessionId != sessionId })
	}
	if len(tokens) == 0 {
		return
	}

	user, err := GetUser(util.GetId(organization, username))
	if err != nil || user == nil {
		return
	}

	sendBackchannelLogoutForTokens(user, tokens, sessionId, host)
}

// sendBackchannelLogoutForTokens is the body of SendBackchannelLogout for callers that already
// hold the user and its active tokens, so that the token table is not queried a second time
func sendBackchannelLogoutForTokens(user *User, tokens []*Token, sessionId string, host string) {
	// Deduplicate applications
	seen := map[string]bool{}
	for _, token := range tokens {
		appId := util.GetId(token.Owner, token.Application)
		if seen[appId] {
			continue
		}
		seen[appId] = true

		application, err := GetApplication(appId)
		if err != nil || application == nil || application.BackchannelLogoutUri == "" {
			continue
		}

		logoutToken, err := generateLogoutToken(application, user, token, sessionId, host)
		if err != nil {
			continue
		}

		go postLogoutForm(application, application.BackchannelLogoutUri, url.Values{"logout_token": {logoutToken}})
	}
}

func getTokenIssuer(token *Token) string {
	for _, tokenString := range []string{token.IdToken, token.AccessToken} {
		if tokenString == "" {
			continue
		}

		jwtToken, err := ParseJwtTokenWithoutValidation(tokenString)
		if err != nil {
			continue
		}
		if claims, ok := jwtToken.Claims.(*Claims); ok && claims.Issuer != "" {
			return claims.Issuer
		}
	}
	return ""
}

func postLogoutForm(application *Application, logoutUrl string, form url.Values) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if application.Organization != "built-in" {
		client.Transport = util.NewNonLocalHttpTransport(30 * time.Second)
	}

	resp, err := client.PostForm(logoutUrl, form)
	if err != nil {
		fmt.Printf("postLogoutForm() error: %s\n", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("postLogoutForm() error: unexpected status: %s from: %s\n", resp.Status, logoutUrl)
	}
}
