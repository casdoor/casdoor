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

// KeycloakSyncerProvider implements SyncerProvider for Keycloak database syncers
// Keycloak syncer extends DatabaseSyncerProvider with special handling for Keycloak schema
type KeycloakSyncerProvider struct {
	DatabaseSyncerProvider
}

// GetOriginalUsers retrieves all users from Keycloak database
// This method overrides the base implementation to handle Keycloak-specific logic
func (p *KeycloakSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	// Use the base database implementation
	return p.DatabaseSyncerProvider.GetOriginalUsers()
}

// Note: Keycloak-specific user mapping is handled in syncer_util.go
// via getOriginalUsersFromMap which checks syncer.Type == "Keycloak"

// GetOriginalGroups retrieves all groups from Keycloak (not implemented yet)
func (p *KeycloakSyncerProvider) GetOriginalGroups() ([]*OriginalGroup, error) {
	// TODO: Implement Keycloak group sync
	return []*OriginalGroup{}, nil
}

// GetOriginalUserGroups retrieves the group IDs that a user belongs to (not implemented yet)
func (p *KeycloakSyncerProvider) GetOriginalUserGroups(userId string) ([]string, error) {
	// TODO: Implement Keycloak user group membership sync
	return []string{}, nil
}
