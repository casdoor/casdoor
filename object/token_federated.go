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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/proxy"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

// FederatedCredential trusts the tokens of an external OIDC issuer (GitHub Actions, Kubernetes,
// a cloud workload identity, another IdP) as a credential of the application.
type FederatedCredential struct {
	Issuer   string `json:"issuer"`
	JwksUri  string `json:"jwksUri"`
	Subject  string `json:"subject"`
	Audience string `json:"audience"`
	User     string `json:"user"`
}

const (
	federatedKeySetMaxAge      = time.Hour
	federatedKeySetMinInterval = time.Minute
	federatedTokenLeeway       = time.Minute
)

var federatedSigningMethods = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "ES256", "ES384", "ES512", "EdDSA"}

type federatedKeySet struct {
	keySet    *jose.JSONWebKeySet
	fetchedAt time.Time
}

var (
	federatedKeySetCache = map[string]*federatedKeySet{}
	federatedKeySetMutex sync.Mutex
)

func fetchFederatedJson(url string, v interface{}) error {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return fmt.Errorf("invalid URL: %s", url)
	}

	client := proxy.GetHttpClient(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch: %s, status: %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func fetchFederatedKeySet(credential *FederatedCredential) (*jose.JSONWebKeySet, error) {
	jwksUri := credential.JwksUri
	if jwksUri == "" {
		discovery := struct {
			Issuer  string `json:"issuer"`
			JwksUri string `json:"jwks_uri"`
		}{}
		err := fetchFederatedJson(strings.TrimSuffix(credential.Issuer, "/")+"/.well-known/openid-configuration", &discovery)
		if err != nil {
			return nil, err
		}
		if discovery.Issuer != credential.Issuer {
			return nil, fmt.Errorf("the issuer in the discovery document: %s does not match: %s", discovery.Issuer, credential.Issuer)
		}
		if discovery.JwksUri == "" {
			return nil, fmt.Errorf("the discovery document of issuer: %s has no jwks_uri", credential.Issuer)
		}
		jwksUri = discovery.JwksUri
	}

	keySet := &jose.JSONWebKeySet{}
	err := fetchFederatedJson(jwksUri, keySet)
	if err != nil {
		return nil, err
	}
	return keySet, nil
}

func findFederatedKey(keySet *jose.JSONWebKeySet, kid string) interface{} {
	keys := []jose.JSONWebKey{}
	for _, key := range keySet.Keys {
		if !key.IsPublic() || (key.Use != "" && key.Use != "sig") {
			continue
		}
		if kid == "" || key.KeyID == kid {
			keys = append(keys, key)
		}
	}

	if len(keys) != 1 {
		return nil
	}
	return keys[0].Key
}

// getFederatedKey reads the issuer's JWKS through a cache. An unknown key ID refetches the set,
// at most once per minute, so a rotated key is picked up without waiting for the cache to expire.
func getFederatedKey(credential *FederatedCredential, kid string) (interface{}, error) {
	federatedKeySetMutex.Lock()
	defer federatedKeySetMutex.Unlock()

	cacheKey := credential.Issuer + "|" + credential.JwksUri
	cached := federatedKeySetCache[cacheKey]
	if cached != nil && time.Since(cached.fetchedAt) < federatedKeySetMaxAge {
		key := findFederatedKey(cached.keySet, kid)
		if key != nil {
			return key, nil
		}
		if time.Since(cached.fetchedAt) < federatedKeySetMinInterval {
			return nil, fmt.Errorf("no signing key matches the key ID: %s for issuer: %s", kid, credential.Issuer)
		}
	}

	keySet, err := fetchFederatedKeySet(credential)
	if err != nil {
		return nil, err
	}
	federatedKeySetCache[cacheKey] = &federatedKeySet{keySet: keySet, fetchedAt: time.Now()}

	key := findFederatedKey(keySet, kid)
	if key == nil {
		return nil, fmt.Errorf("no signing key matches the key ID: %s for issuer: %s", kid, credential.Issuer)
	}
	return key, nil
}

func isFederatedSubjectMatched(pattern string, subject string) bool {
	if pattern == "" || subject == "" {
		return false
	}

	parts := strings.Split(pattern, "*")
	for i, part := range parts {
		parts[i] = regexp.QuoteMeta(part)
	}
	matched, err := regexp.MatchString("^"+strings.Join(parts, ".*")+"$", subject)
	return err == nil && matched
}

func isFederatedAudienceMatched(application *Application, credential *FederatedCredential, audiences []string, host string) bool {
	expected := []string{credential.Audience}
	if credential.Audience == "" {
		_, originBackend := getOriginFromHost(host)
		expected = []string{originBackend, fmt.Sprintf("%s/api/login/oauth/access_token", originBackend), application.ClientId}
	}

	for _, audience := range audiences {
		for _, item := range expected {
			if audience == item {
				return true
			}
		}
	}
	return false
}

// validateFederatedToken verifies a JWT against the federated credentials of the application.
// It returns a nil credential when the token is not a JWT of an issuer the application trusts,
// so that the caller can go on with its own checks; a token of a trusted issuer that fails
// verification is an error.
func validateFederatedToken(application *Application, tokenString string, host string) (*FederatedCredential, jwt.MapClaims, error) {
	if len(application.FederatedCredentials) == 0 || tokenString == "" {
		return nil, nil, nil
	}

	unverifiedClaims := jwt.MapClaims{}
	unverifiedToken, _, err := jwt.NewParser().ParseUnverified(tokenString, unverifiedClaims)
	if err != nil {
		return nil, nil, nil
	}

	issuer, _ := unverifiedClaims.GetIssuer()
	subject, _ := unverifiedClaims.GetSubject()
	if issuer == "" {
		return nil, nil, nil
	}

	var credential *FederatedCredential
	isIssuerTrusted := false
	for _, item := range application.FederatedCredentials {
		if item == nil || item.Issuer != issuer {
			continue
		}
		isIssuerTrusted = true
		if isFederatedSubjectMatched(item.Subject, subject) {
			credential = item
			break
		}
	}
	if !isIssuerTrusted {
		return nil, nil, nil
	}
	if credential == nil {
		return nil, nil, fmt.Errorf("the subject: %s of issuer: %s is not trusted by application: [%s]", subject, issuer, application.GetId())
	}

	kid, _ := unverifiedToken.Header["kid"].(string)
	key, err := getFederatedKey(credential, kid)
	if err != nil {
		return nil, nil, err
	}

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}, jwt.WithValidMethods(federatedSigningMethods), jwt.WithIssuer(credential.Issuer), jwt.WithExpirationRequired(), jwt.WithLeeway(federatedTokenLeeway))
	if err != nil {
		return nil, nil, err
	}

	audiences, err := claims.GetAudience()
	if err != nil {
		return nil, nil, err
	}
	if !isFederatedAudienceMatched(application, credential, audiences, host) {
		return nil, nil, fmt.Errorf("the audience: %v of the token is not accepted by application: [%s]", []string(audiences), application.GetId())
	}

	return credential, claims, nil
}

// getFederatedUser maps a verified federated token to a user of the application. An empty mapping
// returns no user: the token then stands for the application itself, as in client_credentials.
// "{claim}" takes the user from a claim of the token, e.g. "{email}".
func getFederatedUser(application *Application, credential *FederatedCredential, claims jwt.MapClaims) (*User, *TokenError, error) {
	username := credential.User
	if username == "" {
		return nil, nil, nil
	}

	if strings.HasPrefix(username, "{") && strings.HasSuffix(username, "}") {
		claimName := username[1 : len(username)-1]
		username, _ = claims[claimName].(string)
		if username == "" {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: fmt.Sprintf("the token has no claim: %s to map to a user", claimName),
			}, nil
		}
	}

	user, err := GetUserByFieldsForSharedApp(application, application.Organization, username)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the user: %s mapped from the token does not exist", username),
		}, nil
	}
	return user, nil, nil
}
