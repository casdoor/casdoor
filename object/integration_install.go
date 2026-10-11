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

package object

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const defaultMarketplaceUrl = "https://cdn.jsdelivr.net/gh/casdoor/casdoor-integrations@dist"

type MarketplaceVersion struct {
	Version        string `json:"version"`
	CasdoorVersion string `json:"casdoorVersion"`
	Url            string `json:"url"`
	Sha256         string `json:"sha256"`
	PublishedAt    string `json:"publishedAt"`
}

type MarketplaceIntegration struct {
	Id                  string                `json:"id"`
	Type                string                `json:"type"`
	Name                map[string]string     `json:"name"`
	Description         map[string]string     `json:"description"`
	Author              string                `json:"author"`
	License             string                `json:"license"`
	Homepage            string                `json:"homepage"`
	Tags                []string              `json:"tags"`
	Categories          []string              `json:"categories"`
	Logo                string                `json:"logo"`
	Screenshots         []string              `json:"screenshots"`
	Verified            bool                  `json:"verified"`
	RequiresGlobalAdmin bool                  `json:"requiresGlobalAdmin"`
	Latest              string                `json:"latest"`
	Versions            []*MarketplaceVersion `json:"versions"`
}

type MarketplaceRevoked struct {
	Id      string `json:"id"`
	Version string `json:"version"`
}

type MarketplaceIndex struct {
	SchemaVersion int                       `json:"schemaVersion"`
	GeneratedAt   string                    `json:"generatedAt"`
	BaseUrl       string                    `json:"baseUrl"`
	Integrations  []*MarketplaceIntegration `json:"integrations"`
	Revoked       []*MarketplaceRevoked     `json:"revoked"`
}

type IntegrationVariable struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type IntegrationManifest struct {
	Id        string                 `json:"id"`
	Type      string                 `json:"type"`
	Version   string                 `json:"version"`
	Name      map[string]string      `json:"name"`
	Variables []*IntegrationVariable `json:"variables"`
}

type IntegrationBundle struct {
	Manifest *IntegrationManifest       `json:"manifest"`
	Files    map[string]json.RawMessage `json:"files"`
}

type IntegrationInstallRequest struct {
	Owner         string            `json:"owner"`
	Name          string            `json:"name"`
	IntegrationId string            `json:"integrationId"`
	Version       string            `json:"version"`
	Variables     map[string]string `json:"variables"`
	// theme: the application to restyle; provider: the application to add the provider to, optional
	Application string `json:"application"`
	// provider: the client the other identity provider issued
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type applicationTemplate struct {
	RedirectUris     []string `json:"redirectUris"`
	GrantTypes       []string `json:"grantTypes"`
	TokenFormat      string   `json:"tokenFormat"`
	TokenGroupFormat string   `json:"tokenGroupFormat"`
}

type themeFields struct {
	ThemeData     *ThemeData `json:"themeData"`
	FormOffset    int        `json:"formOffset"`
	FormCss       string     `json:"formCss"`
	FormCssMobile string     `json:"formCssMobile"`
	FormSideHtml  string     `json:"formSideHtml"`
	FooterHtml    string     `json:"footerHtml"`

	FormBackgroundUrl       string `json:"formBackgroundUrl"`
	FormBackgroundUrlMobile string `json:"formBackgroundUrlMobile"`
}

var themeColumns = []string{"theme_data", "form_offset", "form_css", "form_css_mobile", "form_side_html", "footer_html", "form_background_url", "form_background_url_mobile"}

var (
	marketplaceIndexMutex sync.Mutex
	marketplaceIndexCache *MarketplaceIndex
	marketplaceIndexTime  time.Time
)

var (
	variablePattern  = regexp.MustCompile(`\{\{\s*([a-zA-Z][a-zA-Z0-9]*)\s*\}\}`)
	colorPattern     = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	installNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
)

func getMarketplaceUrl() string {
	marketplaceUrl := conf.GetConfigString("marketplaceUrl")
	if marketplaceUrl == "" {
		marketplaceUrl = defaultMarketplaceUrl
	}
	return strings.TrimSuffix(marketplaceUrl, "/")
}

func fetchMarketplaceFile(fileUrl string) ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(fileUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %s, status code: %d", fileUrl, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 10<<20))
}

func joinMarketplaceUrl(baseUrl string, path string) string {
	if path == "" || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		return path
	}
	return fmt.Sprintf("%s/%s", baseUrl, strings.TrimPrefix(path, "/"))
}

// GetMarketplaceIndex returns the catalog of the Marketplace with logos and screenshots as absolute URLs.
func GetMarketplaceIndex(refresh bool) (*MarketplaceIndex, error) {
	marketplaceIndexMutex.Lock()
	defer marketplaceIndexMutex.Unlock()

	if !refresh && marketplaceIndexCache != nil && time.Since(marketplaceIndexTime) < 10*time.Minute {
		return marketplaceIndexCache, nil
	}

	baseUrl := getMarketplaceUrl()
	data, err := fetchMarketplaceFile(baseUrl + "/index.json")
	if err != nil {
		return nil, err
	}

	index := &MarketplaceIndex{}
	err = json.Unmarshal(data, index)
	if err != nil {
		return nil, err
	}

	index.BaseUrl = baseUrl
	for _, item := range index.Integrations {
		item.Logo = joinMarketplaceUrl(baseUrl, item.Logo)
		for i, screenshot := range item.Screenshots {
			item.Screenshots[i] = joinMarketplaceUrl(baseUrl, screenshot)
		}
	}

	marketplaceIndexCache = index
	marketplaceIndexTime = time.Now()
	return index, nil
}

func GetMarketplaceIntegration(id string) (*MarketplaceIntegration, error) {
	index, err := GetMarketplaceIndex(false)
	if err != nil {
		return nil, err
	}

	for _, item := range index.Integrations {
		if item.Id == id {
			return item, nil
		}
	}
	return nil, fmt.Errorf("the integration: %s is not in the Marketplace", id)
}

func isIntegrationRevoked(index *MarketplaceIndex, id string, version string) bool {
	for _, revoked := range index.Revoked {
		if revoked.Id == id && (revoked.Version == "" || revoked.Version == version) {
			return true
		}
	}
	return false
}

func getIntegrationBundle(item *MarketplaceIntegration, version string) (*IntegrationBundle, string, error) {
	index, err := GetMarketplaceIndex(false)
	if err != nil {
		return nil, "", err
	}
	if version == "" {
		version = item.Latest
	}
	if isIntegrationRevoked(index, item.Id, version) {
		return nil, "", fmt.Errorf("the integration: %s %s has been withdrawn from the Marketplace", item.Id, version)
	}

	var marketplaceVersion *MarketplaceVersion
	for _, v := range item.Versions {
		if v.Version == version {
			marketplaceVersion = v
		}
	}
	if marketplaceVersion == nil {
		return nil, "", fmt.Errorf("the integration: %s has no version %s", item.Id, version)
	}

	data, err := fetchMarketplaceFile(joinMarketplaceUrl(index.BaseUrl, marketplaceVersion.Url))
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256(data)
	if !strings.EqualFold(hex.EncodeToString(hash[:]), marketplaceVersion.Sha256) {
		return nil, "", fmt.Errorf("the checksum of %s %s does not match the Marketplace index", item.Id, version)
	}

	bundle := &IntegrationBundle{}
	err = json.Unmarshal(data, bundle)
	if err != nil {
		return nil, "", err
	}
	if bundle.Manifest == nil || bundle.Manifest.Id != item.Id || bundle.Manifest.Type != item.Type || bundle.Manifest.Version != version {
		return nil, "", fmt.Errorf("the bundle of %s %s does not match the Marketplace index", item.Id, version)
	}

	return bundle, string(data), nil
}

// GetMarketplaceBundle returns the manifest and content files of an integration version, checked against the index.
func GetMarketplaceBundle(id string, version string) (json.RawMessage, error) {
	item, err := GetMarketplaceIntegration(id)
	if err != nil {
		return nil, err
	}

	_, bundleText, err := getIntegrationBundle(item, version)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(bundleText), nil
}

func getIntegrationVariables(bundle *IntegrationBundle, input map[string]string) (map[string]string, error) {
	res := map[string]string{}
	for _, variable := range bundle.Manifest.Variables {
		value := strings.TrimSpace(input[variable.Name])
		if value == "" {
			if variable.Required {
				return nil, fmt.Errorf("%s is required", variable.Name)
			}
			res[variable.Name] = ""
			continue
		}

		switch variable.Type {
		case "url":
			value = strings.TrimRight(value, "/")
			u, err := url.Parse(value)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || strings.ContainsAny(value, " \"'<>`") {
				return nil, fmt.Errorf("%s is not a valid http(s) URL: %s", variable.Name, value)
			}
		case "color":
			if !colorPattern.MatchString(value) {
				return nil, fmt.Errorf("%s is not a #rrggbb color: %s", variable.Name, value)
			}
		default:
			if len(value) > 500 {
				return nil, fmt.Errorf("%s is too long", variable.Name)
			}
		}
		res[variable.Name] = value
	}
	return res, nil
}

func replaceIntegrationVariables(value interface{}, variables map[string]string, escape func(string) string) interface{} {
	switch v := value.(type) {
	case string:
		return variablePattern.ReplaceAllStringFunc(v, func(match string) string {
			name := variablePattern.FindStringSubmatch(match)[1]
			if res, ok := variables[name]; ok {
				return escape(res)
			}
			return match
		})
	case map[string]interface{}:
		for key, item := range v {
			v[key] = replaceIntegrationVariables(item, variables, escape)
		}
	case []interface{}:
		for i, item := range v {
			v[i] = replaceIntegrationVariables(item, variables, escape)
		}
	}
	return value
}

// renderIntegrationFile fills the {{variables}} of a content file into out. Values going into HTML fields
// are escaped, values going into CSS must not be able to leave the declaration they are in.
func renderIntegrationFile(bundle *IntegrationBundle, file string, variables map[string]string, out interface{}) error {
	raw, ok := bundle.Files[file]
	if !ok {
		return fmt.Errorf("the integration has no %s", file)
	}

	var data map[string]interface{}
	err := json.Unmarshal(raw, &data)
	if err != nil {
		return err
	}

	var cssErr error
	for key, value := range data {
		escape := func(s string) string { return s }
		if strings.HasSuffix(key, "Html") {
			escape = html.EscapeString
		} else if strings.HasPrefix(key, "formCss") || key == "themeData" {
			escape = func(s string) string {
				if strings.ContainsAny(s, "<>{};\\\"'@\r\n") {
					cssErr = fmt.Errorf("the value %q cannot be used in a stylesheet", s)
				}
				return s
			}
		}
		data[key] = replaceIntegrationVariables(value, variables, escape)
	}
	if cssErr != nil {
		return cssErr
	}

	rendered, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(rendered, out)
}

func getIntegrationApplication(owner string, name string, isGlobalAdmin bool) (*Application, error) {
	application, err := getApplication("admin", name)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, fmt.Errorf("the application: %s does not exist", name)
	}
	if application.Organization != owner && !isGlobalAdmin {
		return nil, fmt.Errorf("the application: %s does not belong to the organization: %s", name, owner)
	}
	return application, nil
}

func installAppIntegration(integration *Integration, bundle *IntegrationBundle, variables map[string]string, lang string) error {
	template := applicationTemplate{}
	err := renderIntegrationFile(bundle, "application.json", variables, &template)
	if err != nil {
		return err
	}

	existing, err := getApplication("admin", integration.Name)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("an application named %s already exists, choose another name", integration.Name)
	}

	grantTypes := template.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code", "refresh_token"}
	}
	logo := integration.Logo
	if logo == "" {
		logo = fmt.Sprintf("%s/img/casdoor-logo_1185x256.png", conf.GetConfigString("staticBaseUrl"))
	}

	application := &Application{
		Owner:                "admin",
		Name:                 integration.Name,
		CreatedTime:          util.GetCurrentTime(),
		DisplayName:          integration.DisplayName,
		Category:             "Default",
		Type:                 "All",
		Logo:                 logo,
		HomepageUrl:          variables["appUrl"],
		Organization:         integration.Owner,
		Cert:                 "cert-built-in",
		EnablePassword:       true,
		EnableSignUp:         true,
		Providers:            []*ProviderItem{},
		Scopes:               []*ScopeItem{},
		Tags:                 []string{},
		GrantTypes:           grantTypes,
		RedirectUris:         template.RedirectUris,
		TokenFormat:          template.TokenFormat,
		TokenGroupFormat:     template.TokenGroupFormat,
		ExpireInHours:        24 * 7,
		RefreshExpireInHours: 24 * 7,
		CookieExpireInHours:  24 * 30,
		FormOffset:           2,
	}
	affected, err := AddApplication(application, lang)
	if err != nil {
		return err
	}
	if !affected {
		return fmt.Errorf("failed to add the application: %s", integration.Name)
	}

	integration.Application = application.Name
	return nil
}

func installProviderIntegration(integration *Integration, bundle *IntegrationBundle, variables map[string]string, req *IntegrationInstallRequest, isGlobalAdmin bool, lang string) error {
	if req.ClientId == "" || req.ClientSecret == "" {
		return fmt.Errorf("the client ID and client secret are required")
	}

	var application *Application
	var err error
	if req.Application != "" {
		application, err = getIntegrationApplication(integration.Owner, req.Application, isGlobalAdmin)
		if err != nil {
			return err
		}
	}

	provider := &Provider{}
	err = renderIntegrationFile(bundle, "provider.json", variables, provider)
	if err != nil {
		return err
	}

	existing := &Provider{Name: integration.Name}
	existed, err := ormer.Engine.Get(existing)
	if err != nil {
		return err
	}
	if existed {
		return fmt.Errorf("a provider named %s already exists, choose another name", integration.Name)
	}

	provider.Owner = integration.Owner
	provider.Name = integration.Name
	provider.CreatedTime = util.GetCurrentTime()
	provider.Method = "Normal"
	provider.ClientId = req.ClientId
	provider.ClientSecret = req.ClientSecret
	provider.CustomLogo = integration.Logo
	affected, err := AddProvider(provider)
	if err != nil {
		return err
	}
	if !affected {
		return fmt.Errorf("failed to add the provider: %s", integration.Name)
	}
	integration.Provider = provider.Name

	if application != nil {
		application.Providers = append(application.Providers, &ProviderItem{
			Owner:     provider.Owner,
			Name:      provider.Name,
			CanSignUp: true,
			CanSignIn: true,
			CanUnlink: true,
			Rule:      "None",
		})
		_, err = UpdateApplication(application.GetId(), application, isGlobalAdmin, lang, []string{"providers"})
		if err != nil {
			return err
		}
		integration.Application = application.Name
	}
	return nil
}

func installThemeIntegration(integration *Integration, item *MarketplaceIntegration, bundle *IntegrationBundle, variables map[string]string, req *IntegrationInstallRequest, isGlobalAdmin bool, lang string) error {
	if req.Application == "" {
		return fmt.Errorf("choose the application to apply the theme to")
	}

	application, err := getIntegrationApplication(integration.Owner, req.Application, isGlobalAdmin)
	if err != nil {
		return err
	}

	installed := []*Integration{}
	err = ormer.Engine.Find(&installed, &Integration{Type: "theme", Application: application.Name})
	if err != nil {
		return err
	}
	if len(installed) > 0 {
		return fmt.Errorf("the application: %s already has the theme: %s, uninstall it first", application.Name, installed[0].Name)
	}

	theme := themeFields{}
	err = renderIntegrationFile(bundle, "theme.json", variables, &theme)
	if err != nil {
		return err
	}

	// an organization admin's save would quietly keep the old CSS or HTML instead
	if !isGlobalAdmin {
		prefixes := getCssUrlPrefixes()
		isSafe := isCssSafeFor(theme.FormCss, prefixes) && isCssSafeFor(theme.FormCssMobile, prefixes) &&
			isHtmlSafeFor(theme.FormSideHtml, prefixes) && isHtmlSafeFor(theme.FooterHtml, prefixes)
		if !isSafe {
			return fmt.Errorf("the theme: %s uses HTML or loads files that this Casdoor only lets a global admin set", item.Id)
		}
	}

	backup, err := json.Marshal(themeFields{
		ThemeData:     application.ThemeData,
		FormOffset:    application.FormOffset,
		FormCss:       application.FormCss,
		FormCssMobile: application.FormCssMobile,
		FormSideHtml:  application.FormSideHtml,
		FooterHtml:    application.FooterHtml,

		FormBackgroundUrl:       application.FormBackgroundUrl,
		FormBackgroundUrlMobile: application.FormBackgroundUrlMobile,
	})
	if err != nil {
		return err
	}

	setApplicationTheme(application, &theme)
	_, err = UpdateApplication(application.GetId(), application, isGlobalAdmin, lang, themeColumns)
	if err != nil {
		return err
	}

	integration.Application = application.Name
	integration.Backup = string(backup)
	return nil
}

func setApplicationTheme(application *Application, theme *themeFields) {
	application.ThemeData = theme.ThemeData
	application.FormOffset = theme.FormOffset
	application.FormCss = theme.FormCss
	application.FormCssMobile = theme.FormCssMobile
	application.FormSideHtml = theme.FormSideHtml
	application.FooterHtml = theme.FooterHtml
	application.FormBackgroundUrl = theme.FormBackgroundUrl
	application.FormBackgroundUrlMobile = theme.FormBackgroundUrlMobile
}

// InstallIntegration downloads an integration from the Marketplace, checks it against the index and creates
// the application or provider it describes, or applies the theme to an application.
func InstallIntegration(req *IntegrationInstallRequest, isGlobalAdmin bool, host string, lang string) (*Integration, error) {
	if req.Owner == "" || !installNameRegex.MatchString(req.Name) {
		return nil, fmt.Errorf("the name may only contain letters, digits, - and _")
	}

	organization, err := getOrganization("admin", req.Owner)
	if err != nil {
		return nil, err
	}
	if organization == nil {
		return nil, fmt.Errorf("the organization: %s does not exist", req.Owner)
	}

	existing, err := getIntegration(req.Owner, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("an integration named %s is already installed in %s", req.Name, req.Owner)
	}

	item, err := GetMarketplaceIntegration(req.IntegrationId)
	if err != nil {
		return nil, err
	}

	bundle, bundleText, err := getIntegrationBundle(item, req.Version)
	if err != nil {
		return nil, err
	}

	variables, err := getIntegrationVariables(bundle, req.Variables)
	if err != nil {
		return nil, err
	}
	_, originBackend := getOriginFromHost(host)
	variables["casdoorUrl"] = originBackend

	displayName := bundle.Manifest.Name["en"]
	if displayName == "" {
		displayName = item.Id
	}

	integration := &Integration{
		Owner:         req.Owner,
		Name:          req.Name,
		CreatedTime:   util.GetCurrentTime(),
		UpdatedTime:   util.GetCurrentTime(),
		DisplayName:   displayName,
		IntegrationId: item.Id,
		Type:          item.Type,
		Version:       bundle.Manifest.Version,
		Logo:          item.Logo,
		Variables:     variables,
		Bundle:        bundleText,
	}

	switch item.Type {
	case "app":
		err = installAppIntegration(integration, bundle, variables, lang)
	case "provider":
		err = installProviderIntegration(integration, bundle, variables, req, isGlobalAdmin, lang)
	case "theme":
		err = installThemeIntegration(integration, item, bundle, variables, req, isGlobalAdmin, lang)
	default:
		err = fmt.Errorf("integrations of type: %s cannot be installed yet", item.Type)
	}
	if err != nil {
		return nil, err
	}

	_, err = addIntegration(integration)
	if err != nil {
		return nil, err
	}
	return integration, nil
}

// UninstallIntegration deletes what the integration created, or restores the theme it replaced.
func UninstallIntegration(integration *Integration, isGlobalAdmin bool, lang string) (bool, error) {
	integration, err := getIntegration(integration.Owner, integration.Name)
	if err != nil {
		return false, err
	}
	if integration == nil {
		return false, nil
	}

	var application *Application
	if integration.Application != "" {
		application, err = getApplication("admin", integration.Application)
		if err != nil {
			return false, err
		}
		if application != nil && application.Organization != integration.Owner {
			application = nil
		}
	}

	switch integration.Type {
	case "app":
		if application != nil {
			_, err = DeleteApplication(application)
		}
	case "provider":
		if application != nil {
			providers := []*ProviderItem{}
			for _, providerItem := range application.Providers {
				if providerItem.Name != integration.Provider {
					providers = append(providers, providerItem)
				}
			}
			application.Providers = providers
			_, err = UpdateApplication(application.GetId(), application, isGlobalAdmin, lang, []string{"providers"})
			if err != nil {
				return false, err
			}
		}
		if integration.Provider != "" {
			_, err = DeleteProvider(&Provider{Owner: integration.Owner, Name: integration.Provider})
		}
	case "theme":
		if application != nil && integration.Backup != "" {
			theme := themeFields{}
			err = json.Unmarshal([]byte(integration.Backup), &theme)
			if err != nil {
				return false, err
			}
			setApplicationTheme(application, &theme)
			_, err = UpdateApplication(application.GetId(), application, isGlobalAdmin, lang, themeColumns)
			if err == nil && theme.ThemeData == nil {
				// the update skips a nil JSON column, so put back the missing theme explicitly
				_, err = ormer.Engine.ID(core.PK{application.Owner, application.Name}).Cols("theme_data").Nullable("theme_data").Update(&Application{})
			}
		}
	}
	if err != nil {
		return false, err
	}

	return deleteIntegration(integration)
}
