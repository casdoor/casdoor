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
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
)

const (
	grantTestResourceOrg = "resource-org"
	grantTestPartnerOrg  = "partner-org"
	grantTestOtherOrg    = "other-org"
	grantTestRedirectUri = "https://rp.example.test/callback"
	grantTestHost        = "id.example.test"
)

// newApplicationGrantTestDb gives the test its own SQLite database with synthetic organizations,
// users, groups and applications. The resource organization owns the applications; the users of
// the partner and other organizations have no access to them unless a grant says so.
func newApplicationGrantTestDb(t *testing.T) {
	t.Helper()

	dataSourceName := "file:" + filepath.Join(t.TempDir(), "casdoor.db") + "?_pragma=busy_timeout(5000)"
	// the casbin adapters of the user enforcer read the database from the configuration
	t.Setenv("driverName", "sqlite")
	t.Setenv("dataSourceName", dataSourceName)
	t.Setenv("dbName", "")

	previousOrmer, previousUserEnforcer := ormer, userEnforcer
	t.Cleanup(func() {
		ormer, userEnforcer = previousOrmer, previousUserEnforcer
	})

	var err error
	ormer, err = NewAdapter("sqlite", dataSourceName, "")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	ormer.createTable()
	InitDb()
	InitUserManager()

	for _, name := range []string{grantTestResourceOrg, grantTestPartnerOrg, grantTestOtherOrg} {
		mustAdd(t, AddOrganization, &Organization{Owner: "admin", Name: name, DisplayName: name, PasswordType: "plain", CreatedTime: util.GetCurrentTime()})
	}

	groups := []*Group{
		{Owner: grantTestResourceOrg, Name: "pve-admins", ParentId: grantTestResourceOrg, IsTopGroup: true, IsEnabled: true},
		// same name as the resource organization's admin group, in the partner organization
		{Owner: grantTestPartnerOrg, Name: "pve-admins", ParentId: grantTestPartnerOrg, IsTopGroup: true, IsEnabled: true},
		{Owner: grantTestPartnerOrg, Name: "team", ParentId: grantTestPartnerOrg, IsTopGroup: true, IsEnabled: true},
		{Owner: grantTestPartnerOrg, Name: "contractors", ParentId: "team", IsTopGroup: false, IsEnabled: true},
		{Owner: grantTestOtherOrg, Name: "team", ParentId: grantTestOtherOrg, IsTopGroup: true, IsEnabled: true},
	}
	for _, group := range groups {
		group.CreatedTime = util.GetCurrentTime()
		mustAdd(t, AddGroup, group)
	}

	applications := []*Application{
		// a relying party that keys accounts on the name and reads group names, like Proxmox VE
		{
			Owner: "admin", Name: "pve-like", Organization: grantTestResourceOrg, ClientId: "client-pve", ClientSecret: "secret-pve",
			RedirectUris: []string{grantTestRedirectUri}, GrantTypes: []string{"authorization_code"},
			TokenFormat: "JWT-Custom", TokenFields: []string{"Name", "Groups"}, TokenGroupFormat: "Name", ExpireInHours: 1,
		},
		// a relying party that keys accounts on "sub" and reads group paths, like Grafana
		{
			Owner: "admin", Name: "grafana-like", Organization: grantTestResourceOrg, ClientId: "client-grafana", ClientSecret: "secret-grafana",
			RedirectUris: []string{grantTestRedirectUri}, GrantTypes: []string{"authorization_code", "refresh_token"},
			TokenFormat: "JWT", TokenFields: []string{"Name", "DisplayName", "Email", "Groups"}, TokenGroupFormat: "Path",
			ExpireInHours: 1, RefreshExpireInHours: 1, DisableRefreshRotation: true,
		},
		{Owner: "admin", Name: "partner-portal", Organization: grantTestPartnerOrg, ClientId: "client-partner", ClientSecret: "secret-partner", RedirectUris: []string{grantTestRedirectUri}, GrantTypes: []string{"authorization_code"}, ExpireInHours: 1},
		{Owner: "admin", Name: "other-portal", Organization: grantTestOtherOrg, ClientId: "client-other", ClientSecret: "secret-other", RedirectUris: []string{grantTestRedirectUri}, GrantTypes: []string{"authorization_code"}, ExpireInHours: 1},
		{
			Owner: "admin", Name: "shared-app", Organization: "built-in", IsShared: true, ClientId: "client-shared", ClientSecret: "secret-shared",
			RedirectUris: []string{grantTestRedirectUri}, GrantTypes: []string{"authorization_code"},
			TokenFormat: "JWT", TokenGroupFormat: "Name", ExpireInHours: 1,
		},
	}
	for _, application := range applications {
		application.CreatedTime = util.GetCurrentTime()
		application.TokenSigningMethod = "RS256"
		mustAdd(t, func(a *Application) (bool, error) { return AddApplication(a, "en") }, application)
	}

	users := []*User{
		{Owner: grantTestResourceOrg, Name: "internal-admin", IsAdmin: true, Groups: []string{"resource-org/pve-admins"}},
		{Owner: grantTestResourceOrg, Name: "internal-bob", Groups: []string{"resource-org/pve-admins"}},
		{Owner: grantTestPartnerOrg, Name: "carol", Email: "carol@partner.example.test", Groups: []string{"partner-org/pve-admins"}},
		{Owner: grantTestPartnerOrg, Name: "dave", Groups: []string{"partner-org/pve-admins"}},
		{Owner: grantTestPartnerOrg, Name: "erin", Groups: []string{"partner-org/contractors"}},
		{Owner: grantTestPartnerOrg, Name: "partner-admin", IsAdmin: true},
		// the same name as a user of the resource organization
		{Owner: grantTestPartnerOrg, Name: "internal-bob"},
		{Owner: grantTestOtherOrg, Name: "frank", Groups: []string{"other-org/team"}},
	}
	for _, user := range users {
		user.Id = util.GenerateId()
		user.CreatedTime = util.GetCurrentTime()
		user.DisplayName = user.Name
		user.Password = "synthetic-password"
		user.PasswordType = "plain"
		mustAdd(t, func(u *User) (bool, error) { return AddUser(u, "en") }, user)
	}
}

func mustAdd[T any](t *testing.T, add func(T) (bool, error), item T) {
	t.Helper()
	affected, err := add(item)
	if err != nil || !affected {
		t.Fatalf("add %T: affected=%v err=%v", item, affected, err)
	}
}

func grantTestUser(t *testing.T, owner string, name string) *User {
	t.Helper()
	user, err := getUser(owner, name)
	if err != nil || user == nil {
		t.Fatalf("get user %s/%s: %v", owner, name, err)
	}
	return user
}

func grantTestApplication(t *testing.T, name string) *Application {
	t.Helper()
	application, err := GetApplication(util.GetId("admin", name))
	if err != nil || application == nil {
		t.Fatalf("get application %s: %v", name, err)
	}
	return application
}

func addTestGrant(t *testing.T, applicationName string, principalType string, principal string) *ApplicationGrant {
	t.Helper()
	application := grantTestApplication(t, applicationName)
	grant := &ApplicationGrant{Owner: application.Organization, Application: application.Name, PrincipalType: principalType, Principal: principal}
	if err := CheckApplicationGrant(grant, application, "en"); err != nil {
		t.Fatalf("check grant: %v", err)
	}
	grant.CreatedBy = "resource-org/internal-admin"
	mustAdd(t, AddApplicationGrant, grant)
	return grant
}

func pkcePair() (string, string) {
	verifier := util.GenerateId() + util.GenerateId()
	sum := sha256.Sum256([]byte(verifier))
	return verifier, base64.RawURLEncoding.EncodeToString(sum[:])
}

// authorize runs the authorization request of the code flow with PKCE, state and nonce
func authorize(t *testing.T, user *User, clientId string, challenge string) *Code {
	t.Helper()
	code, err := GetOAuthCode(user.GetId(), clientId, "", "Password", "code", grantTestRedirectUri, "openid profile email", "state-1", "nonce-1", challenge, "", "session-1", time.Now().Unix(), grantTestHost, "en")
	if err != nil {
		t.Fatalf("GetOAuthCode: %v", err)
	}
	return code
}

func parseTestClaims(t *testing.T, token string) jwt.MapClaims {
	t.Helper()
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	return claims
}

func claimStrings(value interface{}) []string {
	res := []string{}
	items, _ := value.([]interface{})
	for _, item := range items {
		res = append(res, item.(string))
	}
	return res
}

func TestApplicationGrantInternalUserUnchanged(t *testing.T) {
	newApplicationGrantTestDb(t)
	bob := grantTestUser(t, grantTestResourceOrg, "internal-bob")
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")

	verifier, challenge := pkcePair()
	code := authorize(t, bob, "client-pve", challenge)
	if code.Message != "" || code.Token == nil {
		t.Fatalf("internal user denied: %q", code.Message)
	}
	if code.Token.IsGrantAccess {
		t.Fatal("token of an internal user is marked as grant access")
	}

	claims := parseTestClaims(t, code.Token.AccessToken)
	if claims["name"] != "internal-bob" || claims["sub"] != bob.Id {
		t.Fatalf("internal claims changed: name=%v sub=%v", claims["name"], claims["sub"])
	}
	if groups := claimStrings(claims["groups"]); len(groups) != 1 || groups[0] != "pve-admins" {
		t.Fatalf("internal Name-format groups changed: %v", groups)
	}

	token, tokenError, err := GetAuthorizationCodeToken(grantTestApplication(t, "pve-like"), "secret-pve", code.Code, verifier, "", "en")
	if err != nil || tokenError != nil || token == nil {
		t.Fatalf("code exchange: %v %+v", err, tokenError)
	}

	userinfo, err := GetUserInfo(bob, "openid profile", "client-pve", grantTestHost)
	if err != nil || userinfo.Name != "internal-bob" || len(userinfo.Groups) != 1 || userinfo.Groups[0] != "pve-admins" {
		t.Fatalf("internal userinfo changed: %+v %v", userinfo, err)
	}
}

func TestApplicationGrantDeniedByDefault(t *testing.T) {
	newApplicationGrantTestDb(t)
	application := grantTestApplication(t, "pve-like")
	for _, name := range []string{"dave", "partner-admin", "internal-bob"} {
		user := grantTestUser(t, grantTestPartnerOrg, name)
		msg, err := CheckApplicationAccess(user, application, "en")
		if err != nil || msg == "" {
			t.Fatalf("%s allowed without a grant: %q %v", name, msg, err)
		}
		_, challenge := pkcePair()
		if code := authorize(t, user, "client-pve", challenge); code.Message == "" || code.Code != "" {
			t.Fatalf("%s got an authorization code without a grant", name)
		}
	}
}

func TestApplicationGrantUserGrantFullFlow(t *testing.T) {
	newApplicationGrantTestDb(t)
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	application := grantTestApplication(t, "pve-like")

	verifier, challenge := pkcePair()
	code := authorize(t, carol, "client-pve", challenge)
	if code.Message != "" || code.Token == nil {
		t.Fatalf("granted user denied: %q", code.Message)
	}
	if !code.Token.IsGrantAccess || code.Token.Organization != grantTestPartnerOrg {
		t.Fatalf("token not bound to the user's own organization: %+v", code.Token)
	}

	// PKCE is still enforced
	if _, tokenError, _ := GetAuthorizationCodeToken(application, "secret-pve", code.Code, "wrong-verifier", "", "en"); tokenError == nil {
		t.Fatal("wrong PKCE verifier accepted")
	}
	token, tokenError, err := GetAuthorizationCodeToken(application, "secret-pve", code.Code, verifier, "", "en")
	if err != nil || tokenError != nil {
		t.Fatalf("code exchange: %v %+v", err, tokenError)
	}
	// the code can't be replayed
	if _, tokenError, _ := GetAuthorizationCodeToken(application, "secret-pve", code.Code, verifier, "", "en"); tokenError == nil {
		t.Fatal("authorization code replayed")
	}

	for _, raw := range []string{token.AccessToken, token.IdToken} {
		claims := parseTestClaims(t, raw)
		if claims["name"] != "carol@partner-org" {
			t.Fatalf("name not qualified: %v", claims["name"])
		}
		if claims["sub"] != carol.Id {
			t.Fatalf("sub is not the user's immutable ID: %v", claims["sub"])
		}
		if aud := claimStrings(claims["aud"]); len(aud) != 1 || aud[0] != "client-pve" {
			t.Fatalf("audience is not the unchanged client: %v", aud)
		}
		groups := claimStrings(claims["groups"])
		if len(groups) != 1 || groups[0] != "partner-org/pve-admins" {
			t.Fatalf("foreign group not qualified, it would collide with the resource group: %v", groups)
		}
	}
	idClaims := parseTestClaims(t, token.IdToken)
	if idClaims["nonce"] != "nonce-1" {
		t.Fatalf("nonce lost: %v", idClaims["nonce"])
	}

	userinfo, err := GetUserInfo(carol, "openid profile", "client-pve", grantTestHost)
	if err != nil {
		t.Fatal(err)
	}
	if userinfo.Name != "carol@partner-org" || userinfo.Sub != carol.Id || len(userinfo.Groups) != 1 || userinfo.Groups[0] != "partner-org/pve-admins" {
		t.Fatalf("userinfo not qualified: %+v", userinfo)
	}
	if active, err := token.IsUserActive(); err != nil || !active {
		t.Fatalf("granted token inactive: %v", err)
	}
}

func TestApplicationGrantFullJwtKeepsHomeOrganization(t *testing.T) {
	newApplicationGrantTestDb(t)
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")
	addTestGrant(t, "grafana-like", ApplicationGrantPrincipalUser, "partner-org/carol")

	_, challenge := pkcePair()
	code := authorize(t, carol, "client-grafana", challenge)
	if code.Message != "" {
		t.Fatal(code.Message)
	}
	claims := parseTestClaims(t, code.Token.AccessToken)
	if claims["owner"] != grantTestPartnerOrg {
		t.Fatalf("home organization not preserved: %v", claims["owner"])
	}
	if groups := claimStrings(claims["groups"]); len(groups) != 1 || groups[0] != "partner-org/pve-admins" {
		t.Fatalf("Path-format groups wrong: %v", groups)
	}
}

func TestApplicationGrantNestedGroupGrant(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalGroup, "partner-org/team")
	application := grantTestApplication(t, "pve-like")
	erin := grantTestUser(t, grantTestPartnerOrg, "erin")

	assertAccess := func(want bool, step string) {
		t.Helper()
		user := grantTestUser(t, erin.Owner, erin.Name)
		msg, err := CheckApplicationAccess(user, application, "en")
		if err != nil {
			t.Fatal(err)
		}
		if (msg == "") != want {
			t.Fatalf("%s: access=%v, want %v (%q)", step, msg == "", want, msg)
		}
	}

	assertAccess(true, "member of a child group of the granted group")
	if granted, _ := HasApplicationGrant(grantTestUser(t, grantTestPartnerOrg, "dave"), application); granted {
		t.Fatal("user outside the granted group tree is granted")
	}
	if granted, _ := HasApplicationGrant(grantTestUser(t, grantTestOtherOrg, "frank"), application); granted {
		t.Fatal("member of a same-named group of another organization is granted")
	}

	setGroupEnabled(t, grantTestPartnerOrg, "contractors", false)
	assertAccess(false, "child group disabled")
	setGroupEnabled(t, grantTestPartnerOrg, "contractors", true)
	setGroupEnabled(t, grantTestPartnerOrg, "team", false)
	assertAccess(false, "granted group disabled")
	setGroupEnabled(t, grantTestPartnerOrg, "team", true)
	assertAccess(true, "groups enabled again")

	erin.Groups = []string{}
	if _, err := UpdateUser(erin.GetId(), erin, []string{"groups"}, true); err != nil {
		t.Fatal(err)
	}
	assertAccess(false, "removed from the group")
}

func setGroupEnabled(t *testing.T, owner string, name string, enabled bool) {
	t.Helper()
	_, err := ormer.Engine.Where("owner = ? and name = ?", owner, name).Cols("is_enabled").Update(&Group{IsEnabled: enabled})
	if err != nil {
		t.Fatal(err)
	}
}

func TestApplicationGrantRevocation(t *testing.T) {
	newApplicationGrantTestDb(t)
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")
	grant := addTestGrant(t, "grafana-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	application := grantTestApplication(t, "grafana-like")

	verifier, challenge := pkcePair()
	code := authorize(t, carol, "client-grafana", challenge)
	token, tokenError, err := GetAuthorizationCodeToken(application, "secret-grafana", code.Code, verifier, "", "en")
	if err != nil || tokenError != nil {
		t.Fatalf("code exchange: %v %+v", err, tokenError)
	}
	refreshed, err := RefreshToken(application, "refresh_token", token.RefreshToken, "", "client-grafana", "secret-grafana", "", grantTestHost, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := refreshed.(*TokenWrapper); !ok {
		t.Fatalf("refresh with a valid grant failed: %+v", refreshed)
	}

	// a code issued before the revocation can't be redeemed after it
	verifier2, challenge2 := pkcePair()
	pendingCode := authorize(t, carol, "client-grafana", challenge2)

	if affected, err := DeleteApplicationGrant(grant); err != nil || !affected {
		t.Fatalf("revoke: %v", err)
	}

	if active, err := token.IsUserActive(); err != nil || active {
		t.Fatalf("token still active after revocation: %v", err)
	}
	refreshed, err = RefreshToken(application, "refresh_token", token.RefreshToken, "", "client-grafana", "secret-grafana", "", grantTestHost, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if tokenError, ok := refreshed.(*TokenError); !ok || tokenError.Error != InvalidGrant {
		t.Fatalf("refresh after revocation not refused: %+v", refreshed)
	}
	if _, tokenError, _ := GetAuthorizationCodeToken(application, "secret-grafana", pendingCode.Code, verifier2, "", "en"); tokenError == nil {
		t.Fatal("code issued before the revocation was redeemed")
	}
	_, challenge3 := pkcePair()
	if code := authorize(t, carol, "client-grafana", challenge3); code.Message == "" {
		t.Fatal("new authorization after revocation succeeded")
	}
}

func TestApplicationGrantGroupMembershipRevocationRefresh(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "grafana-like", ApplicationGrantPrincipalGroup, "partner-org/team")
	application := grantTestApplication(t, "grafana-like")
	erin := grantTestUser(t, grantTestPartnerOrg, "erin")

	verifier, challenge := pkcePair()
	code := authorize(t, erin, "client-grafana", challenge)
	token, tokenError, err := GetAuthorizationCodeToken(application, "secret-grafana", code.Code, verifier, "", "en")
	if err != nil || tokenError != nil {
		t.Fatalf("code exchange: %v %+v", err, tokenError)
	}

	erin.Groups = []string{}
	if _, err := UpdateUser(erin.GetId(), erin, []string{"groups"}, true); err != nil {
		t.Fatal(err)
	}
	if active, _ := token.IsUserActive(); active {
		t.Fatal("token active after the group membership was removed")
	}
	refreshed, err := RefreshToken(application, "refresh_token", token.RefreshToken, "", "client-grafana", "secret-grafana", "", grantTestHost, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := refreshed.(*TokenError); !ok {
		t.Fatalf("refresh after group removal not refused: %+v", refreshed)
	}
}

func TestApplicationGrantIsolation(t *testing.T) {
	newApplicationGrantTestDb(t)
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")

	msg, err := CheckApplicationAccess(carol, grantTestApplication(t, "grafana-like"), "en")
	if err != nil || msg == "" {
		t.Fatal("a grant for one application gave access to another one")
	}
	_, challenge := pkcePair()
	if code := authorize(t, carol, "client-grafana", challenge); code.Message == "" {
		t.Fatal("authorization code issued for an application without a grant")
	}
	// the grant does not make the user a member of the resource organization
	if isUserOfApplication, _ := IsUserOfApplication(carol, grantTestApplication(t, "pve-like")); isUserOfApplication {
		t.Fatal("grant changed the native application membership")
	}
}

func TestApplicationGrantValidation(t *testing.T) {
	newApplicationGrantTestDb(t)
	pve := grantTestApplication(t, "pve-like")
	shared := grantTestApplication(t, "shared-app")
	future := time.Now().Add(time.Hour).Format(time.RFC3339)
	past := time.Now().Add(-time.Hour).Format(time.RFC3339)

	cases := []struct {
		name        string
		grant       ApplicationGrant
		application *Application
	}{
		{"foreign organization claims the application", ApplicationGrant{Owner: grantTestPartnerOrg, Application: "pve-like", PrincipalType: "User", Principal: "partner-org/carol"}, pve},
		{"principal of the application's organization", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "User", Principal: "resource-org/internal-bob"}, pve},
		{"shared application", ApplicationGrant{Owner: "built-in", Application: "shared-app", PrincipalType: "User", Principal: "partner-org/carol"}, shared},
		{"unknown user", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "User", Principal: "partner-org/nobody"}, pve},
		{"unknown group", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "Group", Principal: "partner-org/nobody"}, pve},
		{"unknown organization", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "User", Principal: "missing-org/carol"}, pve},
		{"malformed principal", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "User", Principal: "carol"}, pve},
		{"unknown principal type", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "Role", Principal: "partner-org/carol"}, pve},
		{"expiry in the past", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "User", Principal: "partner-org/carol", ExpireTime: past}, pve},
		{"unreadable expiry", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "User", Principal: "partner-org/carol", ExpireTime: "tomorrow"}, pve},
		{"missing application", ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "User", Principal: "partner-org/carol"}, nil},
	}
	for _, c := range cases {
		grant := c.grant
		if err := CheckApplicationGrant(&grant, c.application, "en"); err == nil {
			t.Errorf("%s: accepted", c.name)
		}
	}

	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	duplicate := ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "User", Principal: "partner-org/carol"}
	if err := CheckApplicationGrant(&duplicate, pve, "en"); err == nil {
		t.Error("duplicate grant accepted")
	}

	valid := ApplicationGrant{Owner: grantTestResourceOrg, Application: "pve-like", PrincipalType: "Group", Principal: "partner-org/team", ExpireTime: future, Name: "client-chosen", PrincipalId: "client-chosen"}
	if err := CheckApplicationGrant(&valid, pve, "en"); err != nil {
		t.Fatalf("valid grant refused: %v", err)
	}
	if valid.Name == "client-chosen" || valid.PrincipalId != "" || valid.CreatedTime == "" {
		t.Fatalf("server-side fields taken from the request: %+v", valid)
	}
}

func TestApplicationGrantExpiry(t *testing.T) {
	newApplicationGrantTestDb(t)
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")
	grant := addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	application := grantTestApplication(t, "pve-like")

	_, err := ormer.Engine.ID([]interface{}{grant.Owner, grant.Name}).Cols("expire_time").Update(&ApplicationGrant{ExpireTime: time.Now().Add(-time.Minute).Format(time.RFC3339)})
	if err != nil {
		t.Fatal(err)
	}
	if granted, _ := HasApplicationGrant(carol, application); granted {
		t.Fatal("expired grant still valid")
	}
}

func TestApplicationGrantMfaOfResourceOrganization(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	application := grantTestApplication(t, "pve-like")

	organization, _ := getOrganization("admin", grantTestResourceOrg)
	organization.MfaItems = []*MfaItem{{Name: TotpType, Rule: "Required"}}
	if _, err := ormer.Engine.ID([]interface{}{"admin", grantTestResourceOrg}).Cols("mfa_items").Update(organization); err != nil {
		t.Fatal(err)
	}
	application = grantTestApplication(t, "pve-like")

	carol := grantTestUser(t, grantTestPartnerOrg, "carol")
	// a user-level override of the user's own organization must not lower the bar
	carol.MfaItems = []*MfaItem{{Name: TotpType, Rule: "Optional"}}
	msg, err := CheckApplicationAccess(carol, application, "en")
	if err != nil || !strings.Contains(msg, "multi-factor") {
		t.Fatalf("granted user without MFA admitted to an MFA-required organization's application: %q %v", msg, err)
	}

	carol.TotpSecret = "synthetic-totp-secret"
	carol.PreferredMfaType = TotpType
	if msg, err := CheckApplicationAccess(carol, application, "en"); err != nil || msg != "" {
		t.Fatalf("granted user with MFA refused: %q %v", msg, err)
	}
	// the user is still authenticated with the MFA policy of their own organization
	home, _ := GetOrganizationByUser(carol)
	if home.Name != grantTestPartnerOrg {
		t.Fatalf("home organization changed: %s", home.Name)
	}
}

func TestApplicationGrantNameCollision(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	application := grantTestApplication(t, "pve-like")

	impostor := &User{Owner: grantTestResourceOrg, Name: "carol@partner-org", Id: util.GenerateId(), CreatedTime: util.GetCurrentTime()}
	mustAdd(t, func(u *User) (bool, error) { return AddUser(u, "en") }, impostor)

	if msg, _ := CheckApplicationAccess(grantTestUser(t, grantTestPartnerOrg, "carol"), application, "en"); msg == "" {
		t.Fatal("qualified name colliding with a resource user admitted")
	}
	grant := ApplicationGrant{Owner: grantTestResourceOrg, Application: "grafana-like", PrincipalType: "User", Principal: "partner-org/carol"}
	if err := CheckApplicationGrant(&grant, grantTestApplication(t, "grafana-like"), "en"); err == nil {
		t.Fatal("grant for a colliding qualified name accepted")
	}
}

func TestApplicationGrantRecreatedUserDoesNotInherit(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	application := grantTestApplication(t, "pve-like")

	// a deletion that bypassed the cascade, e.g. by an older Casdoor version
	if _, err := ormer.Engine.Where("owner = ? and name = ?", grantTestPartnerOrg, "carol").Delete(&User{}); err != nil {
		t.Fatal(err)
	}
	recreated := &User{Owner: grantTestPartnerOrg, Name: "carol", Id: util.GenerateId(), CreatedTime: util.GetCurrentTime()}
	mustAdd(t, func(u *User) (bool, error) { return AddUser(u, "en") }, recreated)

	if granted, _ := HasApplicationGrant(grantTestUser(t, grantTestPartnerOrg, "carol"), application); granted {
		t.Fatal("recreated user inherited the grant of the deleted one")
	}
}

func TestApplicationGrantCascades(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalGroup, "partner-org/pve-admins")
	addTestGrant(t, "grafana-like", ApplicationGrantPrincipalGroup, "partner-org/team")

	countGrants := func(application string) int {
		grants, err := GetApplicationGrants(grantTestResourceOrg, application)
		if err != nil {
			t.Fatal(err)
		}
		return len(grants)
	}

	if _, err := DeleteUser(grantTestUser(t, grantTestPartnerOrg, "carol")); err != nil {
		t.Fatal(err)
	}
	if countGrants("pve-like") != 1 {
		t.Fatal("deleting the user did not delete its grant")
	}

	team, _ := getGroup(grantTestPartnerOrg, "team")
	team.Name = "renamed-team"
	if _, err := UpdateGroup("partner-org/team", team, true, "en"); err != nil {
		t.Fatal(err)
	}
	grants, _ := GetApplicationGrants(grantTestResourceOrg, "grafana-like")
	if len(grants) != 1 || grants[0].Principal != "partner-org/renamed-team" {
		t.Fatalf("group rename not followed: %+v", grants)
	}
	if granted, _ := HasApplicationGrant(grantTestUser(t, grantTestPartnerOrg, "erin"), grantTestApplication(t, "grafana-like")); !granted {
		t.Fatal("member of the renamed group lost the grant")
	}

	dave := grantTestUser(t, grantTestPartnerOrg, "dave")
	dave.Groups = []string{}
	if _, err := UpdateUser(dave.GetId(), dave, []string{"groups"}, true); err != nil {
		t.Fatal(err)
	}
	group, _ := getGroup(grantTestPartnerOrg, "pve-admins")
	if _, err := DeleteGroup(group); err != nil {
		t.Fatal(err)
	}
	if countGrants("pve-like") != 0 {
		t.Fatal("deleting the group did not delete its grant")
	}

	if _, err := DeleteApplication(grantTestApplication(t, "grafana-like")); err != nil {
		t.Fatal(err)
	}
	if countGrants("grafana-like") != 0 {
		t.Fatal("deleting the application did not delete its grants")
	}
}

func TestApplicationGrantLoginResolution(t *testing.T) {
	newApplicationGrantTestDb(t)
	application := grantTestApplication(t, "pve-like")

	if user, _ := GetApplicationGrantUserForLogin(application, "carol", "en"); user != nil {
		t.Fatal("user resolved without a grant")
	}
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")

	for _, field := range []string{"carol", "carol@partner.example.test", "partner-org/carol"} {
		user, err := GetApplicationGrantUserForLogin(application, field, "en")
		if err != nil || user == nil || user.GetId() != "partner-org/carol" {
			t.Fatalf("%s: not resolved to the granted user: %v %v", field, user, err)
		}
	}
	for _, field := range []string{"dave", "partner-org/dave", "other-org/frank", "resource-org/internal-bob", "partner-org/nobody"} {
		if user, _ := GetApplicationGrantUserForLogin(application, field, "en"); user != nil {
			t.Fatalf("%s: resolved to %s without a grant", field, user.GetId())
		}
	}

	// two granted users with the same name in different organizations are not guessed between
	zed := &User{Owner: grantTestOtherOrg, Name: "carol", Id: util.GenerateId(), CreatedTime: util.GetCurrentTime()}
	mustAdd(t, func(u *User) (bool, error) { return AddUser(u, "en") }, zed)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "other-org/carol")
	if _, err := GetApplicationGrantUserForLogin(application, "carol", "en"); err == nil {
		t.Fatal("ambiguous name resolved")
	}
	if user, err := GetApplicationGrantUserForLogin(application, "other-org/carol", "en"); err != nil || user.GetId() != "other-org/carol" {
		t.Fatalf("qualified name not resolved: %v", err)
	}
}

func TestApplicationGrantSharedApplicationUnchanged(t *testing.T) {
	newApplicationGrantTestDb(t)
	carol := grantTestUser(t, grantTestPartnerOrg, "carol")

	_, challenge := pkcePair()
	code := authorize(t, carol, "client-shared-org-partner-org", challenge)
	if code.Message != "" || code.Token == nil {
		t.Fatalf("shared application refused: %q", code.Message)
	}
	if code.Token.IsGrantAccess {
		t.Fatal("shared application token marked as grant access")
	}
	claims := parseTestClaims(t, code.Token.AccessToken)
	if claims["name"] != "carol" {
		t.Fatalf("shared application claims changed: %v", claims["name"])
	}
}

func TestApplicationGrantGlobalAdminUnchanged(t *testing.T) {
	newApplicationGrantTestDb(t)
	admin := grantTestUser(t, "built-in", "admin")

	_, challenge := pkcePair()
	code := authorize(t, admin, "client-pve", challenge)
	if code.Message != "" || code.Token.IsGrantAccess {
		t.Fatalf("global admin flow changed: %q", code.Message)
	}
	if claims := parseTestClaims(t, code.Token.AccessToken); claims["name"] != "admin" {
		t.Fatalf("global admin claims changed: %v", claims["name"])
	}
}

func TestApplicationGrantPrincipalSearch(t *testing.T) {
	newApplicationGrantTestDb(t)
	application := grantTestApplication(t, "pve-like")

	found, err := SearchApplicationGrantPrincipals(application, "User", grantTestPartnerOrg, "car", false)
	if err != nil || len(found) != 1 || found[0].Id != "partner-org/carol" {
		t.Fatalf("substring search: %+v %v", found, err)
	}
	if found, _ := SearchApplicationGrantPrincipals(application, "User", grantTestPartnerOrg, "car", true); len(found) != 0 {
		t.Fatal("exact-only search matched part of a name")
	}
	if found, _ := SearchApplicationGrantPrincipals(application, "User", grantTestPartnerOrg, "carol", true); len(found) != 1 {
		t.Fatal("exact-only search missed the exact name")
	}
	if found, _ := SearchApplicationGrantPrincipals(application, "User", grantTestPartnerOrg, "%", false); len(found) != 0 {
		t.Fatal("LIKE wildcard not escaped")
	}
	if found, _ := SearchApplicationGrantPrincipals(application, "Group", grantTestPartnerOrg, "team", true); len(found) != 1 || found[0].Id != "partner-org/team" {
		t.Fatalf("group search: %+v", found)
	}
	// the application's own users may get application roles, its admins search them like any of their users
	if found, _ := SearchApplicationGrantPrincipals(application, "User", grantTestResourceOrg, "internal", true); len(found) != 2 {
		t.Fatalf("search of the application's own organization: %+v", found)
	}
}

func TestApplicationGrantSubjectCollision(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalGroup, "partner-org/pve-admins")
	application := grantTestApplication(t, "pve-like")
	bob := grantTestUser(t, grantTestResourceOrg, "internal-bob")

	// an admin of the partner organization gives a member of the granted group the ID of a
	// user of the resource organization, which would be the same "sub" at the relying party
	dave := grantTestUser(t, grantTestPartnerOrg, "dave")
	if _, err := ormer.Engine.Where("owner = ? and name = ?", dave.Owner, dave.Name).Cols("id").Update(&User{Id: bob.Id}); err != nil {
		t.Fatal(err)
	}
	if msg, _ := CheckApplicationAccess(grantTestUser(t, grantTestPartnerOrg, "dave"), application, "en"); msg == "" {
		t.Fatal("user sharing the ID of another user admitted")
	}
	grant := ApplicationGrant{Owner: grantTestResourceOrg, Application: "grafana-like", PrincipalType: "User", Principal: "partner-org/dave"}
	if err := CheckApplicationGrant(&grant, grantTestApplication(t, "grafana-like"), "en"); err == nil {
		t.Fatal("grant for a user with a duplicate ID accepted")
	}
}

func TestApplicationGrantClaimsDropHomeAdminFlag(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "grafana-like", ApplicationGrantPrincipalUser, "partner-org/partner-admin")
	admin := grantTestUser(t, grantTestPartnerOrg, "partner-admin")

	_, challenge := pkcePair()
	code := authorize(t, admin, "client-grafana", challenge)
	if code.Message != "" {
		t.Fatal(code.Message)
	}
	if claims := parseTestClaims(t, code.Token.AccessToken); claims["isAdmin"] != false {
		t.Fatalf("admin flag of the home organization carried into the application: %v", claims["isAdmin"])
	}
}

func TestApplicationGrantCaseInsensitiveNameCollision(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	impostor := &User{Owner: grantTestResourceOrg, Name: "Carol@Partner-Org", Id: util.GenerateId(), CreatedTime: util.GetCurrentTime()}
	mustAdd(t, func(u *User) (bool, error) { return AddUser(u, "en") }, impostor)

	if msg, _ := CheckApplicationAccess(grantTestUser(t, grantTestPartnerOrg, "carol"), grantTestApplication(t, "pve-like"), "en"); msg == "" {
		t.Fatal("qualified name colliding without case admitted")
	}
}

func TestApplicationGrantFollowsApplicationRename(t *testing.T) {
	newApplicationGrantTestDb(t)
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")
	application := grantTestApplication(t, "pve-like")
	application.Name = "pve-renamed"
	if _, err := UpdateApplication("admin/pve-like", application, true, "en", nil); err != nil {
		t.Fatal(err)
	}
	if grants, _ := GetApplicationGrants(grantTestResourceOrg, "pve-renamed"); len(grants) != 1 {
		t.Fatal("grant did not follow the renamed application")
	}
	if grants, _ := GetApplicationGrants(grantTestResourceOrg, "pve-like"); len(grants) != 0 {
		t.Fatal("grant left behind for the old application name")
	}
}

func TestApplicationGrantOrganizationDeletionIsExact(t *testing.T) {
	newApplicationGrantTestDb(t)
	mustAdd(t, AddOrganization, &Organization{Owner: "admin", Name: "partner_org", DisplayName: "partner_org", CreatedTime: util.GetCurrentTime()})
	addTestGrant(t, "pve-like", ApplicationGrantPrincipalUser, "partner-org/carol")

	if err := deleteApplicationGrantsByOrganization("partner_org"); err != nil {
		t.Fatal(err)
	}
	if grants, _ := GetApplicationGrants(grantTestResourceOrg, "pve-like"); len(grants) != 1 {
		t.Fatal("deleting one organization removed the grants of a similarly named one")
	}
	if err := deleteApplicationGrantsByOrganization(grantTestPartnerOrg); err != nil {
		t.Fatal(err)
	}
	if grants, _ := GetApplicationGrants(grantTestResourceOrg, "pve-like"); len(grants) != 0 {
		t.Fatal("grants of the deleted principal organization kept")
	}
}
