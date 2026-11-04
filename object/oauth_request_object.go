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
	"net/url"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

var RequestObjectSigningAlgs = []string{"HS256", "HS384", "HS512"}

var requestObjectJwtClaims = map[string]bool{
	"iss": true,
	"aud": true,
	"sub": true,
	"exp": true,
	"iat": true,
	"nbf": true,
	"jti": true,
}

// ResolveOAuthRequestObject verifies the "request" parameter of an authorization request (RFC 9101) with the
// client secret and returns the query with the request object's claims merged in, which take precedence
func ResolveOAuthRequestObject(query url.Values) (url.Values, error) {
	clientId := query.Get("client_id")
	if clientId == "" {
		return nil, fmt.Errorf("client_id is required when the request parameter is used")
	}

	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, fmt.Errorf("the application for client_id: %s does not exist", clientId)
	}
	if application.ClientSecret == "" {
		return nil, fmt.Errorf("the application: %s has no client secret to verify the request object", application.Name)
	}

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(query.Get("request"), claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(application.ClientSecret), nil
	}, jwt.WithValidMethods(RequestObjectSigningAlgs))
	if err != nil {
		return nil, fmt.Errorf("invalid request object: %s", err.Error())
	}

	if value, ok := claims["client_id"]; ok && value != clientId {
		return nil, fmt.Errorf("the client_id in the request object does not match the client_id parameter")
	}

	res := url.Values{}
	for key, values := range query {
		if key != "request" {
			res[key] = values
		}
	}

	for key, value := range claims {
		if requestObjectJwtClaims[key] {
			continue
		}

		switch v := value.(type) {
		case string:
			res.Set(key, v)
		case float64:
			res.Set(key, strconv.FormatFloat(v, 'f', -1, 64))
		case bool:
			res.Set(key, strconv.FormatBool(v))
		default:
			bytes, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			res.Set(key, string(bytes))
		}
	}

	return res, nil
}
