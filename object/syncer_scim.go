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

// SCIMSyncerProvider implements SyncerProvider for SCIM 2.0 API-based syncers
type SCIMSyncerProvider struct {
	Syncer *Syncer
}

// InitAdapter initializes the SCIM syncer (no database adapter needed)
func (p *SCIMSyncerProvider) InitAdapter() error {
	// SCIM syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from SCIM API
func (p *SCIMSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getSCIMUsers()
}

// AddUser adds a new user to SCIM (not supported for read-only API)
func (p *SCIMSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// SCIM syncer is typically read-only
	return false, fmt.Errorf("adding users to SCIM is not supported")
}

// UpdateUser updates an existing user in SCIM (not supported for read-only API)
func (p *SCIMSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// SCIM syncer is typically read-only
	return false, fmt.Errorf("updating users in SCIM is not supported")
}

// TestConnection tests the SCIM API connection
func (p *SCIMSyncerProvider) TestConnection() error {
	// Test by trying to fetch users with a limit of 1
	endpoint := p.buildSCIMEndpoint()
	endpoint = fmt.Sprintf("%s?startIndex=1&count=1", endpoint)

	req, err := p.createSCIMRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SCIM connection test failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// Close closes any open connections (no-op for SCIM API-based syncer)
func (p *SCIMSyncerProvider) Close() error {
	// SCIM syncer doesn't maintain persistent connections
	return nil
}

// SCIMName represents a SCIM user name structure
type SCIMName struct {
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
	Formatted  string `json:"formatted"`
}

// SCIMEmail represents a SCIM user email structure
type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type"`
	Primary bool   `json:"primary"`
}

// SCIMPhoneNumber represents a SCIM user phone number structure
type SCIMPhoneNumber struct {
	Value   string `json:"value"`
	Type    string `json:"type"`
	Primary bool   `json:"primary"`
}

// SCIMAddress represents a SCIM user address structure
type SCIMAddress struct {
	StreetAddress string `json:"streetAddress"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postalCode"`
	Country       string `json:"country"`
	Formatted     string `json:"formatted"`
	Type          string `json:"type"`
	Primary       bool   `json:"primary"`
}

// SCIMUser represents a SCIM 2.0 user resource
type SCIMUser struct {
	ID           string            `json:"id"`
	ExternalID   string            `json:"externalId"`
	UserName     string            `json:"userName"`
	Name         SCIMName          `json:"name"`
	DisplayName  string            `json:"displayName"`
	NickName     string            `json:"nickName"`
	ProfileURL   string            `json:"profileUrl"`
	Title        string            `json:"title"`
	UserType     string            `json:"userType"`
	PreferredLan string            `json:"preferredLanguage"`
	Locale       string            `json:"locale"`
	Timezone     string            `json:"timezone"`
	Active       bool              `json:"active"`
	Emails       []SCIMEmail       `json:"emails"`
	PhoneNumbers []SCIMPhoneNumber `json:"phoneNumbers"`
	Addresses    []SCIMAddress     `json:"addresses"`
}

// SCIMListResponse represents a SCIM list response
type SCIMListResponse struct {
	TotalResults int         `json:"totalResults"`
	ItemsPerPage int         `json:"itemsPerPage"`
	StartIndex   int         `json:"startIndex"`
	Resources    []*SCIMUser `json:"Resources"`
}

// buildSCIMEndpoint builds the SCIM API endpoint URL
func (p *SCIMSyncerProvider) buildSCIMEndpoint() string {
	// syncer.Host should be the SCIM server URL (e.g., https://example.com/scim/v2)
	host := strings.TrimSuffix(p.Syncer.Host, "/")
	return fmt.Sprintf("%s/Users", host)
}

// createSCIMRequest creates an HTTP request with proper authentication
func (p *SCIMSyncerProvider) createSCIMRequest(method, url string, body io.Reader) (*http.Request, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Set SCIM headers
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	// Add authentication
	// syncer.User should be the authentication token or username
	// syncer.Password should be the password or API key
	if p.Syncer.User != "" && p.Syncer.Password != "" {
		// Try Basic Auth
		req.SetBasicAuth(p.Syncer.User, p.Syncer.Password)
	} else if p.Syncer.Password != "" {
		// Try Bearer token (assuming password field contains the token)
		req.Header.Set("Authorization", "Bearer "+p.Syncer.Password)
	} else if p.Syncer.User != "" {
		// Try Bearer token (assuming user field contains the token)
		req.Header.Set("Authorization", "Bearer "+p.Syncer.User)
	}

	return req, nil
}

// getSCIMUsers retrieves all users from SCIM API with pagination
func (p *SCIMSyncerProvider) getSCIMUsers() ([]*OriginalUser, error) {
	allUsers := []*SCIMUser{}
	startIndex := 1
	count := 100 // Fetch 100 users per page

	for {
		endpoint := p.buildSCIMEndpoint()
		endpoint = fmt.Sprintf("%s?startIndex=%d&count=%d", endpoint, startIndex, count)

		req, err := p.createSCIMRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

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

		var listResp SCIMListResponse
		err = json.Unmarshal(body, &listResp)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, listResp.Resources...)

		// Check if we've fetched all users
		if len(allUsers) >= listResp.TotalResults {
			break
		}

		// Move to the next page
		startIndex += count
	}

	// Convert SCIM users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, scimUser := range allUsers {
		originalUser := p.scimUserToOriginalUser(scimUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// scimUserToOriginalUser converts SCIM user to Casdoor OriginalUser
func (p *SCIMSyncerProvider) scimUserToOriginalUser(scimUser *SCIMUser) *OriginalUser {
	user := &OriginalUser{
		Id:          scimUser.ID,
		ExternalId:  scimUser.ExternalID,
		Name:        scimUser.UserName,
		DisplayName: scimUser.DisplayName,
		FirstName:   scimUser.Name.GivenName,
		LastName:    scimUser.Name.FamilyName,
		Title:       scimUser.Title,
		Language:    scimUser.PreferredLan,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// If display name is from name structure
	if user.DisplayName == "" && scimUser.Name.Formatted != "" {
		user.DisplayName = scimUser.Name.Formatted
	}

	// If display name is still empty, construct from first and last name
	if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
		user.DisplayName = strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	}

	// Extract primary email or first email
	if len(scimUser.Emails) > 0 {
		for _, email := range scimUser.Emails {
			if email.Primary {
				user.Email = email.Value
				break
			}
		}
		// If no primary email, use the first one
		if user.Email == "" && len(scimUser.Emails) > 0 {
			user.Email = scimUser.Emails[0].Value
		}
	}

	// Extract primary phone or first phone
	if len(scimUser.PhoneNumbers) > 0 {
		for _, phone := range scimUser.PhoneNumbers {
			if phone.Primary {
				user.Phone = phone.Value
				break
			}
		}
		// If no primary phone, use the first one
		if user.Phone == "" && len(scimUser.PhoneNumbers) > 0 {
			user.Phone = scimUser.PhoneNumbers[0].Value
		}
	}

	// Extract primary address or first address
	if len(scimUser.Addresses) > 0 {
		for _, addr := range scimUser.Addresses {
			if addr.Primary {
				if addr.Formatted != "" {
					user.Address = []string{addr.Formatted}
				} else {
					user.Address = []string{addr.StreetAddress, addr.Locality, addr.Region, addr.PostalCode, addr.Country}
				}
				user.Location = addr.Locality
				user.Region = addr.Region
				break
			}
		}
		// If no primary address, use the first one
		if len(user.Address) == 0 && len(scimUser.Addresses) > 0 {
			addr := scimUser.Addresses[0]
			if addr.Formatted != "" {
				user.Address = []string{addr.Formatted}
			} else {
				user.Address = []string{addr.StreetAddress, addr.Locality, addr.Region, addr.PostalCode, addr.Country}
			}
			user.Location = addr.Locality
			user.Region = addr.Region
		}
	}

	// Set IsForbidden based on Active status
	user.IsForbidden = !scimUser.Active

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}

// GetOriginalGroups retrieves all groups from SCIM (not implemented yet)
func (p *SCIMSyncerProvider) GetOriginalGroups() ([]*OriginalGroup, error) {
	// TODO: Implement SCIM group sync
	return []*OriginalGroup{}, nil
}

// GetOriginalUserGroups retrieves the group IDs that a user belongs to (not implemented yet)
func (p *SCIMSyncerProvider) GetOriginalUserGroups(userId string) ([]string, error) {
	// TODO: Implement SCIM user group membership sync
	return []string{}, nil
}

