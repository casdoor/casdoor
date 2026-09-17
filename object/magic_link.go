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
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
)

// MagicLink is a one-time sign-in link that was mailed to an email address. Only the
// hash of the token is stored, so the link in the mailbox is the only copy of the
// secret, and only the hash of the sign-in session's secret, which ties the link to
// the browser that asked for it.
type MagicLink struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Application string `xorm:"varchar(100) notnull" json:"application"`
	Email       string `xorm:"varchar(100) index notnull" json:"email"`
	RemoteAddr  string `xorm:"varchar(100)" json:"remoteAddr"`
	TokenHash   string `xorm:"varchar(100) index notnull" json:"-"`
	SessionHash string `xorm:"varchar(100) notnull" json:"-"`
	Time        int64  `xorm:"notnull" json:"time"`
	IsUsed      bool   `xorm:"notnull" json:"isUsed"`
}

func generateMagicLinkToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashMagicLinkSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

func getMagicLinkTimeout() int64 {
	return conf.GetVerificationCodeTimeout()
}

func isLoopbackHost(host string) bool {
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	hostname = strings.ToLower(strings.Trim(hostname, "[]"))
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || strings.HasSuffix(hostname, ".localhost")
}

// getMagicLinkOrigin is the site the sign-in link points at. It is never taken from
// the request's "Host" header alone: a forged header would mail the one-time token to
// the attacker's own site, so without a configured origin only a loopback host, which
// nobody else can receive, is trusted.
func getMagicLinkOrigin(host string, lang string) (string, error) {
	if conf.GetConfigString("origin") == "" && conf.GetConfigString("originFrontend") == "" && !isLoopbackHost(host) {
		return "", errors.New(i18n.Translate(lang, "verification:please set \"origin\" in conf/app.conf to send magic links"))
	}

	originFrontend, _ := getOriginFromHost(host)
	return originFrontend, nil
}

// isValidSigninPath keeps the link on Casdoor's own sign-in pages, the path is the one
// the browser that asked for the link was on.
func isValidSigninPath(signinPath string) bool {
	if signinPath == "" || len(signinPath) > 1000 || strings.ContainsAny(signinPath, "#\\") {
		return false
	}

	signinUrl, err := url.ParseRequestURI(signinPath)
	if err != nil || signinUrl.Scheme != "" || signinUrl.Host != "" {
		return false
	}

	path := signinUrl.Path
	return path == "/login" || strings.HasPrefix(path, "/login/") || strings.HasPrefix(path, "/cas/")
}

func getMagicLinkUrl(origin string, signinPath string, application *Application, token string) string {
	if !isValidSigninPath(signinPath) {
		signinPath = fmt.Sprintf("/login/%s", application.Name)
	}

	signinUrl, err := url.ParseRequestURI(signinPath)
	if err != nil {
		return ""
	}

	query := signinUrl.Query()
	query.Set("magicLinkToken", token)
	signinUrl.RawQuery = query.Encode()

	return fmt.Sprintf("%s%s", strings.TrimSuffix(origin, "/"), signinUrl.String())
}

func getMagicLinkEmailContent(provider *Provider, magicLinkUrl string, user *User) string {
	content := provider.Content
	if !strings.Contains(content, "%link") {
		content = getDefaultMagicLinkEmailContent()
	}

	content = strings.ReplaceAll(content, "<reset-link>", "")
	content = strings.ReplaceAll(content, "</reset-link>", "")
	content = strings.ReplaceAll(content, "%link", magicLinkUrl)
	content = strings.ReplaceAll(content, "%expireTime", fmt.Sprintf("%d", getMagicLinkTimeout()))

	userString := "Hi"
	if user != nil {
		userString = user.GetFriendlyName()
	}
	return strings.Replace(content, "%{user.friendlyName}", userString, 1)
}

// getDefaultMagicLinkEmailContent is the mail an Email provider without a "%link" in
// its own content falls back to.
func getDefaultMagicLinkEmailContent() string {
	return `<p>You have requested a sign-in link at Casdoor.</p>
<p><a href="%link">Click here to sign in</a></p>
<p>Or open this link: %link</p>
<p>The link can only be used once, in the browser you asked for it, and expires in %expireTime minutes. If you did not request it, you can ignore this email.</p>`
}

// isAllowSendMagicLink throttles the links the same way IsAllowSend() throttles the
// verification codes, an issued link leaves no verification record of its own.
func isAllowSendMagicLink(application *Application, email string, remoteAddr string) error {
	resendTimeoutInSeconds := int64(60)
	if application != nil && application.CodeResendTimeout > 0 {
		resendTimeoutInSeconds = int64(application.CodeResendTimeout)
	}

	now := time.Now().Unix()
	for _, magicLink := range []*MagicLink{{Email: email}, {RemoteAddr: remoteAddr}} {
		if magicLink.Email == "" && magicLink.RemoteAddr == "" {
			continue
		}

		has, err := ormer.Engine.Desc("time").Get(magicLink)
		if err != nil {
			return err
		}

		if has && now-magicLink.Time < resendTimeoutInSeconds {
			return fmt.Errorf("you can only send one code in %ds", resendTimeoutInSeconds)
		}
	}

	return nil
}

func addMagicLink(organization *Organization, application *Application, email string, remoteAddr string, token string, sessionHash string) error {
	magicLink := &MagicLink{
		Owner:       organization.Name,
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		Application: application.Name,
		Email:       email,
		RemoteAddr:  remoteAddr,
		TokenHash:   HashMagicLinkSecret(token),
		SessionHash: sessionHash,
		Time:        time.Now().Unix(),
		IsUsed:      false,
	}

	_, err := ormer.Engine.Insert(magicLink)
	return err
}

// SendMagicLinkToEmail mails a one-time sign-in link. The link is bound to the browser
// that asked for it through sessionHash, so a link that leaves the mailbox is of no use
// anywhere else.
func SendMagicLinkToEmail(organization *Organization, user *User, provider *Provider, remoteAddr string, dest string, host string, signinPath string, application *Application, sessionHash string, lang string) error {
	origin, err := getMagicLinkOrigin(host, lang)
	if err != nil {
		return err
	}

	if sessionHash == "" {
		return errors.New(i18n.Translate(lang, "verification:Please open the magic link in the browser you requested it from"))
	}

	err = IsAllowSend(user, remoteAddr, provider.Category, application)
	if err != nil {
		return err
	}

	err = isAllowSendMagicLink(application, dest, remoteAddr)
	if err != nil {
		return err
	}

	token, err := generateMagicLinkToken()
	if err != nil {
		return err
	}

	title := provider.Title
	if title == "" {
		title = "Magic Link"
	}

	content := getMagicLinkEmailContent(provider, getMagicLinkUrl(origin, signinPath, application, token), user)
	err = SendEmail(provider, title, content, []string{dest}, organization.DisplayName)
	if err != nil {
		return err
	}

	return addMagicLink(organization, application, dest, remoteAddr, token, sessionHash)
}

// ConsumeMagicLink claims a link for a single sign-in. The claim is one conditional
// UPDATE, so two requests carrying the same token cannot both win it.
func ConsumeMagicLink(token string, sessionHash string, application *Application, lang string) (*MagicLink, error) {
	if token == "" {
		return nil, errors.New(i18n.Translate(lang, "verification:The magic link is invalid"))
	}

	magicLink := &MagicLink{TokenHash: HashMagicLinkSecret(token)}
	existed, err := ormer.Engine.Get(magicLink)
	if err != nil {
		return nil, err
	}
	if !existed || magicLink.Application != application.Name || magicLink.Owner != application.Organization {
		return nil, errors.New(i18n.Translate(lang, "verification:The magic link is invalid"))
	}
	if magicLink.IsUsed {
		return nil, errors.New(i18n.Translate(lang, "verification:The magic link has already been used"))
	}
	if time.Now().Unix()-magicLink.Time > getMagicLinkTimeout()*60 {
		return nil, errors.New(i18n.Translate(lang, "verification:The magic link has expired"))
	}
	if sessionHash == "" || sessionHash != magicLink.SessionHash {
		return nil, errors.New(i18n.Translate(lang, "verification:Please open the magic link in the browser you requested it from"))
	}

	affected, err := ormer.Engine.Where("token_hash = ?", magicLink.TokenHash).And("is_used = ?", false).Cols("is_used").Update(&MagicLink{IsUsed: true})
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, errors.New(i18n.Translate(lang, "verification:The magic link has already been used"))
	}

	magicLink.IsUsed = true
	return magicLink, nil
}

// CheckMagicLinkSignup rejects a magic link signup for an application that asks the
// signup page for more than the link itself can answer.
func CheckMagicLinkSignup(application *Application, lang string) error {
	if !application.EnableSignUp {
		return errors.New(i18n.Translate(lang, "account:The application does not allow to sign up new account"))
	}

	for _, signupItem := range application.SignupItems {
		if signupItem == nil || !signupItem.Required {
			continue
		}

		switch signupItem.Name {
		case "ID", "Username", "Display name", "Email", "Password", "Confirm password", "Agreement", "Invitation code", "Signup button", "Providers":
			continue
		default:
			return fmt.Errorf(i18n.Translate(lang, "verification:The signup item: %s is required, so the magic link cannot sign up a new user"), signupItem.Name)
		}
	}

	return nil
}
