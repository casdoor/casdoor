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
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
)

// LogoutTokenClaims represents the claims in an OIDC Back-Channel Logout token.
// See https://openid.net/specs/openid-connect-backchannel-1_0.html
type LogoutTokenClaims struct {
	Events map[string]interface{} `json:"events"`
	jwt.RegisteredClaims
}

func generateLogoutToken(application *Application, user *User, sessionId string, host string) (string, error) {
	nowTime := time.Now()
	_, originBackend := getOriginFromHost(host)

	events := map[string]interface{}{
		"http://schemas.openid.net/event/backchannel-logout": map[string]interface{}{},
	}

	claims := LogoutTokenClaims{
		Events: events,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   originBackend,
			Subject:  user.Id,
			Audience: []string{application.ClientId},
			IssuedAt: jwt.NewNumericDate(nowTime),
			ID:       util.GetId(application.Owner, util.GenerateId()),
		},
	}

	if sessionId != "" {
		// sid claim carries the session ID per the spec
		claims.RegisteredClaims.ID = sessionId
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return "", err
	}
	if cert == nil {
		return "", fmt.Errorf("no cert found for application: %s", application.GetId())
	}

	var (
		token *jwt.Token
		key   interface{}
	)

	signingMethod := application.TokenSigningMethod
	if strings.Contains(signingMethod, "ES") {
		token = jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		key, err = jwt.ParseECPrivateKeyFromPEM([]byte(cert.PrivateKey))
	} else {
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		key, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(cert.PrivateKey))
	}
	if err != nil {
		return "", err
	}

	token.Header["kid"] = cert.Name
	return token.SignedString(key)
}

// SendBackchannelLogout sends OIDC Back-Channel Logout tokens to all registered
// backchannel_logout_uri endpoints for applications that have active tokens for the user.
// See https://openid.net/specs/openid-connect-backchannel-1_0.html
func SendBackchannelLogout(organization, username, sessionId, host string) {
	tokens, err := GetActiveTokensByUser(organization, username)
	if err != nil || len(tokens) == 0 {
		return
	}

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

		user, err := GetUser(util.GetId(organization, username))
		if err != nil || user == nil {
			continue
		}

		logoutToken, err := generateLogoutToken(application, user, sessionId, host)
		if err != nil {
			continue
		}

		go postBackchannelLogout(application.BackchannelLogoutUri, logoutToken)
	}
}

func postBackchannelLogout(logoutUri, logoutToken string) {
	body := url.Values{"logout_token": {logoutToken}}
	resp, err := http.PostForm(logoutUri, body)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
