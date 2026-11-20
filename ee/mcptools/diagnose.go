// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package mcptools

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/util"
)

type loginRecord struct {
	CreatedTime string `json:"createdTime"`
	User        string `json:"user"`
	Method      string `json:"method"`
	RequestUri  string `json:"requestUri"`
	Action      string `json:"action"`
	Object      string `json:"object"`
	Response    string `json:"response"`
	StatusCode  int    `json:"statusCode"`
}

var (
	recordErrorRegex   = regexp.MustCompile(`"?status"?\s*:\s*"error"`)
	recordMsgRegex     = regexp.MustCompile(`"?(?:msg|error_description)"?\s*:\s*"((?:[^"\\]|\\.)*)"`)
	recordOAuthErrRegx = regexp.MustCompile(`"error"\s*:\s*"([a-z_]+)"`)
)

// getMessage is the error of a failed request in the record, or "" when it didn't fail
func (record *loginRecord) getMessage() string {
	isError := record.StatusCode >= 400 || recordErrorRegex.MatchString(record.Response)
	oauthError := recordOAuthErrRegx.FindStringSubmatch(record.Response)
	if !isError && oauthError == nil {
		return ""
	}

	message := ""
	if match := recordMsgRegex.FindStringSubmatch(record.Response); match != nil {
		message = match[1]
	}
	if oauthError != nil {
		message = strings.TrimSpace(oauthError[1] + " " + message)
	}
	if message == "" {
		message = fmt.Sprintf("HTTP %d", record.StatusCode)
	}
	return truncate(message, 300)
}

var loginHints = []struct {
	keywords []string
	hint     string
}{
	{[]string{"redirect_uri", "redirect uri"}, "The app sent a redirect URL that the application doesn't list: add it to redirectUris with update_application, the scheme, host, port and path must match."},
	{[]string{"grant_type", "grant type"}, "The app uses a grant type that the application doesn't allow: add it to grantTypes with update_application."},
	{[]string{"invalid_client", "client_secret", "client secret"}, "The app sends a wrong client ID or secret: copy them again with get_integration_guide."},
	{[]string{"invalid_grant", "authorization code"}, "The authorization code was used twice, expired or was issued to another redirect URL: exchange it once right after the redirect, and check the server clocks."},
	{[]string{"code_verifier", "pkce"}, "The PKCE code verifier doesn't match the challenge: keep the verifier of the same sign-in in the browser until the callback."},
	{[]string{"password or code is incorrect", "password is wrong", "wrong password"}, "Wrong passwords: the user may need to reset it, repeated failures lock the account for a while."},
	{[]string{"user does not exist", "the user:", "doesn't exist"}, "The user isn't in the organization: create the user, or let people sign up on the application."},
	{[]string{"sign up", "signup"}, "Sign-up failed: check that the application allows sign-up and that its sign-up items can be filled."},
	{[]string{"email", "verification code"}, "Verification codes need an Email or SMS provider on the application (or a default one in the settings)."},
	{[]string{"captcha"}, "The captcha failed or its provider isn't set up."},
	{[]string{"forbidden", "disabled"}, "The user or the application is disabled."},
	{[]string{"unauthorized operation"}, "The request isn't allowed for the signed-in user."},
}

func getLoginHint(message string) string {
	lower := strings.ToLower(message)
	for _, item := range loginHints {
		for _, keyword := range item.keywords {
			if strings.Contains(lower, keyword) {
				return item.hint
			}
		}
	}
	return ""
}

func runDiagnoseLogin(t *toolCall) (string, error) {
	applicationId, err := t.requireString("application")
	if err != nil {
		return "", err
	}
	application, err := t.getAdministeredApplication(applicationId)
	if err != nil {
		return "", err
	}

	findings := t.checkApplication(application)
	lines := []string{fmt.Sprintf("Application %s (organization %s)", applicationId, getString(application, "organization")), "", "## Settings"}
	if len(findings) == 0 {
		lines = append(lines, "- No problem found in the settings.")
	}
	lines = append(lines, findings...)

	minutes := t.getInt("minutes", 1440, 60*24*30)
	failures, scanned, err := t.getLoginFailures(application, minutes, t.getString("user"))
	lines = append(lines, "", fmt.Sprintf("## Failed requests in the last %d minutes", minutes))
	if err != nil {
		lines = append(lines, "- The records can't be read: "+err.Error())
	} else if len(failures) == 0 {
		lines = append(lines, fmt.Sprintf("- None among the latest %d records of the organization.", scanned))
	} else {
		lines = append(lines, failures...)
	}
	return strings.Join(lines, "\n"), nil
}

// checkApplication finds the settings that make the sign-in of an application fail
func (t *toolCall) checkApplication(application map[string]interface{}) []string {
	findings := []string{}
	add := func(format string, args ...interface{}) {
		findings = append(findings, "- "+fmt.Sprintf(format, args...))
	}

	redirectUris := toStrings(application["redirectUris"])
	if len(redirectUris) == 0 {
		add("The application has no redirect URLs, every OAuth sign-in fails with \"redirect_uri doesn't exist\": add the app's callback URL to redirectUris.")
	}
	for _, uri := range redirectUris {
		if parsed, err := url.Parse(uri); err == nil && parsed.Scheme == "http" && parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1" {
			add("The redirect URL %s is plain HTTP: browsers may drop the session cookie, use HTTPS in production.", uri)
		}
	}

	grantTypes := toStrings(application["grantTypes"])
	if !util.InSlice(grantTypes, "authorization_code") {
		add("The grant types (%s) leave out authorization_code, which apps signing users in through the Casdoor page use.", strings.Join(grantTypes, ", "))
	}

	providers, _ := application["providers"].([]interface{})
	signinMethods, _ := application["signinMethods"].([]interface{})
	if application["enablePassword"] == false && len(providers) == 0 && len(signinMethods) == 0 {
		add("Password sign-in is off and there is no other sign-in method or provider: nobody can sign in.")
	}

	if getString(application, "certPublicKey") == "" {
		add("The application's cert %q doesn't exist or has no certificate: tokens can't be signed. Pick an existing cert.", getString(application, "cert"))
	}

	organization := getString(application, "organization")
	hasEmailProvider := false
	for _, item := range providers {
		providerItem, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name := getString(providerItem, "name")
		provider, _ := providerItem["provider"].(map[string]interface{})
		if provider == nil {
			provider, _ = t.client.get(kinds["provider"], util.GetId(organization, name))
		}
		if provider == nil {
			provider, _ = t.client.get(kinds["provider"], util.GetId("admin", name))
		}
		if provider == nil {
			add("The provider %s on the application doesn't exist: remove it from the application or add it again.", name)
			continue
		}
		category := getString(provider, "category")
		if category == "Email" {
			hasEmailProvider = true
		}
		if category == "OAuth" && getString(provider, "clientId") == "" {
			add("The provider %s (%s) has no client ID: \"Sign in with %s\" fails.", name, getString(provider, "type"), getString(provider, "type"))
		}
	}

	if application["enableSignUp"] == false {
		add("Sign-up is off: only existing users of the organization %s can sign in.", organization)
	} else if !hasEmailProvider {
		items, _ := application["signupItems"].([]interface{})
		for _, item := range items {
			signupItem, ok := item.(map[string]interface{})
			if ok && getString(signupItem, "name") == "Email" && signupItem["visible"] == true && getString(signupItem, "rule") == "Normal" {
				add("Sign-up asks for an email verification code, but the application has no Email provider: sign-up fails unless a default email provider is set. Add one, or set the Email sign-up item's rule to \"No verification\".")
				break
			}
		}
	}
	return findings
}

// getLoginFailures summarizes the failed requests that concern the application in the latest records of its
// organization, grouped by error
func (t *toolCall) getLoginFailures(application map[string]interface{}, minutes int, user string) ([]string, int, error) {
	query := url.Values{"p": {"1"}, "pageSize": {"500"}, "organizationName": {getString(application, "organization")}}
	resp, err := t.client.call("GET", "/api/get-records", query, nil, (*controllers.ApiController).GetRecords)
	if err != nil {
		return nil, 0, err
	}
	records := []*loginRecord{}
	if err = json.Unmarshal(resp.Data, &records); err != nil {
		return nil, 0, err
	}

	name, clientId := getString(application, "name"), getString(application, "clientId")
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	type group struct {
		message string
		count   int
		last    string
		users   map[string]bool
		path    string
	}
	groups := map[string]*group{}
	for _, record := range records {
		createdTime, err := time.Parse(time.RFC3339, record.CreatedTime)
		if err == nil && createdTime.Before(since) {
			continue
		}
		if user != "" && record.User != user && !strings.HasSuffix(record.User, "/"+user) {
			continue
		}
		// the tool calls of the MCP server name the application in their arguments, they aren't sign-ins
		if strings.HasPrefix(record.RequestUri, "/api/mcp") {
			continue
		}
		text := record.RequestUri + " " + record.Object
		if !strings.Contains(text, name) && (clientId == "" || !strings.Contains(text, clientId)) {
			continue
		}
		message := record.getMessage()
		if message == "" {
			continue
		}

		g := groups[message]
		if g == nil {
			g = &group{message: message, users: map[string]bool{}, last: record.CreatedTime, path: strings.SplitN(record.RequestUri, "?", 2)[0]}
			groups[message] = g
		}
		g.count++
		if record.User != "" {
			g.users[record.User] = true
		}
	}

	list := []*group{}
	for _, g := range groups {
		list = append(list, g)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].count > list[j].count })

	lines := []string{}
	for _, g := range list {
		users := []string{}
		for u := range g.users {
			users = append(users, u)
		}
		sort.Strings(users)
		if len(users) > 5 {
			users = append(users[:5], "...")
		}
		line := fmt.Sprintf("- %d x %q on %s, last at %s", g.count, g.message, g.path, g.last)
		if len(users) > 0 {
			line += ", users: " + strings.Join(users, ", ")
		}
		lines = append(lines, line)
		if hint := getLoginHint(g.message); hint != "" {
			lines = append(lines, "  Fix: "+hint)
		}
	}
	return lines, len(records), nil
}

func toStrings(value interface{}) []string {
	result := []string{}
	values, _ := value.([]interface{})
	for _, item := range values {
		if text, ok := item.(string); ok && text != "" {
			result = append(result, text)
		}
	}
	return result
}
