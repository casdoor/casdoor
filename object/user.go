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
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/xorm-io/core"
)

const (
	UserPropertiesWechatUnionId = "wechatUnionId"
	UserPropertiesWechatOpenId  = "wechatOpenId"
)

type User struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100) index" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	Id                     string   `xorm:"varchar(100) index" json:"id"`
	Type                   string   `xorm:"varchar(100)" json:"type"`
	Password               string   `xorm:"varchar(100)" json:"password"`
	PasswordChangeRequired bool     `xorm:"varchar(100)" json:"passwordChangeRequired"`
	PasswordSalt           string   `xorm:"varchar(100)" json:"passwordSalt"`
	PasswordType           string   `xorm:"varchar(100)" json:"passwordType"`
	DisplayName            string   `xorm:"varchar(100)" json:"displayName"`
	FirstName              string   `xorm:"varchar(100)" json:"firstName"`
	LastName               string   `xorm:"varchar(100)" json:"lastName"`
	Avatar                 string   `xorm:"varchar(500)" json:"avatar"`
	AvatarType             string   `xorm:"varchar(100)" json:"avatarType"`
	PermanentAvatar        string   `xorm:"varchar(500)" json:"permanentAvatar"`
	Email                  string   `xorm:"varchar(100) index" json:"email"`
	EmailVerified          bool     `json:"emailVerified"`
	Phone                  string   `xorm:"varchar(20) index" json:"phone"`
	CountryCode            string   `xorm:"varchar(6)" json:"countryCode"`
	Region                 string   `xorm:"varchar(100)" json:"region"`
	Location               string   `xorm:"varchar(100)" json:"location"`
	Address                []string `json:"address"`
	Affiliation            string   `xorm:"varchar(100)" json:"affiliation"`
	Title                  string   `xorm:"varchar(100)" json:"title"`
	IdCardType             string   `xorm:"varchar(100)" json:"idCardType"`
	IdCard                 string   `xorm:"varchar(100) index" json:"idCard"`
	Homepage               string   `xorm:"varchar(100)" json:"homepage"`
	Bio                    string   `xorm:"varchar(100)" json:"bio"`
	Tag                    string   `xorm:"varchar(100)" json:"tag"`
	Language               string   `xorm:"varchar(100)" json:"language"`
	Gender                 string   `xorm:"varchar(100)" json:"gender"`
	Birthday               string   `xorm:"varchar(100)" json:"birthday"`
	Education              string   `xorm:"varchar(100)" json:"education"`
	Score                  int      `json:"score"`
	Karma                  int      `json:"karma"`
	Ranking                int      `json:"ranking"`
	IsDefaultAvatar        bool     `json:"isDefaultAvatar"`
	IsOnline               bool     `json:"isOnline"`
	IsAdmin                bool     `json:"isAdmin"`
	IsGlobalAdmin          bool     `json:"isGlobalAdmin"`
	IsForbidden            bool     `json:"isForbidden"`
	IsDeleted              bool     `json:"isDeleted"`
	SignupApplication      string   `xorm:"varchar(100)" json:"signupApplication"`
	Hash                   string   `xorm:"varchar(100)" json:"hash"`
	PreHash                string   `xorm:"varchar(100)" json:"preHash"`
	AccessKey              string   `xorm:"varchar(100)" json:"accessKey"`
	AccessSecret           string   `xorm:"varchar(100)" json:"accessSecret"`

	CreatedIp      string `xorm:"varchar(100)" json:"createdIp"`
	LastSigninTime string `xorm:"varchar(100)" json:"lastSigninTime"`
	LastSigninIp   string `xorm:"varchar(100)" json:"lastSigninIp"`

	GitHub          string `xorm:"github varchar(100)" json:"github"`
	Google          string `xorm:"varchar(100)" json:"google"`
	QQ              string `xorm:"qq varchar(100)" json:"qq"`
	WeChat          string `xorm:"wechat varchar(100)" json:"wechat"`
	Facebook        string `xorm:"facebook varchar(100)" json:"facebook"`
	DingTalk        string `xorm:"dingtalk varchar(100)" json:"dingtalk"`
	Weibo           string `xorm:"weibo varchar(100)" json:"weibo"`
	Gitee           string `xorm:"gitee varchar(100)" json:"gitee"`
	LinkedIn        string `xorm:"linkedin varchar(100)" json:"linkedin"`
	Wecom           string `xorm:"wecom varchar(100)" json:"wecom"`
	Lark            string `xorm:"lark varchar(100)" json:"lark"`
	Gitlab          string `xorm:"gitlab varchar(100)" json:"gitlab"`
	Adfs            string `xorm:"adfs varchar(100)" json:"adfs"`
	Baidu           string `xorm:"baidu varchar(100)" json:"baidu"`
	Alipay          string `xorm:"alipay varchar(100)" json:"alipay"`
	Casdoor         string `xorm:"casdoor varchar(100)" json:"casdoor"`
	Infoflow        string `xorm:"infoflow varchar(100)" json:"infoflow"`
	Apple           string `xorm:"apple varchar(100)" json:"apple"`
	AzureAD         string `xorm:"azuread varchar(100)" json:"azuread"`
	Slack           string `xorm:"slack varchar(100)" json:"slack"`
	Steam           string `xorm:"steam varchar(100)" json:"steam"`
	Bilibili        string `xorm:"bilibili varchar(100)" json:"bilibili"`
	Okta            string `xorm:"okta varchar(100)" json:"okta"`
	Douyin          string `xorm:"douyin varchar(100)" json:"douyin"`
	Line            string `xorm:"line varchar(100)" json:"line"`
	Amazon          string `xorm:"amazon varchar(100)" json:"amazon"`
	Auth0           string `xorm:"auth0 varchar(100)" json:"auth0"`
	BattleNet       string `xorm:"battlenet varchar(100)" json:"battlenet"`
	Bitbucket       string `xorm:"bitbucket varchar(100)" json:"bitbucket"`
	Box             string `xorm:"box varchar(100)" json:"box"`
	CloudFoundry    string `xorm:"cloudfoundry varchar(100)" json:"cloudfoundry"`
	Dailymotion     string `xorm:"dailymotion varchar(100)" json:"dailymotion"`
	Deezer          string `xorm:"deezer varchar(100)" json:"deezer"`
	DigitalOcean    string `xorm:"digitalocean varchar(100)" json:"digitalocean"`
	Discord         string `xorm:"discord varchar(100)" json:"discord"`
	Dropbox         string `xorm:"dropbox varchar(100)" json:"dropbox"`
	EveOnline       string `xorm:"eveonline varchar(100)" json:"eveonline"`
	Fitbit          string `xorm:"fitbit varchar(100)" json:"fitbit"`
	Gitea           string `xorm:"gitea varchar(100)" json:"gitea"`
	Heroku          string `xorm:"heroku varchar(100)" json:"heroku"`
	InfluxCloud     string `xorm:"influxcloud varchar(100)" json:"influxcloud"`
	Instagram       string `xorm:"instagram varchar(100)" json:"instagram"`
	Intercom        string `xorm:"intercom varchar(100)" json:"intercom"`
	Kakao           string `xorm:"kakao varchar(100)" json:"kakao"`
	Lastfm          string `xorm:"lastfm varchar(100)" json:"lastfm"`
	Mailru          string `xorm:"mailru varchar(100)" json:"mailru"`
	Meetup          string `xorm:"meetup varchar(100)" json:"meetup"`
	MicrosoftOnline string `xorm:"microsoftonline varchar(100)" json:"microsoftonline"`
	Naver           string `xorm:"naver varchar(100)" json:"naver"`
	Nextcloud       string `xorm:"nextcloud varchar(100)" json:"nextcloud"`
	OneDrive        string `xorm:"onedrive varchar(100)" json:"onedrive"`
	Oura            string `xorm:"oura varchar(100)" json:"oura"`
	Patreon         string `xorm:"patreon varchar(100)" json:"patreon"`
	Paypal          string `xorm:"paypal varchar(100)" json:"paypal"`
	SalesForce      string `xorm:"salesforce varchar(100)" json:"salesforce"`
	Shopify         string `xorm:"shopify varchar(100)" json:"shopify"`
	Soundcloud      string `xorm:"soundcloud varchar(100)" json:"soundcloud"`
	Spotify         string `xorm:"spotify varchar(100)" json:"spotify"`
	Strava          string `xorm:"strava varchar(100)" json:"strava"`
	Stripe          string `xorm:"stripe varchar(100)" json:"stripe"`
	TikTok          string `xorm:"tiktok varchar(100)" json:"tiktok"`
	Tumblr          string `xorm:"tumblr varchar(100)" json:"tumblr"`
	Twitch          string `xorm:"twitch varchar(100)" json:"twitch"`
	Twitter         string `xorm:"twitter varchar(100)" json:"twitter"`
	Typetalk        string `xorm:"typetalk varchar(100)" json:"typetalk"`
	Uber            string `xorm:"uber varchar(100)" json:"uber"`
	VK              string `xorm:"vk varchar(100)" json:"vk"`
	Wepay           string `xorm:"wepay varchar(100)" json:"wepay"`
	Xero            string `xorm:"xero varchar(100)" json:"xero"`
	Yahoo           string `xorm:"yahoo varchar(100)" json:"yahoo"`
	Yammer          string `xorm:"yammer varchar(100)" json:"yammer"`
	Yandex          string `xorm:"yandex varchar(100)" json:"yandex"`
	Zoom            string `xorm:"zoom varchar(100)" json:"zoom"`
	Custom          string `xorm:"custom varchar(100)" json:"custom"`

	WebauthnCredentials []webauthn.Credential `xorm:"webauthnCredentials blob" json:"webauthnCredentials"`
	PreferredMfaType    string                `xorm:"varchar(100)" json:"preferredMfaType"`
	RecoveryCodes       []string              `xorm:"varchar(1000)" json:"recoveryCodes"`
	TotpSecret          string                `xorm:"varchar(100)" json:"totpSecret"`
	MfaPhoneEnabled     bool                  `json:"mfaPhoneEnabled"`
	MfaEmailEnabled     bool                  `json:"mfaEmailEnabled"`
	MultiFactorAuths    []*MfaProps           `xorm:"-" json:"multiFactorAuths,omitempty"`

	Ldap       string            `xorm:"ldap varchar(100)" json:"ldap"`
	Properties map[string]string `json:"properties"`

	Roles       []*Role       `json:"roles"`
	Permissions []*Permission `json:"permissions"`
	Groups      []string      `xorm:"groups varchar(1000)" json:"groups"`

	LastSigninWrongTime string `xorm:"varchar(100)" json:"lastSigninWrongTime"`
	SigninWrongTimes    int    `json:"signinWrongTimes"`

	ManagedAccounts []ManagedAccount `xorm:"managedAccounts blob" json:"managedAccounts"`
}

type Userinfo struct {
	Sub         string   `json:"sub"`
	Iss         string   `json:"iss"`
	Aud         string   `json:"aud"`
	Name        string   `json:"preferred_username,omitempty"`
	DisplayName string   `json:"name,omitempty"`
	Email       string   `json:"email,omitempty"`
	Avatar      string   `json:"picture,omitempty"`
	Address     string   `json:"address,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	Groups      []string `json:"groups,omitempty"`
}

type ManagedAccount struct {
	Application string `xorm:"varchar(100)" json:"application"`
	Username    string `xorm:"varchar(100)" json:"username"`
	Password    string `xorm:"varchar(100)" json:"password"`
	SigninUrl   string `xorm:"varchar(200)" json:"signinUrl"`
}

func (u *User) setPasswordChangeRequirement() {
	if u.passwordChangingAllowed() {
		u.PasswordChangeRequired = true
	}
}

func (u *User) validateUnsupportedPasswordChange() error {
	if !u.passwordChangingAllowed() && u.PasswordChangeRequired {
		return fmt.Errorf("PasswordChangeRequired not allowed for user '%s' due to be external one", u.Name)
	}
	return nil
}

func (u *User) passwordChangingAllowed() bool {
	return u.Type != "" || u.Ldap == ""
}

func GetGlobalUserCount(field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Count(&User{})
}

func GetGlobalUsers() ([]*User, error) {
	users := []*User{}
	err := adapter.Engine.Desc("created_time").Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetPaginationGlobalUsers(offset, limit int, field, value, sortField, sortOrder string) ([]*User, error) {
	users := []*User{}
	session := GetSessionForUser("", offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserCount(owner, field, value string, groupName string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")

	if groupName != "" {
		return GetGroupUserCount(groupName, field, value)
	}

	return session.Count(&User{})
}

func GetOnlineUserCount(owner string, isOnline int) (int64, error) {
	return adapter.Engine.Where("is_online = ?", isOnline).Count(&User{Owner: owner})
}

func GetUsers(owner string) ([]*User, error) {
	users := []*User{}
	err := adapter.Engine.Desc("created_time").Find(&users, &User{Owner: owner})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUsersByTag(owner string, tag string) ([]*User, error) {
	users := []*User{}
	err := adapter.Engine.Desc("created_time").Find(&users, &User{Owner: owner, Tag: tag})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetSortedUsers(owner string, sorter string, limit int) ([]*User, error) {
	users := []*User{}
	err := adapter.Engine.Desc(sorter).Limit(limit, 0).Find(&users, &User{Owner: owner})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetPaginationUsers(owner string, offset, limit int, field, value, sortField, sortOrder string, groupName string) ([]*User, error) {
	users := []*User{}

	if groupName != "" {
		return GetPaginationGroupUsers(groupName, offset, limit, field, value, sortField, sortOrder)
	}

	session := GetSessionForUser(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func getUser(owner string, name string) (*User, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	user := User{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func getUserById(owner string, id string) (*User, error) {
	if owner == "" || id == "" {
		return nil, nil
	}

	user := User{Owner: owner, Id: id}
	existed, err := adapter.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func getUserByWechatId(owner string, wechatOpenId string, wechatUnionId string) (*User, error) {
	if wechatUnionId == "" {
		wechatUnionId = wechatOpenId
	}
	user := &User{}
	existed, err := adapter.Engine.Where("owner = ?", owner).Where("wechat = ? OR wechat = ?", wechatOpenId, wechatUnionId).Get(user)
	if err != nil {
		return nil, err
	}

	if existed {
		return user, nil
	} else {
		return nil, nil
	}
}

func GetUserByEmail(owner string, email string) (*User, error) {
	if owner == "" || email == "" {
		return nil, nil
	}

	user := User{Owner: owner, Email: email}
	existed, err := adapter.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func GetUserByPhone(owner string, phone string) (*User, error) {
	if owner == "" || phone == "" {
		return nil, nil
	}

	user := User{Owner: owner, Phone: phone}
	existed, err := adapter.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func GetUserByUserId(owner string, userId string) (*User, error) {
	if owner == "" || userId == "" {
		return nil, nil
	}

	user := User{Owner: owner, Id: userId}
	existed, err := adapter.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func GetUserByAccessKey(accessKey string) (*User, error) {
	if accessKey == "" {
		return nil, nil
	}
	user := User{AccessKey: accessKey}
	existed, err := adapter.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func GetUser(id string) (*User, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getUser(owner, name)
}

func GetUserNoCheck(id string) (*User, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getUser(owner, name)
}

func GetMaskedUser(user *User, errs ...error) (*User, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	if user == nil {
		return nil, nil
	}

	if user.Password != "" {
		user.Password = "***"
	}
	if user.AccessSecret != "" {
		user.AccessSecret = "***"
	}
	if user.ManagedAccounts != nil {
		for _, manageAccount := range user.ManagedAccounts {
			manageAccount.Password = "***"
		}
	}

	if user.TotpSecret != "" {
		user.TotpSecret = ""
	}
	if user.RecoveryCodes != nil {
		user.RecoveryCodes = nil
	}

	return user, nil
}

func GetMaskedUsers(users []*User, errs ...error) ([]*User, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	var err error
	for _, user := range users {
		user, err = GetMaskedUser(user)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}

func GetLastUser(owner string) (*User, error) {
	user := User{Owner: owner}
	existed, err := adapter.Engine.Desc("created_time", "id").Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	}

	return nil, nil
}

func UpdateUser(id string, user *User, columns []string, isAdmin bool) (bool, error) {
	var err error
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	oldUser, err := getUser(owner, name)
	if err != nil {
		return false, err
	}
	if oldUser == nil {
		return false, nil
	}

	if name != user.Name {
		err := userChangeTrigger(name, user.Name)
		if err != nil {
			return false, err
		}
	}

	if user.Password == "***" {
		user.Password = oldUser.Password
	}

	if user.Avatar != oldUser.Avatar && user.Avatar != "" && user.PermanentAvatar != "*" {
		user.PermanentAvatar, err = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar, false)
		if err != nil {
			return false, err
		}
	}

	if len(columns) == 0 {
		columns = []string{
			"owner", "display_name", "avatar",
			"location", "address", "country_code", "region", "language", "affiliation", "title", "homepage", "bio", "tag", "language", "gender", "birthday", "education", "score", "karma", "ranking", "signup_application",
			"is_admin", "is_global_admin", "is_forbidden", "is_deleted", "password_change_required", "hash", "is_default_avatar", "properties", "webauthnCredentials", "managedAccounts",
			"signin_wrong_times", "last_signin_wrong_time", "groups", "access_key", "access_secret",
			"github", "google", "qq", "wechat", "facebook", "dingtalk", "weibo", "gitee", "linkedin", "wecom", "lark", "gitlab", "adfs",
			"baidu", "alipay", "casdoor", "infoflow", "apple", "azuread", "slack", "steam", "bilibili", "okta", "douyin", "line", "amazon",
			"auth0", "battlenet", "bitbucket", "box", "cloudfoundry", "dailymotion", "deezer", "digitalocean", "discord", "dropbox",
			"eveonline", "fitbit", "gitea", "heroku", "influxcloud", "instagram", "intercom", "kakao", "lastfm", "mailru", "meetup",
			"microsoftonline", "naver", "nextcloud", "onedrive", "oura", "patreon", "paypal", "salesforce", "shopify", "soundcloud",
			"spotify", "strava", "stripe", "tiktok", "tumblr", "twitch", "twitter", "typetalk", "uber", "vk", "wepay", "xero", "yahoo",
			"yammer", "yandex", "zoom", "custom",
		}
	}
	if isAdmin {
		columns = append(columns, "name", "email", "phone", "country_code")
	}

	affected, err := updateUser(id, user, columns)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func updateUser(id string, user *User, columns []string) (int64, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	err := user.UpdateUserHash()
	if err != nil {
		return 0, err
	}

	err = user.validateUnsupportedPasswordChange()
	if err != nil {
		return 0, err
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).Cols(columns...).Update(user)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func UpdateUserForAllFields(id string, user *User) (bool, error) {
	var err error
	owner, name := util.GetOwnerAndNameFromId(id)
	oldUser, err := getUser(owner, name)
	if err != nil {
		return false, err
	}

	if oldUser == nil {
		return false, nil
	}

	if name != user.Name {
		err := userChangeTrigger(name, user.Name)
		if err != nil {
			return false, nil
		}
	}

	err = user.UpdateUserHash()
	if err != nil {
		return false, err
	}

	if user.Avatar != oldUser.Avatar && user.Avatar != "" {
		user.PermanentAvatar, err = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar, false)
		if err != nil {
			return false, err
		}
	}
	err = user.validateUnsupportedPasswordChange()
	if err != nil {
		return false, err
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddUser(user *User) (bool, error) {
	var err error
	if user.Id == "" {
		user.Id = util.GenerateId()
	}

	if user.Owner == "" || user.Name == "" {
		return false, nil
	}

	organization, _ := GetOrganizationByUser(user)
	if organization == nil {
		return false, nil
	}

	if user.PasswordType == "" && organization.PasswordType != "" {
		user.PasswordType = organization.PasswordType
	}

	user.UpdateUserPassword(organization)

	err = user.UpdateUserHash()
	if err != nil {
		return false, err
	}

	user.PreHash = user.Hash

	updated, err := user.refreshAvatar()
	if err != nil {
		return false, err
	}

	if updated && user.PermanentAvatar != "*" {
		user.PermanentAvatar, err = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar, false)
		if err != nil {
			return false, err
		}
	}

	count, err := GetUserCount(user.Owner, "", "", "")
	if err != nil {
		return false, err
	}
	user.Ranking = int(count + 1)
	user.setPasswordChangeRequirement()

	affected, err := adapter.Engine.Insert(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddUsers(users []*User) (bool, error) {
	var err error
	if len(users) == 0 {
		return false, nil
	}

	// organization := GetOrganizationByUser(users[0])
	for _, user := range users {
		// this function is only used for syncer or batch upload, so no need to encrypt the password
		// user.UpdateUserPassword(organization)

		err = user.UpdateUserHash()
		if err != nil {
			return false, err
		}

		user.PreHash = user.Hash

		user.PermanentAvatar, err = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar, true)
		if err != nil {
			return false, err
		}
		user.setPasswordChangeRequirement()
	}

	affected, err := adapter.Engine.Insert(users)
	if err != nil {
		if !strings.Contains(err.Error(), "Duplicate entry") {
			return false, err
		}
	}

	return affected != 0, nil
}

func AddUsersInBatch(users []*User) (bool, error) {
	batchSize := conf.GetConfigBatchSize()

	if len(users) == 0 {
		return false, nil
	}

	affected := false
	for i := 0; i < (len(users)-1)/batchSize+1; i++ {
		start := i * batchSize
		end := (i + 1) * batchSize
		if end > len(users) {
			end = len(users)
		}

		tmp := users[start:end]
		// TODO: save to log instead of standard output
		// fmt.Printf("Add users: [%d - %d].\n", start, end)
		if ok, err := AddUsers(tmp); err != nil {
			return false, err
		} else if ok {
			affected = true
		}
	}

	return affected, nil
}

func DeleteUser(user *User) (bool, error) {
	// Forced offline the user first
	_, err := DeleteSession(util.GetSessionId(user.Owner, user.Name, CasdoorApplication))
	if err != nil {
		return false, err
	}

	affected, err := adapter.Engine.ID(core.PK{user.Owner, user.Name}).Delete(&User{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func GetUserInfo(user *User, scope string, aud string, host string) *Userinfo {
	_, originBackend := getOriginFromHost(host)

	resp := Userinfo{
		Sub: user.Id,
		Iss: originBackend,
		Aud: aud,
	}
	if strings.Contains(scope, "profile") {
		resp.Name = user.Name
		resp.DisplayName = user.DisplayName
		resp.Avatar = user.Avatar
		resp.Groups = user.Groups
	}
	if strings.Contains(scope, "email") {
		resp.Email = user.Email
	}
	if strings.Contains(scope, "address") {
		resp.Address = user.Location
	}
	if strings.Contains(scope, "phone") {
		resp.Phone = user.Phone
	}
	return &resp
}

func LinkUserAccount(user *User, field string, value string) (bool, error) {
	return SetUserField(user, field, value)
}

func (user *User) GetId() string {
	return fmt.Sprintf("%s/%s", user.Owner, user.Name)
}

func isUserIdGlobalAdmin(userId string) bool {
	return strings.HasPrefix(userId, "built-in/")
}

func ExtendUserWithRolesAndPermissions(user *User) (err error) {
	if user == nil {
		return
	}

	user.Permissions, user.Roles, err = GetPermissionsAndRolesByUser(user.GetId())
	if err != nil {
		return err
	}

	if user.Groups == nil {
		user.Groups = []string{}
	}

	return
}

func userChangeTrigger(oldName string, newName string) error {
	session := adapter.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	var roles []*Role
	err = adapter.Engine.Find(&roles)
	if err != nil {
		return err
	}

	for _, role := range roles {
		for j, u := range role.Users {
			// u = organization/username
			owner, name := util.GetOwnerAndNameFromId(u)
			if name == oldName {
				role.Users[j] = util.GetId(owner, newName)
			}
		}
		_, err = session.Where("name=?", role.Name).And("owner=?", role.Owner).Update(role)
		if err != nil {
			return err
		}
	}

	var permissions []*Permission
	err = adapter.Engine.Find(&permissions)
	if err != nil {
		return err
	}
	for _, permission := range permissions {
		for j, u := range permission.Users {
			// u = organization/username
			owner, name := util.GetOwnerAndNameFromId(u)
			if name == oldName {
				permission.Users[j] = util.GetId(owner, newName)
			}
		}
		_, err = session.Where("name=?", permission.Name).And("owner=?", permission.Owner).Update(permission)
		if err != nil {
			return err
		}
	}

	resource := new(Resource)
	resource.User = newName
	_, err = session.Where("user=?", oldName).Update(resource)
	if err != nil {
		return err
	}

	return session.Commit()
}

func (user *User) IsMfaEnabled() bool {
	if user == nil {
		return false
	}
	return user.PreferredMfaType != ""
}

func (user *User) GetPreferredMfaProps(masked bool) *MfaProps {
	if user == nil || user.PreferredMfaType == "" {
		return nil
	}
	return user.GetMfaProps(user.PreferredMfaType, masked)
}

func AddUserkeys(user *User, isAdmin bool) (bool, error) {
	if user == nil {
		return false, nil
	}

	user.AccessKey = util.GenerateId()
	user.AccessSecret = util.GenerateId()

	return UpdateUser(user.GetId(), user, []string{}, isAdmin)
}
