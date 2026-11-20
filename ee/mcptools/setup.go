// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package mcptools

import (
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// framework is a kind of app that a developer adds Casdoor sign-in to, and how
type framework struct {
	Name     string
	Label    string
	Callback string
	Docs     string
	// Guide returns the steps to add sign-in, given the application and where Casdoor is
	Guide func(g *guideInfo) string
}

type guideInfo struct {
	Organization string
	Application  string
	ClientId     string
	ClientSecret string
	Certificate  string
	Frontend     string
	Backend      string
	RedirectUri  string
}

var frameworks = []*framework{
	{Name: "react", Label: "React, Vue, Angular or another single-page app (casdoor-js-sdk, PKCE)", Callback: "/callback", Docs: "https://casdoor.ai/docs/how-to-connect/sdk", Guide: spaGuide},
	{Name: "nextjs", Label: "Next.js (casdoor-js-sdk)", Callback: "/callback", Docs: "https://casdoor.ai/docs/how-to-connect/nextjs", Guide: spaGuide},
	{Name: "nodejs", Label: "Node.js backend, e.g. Express (casdoor-nodejs-sdk)", Callback: "/callback", Docs: "https://casdoor.ai/docs/how-to-connect/sdk", Guide: nodeGuide},
	{Name: "go", Label: "Go backend (casdoor-go-sdk)", Callback: "/callback", Docs: "https://casdoor.ai/docs/how-to-connect/sdk", Guide: goGuide},
	{Name: "python", Label: "Python backend, e.g. Flask (casdoor)", Callback: "/callback", Docs: "https://casdoor.ai/docs/how-to-connect/sdk", Guide: pythonGuide},
	{Name: "spring-boot", Label: "Spring Boot (casdoor-spring-boot-starter)", Callback: "/callback", Docs: "https://casdoor.ai/docs/how-to-connect/sdk", Guide: springGuide},
	{Name: "dotnet", Label: "ASP.NET Core (Casdoor.AspNetCore)", Callback: "/casdoor/signin-callback", Docs: "https://casdoor.ai/docs/how-to-connect/sdk", Guide: dotnetGuide},
	{Name: "oidc", Label: "Any OpenID Connect client library", Callback: "/callback", Docs: "https://casdoor.ai/docs/how-to-connect/oidc-client", Guide: oidcGuide},
}

func findFramework(name string) *framework {
	for _, f := range frameworks {
		if f.Name == strings.ToLower(name) {
			return f
		}
	}
	return nil
}

func frameworkProperty() map[string]interface{} {
	names := []string{}
	labels := []string{}
	for _, f := range frameworks {
		names = append(names, f.Name)
		labels = append(labels, fmt.Sprintf("%s: %s", f.Name, f.Label))
	}
	return map[string]interface{}{
		"type":        "string",
		"enum":        names,
		"description": "The kind of app: " + strings.Join(labels, "; "),
	}
}

var applicationIdProperty = map[string]interface{}{
	"type":        "string",
	"description": "The ID of the application: \"admin/<name>\"",
}

func getSetupTools() []*toolDef {
	return []*toolDef{
		{
			Name: "setup_application",
			Description: "Create an application for an app that will sign users in with Casdoor, with the defaults the console uses and the redirect URL of the app's framework, " +
				"and return its client ID, client secret and certificate with the install command, settings and code to add sign-in to the app. " +
				"Use this rather than add_application when a developer wants to connect an app.",
			Scope: "application:write",
			Write: true,
			Properties: map[string]interface{}{
				"organization": stringProperty("The organization whose users sign in to the app"),
				"name":         stringProperty("The name of the application, letters, digits, \"-\" and \"_\""),
				"displayName":  stringProperty("The name shown on the sign-in page"),
				"framework":    frameworkProperty(),
				"homepageUrl":  stringProperty("Where the app runs, e.g. http://localhost:3000, the redirect URL is made from it"),
				"redirectUris": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "The redirect URLs, instead of the one made from homepageUrl"},
				"enableSignUp": map[string]interface{}{"type": "boolean", "description": "Whether people can sign up on the sign-in page, true by default"},
			},
			Required: []string{"organization", "name", "framework"},
			Organization: func(t *toolCall) (string, error) {
				return t.getString("organization"), nil
			},
			Run: runSetupApplication,
		},
		{
			Name:        "get_integration_guide",
			Description: "Return the client ID, client secret, certificate, install command, settings and code to add Casdoor sign-in to an app with an existing application.",
			Scope:       "application:read",
			Properties: map[string]interface{}{
				"application": applicationIdProperty,
				"framework":   frameworkProperty(),
			},
			Required: []string{"application", "framework"},
			Organization: func(t *toolCall) (string, error) {
				return t.getKindOrganization(kinds["application"], t.getString("application")), nil
			},
			Run: func(t *toolCall) (string, error) {
				f := findFramework(t.getString("framework"))
				if f == nil {
					return "", fmt.Errorf("unknown framework: %s", t.getString("framework"))
				}
				id, err := t.requireString("application")
				if err != nil {
					return "", err
				}
				application, err := t.getAdministeredApplication(id)
				if err != nil {
					return "", err
				}
				return t.getGuide(application, f), nil
			},
		},
		{
			Name: "add_login_provider",
			Description: "Let the users of an application sign in with a third-party account, e.g. GitHub or Google: add an OAuth provider with the client ID and secret " +
				"from the third party (or reuse an existing provider) and put it on the application's sign-in page. Returns the callback URL to register at the third party.",
			Scope:      "provider:write",
			MoreScopes: []string{"application:write"},
			Write:      true,
			Properties: map[string]interface{}{
				"application":  applicationIdProperty,
				"type":         map[string]interface{}{"type": "string", "enum": loginProviderTypes(), "description": "The third party"},
				"clientId":     stringProperty("The client ID (app ID, app key) that the third party gave"),
				"clientSecret": stringProperty("The client secret (app secret) that the third party gave"),
				"provider":     stringProperty("The ID (\"<owner>/<name>\") of an existing provider to use instead of adding one"),
			},
			Required: []string{"application"},
			Organization: func(t *toolCall) (string, error) {
				return t.getKindOrganization(kinds["application"], t.getString("application")), nil
			},
			Run: runAddLoginProvider,
		},
		{
			Name: "diagnose_login",
			Description: "Find out why users can't sign in to an application: check its settings (redirect URLs, grant types, sign-in methods, cert, providers) " +
				"and summarize the recent failed requests of the organization that concern it, with how to fix each.",
			Scope: "application:read",
			Properties: map[string]interface{}{
				"application": applicationIdProperty,
				"user":        stringProperty("Only the failures of this user (name), optional"),
				"minutes":     map[string]interface{}{"type": "integer", "description": "How far back to look, 1440 (a day) by default"},
			},
			Required: []string{"application"},
			Organization: func(t *toolCall) (string, error) {
				return t.getKindOrganization(kinds["application"], t.getString("application")), nil
			},
			Run: runDiagnoseLogin,
		},
	}
}

// getAdministeredApplication loads an application whose organization the user administers: the API shows the
// applications of other organizations masked, the tools don't work on them
func (t *toolCall) getAdministeredApplication(id string) (map[string]interface{}, error) {
	application, err := t.client.mustGet(kinds["application"], id)
	if err != nil {
		return nil, err
	}
	if _, err = t.client.mustGet(kinds["organization"], util.GetId("admin", getString(application, "organization"))); err != nil {
		return nil, err
	}
	return application, nil
}

func runSetupApplication(t *toolCall) (string, error) {
	f := findFramework(t.getString("framework"))
	if f == nil {
		return "", fmt.Errorf("unknown framework: %s", t.getString("framework"))
	}
	organization, err := t.requireString("organization")
	if err != nil {
		return "", err
	}
	name, err := t.requireString("name")
	if err != nil {
		return "", err
	}
	if _, err = t.client.mustGet(kinds["organization"], util.GetId("admin", organization)); err != nil {
		return "", err
	}

	homepageUrl := strings.TrimRight(t.getString("homepageUrl"), "/")
	redirectUris := t.getStrings("redirectUris")
	if len(redirectUris) == 0 {
		if homepageUrl == "" {
			return "", fmt.Errorf("give homepageUrl or redirectUris")
		}
		redirectUris = []string{homepageUrl + f.Callback}
	}
	displayName := t.getString("displayName")
	if displayName == "" {
		displayName = name
	}
	enableSignUp := true
	if len(t.args["enableSignUp"]) > 0 {
		enableSignUp = t.getBool("enableSignUp")
	}

	application := newApplication(organization, name, displayName, homepageUrl, redirectUris, enableSignUp)
	note := ""
	// sign-up sends an email verification code, which needs an Email provider on the application
	if emailProvider := t.findEmailProvider(organization); emailProvider != "" {
		application["providers"] = []map[string]interface{}{
			{"owner": "", "name": emailProvider, "canSignUp": false, "canSignIn": false, "canUnlink": false, "prompted": false, "signupGroup": "", "rule": "All"},
		}
	} else if enableSignUp {
		setSignupItemRule(application, "Email", "No verification")
		note = "\n\nThe organization has no Email provider, so sign-up doesn't verify email addresses. Add an Email provider to the application and set the Email sign-up item's rule to \"Normal\" to verify them."
	}

	if err = t.apply(kinds["application"], actionAdd, "", nil, application); err != nil {
		return "", err
	}
	if t.dryRun {
		return "The client ID and secret are made when the application is added." + note, nil
	}

	saved, err := t.client.mustGet(kinds["application"], getObjectId(application))
	if err != nil {
		return "", err
	}
	return "Created the application.\n\n" + t.getGuide(saved, f) + note, nil
}

// findEmailProvider is the name of an Email provider that the organization's applications can use, or ""
func (t *toolCall) findEmailProvider(organization string) string {
	for _, owner := range []string{organization, "admin"} {
		providers, _, err := t.client.list(kinds["provider"], owner, 1, 200)
		if err != nil {
			continue
		}
		for _, provider := range providers {
			if getString(provider, "category") == "Email" {
				return getString(provider, "name")
			}
		}
	}
	return ""
}

func setSignupItemRule(application map[string]interface{}, name string, rule string) {
	items, _ := application["signupItems"].([]map[string]interface{})
	for _, item := range items {
		if item["name"] == name {
			item["rule"] = rule
		}
	}
}

// newApplication is an application with the defaults of the console's "Add" button
func newApplication(organization string, name string, displayName string, homepageUrl string, redirectUris []string, enableSignUp bool) map[string]interface{} {
	return map[string]interface{}{
		"owner":          "admin",
		"name":           name,
		"organization":   organization,
		"createdTime":    util.GetCurrentTime(),
		"displayName":    displayName,
		"category":       "Default",
		"type":           "All",
		"logo":           fmt.Sprintf("%s/img/casdoor-logo_1185x256.png", conf.GetConfigString("staticBaseUrl")),
		"homepageUrl":    homepageUrl,
		"enablePassword": true,
		"enableSignUp":   enableSignUp,
		"scopes":         []interface{}{},
		"providers":      []interface{}{},
		"signinMethods": []map[string]interface{}{
			{"name": "Password", "displayName": "Password", "rule": "All"},
			{"name": "Verification code", "displayName": "Verification code", "rule": "All"},
			{"name": "WebAuthn", "displayName": "WebAuthn", "rule": "None"},
		},
		"signupItems": []map[string]interface{}{
			{"name": "ID", "visible": false, "required": true, "rule": "Random"},
			{"name": "Username", "visible": true, "required": true, "rule": "None"},
			{"name": "Display name", "visible": true, "required": true, "rule": "None"},
			{"name": "Password", "visible": true, "required": true, "rule": "None"},
			{"name": "Confirm password", "visible": true, "required": true, "rule": "None"},
			{"name": "Email", "visible": true, "required": true, "rule": "Normal"},
			{"name": "Agreement", "visible": true, "required": true, "rule": "None"},
			{"name": "Signup button", "visible": true, "required": true, "rule": "None"},
			{"name": "Providers", "visible": true, "required": true, "rule": "None"},
		},
		"grantTypes":           []string{"authorization_code", "refresh_token"},
		"cert":                 "cert-built-in",
		"redirectUris":         redirectUris,
		"tokenFormat":          "",
		"tokenFields":          []interface{}{},
		"expireInHours":        24 * 7,
		"refreshExpireInHours": 24 * 7,
		"cookieExpireInHours":  24 * 30,
		"formOffset":           2,
	}
}

func (t *toolCall) getGuide(application map[string]interface{}, f *framework) string {
	frontend, backend := object.GetOriginFromHost(t.c.Ctx.Request.Host)
	g := &guideInfo{
		Organization: getString(application, "organization"),
		Application:  getString(application, "name"),
		ClientId:     getString(application, "clientId"),
		ClientSecret: getString(application, "clientSecret"),
		Certificate:  strings.TrimSpace(getString(application, "certPublicKey")),
		Frontend:     frontend,
		Backend:      backend,
		RedirectUri:  "<the app's URL>" + f.Callback,
	}
	redirectUris := []string{}
	if values, ok := application["redirectUris"].([]interface{}); ok {
		for _, value := range values {
			if text, ok := value.(string); ok {
				redirectUris = append(redirectUris, text)
			}
		}
	}
	for _, uri := range redirectUris {
		if strings.HasSuffix(uri, f.Callback) {
			g.RedirectUri = uri
			break
		}
	}

	lines := []string{
		fmt.Sprintf("Application: admin/%s (organization %s)", g.Application, g.Organization),
		"Casdoor: " + g.Frontend,
		"Client ID: " + g.ClientId,
		"Client secret: " + g.ClientSecret,
		"Redirect URLs: " + strings.Join(redirectUris, ", "),
		fmt.Sprintf("OpenID Connect discovery: %s/.well-known/openid-configuration", g.Backend),
	}
	if g.RedirectUri == "<the app's URL>"+f.Callback {
		lines = append(lines, fmt.Sprintf("Note: no redirect URL ends with %s, add the app's callback URL to the application with update_application.", f.Callback))
	}
	lines = append(lines, "", "## "+f.Label, "", f.Guide(g))
	if g.Certificate != "" && f.Name != "react" && f.Name != "nextjs" && f.Name != "dotnet" && f.Name != "oidc" {
		lines = append(lines, "", "Save the certificate that verifies the tokens as casdoor.pem:", "```", g.Certificate, "```")
	}
	lines = append(lines, "", "Keep the client secret out of the browser and out of the repository. Docs: "+f.Docs)
	return strings.Join(lines, "\n")
}

func spaGuide(g *guideInfo) string {
	return fmt.Sprintf("```shell\nnpm install casdoor-js-sdk\n```\n\n```js\nimport Sdk from \"casdoor-js-sdk\";\n\nconst sdk = new Sdk({\n  serverUrl: %q,\n  clientId: %q,\n  appName: %q,\n  organizationName: %q,\n  redirectPath: \"/callback\",\n});\n\n"+
		"// the sign-in button\nsdk.signin_redirect();\n\n// on the /callback page: exchange the code with PKCE, no client secret in the browser\nconst token = await sdk.exchangeForAccessToken();\nconst user = await sdk.getUserInfo(token.access_token);\n```\n\n"+
		"The redirect URL is the page's origin + redirectPath, it must be one of the application's redirect URLs.",
		g.Frontend, g.ClientId, g.Application, g.Organization)
}

func nodeGuide(g *guideInfo) string {
	return fmt.Sprintf("```shell\nnpm install casdoor-nodejs-sdk\n```\n\n```js\nconst fs = require(\"fs\");\nconst express = require(\"express\");\nconst { SDK } = require(\"casdoor-nodejs-sdk\");\n\n"+
		"const sdk = new SDK({\n  endpoint: %q,\n  clientId: process.env.CASDOOR_CLIENT_ID,\n  clientSecret: process.env.CASDOOR_CLIENT_SECRET,\n  certificate: fs.readFileSync(\"casdoor.pem\", \"utf8\"),\n  orgName: %q,\n  appName: %q,\n});\n\n"+
		"const app = express();\napp.get(\"/login\", (req, res) => res.redirect(sdk.getSignInUrl(%q)));\napp.get(\"/callback\", async (req, res) => {\n  const token = await sdk.getAuthToken(req.query.code);\n  const user = sdk.parseJwtToken(token.access_token);\n  res.json(user);\n});\n```\n\n"+
		"```shell\nCASDOOR_CLIENT_ID=%s\nCASDOOR_CLIENT_SECRET=%s\n```",
		g.Backend, g.Organization, g.Application, g.RedirectUri, g.ClientId, g.ClientSecret)
}

func goGuide(g *guideInfo) string {
	return fmt.Sprintf("```shell\ngo get github.com/casdoor/casdoor-go-sdk@latest\n```\n\n```go\nimport \"github.com/casdoor/casdoor-go-sdk/casdoorsdk\"\n\n"+
		"func init() {\n\tcertificate, _ := os.ReadFile(\"casdoor.pem\")\n\tcasdoorsdk.InitConfig(%q, os.Getenv(\"CASDOOR_CLIENT_ID\"), os.Getenv(\"CASDOOR_CLIENT_SECRET\"), string(certificate), %q, %q)\n}\n\n"+
		"func login(w http.ResponseWriter, r *http.Request) {\n\thttp.Redirect(w, r, casdoorsdk.GetSigninUrl(%q), http.StatusFound)\n}\n\n"+
		"func callback(w http.ResponseWriter, r *http.Request) {\n\ttoken, err := casdoorsdk.GetOAuthToken(r.URL.Query().Get(\"code\"), r.URL.Query().Get(\"state\"))\n\tif err != nil {\n\t\thttp.Error(w, err.Error(), http.StatusBadRequest)\n\t\treturn\n\t}\n"+
		"\tclaims, err := casdoorsdk.ParseJwtToken(token.AccessToken)\n\tif err != nil {\n\t\thttp.Error(w, err.Error(), http.StatusUnauthorized)\n\t\treturn\n\t}\n\tfmt.Fprintf(w, \"hello %%s\", claims.User.Name)\n}\n```\n\n"+
		"```shell\nCASDOOR_CLIENT_ID=%s\nCASDOOR_CLIENT_SECRET=%s\n```",
		g.Backend, g.Organization, g.Application, g.RedirectUri, g.ClientId, g.ClientSecret)
}

func pythonGuide(g *guideInfo) string {
	return fmt.Sprintf("```shell\npip install casdoor\n```\n\n```python\nimport os\nfrom casdoor import CasdoorSDK\n\n"+
		"sdk = CasdoorSDK(\n    endpoint=%q,\n    client_id=os.environ[\"CASDOOR_CLIENT_ID\"],\n    client_secret=os.environ[\"CASDOOR_CLIENT_SECRET\"],\n    certificate=open(\"casdoor.pem\").read(),\n    org_name=%q,\n    application_name=%q,\n    front_endpoint=%q,\n)\n\n"+
		"# the sign-in link, with a random state kept in the session\nurl = sdk.get_auth_link(%q, state=state)\n\n# on the callback\ntoken = sdk.get_oauth_token(code=code)\nuser = sdk.parse_jwt_token(token[\"access_token\"])\n```\n\n"+
		"Always pass front_endpoint, otherwise the SDK guesses the sign-in page's port.\n\n```shell\nCASDOOR_CLIENT_ID=%s\nCASDOOR_CLIENT_SECRET=%s\n```",
		g.Backend, g.Organization, g.Application, g.Frontend, g.RedirectUri, g.ClientId, g.ClientSecret)
}

func springGuide(g *guideInfo) string {
	return fmt.Sprintf("Add the dependency org.casbin:casdoor-spring-boot-starter (the latest version on Maven Central, Spring Boot 3 and Java 17 or later).\n\n"+
		"```yaml\ncasdoor:\n  endpoint: %s\n  client-id: ${CASDOOR_CLIENT_ID}\n  client-secret: ${CASDOOR_CLIENT_SECRET}\n  organization-name: %s\n  application-name: %s\n  certificate: |\n%s\n```\n\n"+
		"```java\n@RestController\nclass AuthController {\n    @Resource\n    private AuthService authService;\n\n    @GetMapping(\"/login\")\n    void login(HttpServletResponse response) throws IOException {\n        response.sendRedirect(authService.getSigninUrl(%q));\n    }\n\n"+
		"    @GetMapping(\"/callback\")\n    String callback(@RequestParam String code, @RequestParam String state) {\n        String accessToken = authService.getOAuthToken(code, state);\n        User user = authService.parseJwtToken(accessToken);\n        return user.name;\n    }\n}\n```\n\n"+
		"```shell\nCASDOOR_CLIENT_ID=%s\nCASDOOR_CLIENT_SECRET=%s\n```",
		g.Backend, g.Organization, g.Application, indent(g.Certificate, "    "), g.RedirectUri, g.ClientId, g.ClientSecret)
}

func dotnetGuide(g *guideInfo) string {
	return fmt.Sprintf("```shell\ndotnet add package Casdoor.AspNetCore\n```\n\n```json\n\"Casdoor\": {\n  \"Endpoint\": %q,\n  \"OrganizationName\": %q,\n  \"ApplicationName\": %q,\n  \"ApplicationType\": \"webapp\",\n  \"ClientId\": %q,\n  \"ClientSecret\": %q,\n  \"CallbackPath\": \"/casdoor/signin-callback\"\n}\n```\n\n"+
		"```csharp\nbuilder.Services.AddAuthentication(CookieAuthenticationDefaults.AuthenticationScheme)\n    .AddCasdoor(builder.Configuration.GetSection(\"Casdoor\"))\n    .AddCookie(CookieAuthenticationDefaults.AuthenticationScheme);\n```\n\n"+
		"ApplicationType must be \"webapp\". Keep the client secret in user secrets or an environment variable (Casdoor__ClientSecret) rather than appsettings.json.",
		g.Backend, g.Organization, g.Application, g.ClientId, g.ClientSecret)
}

func oidcGuide(g *guideInfo) string {
	return fmt.Sprintf("Issuer: %s\nDiscovery: %s/.well-known/openid-configuration\nAuthorization endpoint: %s/login/oauth/authorize\nToken endpoint: %s/api/login/oauth/access_token\nUserinfo endpoint: %s/api/userinfo\nJWKS: %s/.well-known/jwks\n"+
		"Client ID: %s\nClient secret: %s\nRedirect URL: %s\nScopes: openid profile email\nGrant type: authorization_code, with PKCE for public clients",
		g.Backend, g.Backend, g.Frontend, g.Backend, g.Backend, g.Backend, g.ClientId, g.ClientSecret, g.RedirectUri)
}

func indent(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

// loginProviders are the third parties that add_login_provider sets up with a client ID and secret, and where to
// register the callback URL at each
var loginProviders = map[string]string{
	"GitHub":   "https://github.com/settings/developers (OAuth Apps, \"Authorization callback URL\")",
	"Google":   "https://console.cloud.google.com/apis/credentials (OAuth client ID, Web application, \"Authorized redirect URIs\")",
	"GitLab":   "https://gitlab.com/-/user_settings/applications (\"Redirect URI\", scopes read_user, openid, email)",
	"Gitee":    "https://gitee.com/oauth/applications (\"应用回调地址\")",
	"Facebook": "https://developers.facebook.com/apps (Facebook Login, \"Valid OAuth Redirect URIs\")",
	"LinkedIn": "https://www.linkedin.com/developers/apps (Auth, \"Authorized redirect URLs\")",
	"DingTalk": "https://open-dev.dingtalk.com (应用的登录与分享，回调域名)",
	"Lark":     "https://open.feishu.cn/app (安全设置，重定向 URL)",
	"WeChat":   "https://open.weixin.qq.com (网站应用，授权回调域填 Casdoor 的域名)",
	"QQ":       "https://connect.qq.com (网站应用，网站回调域)",
	"Weibo":    "https://open.weibo.com (高级信息，授权回调页)",
}

func loginProviderTypes() []string {
	return []string{"GitHub", "Google", "GitLab", "Gitee", "Facebook", "LinkedIn", "DingTalk", "Lark", "WeChat", "QQ", "Weibo"}
}

func runAddLoginProvider(t *toolCall) (string, error) {
	applicationId, err := t.requireString("application")
	if err != nil {
		return "", err
	}
	application, err := t.getAdministeredApplication(applicationId)
	if err != nil {
		return "", err
	}
	organization := getString(application, "organization")

	var provider map[string]interface{}
	if providerId := t.getString("provider"); providerId != "" {
		provider, err = t.client.mustGet(kinds["provider"], providerId)
		if err != nil {
			return "", err
		}
		// an application finds its providers by name in its organization and among the global ones
		if owner := getString(provider, "owner"); owner != organization && owner != "admin" {
			return "", fmt.Errorf("the provider %s belongs to another organization, the application of %s can only use its own or global providers", providerId, organization)
		}
	} else {
		providerType := t.getString("type")
		if _, ok := loginProviders[providerType]; !ok {
			return "", fmt.Errorf("type must be one of: %s, or give an existing provider", strings.Join(loginProviderTypes(), ", "))
		}
		clientId, err := t.requireString("clientId")
		if err != nil {
			return "", err
		}
		clientSecret, err := t.requireString("clientSecret")
		if err != nil {
			return "", err
		}
		provider = map[string]interface{}{
			"owner":        organization,
			"name":         fmt.Sprintf("provider_%s_%s", strings.ToLower(providerType), util.GetRandomName()),
			"createdTime":  util.GetCurrentTime(),
			"displayName":  providerType,
			"category":     "OAuth",
			"type":         providerType,
			"method":       "Normal",
			"clientId":     clientId,
			"clientSecret": clientSecret,
			"enableSignUp": true,
		}
		if err = t.apply(kinds["provider"], actionAdd, "", nil, provider); err != nil {
			return "", err
		}
	}

	providerName := getString(provider, "name")
	items, _ := application["providers"].([]interface{})
	for _, item := range items {
		if obj, ok := item.(map[string]interface{}); ok && getString(obj, "name") == providerName {
			return "", fmt.Errorf("the application already has the provider %s", providerName)
		}
	}

	updated := map[string]interface{}{}
	for key, value := range application {
		updated[key] = value
	}
	updated["providers"] = append(append([]interface{}{}, items...), map[string]interface{}{
		"owner": "", "name": providerName, "canSignUp": true, "canSignIn": true, "canUnlink": true,
		"prompted": false, "signupGroup": "", "rule": "None",
	})
	if err = t.apply(kinds["application"], actionUpdate, applicationId, application, updated); err != nil {
		return "", err
	}

	frontend, _ := object.GetOriginFromHost(t.c.Ctx.Request.Host)
	lines := []string{fmt.Sprintf("Register this callback URL at the third party: %s/callback", frontend)}
	if !t.dryRun {
		lines = append([]string{fmt.Sprintf("The application %s shows \"Sign in with %s\" with the provider %s.", applicationId, getString(provider, "type"), getObjectId(provider))}, lines...)
	}
	if where, ok := loginProviders[getString(provider, "type")]; ok {
		lines = append(lines, "Where: "+where)
	}
	return strings.Join(lines, "\n"), nil
}
