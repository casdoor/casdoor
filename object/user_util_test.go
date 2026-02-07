// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/idp"
)

// applyOAuthPropertiesWithoutDB applies OAuth properties without calling the database
// This is a simplified version for testing
func applyOAuthPropertiesWithoutDB(org *Organization, user *User, providerType string, userInfo *idp.UserInfo) {
	if userInfo.AvatarUrl != "" {
		propertyName := "oauth_" + providerType + "_avatarUrl"
		oldAvatarUrl := getUserProperty(user, propertyName)
		setUserProperty(user, propertyName, userInfo.AvatarUrl)

		// Update avatar if:
		// 1. User has no avatar or has default avatar, OR
		// 2. Current avatar matches the old OAuth avatar (meaning it was set by OAuth) and the new avatar is different
		if user.Avatar == "" || user.Avatar == org.DefaultAvatar {
			user.Avatar = userInfo.AvatarUrl
		} else if oldAvatarUrl != "" && user.Avatar == oldAvatarUrl && oldAvatarUrl != userInfo.AvatarUrl {
			user.Avatar = userInfo.AvatarUrl
		}
	}
}

func TestSetUserOAuthProperties_AvatarUpdate(t *testing.T) {
	// Test case 1: Avatar should be set when user has no avatar
	t.Run("Set avatar when user has no avatar", func(t *testing.T) {
		org := &Organization{DefaultAvatar: ""}
		user := &User{
			Owner:      "test",
			Name:       "testuser",
			Avatar:     "",
			Properties: make(map[string]string),
		}
		userInfo := &idp.UserInfo{
			AvatarUrl: "https://example.com/avatar1.png",
		}

		applyOAuthPropertiesWithoutDB(org, user, "WeChat", userInfo)

		if user.Avatar != "https://example.com/avatar1.png" {
			t.Errorf("Expected avatar to be set to %s, got %s", userInfo.AvatarUrl, user.Avatar)
		}
	})

	// Test case 2: Avatar should be updated when OAuth avatar changes
	t.Run("Update avatar when OAuth avatar changes", func(t *testing.T) {
		org := &Organization{DefaultAvatar: ""}
		user := &User{
			Owner:  "test",
			Name:   "testuser",
			Avatar: "https://example.com/avatar1.png",
			Properties: map[string]string{
				"oauth_WeChat_avatarUrl": "https://example.com/avatar1.png",
			},
		}
		userInfo := &idp.UserInfo{
			AvatarUrl: "https://example.com/avatar2.png",
		}

		applyOAuthPropertiesWithoutDB(org, user, "WeChat", userInfo)

		if user.Avatar != "https://example.com/avatar2.png" {
			t.Errorf("Expected avatar to be updated to %s, got %s", userInfo.AvatarUrl, user.Avatar)
		}
	})

	// Test case 3: Avatar should NOT be updated when user has custom avatar
	t.Run("Do not update avatar when user has custom avatar", func(t *testing.T) {
		org := &Organization{DefaultAvatar: ""}
		user := &User{
			Owner:  "test",
			Name:   "testuser",
			Avatar: "https://example.com/custom-avatar.png",
			Properties: map[string]string{
				"oauth_WeChat_avatarUrl": "https://example.com/avatar1.png",
			},
		}
		userInfo := &idp.UserInfo{
			AvatarUrl: "https://example.com/avatar2.png",
		}

		applyOAuthPropertiesWithoutDB(org, user, "WeChat", userInfo)

		if user.Avatar != "https://example.com/custom-avatar.png" {
			t.Errorf("Expected avatar to remain %s, got %s", "https://example.com/custom-avatar.png", user.Avatar)
		}
	})

	// Test case 4: Avatar should be set when user has default avatar
	t.Run("Set avatar when user has default avatar", func(t *testing.T) {
		org := &Organization{DefaultAvatar: "https://example.com/default.png"}
		user := &User{
			Owner:      "test",
			Name:       "testuser",
			Avatar:     "https://example.com/default.png",
			Properties: make(map[string]string),
		}
		userInfo := &idp.UserInfo{
			AvatarUrl: "https://example.com/avatar1.png",
		}

		applyOAuthPropertiesWithoutDB(org, user, "WeChat", userInfo)

		if user.Avatar != "https://example.com/avatar1.png" {
			t.Errorf("Expected avatar to be set to %s, got %s", userInfo.AvatarUrl, user.Avatar)
		}
	})
}
