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
	"fmt"
	"time"

	"github.com/casdoor/casdoor/util"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// getGoogleWorkspaceAdminService creates and returns a Google Workspace Admin SDK service
func (syncer *Syncer) getGoogleWorkspaceAdminService() (*admin.Service, error) {
	// syncer.Host should contain the service account JSON credentials
	// syncer.User should contain the admin email for domain-wide delegation

	if syncer.Host == "" {
		return nil, fmt.Errorf("service account credentials (host field) are required for Google Workspace syncer")
	}

	if syncer.User == "" {
		return nil, fmt.Errorf("admin email (user field) is required for Google Workspace syncer")
	}

	// Parse the service account credentials
	credentialsJSON := []byte(syncer.Host)

	// Create JWT config for domain-wide delegation
	config, err := google.JWTConfigFromJSON(credentialsJSON, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account credentials: %v", err)
	}

	// Set the subject (admin user email) for domain-wide delegation
	config.Subject = syncer.User

	// Create admin service
	ctx := context.Background()
	client := config.Client(ctx)

	service, err := admin.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create admin service: %v", err)
	}

	return service, nil
}

// getGoogleWorkspaceUsers retrieves all users from Google Workspace
func (syncer *Syncer) getGoogleWorkspaceUsers(service *admin.Service) ([]*admin.User, error) {
	allUsers := []*admin.User{}

	// Get customer ID (use "my_customer" for the domain)
	customer := "my_customer"

	// List all users with pagination
	pageToken := ""
	for {
		call := service.Users.List().Customer(customer).MaxResults(500)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		users, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %v", err)
		}

		allUsers = append(allUsers, users.Users...)

		// Check if there are more pages
		if users.NextPageToken == "" {
			break
		}
		pageToken = users.NextPageToken
	}

	return allUsers, nil
}

// googleWorkspaceUserToOriginalUser converts Google Workspace user to Casdoor OriginalUser
func (syncer *Syncer) googleWorkspaceUserToOriginalUser(gwUser *admin.User) *OriginalUser {
	user := &OriginalUser{
		Id:          gwUser.Id,
		Name:        gwUser.PrimaryEmail,
		DisplayName: "",
		FirstName:   "",
		LastName:    "",
		Email:       gwUser.PrimaryEmail,
		Phone:       "",
		Avatar:      gwUser.ThumbnailPhotoUrl,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
		IsForbidden: gwUser.Suspended,
	}

	// Extract name information
	if gwUser.Name != nil {
		user.DisplayName = gwUser.Name.FullName
		user.FirstName = gwUser.Name.GivenName
		user.LastName = gwUser.Name.FamilyName
	}

	// If display name is empty, construct from first and last name
	if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
		user.DisplayName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}

	// Extract phone number (use first phone if available)
	if phones, ok := gwUser.Phones.([]interface{}); ok && len(phones) > 0 {
		if phoneMap, ok := phones[0].(map[string]interface{}); ok {
			if val, exists := phoneMap["value"]; exists {
				user.Phone = fmt.Sprintf("%v", val)
			}
		}
	}

	// Extract job title and department from organizations
	if orgs, ok := gwUser.Organizations.([]interface{}); ok && len(orgs) > 0 {
		if orgMap, ok := orgs[0].(map[string]interface{}); ok {
			if title, exists := orgMap["title"]; exists {
				user.Title = fmt.Sprintf("%v", title)
			}
			if dept, exists := orgMap["department"]; exists {
				user.Affiliation = fmt.Sprintf("%v", dept)
			}
		}
	}

	// Extract location
	if locations, ok := gwUser.Locations.([]interface{}); ok && len(locations) > 0 {
		if locMap, ok := locations[0].(map[string]interface{}); ok {
			if area, exists := locMap["area"]; exists && area != "" {
				user.Location = fmt.Sprintf("%v", area)
			} else if building, exists := locMap["buildingId"]; exists && building != "" {
				user.Location = fmt.Sprintf("%v", building)
			}
		}
	}

	// Extract language
	if languages, ok := gwUser.Languages.([]interface{}); ok && len(languages) > 0 {
		if langMap, ok := languages[0].(map[string]interface{}); ok {
			if code, exists := langMap["languageCode"]; exists {
				user.Language = fmt.Sprintf("%v", code)
			}
		}
	}

	// Store organization unit path in properties
	if gwUser.OrgUnitPath != "" {
		user.Properties["orgUnitPath"] = gwUser.OrgUnitPath
	}

	// Parse and set creation time
	if gwUser.CreationTime != "" {
		// Google Workspace returns time in RFC3339 format
		t, err := time.Parse(time.RFC3339, gwUser.CreationTime)
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

// getGoogleWorkspaceOriginalUsers is the main entry point for Google Workspace syncer
func (syncer *Syncer) getGoogleWorkspaceOriginalUsers() ([]*OriginalUser, error) {
	// Get admin service
	service, err := syncer.getGoogleWorkspaceAdminService()
	if err != nil {
		return nil, err
	}

	// Get all users from Google Workspace
	gwUsers, err := syncer.getGoogleWorkspaceUsers(service)
	if err != nil {
		return nil, err
	}

	// Convert Google Workspace users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, gwUser := range gwUsers {
		originalUser := syncer.googleWorkspaceUserToOriginalUser(gwUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// testGoogleWorkspaceConnection tests the connection to Google Workspace
func (syncer *Syncer) testGoogleWorkspaceConnection() error {
	// Try to get admin service and list users (limited to 1)
	service, err := syncer.getGoogleWorkspaceAdminService()
	if err != nil {
		return err
	}

	// Try to list just one user to verify connection
	_, err = service.Users.List().Customer("my_customer").MaxResults(1).Do()
	if err != nil {
		return fmt.Errorf("failed to test connection: %v", err)
	}

	return nil
}
