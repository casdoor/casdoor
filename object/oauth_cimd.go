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
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
)

const (
	cimdTag             = "cimd"
	cimdOrganization    = "built-in"
	cimdMaxDocumentSize = 16 * 1024
	cimdFetchTimeout    = 5 * time.Second
	cimdRefreshInterval = time.Hour
)

// ClientIdMetadataDocument is the document a client publishes at the URL it uses as client_id
// (draft-ietf-oauth-client-id-metadata-document, used by MCP clients instead of registering)
type ClientIdMetadataDocument struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	DynamicClientRegistrationRequest
}

var (
	cimdMutex     sync.Mutex
	cimdFetchedAt = map[string]time.Time{}
)

// IsClientIdMetadataUrl reports whether a client_id is the URL of a client ID metadata document:
// https, with a path, without credentials, fragment or dot segments
func IsClientIdMetadataUrl(clientId string) bool {
	if !strings.HasPrefix(clientId, "https://") || strings.Contains(clientId, "#") {
		return false
	}

	u, err := url.Parse(clientId)
	if err != nil || u.Host == "" || u.User != nil || u.Path == "" || u.Path == "/" {
		return false
	}

	for _, segment := range strings.Split(u.Path, "/") {
		if segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

// IsClientIdMetadataDocumentSupported reports whether URL client_ids are accepted, which follows
// the dynamic client registration switch of the organization they are added to
func IsClientIdMetadataDocumentSupported() bool {
	org, err := GetOrganization(util.GetId("admin", cimdOrganization))
	if err != nil || org == nil {
		return false
	}
	return org.DcrPolicy != "" && org.DcrPolicy != "disabled"
}

// GetApplicationByClientIdMetadata returns the application of a client_id. When the client_id is the URL
// of a client ID metadata document, the document is fetched and kept as an application tagged "cimd",
// fetched again once it is older than cimdRefreshInterval.
func GetApplicationByClientIdMetadata(clientId string, lang string) (*Application, error) {
	if !IsClientIdMetadataUrl(clientId) {
		return GetApplicationByClientId(clientId)
	}

	cimdMutex.Lock()
	defer cimdMutex.Unlock()

	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return nil, err
	}
	if application != nil && !util.InSlice(application.Tags, cimdTag) {
		return application, nil
	}

	if !IsClientIdMetadataDocumentSupported() {
		return nil, fmt.Errorf("client ID metadata documents are not accepted, enable dynamic client registration of the organization: %s first", cimdOrganization)
	}

	if application != nil && time.Since(cimdFetchedAt[clientId]) < cimdRefreshInterval {
		return application, nil
	}

	doc, err := fetchClientIdMetadataDocument(clientId)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(clientId)
	host := u.Hostname()
	displayName := firstNonEmpty(doc.ClientName, host)
	if !strings.EqualFold(displayName, host) {
		displayName = fmt.Sprintf("%s (%s)", displayName, host)
	}
	homepageUrl := fmt.Sprintf("https://%s", u.Host)
	if clientUri, err := url.Parse(doc.ClientUri); err == nil && doc.ClientUri != "" && strings.EqualFold(clientUri.Hostname(), host) {
		homepageUrl = doc.ClientUri
	}

	if application == nil {
		org, err := GetOrganization(util.GetId("admin", cimdOrganization))
		if err != nil {
			return nil, err
		}

		application = newDynamicClientApplication(&doc.DynamicClientRegistrationRequest, org, cimdTag)
		application.ClientId = clientId
		application.DisplayName = displayName
		application.HomepageUrl = homepageUrl

		affected, err := AddApplication(application, lang)
		if err != nil {
			return nil, err
		}
		if !affected {
			return nil, fmt.Errorf("failed to add the application for the client: %s", clientId)
		}
	} else {
		application.DisplayName = displayName
		application.HomepageUrl = homepageUrl
		application.RedirectUris = doc.RedirectUris
		application.GrantTypes = doc.GrantTypes
		application.TermsOfUse = doc.TosUri
		if doc.LogoUri != "" {
			application.Logo = doc.LogoUri
		}

		_, err = UpdateApplication(application.GetId(), application, true, lang, nil)
		if err != nil {
			return nil, err
		}
	}

	cimdFetchedAt[clientId] = time.Now()
	return GetApplicationByClientId(clientId)
}

func fetchClientIdMetadataDocument(clientId string) (*ClientIdMetadataDocument, error) {
	client := util.NewInternetOnlyHttpClient(cimdFetchTimeout)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest(http.MethodGet, clientId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the client metadata document: %s, %s", clientId, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch the client metadata document: %s, status: %s", clientId, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, cimdMaxDocumentSize+1))
	if err != nil {
		return nil, err
	}
	if len(body) > cimdMaxDocumentSize {
		return nil, fmt.Errorf("the client metadata document: %s is larger than %d bytes", clientId, cimdMaxDocumentSize)
	}

	doc := &ClientIdMetadataDocument{}
	err = json.Unmarshal(body, doc)
	if err != nil {
		return nil, fmt.Errorf("the client metadata document: %s is not valid JSON, %s", clientId, err.Error())
	}

	err = checkClientIdMetadataDocument(doc, clientId)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func checkClientIdMetadataDocument(doc *ClientIdMetadataDocument, clientId string) error {
	if doc.ClientId != clientId {
		return fmt.Errorf("the client_id: %s in the client metadata document doesn't match its URL: %s", doc.ClientId, clientId)
	}
	if doc.ClientSecret != "" {
		return fmt.Errorf("the client metadata document: %s must not contain client_secret", clientId)
	}
	if doc.TokenEndpointAuthMethod != "" && doc.TokenEndpointAuthMethod != "none" {
		return fmt.Errorf("the token_endpoint_auth_method: %s of the client: %s is not supported, only \"none\" with PKCE is", doc.TokenEndpointAuthMethod, clientId)
	}
	if len(doc.RedirectUris) == 0 {
		return fmt.Errorf("the client metadata document: %s has no redirect_uris", clientId)
	}

	// Only full URLs, so that a redirect URI is never read as a host pattern
	for _, redirectUri := range doc.RedirectUris {
		u, err := url.Parse(redirectUri)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("the redirect URI: %s of the client: %s must be a full URL", redirectUri, clientId)
		}
	}

	if dcrErr := checkDcrClientMetadata(&doc.DynamicClientRegistrationRequest); dcrErr != nil {
		return fmt.Errorf("%s: %s", dcrErr.Error, dcrErr.ErrorDescription)
	}
	if len(doc.GrantTypes) == 0 {
		doc.GrantTypes = []string{"authorization_code"}
	}
	return nil
}
