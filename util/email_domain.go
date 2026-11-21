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

package util

import (
	_ "embed"
	"net/mail"
	"strings"
	"sync"
)

// from https://github.com/disposable-email-domains/disposable-email-domains (CC0)
//
//go:embed disposable_email_domains.txt
var disposableEmailDomainsText string

var (
	disposableEmailDomains     map[string]struct{}
	disposableEmailDomainsOnce sync.Once
)

func GetEmailDomain(email string) string {
	if address, err := mail.ParseAddress(email); err == nil {
		email = address.Address
	}

	index := strings.LastIndex(email, "@")
	if index < 0 {
		return ""
	}
	return normalizeEmailDomain(email[index+1:])
}

func normalizeEmailDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimPrefix(domain, "@")
	domain = strings.TrimPrefix(domain, "*")
	return strings.Trim(domain, ".")
}

// a domain matches an entry when it is the entry or one of its subdomains
func isEmailDomainMatched(domain string, entry string) bool {
	return entry != "" && (domain == entry || strings.HasSuffix(domain, "."+entry))
}

func IsEmailDomainInList(domain string, list []string) bool {
	domain = normalizeEmailDomain(domain)
	if domain == "" {
		return false
	}

	for _, entry := range list {
		if isEmailDomainMatched(domain, normalizeEmailDomain(entry)) {
			return true
		}
	}
	return false
}

func IsDisposableEmailDomain(domain string) bool {
	disposableEmailDomainsOnce.Do(func() {
		disposableEmailDomains = map[string]struct{}{}
		for _, line := range strings.Split(disposableEmailDomainsText, "\n") {
			if line = normalizeEmailDomain(line); line != "" && !strings.HasPrefix(line, "#") {
				disposableEmailDomains[line] = struct{}{}
			}
		}
	})

	domain = normalizeEmailDomain(domain)
	for domain != "" {
		if _, ok := disposableEmailDomains[domain]; ok {
			return true
		}

		index := strings.Index(domain, ".")
		if index < 0 {
			break
		}
		domain = domain[index+1:]
	}
	return false
}
