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
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/email"
	"github.com/casdoor/casdoor/notification"
	"github.com/casdoor/casdoor/util"
	notify "github.com/casdoor/notify2"
)

// isTenantProvider reports a provider an organization admin manages, as opposed to the global
// ones ("admin") and the ones of "built-in", which only global admins manage. The hosts of a
// tenant provider must not reach the intranet, like the webhooks of the organizations.
func isTenantProvider(provider *Provider) bool {
	return provider.Owner != "admin" && provider.Owner != "built-in"
}

func checkTenantProviderHost(provider *Provider, host string) error {
	if !isTenantProvider(provider) {
		return nil
	}
	return util.CheckInternetHost(host)
}

func getTenantHttpClient(provider *Provider) *http.Client {
	if !isTenantProvider(provider) {
		return nil
	}
	return util.NewInternetOnlyHttpClient(30 * time.Second)
}

func getProviderHttpClient(provider *Provider) *http.Client {
	if client := getTenantHttpClient(provider); client != nil {
		return client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func restrictTenantEmailProvider(organization string, provider *Provider, emailProvider email.EmailProvider) error {
	switch p := emailProvider.(type) {
	case *email.HttpEmailProvider:
		if isTenantProvider(provider) {
			networks, err := getTenantPrivateEmailNetworks(organization, provider)
			if err != nil {
				return err
			}
			if len(networks) != 0 {
				p.SetHttpClient(util.NewPrivateEndpointHttpClient(30*time.Second, networks))
				return nil
			}
		}
		p.SetHttpClient(getTenantHttpClient(provider))
	case *email.Microsoft365EmailProvider:
		p.SetHttpClient(getTenantHttpClient(provider))
	}
	return nil
}

type privateEmailEndpoint struct {
	AllowedCIDRs []string `json:"allowedCIDRs"`
	Headers      []string `json:"headers"`
}

func getTenantPrivateEmailGrants() (map[string]json.RawMessage, error) {
	value := strings.TrimSpace(conf.GetConfigString("tenantPrivateEmailEndpoints"))
	if value == "" {
		return nil, nil
	}

	var grants map[string]json.RawMessage
	if err := json.Unmarshal([]byte(value), &grants); err != nil || grants == nil {
		return nil, errors.New("tenantPrivateEmailEndpoints must be a JSON object")
	}
	return grants, nil
}

func parseTenantPrivateEmailOrganization(organization string, value json.RawMessage) (map[string]json.RawMessage, error) {
	var endpoints map[string]json.RawMessage
	if err := json.Unmarshal(value, &endpoints); err != nil || endpoints == nil {
		return nil, fmt.Errorf("tenantPrivateEmailEndpoints for organization %q must be a JSON object", organization)
	}
	return endpoints, nil
}

func ValidateTenantPrivateEmailConfig() error {
	grants, err := getTenantPrivateEmailGrants()
	if err != nil {
		return err
	}
	var invalid []error
	for organization, value := range grants {
		if _, err := parseTenantPrivateEmailOrganization(organization, value); err != nil {
			invalid = append(invalid, err)
		}
	}
	return errors.Join(invalid...)
}

func getTenantPrivateEmailNetworks(organization string, provider *Provider) ([]netip.Prefix, error) {
	grants, err := getTenantPrivateEmailGrants()
	if err != nil {
		return nil, nil
	}
	organizationGrants, configured := grants[provider.Owner]
	if !configured {
		return nil, nil
	}
	endpoints, err := parseTenantPrivateEmailOrganization(provider.Owner, organizationGrants)
	if err != nil {
		return nil, err
	}
	entry, configured := endpoints[provider.Endpoint]
	if !configured {
		return nil, nil
	}
	if organization == "" || organization != provider.Owner {
		return nil, errors.New("private email endpoints can only be used by the provider's organization")
	}
	var grant privateEmailEndpoint
	if err := json.Unmarshal(entry, &grant); err != nil {
		return nil, errors.New("the private email endpoint configuration is invalid")
	}

	endpoint, err := url.Parse(provider.Endpoint)
	if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() == "" || endpoint.User != nil ||
		endpoint.ForceQuery || endpoint.RawQuery != "" || strings.ContainsAny(provider.Endpoint, "#*") ||
		strings.Contains(endpoint.Hostname(), "%") {
		return nil, errors.New("the private email endpoint must be an exact HTTPS URL")
	}
	if provider.Method != http.MethodPost ||
		(provider.IssuerUrl != "" && provider.IssuerUrl != "application/x-www-form-urlencoded") {
		return nil, errors.New("private email endpoints require a form-encoded POST")
	}

	for header := range provider.HttpHeaders {
		switch http.CanonicalHeaderKey(header) {
		case "Host", "Content-Type":
			return nil, errors.New("private email providers cannot override Host or Content-Type")
		}
		if !slices.ContainsFunc(grant.Headers, func(allowed string) bool {
			return strings.EqualFold(header, allowed)
		}) {
			return nil, errors.New("a custom header is not allowed for the private email endpoint")
		}
	}

	if len(grant.AllowedCIDRs) == 0 {
		return nil, errors.New("the private email endpoint must have at least one allowed CIDR")
	}
	networks := make([]netip.Prefix, 0, len(grant.AllowedCIDRs))
	for _, value := range grant.AllowedCIDRs {
		network, err := netip.ParsePrefix(value)
		if err != nil || network.Bits() == 0 || network.Addr().Is4In6() {
			return nil, errors.New("the private email endpoint has an invalid allowed CIDR")
		}
		networks = append(networks, network)
	}
	return networks, nil
}

func RestrictNotificationClient(client notify.Notifier, httpClient *http.Client) {
	if httpNotificationClient, ok := client.(*notification.HttpNotificationClient); ok {
		httpNotificationClient.SetHttpClient(httpClient)
	}
}
