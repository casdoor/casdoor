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
	Name        string   `json:"name"`
	Visible     bool     `json:"visible"`
	Required    bool     `json:"required"`
	Prompted    bool     `json:"prompted"`
	Type        string   `json:"type"`
	CustomCss   string   `json:"customCss"`
	Label       string   `json:"label"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options"`
	Regex       string   `json:"regex"`
	Rule        string   `json:"rule"`
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

type JwtItem struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Value    string `json:"value"`
	Type     string `json:"type"`
}

type ScopeItem struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"` // MCP tools allowed by this scope
}

type Application struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName                  string          `xorm:"varchar(100)" json:"displayName"`
	Category                     string          `xorm:"varchar(20)" json:"category"`
	Type                         string          `xorm:"varchar(20)" json:"type"`
	Scopes                       []*ScopeItem    `xorm:"mediumtext" json:"scopes"`
	Logo                         string          `xorm:"varchar(200)" json:"logo"`
	Title                        string          `xorm:"varchar(100)" json:"title"`
	Favicon                      string          `xorm:"varchar(200)" json:"favicon"`
	Order                        int             `json:"order"`
	HomepageUrl                  string          `xorm:"varchar(100)" json:"homepageUrl"`
	Description                  string          `xorm:"varchar(100)" json:"description"`
	Organization                 string          `xorm:"varchar(100)" json:"organization"`
	Cert                         string          `xorm:"varchar(100)" json:"cert"`
	DefaultGroup                 string          `xorm:"varchar(100)" json:"defaultGroup"`
	HeaderHtml                   string          `xorm:"mediumtext" json:"headerHtml"`
	EnablePassword               bool            `json:"enablePassword"`
	EnableSignUp                 bool            `json:"enableSignUp"`
	EnableGuestSignin            bool            `json:"enableGuestSignin"`
	DisableSignin                bool            `json:"disableSignin"`
	EnableSigninSession          bool            `json:"enableSigninSession"`
	EnableAutoSignin             bool            `json:"enableAutoSignin"`
	EnableCodeSignin             bool            `json:"enableCodeSignin"`
	EnableExclusiveSignin        bool            `json:"enableExclusiveSignin"`
	EnableSamlCompress           bool            `json:"enableSamlCompress"`
	EnableSamlC14n10             bool            `json:"enableSamlC14n10"`
	EnableSamlPostBinding        bool            `json:"enableSamlPostBinding"`
	DisableSamlAttributes        bool            `json:"disableSamlAttributes"`
	EnableSamlAssertionSignature bool            `json:"enableSamlAssertionSignature"`
	UseEmailAsSamlNameId         bool            `json:"useEmailAsSamlNameId"`
	EnableWebAuthn               bool            `json:"enableWebAuthn"`
	EnableLinkWithEmail          bool            `json:"enableLinkWithEmail"`
	OrgChoiceMode                string          `json:"orgChoiceMode"`
	SamlReplyUrl                 string          `xorm:"varchar(500)" json:"samlReplyUrl"`
	Providers                    []*ProviderItem `xorm:"mediumtext" json:"providers"`
	SigninMethods                []*SigninMethod `xorm:"varchar(2000)" json:"signinMethods"`
	SignupItems                  []*SignupItem   `xorm:"varchar(3000)" json:"signupItems"`
	SigninItems                  []*SigninItem   `xorm:"mediumtext" json:"signinItems"`
	GrantTypes                   []string        `xorm:"varchar(1000)" json:"grantTypes"`
	OrganizationObj              *Organization   `xorm:"-" json:"organizationObj"`
	CertPublicKey                string          `xorm:"-" json:"certPublicKey"`
	Tags                         []string        `xorm:"mediumtext" json:"tags"`
	SamlAttributes               []*SamlItem     `xorm:"varchar(1000)" json:"samlAttributes"`
	SamlHashAlgorithm            string          `xorm:"varchar(20)" json:"samlHashAlgorithm"`
	IsShared                     bool            `json:"isShared"`
	IpRestriction                string          `json:"ipRestriction"`

	ClientId                string     `xorm:"varchar(100)" json:"clientId"`
	ClientSecret            string     `xorm:"varchar(100)" json:"clientSecret"`
	ClientCert              string     `xorm:"varchar(100)" json:"clientCert"`
	RedirectUris            []string   `xorm:"varchar(1000)" json:"redirectUris"`
	ForcedRedirectOrigin    string     `xorm:"varchar(100)" json:"forcedRedirectOrigin"`
	TokenFormat             string     `xorm:"varchar(100)" json:"tokenFormat"`
	TokenSigningMethod      string     `xorm:"varchar(100)" json:"tokenSigningMethod"`
	TokenFields             []string   `xorm:"varchar(1000)" json:"tokenFields"`
	TokenAttributes         []*JwtItem `xorm:"mediumtext" json:"tokenAttributes"`
	ExpireInHours           float64    `json:"expireInHours"`
	RefreshExpireInHours    float64    `json:"refreshExpireInHours"`
	CookieExpireInHours     int64      `json:"cookieExpireInHours"`
	SignupUrl               string     `xorm:"varchar(200)" json:"signupUrl"`
	SigninUrl               string     `xorm:"varchar(200)" json:"signinUrl"`
	ForgetUrl               string     `xorm:"varchar(200)" json:"forgetUrl"`
	AffiliationUrl          string     `xorm:"varchar(100)" json:"affiliationUrl"`
	IpWhitelist             string     `xorm:"varchar(200)" json:"ipWhitelist"`
	TermsOfUse              string     `xorm:"varchar(200)" json:"termsOfUse"`
	SignupHtml              string     `xorm:"mediumtext" json:"signupHtml"`
	SigninHtml              string     `xorm:"mediumtext" json:"signinHtml"`
	ThemeData               *ThemeData `xorm:"json" json:"themeData"`
	FooterHtml              string     `xorm:"mediumtext" json:"footerHtml"`
	FormCss                 string     `xorm:"text" json:"formCss"`
	FormCssMobile           string     `xorm:"text" json:"formCssMobile"`
	FormOffset              int        `json:"formOffset"`
	FormSideHtml            string     `xorm:"mediumtext" json:"formSideHtml"`
	FormBackgroundUrl       string     `xorm:"varchar(200)" json:"formBackgroundUrl"`
	FormBackgroundUrlMobile string     `xorm:"varchar(200)" json:"formBackgroundUrlMobile"`

	FailedSigninLimit      int `json:"failedSigninLimit"`
	FailedSigninFrozenTime int `json:"failedSigninFrozenTime"`
	CodeResendTimeout      int `json:"codeResendTimeout"`

	CustomScopes []*ScopeDescription `xorm:"mediumtext" json:"customScopes"`

	// Reverse proxy fields
	Domain       string   `xorm:"varchar(100)" json:"domain"`
	OtherDomains []string `xorm:"varchar(1000)" json:"otherDomains"`
	UpstreamHost string   `xorm:"varchar(100)" json:"upstreamHost"`
	SslMode      string   `xorm:"varchar(100)" json:"sslMode"`
	SslCert      string   `xorm:"varchar(100)" json:"sslCert"`

	CertObj *Cert `xorm:"-"`
}

func (application *Application) HasSigninMethod(name string) bool {
	if application == nil {
		return false
	}

	for _, signinMethod := range application.SigninMethods {
		if signinMethod != nil && signinMethod.Name == name && signinMethod.Rule != "Hide password" {
			return true
		}
	}

	return false
}

func GetApplicationCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Application{})
}

func GetOrganizationApplicationCount(owner, organization, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Where("organization = ? or is_shared = ? ", organization, true).Count(&Application{})
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
	err := ormer.Engine.Desc("created_time").Where("organization = ? or is_shared = ? ", organization, true).Find(&applications, &Application{})
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
	err := session.Where("organization = ? or is_shared = ? ", organization, true).Find(&applications, &Application{})
	if err != nil {
		return applications, err
	}

	return applications, nil
}

func getApplication(owner string, name string) (*Application, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	realApplicationName, sharedOrg := util.GetSharedOrgFromApp(name)

	application := Application{Owner: owner, Name: realApplicationName}
	existed, err := ormer.Engine.Get(&application)
	if err != nil {
		return nil, err
	}

	if application.IsShared && sharedOrg != "" {
		application.Organization = sharedOrg
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
	_, name, err := util.GetOwnerAndNameFromIdWithError(userId)
	if err != nil {
		return nil, err
	}
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

	realClientId, sharedOrg := util.GetSharedOrgFromApp(clientId)

	existed, err := ormer.Engine.Where("client_id=?", realClientId).Get(&application)
	if err != nil {
		return nil, err
	}

	if application.IsShared && sharedOrg != "" {
		application.Organization = sharedOrg
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
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getApplication(owner, name)
}

func UpdateApplication(id string, application *Application, isGlobalAdmin bool, lang string) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	oldApplication, err := getApplication(owner, name)
	if oldApplication == nil {
		return false, err
	}

	if !isGlobalAdmin && oldApplication.Organization != application.Organization {
		return false, errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
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

	if application.IsShared == true && application.Organization != "built-in" {
		return false, fmt.Errorf("only applications belonging to built-in organization can be shared")
	}

	err = checkMultipleCaptchaProviders(application, lang)
	if err != nil {
		return false, err
	}

	err = validateCustomScopes(application.CustomScopes, lang)
	if err != nil {
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

	// Initialize default values for required fields to prevent UI errors
	err = extendApplicationWithSignupItems(application)
	if err != nil {
		return false, err
	}

	err = extendApplicationWithSigninItems(application)
	if err != nil {
		return false, err
	}

	err = extendApplicationWithSigninMethods(application)
	if err != nil {
		return false, err
	}

	err = validateCustomScopes(application.CustomScopes, "en")
	if err != nil {
		return false, err
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
	affected, err := ormer.Engine.ID(core.PK{application.Owner, application.Name}).Where("organization = ?", application.Organization).Delete(&Application{})
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
