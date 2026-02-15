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

package object

import (
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// ScopeDescription represents a human-readable description of an OAuth scope
type ScopeDescription struct {
	Scope       string `json:"scope"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// DefaultScopeDescriptions provides descriptions for standard OIDC scopes
var DefaultScopeDescriptions = []ScopeDescription{
	{Scope: "openid", DisplayName: "OpenID", Description: "Verify your identity"},
	{Scope: "profile", DisplayName: "Profile", Description: "View your basic profile information"},
	{Scope: "email", DisplayName: "Email", Description: "View your email address"},
	{Scope: "address", DisplayName: "Address", Description: "View your address"},
	{Scope: "phone", DisplayName: "Phone", Description: "View your phone number"},
	{Scope: "offline_access", DisplayName: "Offline Access", Description: "Maintain access when you are not actively using the application"},
}

// ConsentRecord stores user consent for OAuth applications
type ConsentRecord struct {
	Owner          string   `xorm:"varchar(100) notnull pk" json:"owner"`
	Name           string   `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime    string   `xorm:"varchar(100)" json:"createdTime"`
	User           string   `xorm:"varchar(100) index" json:"user"`
	Application    string   `xorm:"varchar(100) index" json:"application"`
	GrantedScopes  []string `xorm:"varchar(1000)" json:"grantedScopes"`
	ConsentTime    string   `xorm:"varchar(100)" json:"consentTime"`
	ExpirationTime string   `xorm:"varchar(100)" json:"expirationTime"`
}

// GetScopeDescriptions returns descriptions for the given scopes
func GetScopeDescriptions(scopes []string) []ScopeDescription {
	var descriptions []ScopeDescription
	scopeMap := make(map[string]ScopeDescription)

	// Build map from default descriptions
	for _, desc := range DefaultScopeDescriptions {
		scopeMap[desc.Scope] = desc
	}

	// Return descriptions for requested scopes
	for _, scope := range scopes {
		if desc, ok := scopeMap[scope]; ok {
			descriptions = append(descriptions, desc)
		} else {
			// For unknown scopes, create a generic description
			descriptions = append(descriptions, ScopeDescription{
				Scope:       scope,
				DisplayName: scope,
				Description: fmt.Sprintf("Access to %s", scope),
			})
		}
	}

	return descriptions
}

// GetConsentRecordCount returns the count of consent records
func GetConsentRecordCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&ConsentRecord{})
}

// GetConsentRecords returns all consent records for an owner
func GetConsentRecords(owner string) ([]*ConsentRecord, error) {
	consents := []*ConsentRecord{}
	err := ormer.Engine.Desc("created_time").Find(&consents, &ConsentRecord{Owner: owner})
	return consents, err
}

// GetPaginationConsentRecords returns paginated consent records
func GetPaginationConsentRecords(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*ConsentRecord, error) {
	consents := []*ConsentRecord{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&consents)
	return consents, err
}

// GetConsentRecord returns a specific consent record
func GetConsentRecord(id string) (*ConsentRecord, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getConsentRecord(owner, name)
}

func getConsentRecord(owner, name string) (*ConsentRecord, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	consent := ConsentRecord{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&consent)
	if err != nil {
		return nil, err
	}

	if existed {
		return &consent, nil
	}

	return nil, nil
}

// GetUserConsentForApplication returns the consent record for a user and application
func GetUserConsentForApplication(user, application string) (*ConsentRecord, error) {
	consents := []*ConsentRecord{}
	err := ormer.Engine.Where("user = ? AND application = ?", user, application).Desc("created_time").Limit(1).Find(&consents)
	if err != nil {
		return nil, err
	}

	if len(consents) > 0 {
		return consents[0], nil
	}

	return nil, nil
}

// AddConsentRecord adds a new consent record
func AddConsentRecord(consent *ConsentRecord) (bool, error) {
	if consent.Owner == "" {
		return false, fmt.Errorf("owner is required")
	}
	if consent.Name == "" {
		consent.Name = util.GenerateId()
	}
	if consent.CreatedTime == "" {
		consent.CreatedTime = util.GetCurrentTime()
	}
	if consent.ConsentTime == "" {
		consent.ConsentTime = util.GetCurrentTime()
	}

	affected, err := ormer.Engine.Insert(consent)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

// UpdateConsentRecord updates an existing consent record
func UpdateConsentRecord(id string, consent *ConsentRecord) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if c, err := getConsentRecord(owner, name); err != nil {
		return false, err
	} else if c == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(consent)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

// DeleteConsentRecord deletes a consent record
func DeleteConsentRecord(consent *ConsentRecord) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{consent.Owner, consent.Name}).Delete(&ConsentRecord{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

// DeleteUserConsentForApplication deletes consent for a specific user and application
func DeleteUserConsentForApplication(user, application string) (bool, error) {
	affected, err := ormer.Engine.Where("user = ? AND application = ?", user, application).Delete(&ConsentRecord{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

// ContainsAllScopes checks if grantedScopes contains all requestedScopes
func ContainsAllScopes(grantedScopes, requestedScopes []string) bool {
	grantedMap := make(map[string]bool)
	for _, scope := range grantedScopes {
		grantedMap[scope] = true
	}

	for _, scope := range requestedScopes {
		if !grantedMap[scope] {
			return false
		}
	}

	return true
}

// ParseScopes converts a space-separated scope string to a slice
func ParseScopes(scopeStr string) []string {
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

// JoinScopes converts a slice of scopes to a space-separated string
func JoinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}

// checkConsentRequired checks if user consent is required for the OAuth flow
func checkConsentRequired(user *User, application *Application, scopeStr string) (bool, error) {
	// Check consent policy
	consentPolicy := application.ConsentPolicy
	if consentPolicy == "" || consentPolicy == "skip" {
		// No consent required
		return false, nil
	}

	if consentPolicy == "always" {
		// Always require consent
		return true, nil
	}

	// Policy is "once" - check if consent already granted
	requestedScopes := ParseScopes(scopeStr)
	existingConsent, err := GetUserConsentForApplication(user.Name, application.Name)
	if err != nil {
		return false, err
	}

	if existingConsent != nil && ContainsAllScopes(existingConsent.GrantedScopes, requestedScopes) {
		// Consent already granted for all requested scopes
		return false, nil
	}

	// Consent required
	return true, nil
}
