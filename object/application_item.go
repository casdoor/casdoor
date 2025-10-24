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

func isProviderItemCountryCodeMatched(providerItem *ProviderItem, countryCode string) bool {
	if len(providerItem.CountryCodes) == 0 {
		return true
	}

	for _, countryCode2 := range providerItem.CountryCodes {
		if countryCode2 == "" || countryCode2 == "All" || countryCode2 == "all" || countryCode2 == countryCode {
			return true
		}
	}
	return false
}

func (application *Application) GetProviderByCategoryAndRule(category string, method string, countryCode string) (*Provider, error) {
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
		if providerItem.Provider != nil && providerItem.Provider.Category == "SMS" {
			if !isProviderItemCountryCodeMatched(providerItem, countryCode) {
				continue
			}
		}

		if providerItem.Rule == method || providerItem.Rule == "" || providerItem.Rule == "All" || providerItem.Rule == "all" || providerItem.Rule == "None" {
			if provider, ok := m[providerItem.Name]; ok {
				return provider, nil
			}
		}
	}

	return nil, nil
}

func (application *Application) GetEmailProvider(method string) (*Provider, error) {
	return application.GetProviderByCategoryAndRule("Email", method, "All")
}

func (application *Application) GetSmsProvider(method string, countryCode string) (*Provider, error) {
	return application.GetProviderByCategoryAndRule("SMS", method, countryCode)
}

func (application *Application) GetStorageProvider() (*Provider, error) {
	return application.GetProviderByCategory("Storage")
}

func (application *Application) getSignupItem(itemName string) *SignupItem {
	for _, signupItem := range application.SignupItems {
		if signupItem.Name == itemName {
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

// HasPromptPageForUser checks if a specific user needs to see the prompt page
// This includes checking for missing required signup items
func (application *Application) HasPromptPageForUser(user *User) bool {
	if application.HasPromptPage() {
		return true
	}

	// Check if the user has missing required signup items
	return application.HasMissingRequiredSignupItems(user)
}

// GetMissingRequiredSignupItems returns a list of required signup items that are not filled for the given user
func (application *Application) GetMissingRequiredSignupItems(user *User) []*SignupItem {
	missing := []*SignupItem{}
	
	for _, signupItem := range application.SignupItems {
		if !signupItem.Required {
			continue
		}
		
		// Check if the field is empty in the user
		isEmpty := false
		switch signupItem.Name {
		case "Email":
			isEmpty = user.Email == ""
		case "Phone":
			isEmpty = user.Phone == ""
		case "Display name":
			isEmpty = user.DisplayName == ""
		case "Affiliation":
			isEmpty = user.Affiliation == ""
		case "ID card":
			isEmpty = user.IdCard == ""
		case "Region":
			isEmpty = user.Region == ""
		case "Country/Region":
			isEmpty = user.Region == ""
		// Add more fields as needed
		default:
			// For custom fields, we can't easily check, so skip
			continue
		}
		
		if isEmpty {
			missing = append(missing, signupItem)
		}
	}
	
	return missing
}

// HasMissingRequiredSignupItems checks if a user has missing required signup items
func (application *Application) HasMissingRequiredSignupItems(user *User) bool {
	return len(application.GetMissingRequiredSignupItems(user)) > 0
}
