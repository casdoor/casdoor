// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"net/url"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
)

func getProviderMap(owner string) (m map[string]*Provider, err error) {
	providers, err := GetProviders(owner)
	if err != nil {
		return nil, err
	}

	m = map[string]*Provider{}
	for _, provider := range providers {
		m[provider.Name] = GetMaskedProvider(provider, true)
	}

	return m, err
}

func extendApplicationWithProviders(application *Application) (err error) {
	m, err := getProviderMap(application.Organization)
	if err != nil {
		return err
	}

	for _, providerItem := range application.Providers {
		if provider, ok := m[providerItem.Name]; ok {
			providerItem.Provider = provider
		}
	}

	return
}

func extendApplicationWithOrg(application *Application) (err error) {
	organization, err := getOrganization(application.Owner, application.Organization)
	application.OrganizationObj = organization
	return
}

func extendApplicationWithSigninItems(application *Application) (err error) {
	if len(application.SigninItems) == 0 {
		signinItem := &SigninItem{
			Name:        "Back button",
			Visible:     true,
			CustomCss:   ".back-button {\n      top: 65px;\n      left: 15px;\n      position: absolute;\n}\n.back-inner-button{}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Languages",
			Visible:     true,
			CustomCss:   ".login-languages {\n    top: 55px;\n    right: 5px;\n    position: absolute;\n}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Logo",
			Visible:     true,
			CustomCss:   ".login-logo-box {}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Signin methods",
			Visible:     true,
			CustomCss:   ".signin-methods {}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Username",
			Visible:     true,
			CustomCss:   ".login-username {}\n.login-username-input{}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Password",
			Visible:     true,
			CustomCss:   ".login-password {}\n.login-password-input{}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Verification code",
			Visible:     true,
			CustomCss:   ".verification-code {}\n.verification-code-input{}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Agreement",
			Visible:     true,
			CustomCss:   ".login-agreement {}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Forgot password?",
			Visible:     true,
			CustomCss:   ".login-forget-password {\n    display: inline-flex;\n    justify-content: space-between;\n    width: 320px;\n    margin-bottom: 25px;\n}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Login button",
			Visible:     true,
			CustomCss:   ".login-button-box {\n    margin-bottom: 5px;\n}\n.login-button {\n    width: 100%;\n}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Signup link",
			Visible:     true,
			CustomCss:   ".login-signup-link {\n    margin-bottom: 24px;\n    display: flex;\n    justify-content: end;\n}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
		signinItem = &SigninItem{
			Name:        "Providers",
			Visible:     true,
			CustomCss:   ".provider-img {\n      width: 30px;\n      margin: 5px;\n}\n.provider-big-img {\n      margin-bottom: 10px;\n}",
			Placeholder: "",
			Rule:        "None",
		}
		application.SigninItems = append(application.SigninItems, signinItem)
	}
	for idx, item := range application.SigninItems {
		if item.Label != "" && item.CustomCss == "" {
			application.SigninItems[idx].CustomCss = item.Label
			application.SigninItems[idx].Label = ""
		}
	}
	return
}

func extendApplicationWithSigninMethods(application *Application) (err error) {
	if len(application.SigninMethods) == 0 {
		if application.EnablePassword {
			signinMethod := &SigninMethod{Name: "Password", DisplayName: "Password", Rule: "All"}
			application.SigninMethods = append(application.SigninMethods, signinMethod)
		}
		if application.EnableCodeSignin {
			signinMethod := &SigninMethod{Name: "Verification code", DisplayName: "Verification code", Rule: "All"}
			application.SigninMethods = append(application.SigninMethods, signinMethod)
		}
		if application.EnableWebAuthn {
			signinMethod := &SigninMethod{Name: "WebAuthn", DisplayName: "WebAuthn", Rule: "None"}
			application.SigninMethods = append(application.SigninMethods, signinMethod)
		}

		signinMethod := &SigninMethod{Name: "Face ID", DisplayName: "Face ID", Rule: "None"}
		application.SigninMethods = append(application.SigninMethods, signinMethod)
	}

	if len(application.SigninMethods) == 0 {
		signinMethod := &SigninMethod{Name: "Password", DisplayName: "Password", Rule: "All"}
		application.SigninMethods = append(application.SigninMethods, signinMethod)
	}

	return
}

func extendApplicationWithSignupItems(application *Application) (err error) {
	if len(application.SignupItems) == 0 {
		application.SignupItems = []*SignupItem{
			{Name: "ID", Visible: false, Required: true, Prompted: false, Rule: "Random"},
			{Name: "Username", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Display name", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Password", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Confirm password", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Email", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Phone", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Agreement", Visible: true, Required: true, Prompted: false, Rule: "None"},
		}
	}
	return
}

func GetMaskedApplication(application *Application, userId string) *Application {
	if application == nil {
		return nil
	}

	if application.TokenFields == nil {
		application.TokenFields = []string{}
	}

	if application.FailedSigninLimit == 0 {
		application.FailedSigninLimit = DefaultFailedSigninLimit
	}
	if application.FailedSigninFrozenTime == 0 {
		application.FailedSigninFrozenTime = DefaultFailedSigninFrozenTime
	}

	isOrgUser := false
	if userId != "" {
		if isUserIdGlobalAdmin(userId) {
			return application
		}

		user, err := GetUser(userId)
		if err != nil {
			panic(err)
		}
		if user != nil {
			if user.IsApplicationAdmin(application) {
				return application
			}

			if user.Owner == application.Organization {
				isOrgUser = true
			}
		}
	}

	application.ClientSecret = "***"
	application.Cert = "***"
	application.EnablePassword = false
	application.EnableSigninSession = false
	application.EnableCodeSignin = false
	application.EnableSamlCompress = false
	application.EnableSamlC14n10 = false
	application.EnableSamlPostBinding = false
	application.DisableSamlAttributes = false
	application.EnableWebAuthn = false
	application.EnableLinkWithEmail = false
	application.SamlReplyUrl = "***"

	providerItems := []*ProviderItem{}
	for _, providerItem := range application.Providers {
		if providerItem.Provider != nil && (providerItem.Provider.Category == "OAuth" || providerItem.Provider.Category == "Web3" || providerItem.Provider.Category == "Captcha" || providerItem.Provider.Category == "SAML" || providerItem.Provider.Category == "Face ID") {
			providerItems = append(providerItems, providerItem)
		}
	}
	application.Providers = providerItems

	application.GrantTypes = []string{}
	application.RedirectUris = []string{}
	application.TokenFormat = "***"
	application.TokenFields = []string{}
	application.ExpireInHours = -1
	application.RefreshExpireInHours = -1
	application.FailedSigninLimit = -1
	application.FailedSigninFrozenTime = -1

	if application.OrganizationObj != nil {
		application.OrganizationObj.MasterPassword = "***"
		application.OrganizationObj.DefaultPassword = "***"
		application.OrganizationObj.MasterVerificationCode = "***"
		application.OrganizationObj.PasswordType = "***"
		application.OrganizationObj.PasswordSalt = "***"
		application.OrganizationObj.InitScore = -1
		application.OrganizationObj.EnableSoftDeletion = false

		if !isOrgUser {
			application.OrganizationObj.MfaItems = nil
			if !application.OrganizationObj.IsProfilePublic {
				application.OrganizationObj.AccountItems = nil
			}
		}
	}

	return application
}

func GetMaskedApplications(applications []*Application, userId string) []*Application {
	if isUserIdGlobalAdmin(userId) {
		return applications
	}

	for _, application := range applications {
		application = GetMaskedApplication(application, userId)
	}
	return applications
}

func GetAllowedApplications(applications []*Application, userId string, lang string) ([]*Application, error) {
	if userId == "" {
		return nil, errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
	}

	if isUserIdGlobalAdmin(userId) {
		return applications, nil
	}

	user, err := GetUser(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
	}

	if user.IsAdmin {
		return applications, nil
	}

	res := []*Application{}
	for _, application := range applications {
		var allowed bool
		allowed, err = CheckLoginPermission(userId, application)
		if err != nil {
			return nil, err
		}

		if allowed {
			res = append(res, application)
		}
	}
	return res, nil
}

func checkMultipleCaptchaProviders(application *Application, lang string) error {
	var captchaProviders []string
	for _, providerItem := range application.Providers {
		if providerItem.Provider != nil && providerItem.Provider.Category == "Captcha" {
			captchaProviders = append(captchaProviders, providerItem.Name)
		}
	}

	if len(captchaProviders) > 1 {
		return fmt.Errorf(i18n.Translate(lang, "general:Multiple captcha providers are not allowed in the same application: %s"), strings.Join(captchaProviders, ", "))
	}

	return nil
}

func (application *Application) GetId() string {
	return fmt.Sprintf("%s/%s", application.Owner, application.Name)
}

func (application *Application) IsRedirectUriValid(redirectUri string) bool {
	isValid, err := util.IsValidOrigin(redirectUri)
	if err != nil {
		panic(err)
	}
	if isValid {
		return true
	}

	redirectUriObj, err := url.Parse(redirectUri)
	if err != nil || redirectUriObj.Host == "" {
		return false
	}

	for _, targetUri := range application.RedirectUris {
		if targetUri == "" {
			continue
		}

		if redirectUri == targetUri {
			return true
		}

		// URL-based comparison: scheme/host/port/path must each match.
		// The host may be the configured host or any subdomain of it.
		// This prevents "double-URL" attacks like https://evil.com/?x=https://legit.example.com/cb
		targetUriObj, err := url.Parse(targetUri)
		if err == nil && targetUriObj.Host != "" {
			if redirectUriMatchesTarget(redirectUriObj, targetUriObj) {
				return true
			}
			// The configured URI is a valid URL; skip regex fallback to avoid false positives.
			continue
		}

		// Fall back to anchored full-string regex for wildcard patterns
		// (e.g. "https://.*\.example\.com/callback") that are not valid URLs.
		// Anchoring with ^...$ prevents partial-match bypasses.
		anchoredPattern := "^(?:" + targetUri + ")$"
		targetUriRegex, err := regexp.Compile(anchoredPattern)
		if err == nil && targetUriRegex.MatchString(redirectUri) {
			return true
		}
	}
	return false
}

// redirectUriMatchesTarget checks if redirectUri matches targetUri with subdomain support.
// Scheme, port, and path must match exactly. The host may be the target host or any subdomain of it.
func redirectUriMatchesTarget(redirectUri, targetUri *url.URL) bool {
	if redirectUri.Scheme != targetUri.Scheme {
		return false
	}
	if redirectUri.Port() != targetUri.Port() {
		return false
	}
	redirectHost := redirectUri.Hostname()
	targetHost := targetUri.Hostname()
	// Allow exact host match or subdomain (e.g. aaa.example.com for target example.com).
	// The "."+targetHost prefix prevents evilexample.com from matching example.com.
	if redirectHost != targetHost && !strings.HasSuffix(redirectHost, "."+targetHost) {
		return false
	}
	if redirectUri.Path != targetUri.Path {
		return false
	}
	return true
}

func (application *Application) IsPasswordEnabled() bool {
	if len(application.SigninMethods) == 0 {
		return application.EnablePassword
	} else {
		for _, signinMethod := range application.SigninMethods {
			if signinMethod.Name == "Password" {
				return true
			}
		}
		return false
	}
}

func (application *Application) IsPasswordWithLdapEnabled() bool {
	if len(application.SigninMethods) == 0 {
		return application.EnablePassword
	} else {
		for _, signinMethod := range application.SigninMethods {
			if signinMethod.Name == "Password" && signinMethod.Rule == "All" {
				return true
			}
		}
		return false
	}
}

func (application *Application) IsCodeSigninViaEmailEnabled() bool {
	if len(application.SigninMethods) == 0 {
		return application.EnableCodeSignin
	} else {
		for _, signinMethod := range application.SigninMethods {
			if signinMethod.Name == "Verification code" && signinMethod.Rule != "Phone only" {
				return true
			}
		}
		return false
	}
}

func (application *Application) IsCodeSigninViaSmsEnabled() bool {
	if len(application.SigninMethods) == 0 {
		return application.EnableCodeSignin
	} else {
		for _, signinMethod := range application.SigninMethods {
			if signinMethod.Name == "Verification code" && signinMethod.Rule != "Email only" {
				return true
			}
		}
		return false
	}
}

func (application *Application) IsLdapEnabled() bool {
	if len(application.SigninMethods) > 0 {
		for _, signinMethod := range application.SigninMethods {
			if signinMethod.Name == "LDAP" {
				return true
			}
		}
	}
	return false
}

func (application *Application) IsFaceIdEnabled() bool {
	if len(application.SigninMethods) > 0 {
		for _, signinMethod := range application.SigninMethods {
			if signinMethod.Name == "Face ID" {
				return true
			}
		}
	}
	return false
}

func IsOriginAllowed(origin string) (bool, error) {
	applications, err := GetApplications("")
	if err != nil {
		return false, err
	}

	for _, application := range applications {
		if application.IsRedirectUriValid(origin) {
			return true, nil
		}
	}
	return false, nil
}

func getApplicationMap(organization string) (map[string]*Application, error) {
	applicationMap := make(map[string]*Application)
	applications, err := GetOrganizationApplications("admin", organization)
	if err != nil {
		return applicationMap, err
	}

	for _, application := range applications {
		applicationMap[application.Name] = application
	}

	return applicationMap, nil
}

func ExtendManagedAccountsWithUser(user *User) (*User, error) {
	if user.ManagedAccounts == nil || len(user.ManagedAccounts) == 0 {
		return user, nil
	}

	applicationMap, err := getApplicationMap(user.Owner)
	if err != nil {
		return user, err
	}

	var managedAccounts []ManagedAccount
	for _, managedAccount := range user.ManagedAccounts {
		application := applicationMap[managedAccount.Application]
		if application != nil {
			managedAccount.SigninUrl = application.SigninUrl
			managedAccounts = append(managedAccounts, managedAccount)
		}
	}
	user.ManagedAccounts = managedAccounts

	return user, nil
}

func applicationChangeTrigger(oldName string, newName string) error {
	session := ormer.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	organization := new(Organization)
	organization.DefaultApplication = newName
	_, err = session.Where("default_application=?", oldName).Update(organization)
	if err != nil {
		return err
	}

	user := new(User)
	user.SignupApplication = newName
	_, err = session.Where("signup_application=?", oldName).Update(user)
	if err != nil {
		return err
	}

	resource := new(Resource)
	resource.Application = newName
	_, err = session.Where("application=?", oldName).Update(resource)
	if err != nil {
		return err
	}

	var permissions []*Permission
	err = ormer.Engine.Find(&permissions)
	if err != nil {
		return err
	}
	for i := 0; i < len(permissions); i++ {
		permissionResoureces := permissions[i].Resources
		for j := 0; j < len(permissionResoureces); j++ {
			if permissionResoureces[j] == oldName {
				permissionResoureces[j] = newName
			}
		}
		permissions[i].Resources = permissionResoureces
		_, err = session.Where("owner=?", permissions[i].Owner).Where("name=?", permissions[i].Name).Update(permissions[i])
		if err != nil {
			return err
		}
	}

	return session.Commit()
}
