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

// OriginalGroup represents a group from an external system
type OriginalGroup struct {
	Id          string
	Name        string
	DisplayName string
	Description string
	Type        string
	Manager     string
	Email       string
}

// SyncerProvider defines the interface that all syncer implementations must satisfy.
// Different syncer types (Database, Keycloak, WeCom, Azure AD) implement this interface.
type SyncerProvider interface {
	// InitAdapter initializes the connection to the external system
	InitAdapter() error

	// GetOriginalUsers retrieves all users from the external system
	GetOriginalUsers() ([]*OriginalUser, error)

	// GetOriginalGroups retrieves all groups from the external system
	GetOriginalGroups() ([]*OriginalGroup, error)

	// GetOriginalUserGroups retrieves the group IDs that a user belongs to
	GetOriginalUserGroups(userId string) ([]string, error)

	// AddUser adds a new user to the external system
	AddUser(user *OriginalUser) (bool, error)

	// UpdateUser updates an existing user in the external system
	UpdateUser(user *OriginalUser) (bool, error)

	// TestConnection tests the connection to the external system
	TestConnection() error

	// Close closes any open connections and releases resources
	Close() error
}

// GetSyncerProvider returns the appropriate SyncerProvider implementation based on syncer type
func GetSyncerProvider(syncer *Syncer) SyncerProvider {
	switch syncer.Type {
	case "WeCom":
		return &WecomSyncerProvider{Syncer: syncer}
	case "Azure AD":
		return &AzureAdSyncerProvider{Syncer: syncer}
	case "Google Workspace":
		return &GoogleWorkspaceSyncerProvider{Syncer: syncer}
	case "Active Directory":
		return &ActiveDirectorySyncerProvider{Syncer: syncer}
	case "DingTalk":
		return &DingtalkSyncerProvider{Syncer: syncer}
	case "Lark":
		return &LarkSyncerProvider{Syncer: syncer}
	case "Okta":
		return &OktaSyncerProvider{Syncer: syncer}
	case "SCIM":
		return &SCIMSyncerProvider{Syncer: syncer}
	case "AWS IAM":
		return &AwsIamSyncerProvider{Syncer: syncer}
	case "Keycloak":
		return &KeycloakSyncerProvider{
			DatabaseSyncerProvider: DatabaseSyncerProvider{Syncer: syncer},
		}
	default:
		// Default to database syncer for "Database" type and any others
		return &DatabaseSyncerProvider{Syncer: syncer}
	}
}
