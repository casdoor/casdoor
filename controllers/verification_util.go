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

package controllers

import (
	"fmt"

	"github.com/casdoor/casdoor/v2/object"
)

func (c *ApiController) checkOrgMasterVerificationCode(user *object.User, code string) (bool, error) {
	organization, err := object.GetOrganizationByUser(user)
	if err != nil {
		return false, err
	}
	if organization == nil {
		return false, fmt.Errorf("The organization: %s does not exist", user.Owner)
	}

	if organization.MasterVerificationCode != "" && organization.MasterVerificationCode == code {
		return true, nil
	}
	return false, nil
}
