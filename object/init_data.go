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
	Models        []*Model        `json:"models"`
	Permissions   []*Permission   `json:"permissions"`
	Payments      []*Payment      `json:"payments"`
	Products      []*Product      `json:"products"`
	Resources     []*Resource     `json:"resources"`
	Roles         []*Role         `json:"roles"`
	Syncers       []*Syncer       `json:"syncers"`
	Tokens        []*Token        `json:"tokens"`
	Webhooks      []*Webhook      `json:"webhooks"`
}

func InitFromFile() {
	initData, err := readInitDataFromFile("./init_data.json")
	if err != nil {
		panic(err)
	}

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
		for _, model := range initData.Models {
			initDefinedModel(model)
		}
		for _, permission := range initData.Permissions {
			initDefinedPermission(permission)
		}
		for _, payment := range initData.Payments {
			initDefinedPayment(payment)
		}
		for _, product := range initData.Products {
			initDefinedProduct(product)
		}
		for _, resource := range initData.Resources {
			initDefinedResource(resource)
		}
		for _, role := range initData.Roles {
			initDefinedRole(role)
		}
		for _, syncer := range initData.Syncers {
			initDefinedSyncer(syncer)
		}
		for _, token := range initData.Tokens {
			initDefinedToken(token)
		}
		for _, webhook := range initData.Webhooks {
			initDefinedWebhook(webhook)
		}
	}
}

func readInitDataFromFile(filePath string) (*InitData, error) {
	if !util.FileExist(filePath) {
		return nil, nil
	}

	s := util.ReadStringFromPath(filePath)

	data := &InitData{
		Organizations: []*Organization{},
		Applications:  []*Application{},
		Users:         []*User{},
		Certs:         []*Cert{},
		Providers:     []*Provider{},
		Ldaps:         []*Ldap{},
		Models:        []*Model{},
		Permissions:   []*Permission{},
		Payments:      []*Payment{},
		Products:      []*Product{},
		Resources:     []*Resource{},
		Roles:         []*Role{},
		Syncers:       []*Syncer{},
		Tokens:        []*Token{},
		Webhooks:      []*Webhook{},
	}
	err := util.JsonToStruct(s, data)
	if err != nil {
		return nil, err
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
	for _, permission := range data.Permissions {
		if permission.Actions == nil {
			permission.Actions = []string{}
		}
		if permission.Resources == nil {
			permission.Resources = []string{}
		}
		if permission.Roles == nil {
			permission.Roles = []string{}
		}
		if permission.Users == nil {
			permission.Users = []string{}
		}
	}
	for _, role := range data.Roles {
		if role.Roles == nil {
			role.Roles = []string{}
		}
		if role.Users == nil {
			role.Users = []string{}
		}
	}
	for _, syncer := range data.Syncers {
		if syncer.TableColumns == nil {
			syncer.TableColumns = []*TableColumn{}
		}
	}
	for _, webhook := range data.Webhooks {
		if webhook.Events == nil {
			webhook.Events = []string{}
		}
		if webhook.Headers == nil {
			webhook.Headers = []*Header{}
		}
	}

	return data, nil
}

func initDefinedOrganization(organization *Organization) {
	existed, err := getOrganization(organization.Owner, organization.Name)
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	organization.CreatedTime = util.GetCurrentTime()
	organization.AccountItems = getBuiltInAccountItems()

	_, err = AddOrganization(organization)
	if err != nil {
		panic(err)
	}
}

func initDefinedApplication(application *Application) {
	existed, err := getApplication(application.Owner, application.Name)
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	application.CreatedTime = util.GetCurrentTime()
	_, err = AddApplication(application)
	if err != nil {
		panic(err)
	}
}

func initDefinedUser(user *User) {
	existed, err := getUser(user.Owner, user.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		return
	}
	user.CreatedTime = util.GetCurrentTime()
	user.Id = util.GenerateId()
	user.Properties = make(map[string]string)
	_, err = AddUser(user)
	if err != nil {
		panic(err)
	}
}

func initDefinedCert(cert *Cert) {
	existed, err := getCert(cert.Owner, cert.Name)
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	cert.CreatedTime = util.GetCurrentTime()
	_, err = AddCert(cert)
	if err != nil {
		panic(err)
	}
}

func initDefinedLdap(ldap *Ldap) {
	existed, err := GetLdap(ldap.Id)
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	_, err = AddLdap(ldap)
	if err != nil {
		panic(err)
	}
}

func initDefinedProvider(provider *Provider) {
	existed, err := GetProvider(util.GetId("admin", provider.Name))
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	_, err = AddProvider(provider)
	if err != nil {
		panic(err)
	}
}

func initDefinedModel(model *Model) {
	existed, err := GetModel(model.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	model.CreatedTime = util.GetCurrentTime()
	_, err = AddModel(model)
	if err != nil {
		panic(err)
	}
}

func initDefinedPermission(permission *Permission) {
	existed, err := GetPermission(permission.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	permission.CreatedTime = util.GetCurrentTime()
	_, err = AddPermission(permission)
	if err != nil {
		panic(err)
	}
}

func initDefinedPayment(payment *Payment) {
	existed, err := GetPayment(payment.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	payment.CreatedTime = util.GetCurrentTime()
	_, err = AddPayment(payment)
	if err != nil {
		panic(err)
	}
}

func initDefinedProduct(product *Product) {
	existed, err := GetProduct(product.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	product.CreatedTime = util.GetCurrentTime()
	_, err = AddProduct(product)
	if err != nil {
		panic(err)
	}
}

func initDefinedResource(resource *Resource) {
	existed, err := GetResource(resource.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	resource.CreatedTime = util.GetCurrentTime()
	_, err = AddResource(resource)
	if err != nil {
		panic(err)
	}
}

func initDefinedRole(role *Role) {
	existed, err := GetRole(role.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	role.CreatedTime = util.GetCurrentTime()
	_, err = AddRole(role)
	if err != nil {
		panic(err)
	}
}

func initDefinedSyncer(syncer *Syncer) {
	existed, err := GetSyncer(syncer.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	syncer.CreatedTime = util.GetCurrentTime()
	_, err = AddSyncer(syncer)
	if err != nil {
		panic(err)
	}
}

func initDefinedToken(token *Token) {
	existed, err := GetToken(token.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	token.CreatedTime = util.GetCurrentTime()
	_, err = AddToken(token)
	if err != nil {
		panic(err)
	}
}

func initDefinedWebhook(webhook *Webhook) {
	existed, err := GetWebhook(webhook.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		return
	}
	webhook.CreatedTime = util.GetCurrentTime()
	_, err = AddWebhook(webhook)
	if err != nil {
		panic(err)
	}
}
