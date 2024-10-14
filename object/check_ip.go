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

func CheckEntryIp(user *User, application *Application, organization *Organization, remoteAddress string, lang string) string {
	entryIp, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		return err.Error()
	}

	if user != nil && !isEntryIpAllowd(user.IpWhitelist, entryIp) {
		return fmt.Sprintf(i18n.Translate(lang, "check:Your IP address: %s has been banned according to the configuration of: %s"), entryIp, user.Name)
	}

	if application != nil && !isEntryIpAllowd(application.IpWhitelist, entryIp) {
		return fmt.Sprintf(i18n.Translate(lang, "check:Your IP address: %s has been banned according to the configuration of: %s"), entryIp, application.Name)
	}

	if organization != nil && !isEntryIpAllowd(organization.IpWhitelist, entryIp) {
		return fmt.Sprintf(i18n.Translate(lang, "check:Your IP address: %s has been banned according to the configuration of: %s"), entryIp, organization.Name)
	}

	return ""
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

func CheckIpWhitelist(ipWhitelistStr string, lang string) string {
	if ipWhitelistStr == "" {
		return ""
	}

	ipWhiteList := strings.Split(ipWhitelistStr, ",")
	for _, ip := range ipWhiteList {
		if _, _, err := net.ParseCIDR(ip); err != nil {
			return fmt.Sprintf(i18n.Translate(lang, "check:%s does not meet the CIDR format requirements: %s"), ip, err.Error())
		}
	}

	return ""
}
