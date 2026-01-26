// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"strings"
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
	return p.getOktaOriginalUsers()
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
	// Try to fetch first page of users to verify connection
	_, _, err := p.getOktaUsers("")
	return err
}

// Close closes any open connections (no-op for Okta API-based syncer)
func (p *OktaSyncerProvider) Close() error {
	// Okta syncer doesn't maintain persistent connections
	return nil
}

// OktaUser represents a user from Okta API
type OktaUser struct {
	Id      string `json:"id"`
	Status  string `json:"status"`
	Created string `json:"created"`
	Profile struct {
		Login             string `json:"login"`
		Email             string `json:"email"`
		FirstName         string `json:"firstName"`
		LastName          string `json:"lastName"`
		DisplayName       string `json:"displayName"`
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
		Title             string `json:"title"`
		Department        string `json:"department"`
		Organization      string `json:"organization"`
	} `json:"profile"`
}

// parseLinkHeader parses the HTTP Link header
// Format: <https://example.com/api/v1/users?after=xyz>; rel="next"
func parseLinkHeader(header string) map[string]string {
	links := make(map[string]string)
	parts := strings.Split(header, ",")
	for _, part := range parts {
		section := strings.Split(strings.TrimSpace(part), ";")
		if len(section) != 2 {
			continue
		}

		url := strings.Trim(strings.TrimSpace(section[0]), "<>")
		rel := strings.TrimSpace(section[1])

		if strings.HasPrefix(rel, "rel=\"") && strings.HasSuffix(rel, "\"") {
			relValue := rel[5 : len(rel)-1]
			links[relValue] = url
		}
	}
	return links
}

// getOktaUsers retrieves users from Okta API with pagination support
// Returns users and the next page link (if any)
func (p *OktaSyncerProvider) getOktaUsers(nextLink string) ([]*OktaUser, string, error) {
	// syncer.Host should be the Okta domain (e.g., "dev-12345.okta.com" or full URL)
	// syncer.Password should be the API token

	domain := p.Syncer.Host
	if domain == "" {
		return nil, "", fmt.Errorf("Okta domain (host field) is required for Okta syncer")
	}

	apiToken := p.Syncer.Password
	if apiToken == "" {
		return nil, "", fmt.Errorf("API token (password field) is required for Okta syncer")
	}

	// Construct API URL
	var apiUrl string
	if nextLink != "" {
		apiUrl = nextLink
	} else {
		// Remove https:// prefix if present in domain
		if len(domain) > 8 && domain[:8] == "https://" {
			domain = domain[8:]
		} else if len(domain) > 7 && domain[:7] == "http://" {
			domain = domain[7:]
		}
		apiUrl = fmt.Sprintf("https://%s/api/v1/users?limit=200", domain)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Authorization", "SSWS "+apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("failed to get users from Okta: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var users []*OktaUser
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, "", err
	}

	// Parse Link header for next page
	// Link header format: <https://...>; rel="next"
	nextPageLink := ""
	linkHeader := resp.Header.Get("Link")
	if linkHeader != "" {
		links := parseLinkHeader(linkHeader)
		if next, ok := links["next"]; ok {
			nextPageLink = next
		}
	}

	return users, nextPageLink, nil
}

// oktaUserToOriginalUser converts Okta user to Casdoor OriginalUser
func (p *OktaSyncerProvider) oktaUserToOriginalUser(oktaUser *OktaUser) *OriginalUser {
	user := &OriginalUser{
		Id:          oktaUser.Id,
		Name:        oktaUser.Profile.Login,
		DisplayName: oktaUser.Profile.DisplayName,
		FirstName:   oktaUser.Profile.FirstName,
		LastName:    oktaUser.Profile.LastName,
		Email:       oktaUser.Profile.Email,
		Phone:       oktaUser.Profile.MobilePhone,
		Title:       oktaUser.Profile.Title,
		Language:    oktaUser.Profile.PreferredLanguage,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Build address from street, city, state, zip
	if oktaUser.Profile.StreetAddress != "" {
		user.Address = append(user.Address, oktaUser.Profile.StreetAddress)
	}
	if oktaUser.Profile.City != "" {
		user.Address = append(user.Address, oktaUser.Profile.City)
	}
	if oktaUser.Profile.State != "" {
		user.Address = append(user.Address, oktaUser.Profile.State)
	}
	if oktaUser.Profile.ZipCode != "" {
		user.Address = append(user.Address, oktaUser.Profile.ZipCode)
	}

	// Store additional properties
	if oktaUser.Profile.Department != "" {
		user.Properties["department"] = oktaUser.Profile.Department
	}
	if oktaUser.Profile.Organization != "" {
		user.Properties["organization"] = oktaUser.Profile.Organization
	}
	if oktaUser.Profile.Timezone != "" {
		user.Properties["timezone"] = oktaUser.Profile.Timezone
	}

	// Set IsForbidden based on status
	// Okta status values: STAGED, PROVISIONED, ACTIVE, RECOVERY, PASSWORD_EXPIRED, LOCKED_OUT, SUSPENDED, DEPROVISIONED
	if oktaUser.Status == "SUSPENDED" || oktaUser.Status == "DEPROVISIONED" || oktaUser.Status == "LOCKED_OUT" {
		user.IsForbidden = true
	} else {
		user.IsForbidden = false
	}

	// If display name is empty, construct from first and last name
	if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
		user.DisplayName = strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	}

	// If email is empty, use login as email (typically login is an email)
	if user.Email == "" && oktaUser.Profile.Login != "" {
		user.Email = oktaUser.Profile.Login
	}

	// If mobile phone is empty, try primary phone
	if user.Phone == "" && oktaUser.Profile.PrimaryPhone != "" {
		user.Phone = oktaUser.Profile.PrimaryPhone
	}

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}

// getOktaOriginalUsers is the main entry point for Okta syncer
func (p *OktaSyncerProvider) getOktaOriginalUsers() ([]*OriginalUser, error) {
	allUsers := []*OktaUser{}
	nextLink := ""

	// Fetch all users with pagination
	for {
		users, next, err := p.getOktaUsers(nextLink)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, users...)

		// If there's no next link, we've fetched all users
		if next == "" {
			break
		}

		nextLink = next
	}

	// Convert Okta users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, oktaUser := range allUsers {
		originalUser := p.oktaUserToOriginalUser(oktaUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}
