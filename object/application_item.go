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
	"github.com/casdoor/casdoor/util"
)

func (application *Application) GetProviderByCategory(category string) (*Provider, error) {
	providers, err := GetProviders(application.Organization)
	if err != nil {
		return nil, err
	}

	m := map[string]*Provider{}
	for _, provider := range providers {
		if provider.Category != category {
			continue
		}

		m[provider.Name] = provider
	}

	for _, providerItem := range application.Providers {
		if provider, ok := m[providerItem.Name]; ok {
			return provider, nil
		}
	}

	return nil, nil
}

func (application *Application) GetEmailProvider() (*Provider, error) {
	return application.GetProviderByCategory("Email")
}

func (application *Application) GetSmsProvider() (*Provider, error) {
	return application.GetProviderByCategory("SMS")
}

func (application *Application) GetStorageProvider() (*Provider, error) {
	return application.GetProviderByCategory("Storage")
}

func (application *Application) getSignupItem(itemName string) *SignupItem {
	for _, signupItem := range application.SignupItems {
		if signupItem.Name == itemName {
			return signupItem
		} else if signupItem.Name == "Display name" {
			if signupItem.Rule == "First, last" && (itemName == "FirstName" || itemName == "LastName") {
				return signupItem
			} else if signupItem.Rule == "Real name" && itemName == "Name" {
				return signupItem
			} else if signupItem.Rule == "None" && itemName == "Name" {
				return signupItem
			}
		} else if signupItem.Name == "Country/Region" && itemName == "Region" {
			return signupItem
		}
	}
	return nil
}

func (application *Application) IsSignupItemVisible(itemName string) bool {
	signupItem := application.getSignupItem(itemName)
	if signupItem == nil {
		return false
	}

	return signupItem.Visible
}

func (application *Application) IsSignupItemRequired(itemName string) bool {
	signupItem := application.getSignupItem(itemName)
	if signupItem == nil {
		return false
	}

	return signupItem.Required
}

func (application *Application) IsSignupItemRegex(itemName string) (bool, string) {
	signupItem := application.getSignupItem(itemName)
	if signupItem == nil || signupItem.Regex == "" {
		return false, ""
	}

	return true, signupItem.Regex
}

func (application *Application) validRegex(formItem string, itemName string) (bool, string) {
	valid := true
	regexNotNull, Regex := application.IsSignupItemRegex(itemName)
	if regexNotNull {
		var errMsg string
		valid, errMsg = util.IsStringValidRegex(formItem, Regex)
		if errMsg != "" {
			return true, errMsg
		}
	}
	return valid, ""
}

func (si *SignupItem) isSignupItemPrompted() bool {
	return si.Visible && si.Prompted
}

func (application *Application) GetSignupItemRule(itemName string) string {
	signupItem := application.getSignupItem(itemName)
	if signupItem == nil {
		return ""
	}

	return signupItem.Rule
}

func (application *Application) getAllPromptedProviderItems() []*ProviderItem {
	res := []*ProviderItem{}
	for _, providerItem := range application.Providers {
		if providerItem.isProviderPrompted() {
			res = append(res, providerItem)
		}
	}
	return res
}

func (application *Application) getAllPromptedSignupItems() []*SignupItem {
	res := []*SignupItem{}
	for _, signupItem := range application.SignupItems {
		if signupItem.isSignupItemPrompted() {
			res = append(res, signupItem)
		}
	}
	return res
}

func (application *Application) isAffiliationPrompted() bool {
	signupItem := application.getSignupItem("Affiliation")
	if signupItem == nil {
		return false
	}

	return signupItem.Prompted
}

func (application *Application) HasPromptPage() bool {
	providerItems := application.getAllPromptedProviderItems()
	if len(providerItems) != 0 {
		return true
	}

	signupItems := application.getAllPromptedSignupItems()
	if len(signupItems) != 0 {
		return true
	}

	return application.isAffiliationPrompted()
}
