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
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
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
		// sid claim carries the session ID per the spec; jti stays a unique token id
		claims.Sid = sessionId
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

// getClientIdFromClaims returns the first non-empty audience (client_id).
func getClientIdFromClaims(claims *Claims) string {
	for _, aud := range claims.Audience {
		if aud != "" {
			return aud
		}
	}
	return ""
}

// ExpireTokenByLogoutHint expires the token referenced by an "id_token_hint" from
// RP-initiated logout. The hint is usually an id token, which won't match the
// access token hash, so in that case we verify it against the application cert and
// expire the active token(s) for that user and application. Only the current
// application's token is expired so the back-channel fan-out can still reach the
// other applications.
func ExpireTokenByLogoutHint(hint string) (bool, *Application, *Token, error) {
	// Some clients send the access token as the hint.
	if affected, application, token, err := ExpireTokenByAccessToken(hint); err == nil && token != nil {
		return affected, application, token, nil
	}

	parsed, err := ParseJwtTokenWithoutValidation(hint)
	if err != nil {
		return false, nil, nil, fmt.Errorf("invalid id_token_hint: %w", err)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || claims.User == nil {
		return false, nil, nil, fmt.Errorf("invalid id_token_hint: missing user claims")
	}

	clientId := getClientIdFromClaims(claims)
	if clientId == "" {
		return false, nil, nil, fmt.Errorf("invalid id_token_hint: missing audience")
	}

	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return false, nil, nil, err
	}
	if application == nil {
		return false, nil, nil, fmt.Errorf("invalid id_token_hint: application not found for client_id %s", clientId)
	}

	// Verify the hint was actually issued and signed by this IdP.
	if _, err = ParseJwtTokenByApplication(hint, application); err != nil {
		return false, application, nil, fmt.Errorf("invalid id_token_hint signature: %w", err)
	}

	affected, token, err := expireActiveTokensByUserApplication(claims.User.Owner, claims.User.Name, application.Name)
	if err != nil {
		return false, application, nil, err
	}

	return affected, application, token, nil
}

// SendBackchannelLogout sends OIDC Back-Channel Logout tokens to all registered
// backchannel_logout_uri endpoints for applications that have active tokens for the user.
// See https://openid.net/specs/openid-connect-backchannel-1_0.html
func SendBackchannelLogout(organization, username, sessionId, host string) {
	tokens, err := GetActiveTokensByUser(organization, username)
	if err != nil {
		logs.Warning("backchannel logout: failed to load active tokens for %s/%s: %v", organization, username, err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	user, err := GetUser(util.GetId(organization, username))
	if err != nil || user == nil {
		logs.Warning("backchannel logout: user %s/%s not found: %v", organization, username, err)
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
		if err != nil || application == nil {
			logs.Warning("backchannel logout: application %s not found: %v", appId, err)
			continue
		}
		if application.BackchannelLogoutUri == "" {
			continue
		}

		logoutToken, err := generateLogoutToken(application, user, sessionId, host)
		if err != nil {
			logs.Warning("backchannel logout: failed to build logout token for %s: %v", appId, err)
			continue
		}

		go postBackchannelLogout(application.BackchannelLogoutUri, logoutToken)
	}
}

func postBackchannelLogout(logoutUri, logoutToken string) {
	body := url.Values{"logout_token": {logoutToken}}

	resp, err := http.PostForm(logoutUri, body)
	if err != nil {
		logs.Warning("backchannel logout request to %s failed: %v", logoutUri, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		logs.Warning("backchannel logout rejected by %s: status=%d body=%s", logoutUri, resp.StatusCode, string(respBody))
	}
}
