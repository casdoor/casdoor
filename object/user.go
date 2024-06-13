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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/xorm-io/builder"
	"github.com/xorm-io/core"
)

const (
	UserPropertiesWechatUnionId = "wechatUnionId"
	UserPropertiesWechatOpenId  = "wechatOpenId"
)

const UserEnforcerId = "built-in/user-enforcer-built-in"

var userEnforcer *UserGroupEnforcer

func InitUserManager() {
	enforcer, err := GetInitializedEnforcer(UserEnforcerId)
	if err != nil {
		panic(err)
	}

	userEnforcer = NewUserGroupEnforcer(enforcer.Enforcer)
}

type User struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100) index" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DeletedTime string `xorm:"varchar(100)" json:"deletedTime"`

	Id                string   `xorm:"varchar(100) index" json:"id"`
	ExternalId        string   `xorm:"varchar(100) index" json:"externalId"`
	Type              string   `xorm:"varchar(100)" json:"type"`
	Password          string   `xorm:"varchar(150)" json:"password"`
	PasswordSalt      string   `xorm:"varchar(100)" json:"passwordSalt"`
	PasswordType      string   `xorm:"varchar(100)" json:"passwordType"`
	DisplayName       string   `xorm:"varchar(100)" json:"displayName"`
	FirstName         string   `xorm:"varchar(100)" json:"firstName"`
	LastName          string   `xorm:"varchar(100)" json:"lastName"`
	Avatar            string   `xorm:"varchar(500)" json:"avatar"`
	AvatarType        string   `xorm:"varchar(100)" json:"avatarType"`
	PermanentAvatar   string   `xorm:"varchar(500)" json:"permanentAvatar"`
	Email             string   `xorm:"varchar(100) index" json:"email"`
	EmailVerified     bool     `json:"emailVerified"`
	Phone             string   `xorm:"varchar(100) index" json:"phone"`
	CountryCode       string   `xorm:"varchar(6)" json:"countryCode"`
	Region            string   `xorm:"varchar(100)" json:"region"`
	Location          string   `xorm:"varchar(100)" json:"location"`
	Address           []string `json:"address"`
	Affiliation       string   `xorm:"varchar(100)" json:"affiliation"`
	Title             string   `xorm:"varchar(100)" json:"title"`
	IdCardType        string   `xorm:"varchar(100)" json:"idCardType"`
	IdCard            string   `xorm:"varchar(100) index" json:"idCard"`
	Homepage          string   `xorm:"varchar(100)" json:"homepage"`
	Bio               string   `xorm:"varchar(100)" json:"bio"`
	Tag               string   `xorm:"varchar(100)" json:"tag"`
	Language          string   `xorm:"varchar(100)" json:"language"`
	Gender            string   `xorm:"varchar(100)" json:"gender"`
	Birthday          string   `xorm:"varchar(100)" json:"birthday"`
	Education         string   `xorm:"varchar(100)" json:"education"`
	Score             int      `json:"score"`
	Karma             int      `json:"karma"`
	Ranking           int      `json:"ranking"`
	Balance           float64  `json:"balance"`
	Currency          string   `xorm:"varchar(100)" json:"currency"`
	IsDefaultAvatar   bool     `json:"isDefaultAvatar"`
	IsOnline          bool     `json:"isOnline"`
	IsAdmin           bool     `json:"isAdmin"`
	IsForbidden       bool     `json:"isForbidden"`
	IsDeleted         bool     `json:"isDeleted"`
	SignupApplication string   `xorm:"varchar(100)" json:"signupApplication"`
	Hash              string   `xorm:"varchar(100)" json:"hash"`
	PreHash           string   `xorm:"varchar(100)" json:"preHash"`
	AccessKey         string   `xorm:"varchar(100)" json:"accessKey"`
	AccessSecret      string   `xorm:"varchar(100)" json:"accessSecret"`
	AccessToken       string   `xorm:"mediumtext" json:"accessToken"`

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
	AzureADB2c      string `xorm:"azureadb2c varchar(100)" json:"azureadb2c"`
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
	MetaMask        string `xorm:"metamask varchar(100)" json:"metamask"`
	Web3Onboard     string `xorm:"web3onboard varchar(100)" json:"web3onboard"`
	Custom          string `xorm:"custom varchar(100)" json:"custom"`

	WebauthnCredentials []webauthn.Credential `xorm:"webauthnCredentials blob" json:"webauthnCredentials"`
	PreferredMfaType    string                `xorm:"varchar(100)" json:"preferredMfaType"`
	RecoveryCodes       []string              `xorm:"varchar(1000)" json:"recoveryCodes"`
	TotpSecret          string                `xorm:"varchar(100)" json:"totpSecret"`
	MfaPhoneEnabled     bool                  `json:"mfaPhoneEnabled"`
	MfaEmailEnabled     bool                  `json:"mfaEmailEnabled"`
	MultiFactorAuths    []*MfaProps           `xorm:"-" json:"multiFactorAuths,omitempty"`
	Invitation          string                `xorm:"varchar(100) index" json:"invitation"`
	InvitationCode      string                `xorm:"varchar(100) index" json:"invitationCode"`
	FaceIds             []*FaceId             `json:"faceIds"`

	Ldap       string            `xorm:"ldap varchar(100)" json:"ldap"`
	Properties map[string]string `json:"properties"`

	Roles       []*Role       `json:"roles"`
	Permissions []*Permission `json:"permissions"`
	Groups      []string      `xorm:"groups varchar(1000)" json:"groups"`

	LastSigninWrongTime string `xorm:"varchar(100)" json:"lastSigninWrongTime"`
	SigninWrongTimes    int    `json:"signinWrongTimes"`

	ManagedAccounts    []ManagedAccount `xorm:"managedAccounts blob" json:"managedAccounts"`
	NeedUpdatePassword bool             `json:"needUpdatePassword"`
}

type Userinfo struct {
	Sub           string   `json:"sub"`
	Iss           string   `json:"iss"`
	Aud           string   `json:"aud"`
	Name          string   `json:"preferred_username,omitempty"`
	DisplayName   string   `json:"name,omitempty"`
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty"`
	Avatar        string   `json:"picture,omitempty"`
	Address       string   `json:"address,omitempty"`
	Phone         string   `json:"phone,omitempty"`
	Groups        []string `json:"groups,omitempty"`
	Roles         []string `json:"roles,omitempty"`
	Permissions   []string `json:"permissions,omitempty"`
}

type ManagedAccount struct {
	Application string `xorm:"varchar(100)" json:"application"`
	Username    string `xorm:"varchar(100)" json:"username"`
	Password    string `xorm:"varchar(100)" json:"password"`
	SigninUrl   string `xorm:"varchar(200)" json:"signinUrl"`
}

type FaceId struct {
	Name       string    `xorm:"varchar(100) notnull pk" json:"name"`
	FaceIdData []float64 `json:"faceIdData"`
}

func GetUserFieldStringValue(user *User, fieldName string) (bool, string, error) {
	val := reflect.ValueOf(*user)
	fieldValue := val.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		return false, "", nil
	}

	if fieldValue.Kind() == reflect.String {
		return true, fieldValue.String(), nil
	}

	marshalValue, err := json.Marshal(fieldValue.Interface())
	if err != nil {
		return false, "", err
	}

	return true, string(marshalValue), nil
}

func GetGlobalUserCount(field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Count(&User{})
}

func GetGlobalUsers() ([]*User, error) {
	users := []*User{}
	err := ormer.Engine.Desc("created_time").Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetGlobalUsersWithFilter(cond builder.Cond) ([]*User, error) {
	users := []*User{}
	session := ormer.Engine.Desc("created_time")
	if cond != nil {
		session = session.Where(cond)
	}
	err := session.Find(&users)
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
		return GetGroupUserCount(util.GetId(owner, groupName), field, value)
	}

	return session.Count(&User{})
}

func GetOnlineUserCount(owner string, isOnline int) (int64, error) {
	return ormer.Engine.Where("is_online = ?", isOnline).Count(&User{Owner: owner})
}

func GetUsers(owner string) ([]*User, error) {
	users := []*User{}
	err := ormer.Engine.Desc("created_time").Find(&users, &User{Owner: owner})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUsersWithFilter(owner string, cond builder.Cond) ([]*User, error) {
	users := []*User{}
	session := ormer.Engine.Desc("created_time")
	if cond != nil {
		session = session.Where(cond)
	}
	err := session.Find(&users, &User{Owner: owner})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUsersByTagWithFilter(owner string, tag string, cond builder.Cond) ([]*User, error) {
	users := []*User{}
	session := ormer.Engine.Desc("created_time")
	if cond != nil {
		session = session.Where(cond)
	}
	err := session.Find(&users, &User{Owner: owner, Tag: tag})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetSortedUsers(owner string, sorter string, limit int) ([]*User, error) {
	users := []*User{}
	err := ormer.Engine.Desc(sorter).Limit(limit, 0).Find(&users, &User{Owner: owner})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetPaginationUsers(owner string, offset, limit int, field, value, sortField, sortOrder string, groupName string) ([]*User, error) {
	users := []*User{}

	if groupName != "" {
		return GetPaginationGroupUsers(util.GetId(owner, groupName), offset, limit, field, value, sortField, sortOrder)
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
	existed, err := ormer.Engine.Get(&user)
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
	existed, err := ormer.Engine.Get(&user)
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
	existed, err := ormer.Engine.Where("owner = ?", owner).Where("wechat = ? OR wechat = ?", wechatOpenId, wechatUnionId).Get(user)
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
	existed, err := ormer.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func GetUserByEmailOnly(email string) (*User, error) {
	if email == "" {
		return nil, nil
	}

	user := User{Email: email}
	existed, err := ormer.Engine.Get(&user)
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
	existed, err := ormer.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func GetUserByPhoneOnly(phone string) (*User, error) {
	if phone == "" {
		return nil, nil
	}

	user := User{Phone: phone}
	existed, err := ormer.Engine.Get(&user)
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
	existed, err := ormer.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func GetUserByUserIdOnly(userId string) (*User, error) {
	if userId == "" {
		return nil, nil
	}

	user := User{Id: userId}
	existed, err := ormer.Engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func GetUserByInvitationCode(owner string, invitationCode string) (*User, error) {
	if owner == "" || invitationCode == "" {
		return nil, nil
	}

	user := User{Owner: owner, InvitationCode: invitationCode}
	existed, err := ormer.Engine.Get(&user)
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
	existed, err := ormer.Engine.Get(&user)
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

func GetMaskedUser(user *User, isAdminOrSelf bool, errs ...error) (*User, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	if user == nil {
		return nil, nil
	}

	if user.Password != "" {
		user.Password = "***"
	}

	if !isAdminOrSelf {
		if user.AccessSecret != "" {
			user.AccessSecret = "***"
		}
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
		user, err = GetMaskedUser(user, false)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}

func getLastUser(owner string) (*User, error) {
	user := User{Owner: owner}
	existed, err := ormer.Engine.Desc("created_time", "id").Get(&user)
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
		return false, fmt.Errorf("the user: %s is not found", id)
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
			"owner", "display_name", "avatar", "first_name", "last_name",
			"location", "address", "country_code", "region", "language", "affiliation", "title", "id_card_type", "id_card", "homepage", "bio", "tag", "language", "gender", "birthday", "education", "score", "karma", "ranking", "signup_application",
			"is_admin", "is_forbidden", "is_deleted", "hash", "is_default_avatar", "properties", "webauthnCredentials", "managedAccounts", "face_ids",
			"signin_wrong_times", "last_signin_wrong_time", "groups", "access_key", "access_secret", "mfa_phone_enabled", "mfa_email_enabled",
			"github", "google", "qq", "wechat", "facebook", "dingtalk", "weibo", "gitee", "linkedin", "wecom", "lark", "gitlab", "adfs",
			"baidu", "alipay", "casdoor", "infoflow", "apple", "azuread", "azureadb2c", "slack", "steam", "bilibili", "okta", "douyin", "line", "amazon",
			"auth0", "battlenet", "bitbucket", "box", "cloudfoundry", "dailymotion", "deezer", "digitalocean", "discord", "dropbox",
			"eveonline", "fitbit", "gitea", "heroku", "influxcloud", "instagram", "intercom", "kakao", "lastfm", "mailru", "meetup",
			"microsoftonline", "naver", "nextcloud", "onedrive", "oura", "patreon", "paypal", "salesforce", "shopify", "soundcloud",
			"spotify", "strava", "stripe", "type", "tiktok", "tumblr", "twitch", "twitter", "typetalk", "uber", "vk", "wepay", "xero", "yahoo",
			"yammer", "yandex", "zoom", "custom", "need_update_password",
		}
	}
	if isAdmin {
		columns = append(columns, "name", "id", "email", "phone", "country_code", "type")
	}

	columns = append(columns, "updated_time")
	user.UpdatedTime = util.GetCurrentTime()

	if len(user.DeletedTime) > 0 {
		columns = append(columns, "deleted_time")
	}

	if util.ContainsString(columns, "groups") {
		_, err := userEnforcer.UpdateGroupsForUser(user.GetId(), user.Groups)
		if err != nil {
			return false, err
		}
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

	affected, err := ormer.Engine.ID(core.PK{owner, name}).Cols(columns...).Update(user)
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
		return false, fmt.Errorf("the user: %s is not found", id)
	}

	if name != user.Name {
		err := userChangeTrigger(name, user.Name)
		if err != nil {
			return false, err
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

	user.UpdatedTime = util.GetCurrentTime()

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddUser(user *User) (bool, error) {
	if user.Id == "" {
		application, err := GetApplicationByUser(user)
		if err != nil {
			return false, err
		}

		id, err := GenerateIdForNewUser(application)
		if err != nil {
			return false, err
		}

		user.Id = id
	}

	if user.Owner == "" || user.Name == "" {
		return false, fmt.Errorf("the user's owner and name should not be empty")
	}

	organization, err := GetOrganizationByUser(user)
	if err != nil {
		return false, err
	}
	if organization == nil {
		return false, fmt.Errorf("the organization: %s is not found", user.Owner)
	}

	if organization.DefaultPassword != "" && user.Password == "123" {
		user.Password = organization.DefaultPassword
	}

	if user.PasswordType == "" || user.PasswordType == "plain" {
		user.UpdateUserPassword(organization)
	}

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

	if user.Groups != nil && len(user.Groups) > 0 {
		_, err = userEnforcer.UpdateGroupsForUser(user.GetId(), user.Groups)
		if err != nil {
			return false, err
		}
	}

	isUsernameLowered := conf.GetConfigBool("isUsernameLowered")
	if isUsernameLowered {
		user.Name = strings.ToLower(user.Name)
	}

	affected, err := ormer.Engine.Insert(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddUsers(users []*User) (bool, error) {
	if len(users) == 0 {
		return false, fmt.Errorf("no users are provided")
	}

	isUsernameLowered := conf.GetConfigBool("isUsernameLowered")

	// organization := GetOrganizationByUser(users[0])
	for _, user := range users {
		// this function is only used for syncer or batch upload, so no need to encrypt the password
		// user.UpdateUserPassword(organization)

		err := user.UpdateUserHash()
		if err != nil {
			return false, err
		}

		user.PreHash = user.Hash

		user.PermanentAvatar, err = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar, true)
		if err != nil {
			return false, err
		}

		if user.Groups != nil && len(user.Groups) > 0 {
			_, err = userEnforcer.UpdateGroupsForUser(user.GetId(), user.Groups)
			if err != nil {
				return false, err
			}
		}

		user.Name = strings.TrimSpace(user.Name)
		if isUsernameLowered {
			user.Name = strings.ToLower(user.Name)
		}
	}

	affected, err := ormer.Engine.Insert(users)
	if err != nil {
		if !strings.Contains(err.Error(), "Duplicate entry") {
			return false, err
		}
	}

	return affected != 0, nil
}

func AddUsersInBatch(users []*User) (bool, error) {
	if len(users) == 0 {
		return false, fmt.Errorf("no users are provided")
	}

	batchSize := conf.GetConfigBatchSize()

	affected := false
	for i := 0; i < len(users); i += batchSize {
		start := i
		end := i + batchSize
		if end > len(users) {
			end = len(users)
		}

		tmp := users[start:end]
		fmt.Printf("The syncer adds users: [%d - %d]\n", start, end)
		if ok, err := AddUsers(tmp); err != nil {
			return false, err
		} else if ok {
			affected = true
		}
	}

	return affected, nil
}

func deleteUser(user *User) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{user.Owner, user.Name}).Delete(&User{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteUser(user *User) (bool, error) {
	// Forced offline the user first
	_, err := DeleteSession(util.GetSessionId(user.Owner, user.Name, CasdoorApplication))
	if err != nil {
		return false, err
	}

	return deleteUser(user)
}

func GetUserInfo(user *User, scope string, aud string, host string) (*Userinfo, error) {
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

		err := ExtendUserWithRolesAndPermissions(user)
		if err != nil {
			return nil, err
		}

		resp.Roles = []string{}
		for _, role := range user.Roles {
			resp.Roles = append(resp.Roles, role.Name)
		}

		resp.Permissions = []string{}
		for _, permission := range user.Permissions {
			resp.Permissions = append(resp.Permissions, permission.Name)
		}
	}

	if strings.Contains(scope, "email") {
		resp.Email = user.Email
		// resp.EmailVerified = user.EmailVerified
		resp.EmailVerified = true
	}

	if strings.Contains(scope, "address") {
		resp.Address = user.Location
	}

	if strings.Contains(scope, "phone") {
		resp.Phone = user.Phone
	}

	return &resp, nil
}

func LinkUserAccount(user *User, field string, value string) (bool, error) {
	return SetUserField(user, field, value)
}

func (user *User) GetId() string {
	return fmt.Sprintf("%s/%s", user.Owner, user.Name)
}

func (user *User) GetFriendlyName() string {
	if user.FirstName != "" && user.LastName != "" {
		return fmt.Sprintf("%s, %s", user.FirstName, user.LastName)
	} else if user.DisplayName != "" {
		return user.DisplayName
	} else if user.Name != "" {
		return user.Name
	} else {
		return user.Id
	}
}

func isUserIdGlobalAdmin(userId string) bool {
	return strings.HasPrefix(userId, "built-in/") || IsAppUser(userId)
}

func ExtendUserWithRolesAndPermissions(user *User) (err error) {
	if user == nil {
		return
	}

	user.Permissions, user.Roles, err = getPermissionsAndRolesByUser(user.GetId())
	if err != nil {
		return err
	}

	if user.Groups == nil {
		user.Groups = []string{}
	}

	return
}

func DeleteGroupForUser(user string, group string) (bool, error) {
	return userEnforcer.DeleteGroupForUser(user, group)
}

func userChangeTrigger(oldName string, newName string) error {
	session := ormer.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	var roles []*Role
	err = ormer.Engine.Find(&roles)
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
	err = ormer.Engine.Find(&permissions)
	if err != nil {
		return err
	}
	for _, permission := range permissions {
		for j, u := range permission.Users {
			if u == "*" {
				continue
			}

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

func AddUserKeys(user *User, isAdmin bool) (bool, error) {
	if user == nil {
		return false, fmt.Errorf("the user is not found")
	}

	user.AccessKey = util.GenerateId()
	user.AccessSecret = util.GenerateId()

	return UpdateUser(user.GetId(), user, []string{}, isAdmin)
}

func (user *User) IsApplicationAdmin(application *Application) bool {
	if user == nil {
		return false
	}

	return (user.Owner == application.Organization && user.IsAdmin) || user.IsGlobalAdmin()
}

func (user *User) IsGlobalAdmin() bool {
	if user == nil {
		return false
	}

	return user.Owner == "built-in"
}

func GenerateIdForNewUser(application *Application) (string, error) {
	if application == nil || application.GetSignupItemRule("ID") != "Incremental" {
		return util.GenerateId(), nil
	}

	lastUser, err := getLastUser(application.Organization)
	if err != nil {
		return "", err
	}

	lastUserId := -1
	if lastUser != nil {
		lastUserId, err = util.ParseIntWithError(lastUser.Id)
		if err != nil {
			return util.GenerateId(), nil
		}
	}

	res := strconv.Itoa(lastUserId + 1)
	return res, nil
}
