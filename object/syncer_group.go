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

func (syncer *Syncer) createGroupFromOriginalGroup(originalGroup *OriginalGroup) *Group {
	group := &Group{
		Owner:       syncer.Organization,
		Name:        originalGroup.Name,
		CreatedTime: util.GetCurrentTime(),
		UpdatedTime: util.GetCurrentTime(),
		DisplayName: originalGroup.DisplayName,
		Type:        originalGroup.Type,
		Manager:     originalGroup.Manager,
		IsEnabled:   true,
		IsTopGroup:  true,
	}

	if originalGroup.Email != "" {
		group.ContactEmail = originalGroup.Email
	}

	return group
}

func (syncer *Syncer) syncGroups() error {
	fmt.Printf("Running syncGroups()..\n")

	// Get existing groups from Casdoor
	groups, err := GetGroups(syncer.Organization)
	if err != nil {
		line := fmt.Sprintf("[%s] %s\n", util.GetCurrentTime(), err.Error())
		_, err2 := updateSyncerErrorText(syncer, line)
		if err2 != nil {
			panic(err2)
		}
		return err
	}

	// Get groups from the external system
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
			newGroup := syncer.createGroupFromOriginalGroup(oGroup)
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
