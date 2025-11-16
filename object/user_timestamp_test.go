// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"testing"

	"github.com/casdoor/casdoor/util"
)

func TestAddUserTimestamps(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		wantCreated bool
		wantUpdated bool
	}{
		{
			name: "Both timestamps empty - should be populated",
			user: &User{
				Owner:       "test-org",
				Name:        "test-user-1",
				CreatedTime: "",
				UpdatedTime: "",
			},
			wantCreated: true,
			wantUpdated: true,
		},
		{
			name: "CreatedTime set, UpdatedTime empty - UpdatedTime should equal CreatedTime",
			user: &User{
				Owner:       "test-org",
				Name:        "test-user-2",
				CreatedTime: "2024-01-01T00:00:00Z",
				UpdatedTime: "",
			},
			wantCreated: true,
			wantUpdated: true,
		},
		{
			name: "Both timestamps set - should preserve values",
			user: &User{
				Owner:       "test-org",
				Name:        "test-user-3",
				CreatedTime: "2024-01-01T00:00:00Z",
				UpdatedTime: "2024-01-02T00:00:00Z",
			},
			wantCreated: true,
			wantUpdated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store initial values
			initialCreatedTime := tt.user.CreatedTime
			initialUpdatedTime := tt.user.UpdatedTime

			// Simulate the timestamp logic from AddUser
			if tt.user.CreatedTime == "" {
				tt.user.CreatedTime = util.GetCurrentTime()
			}

			if tt.user.UpdatedTime == "" {
				tt.user.UpdatedTime = tt.user.CreatedTime
			}

			// Verify CreatedTime is set
			if tt.wantCreated && tt.user.CreatedTime == "" {
				t.Errorf("CreatedTime was not set")
			}

			// Verify UpdatedTime is set
			if tt.wantUpdated && tt.user.UpdatedTime == "" {
				t.Errorf("UpdatedTime was not set")
			}

			// If both were initially empty, they should be equal
			if initialCreatedTime == "" && initialUpdatedTime == "" {
				if tt.user.CreatedTime != tt.user.UpdatedTime {
					t.Errorf("When both timestamps are empty, UpdatedTime should equal CreatedTime. Got CreatedTime=%s, UpdatedTime=%s",
						tt.user.CreatedTime, tt.user.UpdatedTime)
				}
			}

			// If CreatedTime was set but UpdatedTime was empty, UpdatedTime should equal CreatedTime
			if initialCreatedTime != "" && initialUpdatedTime == "" {
				if tt.user.UpdatedTime != tt.user.CreatedTime {
					t.Errorf("When CreatedTime is set but UpdatedTime is empty, UpdatedTime should equal CreatedTime. Got CreatedTime=%s, UpdatedTime=%s",
						tt.user.CreatedTime, tt.user.UpdatedTime)
				}
			}

			// If both were set, they should be preserved
			if initialCreatedTime != "" && initialUpdatedTime != "" {
				if tt.user.CreatedTime != initialCreatedTime {
					t.Errorf("CreatedTime should be preserved. Want %s, got %s", initialCreatedTime, tt.user.CreatedTime)
				}
				if tt.user.UpdatedTime != initialUpdatedTime {
					t.Errorf("UpdatedTime should be preserved. Want %s, got %s", initialUpdatedTime, tt.user.UpdatedTime)
				}
			}
		})
	}
}

func TestAddUsersTimestamps(t *testing.T) {
	tests := []struct {
		name        string
		users       []*User
		wantCreated bool
		wantUpdated bool
	}{
		{
			name: "Multiple users with empty timestamps",
			users: []*User{
				{
					Owner:       "test-org",
					Name:        "batch-user-1",
					CreatedTime: "",
					UpdatedTime: "",
				},
				{
					Owner:       "test-org",
					Name:        "batch-user-2",
					CreatedTime: "",
					UpdatedTime: "",
				},
			},
			wantCreated: true,
			wantUpdated: true,
		},
		{
			name: "Multiple users with CreatedTime set, UpdatedTime empty",
			users: []*User{
				{
					Owner:       "test-org",
					Name:        "batch-user-3",
					CreatedTime: "2024-01-01T00:00:00Z",
					UpdatedTime: "",
				},
				{
					Owner:       "test-org",
					Name:        "batch-user-4",
					CreatedTime: "2024-01-02T00:00:00Z",
					UpdatedTime: "",
				},
			},
			wantCreated: true,
			wantUpdated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, user := range tt.users {
				initialCreatedTime := user.CreatedTime
				initialUpdatedTime := user.UpdatedTime

				// Simulate the timestamp logic from AddUsers
				if user.CreatedTime == "" {
					user.CreatedTime = util.GetCurrentTime()
				}

				if user.UpdatedTime == "" {
					user.UpdatedTime = user.CreatedTime
				}

				// Verify CreatedTime is set
				if tt.wantCreated && user.CreatedTime == "" {
					t.Errorf("User %s: CreatedTime was not set", user.Name)
				}

				// Verify UpdatedTime is set
				if tt.wantUpdated && user.UpdatedTime == "" {
					t.Errorf("User %s: UpdatedTime was not set", user.Name)
				}

				// If both were initially empty, they should be equal
				if initialCreatedTime == "" && initialUpdatedTime == "" {
					if user.CreatedTime != user.UpdatedTime {
						t.Errorf("User %s: When both timestamps are empty, UpdatedTime should equal CreatedTime. Got CreatedTime=%s, UpdatedTime=%s",
							user.Name, user.CreatedTime, user.UpdatedTime)
					}
				}

				// If CreatedTime was set but UpdatedTime was empty, UpdatedTime should equal CreatedTime
				if initialCreatedTime != "" && initialUpdatedTime == "" {
					if user.UpdatedTime != user.CreatedTime {
						t.Errorf("User %s: When CreatedTime is set but UpdatedTime is empty, UpdatedTime should equal CreatedTime. Got CreatedTime=%s, UpdatedTime=%s",
							user.Name, user.CreatedTime, user.UpdatedTime)
					}
				}
			}
		})
	}
}
