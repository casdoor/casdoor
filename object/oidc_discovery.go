// Copyright 2021 The casbin Authors. All Rights Reserved.
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

type OidcDiscovery struct {
	Issuer                                 string   `json:"issuer"`
	AuthorizationEndpoint                  string   `json:"authorization_endpoint"`
	JwksUri                                string   `json:"jwks_uri"`
	ResponseTypesSupported                 []string `json:"response_types_supported"`
	ResponseModesSupported                 []string `json:"response_modes_supported"`
	GrantTypesSupported                    []string `json:"grant_types_supported"`
	SubjectTypesSupported                  []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported       []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                        []string `json:"scopes_supported"`
	ClaimsSupported                        []string `json:"claims_supported"`
	RequestParameterSupported              bool     `json:"request_parameter_supported"`
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported"`
}

var oidcDiscovery OidcDiscovery

func init() {
	oidcDiscovery = OidcDiscovery{
		Issuer:                                 "",
		AuthorizationEndpoint:                  "",
		JwksUri:                                "",
		ResponseTypesSupported:                 nil,
		ResponseModesSupported:                 nil,
		GrantTypesSupported:                    nil,
		SubjectTypesSupported:                  nil,
		IdTokenSigningAlgValuesSupported:       nil,
		ScopesSupported:                        nil,
		ClaimsSupported:                        nil,
		RequestParameterSupported:              false,
		RequestObjectSigningAlgValuesSupported: nil,
	}
}

func GetOidcDiscovery() OidcDiscovery {
	return oidcDiscovery
}
