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
	"net/url"
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
	_, err := p.getSCIMUsers()
	return err
}

// Close closes any open connections (no-op for SCIM API-based syncer)
func (p *SCIMSyncerProvider) Close() error {
	// SCIM syncer doesn't maintain persistent connections
	return nil
}

// SCIMUser represents a user in SCIM 2.0 format
type SCIMUser struct {
	ID       string `json:"id"`
	UserName string `json:"userName"`
	Name     struct {
		FamilyName string `json:"familyName"`
		GivenName  string `json:"givenName"`
		Formatted  string `json:"formatted"`
	} `json:"name"`
	DisplayName string `json:"displayName"`
	NickName    string `json:"nickName"`
	Active      bool   `json:"active"`
	Emails      []struct {
		Value   string `json:"value"`
		Type    string `json:"type"`
		Primary bool   `json:"primary"`
	} `json:"emails"`
	PhoneNumbers []struct {
		Value   string `json:"value"`
		Type    string `json:"type"`
		Primary bool   `json:"primary"`
	} `json:"phoneNumbers"`
	Addresses []struct {
		Formatted     string `json:"formatted"`
		StreetAddress string `json:"streetAddress"`
		Locality      string `json:"locality"`
		Region        string `json:"region"`
		PostalCode    string `json:"postalCode"`
		Country       string `json:"country"`
		Type          string `json:"type"`
		Primary       bool   `json:"primary"`
	} `json:"addresses"`
	Title        string `json:"title"`
	PreferredLanguage string `json:"preferredLanguage"`
	Locale       string `json:"locale"`
	Timezone     string `json:"timezone"`
	Photos       []struct {
		Value   string `json:"value"`
		Type    string `json:"type"`
		Primary bool   `json:"primary"`
	} `json:"photos"`
}

// SCIMListResponse represents the SCIM 2.0 ListResponse format
type SCIMListResponse struct {
	Schemas      []string    `json:"schemas"`
	TotalResults int         `json:"totalResults"`
	StartIndex   int         `json:"startIndex"`
	ItemsPerPage int         `json:"itemsPerPage"`
	Resources    []*SCIMUser `json:"Resources"`
}

// getSCIMUsers gets all users from SCIM API with pagination support
func (p *SCIMSyncerProvider) getSCIMUsers() ([]*OriginalUser, error) {
	// syncer.Host should be the SCIM endpoint URL (e.g., "https://example.com/scim/v2")
	// syncer.User should be the authentication username or client ID (if using basic auth or OAuth)
	// syncer.Password should be the password or access token

	host := p.Syncer.Host
	if host == "" {
		return nil, fmt.Errorf("SCIM endpoint URL (host field) is required for SCIM syncer")
	}

	// Ensure the host doesn't end with a slash
	host = strings.TrimSuffix(host, "/")

	allUsers := []*SCIMUser{}
	startIndex := 1
	count := 100 // SCIM standard pagination count

	for {
		// Build the Users endpoint URL with pagination
		usersURL := fmt.Sprintf("%s/Users?startIndex=%d&count=%d", host, startIndex, count)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", usersURL, nil)
		if err != nil {
			return nil, err
		}

		// Set authorization header
		if p.Syncer.User != "" && p.Syncer.Password != "" {
			// Basic authentication
			req.SetBasicAuth(p.Syncer.User, p.Syncer.Password)
		} else if p.Syncer.Password != "" {
			// Bearer token authentication
			req.Header.Set("Authorization", "Bearer "+p.Syncer.Password)
		}

		req.Header.Set("Content-Type", "application/scim+json")
		req.Header.Set("Accept", "application/scim+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to get users from SCIM: status=%d, body=%s", resp.StatusCode, string(body))
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

		// Check if we need to fetch more pages
		if len(allUsers) >= listResp.TotalResults {
			break
		}

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
		Name:        scimUser.UserName,
		DisplayName: scimUser.DisplayName,
		FirstName:   scimUser.Name.GivenName,
		LastName:    scimUser.Name.FamilyName,
		Title:       scimUser.Title,
		Language:    scimUser.PreferredLanguage,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Get primary email or first email
	if len(scimUser.Emails) > 0 {
		for _, email := range scimUser.Emails {
			if email.Primary {
				user.Email = email.Value
				break
			}
		}
		if user.Email == "" {
			user.Email = scimUser.Emails[0].Value
		}
	}

	// Get primary phone or first phone
	if len(scimUser.PhoneNumbers) > 0 {
		for _, phone := range scimUser.PhoneNumbers {
			if phone.Primary {
				user.Phone = phone.Value
				break
			}
		}
		if user.Phone == "" {
			user.Phone = scimUser.PhoneNumbers[0].Value
		}
	}

	// Get primary address or first address
	if len(scimUser.Addresses) > 0 {
		for _, addr := range scimUser.Addresses {
			if addr.Primary {
				user.Location = addr.Formatted
				if user.Location == "" {
					user.Location = fmt.Sprintf("%s, %s, %s %s", addr.StreetAddress, addr.Locality, addr.Region, addr.PostalCode)
				}
				break
			}
		}
		if user.Location == "" {
			user.Location = scimUser.Addresses[0].Formatted
		}
	}

	// Get primary photo or first photo
	if len(scimUser.Photos) > 0 {
		for _, photo := range scimUser.Photos {
			if photo.Primary {
				user.Avatar = photo.Value
				break
			}
		}
		if user.Avatar == "" {
			user.Avatar = scimUser.Photos[0].Value
		}
	}

	// Set IsForbidden based on active status
	user.IsForbidden = !scimUser.Active

	// If display name is empty, construct from first and last name or use formatted name
	if user.DisplayName == "" {
		if scimUser.Name.Formatted != "" {
			user.DisplayName = scimUser.Name.Formatted
		} else if user.FirstName != "" || user.LastName != "" {
			user.DisplayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		} else if scimUser.NickName != "" {
			user.DisplayName = scimUser.NickName
		} else {
			user.DisplayName = user.Name
		}
	}

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}
