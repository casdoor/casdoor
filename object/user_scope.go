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
	"github.com/casdoor/casdoor/util"
)

// ConsentRecord represents the data for OAuth consent API requests/responses
type ConsentRecord struct {
	// owner/name
	Application   string   `json:"application"`
	GrantedScopes []string `json:"grantedScopes"`
	DeniedScopes  []string `json:"deniedScopes,omitempty"`
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

func getCustomScopeSet(application *Application) map[string]bool {
	res := map[string]bool{}
	for _, customScope := range application.CustomScopes {
		if customScope != nil && customScope.Scope != "" {
			res[customScope.Scope] = true
		}
	}
	return res
}

func getConsentRecord(userObj *User, application *Application) *ConsentRecord {
	appId := application.GetId()
	for i := range userObj.ApplicationScopes {
		if userObj.ApplicationScopes[i].Application == appId {
			return &userObj.ApplicationScopes[i]
		}
	}
	return nil
}

// CheckConsentRequired checks if user consent is required for the OAuth flow
func CheckConsentRequired(userObj *User, application *Application, scopeStr string) (bool, error) {
	// Skip consent when no custom scopes are configured
	if len(application.CustomScopes) == 0 {
		return false, nil
	}

	// Once policy: consent is asked again only for requested custom scopes the user has neither granted nor denied
	customScopeSet := getCustomScopeSet(application)
	validRequestedScopes := []string{}
	for _, scope := range parseScopes(scopeStr) {
		if customScopeSet[scope] {
			validRequestedScopes = append(validRequestedScopes, scope)
		}
	}

	// If no valid requested scopes, no consent required
	if len(validRequestedScopes) == 0 {
		return false, nil
	}

	record := getConsentRecord(userObj, application)
	if record == nil {
		return true, nil
	}

	decided := map[string]bool{}
	for _, scope := range record.GrantedScopes {
		decided[scope] = true
	}
	for _, scope := range record.DeniedScopes {
		decided[scope] = true
	}

	for _, scope := range validRequestedScopes {
		if !decided[scope] {
			return true, nil
		}
	}
	return false, nil
}

// GetConsentGrantedScopes returns the granted scopes submitted on the consent page, or false if they are not a subset of the requested ones
func GetConsentGrantedScopes(scopeStr string, grantedScopes []string) ([]string, bool) {
	requestedSet := map[string]bool{}
	for _, scope := range parseScopes(scopeStr) {
		requestedSet[scope] = true
	}

	res := []string{}
	seen := map[string]bool{}
	for _, scope := range grantedScopes {
		if !requestedSet[scope] {
			return nil, false
		}
		if !seen[scope] {
			seen[scope] = true
			res = append(res, scope)
		}
	}

	if requestedSet["openid"] && !seen["openid"] {
		res = append([]string{"openid"}, res...)
	}
	return res, true
}

// UpdateConsentRecord records which requested scopes the user granted and which custom scopes the user denied
func UpdateConsentRecord(userObj *User, application *Application, scopeStr string, grantedScopes []string) {
	requestedSet := map[string]bool{}
	for _, scope := range parseScopes(scopeStr) {
		requestedSet[scope] = true
	}
	grantedSet := map[string]bool{}
	for _, scope := range grantedScopes {
		grantedSet[scope] = true
	}

	deniedScopes := []string{}
	customScopeSet := getCustomScopeSet(application)
	for _, scope := range parseScopes(scopeStr) {
		if customScopeSet[scope] && !grantedSet[scope] && !util.InSlice(deniedScopes, scope) {
			deniedScopes = append(deniedScopes, scope)
		}
	}

	record := getConsentRecord(userObj, application)
	if record == nil {
		userObj.ApplicationScopes = append(userObj.ApplicationScopes, ConsentRecord{Application: application.GetId()})
		record = &userObj.ApplicationScopes[len(userObj.ApplicationScopes)-1]
	}

	newGranted := []string{}
	for _, scope := range record.GrantedScopes {
		if !requestedSet[scope] || grantedSet[scope] {
			newGranted = append(newGranted, scope)
		}
	}
	for _, scope := range grantedScopes {
		if !util.InSlice(newGranted, scope) {
			newGranted = append(newGranted, scope)
		}
	}

	newDenied := []string{}
	for _, scope := range record.DeniedScopes {
		if !requestedSet[scope] {
			newDenied = append(newDenied, scope)
		}
	}
	newDenied = append(newDenied, deniedScopes...)

	record.GrantedScopes = newGranted
	record.DeniedScopes = newDenied
}

// GetConsentedScope removes the custom scopes the user has denied from the requested scope
func GetConsentedScope(userObj *User, application *Application, scopeStr string) string {
	record := getConsentRecord(userObj, application)
	if record == nil || len(record.DeniedScopes) == 0 {
		return scopeStr
	}

	customScopeSet := getCustomScopeSet(application)
	deniedSet := map[string]bool{}
	for _, scope := range record.DeniedScopes {
		if customScopeSet[scope] {
			deniedSet[scope] = true
		}
	}

	scopes := parseScopes(scopeStr)
	res := []string{}
	for _, scope := range scopes {
		if !deniedSet[scope] {
			res = append(res, scope)
		}
	}
	if len(res) == len(scopes) {
		return scopeStr
	}
	return strings.Join(res, " ")
}

func validateCustomScopes(customScopes []*ScopeDescription, lang string) error {
	for _, scope := range customScopes {
		if scope == nil || strings.TrimSpace(scope.Scope) == "" {
			return fmt.Errorf("%s: custom scope name", i18n.Translate(lang, "general:Missing parameter"))
		}
	}
	return nil
}
