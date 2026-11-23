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
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/ip"
)

// GetOrganizationEmailProvider picks the email provider of the organization's default application, then any
// email provider owned by the organization, then the default Email provider of the global setting
func GetOrganizationEmailProvider(organization *Organization) (*Provider, error) {
	if organization.DefaultApplication != "" {
		application, err := LookupApplication("admin", organization.DefaultApplication)
		if err != nil {
			return nil, err
		}
		if application != nil {
			provider, err := application.GetProviderByCategoryAndRule("Email", "", "All")
			if err != nil {
				return nil, err
			}
			if provider != nil {
				return provider, nil
			}
		}
	}

	providers := []*Provider{}
	err := ormer.Engine.Where("category = ? and (owner = ? or owner = ?)", "Email", organization.Name, "admin").Find(&providers)
	if err != nil {
		return nil, err
	}
	for _, provider := range providers {
		if provider.Owner == organization.Name {
			return provider, nil
		}
	}

	provider, err := GetDefaultEmailProvider()
	if err != nil || provider != nil {
		return provider, err
	}
	if len(providers) > 0 {
		return providers[0], nil
	}
	return nil, nil
}

// SendOrganizationEmail sends an email with the organization's email provider, in the name of the organization
func SendOrganizationEmail(organization *Organization, title string, content string, dest []string) error {
	if len(dest) == 0 {
		return nil
	}

	provider, err := GetOrganizationEmailProvider(organization)
	if err != nil {
		return err
	}
	if provider == nil {
		return fmt.Errorf("the organization: %s has no email provider", organization.Name)
	}

	sender := organization.DisplayName
	if sender == "" {
		sender = organization.Name
	}
	return SendEmail(provider, title, content, dest, sender)
}

var (
	ipDbOnce   sync.Once
	ipDbLoaded bool
)

func loadIpDb() {
	ipDbOnce.Do(func() {
		if file := conf.GetConfigString("ipDbFile"); file != "" {
			if err := ip.Init(file); err != nil {
				logs.Warn("the IP database: %s is not loaded: %s", file, err.Error())
			} else {
				ipDbLoaded = true
			}
		}
	})
}

// IsIpDbLoaded tells whether the IP database configured by "ipDbFile" is loaded
func IsIpDbLoaded() bool {
	loadIpDb()
	return ipDbLoaded
}

// GetIpLocation names the place of an IP: the city from the IP database configured by "ipDbFile" (the ipip.net
// 17mon format), otherwise the IPv4 /16 or IPv6 /48 network, which changes with the place in most cases
func GetIpLocation(clientIp string) string {
	loadIpDb()

	parsed := net.ParseIP(clientIp)
	if parsed == nil {
		return ""
	}
	if parsed.IsLoopback() || parsed.IsPrivate() {
		return "intranet"
	}

	if ipDbLoaded && parsed.To4() != nil {
		if info, err := ip.Find(clientIp); err == nil && info.Country != "" {
			return strings.Trim(strings.Join([]string{info.Country, info.Region, info.City}, " "), " ")
		}
	}

	if v4 := parsed.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.0.0/16", v4[0], v4[1])
	}
	return (&net.IPNet{IP: parsed.Mask(net.CIDRMask(48, 128)), Mask: net.CIDRMask(48, 128)}).String()
}

// GetIpDbCountry names the country of an IP from the IP database, "" without the database
func GetIpDbCountry(clientIp string) string {
	loadIpDb()
	parsed := net.ParseIP(clientIp)
	if !ipDbLoaded || parsed == nil || parsed.To4() == nil || parsed.IsLoopback() || parsed.IsPrivate() {
		return ""
	}
	if info, err := ip.Find(clientIp); err == nil && info.Country != "" && info.Country != ip.Null {
		return info.Country
	}
	return ""
}

// IsIpInNetworks tells whether the IP is one of the IPs or in one of the CIDR ranges
func IsIpInNetworks(clientIp string, networks []string) bool {
	parsed := net.ParseIP(clientIp)
	if parsed == nil {
		return false
	}
	for _, network := range networks {
		network = strings.TrimSpace(network)
		if network == "" {
			continue
		}
		if !strings.Contains(network, "/") {
			if other := net.ParseIP(network); other != nil && other.Equal(parsed) {
				return true
			}
			continue
		}
		if _, ipNet, err := net.ParseCIDR(network); err == nil && ipNet.Contains(parsed) {
			return true
		}
	}
	return false
}
