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

	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type SignupItem struct {
	Name     string `json:"name"`
	Visible  bool   `json:"visible"`
	Required bool   `json:"required"`
	Prompted bool   `json:"prompted"`
	Rule     string `json:"rule"`
}

type Application struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName         string          `xorm:"varchar(100)" json:"displayName"`
	Logo                string          `xorm:"varchar(200)" json:"logo"`
	HomepageUrl         string          `xorm:"varchar(100)" json:"homepageUrl"`
	Description         string          `xorm:"varchar(100)" json:"description"`
	Organization        string          `xorm:"varchar(100)" json:"organization"`
	Cert                string          `xorm:"varchar(100)" json:"cert"`
	EnablePassword      bool            `json:"enablePassword"`
	EnableSignUp        bool            `json:"enableSignUp"`
	EnableSigninSession bool            `json:"enableSigninSession"`
	EnableAutoSignin    bool            `json:"enableAutoSignin"`
	EnableCodeSignin    bool            `json:"enableCodeSignin"`
	EnableSamlCompress  bool            `json:"enableSamlCompress"`
	EnableWebAuthn      bool            `json:"enableWebAuthn"`
	EnableLinkWithEmail bool            `json:"enableLinkWithEmail"`
	SignInStyle         string          `xorm:"varchar(100)" json:"signInStyle"`
	OrgChoiceMode       string          `json:"orgChoiceMode"`
	SamlReplyUrl        string          `xorm:"varchar(100)" json:"samlReplyUrl"`
	Providers           []*ProviderItem `xorm:"mediumtext" json:"providers"`
	SignupItems         []*SignupItem   `xorm:"varchar(1000)" json:"signupItems"`
	GrantTypes          []string        `xorm:"varchar(1000)" json:"grantTypes"`
	OrganizationObj     *Organization   `xorm:"-" json:"organizationObj"`

	ClientId             string     `xorm:"varchar(100)" json:"clientId"`
	ClientSecret         string     `xorm:"varchar(100)" json:"clientSecret"`
	RedirectUris         []string   `xorm:"varchar(1000)" json:"redirectUris"`
	TokenFormat          string     `xorm:"varchar(100)" json:"tokenFormat"`
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
	FormCss              string     `xorm:"text" json:"formCss"`
	FormCssMobile        string     `xorm:"text" json:"formCssMobile"`
	FormOffset           int        `json:"formOffset"`
	FormSideHtml         string     `xorm:"mediumtext" json:"formSideHtml"`
	FormBackgroundUrl    string     `xorm:"varchar(200)" json:"formBackgroundUrl"`
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
	err := adapter.Engine.Desc("created_time").Find(&applications, &Application{Owner: owner})
	if err != nil {
		return applications, err
	}

	return applications, nil
}

func GetOrganizationApplications(owner string, organization string) ([]*Application, error) {
	applications := []*Application{}
	err := adapter.Engine.Desc("created_time").Find(&applications, &Application{Organization: organization})
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
		// Get QRCode only once
		if provider.Type == "WeChat" && provider.DisableSsl && provider.Content == "" {
			provider.Content, err = idp.GetWechatOfficialAccountQRCode(provider.ClientId2, provider.ClientSecret2)
			if err != nil {
				return
			}
			UpdateProvider(provider.Owner+"/"+provider.Name, provider)
		}

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

func getApplication(owner string, name string) (*Application, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	application := Application{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&application)
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

		return &application, nil
	} else {
		return nil, nil
	}
}

func GetApplicationByOrganizationName(organization string) (*Application, error) {
	application := Application{}
	existed, err := adapter.Engine.Where("organization=?", organization).Get(&application)
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
	owner, name := util.GetOwnerAndNameFromId(userId)
	if owner == "app" {
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
	existed, err := adapter.Engine.Where("client_id=?", clientId).Get(&application)
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
	if isUserIdGlobalAdmin(userId) {
		return application
	}

	if application == nil {
		return nil
	}

	if application.ClientSecret != "" {
		application.ClientSecret = "***"
	}

	if application.OrganizationObj != nil {
		if application.OrganizationObj.MasterPassword != "" {
			application.OrganizationObj.MasterPassword = "***"
		}
		if application.OrganizationObj.PasswordType != "" {
			application.OrganizationObj.PasswordType = "***"
		}
		if application.OrganizationObj.PasswordSalt != "" {
			application.OrganizationObj.PasswordSalt = "***"
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

	session := adapter.Engine.ID(core.PK{owner, name}).AllCols()
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

	affected, err := adapter.Engine.Insert(application)
	if err != nil {
		return false, nil
	}

	return affected != 0, nil
}

func DeleteApplication(application *Application) (bool, error) {
	if application.Name == "app-built-in" {
		return false, nil
	}

	affected, err := adapter.Engine.ID(core.PK{application.Owner, application.Name}).Delete(&Application{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (application *Application) GetId() string {
	return fmt.Sprintf("%s/%s", application.Owner, application.Name)
}

func (application *Application) IsRedirectUriValid(redirectUri string) bool {
	isValid := false
	for _, targetUri := range application.RedirectUris {
		targetUriRegex := regexp.MustCompile(targetUri)
		if targetUriRegex.MatchString(redirectUri) || strings.Contains(redirectUri, targetUri) {
			isValid = true
			break
		}
	}
	return isValid
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
	session := adapter.Engine.NewSession()
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
	err = adapter.Engine.Find(&permissions)
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
		_, err = session.Where("name=?", permissions[i].Name).Update(permissions[i])
		if err != nil {
			return err
		}
	}

	return session.Commit()
}
