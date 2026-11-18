// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/object"
)

// FormFillSite is an application whose sign-in form the extension fills for the current user, without the passwords
type FormFillSite struct {
	Application      string   `json:"application"`
	DisplayName      string   `json:"displayName"`
	Logo             string   `json:"logo"`
	Urls             []string `json:"urls"`
	UsernameSelector string   `json:"usernameSelector"`
	PasswordSelector string   `json:"passwordSelector"`
	SubmitSelector   string   `json:"submitSelector"`
	AutoSubmit       bool     `json:"autoSubmit"`
	AllowSave        bool     `json:"allowSave"`
	// the desktop programs, which have no secrets in them either
	Programs             []*object.FormFillProgram `json:"programs"`
	Captcha              bool                      `json:"captcha"`
	CaptchaSelector      string                    `json:"captchaSelector"`
	CaptchaImageSelector string                    `json:"captchaImageSelector"`
	CaptchaCharset       string                    `json:"captchaCharset"`
	// the usernames of the user's managed accounts of the application
	Accounts []string `json:"accounts"`
}

type FormFillAccount struct {
	Application string `json:"application"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

var formFillSchemeRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)

var formFillCaptchaCharsets = map[string]bool{"": true, "Digits": true, "Alphanumeric": true, "Any": true}

// splitFormFillUrl splits a sign-in page pattern or a page URL into its scheme (may be empty), host with port and the
// rest of it
func splitFormFillUrl(value string) (string, string, string) {
	value = strings.TrimSpace(value)
	scheme := ""
	if match := formFillSchemeRegex.FindString(value); match != "" {
		scheme = strings.ToLower(strings.TrimSuffix(match, "://"))
		value = value[len(match):]
	}
	host, rest := value, "/"
	if index := strings.IndexAny(value, "/?#"); index >= 0 {
		host, rest = value[:index], value[index:]
		if !strings.HasPrefix(rest, "/") {
			rest = "/" + rest
		}
	}
	return scheme, strings.ToLower(host), rest
}

func matchFormFillHost(pattern string, host string) bool {
	if pattern == host {
		return true
	}
	// "*.example.com" matches the subdomains of example.com
	if strings.HasPrefix(pattern, "*.") {
		return strings.HasSuffix(host, pattern[1:])
	}
	return false
}

// MatchFormFillUrl tells whether a page is one of the sign-in pages of a pattern. The host has to be the same (or a
// subdomain for "*."), so that a page of another site never gets the password. A path without "*" is a prefix, and
// the part after "#" is compared only when the pattern has it.
func MatchFormFillUrl(pattern string, pageUrl string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	patternScheme, patternHost, patternPath := splitFormFillUrl(pattern)
	pageScheme, pageHost, pagePath := splitFormFillUrl(pageUrl)
	if pageScheme != "http" && pageScheme != "https" {
		return false
	}
	if patternScheme != "" && patternScheme != pageScheme {
		return false
	}
	if !matchFormFillHost(patternHost, pageHost) {
		return false
	}
	// the fragment only counts when the pattern has one, like the "#/login" route of a single-page application
	if index := strings.Index(pagePath, "#"); index >= 0 && !strings.Contains(patternPath, "#") {
		pagePath = pagePath[:index]
	}
	if !strings.Contains(patternPath, "*") {
		return strings.HasPrefix(pagePath, patternPath)
	}
	expression := "^" + strings.ReplaceAll(regexp.QuoteMeta(patternPath), `\*`, ".*") + "$"
	matched, err := regexp.MatchString(expression, pagePath)
	return err == nil && matched
}

// ValidateFormFill checks the sign-in pages of an application's form autofill
func ValidateFormFill(formFill *object.FormFill, lang string) error {
	if !formFill.IsOn() {
		return nil
	}
	if len(formFill.Urls) == 0 && len(formFill.Programs) == 0 {
		return errors.New(i18n.Translate(lang, "application:Form autofill needs at least one sign-in page URL or desktop program"))
	}
	for _, program := range formFill.Programs {
		if err := validateFormFillProgram(program, lang); err != nil {
			return err
		}
	}
	if !formFillCaptchaCharsets[formFill.CaptchaCharset] {
		return fmt.Errorf(i18n.Translate(lang, "application:The captcha characters: %s are not supported"), formFill.CaptchaCharset)
	}
	for _, pattern := range formFill.Urls {
		scheme, host, _ := splitFormFillUrl(pattern)
		if scheme != "" && scheme != "http" && scheme != "https" {
			return fmt.Errorf(i18n.Translate(lang, "application:The sign-in page URL: %s is not an http or https URL"), pattern)
		}
		if host == "" || strings.Contains(strings.TrimPrefix(host, "*."), "*") {
			return fmt.Errorf(i18n.Translate(lang, "application:The sign-in page URL: %s must start with a host, where * is only allowed as *. at the beginning"), pattern)
		}
		if _, err := url.Parse("http://" + strings.TrimPrefix(host, "*.")); err != nil {
			return fmt.Errorf(i18n.Translate(lang, "application:The sign-in page URL: %s is invalid: %s"), pattern, err.Error())
		}
	}
	return nil
}

// formFillApplication is an application that the user may be signed into by form autofill, with the form autofill
// it gets: its own settings in the enterprise edition, and its Signin URL with the managed accounts as they are in the
// other editions
type formFillApplication struct {
	application *object.Application
	formFill    *object.FormFill
}

// IsFullFormFill tells whether the form autofill of the enterprise edition is available. The other editions only fill
// the managed accounts into the Signin URL of their applications and submit them, like the first browser extensions.
func IsFullFormFill() bool {
	return edition.Require(edition.Enterprise, edition.FeatureFormFill) == nil
}

func getBasicFormFill(application *object.Application) *object.FormFill {
	scheme, host, _ := splitFormFillUrl(application.SigninUrl)
	if (scheme != "http" && scheme != "https") || host == "" || strings.ContainsAny(host, `*\`) {
		return nil
	}
	return &object.FormFill{Enabled: true, Urls: []string{application.SigninUrl}, AutoSubmit: true}
}

// getFormFillApplications returns the applications of the user's organization that the user may be signed into by
// form autofill
func getFormFillApplications(user *object.User) ([]*formFillApplication, error) {
	applications, err := object.GetOrganizationApplications("admin", user.Owner)
	if err != nil {
		return nil, err
	}

	full := IsFullFormFill()
	res := []*formFillApplication{}
	for _, application := range applications {
		formFill := application.FormFill
		if !full {
			formFill = nil
			if user.HasManagedAccountOf(application.Name) {
				formFill = getBasicFormFill(application)
			}
		}
		if !formFill.IsOn() {
			continue
		}
		allowed, err := object.CheckLoginPermission(user.GetId(), application)
		if err != nil {
			return nil, err
		}
		if allowed {
			res = append(res, &formFillApplication{application: application, formFill: formFill})
		}
	}
	return res, nil
}

func GetFormFillSites(user *object.User) ([]*FormFillSite, error) {
	applications, err := getFormFillApplications(user)
	if err != nil {
		return nil, err
	}

	sites := []*FormFillSite{}
	for _, item := range applications {
		application, formFill := item.application, item.formFill
		site := &FormFillSite{
			Application:          application.Name,
			DisplayName:          application.DisplayName,
			Logo:                 application.Logo,
			Urls:                 formFill.Urls,
			UsernameSelector:     formFill.UsernameSelector,
			PasswordSelector:     formFill.PasswordSelector,
			SubmitSelector:       formFill.SubmitSelector,
			AutoSubmit:           formFill.AutoSubmit,
			AllowSave:            formFill.AllowSave,
			Programs:             formFill.Programs,
			Captcha:              formFill.Captcha,
			CaptchaSelector:      formFill.CaptchaSelector,
			CaptchaImageSelector: formFill.CaptchaImageSelector,
			CaptchaCharset:       formFill.CaptchaCharset,
			Accounts:             []string{},
		}
		if site.Programs == nil {
			site.Programs = []*object.FormFillProgram{}
		}
		for _, account := range user.ManagedAccounts {
			if account.Application == application.Name && account.Username != "" {
				site.Accounts = append(site.Accounts, account.Username)
			}
		}
		if len(site.Accounts) > 0 || site.AllowSave {
			sites = append(sites, site)
		}
	}
	return sites, nil
}

func getFormFillApplication(user *object.User, applicationName string, target *object.FormFillTarget, lang string) (*formFillApplication, error) {
	// desktop programs and saving accounts are only in the enterprise edition
	if target.Program != "" && !IsFullFormFill() {
		return nil, edition.Require(edition.Enterprise, edition.FeatureFormFill)
	}
	applications, err := getFormFillApplications(user)
	if err != nil {
		return nil, err
	}
	for _, item := range applications {
		if item.application.Name != applicationName {
			continue
		}
		if !matchFormFillTarget(item.formFill, target) {
			if target.Program != "" {
				return nil, fmt.Errorf(i18n.Translate(lang, "application:The window of the program: %s is not a sign-in window of the application: %s"), target.String(), applicationName)
			}
			return nil, fmt.Errorf(i18n.Translate(lang, "application:The page: %s is not a sign-in page of the application: %s"), target.Url, applicationName)
		}
		return item, nil
	}
	return nil, fmt.Errorf(i18n.Translate(lang, "application:Form autofill is not available for the application: %s"), applicationName)
}

// GetFormFillAccount hands out the user's managed account of an application, for the sign-in page of the
// application or the sign-in window of its desktop program that the client is on
func GetFormFillAccount(user *object.User, applicationName string, username string, target *object.FormFillTarget, lang string) (*FormFillAccount, error) {
	_, err := getFormFillApplication(user, applicationName, target, lang)
	if err != nil {
		return nil, err
	}

	for _, account := range user.ManagedAccounts {
		if account.Application == applicationName && (username == "" || account.Username == username) {
			return &FormFillAccount{Application: applicationName, Username: account.Username, Password: account.Password}, nil
		}
	}
	return nil, fmt.Errorf(i18n.Translate(lang, "application:You have no managed account of the application: %s"), applicationName)
}

// SaveFormFillAccount keeps the account that the user signed into an application with by hand as a managed account
func SaveFormFillAccount(user *object.User, applicationName string, username string, password string, target *object.FormFillTarget, lang string) error {
	if !IsFullFormFill() {
		return edition.Require(edition.Enterprise, edition.FeatureFormFill)
	}
	if username == "" || password == "" {
		return errors.New(i18n.Translate(lang, "application:The username and the password must not be empty"))
	}
	item, err := getFormFillApplication(user, applicationName, target, lang)
	if err != nil {
		return err
	}
	if !item.formFill.AllowSave {
		return fmt.Errorf(i18n.Translate(lang, "application:The application: %s does not let users save their accounts"), applicationName)
	}

	saved := false
	for i := range user.ManagedAccounts {
		if user.ManagedAccounts[i].Application == applicationName && user.ManagedAccounts[i].Username == username {
			user.ManagedAccounts[i].Password = password
			saved = true
		}
	}
	if !saved {
		user.ManagedAccounts = append(user.ManagedAccounts, object.ManagedAccount{Application: applicationName, Username: username, Password: password})
	}

	_, err = object.UpdateUser(user.GetId(), user, []string{"managedAccounts"}, false)
	return err
}

// keepMaskedAccountSecrets keeps the passwords of the managed accounts that come back masked from the console, matched
// by the account, or by the row when only the username was changed
func keepMaskedAccountSecrets(user *object.User, oldUser *object.User) {
	for i := range user.ManagedAccounts {
		account := &user.ManagedAccounts[i]
		if account.Password != "***" {
			continue
		}
		account.Password = ""
		for _, oldAccount := range oldUser.ManagedAccounts {
			if oldAccount.Application == account.Application && oldAccount.Username == account.Username {
				account.Password = oldAccount.Password
				break
			}
		}
		if account.Password == "" && i < len(oldUser.ManagedAccounts) && oldUser.ManagedAccounts[i].Application == account.Application {
			account.Password = oldUser.ManagedAccounts[i].Password
		}
	}
	for i := range user.MfaAccounts {
		account := &user.MfaAccounts[i]
		if account.SecretKey != "***" {
			continue
		}
		account.SecretKey = ""
		for _, oldAccount := range oldUser.MfaAccounts {
			if oldAccount.AccountName == account.AccountName && oldAccount.Issuer == account.Issuer {
				account.SecretKey = oldAccount.SecretKey
				break
			}
		}
	}
}
