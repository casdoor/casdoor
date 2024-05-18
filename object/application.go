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
	"fmt"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type SigninMethod struct {
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Rule        string `json:"rule"`
}

type SignupItem struct {
	Name        string `json:"name"`
	Visible     bool   `json:"visible"`
	Required    bool   `json:"required"`
	Prompted    bool   `json:"prompted"`
	CustomCss   string `json:"customCss"`
	Label       string `json:"label"`
	Placeholder string `json:"placeholder"`
	Regex       string `json:"regex"`
	Rule        string `json:"rule"`
}

type SigninItem struct {
	Name        string `json:"name"`
	Visible     bool   `json:"visible"`
	Label       string `json:"label"`
	CustomCss   string `json:"customCss"`
	Placeholder string `json:"placeholder"`
	Rule        string `json:"rule"`
	IsCustom    bool   `json:"isCustom"`
}

type SamlItem struct {
	Name       string `json:"name"`
	NameFormat string `json:"nameFormat"`
	Value      string `json:"value"`
}

type Application struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName           string          `xorm:"varchar(100)" json:"displayName"`
	Logo                  string          `xorm:"varchar(200)" json:"logo"`
	HomepageUrl           string          `xorm:"varchar(100)" json:"homepageUrl"`
	Description           string          `xorm:"varchar(100)" json:"description"`
	Organization          string          `xorm:"varchar(100)" json:"organization"`
	Cert                  string          `xorm:"varchar(100)" json:"cert"`
	HeaderHtml            string          `xorm:"mediumtext" json:"headerHtml"`
	EnablePassword        bool            `json:"enablePassword"`
	EnableSignUp          bool            `json:"enableSignUp"`
	EnableSigninSession   bool            `json:"enableSigninSession"`
	EnableAutoSignin      bool            `json:"enableAutoSignin"`
	EnableCodeSignin      bool            `json:"enableCodeSignin"`
	EnableSamlCompress    bool            `json:"enableSamlCompress"`
	EnableSamlC14n10      bool            `json:"enableSamlC14n10"`
	EnableSamlPostBinding bool            `json:"enableSamlPostBinding"`
	EnableWebAuthn        bool            `json:"enableWebAuthn"`
	EnableLinkWithEmail   bool            `json:"enableLinkWithEmail"`
	OrgChoiceMode         string          `json:"orgChoiceMode"`
	SamlReplyUrl          string          `xorm:"varchar(100)" json:"samlReplyUrl"`
	Providers             []*ProviderItem `xorm:"mediumtext" json:"providers"`
	SigninMethods         []*SigninMethod `xorm:"varchar(2000)" json:"signinMethods"`
	SignupItems           []*SignupItem   `xorm:"varchar(2000)" json:"signupItems"`
	SigninItems           []*SigninItem   `xorm:"mediumtext" json:"signinItems"`
	GrantTypes            []string        `xorm:"varchar(1000)" json:"grantTypes"`
	OrganizationObj       *Organization   `xorm:"-" json:"organizationObj"`
	CertPublicKey         string          `xorm:"-" json:"certPublicKey"`
	Tags                  []string        `xorm:"mediumtext" json:"tags"`
	SamlAttributes        []*SamlItem     `xorm:"varchar(1000)" json:"samlAttributes"`

	ClientId             string     `xorm:"varchar(100)" json:"clientId"`
	ClientSecret         string     `xorm:"varchar(100)" json:"clientSecret"`
	RedirectUris         []string   `xorm:"varchar(1000)" json:"redirectUris"`
	TokenFormat          string     `xorm:"varchar(100)" json:"tokenFormat"`
	TokenFields          []string   `xorm:"varchar(1000)" json:"tokenFields"`
	ExpireInHours        int        `json:"expireInHours"`
	RefreshExpireInHours int        `json:"refreshExpireInHours"`
	SignupUrl            string     `xorm:"varchar(200)" json:"signupUrl"`
	SigninUrl            string     `xorm:"varchar(200)" json:"signinUrl"`
	ForgetUrl            string     `xorm:"varchar(200)" json:"forgetUrl"`
	AffiliationUrl       string     `xorm:"varchar(100)" json:"affiliationUrl"`
	TermsOfUse           string     `xorm:"varchar(100)" json:"termsOfUse"`
	SignupHtml           string     `xorm:"mediumtext" json:"signupHtml"`
	SigninHtml           string     `xorm:"mediumtext" json:"signinHtml"`
	ThemeData            *ThemeData `xorm:"json" json:"themeData"`
	FooterHtml           string     `xorm:"mediumtext" json:"footerHtml"`
	FormCss              string     `xorm:"text" json:"formCss"`
	FormCssMobile        string     `xorm:"text" json:"formCssMobile"`
	FormOffset           int        `json:"formOffset"`
	FormSideHtml         string     `xorm:"mediumtext" json:"formSideHtml"`
	FormBackgroundUrl    string     `xorm:"varchar(200)" json:"formBackgroundUrl"`

	FailedSigninLimit      int `json:"failedSigninLimit"`
	FailedSigninFrozenTime int `json:"failedSigninFrozenTime"`
}

func GetApplicationCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Application{})
}

func GetOrganizationApplicationCount(owner, Organization, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Application{Organization: Organization})
}

func GetApplications(owner string) ([]*Application, error) {
	applications := []*Application{}
	err := ormer.Engine.Desc("created_time").Find(&applications, &Application{Owner: owner})
	if err != nil {
		return applications, err
	}

	return applications, nil
}

func GetOrganizationApplications(owner string, organization string) ([]*Application, error) {
	applications := []*Application{}
	err := ormer.Engine.Desc("created_time").Find(&applications, &Application{Organization: organization})
	if err != nil {
		return applications, err
	}

	return applications, nil
}

func GetPaginationApplications(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Application, error) {
	var applications []*Application
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&applications)
	if err != nil {
		return applications, err
	}

	return applications, nil
}

func GetPaginationOrganizationApplications(owner, organization string, offset, limit int, field, value, sortField, sortOrder string) ([]*Application, error) {
	applications := []*Application{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&applications, &Application{Organization: organization})
	if err != nil {
		return applications, err
	}

	return applications, nil
}

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

func getApplication(owner string, name string) (*Application, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	application := Application{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&application)
	if err != nil {
		return nil, err
	}

	if existed {
		err = extendApplicationWithProviders(&application)
		if err != nil {
			return nil, err
		}

		err = extendApplicationWithOrg(&application)
		if err != nil {
			return nil, err
		}

		err = extendApplicationWithSigninMethods(&application)
		if err != nil {
			return nil, err
		}
		err = extendApplicationWithSigninItems(&application)
		if err != nil {
			return nil, err
		}

		return &application, nil
	} else {
		return nil, nil
	}
}

func GetApplicationByOrganizationName(organization string) (*Application, error) {
	application := Application{}
	existed, err := ormer.Engine.Where("organization=?", organization).Get(&application)
	if err != nil {
		return nil, nil
	}

	if existed {
		err = extendApplicationWithProviders(&application)
		if err != nil {
			return nil, err
		}

		err = extendApplicationWithOrg(&application)
		if err != nil {
			return nil, err
		}

		err = extendApplicationWithSigninMethods(&application)
		if err != nil {
			return nil, err
		}

		err = extendApplicationWithSigninItems(&application)
		if err != nil {
			return nil, err
		}

		return &application, nil
	} else {
		return nil, nil
	}
}

func GetApplicationByUser(user *User) (*Application, error) {
	if user.SignupApplication != "" {
		return getApplication("admin", user.SignupApplication)
	} else {
		return GetApplicationByOrganizationName(user.Owner)
	}
}

func GetApplicationByUserId(userId string) (application *Application, err error) {
	_, name := util.GetOwnerAndNameFromId(userId)
	if IsAppUser(userId) {
		application, err = getApplication("admin", name)
		return
	}

	user, err := GetUser(userId)
	if err != nil {
		return nil, err
	}
	application, err = GetApplicationByUser(user)
	return
}

func GetApplicationByClientId(clientId string) (*Application, error) {
	application := Application{}
	existed, err := ormer.Engine.Where("client_id=?", clientId).Get(&application)
	if err != nil {
		return nil, err
	}

	if existed {
		err = extendApplicationWithProviders(&application)
		if err != nil {
			return nil, err
		}

		err = extendApplicationWithOrg(&application)
		if err != nil {
			return nil, err
		}

		err = extendApplicationWithSigninMethods(&application)
		if err != nil {
			return nil, err
		}

		err = extendApplicationWithSigninItems(&application)
		if err != nil {
			return nil, err
		}

		return &application, nil
	} else {
		return nil, nil
	}
}

func GetApplication(id string) (*Application, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getApplication(owner, name)
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
	application.EnableWebAuthn = false
	application.EnableLinkWithEmail = false
	application.SamlReplyUrl = "***"

	providerItems := []*ProviderItem{}
	for _, providerItem := range application.Providers {
		if providerItem.Provider != nil && (providerItem.Provider.Category == "OAuth" || providerItem.Provider.Category == "Web3" || providerItem.Provider.Category == "Captcha") {
			providerItems = append(providerItems, providerItem)
		}
	}
	application.Providers = providerItems

	application.GrantTypes = nil
	application.Tags = nil
	application.RedirectUris = nil
	application.TokenFormat = "***"
	application.TokenFields = nil
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
		return nil, fmt.Errorf(i18n.Translate(lang, "auth:Unauthorized operation"))
	}

	if isUserIdGlobalAdmin(userId) {
		return applications, nil
	}

	user, err := GetUser(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf(i18n.Translate(lang, "auth:Unauthorized operation"))
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

func UpdateApplication(id string, application *Application) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	oldApplication, err := getApplication(owner, name)
	if oldApplication == nil {
		return false, err
	}

	if name == "app-built-in" {
		application.Name = name
	}

	if name != application.Name {
		err = applicationChangeTrigger(name, application.Name)
		if err != nil {
			return false, err
		}
	}

	applicationByClientId, err := GetApplicationByClientId(application.ClientId)
	if err != nil {
		return false, err
	}

	if oldApplication.ClientId != application.ClientId && applicationByClientId != nil {
		return false, err
	}

	for _, providerItem := range application.Providers {
		providerItem.Provider = nil
	}

	session := ormer.Engine.ID(core.PK{owner, name}).AllCols()
	if application.ClientSecret == "***" {
		session.Omit("client_secret")
	}
	affected, err := session.Update(application)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddApplication(application *Application) (bool, error) {
	if application.Owner == "" {
		application.Owner = "admin"
	}
	if application.Organization == "" {
		application.Organization = "built-in"
	}
	if application.ClientId == "" {
		application.ClientId = util.GenerateClientId()
	}
	if application.ClientSecret == "" {
		application.ClientSecret = util.GenerateClientSecret()
	}

	app, err := GetApplicationByClientId(application.ClientId)
	if err != nil {
		return false, err
	}

	if app != nil {
		return false, nil
	}

	for _, providerItem := range application.Providers {
		providerItem.Provider = nil
	}

	affected, err := ormer.Engine.Insert(application)
	if err != nil {
		return false, nil
	}

	return affected != 0, nil
}

func deleteApplication(application *Application) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{application.Owner, application.Name}).Delete(&Application{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteApplication(application *Application) (bool, error) {
	if application.Name == "app-built-in" {
		return false, nil
	}

	return deleteApplication(application)
}

func (application *Application) GetId() string {
	return fmt.Sprintf("%s/%s", application.Owner, application.Name)
}

func (application *Application) IsRedirectUriValid(redirectUri string) bool {
	redirectUris := append([]string{"http://localhost:", "https://localhost:", "http://127.0.0.1:", "http://casdoor-app", ".chromiumapp.org"}, application.RedirectUris...)
	for _, targetUri := range redirectUris {
		targetUriRegex := regexp.MustCompile(targetUri)
		if targetUriRegex.MatchString(redirectUri) || strings.Contains(redirectUri, targetUri) {
			return true
		}
	}
	return false
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
