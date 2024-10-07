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

func CheckEntryIpByApplicationIdAndOrganizationId(applicationId string, organizationId string, remoteAddress string, lang string) error {
	var organization *Organization
	var application *Application
	var err error

	if organizationId != "" && organizationId != "/" {
		organization, err = GetOrganization(organizationId)
		if err != nil {
			return err
		}
		if organization == nil {
			return fmt.Errorf(i18n.Translate(lang, "auth:The organization: %s does not exist"), organizationId)
		}
	}

	if applicationId != "" && applicationId != "/" {
		application, err = GetApplication(applicationId)
		if err != nil {
			return err
		}
		if application == nil {
			return fmt.Errorf(i18n.Translate(lang, "auth:The application: %s does not exist"), applicationId)
		}
	}

	return checkEntryIpByObject(nil, organization, application, remoteAddress, lang)
}

func CheckEntryIpByUser(user *User, remoteAddress string, lang string) error {
	if user == nil {
		return fmt.Errorf(i18n.Translate(lang, "general:User doesn't exist"))
	}

	organization, err := GetOrganizationByUser(user)
	if err != nil {
		return err
	}
	if organization == nil {
		return fmt.Errorf(i18n.Translate(lang, "auth:The organization: %s does not exist"), user.Owner)
	}

	application, err := GetApplicationByUser(user)
	if err != nil {
		return err
	}
	if application == nil {
		return fmt.Errorf(i18n.Translate(lang, "util:No application is found for userId: %s"), user.GetId())
	}

	return checkEntryIpByObject(user, organization, application, remoteAddress, lang)
}

func checkEntryIpByObject(user *User, organization *Organization, application *Application, remoteAddress string, lang string) error {
	entryIp, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		return err
	}

	if user != nil && !isEntryIpAllowd(user.IpWhitelist, entryIp) {
		return fmt.Errorf(i18n.Translate(lang, "check:Your IP address %s has been banned according to the configuration of user %s"), entryIp, user.Name)
	}

	if application != nil && !isEntryIpAllowd(application.IpWhitelist, entryIp) {
		return fmt.Errorf(i18n.Translate(lang, "check:Your IP address %s has been banned according to the configuration of application %s"), entryIp, application.Name)
	}

	if organization != nil && !isEntryIpAllowd(organization.IpWhitelist, entryIp) {
		return fmt.Errorf(i18n.Translate(lang, "check:Your IP address %s has been banned according to the configuration of organization %s"), entryIp, organization.Name)
	}

	return nil
}

func isEntryIpAllowd(ipWhitelistStr string, entryIp string) bool {
	if ipWhitelistStr == "" {
		return true
	}

	ipWhitelist := strings.Split(ipWhitelistStr, ",")
	for _, ip := range ipWhitelist {
		_, ipNet, _ := net.ParseCIDR(ip)
		if ipNet != nil && ipNet.Contains(net.ParseIP(entryIp)) {
			return true
		}
	}

	return false
}
