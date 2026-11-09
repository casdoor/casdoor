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

package object

import (
	"encoding/json"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
)

// The settings of paid features are refused when saved without a license, so that they are never kept
// while silently not taking effect

func CheckApplicationEdition(application *Application) error {
	if application == nil || edition.Has(edition.Commercial) {
		return nil
	}
	if application.EnableExclusiveSignin {
		return edition.Require(edition.Commercial, edition.FeatureExclusiveSignin)
	}
	if application.InactiveTimeoutMinutes > 0 {
		return edition.Require(edition.Commercial, edition.FeatureInactiveTimeout)
	}
	if util.InSlice(application.GrantTypes, MobileOneClickGrantType) {
		return edition.Require(edition.Commercial, edition.FeatureMobileOneClick)
	}
	return nil
}

func CheckOrganizationEdition(organization *Organization) error {
	if organization == nil || edition.Has(edition.Commercial) {
		return nil
	}
	if organization.EnableExclusiveSignin {
		return edition.Require(edition.Commercial, edition.FeatureExclusiveSignin)
	}
	if len(organization.IdJagIssuers) > 0 {
		return edition.Require(edition.Commercial, edition.FeatureIdJag)
	}
	if !organization.SecurityPolicy.isEmpty() {
		return edition.Require(edition.Commercial, edition.FeatureSecurityPolicy)
	}
	return nil
}

// CheckSyncerEdition is checked when a syncer is saved, isNew when it is added
func CheckSyncerEdition(syncer *Syncer, isNew bool) error {
	if syncer == nil || edition.Has(edition.Commercial) {
		return nil
	}
	if syncer.Type == "Keycloak" {
		return edition.Require(edition.Commercial, edition.FeatureKeycloakSyncer)
	}
	if isNew {
		count, err := GetSyncerCount("", "", "", "")
		if err != nil {
			return err
		}
		if count >= 1 {
			return edition.Require(edition.Commercial, edition.FeatureMultipleSyncers)
		}
	}
	return nil
}

// isEmpty tells whether no setting of the policy is turned on, as the organization page sends an object of
// zero values once the security tab is opened
func (policy *SecurityPolicy) isEmpty() bool {
	if policy == nil {
		return true
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return false
	}
	fields := map[string]interface{}{}
	if err = json.Unmarshal(data, &fields); err != nil {
		return false
	}
	for _, value := range fields {
		switch v := value.(type) {
		case nil:
		case bool:
			if v {
				return false
			}
		case float64:
			if v != 0 {
				return false
			}
		case string:
			if v != "" {
				return false
			}
		case []interface{}:
			if len(v) > 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}
