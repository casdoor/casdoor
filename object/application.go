// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"github.com/casbin/casdoor/util"
	"xorm.io/core"
)

type Application struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName      string          `xorm:"varchar(100)" json:"displayName"`
	Logo             string          `xorm:"varchar(100)" json:"logo"`
	HomepageUrl      string          `xorm:"varchar(100)" json:"homepageUrl"`
	Description      string          `xorm:"varchar(100)" json:"description"`
	Organization     string          `xorm:"varchar(100)" json:"organization"`
	EnablePassword   bool            `json:"enablePassword"`
	EnableSignUp     bool            `json:"enableSignUp"`
	EnableCodeSignin bool            `json:"enableCodeSignin"`
	Providers        []*ProviderItem `xorm:"varchar(10000)" json:"providers"`
	SignupItems      []*SignupItem   `xorm:"varchar(1000)" json:"signupItems"`
	OrganizationObj  *Organization   `xorm:"-" json:"organizationObj"`

	ClientId       string   `xorm:"varchar(100)" json:"clientId"`
	ClientSecret   string   `xorm:"varchar(100)" json:"clientSecret"`
	RedirectUris   []string `xorm:"varchar(1000)" json:"redirectUris"`
	ExpireInHours  int      `json:"expireInHours"`
	SignupUrl      string   `xorm:"varchar(1000)" json:"signupUrl"`
	SigninUrl      string   `xorm:"varchar(1000)" json:"signinUrl"`
	ForgetUrl      string   `xorm:"varchar(1000)" json:"forgetUrl"`
	AffiliationUrl string   `xorm:"varchar(1000)" json:"affiliationUrl"`
	TermsOfUse     string   `xorm:"varchar(1000)" json:"termsOfUse"`
	SignupHtml     string   `xorm:"mediumtext" json:"signupHtml"`
	SigninHtml     string   `xorm:"mediumtext" json:"signinHtml"`
}

func GetApplicationCount(owner string) int {
	count, err := adapter.Engine.Count(&Application{Owner: owner})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetApplications(owner string) []*Application {
	applications := []*Application{}
	err := adapter.Engine.Desc("created_time").Find(&applications, &Application{Owner: owner})
	if err != nil {
		panic(err)
	}

	return applications
}

func GetPaginationApplications(owner string, offset, limit int) []*Application {
	applications := []*Application{}
	err := adapter.Engine.Desc("created_time").Limit(limit, offset).Find(&applications, &Application{Owner: owner})
	if err != nil {
		panic(err)
	}

	return applications
}

func getProviderMap(owner string) map[string]*Provider {
	providers := GetProviders(owner)
	m := map[string]*Provider{}
	for _, provider := range providers {
		//if provider.Category != "OAuth" {
		//	continue
		//}

		m[provider.Name] = getMaskedProvider(provider)
	}
	return m
}

func extendApplicationWithProviders(application *Application) {
	m := getProviderMap(application.Owner)
	for _, providerItem := range application.Providers {
		if provider, ok := m[providerItem.Name]; ok {
			providerItem.Provider = provider
		}
	}
}

func extendApplicationWithOrg(application *Application) {
	organization := getOrganization(application.Owner, application.Organization)
	application.OrganizationObj = organization
}

func getApplication(owner string, name string) *Application {
	if owner == "" || name == "" {
		return nil
	}

	application := Application{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func GetApplicationByOrganizationName(organization string) *Application {
	application := Application{}
	existed, err := adapter.Engine.Where("organization=?", organization).Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func GetApplicationByUser(user *User) *Application {
	if user.SignupApplication != "" {
		return getApplication("admin", user.SignupApplication)
	} else {
		return GetApplicationByOrganizationName(user.Owner)
	}
}

func GetApplicationByUserId(userId string) (*Application, *User) {
	var application *Application

	owner, name := util.GetOwnerAndNameFromId(userId)
	if owner == "app" {
		application = getApplication("admin", name)
		return application, nil
	}

	user := GetUser(userId)
	application = GetApplicationByUser(user)

	return application, user
}

func GetApplicationByClientId(clientId string) *Application {
	application := Application{}
	existed, err := adapter.Engine.Where("client_id=?", clientId).Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func GetApplicationByClientIdAndSecret(clientId, clientSecret string) *Application {
	if util.IsStrsEmpty(clientId, clientSecret) {
		return nil
	}

	app := GetApplicationByClientId(clientId)
	if app == nil || app.ClientSecret != clientSecret {
		return nil
	}

	return app
}

func GetApplication(id string) *Application {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getApplication(owner, name)
}

func UpdateApplication(id string, application *Application) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getApplication(owner, name) == nil {
		return false
	}

	for _, providerItem := range application.Providers {
		providerItem.Provider = nil
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(application)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddApplication(application *Application) bool {
	application.ClientId = util.GenerateClientId()
	application.ClientSecret = util.GenerateClientSecret()
	for _, providerItem := range application.Providers {
		providerItem.Provider = nil
	}

	affected, err := adapter.Engine.Insert(application)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteApplication(application *Application) bool {
	affected, err := adapter.Engine.ID(core.PK{application.Owner, application.Name}).Delete(&Application{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}
