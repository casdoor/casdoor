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

	"github.com/casdoor/casdoor/util"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// GoogleWorkspaceSyncerProvider implements SyncerProvider for Google Workspace API-based syncers
type GoogleWorkspaceSyncerProvider struct {
	Syncer *Syncer
}

// InitAdapter initializes the Google Workspace syncer (no database adapter needed)
func (p *GoogleWorkspaceSyncerProvider) InitAdapter() error {
	// Google Workspace syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from Google Workspace API
func (p *GoogleWorkspaceSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getGoogleWorkspaceOriginalUsers()
}

// AddUser adds a new user to Google Workspace (not supported for read-only API)
func (p *GoogleWorkspaceSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// Google Workspace syncer is typically read-only
	return false, fmt.Errorf("adding users to Google Workspace is not supported")
}

// UpdateUser updates an existing user in Google Workspace (not supported for read-only API)
func (p *GoogleWorkspaceSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// Google Workspace syncer is typically read-only
	return false, fmt.Errorf("updating users in Google Workspace is not supported")
}

// TestConnection tests the Google Workspace API connection
func (p *GoogleWorkspaceSyncerProvider) TestConnection() error {
	_, err := p.getAdminService()
	return err
}

// Close closes any open connections (no-op for Google Workspace API-based syncer)
func (p *GoogleWorkspaceSyncerProvider) Close() error {
	// Google Workspace syncer doesn't maintain persistent connections
	return nil
}

// getAdminService creates and returns a Google Workspace Admin SDK service
func (p *GoogleWorkspaceSyncerProvider) getAdminService() (*admin.Service, error) {
	// syncer.Host should be the admin email (impersonation account)
	// syncer.User should be the service account email or client_email
	// syncer.Password should be the service account private key (JSON key file content)

	adminEmail := p.Syncer.Host
	if adminEmail == "" {
		return nil, fmt.Errorf("admin email (host field) is required for Google Workspace syncer")
	}

	// Parse the service account credentials from the password field
	serviceAccountKey := p.Syncer.Password
	if serviceAccountKey == "" {
		return nil, fmt.Errorf("service account key (password field) is required for Google Workspace syncer")
	}

	// Parse the JSON key
	var serviceAccount struct {
		Type        string `json:"type"`
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}

	err := json.Unmarshal([]byte(serviceAccountKey), &serviceAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account key: %v", err)
	}

	// Create JWT config for service account with domain-wide delegation
	config := &jwt.Config{
		Email:      serviceAccount.ClientEmail,
		PrivateKey: []byte(serviceAccount.PrivateKey),
		Scopes: []string{
			admin.AdminDirectoryUserReadonlyScope,
			admin.AdminDirectoryGroupReadonlyScope,
		},
		TokenURL: google.JWTTokenURL,
		Subject:  adminEmail, // Impersonate the admin user
	}

	client := config.Client(context.Background())

	// Create Admin SDK service
	service, err := admin.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create admin service: %v", err)
	}

	return service, nil
}

// getGoogleWorkspaceUsers gets all users from Google Workspace using Admin SDK API
func (p *GoogleWorkspaceSyncerProvider) getGoogleWorkspaceUsers(service *admin.Service) ([]*admin.User, error) {
	allUsers := []*admin.User{}
	pageToken := ""

	// Get the customer ID (use "my_customer" for the domain)
	customer := "my_customer"

	for {
		call := service.Users.List().Customer(customer).MaxResults(500)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %v", err)
		}

		allUsers = append(allUsers, resp.Users...)

		// Handle pagination
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allUsers, nil
}

// googleWorkspaceUserToOriginalUser converts Google Workspace user to Casdoor OriginalUser
func (p *GoogleWorkspaceSyncerProvider) googleWorkspaceUserToOriginalUser(gwUser *admin.User) *OriginalUser {
	user := &OriginalUser{
		Id:         gwUser.Id,
		Name:       gwUser.PrimaryEmail,
		Email:      gwUser.PrimaryEmail,
		Avatar:     gwUser.ThumbnailPhotoUrl,
		Address:    []string{},
		Properties: map[string]string{},
		Groups:     []string{},
	}

	// Set name fields if Name is not nil
	if gwUser.Name != nil {
		user.DisplayName = gwUser.Name.FullName
		user.FirstName = gwUser.Name.GivenName
		user.LastName = gwUser.Name.FamilyName
	}

	// Set IsForbidden based on account status
	user.IsForbidden = gwUser.Suspended

	// Set IsAdmin
	user.IsAdmin = gwUser.IsAdmin

	// If display name is empty, construct from first and last name
	if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
		user.DisplayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}

	// Set CreatedTime from Google or current time
	if gwUser.CreationTime != "" {
		user.CreatedTime = gwUser.CreationTime
	} else {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}

// getGoogleWorkspaceOriginalUsers is the main entry point for Google Workspace syncer
func (p *GoogleWorkspaceSyncerProvider) getGoogleWorkspaceOriginalUsers() ([]*OriginalUser, error) {
	// Get Admin SDK service
	service, err := p.getAdminService()
	if err != nil {
		return nil, err
	}

	// Get all users from Google Workspace
	gwUsers, err := p.getGoogleWorkspaceUsers(service)
	if err != nil {
		return nil, err
	}

	// Get all groups and their members to build a user-to-groups mapping
	// This avoids N+1 queries by fetching group memberships upfront
	userGroupsMap, err := p.buildUserGroupsMap(service)
	if err != nil {
		fmt.Printf("Warning: failed to fetch group memberships: %v. Users will have no groups assigned.\n", err)
		userGroupsMap = make(map[string][]string)
	}

	// Convert Google Workspace users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, gwUser := range gwUsers {
		originalUser := p.googleWorkspaceUserToOriginalUser(gwUser)

		// Assign groups from the pre-built map
		if groups, exists := userGroupsMap[gwUser.PrimaryEmail]; exists {
			originalUser.Groups = groups
		} else {
			originalUser.Groups = []string{}
		}

		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// buildUserGroupsMap builds a map of user email to group emails by iterating through all groups
// and their members. This is more efficient than querying groups for each user individually.
func (p *GoogleWorkspaceSyncerProvider) buildUserGroupsMap(service *admin.Service) (map[string][]string, error) {
	userGroupsMap := make(map[string][]string)

	// Get all groups
	groups, err := p.getGoogleWorkspaceGroups(service)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch groups: %v", err)
	}

	// For each group, get its members and populate the user-to-groups map
	for _, group := range groups {
		members, err := p.getGroupMembers(service, group.Id)
		if err != nil {
			fmt.Printf("Warning: failed to get members for group %s: %v\n", group.Email, err)
			continue
		}

		// Add this group to each member's group list
		for _, member := range members {
			userGroupsMap[member.Email] = append(userGroupsMap[member.Email], group.Email)
		}
	}

	return userGroupsMap, nil
}

// getGroupMembers retrieves all members of a specific group
func (p *GoogleWorkspaceSyncerProvider) getGroupMembers(service *admin.Service, groupId string) ([]*admin.Member, error) {
	allMembers := []*admin.Member{}
	pageToken := ""

	for {
		call := service.Members.List(groupId).MaxResults(500)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list members: %v", err)
		}

		allMembers = append(allMembers, resp.Members...)

		// Handle pagination
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allMembers, nil
}

// GetOriginalGroups retrieves all groups from Google Workspace
func (p *GoogleWorkspaceSyncerProvider) GetOriginalGroups() ([]*OriginalGroup, error) {
	// Get Admin SDK service
	service, err := p.getAdminService()
	if err != nil {
		return nil, err
	}

	// Get all groups from Google Workspace
	gwGroups, err := p.getGoogleWorkspaceGroups(service)
	if err != nil {
		return nil, err
	}

	// Convert Google Workspace groups to Casdoor OriginalGroup
	originalGroups := []*OriginalGroup{}
	for _, gwGroup := range gwGroups {
		originalGroup := p.googleWorkspaceGroupToOriginalGroup(gwGroup)
		originalGroups = append(originalGroups, originalGroup)
	}

	return originalGroups, nil
}

// GetOriginalUserGroups retrieves the group IDs that a user belongs to
func (p *GoogleWorkspaceSyncerProvider) GetOriginalUserGroups(userId string) ([]string, error) {
	// Get Admin SDK service
	service, err := p.getAdminService()
	if err != nil {
		return nil, err
	}

	// Get groups for the user
	groupIds := []string{}
	pageToken := ""

	for {
		call := service.Groups.List().UserKey(userId)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list user groups: %v", err)
		}

		for _, group := range resp.Groups {
			groupIds = append(groupIds, group.Email)
		}

		// Handle pagination
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return groupIds, nil
}

// getGoogleWorkspaceGroups gets all groups from Google Workspace using Admin SDK API
func (p *GoogleWorkspaceSyncerProvider) getGoogleWorkspaceGroups(service *admin.Service) ([]*admin.Group, error) {
	allGroups := []*admin.Group{}
	pageToken := ""

	// Get the customer ID (use "my_customer" for the domain)
	customer := "my_customer"

	for {
		call := service.Groups.List().Customer(customer).MaxResults(500)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list groups: %v", err)
		}

		allGroups = append(allGroups, resp.Groups...)

		// Handle pagination
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allGroups, nil
}

// googleWorkspaceGroupToOriginalGroup converts Google Workspace group to Casdoor OriginalGroup
func (p *GoogleWorkspaceSyncerProvider) googleWorkspaceGroupToOriginalGroup(gwGroup *admin.Group) *OriginalGroup {
	group := &OriginalGroup{
		Id:          gwGroup.Id,
		Name:        gwGroup.Email,
		DisplayName: gwGroup.Name,
		Description: gwGroup.Description,
		Email:       gwGroup.Email,
	}

	return group
}
