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
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/casdoor/casdoor/util"
)

// newAppRoleTestDb adds to the grant test database what the application role tests use: a nested
// group of the resource organization (resource-org/ops/grafana-admins), a Casdoor role
// (resource-org/viewer) and an application with custom claims (claims-like).
func newAppRoleTestDb(t *testing.T) {
	t.Helper()
	newApplicationGrantTestDb(t)

	for _, group := range []*Group{
		{Owner: grantTestResourceOrg, Name: "ops", ParentId: grantTestResourceOrg, IsTopGroup: true, IsEnabled: true},
		{Owner: grantTestResourceOrg, Name: "grafana-admins", ParentId: "ops", IsEnabled: true},
		{Owner: grantTestResourceOrg, Name: "readers", ParentId: grantTestResourceOrg, IsTopGroup: true, IsEnabled: true},
	} {
		group.CreatedTime = util.GetCurrentTime()
		mustAdd(t, AddGroup, group)
	}
	mustAdd(t, AddRole, &Role{Owner: grantTestResourceOrg, Name: "viewer", CreatedTime: util.GetCurrentTime(), IsEnabled: true})
	mustAdd(t, AddRole, &Role{Owner: grantTestPartnerOrg, Name: "viewer", CreatedTime: util.GetCurrentTime(), IsEnabled: true})

	mustAdd(t, func(a *Application) (bool, error) { return AddApplication(a, "en") }, &Application{
		Owner: "admin", Name: "claims-like", Organization: grantTestResourceOrg, ClientId: "client-claims", ClientSecret: "secret-claims",
		RedirectUris: []string{grantTestRedirectUri}, GrantTypes: []string{"authorization_code", "refresh_token"},
		TokenFormat: "JWT-Custom", TokenFields: []string{"Name", "Groups", "Roles"}, TokenGroupFormat: "", ExpireInHours: 1, RefreshExpireInHours: 1,
		TokenSigningMethod: "RS256", CreatedTime: util.GetCurrentTime(),
		TokenAttributes: []*JwtItem{{Name: "tenant", Category: "Static Value", Value: "resource", Type: "String"}},
	})
}

func setAppRoles(t *testing.T, applicationName string, roles ...*AppRole) {
	t.Helper()
	if _, err := UpdateApplicationAppRoles(grantTestApplication(t, applicationName), roles, "en"); err != nil {
		t.Fatalf("set application roles of %s: %v", applicationName, err)
	}
}

func addTestGrantWithRoles(t *testing.T, applicationName string, principalType string, principal string, roles ...string) *ApplicationGrant {
	t.Helper()
	application := grantTestApplication(t, applicationName)
	grant := &ApplicationGrant{Owner: application.Organization, Application: application.Name, PrincipalType: principalType, Principal: principal, Roles: roles}
	if err := CheckApplicationGrant(grant, application, "en"); err != nil {
		t.Fatalf("check grant: %v", err)
	}
	grant.CreatedBy = "resource-org/internal-admin"
	mustAdd(t, AddApplicationGrant, grant)
	return grant
}

// tokenClaimsFor runs the code flow for the user and returns the claims of the access token and the
// ID token and the userinfo of the application.
func tokenClaimsFor(t *testing.T, user *User, applicationName string, clientId string) (map[string]interface{}, map[string]interface{}, *Userinfo) {
	t.Helper()
	_, challenge := pkcePair()
	code := authorize(t, user, clientId, challenge)
	if code.Message != "" {
		t.Fatalf("%s denied for %s: %s", user.GetId(), applicationName, code.Message)
	}
	userinfo, err := GetUserInfo(grantTestUser(t, user.Owner, user.Name), "openid profile email", clientId, grantTestHost)
	if err != nil {
		t.Fatal(err)
	}
	return parseTestClaims(t, code.Token.AccessToken), parseTestClaims(t, code.Token.IdToken), userinfo
}

// claimRoleNames reads a roles claim, which holds names or, in the JWT-Custom format, role objects
func claimRoleNames(value interface{}) []string {
	res := []string{}
	items, _ := value.([]interface{})
	for _, item := range items {
		switch v := item.(type) {
		case string:
			res = append(res, v)
		case map[string]interface{}:
			if name, ok := v["name"].(string); ok {
				res = append(res, name)
			}
		}
	}
	return res
}

func sortedStrings(values []string) []string {
	res := append([]string{}, values...)
	sort.Strings(res)
	return res
}

// pveSyncGroups mirrors how Proxmox VE's OpenID realm uses a groups claim: names with characters
// outside its group name pattern are skipped, the realm is appended, and only groups that exist in
// Proxmox are kept (groups-autocreate off).
func pveSyncGroups(groupsClaim []string, realm string, pveGroups []string) []string {
	valid := regexp.MustCompile(`^[A-Za-z0-9.\-_]+$`)
	res := []string{}
	for _, group := range groupsClaim {
		if valid.MatchString(group) && util.InSlice(pveGroups, group+"-"+realm) {
			res = append(res, group+"-"+realm)
		}
	}
	return res
}

// grafanaStrictRole mirrors a strict Grafana generic OAuth setup: the user must be in one of the
// allowed groups, and the role expression must return a role, otherwise the sign-in fails.
func grafanaStrictRole(groupsClaim []string, allowedGroup string, roleForGroup string) (string, string) {
	if !util.InSlice(groupsClaim, allowedGroup) {
		return "", "user not a member of one of the required groups"
	}
	return roleForGroup, ""
}

func TestAppRoleLoginOnlyGrantAddsNothing(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "pve-like", &AppRole{Name: "pve-admin", Groups: []string{"resource-org/pve-admins"}, IsPrivileged: true})
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")

	access, id, userinfo := tokenClaimsFor(t, carol, "pve-like", "client-pve")
	for _, claims := range []map[string]interface{}{access, id} {
		if groups := claimStrings(claims["groups"]); !reflect.DeepEqual(groups, []string{"partner-org/pve-admins"}) {
			t.Fatalf("a sign-in only grant projected groups: %v", groups)
		}
	}
	if !reflect.DeepEqual(userinfo.Groups, []string{"partner-org/pve-admins"}) || len(userinfo.CustomClaims) != 0 {
		t.Fatalf("a sign-in only grant changed the userinfo: %+v", userinfo)
	}
	if got := pveSyncGroups(claimStrings(access["groups"]), "albweb-id", []string{"pve-admins-albweb-id"}); len(got) != 0 {
		t.Fatalf("Proxmox would sync a sign-in only user into %v", got)
	}
}

func TestAppRoleProxmoxStyleNameFormat(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "pve-like", &AppRole{Name: "pve-admin", DisplayName: "PVE administrators", Groups: []string{"resource-org/pve-admins"}, IsPrivileged: true})
	addTestGrantWithRoles(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol", "pve-admin")
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/erin")
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")
	bob := grantTestUser(t, grantTestResourceOrg, "internal-bob")

	internalAccess, _, _ := tokenClaimsFor(t, bob, "pve-like", "client-pve")
	access, _, userinfo := tokenClaimsFor(t, carol, "pve-like", "client-pve")

	if access["name"] != "carol@partner-org" || access["sub"] != carol.Id {
		t.Fatalf("identity claims changed by the role: %v %v", access["name"], access["sub"])
	}
	groups := claimStrings(access["groups"])
	if !util.InSlice(groups, "pve-admins") || !util.InSlice(groups, "partner-org/pve-admins") {
		t.Fatalf("expected the projected group and the user's own qualified group: %v", groups)
	}
	if !util.InSlice(userinfo.Groups, "pve-admins") {
		t.Fatalf("projected group missing from userinfo: %v", userinfo.Groups)
	}

	pveGroups := []string{"pve-admins-albweb-id"}
	internalSync := pveSyncGroups(claimStrings(internalAccess["groups"]), "albweb-id", pveGroups)
	externalSync := pveSyncGroups(groups, "albweb-id", pveGroups)
	if !reflect.DeepEqual(internalSync, externalSync) || len(externalSync) != 1 {
		t.Fatalf("the external user does not get what an internal member gets: internal %v, external %v", internalSync, externalSync)
	}

	// the user's actual memberships are untouched
	stored := grantTestUser(t, grantTestPartnerOrg, "carol")
	if !reflect.DeepEqual(stored.Groups, []string{"partner-org/pve-admins"}) {
		t.Fatalf("the role changed the user's real groups: %v", stored.Groups)
	}
	if isUser, _ := IsUserOfApplication(stored, grantTestApplication(t, "pve-like")); isUser {
		t.Fatal("the role made the user a user of the application's organization")
	}

	erin := grantTestUser(t, grantTestPartnerOrg, "erin")
	erinAccess, _, _ := tokenClaimsFor(t, erin, "pve-like", "client-pve")
	if got := pveSyncGroups(claimStrings(erinAccess["groups"]), "albweb-id", pveGroups); len(got) != 0 {
		t.Fatalf("a user granted without the role is synced into %v", got)
	}
}

func TestAppRoleGrafanaStylePathFormat(t *testing.T) {
	newAppRoleTestDb(t)
	adminPath := "resource-org/ops/grafana-admins"
	setAppRoles(t, "grafana-like", &AppRole{Name: "grafana-admin", Groups: []string{"resource-org/grafana-admins"}, IsPrivileged: true})
	addTestGrant(t, "grafana-like", ApplicationGrantPrincipalUser, "partner-org/dave")
	addTestGrantWithRoles(t, "grafana-like", ApplicationGrantPrincipalUser, "partner-org/carol", "grafana-admin")

	bob := grantTestUser(t, grantTestResourceOrg, "internal-bob")
	bob.Groups = []string{"resource-org/grafana-admins"}
	if _, err := UpdateUser(bob.GetId(), bob, []string{"groups"}, true); err != nil {
		t.Fatal(err)
	}
	internalAccess, _, _ := tokenClaimsFor(t, grantTestUser(t, grantTestResourceOrg, "internal-bob"), "grafana-like", "client-grafana")
	if !util.InSlice(claimStrings(internalAccess["groups"]), adminPath) {
		t.Fatalf("fixture: the internal member's path is %v", internalAccess["groups"])
	}

	_, _, daveUserinfo := tokenClaimsFor(t, grantTestUser(t, grantTestPartnerOrg, "dave"), "grafana-like", "client-grafana")
	if role, msg := grafanaStrictRole(daveUserinfo.Groups, adminPath, "Admin"); role != "" || msg == "" {
		t.Fatal("a sign-in only external user passed Grafana's strict check")
	}

	access, id, userinfo := tokenClaimsFor(t, grantTestUser(t, grantTestPartnerOrg, "carol"), "grafana-like", "client-grafana")
	for _, groups := range [][]string{claimStrings(access["groups"]), claimStrings(id["groups"]), userinfo.Groups} {
		if role, msg := grafanaStrictRole(groups, adminPath, "Admin"); role != "Admin" {
			t.Fatalf("the external user with the role did not get the Grafana role: %s %v", msg, groups)
		}
	}
	if access["owner"] != grantTestPartnerOrg || access["isAdmin"] != false {
		t.Fatalf("home organization or admin flag changed: %v %v", access["owner"], access["isAdmin"])
	}
}

func TestAppRoleIdFormatAndCasdoorRoles(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "claims-like", &AppRole{Name: "reader", Groups: []string{"resource-org/readers"}, Roles: []string{"resource-org/viewer"}})
	addTestGrantWithRoles(t, "claims-like", ApplicationGrantPrincipalUser, "partner-org/carol", "reader")

	access, _, userinfo := tokenClaimsFor(t, grantTestUser(t, grantTestPartnerOrg, "carol"), "claims-like", "client-claims")
	if groups := claimStrings(access["groups"]); !util.InSlice(groups, "resource-org/readers") {
		t.Fatalf("ID-format projection missing: %v", groups)
	}
	if roles := claimRoleNames(access["roles"]); !util.InSlice(roles, "viewer") {
		t.Fatalf("projected Casdoor role missing from the token: %v", roles)
	}
	if !util.InSlice(userinfo.Roles, "viewer") {
		t.Fatalf("projected Casdoor role missing from userinfo: %v", userinfo.Roles)
	}
	// a role of the same name in the partner organization is not what the role refers to
	if roles, _ := getRolesByUser("partner-org/carol"); len(roles) != 0 {
		t.Fatalf("the application role created a Casdoor role membership: %v", roles)
	}
}

func TestAppRoleCustomClaims(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "claims-like",
		&AppRole{Name: "editor", Claims: []*JwtItem{{Name: "app_role", Value: "Editor", Type: "String"}, {Name: "entitlements", Value: "write", Type: "Array"}}},
		&AppRole{Name: "auditor", Claims: []*JwtItem{{Name: "entitlements", Value: "audit", Type: "Array"}}},
		&AppRole{Name: "viewer", Claims: []*JwtItem{{Name: "app_role", Value: "Viewer", Type: "String"}}},
	)
	addTestGrantWithRoles(t, "claims-like", ApplicationGrantPrincipalUser, "partner-org/carol", "editor", "auditor")

	access, id, userinfo := tokenClaimsFor(t, grantTestUser(t, grantTestPartnerOrg, "carol"), "claims-like", "client-claims")
	for _, claims := range []map[string]interface{}{access, id, userinfo.CustomClaims} {
		if claims["app_role"] != "Editor" {
			t.Fatalf("String claim wrong: %v", claims["app_role"])
		}
		var entitlements []string
		switch v := claims["entitlements"].(type) {
		case []string:
			entitlements = v
		default:
			entitlements = claimStrings(v)
		}
		if !reflect.DeepEqual(entitlements, []string{"audit", "write"}) {
			t.Fatalf("Array claim not the sorted union: %v", claims["entitlements"])
		}
		if claims["tenant"] != "resource" {
			t.Fatalf("the application's own token attribute was lost: %v", claims["tenant"])
		}
	}

	// two roles giving a single-valued claim different values: the claim is dropped, not guessed
	application := grantTestApplication(t, "claims-like")
	entitlements, err := GetApplicationEntitlements(grantTestUser(t, grantTestPartnerOrg, "carol"), application, []string{"viewer"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := entitlements.Claims["app_role"]; ok || len(entitlements.Warnings) == 0 {
		t.Fatalf("conflicting String claim not dropped: %+v", entitlements)
	}
}

func TestAppRoleFullJwtFormatCarriesCustomClaims(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "grafana-like", &AppRole{Name: "editor", Claims: []*JwtItem{{Name: "grafana_role", Value: "Editor", Type: "String"}}})
	addTestGrantWithRoles(t, "grafana-like", ApplicationGrantPrincipalUser, "partner-org/carol", "editor")

	access, id, _ := tokenClaimsFor(t, grantTestUser(t, grantTestPartnerOrg, "carol"), "grafana-like", "client-grafana")
	if access["grafana_role"] != "Editor" || id["grafana_role"] != "Editor" {
		t.Fatalf("custom claim missing from the JWT format: %v %v", access["grafana_role"], id["grafana_role"])
	}
	if access["name"] != "carol@partner-org" || access["owner"] != grantTestPartnerOrg || access["tokenType"] != "access-token" || id["tokenType"] != "id-token" {
		t.Fatal("the custom claim changed other claims of the token")
	}
}

func TestAppRoleValidation(t *testing.T) {
	newAppRoleTestDb(t)
	application := grantTestApplication(t, "claims-like")
	invalid := map[string]*AppRole{
		"group of another organization": {Name: "x", Groups: []string{"partner-org/pve-admins"}},
		"missing group":                 {Name: "x", Groups: []string{"resource-org/nope"}},
		"role of another organization":  {Name: "x", Roles: []string{"partner-org/viewer"}},
		"missing role":                  {Name: "x", Roles: []string{"resource-org/nope"}},
		"subject claim":                 {Name: "x", Claims: []*JwtItem{{Name: "sub", Value: "admin", Type: "String"}}},
		"owner claim":                   {Name: "x", Claims: []*JwtItem{{Name: "owner", Value: "resource-org", Type: "String"}}},
		"name claim":                    {Name: "x", Claims: []*JwtItem{{Name: "preferred_username", Value: "admin", Type: "String"}}},
		"admin flag":                    {Name: "x", Claims: []*JwtItem{{Name: "isAdmin", Value: "true", Type: "String"}}},
		"groups claim as literal":       {Name: "x", Claims: []*JwtItem{{Name: "groups", Value: "pve-admins", Type: "Array"}}},
		"roles claim as literal":        {Name: "x", Claims: []*JwtItem{{Name: "Roles", Value: "admin", Type: "Array"}}},
		"application token attribute":   {Name: "x", Claims: []*JwtItem{{Name: "tenant", Value: "other", Type: "String"}}},
		"placeholder value":             {Name: "x", Claims: []*JwtItem{{Name: "who", Value: "$user.owner", Type: "String"}}},
		"empty value":                   {Name: "x", Claims: []*JwtItem{{Name: "who", Value: "", Type: "String"}}},
		"unknown type":                  {Name: "x", Claims: []*JwtItem{{Name: "who", Value: "a", Type: "Object"}}},
		"invalid role name":             {Name: "bad name/"},
	}
	for name, role := range invalid {
		if _, err := UpdateApplicationAppRoles(application, []*AppRole{role}, "en"); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := UpdateApplicationAppRoles(application, []*AppRole{{Name: "a"}, {Name: "a"}}, "en"); err == nil {
		t.Error("duplicate role names accepted")
	}
	shared := grantTestApplication(t, "shared-app")
	shared.AppRoles = []*AppRole{{Name: "a"}}
	if err := CheckApplicationAppRoles(shared, "en"); err == nil {
		t.Error("application roles on a shared application accepted")
	}
	// the full application update validates them too
	application.AppRoles = []*AppRole{{Name: "x", Groups: []string{"partner-org/pve-admins"}}}
	if _, err := UpdateApplication("admin/claims-like", application, false, "en", nil); err == nil {
		t.Error("update-application accepted a role with a foreign group")
	}

	setAppRoles(t, "claims-like", &AppRole{Name: "reader", Groups: []string{"resource-org/readers"}})
	grant := ApplicationGrant{Owner: grantTestResourceOrg, Application: "claims-like", PrincipalType: "User", Principal: "partner-org/carol", Roles: []string{"admin"}}
	if err := CheckApplicationGrant(&grant, grantTestApplication(t, "claims-like"), "en"); err == nil {
		t.Error("grant with an undefined application role accepted")
	}
	addTestGrantWithRoles(t, "claims-like", ApplicationGrantPrincipalUser, "partner-org/carol", "reader")
	if _, err := UpdateApplicationAppRoles(grantTestApplication(t, "claims-like"), nil, "en"); err == nil {
		t.Error("removed an application role still assigned by a grant")
	}
}

func TestAppRoleIsolationBetweenApplications(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "pve-like", &AppRole{Name: "pve-admin", Groups: []string{"resource-org/pve-admins"}})
	setAppRoles(t, "claims-like", &AppRole{Name: "reader", Claims: []*JwtItem{{Name: "app_role", Value: "Reader", Type: "String"}}})
	addTestGrantWithRoles(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol", "pve-admin")
	addTestGrant(t, "grafana-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	addTestGrant(t, "claims-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")

	grafana, _, _ := tokenClaimsFor(t, carol, "grafana-like", "client-grafana")
	claims, _, _ := tokenClaimsFor(t, carol, "claims-like", "client-claims")
	for _, groups := range [][]string{claimStrings(grafana["groups"]), claimStrings(claims["groups"])} {
		for _, group := range groups {
			if strings.Contains(group, "pve-admins") && group != "partner-org/pve-admins" {
				t.Fatalf("the pve-like role leaked into another application: %v", groups)
			}
		}
	}
	if _, ok := claims["app_role"]; ok {
		t.Fatal("a role defined but not granted was applied")
	}

	// a grant of one application can't name the roles of another
	grant := ApplicationGrant{Owner: grantTestResourceOrg, Application: "grafana-like", PrincipalType: "User", Principal: "partner-org/dave", Roles: []string{"pve-admin"}}
	if err := CheckApplicationGrant(&grant, grantTestApplication(t, "grafana-like"), "en"); err == nil {
		t.Fatal("grant assigned a role of another application")
	}

	// tenant isolation: a group grant of one organization does not reach a same-named group of another
	addTestGrantWithRoles(t, "pve-like", ApplicationGrantPrincipalGroup, "partner-org/team", "pve-admin")
	if entitlements, _ := GetApplicationEntitlements(grantTestUser(t, grantTestOtherOrg, "frank"), grantTestApplication(t, "pve-like"), nil); len(entitlements.Groups) != 0 {
		t.Fatalf("other-org/team received partner-org/team's role: %v", entitlements.Groups)
	}
}

func TestAppRoleOverlappingGrantsAndRevocation(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "grafana-like",
		&AppRole{Name: "a", Claims: []*JwtItem{{Name: "perm", Value: "a", Type: "Array"}}},
		&AppRole{Name: "b", Claims: []*JwtItem{{Name: "perm", Value: "b", Type: "Array"}}},
	)
	direct := addTestGrantWithRoles(t, "grafana-like", ApplicationGrantPrincipalUser, "partner-org/erin", "a")
	viaGroup := addTestGrantWithRoles(t, "grafana-like", ApplicationGrantPrincipalGroup, "partner-org/team", "b")
	application := grantTestApplication(t, "grafana-like")

	perms := func() []string {
		t.Helper()
		entitlements, err := GetApplicationEntitlements(grantTestUser(t, grantTestPartnerOrg, "erin"), application, nil)
		if err != nil {
			t.Fatal(err)
		}
		values, _ := entitlements.Claims["perm"].([]string)
		return values
	}
	if got := perms(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("overlapping grants are not the union: %v", got)
	}

	// tokens issued with both roles; the refresh token is used after the changes below
	erin := grantTestUser(t, grantTestPartnerOrg, "erin")
	verifier, challenge := pkcePair()
	code := authorize(t, erin, "client-grafana", challenge)
	token, tokenError, err := GetAuthorizationCodeToken(application, "secret-grafana", code.Code, verifier, "", "en")
	if err != nil || tokenError != nil {
		t.Fatalf("code exchange: %v %+v", err, tokenError)
	}

	expired := time.Now().Add(-time.Minute).Format(time.RFC3339)
	if _, err := ormer.Engine.ID([]interface{}{direct.Owner, direct.Name}).Cols("expire_time").Update(&ApplicationGrant{ExpireTime: expired}); err != nil {
		t.Fatal(err)
	}
	if got := perms(); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("expired grant still contributes: %v", got)
	}
	if granted, _ := HasApplicationGrant(grantTestUser(t, grantTestPartnerOrg, "erin"), application); !granted {
		t.Fatal("the remaining group grant no longer allows the sign-in")
	}

	refreshed, err := RefreshToken(application, "refresh_token", token.RefreshToken, "", "client-grafana", "secret-grafana", "", grantTestHost, "", "")
	if err != nil {
		t.Fatal(err)
	}
	wrapper, ok := refreshed.(*TokenWrapper)
	if !ok {
		t.Fatalf("refresh failed: %+v", refreshed)
	}
	if got := claimStrings(parseTestClaims(t, wrapper.AccessToken)["perm"]); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("refresh did not recompute the roles: %v", got)
	}

	// group membership removed: the group grant's role and the sign-in end
	erin = grantTestUser(t, grantTestPartnerOrg, "erin")
	erin.Groups = []string{}
	if _, err := UpdateUser(erin.GetId(), erin, []string{"groups"}, true); err != nil {
		t.Fatal(err)
	}
	if got := perms(); len(got) != 0 {
		t.Fatalf("roles kept after leaving the group: %v", got)
	}
	_ = viaGroup
}

func TestAppRoleGrantUpdate(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "pve-like", &AppRole{Name: "pve-admin", Groups: []string{"resource-org/pve-admins"}, IsPrivileged: true}, &AppRole{Name: "reader", Claims: []*JwtItem{{Name: "app_role", Value: "Reader", Type: "String"}}})
	grant := addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	application := grantTestApplication(t, "pve-like")
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")

	update := ApplicationGrant{Roles: []string{"pve-admin"}, Principal: "partner-org/dave", Description: "ticket 1"}
	stored, _ := GetApplicationGrant(grant.GetId())
	if err := CheckApplicationGrantUpdate(stored, &update, application, "en"); err != nil {
		t.Fatal(err)
	}
	stored.UpdatedBy = "resource-org/internal-admin"
	if _, err := UpdateApplicationGrant(stored); err != nil {
		t.Fatal(err)
	}
	stored, _ = GetApplicationGrant(grant.GetId())
	if stored.Principal != "partner-org/carol" || !reflect.DeepEqual(stored.Roles, []string{"pve-admin"}) || stored.UpdatedBy == "" || stored.UpdatedTime == "" {
		t.Fatalf("update not applied as expected: %+v", stored)
	}
	userinfo, _ := GetUserInfo(carol, "openid profile", "client-pve", grantTestHost)
	if !util.InSlice(userinfo.Groups, "pve-admins") {
		t.Fatal("userinfo does not reflect the added role at once")
	}

	update = ApplicationGrant{Roles: []string{}}
	if err := CheckApplicationGrantUpdate(stored, &update, application, "en"); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateApplicationGrant(stored); err != nil {
		t.Fatal(err)
	}
	userinfo, _ = GetUserInfo(carol, "openid profile", "client-pve", grantTestHost)
	if util.InSlice(userinfo.Groups, "pve-admins") {
		t.Fatal("userinfo still shows a removed role")
	}
	if err := CheckApplicationGrantUpdate(stored, &ApplicationGrant{Roles: []string{"nope"}}, application, "en"); err == nil {
		t.Fatal("update with an undefined role accepted")
	}
}

func TestAppRoleInternalPrincipal(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "claims-like", &AppRole{Name: "reader", Claims: []*JwtItem{{Name: "app_role", Value: "Reader", Type: "String"}}})
	application := grantTestApplication(t, "claims-like")

	signInOnly := ApplicationGrant{Owner: grantTestResourceOrg, Application: "claims-like", PrincipalType: "User", Principal: "resource-org/internal-bob"}
	if err := CheckApplicationGrant(&signInOnly, application, "en"); err == nil {
		t.Fatal("sign-in only grant for a user of the application's organization accepted")
	}
	addTestGrantWithRoles(t, "claims-like", ApplicationGrantPrincipalUser, "resource-org/internal-bob", "reader")

	bob := grantTestUser(t, grantTestResourceOrg, "internal-bob")
	_, challenge := pkcePair()
	code := authorize(t, bob, "client-claims", challenge)
	if code.Message != "" || code.Token.IsGrantAccess {
		t.Fatalf("internal sign-in changed: %q", code.Message)
	}
	claims := parseTestClaims(t, code.Token.AccessToken)
	if claims["name"] != "internal-bob" || claims["app_role"] != "Reader" {
		t.Fatalf("internal user with an application role: %v %v", claims["name"], claims["app_role"])
	}
	// an internal role assignment gives no access to users of other organizations
	if msg, _ := CheckApplicationAccess(grantTestUser(t, grantTestPartnerOrg, "internal-bob"), application, "en"); msg == "" {
		t.Fatal("same-named user of another organization admitted")
	}
}

func TestAppRoleDanglingAndTamperedReferences(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "pve-like", &AppRole{Name: "pve-admin", Groups: []string{"resource-org/pve-admins", "resource-org/readers"}})
	addTestGrantWithRoles(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol", "pve-admin")
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")

	setGroupEnabled(t, grantTestResourceOrg, "readers", false)
	// written into the database directly, bypassing validation
	application := grantTestApplication(t, "pve-like")
	application.AppRoles[0].Groups = append(application.AppRoles[0].Groups, "partner-org/team", "resource-org/deleted")
	if _, err := ormer.Engine.ID([]interface{}{"admin", "pve-like"}).Cols("app_roles").Update(application); err != nil {
		t.Fatal(err)
	}

	entitlements, err := GetApplicationEntitlements(carol, grantTestApplication(t, "pve-like"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(entitlements.Groups, []string{"pve-admins"}) {
		t.Fatalf("disabled, foreign or missing groups projected: %v", entitlements.Groups)
	}
	if len(entitlements.Warnings) != 3 {
		t.Fatalf("expected a warning per ignored reference: %v", entitlements.Warnings)
	}
}

func TestAppRoleReferencesFollowRenames(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "claims-like", &AppRole{Name: "reader", Groups: []string{"resource-org/readers"}, Roles: []string{"resource-org/viewer"}})

	group, _ := getGroup(grantTestResourceOrg, "readers")
	group.Name = "readers-2"
	if _, err := UpdateGroup("resource-org/readers", group, true, "en"); err != nil {
		t.Fatal(err)
	}
	role, _ := getRole(grantTestResourceOrg, "viewer")
	role.Name = "viewer-2"
	if _, err := UpdateRole("resource-org/viewer", role, true, "en"); err != nil {
		t.Fatal(err)
	}
	appRole := grantTestApplication(t, "claims-like").getAppRole("reader")
	if !reflect.DeepEqual(appRole.Groups, []string{"resource-org/readers-2"}) || !reflect.DeepEqual(appRole.Roles, []string{"resource-org/viewer-2"}) {
		t.Fatalf("references did not follow the renames: %+v", appRole)
	}
}

func TestAppRolePreviewAndMasking(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "pve-like", &AppRole{Name: "pve-admin", Groups: []string{"resource-org/pve-admins"}, IsPrivileged: true})
	application := grantTestApplication(t, "pve-like")

	// before any grant: what the role would add, without the user's own memberships
	preview, err := GetApplicationAccessPreview(grantTestUser(t, grantTestPartnerOrg, "carol"), application, []string{"pve-admin"}, true, "en")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Access != "Pending" || len(preview.OwnGroups) != 0 || !preview.Privileged || !reflect.DeepEqual(preview.Groups, []string{"pve-admins"}) || preview.PreferredName != "carol@partner-org" {
		t.Fatalf("draft preview: %+v", preview)
	}

	addTestGrantWithRoles(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol", "pve-admin")
	preview, err = GetApplicationAccessPreview(grantTestUser(t, grantTestPartnerOrg, "carol"), application, nil, false, "en")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Access != "Grant" || len(preview.Entitlements.Grants) != 1 || !reflect.DeepEqual(sortedStrings(preview.Groups), []string{"partner-org/pve-admins", "pve-admins"}) || !reflect.DeepEqual(preview.OwnGroups, []string{"partner-org/pve-admins"}) {
		t.Fatalf("granted preview: %+v", preview)
	}

	claims := grantTestApplication(t, "claims-like")
	claims.TokenFields = []string{"Name"}
	if p, _ := GetApplicationAccessPreview(grantTestUser(t, grantTestResourceOrg, "internal-bob"), claims, nil, false, "en"); p.EmitsGroups || p.Access != "Native" {
		t.Fatalf("preview should tell that the application does not emit groups: %+v", p)
	}

	masked := GetMaskedApplication(grantTestApplication(t, "pve-like"), "")
	if len(masked.AppRoles) != 0 {
		t.Fatal("application roles exposed to anonymous users")
	}
}

func TestAppRoleDraftPreviewDoesNotProbe(t *testing.T) {
	newAppRoleTestDb(t)
	organization, _ := getOrganization("admin", grantTestResourceOrg)
	organization.MfaItems = []*MfaItem{{Name: TotpType, Rule: "Required"}}
	if _, err := ormer.Engine.ID([]interface{}{"admin", grantTestResourceOrg}).Cols("mfa_items").Update(organization); err != nil {
		t.Fatal(err)
	}
	// before the grant, the MFA set-up of a user of another organization is not revealed
	preview, err := GetApplicationAccessPreview(grantTestUser(t, grantTestPartnerOrg, "carol"), grantTestApplication(t, "pve-like"), nil, true, "en")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Access != "Pending" || preview.Reason != "" {
		t.Fatalf("a draft preview revealed a precondition of a user of another organization: %+v", preview)
	}
	// once granted, explaining the access shows why the user can't sign in
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	preview, err = GetApplicationAccessPreview(grantTestUser(t, grantTestPartnerOrg, "carol"), grantTestApplication(t, "pve-like"), nil, false, "en")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Access != "Denied" || !strings.Contains(preview.Reason, "multi-factor") {
		t.Fatalf("the MFA requirement is not explained for a granted user: %+v", preview)
	}
	// a global admin signs in natively, but their memberships are not the application's business
	admin := grantTestUser(t, "built-in", "admin")
	admin.Groups = []string{"built-in/secret-team"}
	preview, err = GetApplicationAccessPreview(admin, grantTestApplication(t, "pve-like"), nil, false, "en")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Access != "Native" || len(preview.OwnGroups) != 0 {
		t.Fatalf("the memberships of a global admin were shown: %+v", preview)
	}
}

func TestAppRolePrivilegedRolesStayOffForeignGroups(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "pve-like", &AppRole{Name: "pve-admin", Groups: []string{"resource-org/pve-admins"}, IsPrivileged: true}, &AppRole{Name: "reader", Groups: []string{"resource-org/readers"}})
	application := grantTestApplication(t, "pve-like")

	grant := ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "Group", Principal: "partner-org/team", Roles: []string{"pve-admin"}}
	if err := CheckApplicationGrant(&grant, application, "en"); err == nil {
		t.Fatal("a privileged role was granted to a group another organization controls")
	}
	// a user of the other organization, or a group of the application's own, may get it
	addTestGrantWithRoles(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol", "pve-admin")
	addTestGrantWithRoles(t, "pve-like", ApplicationGrantPrincipalGroup, "resource-org/readers", "pve-admin")

	groupGrant := addTestGrantWithRoles(t, "pve-like", ApplicationGrantPrincipalGroup, "partner-org/team", "reader")
	stored, _ := GetApplicationGrant(groupGrant.GetId())
	if err := CheckApplicationGrantUpdate(stored, &ApplicationGrant{Roles: []string{"reader", "pve-admin"}}, application, "en"); err == nil {
		t.Fatal("an edit added a privileged role to a foreign group grant")
	}
	if _, err := UpdateApplicationAppRoles(application, []*AppRole{{Name: "pve-admin", Groups: []string{"resource-org/pve-admins"}, IsPrivileged: true}, {Name: "reader", Groups: []string{"resource-org/readers"}, IsPrivileged: true}}, "en"); err == nil {
		t.Fatal("a role held by a foreign group grant was made privileged")
	}
}

func TestAppRoleReferencesRemovedWithTheirGroupOrRole(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "claims-like", &AppRole{Name: "reader", Groups: []string{"resource-org/readers"}, Roles: []string{"resource-org/viewer"}})

	group, _ := getGroup(grantTestResourceOrg, "readers")
	if _, err := DeleteGroup(group); err != nil {
		t.Fatal(err)
	}
	role, _ := getRole(grantTestResourceOrg, "viewer")
	if _, err := DeleteRole(role); err != nil {
		t.Fatal(err)
	}
	appRole := grantTestApplication(t, "claims-like").getAppRole("reader")
	if len(appRole.Groups) != 0 || len(appRole.Roles) != 0 {
		t.Fatalf("references to deleted objects kept, a recreated one would be projected: %+v", appRole)
	}
}

func TestAppRolePartialUpdateValidatesAgainstStoredApplication(t *testing.T) {
	newAppRoleTestDb(t)
	setAppRoles(t, "claims-like", &AppRole{Name: "reader", Claims: []*JwtItem{{Name: "level", Value: "1", Type: "String"}}})
	stored := grantTestApplication(t, "claims-like")

	// a request that only changes the token attributes, colliding with a stored role claim
	request := *stored
	request.TokenAttributes = append(request.TokenAttributes, &JwtItem{Name: "level", Category: "Static Value", Value: "9", Type: "String"})
	if _, err := UpdateApplication("admin/claims-like", &request, true, "en", []string{"token_attributes"}); err == nil {
		t.Fatal("a token attribute colliding with a role claim was saved")
	}
	// and a role claim colliding with a Properties.* token field
	request = *stored
	request.TokenFields = append(request.TokenFields, "Properties.level")
	if _, err := UpdateApplication("admin/claims-like", &request, true, "en", nil); err == nil {
		t.Fatal("a Properties token field colliding with a role claim was saved")
	}
}
