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

package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type OidcDiscovery struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JwksUri               string   `json:"jwks_uri"`
	EndSessionEndpoint    string   `json:"end_session_endpoint"`
	ScopesSupported       []string `json:"scopes_supported"`
}

func GetOidcDiscovery(issuer string) (*OidcDiscovery, error) {
	issuer = strings.TrimSuffix(strings.TrimSpace(issuer), "/")
	if issuer == "" {
		return nil, fmt.Errorf("the issuer is empty")
	}
	if !strings.HasPrefix(issuer, "http://") && !strings.HasPrefix(issuer, "https://") {
		issuer = fmt.Sprintf("https://%s", issuer)
	}

	discoveryUrl := issuer
	if !strings.HasSuffix(discoveryUrl, "/.well-known/openid-configuration") {
		discoveryUrl = fmt.Sprintf("%s/.well-known/openid-configuration", discoveryUrl)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(discoveryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get OIDC discovery from: %s, status: %d", discoveryUrl, resp.StatusCode)
	}

	discovery := &OidcDiscovery{}
	err = json.Unmarshal(data, discovery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OIDC discovery from: %s, error: %v", discoveryUrl, err)
	}

	if discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" {
		return nil, fmt.Errorf("the OIDC discovery from: %s doesn't contain the required endpoints", discoveryUrl)
	}

	return discovery, nil
}

type OidcIdProvider struct {
	CustomIdProvider
}

func NewOidcIdProvider(idpInfo *ProviderInfo, redirectUrl string) *OidcIdProvider {
	idp := &OidcIdProvider{}

	idp.Config = &oauth2.Config{
		ClientID:     idpInfo.ClientId,
		ClientSecret: idpInfo.ClientSecret,
		RedirectURL:  redirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  idpInfo.AuthURL,
			TokenURL: idpInfo.TokenURL,
		},
	}
	idp.UserInfoURL = idpInfo.UserInfoURL
	idp.UserMapping = idpInfo.UserMapping
	idp.Issuer = strings.TrimSuffix(idpInfo.HostUrl, "/")
	idp.CodeVerifier = idpInfo.CodeVerifier
	return idp
}

// resolveEndpoints fills the missing endpoints from the issuer's OIDC discovery document
func (idp *OidcIdProvider) resolveEndpoints() error {
	if idp.Issuer == "" || (idp.Config.Endpoint.TokenURL != "" && idp.UserInfoURL != "") {
		return nil
	}

	discovery, err := GetOidcDiscovery(idp.Issuer)
	if err != nil {
		return err
	}

	if idp.Config.Endpoint.AuthURL == "" {
		idp.Config.Endpoint.AuthURL = discovery.AuthorizationEndpoint
	}
	if idp.Config.Endpoint.TokenURL == "" {
		idp.Config.Endpoint.TokenURL = discovery.TokenEndpoint
	}
	if idp.UserInfoURL == "" {
		idp.UserInfoURL = discovery.UserinfoEndpoint
	}
	return nil
}

func (idp *OidcIdProvider) GetToken(code string) (*oauth2.Token, error) {
	err := idp.resolveEndpoints()
	if err != nil {
		return nil, err
	}
	return idp.CustomIdProvider.GetToken(code)
}

func (idp *OidcIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	err := idp.resolveEndpoints()
	if err != nil {
		return nil, err
	}
	return idp.CustomIdProvider.GetUserInfo(token)
}
