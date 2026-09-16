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
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	MagicLinkStatusCreated = "created"
	MagicLinkStatusSent    = "sent"
	MagicLinkStatusUsed    = "used"
	MagicLinkStatusExpired = "expired"
	MagicLinkStatusFailed  = "failed"
	MagicLinkStatusRevoked = "revoked"

	MagicLinkAuthActionSigninExistingUser = "signin_existing_user"
	MagicLinkAuthActionSignupNewUser      = "signup_new_user"

	// MagicLinkProviderRule is the Email provider rule a magic link mail is sent
	// with, an application without such a row falls back to its "All" row.
	MagicLinkProviderRule = "magicLink"

	MagicLinkDefaultExpireMinutes = 10
	MagicLinkMinExpireMinutes     = 2
	MagicLinkMaxExpireMinutes     = 43200

	MagicLinkDefaultRateLimitEmail       = 3
	MagicLinkDefaultRateLimitIp          = 10
	MagicLinkDefaultRateLimitApplication = 100
)

// MagicLink is one issued sign-in link. Only the SHA-256 hash of the token is
// stored, so the link in the mailbox is the only copy of the secret.
type MagicLink struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Application string `xorm:"varchar(100) index" json:"application"`
	Email       string `xorm:"varchar(100) index" json:"email"`
	Requester   string `xorm:"varchar(100)" json:"requester"`
	RemoteAddr  string `xorm:"varchar(100)" json:"remoteAddr"`
	AuthAction  string `xorm:"varchar(100)" json:"authAction"`
	TokenHash   string `xorm:"varchar(100) index" json:"-"`
	Status      string `xorm:"varchar(20) index" json:"status"`
	ExpireTime  string `xorm:"varchar(100)" json:"expireTime"`
	ExpireAt    int64  `xorm:"index" json:"expireAt"`
	UsedTime    string `xorm:"varchar(100)" json:"usedTime"`
	LastError   string `xorm:"varchar(500)" json:"lastError"`

	ClientId            string `xorm:"varchar(100)" json:"clientId"`
	ResponseType        string `xorm:"varchar(100)" json:"responseType"`
	RedirectUri         string `xorm:"varchar(500)" json:"redirectUri"`
	Scope               string `xorm:"varchar(1000)" json:"scope"`
	State               string `xorm:"varchar(1000)" json:"state"`
	Nonce               string `xorm:"varchar(1000)" json:"nonce"`
	CodeChallengeMethod string `xorm:"varchar(100)" json:"codeChallengeMethod"`
	CodeChallenge       string `xorm:"varchar(500)" json:"codeChallenge"`
}

func (application *Application) IsMagicLinkEnabled() bool {
	return application != nil && application.EnableMagicLink
}

func (application *Application) IsMagicLinkSignupEnabled() bool {
	return application.IsMagicLinkEnabled() && application.EnableMagicLinkSignup
}

func (application *Application) GetMagicLinkExpireMinutes() int {
	if application == nil || application.MagicLinkExpireMinutes <= 0 {
		return MagicLinkDefaultExpireMinutes
	}
	return application.MagicLinkExpireMinutes
}

func (application *Application) GetMagicLinkRateLimitEmail() int {
	if application == nil || application.MagicLinkRateLimitEmail <= 0 {
		return MagicLinkDefaultRateLimitEmail
	}
	return application.MagicLinkRateLimitEmail
}

func (application *Application) GetMagicLinkRateLimitIp() int {
	if application == nil || application.MagicLinkRateLimitIp <= 0 {
		return MagicLinkDefaultRateLimitIp
	}
	return application.MagicLinkRateLimitIp
}

func (application *Application) GetMagicLinkRateLimitApplication() int {
	if application == nil || application.MagicLinkRateLimitApplication <= 0 {
		return MagicLinkDefaultRateLimitApplication
	}
	return application.MagicLinkRateLimitApplication
}

func GenerateMagicLinkToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashMagicLinkToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func GetMagicLinkCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&MagicLink{Owner: owner})
}

func GetMagicLinks(owner string) ([]*MagicLink, error) {
	magicLinks := []*MagicLink{}
	err := ormer.Engine.Desc("created_time").Find(&magicLinks, &MagicLink{Owner: owner})
	if err != nil {
		return nil, err
	}

	return magicLinks, nil
}

func GetPaginationMagicLinks(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*MagicLink, error) {
	magicLinks := []*MagicLink{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&magicLinks, &MagicLink{Owner: owner})
	if err != nil {
		return nil, err
	}

	return magicLinks, nil
}

func GetMagicLink(id string) (*MagicLink, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	magicLink := MagicLink{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&magicLink)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, nil
	}

	return &magicLink, nil
}

func AddMagicLink(magicLink *MagicLink) (bool, error) {
	affected, err := ormer.Engine.Insert(magicLink)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func UpdateMagicLinkStatus(magicLink *MagicLink, status string, lastError string) error {
	magicLink.Status = status
	magicLink.LastError = lastError
	columns := []string{"status", "last_error"}
	if status == MagicLinkStatusUsed && magicLink.UsedTime == "" {
		magicLink.UsedTime = util.GetCurrentTime()
		columns = append(columns, "used_time")
	}

	_, err := ormer.Engine.ID(core.PK{magicLink.Owner, magicLink.Name}).Cols(columns...).Update(magicLink)
	return err
}

// RevokeMagicLink invalidates a link that has not been used yet.
func RevokeMagicLink(id string) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	magicLink := &MagicLink{Status: MagicLinkStatusRevoked}
	affected, err := ormer.Engine.ID(core.PK{owner, name}).In("status", MagicLinkStatusCreated, MagicLinkStatusSent).Cols("status").Update(magicLink)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteMagicLink(id string) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	affected, err := ormer.Engine.ID(core.PK{owner, name}).Delete(&MagicLink{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

// GetMagicLinkPendingCounts counts the links of an application that are still
// usable, per email, per client IP and in total. They are the rate limit's unit:
// a link that has been used, revoked or has expired no longer counts.
func GetMagicLinkPendingCounts(application *Application, email string, remoteAddr string) (int64, int64, int64, error) {
	applicationId := application.GetId()
	now := time.Now().Unix()

	emailCount, err := ormer.Engine.Where("application = ?", applicationId).And("email = ?", email).And("expire_at > ?", now).In("status", MagicLinkStatusCreated, MagicLinkStatusSent).Count(&MagicLink{})
	if err != nil {
		return 0, 0, 0, err
	}

	ipCount, err := ormer.Engine.Where("application = ?", applicationId).And("remote_addr = ?", remoteAddr).And("expire_at > ?", now).In("status", MagicLinkStatusCreated, MagicLinkStatusSent).Count(&MagicLink{})
	if err != nil {
		return 0, 0, 0, err
	}

	applicationCount, err := ormer.Engine.Where("application = ?", applicationId).And("expire_at > ?", now).In("status", MagicLinkStatusCreated, MagicLinkStatusSent).Count(&MagicLink{})
	if err != nil {
		return 0, 0, 0, err
	}

	return emailCount, ipCount, applicationCount, nil
}

func CheckMagicLinkRateLimit(application *Application, email string, remoteAddr string) error {
	emailCount, ipCount, applicationCount, err := GetMagicLinkPendingCounts(application, email, remoteAddr)
	if err != nil {
		return err
	}

	return checkMagicLinkRateLimitCounts(application, emailCount, ipCount, applicationCount)
}

func checkMagicLinkRateLimitCounts(application *Application, emailCount int64, ipCount int64, applicationCount int64) error {
	if emailCount >= int64(application.GetMagicLinkRateLimitEmail()) {
		return fmt.Errorf("too many magic links have been requested for this email")
	}
	if ipCount >= int64(application.GetMagicLinkRateLimitIp()) {
		return fmt.Errorf("too many magic links have been requested from this IP")
	}
	if applicationCount >= int64(application.GetMagicLinkRateLimitApplication()) {
		return fmt.Errorf("too many magic links have been requested for this application")
	}

	return nil
}

// ValidateMagicLinkConfig rejects an application whose magic link settings
// contradict each other, it is called when an application is added or updated.
func ValidateMagicLinkConfig(application *Application) error {
	if application == nil {
		return nil
	}

	if application.EnableMagicLinkSignup && !application.EnableMagicLink {
		return fmt.Errorf("enableMagicLinkSignup requires enableMagicLink")
	}
	if application.MagicLinkExpireMinutes != 0 && (application.MagicLinkExpireMinutes < MagicLinkMinExpireMinutes || application.MagicLinkExpireMinutes > MagicLinkMaxExpireMinutes) {
		return fmt.Errorf("magicLinkExpireMinutes must be between %d and %d", MagicLinkMinExpireMinutes, MagicLinkMaxExpireMinutes)
	}
	if application.MagicLinkRateLimitEmail < 0 {
		return fmt.Errorf("magicLinkRateLimitEmail must be greater than or equal to 0")
	}
	if application.MagicLinkRateLimitIp < 0 {
		return fmt.Errorf("magicLinkRateLimitIp must be greater than or equal to 0")
	}
	if application.MagicLinkRateLimitApplication < 0 {
		return fmt.Errorf("magicLinkRateLimitApplication must be greater than or equal to 0")
	}

	return nil
}

func NewMagicLink(application *Application, email string, remoteAddr string, requester string, token string, oauth map[string]string, expireAt time.Time) *MagicLink {
	if expireAt.IsZero() {
		expireAt = time.Now().Add(time.Duration(application.GetMagicLinkExpireMinutes()) * time.Minute)
	}

	magicLink := &MagicLink{
		Owner:               application.Organization,
		Name:                util.GenerateId(),
		CreatedTime:         util.GetCurrentTime(),
		Application:         application.GetId(),
		Email:               email,
		Requester:           requester,
		RemoteAddr:          remoteAddr,
		Status:              MagicLinkStatusCreated,
		ExpireTime:          util.Time2String(expireAt),
		ExpireAt:            expireAt.Unix(),
		ClientId:            oauth["clientId"],
		ResponseType:        oauth["responseType"],
		RedirectUri:         oauth["redirectUri"],
		Scope:               oauth["scope"],
		State:               oauth["state"],
		Nonce:               oauth["nonce"],
		CodeChallengeMethod: oauth["codeChallengeMethod"],
		CodeChallenge:       oauth["codeChallenge"],
	}
	if magicLink.ResponseType == "" {
		magicLink.ResponseType = "login"
	}
	if token != "" {
		magicLink.TokenHash = HashMagicLinkToken(token)
	}

	return magicLink
}

func BuildMagicLinkUrl(magicLink *MagicLink, token string, host string) string {
	originFrontend, _ := getOriginFromHost(host)

	query := url.Values{}
	query.Set("token", token)
	for key, value := range map[string]string{
		"clientId":              magicLink.ClientId,
		"responseType":          magicLink.ResponseType,
		"redirectUri":           magicLink.RedirectUri,
		"scope":                 magicLink.Scope,
		"state":                 magicLink.State,
		"nonce":                 magicLink.Nonce,
		"code_challenge_method": magicLink.CodeChallengeMethod,
		"code_challenge":        magicLink.CodeChallenge,
	} {
		if value != "" {
			query.Set(key, value)
		}
	}

	return fmt.Sprintf("%s/magic-link/callback?%s", strings.TrimRight(originFrontend, "/"), query.Encode())
}

func SendMagicLinkEmail(organization *Organization, provider *Provider, email string, magicLinkUrl string, expireTime string) error {
	title := provider.Title
	if title == "" {
		title = "Magic Link"
	}

	content := provider.MagicLinkContent
	if content == "" {
		content = GetDefaultMagicLinkEmailContent()
	}
	content = strings.ReplaceAll(content, "%link", magicLinkUrl)
	content = strings.ReplaceAll(content, "%expireTime", expireTime)

	return SendEmail(provider, title, content, []string{email}, organization.DisplayName)
}

// GetDefaultMagicLinkEmailContent is the mail body an Email provider without a
// "Magic link content" of its own falls back to. "%link" is the sign-in link and
// "%expireTime" the moment it stops working.
func GetDefaultMagicLinkEmailContent() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Magic Link</title>
</head>
<body style="margin: 0; padding: 0; font-family: Arial, Helvetica, sans-serif; color: #1a1a1a;">
  <div style="max-width: 600px; margin: 0 auto; padding: 24px;">
    <div style="font-size: 18px; font-weight: 600; margin-bottom: 12px; text-align: center;">
      Sign in to your account
    </div>
    <div style="font-size: 14px; margin-bottom: 20px; line-height: 1.5; text-align: center;">
      Use the one-time link below to sign in.
    </div>
    <div style="text-align: center; margin: 28px 0;">
      <a href="%link" style="display: inline-block; background: #5734d3; color: #ffffff; font-size: 15px; font-weight: 700; text-decoration: none; padding: 14px 28px; border-radius: 8px;">Sign in</a>
    </div>
    <div style="text-align: center; font-size: 14px; word-break: break-all;">
      Or open this link: <a href="%link">%link</a>
    </div>
    <div style="font-size: 14px; margin-top: 24px; text-align: center;">
      The link can be used once and expires at %expireTime.
    </div>
    <div style="font-size: 12px; text-align: center; color: #777; margin-top: 36px; line-height: 1.4;">
      If you did not request this email, you can safely ignore it.
    </div>
  </div>
</body>
</html>`
}

// ConsumeMagicLink claims a link for a single sign-in. The claim is one
// conditional UPDATE, so two parallel requests with the same token cannot both
// win it.
func ConsumeMagicLink(token string) (*MagicLink, error) {
	tokenHash := HashMagicLinkToken(token)
	magicLink := &MagicLink{TokenHash: tokenHash}
	existed, err := ormer.Engine.Get(magicLink)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, fmt.Errorf("the magic link is invalid")
	}

	nowUnix := time.Now().Unix()
	err = CheckMagicLinkState(magicLink, nowUnix)
	if err != nil {
		if err.Error() == "the magic link has expired" {
			magicLink.Status = MagicLinkStatusExpired
			_, _ = ormer.Engine.ID(core.PK{magicLink.Owner, magicLink.Name}).Cols("status").Update(magicLink)
		}
		return nil, err
	}

	now := util.GetCurrentTime()
	claimed := &MagicLink{
		Status:   MagicLinkStatusUsed,
		UsedTime: now,
	}
	affected, err := ormer.Engine.Where("token_hash = ?", tokenHash).And("expire_at > ?", nowUnix).In("status", MagicLinkStatusCreated, MagicLinkStatusSent).Cols("status", "used_time").Update(claimed)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, fmt.Errorf("the magic link has already been used")
	}

	magicLink.Status = MagicLinkStatusUsed
	magicLink.UsedTime = now
	return magicLink, nil
}

func CheckMagicLinkState(magicLink *MagicLink, nowUnix int64) error {
	switch {
	case magicLink.Status == MagicLinkStatusRevoked:
		return fmt.Errorf("the magic link has been revoked")
	case magicLink.Status == MagicLinkStatusUsed:
		return fmt.Errorf("the magic link has already been used")
	case magicLink.Status == MagicLinkStatusExpired || magicLink.ExpireAt <= nowUnix:
		return fmt.Errorf("the magic link has expired")
	case magicLink.Status != MagicLinkStatusCreated && magicLink.Status != MagicLinkStatusSent:
		return fmt.Errorf("the magic link is invalid")
	}

	return nil
}

// ResolveMagicLinkExpireTime is the moment a link issued now stops working.
func ResolveMagicLinkExpireTime(application *Application, now time.Time) time.Time {
	if now.IsZero() {
		now = time.Now()
	}

	expireMinutes := application.GetMagicLinkExpireMinutes()
	if expireMinutes < MagicLinkMinExpireMinutes {
		expireMinutes = MagicLinkMinExpireMinutes
	}
	if expireMinutes > MagicLinkMaxExpireMinutes {
		expireMinutes = MagicLinkMaxExpireMinutes
	}

	return now.Add(time.Duration(expireMinutes) * time.Minute)
}
