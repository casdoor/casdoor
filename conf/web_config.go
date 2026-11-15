// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package conf

import "sync/atomic"

// showCloudPromotion mirrors the global setting of the same name, which lives in the database this package cannot read
var showCloudPromotion atomic.Bool

// getEdition is set by the edition package, which reads the license through this package
var getEdition func() string

func SetEditionGetter(getter func() string) {
	getEdition = getter
}

// getHideUpgrades is set by the edition package too
var getHideUpgrades func() bool

func SetHideUpgradesGetter(getter func() bool) {
	getHideUpgrades = getter
}

func SetShowCloudPromotion(show bool) {
	showCloudPromotion.Store(show)
}

type customHtmlAllowed struct {
	tags []string
	urls []string
}

// customHtmlAllowedState mirrors the global settings of what organization admins may use in custom HTML
var customHtmlAllowedState atomic.Pointer[customHtmlAllowed]

func SetCustomHtmlAllowed(tags []string, urls []string) {
	customHtmlAllowedState.Store(&customHtmlAllowed{tags: tags, urls: urls})
}

// GetCustomHtmlAllowedTags returns the elements that organization admins may use in the custom HTML of
// applications besides the built-in ones, set in the global settings. A script still needs custom scripts enabled.
func GetCustomHtmlAllowedTags() []string {
	res := []string{}
	if state := customHtmlAllowedState.Load(); state != nil {
		for _, tag := range state.tags {
			if tag != "script" {
				res = append(res, tag)
			}
		}
	}
	return res
}

// GetCustomHtmlAllowedUrls returns the URL prefixes, set in the global settings, from which organization admins
// may load images, fonts, frames and other files in the custom HTML and CSS of applications
func GetCustomHtmlAllowedUrls() []string {
	if state := customHtmlAllowedState.Load(); state != nil {
		return state.urls
	}
	return nil
}

type WebConfig struct {
	ShowGithubCorner    bool   `json:"showGithubCorner"`
	ForceLanguage       string `json:"forceLanguage"`
	DefaultLanguage     string `json:"defaultLanguage"`
	IsDemoMode          bool   `json:"isDemoMode"`
	StaticBaseUrl       string `json:"staticBaseUrl"`
	DefaultApplication  string `json:"defaultApplication"`
	MaxItemsForFlatMenu int64  `json:"maxItemsForFlatMenu"`
	ThemeShowcase       bool   `json:"themeShowcase"`
	MarketplaceUrl      string `json:"marketplaceUrl"`
	EditionDemoUrls     string `json:"editionDemoUrls"`
	ShowCloudPromotion  bool   `json:"showCloudPromotion"`
	Edition             string `json:"edition"`
	HideUpgrades        bool   `json:"hideUpgrades"`

	CustomHtmlAllowedTags []string `json:"customHtmlAllowedTags"`
}

func GetWebConfig() *WebConfig {
	config := &WebConfig{}

	config.ShowGithubCorner = GetConfigBool("showGithubCorner")
	// an empty forceLanguage means "don't force any language", it must not become "en"
	config.ForceLanguage = GetConfigString("forceLanguage")
	if config.ForceLanguage != "" {
		config.ForceLanguage = GetLanguage(config.ForceLanguage)
	}
	config.DefaultLanguage = GetLanguage(GetConfigString("defaultLanguage"))
	config.IsDemoMode = IsDemoMode()
	config.StaticBaseUrl = GetConfigString("staticBaseUrl")
	config.DefaultApplication = GetConfigString("defaultApplication")
	if config.DefaultApplication == "" {
		config.DefaultApplication = "app-built-in"
	}

	maxItemsForFlatMenu, err := GetConfigInt64("maxItemsForFlatMenu")
	if err != nil {
		maxItemsForFlatMenu = 7
	}
	config.MaxItemsForFlatMenu = maxItemsForFlatMenu

	config.ThemeShowcase = GetConfigBool("themeShowcase") || IsDemoMode() || IsDemoDatabase()
	if config.ThemeShowcase {
		config.MarketplaceUrl = GetConfigString("marketplaceUrl")
		// where the demo site of each edition is, e.g. "Commercial=https://demo-commercial.{domain},Enterprise=..."
		config.EditionDemoUrls = GetConfigString("editionDemoUrls")
	}

	config.Edition = "Community"
	if getEdition != nil {
		config.Edition = getEdition()
	}

	// a licensed deployment has already bought Casdoor, so it is not shown the hosted offering,
	// while the demo sites keep it in every edition
	// an on-premise license was sold by contract, its console does not offer the higher editions; the demo sites always do
	// and a self-hosted community edition does not offer them until a global admin turns on the promotion
	config.HideUpgrades = !config.ThemeShowcase &&
		((getHideUpgrades != nil && getHideUpgrades()) || (config.Edition == "Community" && !showCloudPromotion.Load()))

	config.CustomHtmlAllowedTags = GetCustomHtmlAllowedTags()

	config.ShowCloudPromotion = config.ThemeShowcase || (showCloudPromotion.Load() && config.Edition == "Community")

	return config
}
