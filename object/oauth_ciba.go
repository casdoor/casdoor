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
	"html"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/casdoor/casdoor/util"
)

const (
	CibaGrantType          = "urn:openid:params:grant-type:ciba"
	CibaDefaultExpiresIn   = 300
	CibaMaxExpiresIn       = 1800
	CibaInterval           = 5
	cibaMaxBindingMessage  = 100
	CibaStatusPending      = "Pending"
	CibaStatusApproved     = "Approved"
	CibaStatusDenied       = "Denied"
	CibaStatusExpired      = "Expired"
	cibaErrorUnknownUserId = "unknown_user_id"
)

type CibaResponse struct {
	AuthReqId string `json:"auth_req_id"`
	ExpiresIn int    `json:"expires_in"`
	Interval  int    `json:"interval"`
}

type CibaRequest struct {
	AuthReqId      string    `json:"-"`
	Id             string    `json:"id"`
	ClientId       string    `json:"-"`
	Organization   string    `json:"organization"`
	Application    string    `json:"application"`
	AppDisplayName string    `json:"appDisplayName"`
	AppLogo        string    `json:"appLogo"`
	UserId         string    `json:"userId"`
	Scope          string    `json:"scope"`
	BindingMessage string    `json:"bindingMessage"`
	ClientIp       string    `json:"clientIp"`
	Status         string    `json:"status"`
	CreatedTime    time.Time `json:"createdTime"`
	ExpireTime     time.Time `json:"expireTime"`
	Interval       int       `json:"-"`
	LastPollTime   time.Time `json:"-"`
}

var (
	cibaMutex    sync.Mutex
	cibaRequests = map[string]*CibaRequest{}
	// the id in the link sent to the user, which is not the auth_req_id that the client polls with
	cibaApprovalIds = map[string]string{}
)

// StartCibaRequest handles a backchannel authentication request (OpenID Connect CIBA Core 1.0, poll mode): it finds
// the user from the hint and sends the user a link to approve the sign-in on their own device
func StartCibaRequest(application *Application, scope string, loginHint string, idTokenHint string, loginHintToken string, bindingMessage string, requestedExpiry string, clientIp string, host string) (*CibaResponse, *TokenError, error) {
	if !IsGrantTypeValid(CibaGrantType, application.GrantTypes) {
		return nil, &TokenError{Error: UnauthorizedClient, ErrorDescription: fmt.Sprintf("grant_type: %s is not enabled for this application", CibaGrantType)}, nil
	}

	if !util.InSlice(strings.Fields(scope), "openid") {
		return nil, &TokenError{Error: InvalidScope, ErrorDescription: "the scope must contain openid"}, nil
	}
	if !IsScopeValid(scope, application) {
		return nil, &TokenError{Error: InvalidScope, ErrorDescription: "the requested scope is invalid or not defined in the application"}, nil
	}

	hintCount := 0
	for _, hint := range []string{loginHint, idTokenHint, loginHintToken} {
		if hint != "" {
			hintCount++
		}
	}
	if hintCount != 1 {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: "exactly one of login_hint, id_token_hint and login_hint_token is required"}, nil
	}
	if loginHintToken != "" {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: "login_hint_token is not supported, use login_hint or id_token_hint"}, nil
	}

	if utf8.RuneCountInString(bindingMessage) > cibaMaxBindingMessage {
		return nil, &TokenError{Error: "invalid_binding_message", ErrorDescription: fmt.Sprintf("binding_message can have at most %d characters", cibaMaxBindingMessage)}, nil
	}

	expiresIn := CibaDefaultExpiresIn
	if requestedExpiry != "" {
		value, err := strconv.Atoi(requestedExpiry)
		if err != nil || value <= 0 {
			return nil, &TokenError{Error: InvalidRequest, ErrorDescription: "requested_expiry must be a positive integer"}, nil
		}
		expiresIn = min(value, CibaMaxExpiresIn)
	}

	user, tokenError, err := getCibaUser(application, loginHint, idTokenHint)
	if err != nil || tokenError != nil {
		return nil, tokenError, err
	}

	msg, err := checkOAuthCodeUser(user, application, "en")
	if err != nil {
		return nil, nil, err
	}
	if msg != "" || user.IsForbidden {
		return nil, &TokenError{Error: AccessDenied, ErrorDescription: "the user can't sign in to this application"}, nil
	}

	now := time.Now()
	request := &CibaRequest{
		AuthReqId:      util.GenerateAuthorizationCode(),
		Id:             util.GenerateClientSecret(),
		ClientId:       application.ClientId,
		Organization:   user.Owner,
		Application:    application.Name,
		AppDisplayName: application.DisplayName,
		AppLogo:        application.Logo,
		UserId:         user.GetId(),
		Scope:          scope,
		BindingMessage: bindingMessage,
		ClientIp:       clientIp,
		Status:         CibaStatusPending,
		CreatedTime:    now,
		ExpireTime:     now.Add(time.Duration(expiresIn) * time.Second),
		Interval:       CibaInterval,
	}

	tokenError, err = sendCibaNotification(application, user, request, host)
	if err != nil || tokenError != nil {
		return nil, tokenError, err
	}

	cibaMutex.Lock()
	removeExpiredCibaRequests(now)
	cibaRequests[request.AuthReqId] = request
	cibaApprovalIds[request.Id] = request.AuthReqId
	cibaMutex.Unlock()

	return &CibaResponse{AuthReqId: request.AuthReqId, ExpiresIn: expiresIn, Interval: CibaInterval}, nil, nil
}

func getCibaUser(application *Application, loginHint string, idTokenHint string) (*User, *TokenError, error) {
	if idTokenHint != "" {
		claims, err := ParseJwtTokenByApplication(idTokenHint, application)
		if err != nil || claims == nil || claims.User == nil {
			return nil, &TokenError{Error: cibaErrorUnknownUserId, ErrorDescription: "id_token_hint is invalid"}, nil
		}

		user, err := GetUser(util.GetId(claims.User.Owner, claims.User.Name))
		if err != nil {
			return nil, nil, err
		}
		if user == nil || user.IsDeleted {
			return nil, &TokenError{Error: cibaErrorUnknownUserId, ErrorDescription: "the user of id_token_hint does not exist"}, nil
		}
		return user, nil, nil
	}

	user, err := GetUserByFields(application.Organization, loginHint)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || user.IsDeleted {
		return nil, &TokenError{Error: cibaErrorUnknownUserId, ErrorDescription: "no user matches login_hint"}, nil
	}
	return user, nil, nil
}

func sendCibaNotification(application *Application, user *User, request *CibaRequest, host string) (*TokenError, error) {
	if user.Email == "" {
		return &TokenError{Error: AccessDenied, ErrorDescription: "the user has no email address to receive the authentication request"}, nil
	}

	provider, err := application.GetEmailProvider("All")
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &TokenError{Error: UnauthorizedClient, ErrorDescription: "the application has no Email provider to notify the user"}, nil
	}

	appName := application.DisplayName
	if appName == "" {
		appName = application.Name
	}

	originFrontend, _ := GetOriginFromHost(host)
	link := fmt.Sprintf("%s/ciba/%s/%s", originFrontend, url.PathEscape(user.Owner), request.Id)
	minutes := max(1, int(request.ExpireTime.Sub(request.CreatedTime).Minutes()))

	title := fmt.Sprintf("Approve the sign-in to %s", appName)
	content := fmt.Sprintf("<p>%s is asking you to sign in as %s.</p>", html.EscapeString(appName), html.EscapeString(user.Name))
	if request.BindingMessage != "" {
		content += fmt.Sprintf("<p>Message: <b>%s</b></p>", html.EscapeString(request.BindingMessage))
	}
	expiry := fmt.Sprintf("%d minutes", minutes)
	if minutes == 1 {
		expiry = "1 minute"
	}
	content += fmt.Sprintf("<p>If you asked for it, open the link and approve it within %s. Otherwise ignore this email.</p><p><a href=\"%s\">%s</a></p>", expiry, link, link)

	organizationName := user.Owner
	organization, err := GetOrganization(util.GetId("admin", user.Owner))
	if err == nil && organization != nil && organization.DisplayName != "" {
		organizationName = organization.DisplayName
	}

	err = SendEmail(provider, title, content, []string{user.Email}, organizationName)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// GetCibaRequest returns a backchannel authentication request for the approval page of its user
func GetCibaRequest(id string, userId string) (*CibaRequest, error) {
	cibaMutex.Lock()
	defer cibaMutex.Unlock()

	request := cibaRequests[cibaApprovalIds[id]]
	if request == nil || request.UserId != userId {
		return nil, fmt.Errorf("the sign-in request does not exist or has expired")
	}

	res := *request
	if res.Status == CibaStatusPending && time.Now().After(res.ExpireTime) {
		res.Status = CibaStatusExpired
	}
	return &res, nil
}

// AnswerCibaRequest records the user's answer to a backchannel authentication request
func AnswerCibaRequest(id string, userId string, approve bool) (string, error) {
	cibaMutex.Lock()
	defer cibaMutex.Unlock()

	request := cibaRequests[cibaApprovalIds[id]]
	if request == nil || request.UserId != userId {
		return "", fmt.Errorf("the sign-in request does not exist or has expired")
	}
	if time.Now().After(request.ExpireTime) {
		return CibaStatusExpired, fmt.Errorf("the sign-in request has expired")
	}
	if request.Status != CibaStatusPending {
		return request.Status, nil
	}

	if approve {
		request.Status = CibaStatusApproved
	} else {
		request.Status = CibaStatusDenied
	}
	return request.Status, nil
}

// GetCibaToken handles the token request of a client that polls with an auth_req_id
func GetCibaToken(clientId string, clientSecret string, clientAssertion string, clientAssertionType string, authReqId string, host string) (interface{}, error) {
	application, tokenError, err := AuthenticateOAuthClient(clientId, clientSecret, clientAssertion, clientAssertionType, host, false)
	if err != nil {
		return nil, err
	}
	if tokenError != nil {
		return tokenError, nil
	}

	if !IsGrantTypeValid(CibaGrantType, application.GrantTypes) {
		return &TokenError{Error: UnauthorizedClient, ErrorDescription: fmt.Sprintf("grant_type: %s is not enabled for this application", CibaGrantType)}, nil
	}
	if authReqId == "" {
		return &TokenError{Error: InvalidRequest, ErrorDescription: "auth_req_id is required"}, nil
	}

	request, tokenError := claimCibaRequest(application.ClientId, authReqId)
	if tokenError != nil {
		return tokenError, nil
	}

	user, err := GetUser(request.UserId)
	if err != nil {
		return nil, err
	}

	token, tokenError, err := mintTokenForUser(application, user, request.Scope, "", host, "")
	if err != nil {
		return nil, err
	}
	if tokenError != nil {
		return tokenError, nil
	}

	token.CodeIsUsed = true
	token.GrantType = CibaGrantType
	_, err = updateUsedByCode(token)
	if err != nil {
		return nil, err
	}

	return &TokenWrapper{
		AccessToken:  token.AccessToken,
		IdToken:      token.IdToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
		Scope:        token.Scope,
	}, nil
}

// claimCibaRequest answers a poll with the state of the request, and hands an approved request out once
func claimCibaRequest(clientId string, authReqId string) (*CibaRequest, *TokenError) {
	cibaMutex.Lock()
	defer cibaMutex.Unlock()

	request := cibaRequests[authReqId]
	if request == nil || request.ClientId != clientId {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "auth_req_id is invalid"}
	}

	now := time.Now()
	if now.After(request.ExpireTime) {
		deleteCibaRequest(request)
		return nil, &TokenError{Error: "expired_token", ErrorDescription: "the authentication request has expired"}
	}

	switch request.Status {
	case CibaStatusApproved:
		deleteCibaRequest(request)
		return request, nil
	case CibaStatusDenied:
		deleteCibaRequest(request)
		return nil, &TokenError{Error: AccessDenied, ErrorDescription: "the user denied the authentication request"}
	}

	if !request.LastPollTime.IsZero() && now.Sub(request.LastPollTime) < time.Duration(request.Interval)*time.Second {
		request.Interval += CibaInterval
		request.LastPollTime = now
		return nil, &TokenError{Error: "slow_down", ErrorDescription: fmt.Sprintf("poll at most once every %d seconds", request.Interval)}
	}
	request.LastPollTime = now
	return nil, &TokenError{Error: "authorization_pending", ErrorDescription: "the user hasn't approved the authentication request yet"}
}

func deleteCibaRequest(request *CibaRequest) {
	delete(cibaRequests, request.AuthReqId)
	delete(cibaApprovalIds, request.Id)
}

func removeExpiredCibaRequests(now time.Time) {
	for _, request := range cibaRequests {
		if now.After(request.ExpireTime) {
			deleteCibaRequest(request)
		}
	}
}
