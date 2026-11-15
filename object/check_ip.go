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
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/casdoor/casdoor/i18n"
)

func CheckEntryIp(clientIp string, user *User, application *Application, organization *Organization, lang string) error {
	entryIp := net.ParseIP(clientIp)
	if entryIp == nil {
		return fmt.Errorf(i18n.Translate(lang, "check:Failed to parse client IP: %s"), clientIp)
	} else if entryIp.IsLoopback() {
		return nil
	}

	var err error
	if user != nil {
		err = isEntryIpAllowd(user.IpWhitelist, entryIp, lang)
		if err != nil {
			return errors.New(err.Error() + user.Name)
		}
	}

	if application != nil {
		err = isEntryIpAllowd(application.IpWhitelist, entryIp, lang)
		if err != nil {
			application.IpRestriction = err.Error() + application.Name
			return errors.New(err.Error() + application.Name)
		} else {
			application.IpRestriction = ""
		}

		if organization == nil && application.OrganizationObj != nil {
			organization = application.OrganizationObj
		}
	}

	if organization != nil {
		err = isEntryIpAllowd(organization.IpWhitelist, entryIp, lang)
		if err != nil {
			organization.IpRestriction = err.Error() + organization.Name
			return errors.New(err.Error() + organization.Name)
		} else {
			organization.IpRestriction = ""
		}
	}

	return nil
}

func isEntryIpAllowd(ipWhitelistStr string, entryIp net.IP, lang string) error {
	if ipWhitelistStr == "" {
		return nil
	}

	ipWhitelist := splitIpWhitelist(ipWhitelistStr)
	if len(ipWhitelist) == 0 {
		return nil
	}

	for _, ip := range ipWhitelist {
		ipNet, err := parseIpWhitelistEntry(ip)
		if err != nil {
			return err
		}
		if ipNet == nil {
			return fmt.Errorf(i18n.Translate(lang, "check:CIDR for IP: %s should not be empty"), entryIp.String())
		}

		if ipNet.Contains(entryIp) {
			return nil
		}
	}

	return fmt.Errorf(i18n.Translate(lang, "check:Your IP address: %s has been banned according to the configuration of: "), entryIp.String())
}

func CheckIpWhitelist(ipWhitelistStr string, lang string) error {
	if ipWhitelistStr == "" {
		return nil
	}

	for _, ip := range splitIpWhitelist(ipWhitelistStr) {
		if _, err := parseIpWhitelistEntry(ip); err != nil {
			return fmt.Errorf(i18n.Translate(lang, "check:%s does not meet the CIDR format requirements: %s"), ip, err.Error())
		}
	}

	return nil
}

func splitIpWhitelist(ipWhitelistStr string) []string {
	res := []string{}
	for _, ip := range strings.FieldsFunc(ipWhitelistStr, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	}) {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			res = append(res, ip)
		}
	}
	return res
}

func parseIpWhitelistEntry(ip string) (*net.IPNet, error) {
	if !strings.Contains(ip, "/") {
		parsedIp := net.ParseIP(ip)
		if parsedIp == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ip)
		}
		if parsedIp.To4() != nil {
			return &net.IPNet{IP: parsedIp.To4(), Mask: net.CIDRMask(32, 32)}, nil
		}
		return &net.IPNet{IP: parsedIp, Mask: net.CIDRMask(128, 128)}, nil
	}

	_, ipNet, err := net.ParseCIDR(ip)
	return ipNet, err
}
