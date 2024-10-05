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
	"fmt"
	"net"
	"strings"

	"github.com/casdoor/casdoor/i18n"
)

func CheckEntryIp(userId string, organizationId string, applicationId string, remoteAddress string, lang string) error {
	entryIp, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		return err
	}

	userValid, userCheckErr := checkEntryIpForUser(userId, entryIp, lang)
	organizationValid, organizationCheckErr := checkEntryIpForOrganization(organizationId, entryIp, lang)
	applicationValid, applicationCheckErr := checkEntryIpForApplcation(applicationId, entryIp, lang)

	if userCheckErr == nil && organizationCheckErr == nil && applicationCheckErr == nil {
		if userValid && organizationValid && applicationValid {
			return nil
		} else {
			return fmt.Errorf(i18n.Translate(lang, "check:Your IP address %s has been banned. If you think this is a mistake, please contact the administrator."), entryIp)
		}
	} else {
		checkErrorMsg := i18n.Translate(lang, "check:Failed to check entry ip: ")
		if userCheckErr != nil {
			checkErrorMsg += fmt.Sprintf(i18n.Translate(lang, "check:user check error: %s. "), userCheckErr.Error())
		}
		if organizationCheckErr != nil {
			checkErrorMsg += fmt.Sprintf(i18n.Translate(lang, "check:organization check error: %s. "), organizationCheckErr.Error())
		}
		if applicationCheckErr != nil {
			checkErrorMsg += fmt.Sprintf(i18n.Translate(lang, "check:application check error: %s. "), applicationCheckErr.Error())
		}
		return fmt.Errorf(checkErrorMsg)
	}
}

func checkEntryIpForUser(userId string, entryIp string, lang string) (bool, error) {
	if userId == "" || userId == "/" {
		return true, nil
	}

	user, err := GetUser(userId)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, fmt.Errorf(i18n.Translate(lang, "general:The user: %s doesn't exist"), userId)
	}

	return isRemoteAddressAllowd(user.LimitedIps, entryIp), nil
}

func checkEntryIpForOrganization(organizationId string, entryIp string, lang string) (bool, error) {
	if organizationId == "" || organizationId == "/" {
		return true, nil
	}

	organization, err := GetOrganization(organizationId)
	if err != nil {
		return false, err
	}
	if organization == nil {
		return false, fmt.Errorf(i18n.Translate(lang, "auth:The organization: %s does not exist"), organizationId)
	}

	return isRemoteAddressAllowd(organization.LimitedIps, entryIp), nil
}

func checkEntryIpForApplcation(applicationId string, entryIp string, lang string) (bool, error) {
	if applicationId == "" || applicationId == "/" {
		return true, nil
	}

	application, err := GetApplication(applicationId)
	if err != nil {
		return false, err
	}
	if application == nil {
		return false, fmt.Errorf(i18n.Translate(lang, "auth:The application: %s does not exist"), applicationId)
	}

	return isRemoteAddressAllowd(application.LimitedIps, entryIp), nil
}

func isRemoteAddressAllowd(limitedIpsStr string, entryIp string) bool {
	if limitedIpsStr == "" {
		return true
	}

	limitedIps := strings.Split(limitedIpsStr, ",")
	for _, limitedIp := range limitedIps {
		_, limitedIpNet, _ := net.ParseCIDR(limitedIp)
		if limitedIpNet != nil && limitedIpNet.Contains(net.ParseIP(entryIp)) {
			return false
		}
	}

	return true
}
