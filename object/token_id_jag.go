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
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
)

// Identity Assertion JWT Authorization Grant (ID-JAG), as used by MCP Enterprise-Managed Authorization.
// See https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/
const (
	IdJagTokenType = "urn:ietf:params:oauth:token-type:id-jag"
	idJagExpiresIn = 300
)

type IdJagClaims struct {
	Email    string `json:"email,omitempty"`
	ClientId string `json:"client_id"`
	Resource string `json:"resource,omitempty"`
	Scope    string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type IdJagTokenWrapper struct {
	IssuedTokenType string `json:"issued_token_type"`
	AccessToken     string `json:"access_token"`
	TokenType       string `json:"token_type"`
	ExpiresIn       int    `json:"expires_in"`
	Scope           string `json:"scope,omitempty"`
}

// GetIdJagToken exchanges an ID token of the user for an ID-JAG, which the client then redeems at the
// authorization server of an MCP server. The MCP server is a Casdoor server (or application), and only
// the users granted by an "Allow" permission on its application get one.
func GetIdJagToken(application *Application, clientSecret string, subjectToken string, subjectTokenType string, audience string, resource string, scope string, host string) (*IdJagTokenWrapper, *TokenError, error) {
	if application.ClientSecret != clientSecret {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}

	if subjectToken == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "subject_token is required",
		}, nil
	}
	if subjectTokenType != "urn:ietf:params:oauth:token-type:id_token" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: fmt.Sprintf("unsupported subject_token_type for an ID-JAG: %s", subjectTokenType),
		}, nil
	}
	if audience == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "audience is required for an ID-JAG",
		}, nil
	}

	user, _, tokenError, err := getSubjectTokenUser(application, subjectToken, subjectTokenType)
	if err != nil || tokenError != nil {
		return nil, tokenError, err
	}

	targetApplication, tokenError, err := getIdJagTargetApplication(application, user, audience, resource)
	if err != nil || tokenError != nil {
		return nil, tokenError, err
	}

	isAllowed, err := checkIdJagPermission(user, targetApplication)
	if err != nil {
		return nil, nil, err
	}
	if !isAllowed {
		return nil, &TokenError{
			Error:            InvalidTarget,
			ErrorDescription: fmt.Sprintf("the user: %s is not allowed to access the application: %s", user.GetId(), targetApplication.Name),
		}, nil
	}

	expandedScope, ok := IsScopeValidAndExpand(scope, targetApplication)
	if !ok {
		return nil, &TokenError{
			Error:            InvalidScope,
			ErrorDescription: fmt.Sprintf("the requested scope is invalid or not defined in the application: %s", targetApplication.Name),
		}, nil
	}

	nowTime := time.Now()
	issuer, err := GetOidcIssuerByApplication(host, application)
	if err != nil {
		return nil, nil, err
	}
	name := util.GenerateId()
	claims := IdJagClaims{
		Email:    user.Email,
		ClientId: application.ClientId,
		Resource: resource,
		Scope:    expandedScope,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   user.Id,
			Audience:  []string{audience},
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(idJagExpiresIn * time.Second)),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			ID:        util.GetId(application.Owner, name),
		},
	}

	idJag, err := signIdJag(application, claims)
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate ID-JAG error: %s", err.Error()),
		}, nil
	}

	token := &Token{
		Owner:        application.Owner,
		Name:         name,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateAuthorizationCode(),
		AccessToken:  idJag,
		ExpiresIn:    idJagExpiresIn,
		Scope:        expandedScope,
		TokenType:    "ID-JAG",
		CodeIsUsed:   true,
		Resource:     util.ReturnAnyNotEmpty(resource, audience),
		GrantType:    "urn:ietf:params:oauth:grant-type:token-exchange",
	}
	_, err = AddToken(token)
	if err != nil {
		return nil, nil, err
	}

	return &IdJagTokenWrapper{
		IssuedTokenType: IdJagTokenType,
		AccessToken:     idJag,
		TokenType:       "N_A",
		ExpiresIn:       idJagExpiresIn,
		Scope:           expandedScope,
	}, nil, nil
}

// getIdJagTargetApplication finds the application of the MCP server: by the server whose URL is the resource,
// or else by the audience as a client ID or application name
func getIdJagTargetApplication(application *Application, user *User, audience string, resource string) (*Application, *TokenError, error) {
	var targetApplication *Application
	if resource != "" {
		servers, err := GetServers(user.Owner)
		if err != nil {
			return nil, nil, err
		}
		for _, server := range servers {
			if server.Application != "" && strings.TrimSuffix(server.Url, "/") == strings.TrimSuffix(resource, "/") {
				targetApplication, err = getApplication("admin", server.Application)
				if err != nil {
					return nil, nil, err
				}
				break
			}
		}
		if targetApplication == nil {
			return nil, &TokenError{
				Error:            InvalidTarget,
				ErrorDescription: fmt.Sprintf("the requested resource is not a known server with an application: %s", resource),
			}, nil
		}
	} else {
		var err error
		targetApplication, err = GetApplicationByClientId(audience)
		if err != nil {
			return nil, nil, err
		}
		if targetApplication == nil {
			targetApplication, err = getApplication(application.Owner, audience)
			if err != nil {
				return nil, nil, err
			}
		}
		if targetApplication == nil {
			return nil, &TokenError{
				Error:            InvalidTarget,
				ErrorDescription: fmt.Sprintf("the requested audience is not a known application: %s, pass the MCP server URL as the resource", audience),
			}, nil
		}
	}

	if user.Owner != targetApplication.Organization && !targetApplication.IsShared {
		return nil, &TokenError{
			Error:            InvalidTarget,
			ErrorDescription: fmt.Sprintf("the application: %s does not serve the organization of the user: %s", targetApplication.Name, user.GetId()),
		}, nil
	}

	return targetApplication, nil, nil
}

// checkIdJagPermission is the sign-in permission check, except that an application nobody is granted stays closed
func checkIdJagPermission(user *User, application *Application) (bool, error) {
	permissionOrganization := application.Organization
	if application.IsShared {
		permissionOrganization = user.Owner
	}

	permissions, err := GetPermissions(permissionOrganization)
	if err != nil {
		return false, err
	}

	isGranted := false
	for _, permission := range permissions {
		if permission.IsEnabled && permission.State == "Approved" && permission.ResourceType == "Application" && permission.Effect == "Allow" && permission.isResourceHit(application.Name) {
			isGranted = true
			break
		}
	}
	if !isGranted {
		return false, nil
	}

	return CheckLoginPermission(user.GetId(), application)
}

func signIdJag(application *Application, claims IdJagClaims) (string, error) {
	cert, err := getCertByApplication(application)
	if err != nil {
		return "", err
	}
	if cert == nil {
		return "", fmt.Errorf("no cert found for application: %s", application.GetId())
	}

	var key interface{}
	token := jwt.NewWithClaims(getJwtSigningMethod(application), claims)
	if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok {
		key, err = jwt.ParseECPrivateKeyFromPEM([]byte(cert.PrivateKey))
	} else {
		key, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(cert.PrivateKey))
	}
	if err != nil {
		return "", err
	}

	token.Header["typ"] = "oauth-id-jag+jwt"
	token.Header["kid"] = cert.Name
	return token.SignedString(key)
}
