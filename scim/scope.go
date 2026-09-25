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

package scim

import (
	"context"
	"net/http"

	"github.com/elimity-com/scim/errors"
)

type organizationKey struct{}

// WithOrganization limits the SCIM request to the users and groups of the organization,
// an empty organization (a global admin) to none. A request without it is denied.
func WithOrganization(r *http.Request, organization string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), organizationKey{}, organization))
}

func getOrganization(r *http.Request) (string, error) {
	organization, ok := r.Context().Value(organizationKey{}).(string)
	if !ok {
		return "", errors.ScimError{Detail: "the SCIM request has no organization", Status: http.StatusForbidden}
	}
	return organization, nil
}

func isOrganizationAllowed(r *http.Request, owner string) bool {
	organization, err := getOrganization(r)
	return err == nil && (organization == "" || organization == owner)
}

func checkOrganization(r *http.Request, owner string) error {
	if !isOrganizationAllowed(r, owner) {
		return errors.ScimError{Detail: "the organization: " + owner + " is not allowed", Status: http.StatusForbidden}
	}
	return nil
}
