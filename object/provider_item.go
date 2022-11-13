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

type ProviderItem struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`

	CanSignUp bool      `json:"canSignUp"`
	CanSignIn bool      `json:"canSignIn"`
	CanUnlink bool      `json:"canUnlink"`
	Prompted  bool      `json:"prompted"`
	AlertType string    `json:"alertType"`
	Rule      string    `json:"rule"`
	Provider  *Provider `json:"provider"`
}

func (application *Application) GetProviderItem(providerName string) *ProviderItem {
	for _, providerItem := range application.Providers {
		if providerItem.Name == providerName {
			return providerItem
		}
	}
	return nil
}

func (application *Application) GetProviderItemByType(providerType string) *ProviderItem {
	for _, item := range application.Providers {
		if item.Provider.Type == providerType {
			return item
		}
	}
	return nil
}

func (pi *ProviderItem) IsProviderVisible() bool {
	if pi.Provider == nil {
		return false
	}
	return pi.Provider.Category == "OAuth" || pi.Provider.Category == "SAML"
}

func (pi *ProviderItem) isProviderPrompted() bool {
	return pi.IsProviderVisible() && pi.Prompted
}
