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

import (
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
)

type InitData struct {
	Organizations []*Organization       `json:"organizations"`
	Applications  []*Application        `json:"applications"`
	Users         []*User               `json:"users"`
	Certs         []*Cert               `json:"certs"`
	Providers     []*Provider           `json:"providers"`
	Ldaps         []*Ldap               `json:"ldaps"`
	Models        []*Model              `json:"models"`
	Permissions   []*Permission         `json:"permissions"`
	Payments      []*Payment            `json:"payments"`
	Products      []*Product            `json:"products"`
	Resources     []*Resource           `json:"resources"`
	Roles         []*Role               `json:"roles"`
	Syncers       []*Syncer             `json:"syncers"`
	Tokens        []*Token              `json:"tokens"`
	Webhooks      []*Webhook            `json:"webhooks"`
	Groups        []*Group              `json:"groups"`
	Adapters      []*Adapter            `json:"adapters"`
	Enforcers     []*Enforcer           `json:"enforcers"`
	Plans         []*Plan               `json:"plans"`
	Pricings      []*Pricing            `json:"pricings"`
	Invitations   []*Invitation         `json:"invitations"`
	Records       []*casvisorsdk.Record `json:"records"`
	Sessions      []*Session            `json:"sessions"`
	Subscriptions []*Subscription       `json:"subscriptions"`
	Transactions  []*Transaction        `json:"transactions"`
}

var initDataNewOnly bool

func InitFromFile() {
	initDataFile := conf.GetConfigString("initDataFile")
	if initDataFile == "" {
		return
	}

	initDataNewOnly = conf.GetConfigBool("initDataNewOnly")

	initData, err := readInitDataFromFile(initDataFile)
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
		for _, application := range initData.Applications {
			initDefinedApplication(application)
		}
		for _, user := range initData.Users {
			initDefinedUser(user)
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
		for _, group := range initData.Groups {
			initDefinedGroup(group)
		}
		for _, adapter := range initData.Adapters {
			initDefinedAdapter(adapter)
		}
		for _, enforcer := range initData.Enforcers {
			initDefinedEnforcer(enforcer)
		}
		for _, plan := range initData.Plans {
			initDefinedPlan(plan)
		}
		for _, pricing := range initData.Pricings {
			initDefinedPricing(pricing)
		}
		for _, invitation := range initData.Invitations {
			initDefinedInvitation(invitation)
		}
		for _, record := range initData.Records {
			initDefinedRecord(record)
		}
		for _, session := range initData.Sessions {
			initDefinedSession(session)
		}
		for _, subscription := range initData.Subscriptions {
			initDefinedSubscription(subscription)
		}
		for _, transaction := range initData.Transactions {
			initDefinedTransaction(transaction)
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
		Groups:        []*Group{},
		Adapters:      []*Adapter{},
		Enforcers:     []*Enforcer{},
		Plans:         []*Plan{},
		Pricings:      []*Pricing{},
		Invitations:   []*Invitation{},
		Records:       []*casvisorsdk.Record{},
		Sessions:      []*Session{},
		Subscriptions: []*Subscription{},
		Transactions:  []*Transaction{},
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
		if organization.AccountItems == nil {
			organization.AccountItems = []*AccountItem{}
		}
	}
	for _, application := range data.Applications {
		if application.Providers == nil {
			application.Providers = []*ProviderItem{}
		}
		if application.SigninMethods == nil {
			application.SigninMethods = []*SigninMethod{}
		}
		if application.SignupItems == nil {
			application.SignupItems = []*SignupItem{}
		}
		if application.GrantTypes == nil {
			application.GrantTypes = []string{}
		}
		if application.Tags == nil {
			application.Tags = []string{}
		}
		if application.RedirectUris == nil {
			application.RedirectUris = []string{}
		}
		if application.TokenFields == nil {
			application.TokenFields = []string{}
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
	for _, plan := range data.Plans {
		if plan.PaymentProviders == nil {
			plan.PaymentProviders = []string{}
		}
	}
	for _, pricing := range data.Pricings {
		if pricing.Plans == nil {
			pricing.Plans = []string{}
		}
	}
	for _, session := range data.Sessions {
		if session.SessionId == nil {
			session.SessionId = []string{}
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
		if initDataNewOnly {
			return
		}
		affected, err := deleteOrganization(organization)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete organization")
		}
	}
	organization.CreatedTime = util.GetCurrentTime()
	if len(organization.AccountItems) == 0 {
		organization.AccountItems = getBuiltInAccountItems()
	}

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
		if initDataNewOnly {
			return
		}
		affected, err := deleteApplication(application)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete application")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := deleteUser(user)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete user")
		}
	}
	user.CreatedTime = util.GetCurrentTime()
	user.Id = util.GenerateId()
	if user.Properties == nil {
		user.Properties = make(map[string]string)
	}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteCert(cert)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete cert")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteLdap(ldap)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete ldap")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteProvider(provider)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete provider")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteModel(model)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete provider")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := deletePermission(permission)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete permission")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeletePayment(payment)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete payment")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteProduct(product)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete product")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteResource(resource)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete resource")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := deleteRole(role)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete role")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteSyncer(syncer)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete role")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteToken(token)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete token")
		}
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
		if initDataNewOnly {
			return
		}
		affected, err := DeleteWebhook(webhook)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete webhook")
		}
	}
	webhook.CreatedTime = util.GetCurrentTime()
	_, err = AddWebhook(webhook)
	if err != nil {
		panic(err)
	}
}

func initDefinedGroup(group *Group) {
	existed, err := getGroup(group.Owner, group.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		affected, err := deleteGroup(group)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete group")
		}
	}
	group.CreatedTime = util.GetCurrentTime()
	_, err = AddGroup(group)
	if err != nil {
		panic(err)
	}
}

func initDefinedAdapter(adapter *Adapter) {
	existed, err := getAdapter(adapter.Owner, adapter.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		affected, err := DeleteAdapter(adapter)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete adapter")
		}
	}
	adapter.CreatedTime = util.GetCurrentTime()
	_, err = AddAdapter(adapter)
	if err != nil {
		panic(err)
	}
}

func initDefinedEnforcer(enforcer *Enforcer) {
	existed, err := getEnforcer(enforcer.Owner, enforcer.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		affected, err := DeleteEnforcer(enforcer)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete enforcer")
		}
	}
	enforcer.CreatedTime = util.GetCurrentTime()
	_, err = AddEnforcer(enforcer)
	if err != nil {
		panic(err)
	}
}

func initDefinedPlan(plan *Plan) {
	existed, err := getPlan(plan.Owner, plan.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		affected, err := DeletePlan(plan)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete plan")
		}
	}
	plan.CreatedTime = util.GetCurrentTime()
	_, err = AddPlan(plan)
	if err != nil {
		panic(err)
	}
}

func initDefinedPricing(pricing *Pricing) {
	existed, err := getPricing(pricing.Owner, pricing.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		affected, err := DeletePricing(pricing)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete pricing")
		}
	}
	pricing.CreatedTime = util.GetCurrentTime()
	_, err = AddPricing(pricing)
	if err != nil {
		panic(err)
	}
}

func initDefinedInvitation(invitation *Invitation) {
	existed, err := getInvitation(invitation.Owner, invitation.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		affected, err := DeleteInvitation(invitation)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete invitation")
		}
	}
	invitation.CreatedTime = util.GetCurrentTime()
	_, err = AddInvitation(invitation, "en")
	if err != nil {
		panic(err)
	}
}

func initDefinedRecord(record *casvisorsdk.Record) {
	record.Id = 0
	record.CreatedTime = util.GetCurrentTime()
	_ = AddRecord(record)
}

func initDefinedSession(session *Session) {
	session.CreatedTime = util.GetCurrentTime()
	_, err := AddSession(session)
	if err != nil {
		panic(err)
	}
}

func initDefinedSubscription(subscription *Subscription) {
	existed, err := getSubscription(subscription.Owner, subscription.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		affected, err := DeleteSubscription(subscription)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete subscription")
		}
	}
	subscription.CreatedTime = util.GetCurrentTime()
	_, err = AddSubscription(subscription)
	if err != nil {
		panic(err)
	}
}

func initDefinedTransaction(transaction *Transaction) {
	existed, err := getTransaction(transaction.Owner, transaction.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		affected, err := DeleteTransaction(transaction)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete transaction")
		}
	}
	transaction.CreatedTime = util.GetCurrentTime()
	_, err = AddTransaction(transaction)
	if err != nil {
		panic(err)
	}
}
