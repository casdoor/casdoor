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

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
)

// AcceptOrganizationInvitation allows an existing user to accept an invitation to join another organization
func AcceptOrganizationInvitation(userId string, invitationCode string, organizationName string, lang string) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(userId)

	// Get the user
	user, err := GetUser(userId)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, fmt.Errorf(i18n.Translate(lang, "general:The user does not exist"))
	}

	// Verify invitation code
	invitation, msg := GetInvitationByCode(invitationCode, organizationName, lang)
	if invitation == nil {
		if msg != "" {
			return false, fmt.Errorf(msg)
		}
		return false, fmt.Errorf(i18n.Translate(lang, "check:Invitation code is invalid"))
	}

	// Verify invitation is for this organization
	if invitation.Owner != organizationName {
		return false, fmt.Errorf(i18n.Translate(lang, "general:Invitation is not for this organization"))
	}

	// Check if the invitation has email restriction
	if invitation.Email != "" && invitation.Email != user.Email {
		return false, fmt.Errorf(i18n.Translate(lang, "check:This invitation is for a different email address"))
	}

	// Check if user is already a member
	existingMembership, err := GetUserOrganization(owner, name, organizationName)
	if err != nil {
		return false, err
	}
	if existingMembership != nil {
		return false, fmt.Errorf(i18n.Translate(lang, "general:User is already a member of this organization"))
	}

	// Create the user-organization relationship
	userOrg := &UserOrganization{
		Owner:        owner,
		Name:         name,
		Organization: organizationName,
		CreatedTime:  util.GetCurrentTime(),
		IsDefault:    false,
	}

	affected, err := AddUserOrganization(userOrg)
	if err != nil {
		return false, err
	}

	if affected {
		// Update invitation usage count
		invitation.UsedCount++
		_, err = UpdateInvitation(invitation.GetId(), invitation, lang)
		if err != nil {
			return false, err
		}
	}

	return affected, nil
}

// GetUserOrganizationNames returns the list of organization names a user belongs to
func GetUserOrganizationNames(owner, name string) ([]string, error) {
	userOrganizations, err := GetUserOrganizations(owner, name)
	if err != nil {
		return nil, err
	}

	organizations := make([]string, len(userOrganizations))
	for i, uo := range userOrganizations {
		organizations[i] = uo.Organization
	}

	return organizations, nil
}

// IsUserMemberOfOrganization checks if a user is a member of an organization
func IsUserMemberOfOrganization(owner, name, organization string) (bool, error) {
	userOrg, err := GetUserOrganization(owner, name, organization)
	if err != nil {
		return false, err
	}
	return userOrg != nil, nil
}
