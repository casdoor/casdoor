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
	"regexp"
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/casdoor/casdoor/xlsx"
)

// ApplicationUploadResult tells what an import of applications did, row by row for the failed ones
type ApplicationUploadResult struct {
	Added   int      `json:"added"`
	Skipped []string `json:"skipped"`
	Errors  []string `json:"errors"`
}

var (
	applicationUploadLineSeparator = regexp.MustCompile(`[\r\n;]+`)
	applicationUploadListSeparator = regexp.MustCompile(`[\s,;]+`)
)

// normalizeUploadHeader turns a header such as "Client ID#client_id" or "Redirect URIs" into "redirecturis"
func normalizeUploadHeader(header string) string {
	if i := strings.LastIndex(header, "#"); i >= 0 {
		header = header[i+1:]
	}
	header = strings.ToLower(header)
	return strings.NewReplacer(" ", "", "_", "", "-", "").Replace(header)
}

// splitRedirectUris reads one URI a line or separated by ";", and by "," only when every part is a URI, as a
// regular expression such as "\d{1,3}" may have a comma in it
func splitRedirectUris(value string) []string {
	res := []string{}
	for _, line := range applicationUploadLineSeparator.Split(value, -1) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		isList := len(parts) > 1
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if !strings.Contains(part, "://") && !strings.HasPrefix(part, "/") {
				isList = false
			}
		}
		if !isList {
			res = append(res, line)
			continue
		}
		for _, part := range parts {
			res = append(res, strings.TrimSpace(part))
		}
	}
	return res
}

func splitUploadList(value string) []string {
	res := []string{}
	for _, item := range applicationUploadListSeparator.Split(value, -1) {
		if item != "" {
			res = append(res, item)
		}
	}
	return res
}

// newUploadedApplication is a new application with the same defaults as one added on the console
func newUploadedApplication(name string, organization string) *Application {
	return &Application{
		Owner:          "admin",
		Name:           name,
		Organization:   organization,
		CreatedTime:    util.GetCurrentTime(),
		DisplayName:    name,
		Category:       "Default",
		Type:           "All",
		Scopes:         []*ScopeItem{},
		EnablePassword: true,
		EnableSignUp:   true,
		Providers:      []*ProviderItem{{Name: "provider_captcha_default"}},
		SigninMethods: []*SigninMethod{
			{Name: "Password", DisplayName: "Password", Rule: "All"},
			{Name: "Verification code", DisplayName: "Verification code", Rule: "All"},
			{Name: "WebAuthn", DisplayName: "WebAuthn", Rule: "None"},
			{Name: "Face ID", DisplayName: "Face ID", Rule: "None"},
		},
		SignupItems: []*SignupItem{
			{Name: "ID", Visible: false, Required: true, Rule: "Random"},
			{Name: "Username", Visible: true, Required: true, Rule: "None"},
			{Name: "Display name", Visible: true, Required: true, Rule: "None"},
			{Name: "Password", Visible: true, Required: true, Rule: "None"},
			{Name: "Confirm password", Visible: true, Required: true, Rule: "None"},
			{Name: "Email", Visible: true, Required: true, Rule: "Normal"},
			{Name: "Phone", Visible: true, Required: true, Rule: "None"},
			{Name: "Agreement", Visible: true, Required: true, Rule: "None"},
			{Name: "Signup button", Visible: true, Required: true, Rule: "None"},
			{Name: "Providers", Visible: true, Required: true, Rule: "None"},
		},
		GrantTypes:           []string{"authorization_code", "refresh_token"},
		Cert:                 "cert-built-in",
		RedirectUris:         []string{},
		ExpireInHours:        24 * 7,
		RefreshExpireInHours: 24 * 7,
		CookieExpireInHours:  24 * 30,
		FormOffset:           2,
	}
}

// UploadApplications adds the applications of a spreadsheet, keeping the client IDs, client secrets and redirect URIs
// that the systems already use, so that moving them from another SSO only changes the address they sign in at. The
// applications that already exist are skipped. An organization admin can only add applications to its organization.
func UploadApplications(path string, organization string, lang string, checkQuota func(count int) error) (*ApplicationUploadResult, error) {
	table, err := xlsx.ReadXlsxFile(path)
	if err != nil {
		return nil, err
	}
	if len(table) < 2 {
		return nil, fmt.Errorf("the file has no application")
	}

	columns := map[string]int{}
	for i, header := range table[0] {
		columns[normalizeUploadHeader(header)] = i
	}
	if _, ok := columns["name"]; !ok {
		return nil, fmt.Errorf("the file has no \"name\" column")
	}
	cell := func(row []string, names ...string) string {
		for _, name := range names {
			if i, ok := columns[name]; ok && i < len(row) {
				return strings.TrimSpace(row[i])
			}
		}
		return ""
	}

	res := &ApplicationUploadResult{Skipped: []string{}, Errors: []string{}}
	for i, row := range table[1:] {
		line := i + 2
		name := cell(row, "name")
		if name == "" {
			continue
		}

		applicationOrganization := cell(row, "organization", "owner")
		if applicationOrganization == "" {
			applicationOrganization = organization
		}
		if organization != "" && applicationOrganization != organization {
			res.Errors = append(res.Errors, fmt.Sprintf("line %d: %s: only applications of the organization: %s can be added", line, name, organization))
			continue
		}
		if applicationOrganization == "" {
			applicationOrganization = "built-in"
		}

		existing, err := getApplication("admin", name)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			res.Skipped = append(res.Skipped, name)
			continue
		}

		application := newUploadedApplication(name, applicationOrganization)
		if value := cell(row, "displayname"); value != "" {
			application.DisplayName = value
		}
		application.ClientId = cell(row, "clientid")
		application.ClientSecret = cell(row, "clientsecret")
		application.RedirectUris = splitRedirectUris(cell(row, "redirecturis", "redirecturi", "callbackurl"))
		application.HomepageUrl = cell(row, "homepageurl", "homepage")
		application.Description = cell(row, "description")
		application.SamlReplyUrl = cell(row, "samlreplyurl", "acsurl")
		if value := cell(row, "logo"); value != "" {
			application.Logo = value
		}
		if value := cell(row, "cert"); value != "" {
			application.Cert = value
		}
		if value := cell(row, "granttypes"); value != "" {
			application.GrantTypes = splitUploadList(value)
		}
		if value := cell(row, "expireinhours"); value != "" {
			hours, err := strconv.ParseFloat(value, 64)
			if err != nil {
				res.Errors = append(res.Errors, fmt.Sprintf("line %d: %s: invalid expire in hours: %s", line, name, value))
				continue
			}
			application.ExpireInHours = hours
		}

		count, err := GetApplicationCount("", "", "")
		if err != nil {
			return nil, err
		}
		if err = checkQuota(int(count)); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("line %d: %s: %s", line, name, err.Error()))
			break
		}
		if err = CheckApplicationEdition(application); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("line %d: %s: %s", line, name, err.Error()))
			continue
		}
		if _, err = AddApplication(application, lang); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("line %d: %s: %s", line, name, err.Error()))
			continue
		}
		res.Added++
	}
	return res, nil
}
