// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

import "github.com/casdoor/casdoor/util"

type InitData struct {
	Organizations []*Organization `json:"organizations"`
	Applications  []*Application  `json:"applications"`
	Users         []*User         `json:"users"`
	Certs         []*Cert         `json:"certs"`
	Providers     []*Provider     `json:"providers"`
	Ldaps         []*Ldap         `json:"ldaps"`
}

func InitFromFile() {
	initData := readInitDataFromFile("./init_data.json")
	if initData != nil {
		for _, organization := range initData.Organizations {
			initDefinedOrganization(organization)
		}
		for _, provider := range initData.Providers {
			initDefinedProvider(provider)
		}
		for _, user := range initData.Users {
			initDefinedUser(user)
		}
		for _, application := range initData.Applications {
			initDefinedApplication(application)
		}
		for _, cert := range initData.Certs {
			initDefinedCert(cert)
		}
		for _, ldap := range initData.Ldaps {
			initDefinedLdap(ldap)
		}
	}
}

func readInitDataFromFile(filePath string) *InitData {
	if !util.FileExist(filePath) {
		return nil
	}

	s := util.ReadStringFromPath(filePath)

	data := &InitData{
		Organizations: []*Organization{},
		Applications:  []*Application{},
		Users:         []*User{},
		Certs:         []*Cert{},
		Providers:     []*Provider{},
		Ldaps:         []*Ldap{},
	}
	err := util.JsonToStruct(s, data)
	if err != nil {
		panic(err)
	}

	// transform nil slice to empty slice
	for _, organization := range data.Organizations {
		if organization.Tags == nil {
			organization.Tags = []string{}
		}
	}
	for _, application := range data.Applications {
		if application.Providers == nil {
			application.Providers = []*ProviderItem{}
		}
		if application.SignupItems == nil {
			application.SignupItems = []*SignupItem{}
		}
		if application.GrantTypes == nil {
			application.GrantTypes = []string{}
		}
		if application.RedirectUris == nil {
			application.RedirectUris = []string{}
		}
	}

	return data
}

func initDefinedOrganization(organization *Organization) {
	existed := getOrganization(organization.Owner, organization.Name)
	if existed != nil {
		return
	}
	organization.CreatedTime = util.GetCurrentTime()
	organization.AccountItems = []*AccountItem{
		{Name: "Organization", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "ID", Visible: true, ViewRule: "Public", ModifyRule: "Immutable"},
		{Name: "Name", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Display name", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Avatar", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "User type", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Password", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
		{Name: "Email", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Phone", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Country/Region", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Location", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Affiliation", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Title", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Homepage", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Bio", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Tag", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Signup application", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Roles", Visible: true, ViewRule: "Public", ModifyRule: "Immutable"},
		{Name: "Permissions", Visible: true, ViewRule: "Public", ModifyRule: "Immutable"},
		{Name: "3rd-party logins", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
		{Name: "Properties", Visible: false, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is admin", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is global admin", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is forbidden", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is deleted", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
	}

	AddOrganization(organization)
}

func initDefinedApplication(application *Application) {
	existed := getApplication(application.Owner, application.Name)
	if existed != nil {
		return
	}
	application.CreatedTime = util.GetCurrentTime()
	AddApplication(application)
}

func initDefinedUser(user *User) {
	existed := getUser(user.Owner, user.Name)
	if existed != nil {
		return
	}
	user.CreatedTime = util.GetCurrentTime()
	user.Id = util.GenerateId()
	user.Properties = make(map[string]string)
	AddUser(user)
}

func initDefinedCert(cert *Cert) {
	existed := getCert(cert.Owner, cert.Name)
	if existed != nil {
		return
	}
	cert.CreatedTime = util.GetCurrentTime()
	AddCert(cert)
}

func initDefinedLdap(ldap *Ldap) {
	existed := GetLdap(ldap.Id)
	if existed != nil {
		return
	}
	AddLdap(ldap)
}

func initDefinedProvider(provider *Provider) {
	existed := GetProvider(util.GetId(provider.Name))
	if existed != nil {
		return
	}
	AddProvider(provider)
}
