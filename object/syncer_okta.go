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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/casdoor/casdoor/util"
)

// OktaSyncerProvider implements SyncerProvider for Okta API-based syncers
type OktaSyncerProvider struct {
	Syncer *Syncer
}

// InitAdapter initializes the Okta syncer (no database adapter needed)
func (p *OktaSyncerProvider) InitAdapter() error {
	// Okta syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from Okta API
func (p *OktaSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getOktaUsers()
}

// AddUser adds a new user to Okta (not supported for read-only API)
func (p *OktaSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// Okta syncer is typically read-only
	return false, fmt.Errorf("adding users to Okta is not supported")
}

// UpdateUser updates an existing user in Okta (not supported for read-only API)
func (p *OktaSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// Okta syncer is typically read-only
	return false, fmt.Errorf("updating users in Okta is not supported")
}

// TestConnection tests the Okta API connection
func (p *OktaSyncerProvider) TestConnection() error {
	// Test connection by trying to fetch a single user
	_, err := p.fetchOktaUsers(1)
	return err
}

// Close closes any open connections (no-op for Okta API-based syncer)
func (p *OktaSyncerProvider) Close() error {
	// Okta syncer doesn't maintain persistent connections
	return nil
}

type OktaUser struct {
	Id              string              `json:"id"`
	Status          string              `json:"status"`
	Created         string              `json:"created"`
	Activated       string              `json:"activated"`
	StatusChanged   string              `json:"statusChanged"`
	LastLogin       string              `json:"lastLogin"`
	LastUpdated     string              `json:"lastUpdated"`
	PasswordChanged string              `json:"passwordChanged"`
	Profile         OktaUserProfile     `json:"profile"`
	Credentials     *OktaUserCredential `json:"credentials,omitempty"`
}

type OktaUserProfile struct {
	Login             string `json:"login"`
	Email             string `json:"email"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	DisplayName       string `json:"displayName"`
	MiddleName        string `json:"middleName"`
	NickName          string `json:"nickName"`
	ProfileUrl        string `json:"profileUrl"`
	SecondEmail       string `json:"secondEmail"`
	MobilePhone       string `json:"mobilePhone"`
	PrimaryPhone      string `json:"primaryPhone"`
	StreetAddress     string `json:"streetAddress"`
	City              string `json:"city"`
	State             string `json:"state"`
	ZipCode           string `json:"zipCode"`
	CountryCode       string `json:"countryCode"`
	PostalAddress     string `json:"postalAddress"`
	PreferredLanguage string `json:"preferredLanguage"`
	Locale            string `json:"locale"`
	Timezone          string `json:"timezone"`
	UserType          string `json:"userType"`
	EmployeeNumber    string `json:"employeeNumber"`
	CostCenter        string `json:"costCenter"`
	Organization      string `json:"organization"`
	Division          string `json:"division"`
	Department        string `json:"department"`
	ManagerId         string `json:"managerId"`
	Manager           string `json:"manager"`
	Title             string `json:"title"`
}

type OktaUserCredential struct {
	Password *OktaUserPassword `json:"password,omitempty"`
}

type OktaUserPassword struct {
	Value string `json:"value,omitempty"`
}

// fetchOktaUsers fetches users from Okta API with pagination
// limit: number of users to fetch (0 for all users)
func (p *OktaSyncerProvider) fetchOktaUsers(limit int) ([]*OktaUser, error) {
	// syncer.Host should be the Okta domain (e.g., "dev-123456.okta.com")
	// syncer.Password should be the API token
	oktaDomain := p.Syncer.Host
	if oktaDomain == "" {
		return nil, fmt.Errorf("Okta domain (host field) is required for Okta syncer")
	}

	apiToken := p.Syncer.Password
	if apiToken == "" {
		return nil, fmt.Errorf("API token (password field) is required for Okta syncer")
	}

	allUsers := []*OktaUser{}
	pageLimit := 200
	if limit > 0 && limit < pageLimit {
		pageLimit = limit
	}

	nextUrl := fmt.Sprintf("https://%s/api/v1/users?limit=%d", oktaDomain, pageLimit)

	for nextUrl != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", nextUrl, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "SSWS "+apiToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to get users: status=%d, body=%s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var users []*OktaUser
		err = json.Unmarshal(body, &users)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, users...)

		// Check if we've reached the limit
		if limit > 0 && len(allUsers) >= limit {
			break
		}

		// Check for pagination link in response headers
		linkHeader := resp.Header.Get("Link")
		nextUrl = ""
		if linkHeader != "" {
			nextUrl = p.parseNextLink(linkHeader)
		}
	}

	return allUsers, nil
}

// parseNextLink extracts the next page URL from the Link header
func (p *OktaSyncerProvider) parseNextLink(linkHeader string) string {
	// Link header format: <https://...>; rel="next", <https://...>; rel="self"
	// We want to extract the URL with rel="next"
	links := parseLinkHeader(linkHeader)
	if nextLink, ok := links["next"]; ok {
		return nextLink
	}
	return ""
}

// parseLinkHeader parses the HTTP Link header and returns a map of rel to URL
func parseLinkHeader(linkHeader string) map[string]string {
	links := make(map[string]string)
	
	// Simple parsing: split by comma, then extract URL and rel
	parts := splitByCommaOutsideAngleBrackets(linkHeader)
	
	for _, part := range parts {
		// Extract URL between < and >
		urlStart := -1
		urlEnd := -1
		for i, ch := range part {
			if ch == '<' {
				urlStart = i + 1
			} else if ch == '>' {
				urlEnd = i
				break
			}
		}
		
		if urlStart == -1 || urlEnd == -1 {
			continue
		}
		
		urlStr := part[urlStart:urlEnd]
		
		// Extract rel value
		relStart := -1
		for i := urlEnd; i < len(part)-4; i++ {
			if part[i:i+5] == "rel=\"" {
				relStart = i + 5
				break
			}
		}
		
		if relStart == -1 {
			continue
		}
		
		relEnd := -1
		for i := relStart; i < len(part); i++ {
			if part[i] == '"' {
				relEnd = i
				break
			}
		}
		
		if relEnd == -1 {
			continue
		}
		
		rel := part[relStart:relEnd]
		links[rel] = urlStr
	}
	
	return links
}

// splitByCommaOutsideAngleBrackets splits a string by commas that are not inside angle brackets
func splitByCommaOutsideAngleBrackets(s string) []string {
	var parts []string
	var current string
	inBrackets := false
	
	for _, ch := range s {
		if ch == '<' {
			inBrackets = true
			current += string(ch)
		} else if ch == '>' {
			inBrackets = false
			current += string(ch)
		} else if ch == ',' && !inBrackets {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}

// getOktaUsers gets all users from Okta API
func (p *OktaSyncerProvider) getOktaUsers() ([]*OriginalUser, error) {
	// Get all users from Okta
	oktaUsers, err := p.fetchOktaUsers(0)
	if err != nil {
		return nil, err
	}

	// Convert Okta users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, oktaUser := range oktaUsers {
		originalUser := p.oktaUserToOriginalUser(oktaUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// oktaUserToOriginalUser converts Okta user to Casdoor OriginalUser
func (p *OktaSyncerProvider) oktaUserToOriginalUser(oktaUser *OktaUser) *OriginalUser {
	profile := oktaUser.Profile

	// Build display name
	displayName := profile.DisplayName
	if displayName == "" && (profile.FirstName != "" || profile.LastName != "") {
		displayName = fmt.Sprintf("%s %s", profile.FirstName, profile.LastName)
	}
	if displayName == "" {
		displayName = profile.Login
	}

	// Build address from available fields
	address := []string{}
	if profile.StreetAddress != "" {
		address = append(address, profile.StreetAddress)
	}
	if profile.City != "" {
		address = append(address, profile.City)
	}
	if profile.State != "" {
		address = append(address, profile.State)
	}
	if profile.ZipCode != "" {
		address = append(address, profile.ZipCode)
	}
	if profile.CountryCode != "" {
		address = append(address, profile.CountryCode)
	}

	user := &OriginalUser{
		Id:          oktaUser.Id,
		Name:        profile.Login,
		DisplayName: displayName,
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		Email:       profile.Email,
		Phone:       profile.MobilePhone,
		Title:       profile.Title,
		Location:    profile.City,
		Language:    profile.PreferredLanguage,
		Address:     address,
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Add additional properties
	if profile.Department != "" {
		user.Properties["department"] = profile.Department
	}
	if profile.Organization != "" {
		user.Properties["organization"] = profile.Organization
	}
	if profile.Division != "" {
		user.Properties["division"] = profile.Division
	}
	if profile.EmployeeNumber != "" {
		user.Properties["employeeNumber"] = profile.EmployeeNumber
	}
	if profile.Manager != "" {
		user.Properties["manager"] = profile.Manager
	}
	if profile.UserType != "" {
		user.Properties["userType"] = profile.UserType
	}

	// Set IsForbidden based on status
	// Okta statuses: STAGED, PROVISIONED, ACTIVE, RECOVERY, PASSWORD_EXPIRED, LOCKED_OUT, SUSPENDED, DEPROVISIONED
	switch oktaUser.Status {
	case "ACTIVE", "PROVISIONED", "RECOVERY", "PASSWORD_EXPIRED":
		user.IsForbidden = false
	default:
		user.IsForbidden = true
	}

	// Parse created time from Okta format (ISO 8601)
	if oktaUser.Created != "" {
		// Okta returns time in RFC3339 format, convert to Casdoor format
		t, err := time.Parse(time.RFC3339, oktaUser.Created)
		if err == nil {
			user.CreatedTime = t.Format("2006-01-02T15:04:05-07:00")
		}
	}

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}
