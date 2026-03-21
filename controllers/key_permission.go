// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package controllers

import (
	"errors"
	"fmt"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) canManageKey(key *object.Key) (bool, error) {
	if key == nil {
		return false, errors.New(c.T("auth:Unauthorized operation"))
	}

	username := c.GetSessionUsername()
	if username == "" {
		return false, errors.New(c.T("general:Please login first"))
	}

	if object.IsAppUser(username) {
		return true, nil
	}

	currentUser := c.getCurrentUser()
	if currentUser == nil {
		return false, errors.New(c.T("general:Please login first"))
	}

	if currentUser.IsGlobalAdmin() {
		return true, nil
	}

	targetOrganization, err := getKeyTargetOrganization(key)
	if err != nil {
		return false, err
	}

	if currentUser.IsAdmin {
		if targetOrganization == currentUser.Owner {
			return true, nil
		}
		return false, errors.New(c.T("auth:Unauthorized operation"))
	}

	if key.Type == object.KeyTypeUser && targetOrganization == currentUser.Owner && key.User == currentUser.Name {
		return true, nil
	}

	return false, errors.New(c.T("auth:Unauthorized operation"))
}

func getKeyTargetOrganization(key *object.Key) (string, error) {
	switch key.Type {
	case object.KeyTypeOrganization, object.KeyTypeUser:
		if key.Organization == "" {
			return "", fmt.Errorf("key organization cannot be empty")
		}
		return key.Organization, nil
	case object.KeyTypeApplication, object.KeyTypeGeneral:
		if key.Application == "" {
			return "", fmt.Errorf("key application cannot be empty")
		}

		application, err := object.GetApplication(util.GetId("admin", key.Application))
		if err != nil {
			return "", err
		}
		if application == nil {
			return "", fmt.Errorf("the application: %s does not exist", key.Application)
		}
		if application.Organization == "" {
			return "", fmt.Errorf("the application: %s has no organization", key.Application)
		}

		return application.Organization, nil
	default:
		return "", fmt.Errorf("unsupported key type: %s", key.Type)
	}
}

func (c *ApiController) filterAuthorizedKeys(keys []*object.Key) ([]*object.Key, error) {
	filteredKeys := make([]*object.Key, 0, len(keys))
	for _, key := range keys {
		ok, err := c.canManageKey(key)
		if err != nil {
			return nil, err
		}
		if ok {
			filteredKeys = append(filteredKeys, key)
		}
	}

	return filteredKeys, nil
}
