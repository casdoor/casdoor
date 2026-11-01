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
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
)

var (
	idJagUsedJtis      = map[string]time.Time{}
	idJagUsedJtisMutex sync.Mutex
)

// IsIdJag reports whether a JWT bearer assertion is an ID-JAG, by its "typ" header
func IsIdJag(assertion string) bool {
	token, _, err := jwt.NewParser().ParseUnverified(assertion, jwt.MapClaims{})
	if err != nil {
		return false
	}
	typ, _ := token.Header["typ"].(string)
	return strings.EqualFold(typ, "oauth-id-jag+jwt")
}

// GetIdJagGrantToken redeems an ID-JAG at the JWT bearer grant (RFC 7523): Casdoor is the authorization server
// of an MCP server, and the ID-JAG is issued by an enterprise IdP that the organization of its audience trusts
func GetIdJagGrantToken(application *Application, clientSecret string, assertion string, scope string, host string, clientIp string, lang string) (*Token, *TokenError, error) {
	// a public client of a client ID metadata document (Claude) has nothing to authenticate with, its ID-JAG is
	// bound to its client_id and redeemed only once
	isPublicClient := util.InSlice(application.Tags, cimdTag) && getPrivateKeyJwtCredential(application) == nil
	if !isPublicClient && application.ClientSecret != clientSecret {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "the client must authenticate to redeem an ID-JAG",
		}, nil
	}

	unverifiedClaims := jwt.MapClaims{}
	unverifiedToken, _, err := jwt.NewParser().ParseUnverified(assertion, unverifiedClaims)
	if err != nil {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: err.Error()}, nil
	}

	audiences, err := unverifiedClaims.GetAudience()
	if err != nil {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: err.Error()}, nil
	}
	organization, err := getIdJagOrganization(application, audiences, host)
	if err != nil {
		return nil, nil, err
	}
	if organization == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the audience: %v of the ID-JAG is not this authorization server", []string(audiences)),
		}, nil
	}

	if application.Organization != organization.Name {
		if !application.IsShared || (isDynamicClient(application) && !IsClientIdMetadataDocumentSupported(organization.Name)) {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: fmt.Sprintf("the application: %s does not serve the organization: %s", application.Name, organization.Name),
			}, nil
		}
		sharedApplication := *application
		sharedApplication.Organization = organization.Name
		application = &sharedApplication
	}

	issuer, _ := unverifiedClaims.GetIssuer()
	var credential *FederatedCredential
	for _, item := range organization.IdJagIssuers {
		if item != nil && item.Issuer != "" && item.Issuer == issuer {
			credential = item
			break
		}
	}
	if credential == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the issuer: %s of the ID-JAG is not trusted by the organization: %s", issuer, organization.Name),
		}, nil
	}

	kid, _ := unverifiedToken.Header["kid"].(string)
	key, err := getFederatedKey(credential, kid, false)
	if err != nil {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: err.Error()}, nil
	}

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(assertion, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}, jwt.WithValidMethods(federatedSigningMethods), jwt.WithIssuer(credential.Issuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(federatedTokenLeeway))
	if err != nil {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: err.Error()}, nil
	}

	clientId, _ := claims["client_id"].(string)
	if clientId != application.ClientId {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the ID-JAG was issued to the client: %s, not: %s", clientId, application.ClientId),
		}, nil
	}

	jti, _ := claims["jti"].(string)
	if jti == "" {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "the ID-JAG has no jti"}, nil
	}
	expiresAt, _ := claims.GetExpirationTime()
	if !useIdJagJti(issuer+"|"+jti, expiresAt.Time.Add(federatedTokenLeeway)) {
		return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "the ID-JAG has already been used"}, nil
	}

	user, tokenError, err := getIdJagUser(organization, credential, claims)
	if err != nil || tokenError != nil {
		return nil, tokenError, err
	}
	if tokenError = checkGrantUserSignin(application, user, clientIp, lang); tokenError != nil {
		return nil, tokenError, nil
	}

	resource, _ := claims["resource"].(string)
	if err = ValidateResourceURI(resource); err != nil {
		return nil, &TokenError{Error: InvalidTarget, ErrorDescription: err.Error()}, nil
	}

	grantedScope, _ := claims["scope"].(string)
	if scope == "" {
		scope = grantedScope
	} else {
		for _, item := range strings.Fields(scope) {
			if !util.InSlice(strings.Fields(grantedScope), item) {
				return nil, &TokenError{
					Error:            InvalidScope,
					ErrorDescription: fmt.Sprintf("the scope: %s is not granted by the ID-JAG", item),
				}, nil
			}
		}
	}

	return mintTokenForUser(application, user, scope, "", host, resource)
}

// getIdJagOrganization finds the organization whose issuer is the audience of an ID-JAG: an organization with
// its own issuer, or the organization of the application for the global issuer
func getIdJagOrganization(application *Application, audiences []string, host string) (*Organization, error) {
	globalIssuer := GetOidcIssuer(host)
	for _, audience := range audiences {
		organizationName := application.Organization
		if audience != globalIssuer {
			if !strings.HasPrefix(audience, globalIssuer+"/orgs/") {
				continue
			}
			name, err := url.PathUnescape(strings.TrimPrefix(audience, globalIssuer+"/orgs/"))
			if err != nil || name == "" || strings.Contains(name, "/") {
				continue
			}
			organizationName = name
		}

		organization, err := getOrganization("admin", organizationName)
		if err != nil {
			return nil, err
		}
		if organization != nil && getOrganizationIssuerOrGlobal(host, organization) == audience {
			return organization, nil
		}
	}
	return nil, nil
}

// getIdJagUser maps the user of an ID-JAG to a user of the organization, by email unless the trust says otherwise
func getIdJagUser(organization *Organization, credential *FederatedCredential, claims jwt.MapClaims) (*User, *TokenError, error) {
	username := credential.User
	if username == "" {
		username = "{email}"
	}

	if strings.HasPrefix(username, "{") && strings.HasSuffix(username, "}") {
		claimName := username[1 : len(username)-1]
		username, _ = claims[claimName].(string)
		if username == "" {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: fmt.Sprintf("the ID-JAG has no claim: %s to map to a user", claimName),
			}, nil
		}
	}

	user, err := GetUserByFields(organization.Name, username)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the user: %s of the ID-JAG does not exist in the organization: %s", username, organization.Name),
		}, nil
	}
	return user, nil, nil
}

// useIdJagJti marks the jti of an ID-JAG as used, so that an ID-JAG is redeemed only once
func useIdJagJti(jti string, expiresAt time.Time) bool {
	idJagUsedJtisMutex.Lock()
	defer idJagUsedJtisMutex.Unlock()

	now := time.Now()
	for key, value := range idJagUsedJtis {
		if now.After(value) {
			delete(idJagUsedJtis, key)
		}
	}

	if _, ok := idJagUsedJtis[jti]; ok {
		return false
	}
	idJagUsedJtis[jti] = expiresAt
	return true
}
