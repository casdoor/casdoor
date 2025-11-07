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

package idp

import (
	"testing"
)

func TestGetIdProvider_WeChat(t *testing.T) {
	tests := []struct {
		name        string
		subType     string
		wantType    string
		shouldError bool
	}{
		{
			name:        "WeChat Web (default)",
			subType:     "",
			wantType:    "*idp.WeChatIdProvider",
			shouldError: false,
		},
		{
			name:        "WeChat Web (explicit)",
			subType:     "Web",
			wantType:    "*idp.WeChatIdProvider",
			shouldError: false,
		},
		{
			name:        "WeChat Mobile",
			subType:     "Mobile",
			wantType:    "*idp.WeChatMobileIdProvider",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idpInfo := &ProviderInfo{
				Type:         "WeChat",
				SubType:      tt.subType,
				ClientId:     "test_client_id",
				ClientSecret: "test_client_secret",
			}

			provider, err := GetIdProvider(idpInfo, "http://localhost/callback")

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Errorf("expected provider but got nil")
				return
			}

			// Verify the correct provider type was returned
			switch tt.subType {
			case "Mobile":
				if _, ok := provider.(*WeChatMobileIdProvider); !ok {
					t.Errorf("expected WeChatMobileIdProvider but got %T", provider)
				}
			default:
				if _, ok := provider.(*WeChatIdProvider); !ok {
					t.Errorf("expected WeChatIdProvider but got %T", provider)
				}
			}
		})
	}
}

func TestGetIdProvider_WeCom(t *testing.T) {
	tests := []struct {
		name        string
		subType     string
		shouldError bool
	}{
		{
			name:        "WeCom Internal",
			subType:     "Internal",
			shouldError: false,
		},
		{
			name:        "WeCom Third-party",
			subType:     "Third-party",
			shouldError: false,
		},
		{
			name:        "WeCom Invalid",
			subType:     "Invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idpInfo := &ProviderInfo{
				Type:         "WeCom",
				SubType:      tt.subType,
				ClientId:     "test_client_id",
				ClientSecret: "test_client_secret",
			}

			provider, err := GetIdProvider(idpInfo, "http://localhost/callback")

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Errorf("expected provider but got nil")
			}
		})
	}
}
