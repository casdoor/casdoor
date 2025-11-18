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
	"fmt"
	"testing"
)

func TestGetSyncerProvider(t *testing.T) {
	tests := []struct {
		name         string
		syncerType   string
		expectedType string
	}{
		{
			name:         "WeCom syncer",
			syncerType:   "WeCom",
			expectedType: "*object.WecomSyncerProvider",
		},
		{
			name:         "Azure AD syncer",
			syncerType:   "Azure AD",
			expectedType: "*object.AzureAdSyncerProvider",
		},
		{
			name:         "Google Workspace syncer",
			syncerType:   "Google Workspace",
			expectedType: "*object.GoogleWorkspaceSyncerProvider",
		},
		{
			name:         "Keycloak syncer",
			syncerType:   "Keycloak",
			expectedType: "*object.KeycloakSyncerProvider",
		},
		{
			name:         "Database syncer",
			syncerType:   "Database",
			expectedType: "*object.DatabaseSyncerProvider",
		},
		{
			name:         "Default to database syncer",
			syncerType:   "Unknown",
			expectedType: "*object.DatabaseSyncerProvider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer := &Syncer{
				Type: tt.syncerType,
			}
			provider := GetSyncerProvider(syncer)
			if provider == nil {
				t.Errorf("GetSyncerProvider() returned nil for type %s", tt.syncerType)
				return
			}

			// Check the type of the provider
			providerType := getTypeName(provider)
			if providerType != tt.expectedType {
				t.Errorf("GetSyncerProvider() for type %s returned %s, expected %s", tt.syncerType, providerType, tt.expectedType)
			}
		})
	}
}

func getTypeName(i interface{}) string {
	return fmt.Sprintf("%T", i)
}
