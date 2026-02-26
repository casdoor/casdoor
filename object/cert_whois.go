// Copyright 2023 The casbin Authors. All Rights Reserved.
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
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

func getDomainExpireTime(domainName string) (string, error) {
	domainName, err := util.GetBaseDomain(domainName)
	if err != nil {
		return "", err
	}

	server := ""
	if strings.HasSuffix(domainName, ".com") || strings.HasSuffix(domainName, ".net") {
		server = "whois.verisign-grs.com"
	} else if strings.HasSuffix(domainName, ".org") {
		server = "whois.pir.org"
	} else if strings.HasSuffix(domainName, ".io") {
		server = "whois.nic.io"
	} else if strings.HasSuffix(domainName, ".co") {
		server = "whois.nic.co"
	} else if strings.HasSuffix(domainName, ".cn") {
		server = "whois.cnnic.cn"
	} else if strings.HasSuffix(domainName, ".run") {
		server = "whois.nic.run"
	} else {
		server = "grs-whois.hichina.com" // com, net, cc, tv
	}

	client := whois.NewClient()
	//if server != "whois.cnnic.cn" && server != "grs-whois.hichina.com" {
	//	dialer := proxy.GetProxyDialer()
	//	if dialer != nil {
	//		client.SetDialer(dialer)
	//	}
	//}

	data, err := client.Whois(domainName, server)
	if err != nil {
		if !strings.HasSuffix(domainName, ".run") || data == "" {
			return "", err
		}
	}

	whoisInfo, err := whoisparser.Parse(data)
	if err != nil {
		return "", err
	}

	res := whoisInfo.Domain.ExpirationDateInTime.Local().Format(time.RFC3339)
	return res, nil
}

func GetDomainExpireTime(domainName string) (string, error) {
	return getDomainExpireTime(domainName)
}
