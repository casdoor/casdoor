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
	"strings"

	"github.com/casdoor/casdoor/util"
)

func addNativeSsoDeviceSecret(token *Token) string {
	for _, item := range strings.Fields(token.Scope) {
		if item == "device_sso" {
			deviceSecret := util.GenerateClientSecret()
			token.DeviceSecretHash = getTokenHash(deviceSecret)
			token.DeviceSecretExpiresIn = token.ExpiresIn
			return deviceSecret
		}
	}

	return ""
}

func getTokenByDeviceSecret(deviceSecret string) (*Token, error) {
	if deviceSecret == "" {
		return nil, nil
	}

	token := Token{DeviceSecretHash: getTokenHash(deviceSecret)}
	existed, err := ormer.Engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, nil
	}
	return &token, nil
}

func getNativeSsoTokenExchangeToken(application *Application, clientSecret string, subjectToken string, subjectTokenType string, actorToken string, actorTokenType string, scope string, host string) (*Token, *TokenError, error) {
	if clientSecret != "" && application.ClientSecret != clientSecret {
		return nil, &TokenError{Error: InvalidClient, ErrorDescription: "client_secret is invalid"}, nil
	}
	if subjectToken == "" {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: "subject_token is required"}, nil
	}
	if subjectTokenType != "urn:ietf:params:oauth:token-type:access_token" {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: fmt.Sprintf("unsupported subject_token_type for native sso: %s", subjectTokenType)}, nil
	}
	if actorToken == "" {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: "actor_token is required"}, nil
	}
	if actorTokenType != "urn:openid:params:token-type:device-secret" {
		return nil, &TokenError{Error: InvalidRequest, ErrorDescription: fmt.Sprintf("unsupported actor_token_type: %s", actorTokenType)}, nil
	}

	deviceToken, err := getTokenByDeviceSecret(actorToken)
	if err != nil {
		return nil, nil, err
	}
	if deviceToken == nil || deviceToken.DeviceSecretExpiresIn <= 0 {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "device_secret is invalid"}, nil
	}
	if expired, _ := util.IsTokenExpired(deviceToken.CreatedTime, deviceToken.DeviceSecretExpiresIn); expired {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "device_secret is expired"}, nil
	}
	subjectTokenRecord, err := GetTokenByAccessToken(subjectToken)
	if err != nil {
		return nil, nil, err
	}
	if subjectTokenRecord == nil || subjectTokenRecord.GetId() != deviceToken.GetId() {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "subject_token is not bound to device_secret"}, nil
	}
	if expired, _ := util.IsTokenExpired(subjectTokenRecord.CreatedTime, subjectTokenRecord.ExpiresIn); expired {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "subject_token is expired"}, nil
	}
	if clientSecret == "" {
		deviceApplication, err := getApplication(deviceToken.Owner, deviceToken.Application)
		if err != nil {
			return nil, nil, err
		}
		if deviceApplication == nil || deviceApplication.Organization != application.Organization {
			return nil, &TokenError{Error: InvalidClient, ErrorDescription: "client_secret is required for native sso across organizations"}, nil
		}
	}

	if deviceToken.Organization != subjectTokenRecord.Organization || deviceToken.User != subjectTokenRecord.User {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "subject_token is not bound to device_secret"}, nil
	}

	if scope == "" {
		items := []string{}
		for _, item := range strings.Fields(deviceToken.Scope) {
			if item != "device_sso" {
				items = append(items, item)
			}
		}
		scope = strings.Join(items, " ")
	}
	return mintNativeSsoToken(application, subjectTokenRecord.Organization, subjectTokenRecord.User, scope, host)
}

func mintNativeSsoToken(application *Application, owner string, name string, scope string, host string) (*Token, *TokenError, error) {
	user, err := getUser(owner, name)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("user from subject_token does not exist: %s", util.GetId(owner, name))}, nil
	}
	if user.IsForbidden || user.IsDeleted {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "the user is forbidden to sign in, please contact the administrator"}, nil
	}
	if application.DisableSignin {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("the application: %s has disabled users to signin", application.Name)}, nil
	}
	if application.OrganizationObj != nil && application.OrganizationObj.DisableSignin {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("the organization: %s has disabled users to signin", application.Organization)}, nil
	}
	allowed, err := CheckLoginPermission(user.GetId(), application)
	if err != nil {
		return nil, nil, err
	}
	if !allowed {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "unauthorized operation"}, nil
	}
	if !IsScopeValid(scope, application) {
		return nil, &TokenError{Error: InvalidScope, ErrorDescription: "invalid scope"}, nil
	}
	if err = ExtendUserWithRolesAndPermissions(user); err != nil {
		return nil, nil, err
	}

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, "", "", "", scope, "", host)
	if err != nil {
		return nil, &TokenError{Error: EndpointError, ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error())}, nil
	}
	token := &Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateClientId(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(application.ExpireInHours * float64(hourSeconds)),
		Scope:        scope,
		TokenType:    "Bearer",
		GrantType:    "urn:ietf:params:oauth:grant-type:token-exchange",
		CodeIsUsed:   true,
	}
	if _, err = AddToken(token); err != nil {
		return nil, nil, err
	}
	return token, nil, nil
}
