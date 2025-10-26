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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// UserOrganization represents the membership of a user in an organization
type UserOrganization struct {
	Owner        string `xorm:"varchar(100) notnull pk" json:"owner"` // User's owner (original organization)
	Name         string `xorm:"varchar(100) notnull pk" json:"name"`  // User's name
	Organization string `xorm:"varchar(100) notnull pk" json:"organization"`
	CreatedTime  string `xorm:"varchar(100)" json:"createdTime"`
	IsDefault    bool   `json:"isDefault"` // Is this the user's primary organization
}

// GetUserOrganizations returns all organizations a user belongs to
func GetUserOrganizations(owner, name string) ([]*UserOrganization, error) {
	userOrganizations := []*UserOrganization{}
	err := ormer.Engine.Find(&userOrganizations, &UserOrganization{Owner: owner, Name: name})
	if err != nil {
		return nil, err
	}
	return userOrganizations, nil
}

// GetUserOrganization returns a specific user-organization relationship
func GetUserOrganization(owner, name, organization string) (*UserOrganization, error) {
	userOrganization := UserOrganization{Owner: owner, Name: name, Organization: organization}
	existed, err := ormer.Engine.Get(&userOrganization)
	if err != nil {
		return nil, err
	}

	if existed {
		return &userOrganization, nil
	}
	return nil, nil
}

// AddUserOrganization adds a user to an organization
func AddUserOrganization(userOrganization *UserOrganization) (bool, error) {
	affected, err := ormer.Engine.Insert(userOrganization)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// DeleteUserOrganization removes a user from an organization
func DeleteUserOrganization(owner, name, organization string) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{owner, name, organization}).Delete(&UserOrganization{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// UpdateUserOrganization updates a user-organization relationship
func UpdateUserOrganization(owner, name, organization string, userOrganization *UserOrganization) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{owner, name, organization}).AllCols().Update(userOrganization)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// GetUsersByOrganization returns all users in an organization
func GetUsersByOrganization(organization string) ([]*UserOrganization, error) {
	userOrganizations := []*UserOrganization{}
	err := ormer.Engine.Find(&userOrganizations, &UserOrganization{Organization: organization})
	if err != nil {
		return nil, err
	}
	return userOrganizations, nil
}

// GetId returns the ID of the user-organization relationship
func (uo *UserOrganization) GetId() string {
	return fmt.Sprintf("%s/%s/%s", uo.Owner, uo.Name, uo.Organization)
}

// EnsureUserOrganizationExists ensures a user-organization relationship exists
// This is called when a user is created to add them to their default organization
func EnsureUserOrganizationExists(user *User) error {
	// Check if user already has organization membership
	existing, err := GetUserOrganizations(user.Owner, user.Name)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

// PopulateUserOrganizations populates the Organizations field of a user
func PopulateUserOrganizations(user *User) error {
	if user == nil {
		return nil
	}

	organizations, err := GetUserOrganizationNames(user.Owner, user.Name)
	if err != nil {
		return err
	}

	user.Organizations = organizations
	return nil
}
