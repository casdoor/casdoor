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
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/ip"
	"github.com/casdoor/casdoor/util"
)

// SigninTrace is a location or a device that a user has signed in from
type SigninTrace struct {
	Owner    string `xorm:"varchar(100) notnull pk" json:"owner"`
	UserName string `xorm:"varchar(100) notnull pk" json:"user"`
	Kind     string `xorm:"varchar(20) notnull pk" json:"kind"`
	TraceKey string `xorm:"varchar(200) notnull pk" json:"traceKey"`

	CreatedTime  string `xorm:"varchar(100)" json:"createdTime"`
	LastSeenTime string `xorm:"varchar(100)" json:"lastSeenTime"`
	Ip           string `xorm:"varchar(100)" json:"ip"`
	UserAgent    string `xorm:"varchar(500)" json:"userAgent"`
	Count        int    `json:"count"`
}

var (
	ipDbOnce   sync.Once
	ipDbLoaded bool
)

// getIpLocation names the place of an IP: the city from the IP database configured by "ipDbFile" (the ipip.net
// 17mon format), otherwise the IPv4 /16 or IPv6 /48 network, which changes with the place in most cases
func getIpLocation(clientIp string) string {
	ipDbOnce.Do(func() {
		if file := conf.GetConfigString("ipDbFile"); file != "" {
			if err := ip.Init(file); err != nil {
				logs.Warn("the IP database: %s is not loaded: %s", file, err.Error())
			} else {
				ipDbLoaded = true
			}
		}
	})

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

func truncateString(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

func getSigninTraces(owner string, user string) ([]*SigninTrace, error) {
	traces := []*SigninTrace{}
	err := ormer.Engine.Find(&traces, &SigninTrace{Owner: owner, UserName: user})
	return traces, err
}

func upsertSigninTrace(user *User, kind string, key string, clientIp string, userAgent string, existing *SigninTrace) error {
	now := util.GetCurrentTime()
	if existing != nil {
		existing.LastSeenTime = now
		existing.Ip = clientIp
		existing.UserAgent = truncateString(userAgent, 500)
		existing.Count++
		_, err := ormer.Engine.Cols("last_seen_time", "ip", "user_agent", "count").
			Update(existing, &SigninTrace{Owner: user.Owner, UserName: user.Name, Kind: kind, TraceKey: key})
		return err
	}

	_, err := ormer.Engine.Insert(&SigninTrace{
		Owner:        user.Owner,
		UserName:     user.Name,
		Kind:         kind,
		TraceKey:     key,
		CreatedTime:  now,
		LastSeenTime: now,
		Ip:           clientIp,
		UserAgent:    truncateString(userAgent, 500),
		Count:        1,
	})
	return err
}

// CheckSigninRisk records where and on which device the user signs in, and emails the user when it is a new place
// or a new device under the organization's security policy. The first sign-in only records.
func CheckSigninRisk(user *User, clientIp string, deviceId string, userAgent string) {
	policy := GetCachedSecurityPolicy(user.Owner)
	if policy == nil || (!policy.AlertOnNewLocation && !policy.AlertOnNewDevice) {
		return
	}

	go func() {
		traces, err := getSigninTraces(user.Owner, user.Name)
		if err != nil {
			logs.Warn("the sign-in traces of the user: %s are not read: %s", user.GetId(), err.Error())
			return
		}

		hasKind := map[string]bool{}
		existing := map[string]*SigninTrace{}
		for _, trace := range traces {
			hasKind[trace.Kind] = true
			existing[trace.Kind+"/"+trace.TraceKey] = trace
		}

		alerts := []string{}
		check := func(enabled bool, kind string, key string, message string) {
			if !enabled || key == "" {
				return
			}
			trace := existing[kind+"/"+key]
			if trace == nil && hasKind[kind] {
				alerts = append(alerts, message)
			}
			if err := upsertSigninTrace(user, kind, key, clientIp, userAgent, trace); err != nil {
				logs.Warn("the sign-in trace of the user: %s is not saved: %s", user.GetId(), err.Error())
			}
		}

		location := getIpLocation(clientIp)
		check(policy.AlertOnNewLocation, "location", location, fmt.Sprintf("New location: %s (IP address %s)", location, clientIp))
		check(policy.AlertOnNewDevice, "device", truncateString(deviceId, 200), fmt.Sprintf("New device: %s", userAgent))

		if len(alerts) == 0 || user.Email == "" {
			return
		}

		organization, err := getOrganization("admin", user.Owner)
		if err != nil || organization == nil {
			return
		}
		title := fmt.Sprintf("New sign-in to your %s account", organization.DisplayName)
		content := fmt.Sprintf("Your account %s signed in at %s.\n\n%s\n\nIf this wasn't you, change your password right away and contact your administrator.",
			user.Name, time.Now().Format("2006-01-02 15:04:05 MST"), strings.Join(alerts, "\n"))
		err = sendOrganizationEmail(organization, title, content, []string{user.Email})
		if err != nil {
			logs.Warn("the sign-in alert to the user: %s is not sent: %s", user.GetId(), err.Error())
		}
	}()
}
