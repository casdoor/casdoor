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
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"gopkg.in/yaml.v3"
)

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
	Groups        []*Group        `json:"groups"`
	Adapters      []*Adapter      `json:"adapters"`
	Enforcers     []*Enforcer     `json:"enforcers"`
	Plans         []*Plan         `json:"plans"`
	Pricings      []*Pricing      `json:"pricings"`
	Invitations   []*Invitation   `json:"invitations"`
	Records       []*Record       `json:"records"`
	Sessions      []*Session      `json:"sessions"`
	Subscriptions []*Subscription `json:"subscriptions"`
	Transactions  []*Transaction  `json:"transactions"`
	Sites         []*Site         `json:"sites"`
	Rules         []*Rule         `json:"rules"`

	ThirdPartyLinks []*ThirdPartyLink `json:"third_party_links"`

	EnforcerPolicies map[string][][]string `json:"enforcerPolicies"`
}

var (
	initDataNewOnly bool
	initDataMerge   bool
	initDataApplied string
)

func InitFromFile() {
	initDataFile := conf.GetConfigString("initDataFile")
	if initDataFile == "" {
		return
	}

	initDataNewOnly = conf.GetConfigBool("initDataNewOnly")
	initDataMerge = conf.GetConfigBool("initDataMerge")

	s, err := readInitDataFile(initDataFile)
	if err != nil {
		panic(err)
	}

	if s != "" {
		err = applyInitData(s)
		if err != nil {
			panic(err)
		}
		initDataApplied = s
	}

	interval, err := conf.GetConfigInt64("initDataWatchInterval")
	if err == nil && interval > 0 {
		startInitDataWatchLoop(initDataFile, interval)
	}
}

// startInitDataWatchLoop applies the init data file again whenever its content changes, so the
// objects in it can be managed declaratively (e.g. from a Kubernetes ConfigMap or Secret)
func startInitDataWatchLoop(initDataFile string, interval int64) {
	fmt.Printf("startInitDataWatchLoop() Start!\n\n")
	util.SafeGoroutine(func() {
		lastError := ""
		for {
			time.Sleep(time.Duration(interval) * time.Second)

			applied, err := applyInitDataFileIfChanged(initDataFile)
			if err != nil {
				if err.Error() != lastError {
					fmt.Printf("[%s] Failed to apply the init data file: %s, error: %v\n", util.GetCurrentTime(), initDataFile, err)
				}
				lastError = err.Error()
			} else if applied {
				lastError = ""
				fmt.Printf("[%s] Applied the changed init data file: %s\n", util.GetCurrentTime(), initDataFile)
			}
		}
	})
}

// applyInitDataFileIfChanged doesn't remember a content that failed to apply, so it is retried in the next round
func applyInitDataFileIfChanged(initDataFile string) (applied bool, err error) {
	s, err := readInitDataFile(initDataFile)
	if err != nil || s == "" || s == initDataApplied {
		return false, err
	}

	defer func() {
		if r := recover(); r != nil {
			applied = false
			err = fmt.Errorf("%v", r)
		}
	}()

	err = applyInitData(s)
	if err != nil {
		return false, err
	}

	initDataApplied = s
	return true, nil
}

func applyInitData(s string) error {
	initData, err := parseInitData(s)
	if err != nil {
		return err
	}

	raws, err := parseInitDataRaws(s)
	if err != nil {
		return err
	}

	for i, organization := range initData.Organizations {
		initDefinedOrganization(organization, getInitDataRaw(raws, "organizations", i))
	}
	for i, provider := range initData.Providers {
		initDefinedProvider(provider, getInitDataRaw(raws, "providers", i))
	}
	for i, application := range initData.Applications {
		initDefinedApplication(application, getInitDataRaw(raws, "applications", i))
	}
	for i, user := range initData.Users {
		initDefinedUser(user, getInitDataRaw(raws, "users", i))
	}
	for i, cert := range initData.Certs {
		initDefinedCert(cert, getInitDataRaw(raws, "certs", i))
	}
	for i, ldap := range initData.Ldaps {
		initDefinedLdap(ldap, getInitDataRaw(raws, "ldaps", i))
	}
	for i, model := range initData.Models {
		initDefinedModel(model, getInitDataRaw(raws, "models", i))
	}
	for i, payment := range initData.Payments {
		initDefinedPayment(payment, getInitDataRaw(raws, "payments", i))
	}
	for i, product := range initData.Products {
		initDefinedProduct(product, getInitDataRaw(raws, "products", i))
	}
	for i, resource := range initData.Resources {
		initDefinedResource(resource, getInitDataRaw(raws, "resources", i))
	}
	for i, role := range initData.Roles {
		initDefinedRole(role, getInitDataRaw(raws, "roles", i))
	}
	for i, syncer := range initData.Syncers {
		initDefinedSyncer(syncer, getInitDataRaw(raws, "syncers", i))
	}
	for i, token := range initData.Tokens {
		initDefinedToken(token, getInitDataRaw(raws, "tokens", i))
	}
	for i, webhook := range initData.Webhooks {
		initDefinedWebhook(webhook, getInitDataRaw(raws, "webhooks", i))
	}
	for i, group := range initData.Groups {
		initDefinedGroup(group, getInitDataRaw(raws, "groups", i))
	}
	for i, adapter := range initData.Adapters {
		initDefinedAdapter(adapter, getInitDataRaw(raws, "adapters", i))
	}
	for i, enforcer := range initData.Enforcers {
		policies := initData.EnforcerPolicies[enforcer.GetId()]
		initDefinedEnforcer(enforcer, policies, getInitDataRaw(raws, "enforcers", i))
	}
	for i, permission := range initData.Permissions {
		initDefinedPermission(permission, getInitDataRaw(raws, "permissions", i))
	}
	for i, plan := range initData.Plans {
		initDefinedPlan(plan, getInitDataRaw(raws, "plans", i))
	}
	for i, pricing := range initData.Pricings {
		initDefinedPricing(pricing, getInitDataRaw(raws, "pricings", i))
	}
	for i, invitation := range initData.Invitations {
		initDefinedInvitation(invitation, getInitDataRaw(raws, "invitations", i))
	}
	for _, record := range initData.Records {
		initDefinedRecord(record)
	}
	for _, session := range initData.Sessions {
		initDefinedSession(session)
	}
	for i, subscription := range initData.Subscriptions {
		initDefinedSubscription(subscription, getInitDataRaw(raws, "subscriptions", i))
	}
	for i, transaction := range initData.Transactions {
		initDefinedTransaction(transaction, getInitDataRaw(raws, "transactions", i))
	}
	for i, rule := range initData.Rules {
		initDefinedRule(rule, getInitDataRaw(raws, "rules", i))
	}
	for i, site := range initData.Sites {
		initDefinedSite(site, getInitDataRaw(raws, "sites", i))
	}
	for _, link := range initData.ThirdPartyLinks {
		initThirdPartyLinks(link)
	}
	return nil
}

// readInitDataFile returns the content of the init data file as JSON, a .yaml or .yml file is converted to JSON
func readInitDataFile(filePath string) (string, error) {
	if !util.FileExist(filePath) {
		return "", nil
	}

	s := util.ReadStringFromPath(filePath)

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".yaml" || ext == ".yml" {
		var data interface{}
		err := yaml.Unmarshal([]byte(s), &data)
		if err != nil {
			return "", err
		}
		if data == nil {
			return "", nil
		}

		bytes, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		s = string(bytes)
	}

	return s, nil
}

// parseInitDataRaws keeps the fields of each object as they are written in the file, so the merge mode only
// overwrites the fields that the file sets
func parseInitDataRaws(s string) (map[string][]json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	err := json.Unmarshal([]byte(s), &fields)
	if err != nil {
		return nil, err
	}

	res := map[string][]json.RawMessage{}
	for key, value := range fields {
		items := []json.RawMessage{}
		if json.Unmarshal(value, &items) == nil {
			res[strings.ToLower(key)] = items
		}
	}
	return res, nil
}

func getInitDataRaw(raws map[string][]json.RawMessage, key string, i int) json.RawMessage {
	items := raws[key]
	if i >= len(items) {
		return nil
	}
	return items[i]
}

// mergeInitObject overwrites the existing object with the fields set in the file, then saves it with its update function
func mergeInitObject(existed interface{}, raw json.RawMessage, update func() (bool, error)) {
	if len(raw) != 0 {
		err := json.Unmarshal(raw, existed)
		if err != nil {
			panic(err)
		}
	}

	_, err := update()
	if err != nil {
		panic(err)
	}
}

func parseInitData(s string) (*InitData, error) {
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
		Records:       []*Record{},
		Sessions:      []*Session{},
		Subscriptions: []*Subscription{},
		Transactions:  []*Transaction{},
		Sites:         []*Site{},
		Rules:         []*Rule{},

		ThirdPartyLinks: []*ThirdPartyLink{},

		EnforcerPolicies: map[string][][]string{},
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

func initDefinedOrganization(organization *Organization, raw json.RawMessage) {
	existed, err := getOrganization(organization.Owner, organization.Name)
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			// the masked secrets are kept as they are, as when an admin saves the organization in the UI
			existed, _ = GetMaskedOrganization(true, existed)
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateOrganization(util.GetId(organization.Owner, organization.Name), existed, true, "en")
			})
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
		organization.AccountItems = GetDefaultAccountItems()
	}

	_, err = AddOrganization(organization)
	if err != nil {
		panic(err)
	}
}

func initDefinedApplication(application *Application, raw json.RawMessage) {
	existed, err := getApplication(application.Owner, application.Name)
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateApplication(util.GetId(application.Owner, application.Name), existed, true, "en", nil)
			})
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
	_, err = AddApplication(application, "en")
	if err != nil {
		panic(err)
	}
}

func initDefinedUser(user *User, raw json.RawMessage) {
	existed, err := getUser(user.Owner, user.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			// the password in the file is only the initial one, users may have changed it since
			password := existed.Password
			mergeInitObject(existed, raw, func() (bool, error) {
				existed.Password = password
				return UpdateUser(util.GetId(user.Owner, user.Name), existed, nil, true)
			})
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
	_, err = AddUser(user, "en")
	if err != nil {
		panic(err)
	}
}

func initDefinedCert(cert *Cert, raw json.RawMessage) {
	existed, err := getCert(cert.Owner, cert.Name)
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateCert(util.GetId(cert.Owner, cert.Name), existed)
			})
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

func initDefinedLdap(ldap *Ldap, raw json.RawMessage) {
	existed, err := GetLdap(ldap.Id)
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateLdap(existed)
			})
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

func initDefinedProvider(provider *Provider, raw json.RawMessage) {
	existed, err := GetProvider(util.GetId("admin", provider.Name))
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateProvider(util.GetId("admin", provider.Name), existed)
			})
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

func initDefinedModel(model *Model, raw json.RawMessage) {
	existed, err := GetModel(model.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateModel(model.GetId(), existed)
			})
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

func initDefinedPermission(permission *Permission, raw json.RawMessage) {
	existed, err := GetPermission(permission.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdatePermission(permission.GetId(), existed)
			})
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

func initDefinedPayment(payment *Payment, raw json.RawMessage) {
	existed, err := GetPayment(payment.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdatePayment(payment.GetId(), existed)
			})
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

func initDefinedProduct(product *Product, raw json.RawMessage) {
	existed, err := GetProduct(product.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateProduct(product.GetId(), existed)
			})
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

func initDefinedResource(resource *Resource, raw json.RawMessage) {
	existed, err := GetResource(resource.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateResource(resource.GetId(), existed)
			})
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

func initDefinedRole(role *Role, raw json.RawMessage) {
	existed, err := GetRole(role.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateRole(role.GetId(), existed, true, "en")
			})
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

func initDefinedSyncer(syncer *Syncer, raw json.RawMessage) {
	existed, err := GetSyncer(syncer.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateSyncer(syncer.GetId(), existed, true, "en")
			})
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

func initDefinedToken(token *Token, raw json.RawMessage) {
	existed, err := GetToken(token.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateToken(token.GetId(), existed, true)
			})
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

func initDefinedWebhook(webhook *Webhook, raw json.RawMessage) {
	existed, err := GetWebhook(webhook.GetId())
	if err != nil {
		panic(err)
	}

	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateWebhook(webhook.GetId(), existed, true, "en")
			})
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

func initDefinedGroup(group *Group, raw json.RawMessage) {
	existed, err := getGroup(group.Owner, group.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateGroup(util.GetId(group.Owner, group.Name), existed, true, "en")
			})
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

func initDefinedAdapter(adapter *Adapter, raw json.RawMessage) {
	existed, err := getAdapter(adapter.Owner, adapter.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateAdapter(util.GetId(adapter.Owner, adapter.Name), existed)
			})
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

func initDefinedEnforcer(enforcer *Enforcer, policies [][]string, raw json.RawMessage) {
	existed, err := getEnforcer(enforcer.Owner, enforcer.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateEnforcer(util.GetId(enforcer.Owner, enforcer.Name), existed)
			})
			initEnforcerPolicies(existed, policies)
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

	initEnforcerPolicies(enforcer, policies)
}

func initEnforcerPolicies(enforcer *Enforcer, policies [][]string) {
	err := enforcer.InitEnforcer()
	if err != nil {
		panic(err)
	}

	for _, policy := range policies {
		if enforcer.HasPolicy(policy) {
			continue
		}

		_, err = enforcer.AddPolicy(policy)
		if err != nil {
			panic(err)
		}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		panic(err)
	}
}

func initDefinedPlan(plan *Plan, raw json.RawMessage) {
	existed, err := getPlan(plan.Owner, plan.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdatePlan(util.GetId(plan.Owner, plan.Name), existed)
			})
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

func initDefinedPricing(pricing *Pricing, raw json.RawMessage) {
	existed, err := getPricing(pricing.Owner, pricing.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdatePricing(util.GetId(pricing.Owner, pricing.Name), existed)
			})
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

func initDefinedInvitation(invitation *Invitation, raw json.RawMessage) {
	existed, err := getInvitation(invitation.Owner, invitation.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateInvitation(util.GetId(invitation.Owner, invitation.Name), existed, "en")
			})
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

func initDefinedRecord(record *Record) {
	if initDataMerge {
		return
	}

	record.Id = 0
	record.CreatedTime = util.GetCurrentTime()
	_ = AddRecord(record)
}

func initDefinedSession(session *Session) {
	if initDataMerge {
		return
	}

	session.CreatedTime = util.GetCurrentTime()
	_, err := AddSession(session)
	if err != nil {
		panic(err)
	}
}

func initDefinedSubscription(subscription *Subscription, raw json.RawMessage) {
	existed, err := getSubscription(subscription.Owner, subscription.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateSubscription(util.GetId(subscription.Owner, subscription.Name), existed)
			})
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

func initDefinedTransaction(transaction *Transaction, raw json.RawMessage) {
	existed, err := getTransaction(transaction.Owner, transaction.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateTransaction(util.GetId(transaction.Owner, transaction.Name), existed, "en")
			})
			return
		}
		affected, err := DeleteTransaction(transaction, "en")
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete transaction")
		}
	}
	transaction.CreatedTime = util.GetCurrentTime()
	_, _, err = AddTransaction(transaction, "en", false)
	if err != nil {
		panic(err)
	}
}

func initDefinedSite(site *Site, raw json.RawMessage) {
	existed, err := getSite(site.Owner, site.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateSite(util.GetId(site.Owner, site.Name), existed)
			})
			return
		}
		affected, err := DeleteSite(site)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete site")
		}
	}
	site.CreatedTime = util.GetCurrentTime()
	_, err = AddSite(site)
	if err != nil {
		panic(err)
	}
}

func initDefinedRule(rule *Rule, raw json.RawMessage) {
	existed, err := getRule(rule.Owner, rule.Name)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			mergeInitObject(existed, raw, func() (bool, error) {
				return UpdateRule(util.GetId(rule.Owner, rule.Name), existed)
			})
			return
		}
		affected, err := DeleteRule(rule)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete rule")
		}
	}
	rule.CreatedTime = util.GetCurrentTime()
	_, err = AddRule(rule)
	if err != nil {
		panic(err)
	}
}

func initThirdPartyLinks(link *ThirdPartyLink) {
	existed, err := GetThirdPartyLink(link.Owner, link.UserName, link.ProviderName)
	if err != nil {
		panic(err)
	}
	if existed != nil {
		if initDataNewOnly {
			return
		}
		if initDataMerge {
			return
		}
		affected, err := DeleteThirdPartyLink(link.Owner, link.UserName, link.ProviderName)
		if err != nil {
			panic(err)
		}
		if !affected {
			panic("Fail to delete third party link")
		}
	}
	link.CreatedTime = util.GetCurrentTime()
	_, err = AddThirdPartyLink(link)
	if err != nil {
		panic(err)
	}
}
