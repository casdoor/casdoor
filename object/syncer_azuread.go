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
	"time"

	"github.com/casdoor/casdoor/util"
)

type AzureAdAccessTokenResp struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type AzureAdUser struct {
	Id                string `json:"id"`
	UserPrincipalName string `json:"userPrincipalName"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	Mail              string `json:"mail"`
	MobilePhone       string `json:"mobilePhone"`
	JobTitle          string `json:"jobTitle"`
	OfficeLocation    string `json:"officeLocation"`
	PreferredLanguage string `json:"preferredLanguage"`
	AccountEnabled    bool   `json:"accountEnabled"`
}

type AzureAdUserListResp struct {
	OdataContext  string         `json:"@odata.context"`
	OdataNextLink string         `json:"@odata.nextLink"`
	Value         []*AzureAdUser `json:"value"`
}

// getAzureAdAccessToken gets access token from Azure AD API using client credentials flow
func (syncer *Syncer) getAzureAdAccessToken() (string, error) {
	// syncer.Host should be the tenant ID or tenant domain
	// syncer.User should be the client ID (application ID)
	// syncer.Password should be the client secret

	tenantId := syncer.Host
	if tenantId == "" {
		return "", fmt.Errorf("tenant ID (host field) is required for Azure AD syncer")
	}

	clientId := syncer.User
	if clientId == "" {
		return "", fmt.Errorf("client ID (user field) is required for Azure AD syncer")
	}

	clientSecret := syncer.Password
	if clientSecret == "" {
		return "", fmt.Errorf("client secret (password field) is required for Azure AD syncer")
	}

	tokenUrl := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantId)

	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("grant_type", "client_credentials")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", tokenUrl, nil)
	if err != nil {
		return "", err
	}

	req.URL.RawQuery = data.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	var tokenResp AzureAdAccessTokenResp
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return "", err
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("failed to get access token: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("access token is empty in response")
	}

	return tokenResp.AccessToken, nil
}

// getAzureAdUsers gets all users from Azure AD using Microsoft Graph API
func (syncer *Syncer) getAzureAdUsers(accessToken string) ([]*AzureAdUser, error) {
	allUsers := []*AzureAdUser{}
	nextLink := "https://graph.microsoft.com/v1.0/users?$top=999"

	for nextLink != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", nextLink, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

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

		var userResp AzureAdUserListResp
		err = json.Unmarshal(body, &userResp)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, userResp.Value...)

		// Handle pagination
		nextLink = userResp.OdataNextLink
	}

	return allUsers, nil
}

// azureAdUserToOriginalUser converts Azure AD user to Casdoor OriginalUser
func (syncer *Syncer) azureAdUserToOriginalUser(azureUser *AzureAdUser) *OriginalUser {
	user := &OriginalUser{
		Id:          azureUser.Id,
		Name:        azureUser.UserPrincipalName,
		DisplayName: azureUser.DisplayName,
		FirstName:   azureUser.GivenName,
		LastName:    azureUser.Surname,
		Email:       azureUser.Mail,
		Phone:       azureUser.MobilePhone,
		Title:       azureUser.JobTitle,
		Location:    azureUser.OfficeLocation,
		Language:    azureUser.PreferredLanguage,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Set IsForbidden based on AccountEnabled
	user.IsForbidden = !azureUser.AccountEnabled

	// If display name is empty, construct from first and last name
	if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
		user.DisplayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}

	// If email is empty, use UserPrincipalName as email
	if user.Email == "" && azureUser.UserPrincipalName != "" {
		user.Email = azureUser.UserPrincipalName
	}

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}

// getAzureAdOriginalUsers is the main entry point for Azure AD syncer
func (syncer *Syncer) getAzureAdOriginalUsers() ([]*OriginalUser, error) {
	// Get access token
	accessToken, err := syncer.getAzureAdAccessToken()
	if err != nil {
		return nil, err
	}

	// Get all users from Azure AD
	azureUsers, err := syncer.getAzureAdUsers(accessToken)
	if err != nil {
		return nil, err
	}

	// Convert Azure AD users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, azureUser := range azureUsers {
		originalUser := syncer.azureAdUserToOriginalUser(azureUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}
