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
	"time"

	"github.com/casdoor/casdoor/util"
)

// JumpCloudSyncerProvider implements SyncerProvider for JumpCloud API-based syncers
type JumpCloudSyncerProvider struct {
	Syncer *Syncer
}

// InitAdapter initializes the JumpCloud syncer (no database adapter needed)
func (p *JumpCloudSyncerProvider) InitAdapter() error {
	// JumpCloud syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from JumpCloud API
func (p *JumpCloudSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getJumpCloudOriginalUsers()
}

// AddUser adds a new user to JumpCloud (not supported for read-only API)
func (p *JumpCloudSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// JumpCloud syncer is typically read-only
	return false, fmt.Errorf("adding users to JumpCloud is not supported")
}

// UpdateUser updates an existing user in JumpCloud (not supported for read-only API)
func (p *JumpCloudSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// JumpCloud syncer is typically read-only
	return false, fmt.Errorf("updating users in JumpCloud is not supported")
}

// TestConnection tests the JumpCloud API connection
func (p *JumpCloudSyncerProvider) TestConnection() error {
	// Try to get users with limit 1 to test connection
	_, err := p.getJumpCloudUsers(1, 0)
	return err
}

// Close closes any open connections (no-op for JumpCloud API-based syncer)
func (p *JumpCloudSyncerProvider) Close() error {
	// JumpCloud syncer doesn't maintain persistent connections
	return nil
}

// JumpCloudUser represents a user object from JumpCloud API
type JumpCloudUser struct {
	ID           string `json:"_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Firstname    string `json:"firstname"`
	Lastname     string `json:"lastname"`
	DisplayName  string `json:"displayname"`
	JobTitle     string `json:"jobTitle"`
	Department   string `json:"department"`
	Company      string `json:"company"`
	Location     string `json:"location"`
	MiddleName   string `json:"middlename"`
	Activated    bool   `json:"activated"`
	Suspended    bool   `json:"suspended"`
	MfaEnabled   bool   `json:"mfa"`
	PhoneNumbers []struct {
		Type   string `json:"type"`
		Number string `json:"number"`
	} `json:"phoneNumbers"`
	Addresses []struct {
		Type            string `json:"type"`
		StreetAddress   string `json:"streetAddress"`
		Locality        string `json:"locality"`
		Region          string `json:"region"`
		PostalCode      string `json:"postalCode"`
		Country         string `json:"country"`
		ExtendedAddress string `json:"extendedAddress"`
	} `json:"addresses"`
}

// getJumpCloudUsers retrieves users from JumpCloud API with pagination
func (p *JumpCloudSyncerProvider) getJumpCloudUsers(limit, skip int) ([]*JumpCloudUser, error) {
	// syncer.User should be the API key (X-API-KEY)
	// syncer.Host can be optionally used for custom JumpCloud console URL
	// Default to console.jumpcloud.com if not specified

	apiKey := p.Syncer.User
	if apiKey == "" {
		return nil, fmt.Errorf("API key (user field) is required for JumpCloud syncer")
	}

	// JumpCloud API v1 endpoint for system users
	apiUrl := "https://console.jumpcloud.com/api/systemusers"
	if p.Syncer.Host != "" {
		apiUrl = fmt.Sprintf("https://%s/api/systemusers", p.Syncer.Host)
	}

	// Add pagination parameters
	apiUrl = fmt.Sprintf("%s?limit=%d&skip=%d", apiUrl, limit, skip)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	// Set required headers
	req.Header.Set("x-api-key", apiKey)
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

	var users []*JumpCloudUser
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// jumpCloudUserToOriginalUser converts JumpCloud user to Casdoor OriginalUser
func (p *JumpCloudSyncerProvider) jumpCloudUserToOriginalUser(jcUser *JumpCloudUser) *OriginalUser {
	user := &OriginalUser{
		Id:          jcUser.ID,
		Name:        jcUser.Username,
		DisplayName: jcUser.DisplayName,
		FirstName:   jcUser.Firstname,
		LastName:    jcUser.Lastname,
		Email:       jcUser.Email,
		Title:       jcUser.JobTitle,
		Location:    jcUser.Location,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Set phone number if available
	if len(jcUser.PhoneNumbers) > 0 {
		user.Phone = jcUser.PhoneNumbers[0].Number
	}

	// Set IsForbidden based on activated and suspended status
	// User is forbidden if not activated or if suspended
	user.IsForbidden = !jcUser.Activated || jcUser.Suspended

	// If display name is empty, construct from first and last name
	if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
		user.DisplayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}

	// Store additional properties
	if jcUser.Department != "" {
		user.Properties["department"] = jcUser.Department
	}
	if jcUser.Company != "" {
		user.Properties["company"] = jcUser.Company
	}
	if jcUser.MiddleName != "" {
		user.Properties["middleName"] = jcUser.MiddleName
	}

	// Store address information
	if len(jcUser.Addresses) > 0 {
		addr := jcUser.Addresses[0]
		if addr.StreetAddress != "" {
			user.Address = append(user.Address, addr.StreetAddress)
		}
		if addr.ExtendedAddress != "" {
			user.Address = append(user.Address, addr.ExtendedAddress)
		}
		if addr.Locality != "" {
			user.Address = append(user.Address, addr.Locality)
		}
		if addr.Region != "" {
			user.Address = append(user.Address, addr.Region)
		}
		if addr.PostalCode != "" {
			user.Address = append(user.Address, addr.PostalCode)
		}
		if addr.Country != "" {
			user.Address = append(user.Address, addr.Country)
		}
	}

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}

// getJumpCloudOriginalUsers is the main entry point for JumpCloud syncer
func (p *JumpCloudSyncerProvider) getJumpCloudOriginalUsers() ([]*OriginalUser, error) {
	// JumpCloud API returns paginated results
	// We'll fetch all users by iterating through pages
	allUsers := []*JumpCloudUser{}
	limit := 100 // JumpCloud API allows up to 100 results per page
	skip := 0

	for {
		users, err := p.getJumpCloudUsers(limit, skip)
		if err != nil {
			return nil, err
		}

		if len(users) == 0 {
			// No more users to fetch
			break
		}

		allUsers = append(allUsers, users...)

		// If we got fewer results than the limit, we've reached the end
		if len(users) < limit {
			break
		}

		// Move to next page
		skip += limit
	}

	// Convert JumpCloud users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, jcUser := range allUsers {
		originalUser := p.jumpCloudUserToOriginalUser(jcUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}
