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
)

func (syncer *Syncer) getOriginalGroups() ([]*OriginalGroup, error) {
	provider := GetSyncerProvider(syncer)
	return provider.GetOriginalGroups()
}

func (syncer *Syncer) createGroupFromOriginalGroup(originalGroup *OriginalGroup, targetOrg string) *Group {
	group := &Group{
		Owner:       targetOrg,
		Name:        originalGroup.Name,
		CreatedTime: util.GetCurrentTime(),
		UpdatedTime: util.GetCurrentTime(),
		DisplayName: originalGroup.DisplayName,
		Type:        originalGroup.Type,
		Manager:     originalGroup.Manager,
		IsEnabled:   true,
		IsTopGroup:  originalGroup.IsTopGroup,
		ParentId:    originalGroup.ParentId,
	}

	// If no parent specified, set as top group with organization as parent
	if group.ParentId == "" {
		group.IsTopGroup = true
		group.ParentId = targetOrg
	}

	if originalGroup.Email != "" {
		group.ContactEmail = originalGroup.Email
	}

	return group
}

func (syncer *Syncer) syncGroups() error {
	fmt.Printf("Running syncGroups()..\n")
	fmt.Printf("Syncer type: %s, Organization: %s\n", syncer.Type, syncer.Organization)

	// Check if the provider supports company info
	provider := GetSyncerProvider(syncer)
	targetOrg := syncer.Organization

	// If provider supports CompanyInfoProvider, create/get Organization from company info
	if companyProvider, ok := provider.(CompanyInfoProvider); ok {
		fmt.Printf("Provider supports CompanyInfoProvider, syncing company to organization...\n")
		newOrg, err := syncer.syncCompanyToOrganization(companyProvider)
		if err != nil {
			fmt.Printf("Warning: failed to sync company to organization: %v\n", err)
			// Continue with original organization
		} else if newOrg != "" {
			// Use the new organization name for groups
			targetOrg = newOrg
			fmt.Printf("Using organization from company: %s\n", targetOrg)
		} else {
			fmt.Printf("syncCompanyToOrganization returned empty org name, using original: %s\n", targetOrg)
		}
	} else {
		fmt.Printf("Provider does NOT support CompanyInfoProvider, using original organization: %s\n", targetOrg)
	}

	// Get existing groups from Casdoor (from target organization)
	groups, err := GetGroups(targetOrg)
	if err != nil {
		line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
		_, err2 := updateSyncerErrorText(syncer, line)
		if err2 != nil {
			panic(err2)
		}
		return err
	}

	// Get groups from the external system (sub-departments, excluding company/root)
	oGroups, err := syncer.getOriginalGroups()
	if err != nil {
		line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
		_, err2 := updateSyncerErrorText(syncer, line)
		if err2 != nil {
			panic(err2)
		}
		return err
	}

	fmt.Printf("Groups: %d, oGroups: %d\n", len(groups), len(oGroups))

	// Create a map of existing groups by name
	myGroups := map[string]*Group{}
	for _, group := range groups {
		myGroups[group.Name] = group
	}

	// Sync groups from external system to Casdoor
	newGroups := []*Group{}
	for _, oGroup := range oGroups {
		if _, ok := myGroups[oGroup.Name]; !ok {
			newGroup := syncer.createGroupFromOriginalGroup(oGroup, targetOrg)
			fmt.Printf("New group: %v\n", newGroup)
			newGroups = append(newGroups, newGroup)
		} else {
			// Group already exists, could update it here if needed
			existingGroup := myGroups[oGroup.Name]

			// Update group display name and other fields if they've changed
			if existingGroup.DisplayName != oGroup.DisplayName {
				existingGroup.DisplayName = oGroup.DisplayName
				existingGroup.UpdatedTime = util.GetCurrentTime()
				_, err = UpdateGroup(existingGroup.GetId(), existingGroup)
				if err != nil {
					fmt.Printf("Failed to update group %s: %v\n", existingGroup.Name, err)
				} else {
					fmt.Printf("Updated group: %s\n", existingGroup.Name)
				}
			}
		}
	}

	if len(newGroups) != 0 {
		_, err = AddGroupsInBatch(newGroups)
		if err != nil {
			return err
		}
	}

	return nil
}

func (syncer *Syncer) syncGroupsNoError() {
	err := syncer.syncGroups()
	if err != nil {
		fmt.Printf("syncGroupsNoError() error: %s\n", err.Error())
	}
}

// syncCompanyToOrganization creates or updates Organization based on company info from external system
// Returns the organization name to use for syncing groups and users
func (syncer *Syncer) syncCompanyToOrganization(provider CompanyInfoProvider) (string, error) {
	fmt.Printf("Syncing company info to create/update organization...\n")

	// Get company info from external system
	company, err := provider.GetCompanyInfo()
	if err != nil {
		return "", fmt.Errorf("failed to get company info: %w", err)
	}

	if company == nil || company.Name == "" {
		fmt.Printf("No company info returned from provider\n")
		return "", nil
	}

	// Use company name as organization name
	orgName := company.Name

	fmt.Printf("Company name from external system: %s\n", orgName)

	// Check if organization with this name already exists
	existingOrg, err := GetOrganization(util.GetId("admin", orgName))
	if err != nil {
		return "", fmt.Errorf("failed to check organization %s: %w", orgName, err)
	}

	if existingOrg != nil {
		// Organization already exists, update it if needed
		fmt.Printf("Organization %s already exists, checking for updates...\n", orgName)

		needUpdate := false

		if company.DisplayName != "" && existingOrg.DisplayName != company.DisplayName {
			fmt.Printf("Updating organization DisplayName: %s -> %s\n", existingOrg.DisplayName, company.DisplayName)
			existingOrg.DisplayName = company.DisplayName
			needUpdate = true
		}

		if company.Logo != "" && existingOrg.Logo != company.Logo {
			fmt.Printf("Updating organization Logo: %s -> %s\n", existingOrg.Logo, company.Logo)
			existingOrg.Logo = company.Logo
			needUpdate = true
		}

		if company.WebsiteUrl != "" && existingOrg.WebsiteUrl != company.WebsiteUrl {
			fmt.Printf("Updating organization WebsiteUrl: %s -> %s\n", existingOrg.WebsiteUrl, company.WebsiteUrl)
			existingOrg.WebsiteUrl = company.WebsiteUrl
			needUpdate = true
		}

		if needUpdate {
			orgId := util.GetId(existingOrg.Owner, existingOrg.Name)
			_, err = UpdateOrganization(orgId, existingOrg, true)
			if err != nil {
				return "", fmt.Errorf("failed to update organization: %w", err)
			}
			fmt.Printf("Organization %s updated\n", orgName)
		} else {
			fmt.Printf("Organization %s already up to date\n", orgName)
		}

		// Update syncer to point to this organization
		if syncer.Organization != orgName {
			syncer.Organization = orgName
			_, err = UpdateSyncer(syncer.GetId(), syncer, true, "en")
			if err != nil {
				fmt.Printf("Warning: failed to update syncer organization: %v\n", err)
			} else {
				fmt.Printf("Syncer organization updated to: %s\n", orgName)
			}
		}

		return orgName, nil
	}

	// Organization doesn't exist, create a new one
	fmt.Printf("Creating new organization: %s\n", orgName)

	newOrg := &Organization{
		Owner:       "admin",
		Name:        orgName,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: company.DisplayName,
		Logo:        company.Logo,
		WebsiteUrl:  company.WebsiteUrl,
	}

	// Set default display name if not provided
	if newOrg.DisplayName == "" {
		newOrg.DisplayName = orgName
	}

	_, err = AddOrganization(newOrg)
	if err != nil {
		return "", fmt.Errorf("failed to create organization %s: %w", orgName, err)
	}

	fmt.Printf("Organization %s created successfully\n", orgName)

	// Update syncer to point to the new organization
	if syncer.Organization != orgName {
		syncer.Organization = orgName
		_, err = UpdateSyncer(syncer.GetId(), syncer, true, "en")
		if err != nil {
			fmt.Printf("Warning: failed to update syncer organization: %v\n", err)
		} else {
			fmt.Printf("Syncer organization updated to: %s\n", orgName)
		}
	}

	return orgName, nil
}
