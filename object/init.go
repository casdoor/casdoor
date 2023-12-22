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
	"encoding/gob"
	"fmt"
	"os"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/go-webauthn/webauthn/webauthn"
)

func InitDb() {
	existed := initBuiltInOrganization()
	if !existed {
		initBuiltInPermission()
		initBuiltInProvider()
		initBuiltInUser()
		initBuiltInApplication()
		initBuiltInCert()
		initBuiltInLdap()
	}

	existed = initBuiltInApiModel()
	if !existed {
		initBuiltInApiAdapter()
		initBuiltInApiEnforcer()
		initBuiltInUserModel()
		initBuiltInUserAdapter()
		initBuiltInUserEnforcer()
	}

	initWebAuthn()
}

func getBuiltInAccountItems() []*AccountItem {
	return []*AccountItem{
		{Name: "Organization", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "ID", Visible: true, ViewRule: "Public", ModifyRule: "Immutable"},
		{Name: "Name", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Display name", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Avatar", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "User type", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "Password", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
		{Name: "Email", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Phone", Visible: true, ViewRule: "Public", ModifyRule: "Self"},
		{Name: "Country code", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
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
		{Name: "Groups", Visible: true, ViewRule: "Public", ModifyRule: "Admin"},
		{Name: "3rd-party logins", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
		{Name: "Properties", Visible: false, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is admin", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is forbidden", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Is deleted", Visible: true, ViewRule: "Admin", ModifyRule: "Admin"},
		{Name: "Multi-factor authentication", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
		{Name: "WebAuthn credentials", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
		{Name: "Managed accounts", Visible: true, ViewRule: "Self", ModifyRule: "Self"},
	}
}

func initBuiltInOrganization() bool {
	organization, err := getOrganization("admin", "built-in")
	if err != nil {
		panic(err)
	}

	if organization != nil {
		return true
	}

	organization = &Organization{
		Owner:              "admin",
		Name:               "built-in",
		CreatedTime:        util.GetCurrentTime(),
		DisplayName:        "Built-in Organization",
		WebsiteUrl:         "https://example.com",
		Favicon:            fmt.Sprintf("%s/img/casbin/favicon.ico", conf.GetConfigString("staticBaseUrl")),
		PasswordType:       "plain",
		PasswordOptions:    []string{"AtLeast6"},
		CountryCodes:       []string{"US", "ES", "FR", "DE", "GB", "CN", "JP", "KR", "VN", "ID", "SG", "IN"},
		DefaultAvatar:      fmt.Sprintf("%s/img/casbin.svg", conf.GetConfigString("staticBaseUrl")),
		Tags:               []string{},
		Languages:          []string{"en", "zh", "es", "fr", "de", "id", "ja", "ko", "ru", "vi", "pt"},
		InitScore:          2000,
		AccountItems:       getBuiltInAccountItems(),
		EnableSoftDeletion: false,
		IsProfilePublic:    false,
	}
	_, err = AddOrganization(organization)
	if err != nil {
		panic(err)
	}

	return false
}

func initBuiltInUser() {
	user, err := getUser("built-in", "admin")
	if err != nil {
		panic(err)
	}
	if user != nil {
		return
	}

	user = &User{
		Owner:             "built-in",
		Name:              "admin",
		CreatedTime:       util.GetCurrentTime(),
		Id:                util.GenerateId(),
		Type:              "normal-user",
		Password:          "123",
		DisplayName:       "Admin",
		Avatar:            fmt.Sprintf("%s/img/casbin.svg", conf.GetConfigString("staticBaseUrl")),
		Email:             "admin@example.com",
		Phone:             "12345678910",
		CountryCode:       "US",
		Address:           []string{},
		Affiliation:       "Example Inc.",
		Tag:               "staff",
		Score:             2000,
		Ranking:           1,
		IsAdmin:           true,
		IsForbidden:       false,
		IsDeleted:         false,
		SignupApplication: "app-built-in",
		CreatedIp:         "127.0.0.1",
		Properties:        make(map[string]string),
	}
	_, err = AddUser(user)
	if err != nil {
		panic(err)
	}
}

func initBuiltInApplication() {
	application, err := getApplication("admin", "app-built-in")
	if err != nil {
		panic(err)
	}

	if application != nil {
		return
	}

	application = &Application{
		Owner:          "admin",
		Name:           "app-built-in",
		CreatedTime:    util.GetCurrentTime(),
		DisplayName:    "Casdoor",
		Logo:           fmt.Sprintf("%s/img/casdoor-logo_1185x256.png", conf.GetConfigString("staticBaseUrl")),
		HomepageUrl:    "https://casdoor.org",
		Organization:   "built-in",
		Cert:           "cert-built-in",
		EnablePassword: true,
		EnableSignUp:   true,
		Providers: []*ProviderItem{
			{Name: "provider_captcha_default", CanSignUp: false, CanSignIn: false, CanUnlink: false, Prompted: false, SignupGroup: "", Rule: "None", Provider: nil},
		},
		SignupItems: []*SignupItem{
			{Name: "ID", Visible: false, Required: true, Prompted: false, Rule: "Random"},
			{Name: "Username", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Display name", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Password", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Confirm password", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Email", Visible: true, Required: true, Prompted: false, Rule: "Normal"},
			{Name: "Phone", Visible: true, Required: true, Prompted: false, Rule: "None"},
			{Name: "Agreement", Visible: true, Required: true, Prompted: false, Rule: "None"},
		},
		Tags:          []string{},
		RedirectUris:  []string{},
		ExpireInHours: 168,
		FormOffset:    2,
	}
	_, err = AddApplication(application)
	if err != nil {
		panic(err)
	}
}

func readTokenFromFile() (string, string) {
	pemPath := "./object/token_jwt_key.pem"
	keyPath := "./object/token_jwt_key.key"
	pem, err := os.ReadFile(pemPath)
	if err != nil {
		return "", ""
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return "", ""
	}
	return string(pem), string(key)
}

func initBuiltInCert() {
	tokenJwtCertificate, tokenJwtPrivateKey := readTokenFromFile()
	cert, err := getCert("admin", "cert-built-in")
	if err != nil {
		panic(err)
	}

	if cert != nil {
		return
	}

	cert = &Cert{
		Owner:           "admin",
		Name:            "cert-built-in",
		CreatedTime:     util.GetCurrentTime(),
		DisplayName:     "Built-in Cert",
		Scope:           "JWT",
		Type:            "x509",
		CryptoAlgorithm: "RS256",
		BitSize:         4096,
		ExpireInYears:   20,
		Certificate:     tokenJwtCertificate,
		PrivateKey:      tokenJwtPrivateKey,
	}
	_, err = AddCert(cert)
	if err != nil {
		panic(err)
	}
}

func initBuiltInLdap() {
	ldap, err := GetLdap("ldap-built-in")
	if err != nil {
		panic(err)
	}

	if ldap != nil {
		return
	}

	ldap = &Ldap{
		Id:         "ldap-built-in",
		Owner:      "built-in",
		ServerName: "BuildIn LDAP Server",
		Host:       "example.com",
		Port:       389,
		Username:   "cn=buildin,dc=example,dc=com",
		Password:   "123",
		BaseDn:     "ou=BuildIn,dc=example,dc=com",
		AutoSync:   0,
		LastSync:   "",
	}
	_, err = AddLdap(ldap)
	if err != nil {
		panic(err)
	}
}

func initBuiltInProvider() {
	provider, err := GetProvider(util.GetId("admin", "provider_captcha_default"))
	if err != nil {
		panic(err)
	}

	if provider != nil {
		return
	}

	provider = &Provider{
		Owner:       "admin",
		Name:        "provider_captcha_default",
		CreatedTime: util.GetCurrentTime(),
		DisplayName: "Captcha Default",
		Category:    "Captcha",
		Type:        "Default",
	}
	_, err = AddProvider(provider)
	if err != nil {
		panic(err)
	}
}

func initWebAuthn() {
	gob.Register(webauthn.SessionData{})
}

func initBuiltInUserModel() {
	model, err := GetModel("built-in/user-model-built-in")
	if err != nil {
		panic(err)
	}

	if model != nil {
		return
	}

	model = &Model{
		Owner:       "built-in",
		Name:        "user-model-built-in",
		CreatedTime: util.GetCurrentTime(),
		DisplayName: "Built-in Model",
		ModelText: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`,
	}
	_, err = AddModel(model)
	if err != nil {
		panic(err)
	}
}

func initBuiltInApiModel() bool {
	model, err := GetModel("built-in/api-model-built-in")
	if err != nil {
		panic(err)
	}

	if model != nil {
		return true
	}

	modelText := `[request_definition]
r = subOwner, subName, method, urlPath, objOwner, objName

[policy_definition]
p = subOwner, subName, method, urlPath, objOwner, objName

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (r.subOwner == p.subOwner || p.subOwner == "*") && \
    (r.subName == p.subName || p.subName == "*" || r.subName != "anonymous" && p.subName == "!anonymous") && \
    (r.method == p.method || p.method == "*") && \
    (r.urlPath == p.urlPath || p.urlPath == "*") && \
    (r.objOwner == p.objOwner || p.objOwner == "*") && \
    (r.objName == p.objName || p.objName == "*") || \
    (r.subOwner == r.objOwner && r.subName == r.objName)`

	model = &Model{
		Owner:       "built-in",
		Name:        "api-model-built-in",
		CreatedTime: util.GetCurrentTime(),
		DisplayName: "API Model",
		ModelText:   modelText,
	}
	_, err = AddModel(model)
	if err != nil {
		panic(err)
	}
	return false
}

func initBuiltInPermission() {
	permission, err := GetPermission("built-in/permission-built-in")
	if err != nil {
		panic(err)
	}
	if permission != nil {
		return
	}

	permission = &Permission{
		Owner:        "built-in",
		Name:         "permission-built-in",
		CreatedTime:  util.GetCurrentTime(),
		DisplayName:  "Built-in Permission",
		Description:  "Built-in Permission",
		Users:        []string{"built-in/*"},
		Groups:       []string{},
		Roles:        []string{},
		Domains:      []string{},
		Model:        "model-built-in",
		Adapter:      "",
		ResourceType: "Application",
		Resources:    []string{"app-built-in"},
		Actions:      []string{"Read", "Write", "Admin"},
		Effect:       "Allow",
		IsEnabled:    true,
		Submitter:    "admin",
		Approver:     "admin",
		ApproveTime:  util.GetCurrentTime(),
		State:        "Approved",
	}
	_, err = AddPermission(permission)
	if err != nil {
		panic(err)
	}
}

func initBuiltInUserAdapter() {
	adapter, err := GetAdapter("built-in/user-adapter-built-in")
	if err != nil {
		panic(err)
	}

	if adapter != nil {
		return
	}

	adapter = &Adapter{
		Owner:       "built-in",
		Name:        "user-adapter-built-in",
		CreatedTime: util.GetCurrentTime(),
		Table:       "casbin_user_rule",
		UseSameDb:   true,
	}
	_, err = AddAdapter(adapter)
	if err != nil {
		panic(err)
	}
}

func initBuiltInApiAdapter() {
	adapter, err := GetAdapter("built-in/api-adapter-built-in")
	if err != nil {
		panic(err)
	}

	if adapter != nil {
		return
	}

	adapter = &Adapter{
		Owner:       "built-in",
		Name:        "api-adapter-built-in",
		CreatedTime: util.GetCurrentTime(),
		Table:       "casbin_api_rule",
		UseSameDb:   true,
	}
	_, err = AddAdapter(adapter)
	if err != nil {
		panic(err)
	}
}

func initBuiltInUserEnforcer() {
	enforcer, err := GetEnforcer("built-in/user-enforcer-built-in")
	if err != nil {
		panic(err)
	}

	if enforcer != nil {
		return
	}

	enforcer = &Enforcer{
		Owner:       "built-in",
		Name:        "user-enforcer-built-in",
		CreatedTime: util.GetCurrentTime(),
		DisplayName: "User Enforcer",
		Model:       "built-in/user-model-built-in",
		Adapter:     "built-in/user-adapter-built-in",
	}

	_, err = AddEnforcer(enforcer)
	if err != nil {
		panic(err)
	}
}

func initBuiltInApiEnforcer() {
	enforcer, err := GetEnforcer("built-in/api-enforcer-built-in")
	if err != nil {
		panic(err)
	}

	if enforcer != nil {
		return
	}

	enforcer = &Enforcer{
		Owner:       "built-in",
		Name:        "api-enforcer-built-in",
		CreatedTime: util.GetCurrentTime(),
		DisplayName: "API Enforcer",
		Model:       "built-in/api-model-built-in",
		Adapter:     "built-in/api-adapter-built-in",
	}

	_, err = AddEnforcer(enforcer)
	if err != nil {
		panic(err)
	}
}
