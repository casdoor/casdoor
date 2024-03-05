// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

func DumpToFile(filePath string) error {
	return writeInitDataToFile(filePath)
}

func writeInitDataToFile(filePath string) error {
	organizations, err := GetOrganizations("admin")
	if err != nil {
		return err
	}

	applications, err := GetApplications("admin")
	if err != nil {
		return err
	}

	users, err := GetGlobalUsers()
	if err != nil {
		return err
	}

	certs, err := GetCerts("")
	if err != nil {
		return err
	}

	providers, err := GetGlobalProviders()
	if err != nil {
		return err
	}

	ldaps, err := GetLdaps("")
	if err != nil {
		return err
	}

	models, err := GetModels("")
	if err != nil {
		return err
	}

	permissions, err := GetPermissions("")
	if err != nil {
		return err
	}

	payments, err := GetPayments("")
	if err != nil {
		return err
	}

	products, err := GetProducts("")
	if err != nil {
		return err
	}

	resources, err := GetResources("", "")
	if err != nil {
		return err
	}

	roles, err := GetRoles("")
	if err != nil {
		return err
	}

	syncers, err := GetSyncers("")
	if err != nil {
		return err
	}

	tokens, err := GetTokens("", "")
	if err != nil {
		return err
	}

	webhooks, err := GetWebhooks("", "")
	if err != nil {
		return err
	}

	groups, err := GetGroups("")
	if err != nil {
		return err
	}

	adapters, err := GetAdapters("")
	if err != nil {
		return err
	}

	enforcers, err := GetEnforcers("")
	if err != nil {
		return err
	}

	plans, err := GetPlans("")
	if err != nil {
		return err
	}

	pricings, err := GetPricings("")
	if err != nil {
		return err
	}

	invitations, err := GetInvitations("")
	if err != nil {
		return err
	}

	records, err := GetRecords()
	if err != nil {
		return err
	}

	sessions, err := GetSessions("")
	if err != nil {
		return err
	}

	subscriptions, err := GetSubscriptions("")
	if err != nil {
		return err
	}

	transactions, err := GetTransactions("")
	if err != nil {
		return err
	}

	data := &InitData{
		Organizations: organizations,
		Applications:  applications,
		Users:         users,
		Certs:         certs,
		Providers:     providers,
		Ldaps:         ldaps,
		Models:        models,
		Permissions:   permissions,
		Payments:      payments,
		Products:      products,
		Resources:     resources,
		Roles:         roles,
		Syncers:       syncers,
		Tokens:        tokens,
		Webhooks:      webhooks,
		Groups:        groups,
		Adapters:      adapters,
		Enforcers:     enforcers,
		Plans:         plans,
		Pricings:      pricings,
		Invitations:   invitations,
		Records:       records,
		Sessions:      sessions,
		Subscriptions: subscriptions,
		Transactions:  transactions,
	}

	text := util.StructToJsonFormatted(data)
	util.WriteStringToPath(text, filePath)

	return nil
}
