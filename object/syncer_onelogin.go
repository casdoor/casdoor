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

// OneLoginSyncerProvider implements SyncerProvider for OneLogin API-based syncers
type OneLoginSyncerProvider struct {
	Syncer *Syncer
}

// InitAdapter initializes the OneLogin syncer (no database adapter needed)
func (p *OneLoginSyncerProvider) InitAdapter() error {
	// OneLogin syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from OneLogin API
func (p *OneLoginSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getOneLoginOriginalUsers()
}

// AddUser adds a new user to OneLogin (not supported for read-only API)
func (p *OneLoginSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// OneLogin syncer is typically read-only
	return false, fmt.Errorf("adding users to OneLogin is not supported")
}

// UpdateUser updates an existing user in OneLogin (not supported for read-only API)
func (p *OneLoginSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// OneLogin syncer is typically read-only
	return false, fmt.Errorf("updating users in OneLogin is not supported")
}

// TestConnection tests the OneLogin API connection
func (p *OneLoginSyncerProvider) TestConnection() error {
	_, err := p.getOneLoginAccessToken()
	return err
}

// Close closes any open connections (no-op for OneLogin API-based syncer)
func (p *OneLoginSyncerProvider) Close() error {
	// OneLogin syncer doesn't maintain persistent connections
	return nil
}

type OneLoginAccessTokenResp struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data []struct {
		AccessToken string `json:"access_token"`
		CreatedAt   string `json:"created_at"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	} `json:"data"`
}

type OneLoginUser struct {
	Id                int    `json:"id"`
	ExternalId        string `json:"external_id"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	Firstname         string `json:"firstname"`
	Lastname          string `json:"lastname"`
	DistinguishedName string `json:"distinguished_name"`
	Phone             string `json:"phone"`
	Company           string `json:"company"`
	Department        string `json:"department"`
	Title             string `json:"title"`
	Status            int    `json:"status"`
	State             int    `json:"state"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	ActivatedAt       string `json:"activated_at"`
	LockedUntil       string `json:"locked_until"`
	LastLogin         string `json:"last_login"`
	InvalidLoginCount int    `json:"invalid_login_attempts"`
	Manager           string `json:"manager"`
	ManagerAdId       string `json:"manager_ad_id"`
	DirectoryId       int    `json:"directory_id"`
}

type OneLoginUserListResp struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data []OneLoginUser `json:"data"`
	Pagination struct {
		BeforeCursor string `json:"before_cursor"`
		AfterCursor  string `json:"after_cursor"`
		PreviousLink string `json:"previous_link"`
		NextLink     string `json:"next_link"`
	} `json:"pagination"`
}

// getOneLoginAccessToken gets access token from OneLogin API using client credentials flow
func (p *OneLoginSyncerProvider) getOneLoginAccessToken() (string, error) {
	// syncer.Host should be the OneLogin region (e.g., "us" or "eu") or full domain
	// syncer.User should be the client ID
	// syncer.Password should be the client secret

	region := p.Syncer.Host
	if region == "" {
		return "", fmt.Errorf("OneLogin region (host field) is required for OneLogin syncer")
	}

	clientId := p.Syncer.User
	if clientId == "" {
		return "", fmt.Errorf("client ID (user field) is required for OneLogin syncer")
	}

	clientSecret := p.Syncer.Password
	if clientSecret == "" {
		return "", fmt.Errorf("client secret (password field) is required for OneLogin syncer")
	}

	// Determine the token URL based on region
	var tokenUrl string
	if strings.Contains(region, ".") {
		// Full domain provided
		tokenUrl = fmt.Sprintf("https://%s/auth/oauth2/v2/token", region)
	} else {
		// Region code provided (e.g., "us", "eu")
		tokenUrl = fmt.Sprintf("https://api.%s.onelogin.com/auth/oauth2/v2/token", region)
	}

	// Prepare request body
	requestBody := map[string]string{
		"grant_type": "client_credentials",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", tokenUrl, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("client_id:%s, client_secret:%s", clientId, clientSecret))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp OneLoginAccessTokenResp
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return "", err
	}

	if tokenResp.Status.Error {
		return "", fmt.Errorf("failed to get access token: %s - %s", tokenResp.Status.Type, tokenResp.Status.Message)
	}

	if len(tokenResp.Data) == 0 || tokenResp.Data[0].AccessToken == "" {
		return "", fmt.Errorf("access token is empty in response")
	}

	return tokenResp.Data[0].AccessToken, nil
}

// getOneLoginUsers gets all users from OneLogin API
func (p *OneLoginSyncerProvider) getOneLoginUsers(accessToken string) ([]OneLoginUser, error) {
	allUsers := []OneLoginUser{}

	// Determine the API URL based on region
	region := p.Syncer.Host
	var apiUrl string
	if strings.Contains(region, ".") {
		// Full domain provided
		apiUrl = fmt.Sprintf("https://%s/api/2/users", region)
	} else {
		// Region code provided (e.g., "us", "eu")
		apiUrl = fmt.Sprintf("https://api.%s.onelogin.com/api/2/users", region)
	}

	// OneLogin API supports pagination, fetch all pages
	nextLink := apiUrl

	for nextLink != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		req, err := http.NewRequestWithContext(ctx, "GET", nextLink, nil)
		if err != nil {
			cancel()
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			cancel()
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			cancel()
			return nil, fmt.Errorf("failed to get users: status=%d, body=%s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()

		if err != nil {
			return nil, err
		}

		var userResp OneLoginUserListResp
		err = json.Unmarshal(body, &userResp)
		if err != nil {
			return nil, err
		}

		if userResp.Status.Error {
			return nil, fmt.Errorf("failed to get users: %s - %s", userResp.Status.Type, userResp.Status.Message)
		}

		allUsers = append(allUsers, userResp.Data...)

		// Handle pagination
		nextLink = userResp.Pagination.NextLink
	}

	return allUsers, nil
}

// oneLoginUserToOriginalUser converts OneLogin user to Casdoor OriginalUser
func (p *OneLoginSyncerProvider) oneLoginUserToOriginalUser(oneLoginUser *OneLoginUser) *OriginalUser {
	user := &OriginalUser{
		Id:          fmt.Sprintf("%d", oneLoginUser.Id),
		Name:        oneLoginUser.Username,
		DisplayName: fmt.Sprintf("%s %s", oneLoginUser.Firstname, oneLoginUser.Lastname),
		FirstName:   oneLoginUser.Firstname,
		LastName:    oneLoginUser.Lastname,
		Email:       oneLoginUser.Email,
		Phone:       oneLoginUser.Phone,
		Title:       oneLoginUser.Title,
		Location:    oneLoginUser.Company,
	}

	// Map department to Affiliation
	if oneLoginUser.Department != "" {
		user.Affiliation = oneLoginUser.Department
	}

	// Set IsForbidden based on status (0 = Unactivated, 1 = Active, 2 = Suspended, 3 = Locked, 4 = Password expired, 5 = Awaiting password reset)
	// Only status 1 (Active) is considered not forbidden
	user.IsForbidden = oneLoginUser.Status != 1

	// If display name construction is empty, use email or username
	if strings.TrimSpace(user.DisplayName) == "" {
		if oneLoginUser.Email != "" {
			user.DisplayName = oneLoginUser.Email
		} else {
			user.DisplayName = oneLoginUser.Username
		}
	}

	// If email is empty, try to use username as email if it contains @
	if user.Email == "" && strings.Contains(oneLoginUser.Username, "@") {
		user.Email = oneLoginUser.Username
	}

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}

// getOneLoginOriginalUsers is the main entry point for OneLogin syncer
func (p *OneLoginSyncerProvider) getOneLoginOriginalUsers() ([]*OriginalUser, error) {
	// Get access token
	accessToken, err := p.getOneLoginAccessToken()
	if err != nil {
		return nil, err
	}

	// Get all users from OneLogin
	oneLoginUsers, err := p.getOneLoginUsers(accessToken)
	if err != nil {
		return nil, err
	}

	// Convert OneLogin users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for i := range oneLoginUsers {
		originalUser := p.oneLoginUserToOriginalUser(&oneLoginUsers[i])
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}
