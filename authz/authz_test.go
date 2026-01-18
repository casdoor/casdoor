// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package authz

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/qiangmzsx/string-adapter/v2"
)

func TestWellKnownEndpointsWithKeyMatch2(t *testing.T) {
	// Create the model with keyMatch2 for URL matching
	modelText := `[request_definition]
r = subOwner, subName, method, urlPath, objOwner, objName

[policy_definition]
p = subOwner, subName, method, urlPath, objOwner, objName

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (r.subOwner == p.subOwner || p.subOwner == "*") && \
    (r.subName == p.subName || p.subName == "*" || r.subName != "anonymous" && p.subName == "!anonymous") && \
    (r.method == p.method || p.method == "*") && \
    (keyMatch2(r.urlPath, p.urlPath) || p.urlPath == "*") && \
    (r.objOwner == p.objOwner || p.objOwner == "*") && \
    (r.objName == p.objName || p.objName == "*") || \
    (r.subOwner == r.objOwner && r.subName == r.objName)`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create policy rules for well-known endpoints
	policyText := `
p, *, *, GET, /.well-known/openid-configuration, *, *
p, *, *, GET, /.well-known/webfinger, *, *
p, *, *, *, /.well-known/jwks, *, *
p, *, *, GET, /.well-known/:application/openid-configuration, *, *
p, *, *, GET, /.well-known/:application/webfinger, *, *
p, *, *, *, /.well-known/:application/jwks, *, *
`

	sa := stringadapter.NewAdapter(policyText)
	e, err := casbin.NewEnforcer(m, sa)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	tests := []struct {
		name     string
		subOwner string
		subName  string
		method   string
		urlPath  string
		objOwner string
		objName  string
		expected bool
	}{
		{
			name:     "Anonymous user accessing well-known openid-configuration without application param",
			subOwner: "anonymous",
			subName:  "anonymous",
			method:   "GET",
			urlPath:  "/.well-known/openid-configuration",
			objOwner: "",
			objName:  "",
			expected: true,
		},
		{
			name:     "Anonymous user accessing well-known openid-configuration with application param",
			subOwner: "anonymous",
			subName:  "anonymous",
			method:   "GET",
			urlPath:  "/.well-known/my-app/openid-configuration",
			objOwner: "",
			objName:  "",
			expected: true,
		},
		{
			name:     "Anonymous user accessing well-known webfinger with application param",
			subOwner: "anonymous",
			subName:  "anonymous",
			method:   "GET",
			urlPath:  "/.well-known/my-app/webfinger",
			objOwner: "",
			objName:  "",
			expected: true,
		},
		{
			name:     "Anonymous user accessing well-known jwks with application param",
			subOwner: "anonymous",
			subName:  "anonymous",
			method:   "GET",
			urlPath:  "/.well-known/my-app/jwks",
			objOwner: "",
			objName:  "",
			expected: true,
		},
		{
			name:     "Anonymous user accessing well-known jwks with application param (POST method)",
			subOwner: "anonymous",
			subName:  "anonymous",
			method:   "POST",
			urlPath:  "/.well-known/my-app/jwks",
			objOwner: "",
			objName:  "",
			expected: true,
		},
		{
			name:     "Anonymous user accessing non-whitelisted endpoint",
			subOwner: "anonymous",
			subName:  "anonymous",
			method:   "GET",
			urlPath:  "/api/add-user",
			objOwner: "",
			objName:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.Enforce(tt.subOwner, tt.subName, tt.method, tt.urlPath, tt.objOwner, tt.objName)
			if err != nil {
				t.Errorf("Enforce() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Enforce() = %v, expected %v for URL path: %s", result, tt.expected, tt.urlPath)
			}
		})
	}
}
