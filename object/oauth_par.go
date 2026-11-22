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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
)

const (
	ParRequestUriPrefix = "urn:ietf:params:oauth:request_uri:"
	ParExpiresIn        = 90
	// the user may take a while to sign in after the authorization request arrives
	parAuthorizationLifetime = 30 * time.Minute
)

type ParResponse struct {
	RequestUri string `json:"request_uri"`
	ExpiresIn  int    `json:"expires_in"`
}

type pushedAuthRequest struct {
	ClientId   string
	Query      url.Values
	ExpireTime time.Time
}

var (
	parMutex          sync.Mutex
	pushedAuthReqs    = map[string]*pushedAuthRequest{}
	parAuthorizations = map[string]time.Time{}
)

// parClientParams are the client authentication parameters, which never go into the authorization request
var parClientParams = []string{"client_secret", "client_assertion", "client_assertion_type"}

// AuthenticateOAuthClient authenticates a client at a back-channel endpoint with client_secret_basic,
// client_secret_post or a client assertion. A public client without a secret is let through only when allowPublic is set
func AuthenticateOAuthClient(clientId string, clientSecret string, clientAssertion string, clientAssertionType string, host string, allowPublic bool) (*Application, *TokenError, error) {
	if clientAssertionType == "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		ok, application, err := ValidateClientAssertion(clientAssertion, clientId, host)
		if err != nil {
			return nil, nil, err
		}
		if !ok || application == nil {
			return nil, &TokenError{Error: InvalidClient, ErrorDescription: "client_assertion is invalid"}, nil
		}
		return application, nil, nil
	}

	if clientId == "" {
		return nil, &TokenError{Error: InvalidClient, ErrorDescription: "client_id is required"}, nil
	}

	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return nil, nil, err
	}
	if application == nil {
		return nil, &TokenError{Error: InvalidClient, ErrorDescription: "client_id is invalid"}, nil
	}

	if clientSecret == "" && allowPublic {
		return application, nil, nil
	}
	if !application.IsClientSecretValid(clientSecret) {
		return nil, &TokenError{Error: InvalidClient, ErrorDescription: "client_secret is invalid"}, nil
	}
	return application, nil, nil
}

// PushAuthorizationRequest stores the authorization request a client pushed to the PAR endpoint (RFC 9126)
// and returns the request_uri that the client sends the user agent to the authorization endpoint with
func PushAuthorizationRequest(application *Application, params url.Values, isPublicClient bool, lang string) (*ParResponse, *TokenError, error) {
	if params.Get("request_uri") != "" {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: "request_uri must not be pushed"}, nil
	}

	query := url.Values{}
	for key, values := range params {
		if !util.InSlice(parClientParams, key) {
			query[key] = values
		}
	}
	query.Set("client_id", application.ClientId)

	if query.Get("request") != "" {
		resolved, err := ResolveOAuthRequestObject(query)
		if err != nil {
			return nil, &TokenError{Error: "invalid_request_object", ErrorDescription: err.Error()}, nil
		}
		query = resolved
	}

	// a public client proves it is the one that redeems the code with PKCE
	if isPublicClient && query.Get("code_challenge") == "" {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: "a client without a client secret must use PKCE (code_challenge)"}, nil
	}

	msg, _, err := CheckOAuthLogin(application.ClientId, "", query.Get("response_type"), query.Get("redirect_uri"), query.Get("scope"), query.Get("state"), query.Get("nonce"), lang)
	if err != nil {
		return nil, nil, err
	}
	if msg != "" {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: msg}, nil
	}

	requestUri := ParRequestUriPrefix + util.GenerateClientSecret()

	parMutex.Lock()
	defer parMutex.Unlock()

	now := time.Now()
	removeExpiredParEntries(now)
	pushedAuthReqs[requestUri] = &pushedAuthRequest{
		ClientId:   application.ClientId,
		Query:      query,
		ExpireTime: now.Add(ParExpiresIn * time.Second),
	}

	return &ParResponse{RequestUri: requestUri, ExpiresIn: ParExpiresIn}, nil, nil
}

// ResolvePushedAuthorizationRequest replaces the request_uri of an authorization request with the parameters that
// were pushed for it. A request_uri can be used once, and the expanded request is remembered as one that came by PAR
func ResolvePushedAuthorizationRequest(query url.Values) (url.Values, error) {
	requestUri := query.Get("request_uri")
	clientId := query.Get("client_id")
	if clientId == "" {
		return nil, fmt.Errorf("client_id is required when the request_uri parameter is used")
	}
	if !strings.HasPrefix(requestUri, ParRequestUriPrefix) {
		return nil, fmt.Errorf("request_uri: %s was not issued by the pushed authorization request endpoint", requestUri)
	}

	parMutex.Lock()
	defer parMutex.Unlock()

	now := time.Now()
	removeExpiredParEntries(now)

	pushed, ok := pushedAuthReqs[requestUri]
	if !ok {
		return nil, fmt.Errorf("request_uri is invalid, expired or already used")
	}
	if pushed.ClientId != clientId {
		return nil, fmt.Errorf("the client_id does not match the pushed authorization request")
	}
	delete(pushedAuthReqs, requestUri)

	res := url.Values{}
	for key, values := range pushed.Query {
		res[key] = values
	}

	parAuthorizations[getParAuthorizationKey(clientId, res.Get("redirect_uri"), res.Get("state"), res.Get("nonce"), res.Get("code_challenge"))] = now.Add(parAuthorizationLifetime)
	return res, nil
}

// consumeParAuthorization tells whether an authorization request came through the PAR endpoint, for an
// application that accepts no other authorization requests
func consumeParAuthorization(clientId string, redirectUri string, state string, nonce string, challenge string) bool {
	key := getParAuthorizationKey(clientId, redirectUri, state, nonce, challenge)

	parMutex.Lock()
	defer parMutex.Unlock()

	expireTime, ok := parAuthorizations[key]
	if !ok {
		return false
	}
	delete(parAuthorizations, key)
	return time.Now().Before(expireTime)
}

func getParAuthorizationKey(clientId string, redirectUri string, state string, nonce string, challenge string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{clientId, redirectUri, state, nonce, challenge}, "\n")))
	return hex.EncodeToString(sum[:])
}

func removeExpiredParEntries(now time.Time) {
	for key, pushed := range pushedAuthReqs {
		if now.After(pushed.ExpireTime) {
			delete(pushedAuthReqs, key)
		}
	}
	for key, expireTime := range parAuthorizations {
		if now.After(expireTime) {
			delete(parAuthorizations, key)
		}
	}
}
