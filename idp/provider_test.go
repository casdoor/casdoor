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

func TestApplyUserMapping(t *testing.T) {
	tests := []struct {
		name        string
		rawData     map[string]interface{}
		userMapping map[string]string
		expectNil   bool
		wantId      string
		wantEmail   string
		wantError   bool
	}{
		{
			name:        "nil mapping returns nil",
			rawData:     map[string]interface{}{"sub": "123", "email": "test@example.com"},
			userMapping: nil,
			expectNil:   true,
		},
		{
			name:        "empty mapping returns nil",
			rawData:     map[string]interface{}{"sub": "123", "email": "test@example.com"},
			userMapping: map[string]string{},
			expectNil:   true,
		},
		{
			name: "simple mapping",
			rawData: map[string]interface{}{
				"sub":        "user123",
				"email":      "user@example.com",
				"name":       "John Doe",
				"given_name": "John",
			},
			userMapping: map[string]string{
				"id":          "sub",
				"email":       "email",
				"username":    "given_name",
				"displayName": "name",
			},
			expectNil: false,
			wantId:    "user123",
			wantEmail: "user@example.com",
		},
		{
			name: "nested field mapping",
			rawData: map[string]interface{}{
				"sub": "user456",
				"profile": map[string]interface{}{
					"email": "nested@example.com",
					"name":  "Jane Doe",
				},
			},
			userMapping: map[string]string{
				"id":          "sub",
				"email":       "profile.email",
				"displayName": "profile.name",
				"username":    "profile.name",
			},
			expectNil: false,
			wantId:    "user456",
			wantEmail: "nested@example.com",
		},
		{
			name: "partial mapping with missing fields",
			rawData: map[string]interface{}{
				"sub":   "user789",
				"email": "partial@example.com",
			},
			userMapping: map[string]string{
				"id":          "sub",
				"email":       "email",
				"username":    "nonexistent",
				"displayName": "alsoMissing",
			},
			expectNil: false,
			wantId:    "user789",
			wantEmail: "partial@example.com",
		},
		{
			name: "numeric id conversion",
			rawData: map[string]interface{}{
				"id":    123456,
				"email": "numeric@example.com",
			},
			userMapping: map[string]string{
				"id":          "id",
				"email":       "email",
				"username":    "email",
				"displayName": "email",
			},
			expectNil: false,
			wantId:    "123456",
			wantEmail: "numeric@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyUserMapping(tt.rawData, tt.userMapping)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectNil {
				if result != nil {
					t.Errorf("expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected non-nil result, got nil")
				return
			}

			if tt.wantId != "" && result.Id != tt.wantId {
				t.Errorf("expected Id=%s, got %s", tt.wantId, result.Id)
			}

			if tt.wantEmail != "" && result.Email != tt.wantEmail {
				t.Errorf("expected Email=%s, got %s", tt.wantEmail, result.Email)
			}
		})
	}
}

func TestGetNestedValue(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]interface{}
		path      string
		wantValue interface{}
		wantError bool
	}{
		{
			name:      "simple key",
			data:      map[string]interface{}{"key": "value"},
			path:      "key",
			wantValue: "value",
		},
		{
			name: "nested key",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"email": "test@example.com",
					},
				},
			},
			path:      "user.profile.email",
			wantValue: "test@example.com",
		},
		{
			name:      "non-existent key",
			data:      map[string]interface{}{"key": "value"},
			path:      "missing",
			wantError: true,
		},
		{
			name: "non-existent nested key",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
				},
			},
			path:      "user.profile.email",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getNestedValue(tt.data, tt.path)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.wantValue {
				t.Errorf("expected %v, got %v", tt.wantValue, result)
			}
		})
	}
}
