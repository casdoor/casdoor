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
	"regexp"

	"github.com/casdoor/casdoor/edition"
)

// Casbin escapes "r.sub" in matchers to "r_sub", so both forms are matched
var abacMatcherRegex = regexp.MustCompile(`\b[rp][0-9]*[._][A-Za-z_]\w*\.[A-Za-z_]|\beval\(`)

func isAbacAllowed() bool {
	return edition.Allows(edition.Commercial, edition.FeatureAbac)
}

func isAbacModelText(modelText string) bool {
	return abacMatcherRegex.MatchString(modelText)
}

// CheckPermissionEdition is checked when a permission is saved, oldPermission is nil for a new one.
// Conditions already saved keep working without a license, only adding them needs it
func CheckPermissionEdition(permission *Permission, oldPermission *Permission) error {
	if permission == nil {
		return nil
	}
	if permission.ResourceType == PermissionResourceTypeMcpTool || len(permission.Agents) > 0 {
		if err := edition.Require(edition.Commercial, edition.FeatureMcpGateway); err != nil {
			return err
		}
	}
	if len(permission.Conditions) > 0 && (oldPermission == nil || len(oldPermission.Conditions) == 0) {
		if err := edition.Require(edition.Commercial, edition.FeatureAbac); err != nil {
			return err
		}
	}
	return nil
}

// CheckModelEdition is checked when a model is saved, oldModel is nil for a new one.
// An ABAC model already saved keeps working without a license, only adding one needs it
func CheckModelEdition(m *Model, oldModel *Model) error {
	if m == nil || isAbacAllowed() || !isAbacModelText(m.ModelText) {
		return nil
	}
	if oldModel != nil && isAbacModelText(oldModel.ModelText) {
		return nil
	}
	return edition.Require(edition.Commercial, edition.FeatureAbac)
}
