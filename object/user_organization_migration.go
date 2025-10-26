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
	"github.com/beego/beego/logs"
	"github.com/casdoor/casdoor/util"
)

// MigrateUserOrganizations creates UserOrganization entries for existing users
// This should be called during startup to ensure all existing users have organization memberships
func MigrateUserOrganizations() error {
	// Get all users
	users := []*User{}
	err := ormer.Engine.Find(&users)
	if err != nil {
		return err
	}

	logs.Info("Starting user-organization migration for %d users", len(users))

	migratedCount := 0
	for _, user := range users {
		// Check if user already has organization membership
		existing, err := GetUserOrganizations(user.Owner, user.Name)
		if err != nil {
			logs.Error("Error checking organizations for user %s/%s: %v", user.Owner, user.Name, err)
			continue
		}

		// If no memberships exist, create the default one
		if len(existing) == 0 {
			userOrg := &UserOrganization{
				Owner:        user.Owner,
				Name:         user.Name,
				Organization: user.Owner,
				CreatedTime:  util.GetCurrentTime(),
				IsDefault:    true,
			}
			_, err := AddUserOrganization(userOrg)
			if err != nil {
				logs.Error("Error creating organization membership for user %s/%s: %v", user.Owner, user.Name, err)
				continue
			}
			migratedCount++
		}
	}

	logs.Info("User-organization migration completed. Migrated %d users", migratedCount)
	return nil
}
