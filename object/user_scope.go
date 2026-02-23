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

	"github.com/casdoor/casdoor/i18n"
)

// ConsentRecord represents the data for OAuth consent API requests/responses
type ConsentRecord struct {
	// owner/name
	Application   string   `json:"application"`
	GrantedScopes []string `json:"grantedScopes"`
}

// ScopeDescription represents a human-readable description of an OAuth scope
type ScopeDescription struct {
	Scope       string `json:"scope"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// parseScopes converts a space-separated scope string to a slice
func parseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return []string{}
	}
	scopes := strings.Split(scopeStr, " ")
	var result []string
	for _, scope := range scopes {
		trimmed := strings.TrimSpace(scope)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// CheckConsentRequired checks if user consent is required for the OAuth flow
func CheckConsentRequired(userObj *User, application *Application, scopeStr string) (bool, error) {
	// Skip consent when no custom scopes are configured
	if len(application.CustomScopes) == 0 {
		return false, nil
	}

	// Once policy: check if consent already granted
	requestedScopes := parseScopes(scopeStr)
	appId := application.GetId()

	// Filter requestedScopes to only include scopes defined in application.CustomScopes
	customScopesMap := make(map[string]bool)
	for _, customScope := range application.CustomScopes {
		if customScope.Scope != "" {
			customScopesMap[customScope.Scope] = true
		}
	}

	validRequestedScopes := []string{}
	for _, scope := range requestedScopes {
		if customScopesMap[scope] {
			validRequestedScopes = append(validRequestedScopes, scope)
		}
	}

	// If no valid requested scopes, no consent required
	if len(validRequestedScopes) == 0 {
		return false, nil
	}

	for _, record := range userObj.ApplicationScopes {
		if record.Application == appId {
			// Check if grantedScopes contains all validRequestedScopes
			grantedMap := make(map[string]bool)
			for _, scope := range record.GrantedScopes {
				grantedMap[scope] = true
			}

			allGranted := true
			for _, scope := range validRequestedScopes {
				if !grantedMap[scope] {
					allGranted = false
					break
				}
			}

			if allGranted {
				// Consent already granted for all valid requested scopes
				return false, nil
			}
		}
	}

	// Consent required
	return true, nil
}

func validateCustomScopes(customScopes []*ScopeDescription, lang string) error {
	for _, scope := range customScopes {
		if scope == nil || strings.TrimSpace(scope.Scope) == "" {
			return fmt.Errorf("%s: custom scope name", i18n.Translate(lang, "general:Missing parameter"))
		}
	}
	return nil
}
