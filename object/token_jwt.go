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
	"reflect"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	*User
	TokenType string `json:"tokenType,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
	Tag       string `json:"tag"`
	Scope     string `json:"scope,omitempty"`
	// the `azp` (Authorized Party) claim. Optional. See https://openid.net/specs/openid-connect-core-1_0.html#IDToken
	Azp string `json:"azp,omitempty"`
	jwt.RegisteredClaims
}

type UserShort struct {
	Owner string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name  string `xorm:"varchar(100) notnull pk" json:"name"`

	Id          string `xorm:"varchar(100) index" json:"id"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Avatar      string `xorm:"varchar(500)" json:"avatar"`
	Email       string `xorm:"varchar(100) index" json:"email"`
	Phone       string `xorm:"varchar(100) index" json:"phone"`
}

type UserWithoutThirdIdp struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100) index" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DeletedTime string `xorm:"varchar(100)" json:"deletedTime"`

	Id                string   `xorm:"varchar(100) index" json:"id"`
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

	GitHub   string `xorm:"github varchar(100)" json:"github"`
	Google   string `xorm:"varchar(100)" json:"google"`
	QQ       string `xorm:"qq varchar(100)" json:"qq"`
	WeChat   string `xorm:"wechat varchar(100)" json:"wechat"`
	Facebook string `xorm:"facebook varchar(100)" json:"facebook"`
	DingTalk string `xorm:"dingtalk varchar(100)" json:"dingtalk"`
	Weibo    string `xorm:"weibo varchar(100)" json:"weibo"`
	Gitee    string `xorm:"gitee varchar(100)" json:"gitee"`
	LinkedIn string `xorm:"linkedin varchar(100)" json:"linkedin"`
	Wecom    string `xorm:"wecom varchar(100)" json:"wecom"`
	Lark     string `xorm:"lark varchar(100)" json:"lark"`
	Gitlab   string `xorm:"gitlab varchar(100)" json:"gitlab"`

	CreatedIp      string `xorm:"varchar(100)" json:"createdIp"`
	LastSigninTime string `xorm:"varchar(100)" json:"lastSigninTime"`
	LastSigninIp   string `xorm:"varchar(100)" json:"lastSigninIp"`

	// WebauthnCredentials []webauthn.Credential `xorm:"webauthnCredentials blob" json:"webauthnCredentials"`
	PreferredMfaType string   `xorm:"varchar(100)" json:"preferredMfaType"`
	RecoveryCodes    []string `xorm:"varchar(1000)" json:"recoveryCodes"`
	TotpSecret       string   `xorm:"varchar(100)" json:"totpSecret"`
	MfaPhoneEnabled  bool     `json:"mfaPhoneEnabled"`
	MfaEmailEnabled  bool     `json:"mfaEmailEnabled"`
	// MultiFactorAuths    []*MfaProps           `xorm:"-" json:"multiFactorAuths,omitempty"`

	Ldap       string            `xorm:"ldap varchar(100)" json:"ldap"`
	Properties map[string]string `json:"properties"`

	Roles       []*Role       `json:"roles"`
	Permissions []*Permission `json:"permissions"`
	Groups      []string      `xorm:"groups varchar(1000)" json:"groups"`

	LastSigninWrongTime string `xorm:"varchar(100)" json:"lastSigninWrongTime"`
	SigninWrongTimes    int    `json:"signinWrongTimes"`

	ManagedAccounts []ManagedAccount `xorm:"managedAccounts blob" json:"managedAccounts"`
}

type ClaimsShort struct {
	*UserShort
	TokenType string `json:"tokenType,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
	Scope     string `json:"scope,omitempty"`
	Azp       string `json:"azp,omitempty"`
	jwt.RegisteredClaims
}

type OIDCAddress struct {
	Formatted     string `json:"formatted"`
	StreetAddress string `json:"street_address"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

type ClaimsWithoutThirdIdp struct {
	*UserWithoutThirdIdp
	TokenType string `json:"tokenType,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
	Tag       string `json:"tag"`
	Scope     string `json:"scope,omitempty"`
	Azp       string `json:"azp,omitempty"`
	jwt.RegisteredClaims
}

func getShortUser(user *User) *UserShort {
	res := &UserShort{
		Owner: user.Owner,
		Name:  user.Name,

		Id:          user.Id,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Email:       user.Email,
		Phone:       user.Phone,
	}
	return res
}

func getUserWithoutThirdIdp(user *User) *UserWithoutThirdIdp {
	res := &UserWithoutThirdIdp{
		Owner:       user.Owner,
		Name:        user.Name,
		CreatedTime: user.CreatedTime,
		UpdatedTime: user.UpdatedTime,
		DeletedTime: user.DeletedTime,

		Id:                user.Id,
		Type:              user.Type,
		Password:          user.Password,
		PasswordSalt:      user.PasswordSalt,
		PasswordType:      user.PasswordType,
		DisplayName:       user.DisplayName,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Avatar:            user.Avatar,
		AvatarType:        user.AvatarType,
		PermanentAvatar:   user.PermanentAvatar,
		Email:             user.Email,
		EmailVerified:     user.EmailVerified,
		Phone:             user.Phone,
		CountryCode:       user.CountryCode,
		Region:            user.Region,
		Location:          user.Location,
		Address:           user.Address,
		Affiliation:       user.Affiliation,
		Title:             user.Title,
		IdCardType:        user.IdCardType,
		IdCard:            user.IdCard,
		Homepage:          user.Homepage,
		Bio:               user.Bio,
		Tag:               user.Tag,
		Language:          user.Language,
		Gender:            user.Gender,
		Birthday:          user.Birthday,
		Education:         user.Education,
		Score:             user.Score,
		Karma:             user.Karma,
		Ranking:           user.Ranking,
		IsDefaultAvatar:   user.IsDefaultAvatar,
		IsOnline:          user.IsOnline,
		IsAdmin:           user.IsAdmin,
		IsForbidden:       user.IsForbidden,
		IsDeleted:         user.IsDeleted,
		SignupApplication: user.SignupApplication,
		Hash:              user.Hash,
		PreHash:           user.PreHash,
		AccessKey:         user.AccessKey,
		AccessSecret:      user.AccessSecret,

		GitHub:   user.GitHub,
		Google:   user.Google,
		QQ:       user.QQ,
		WeChat:   user.WeChat,
		Facebook: user.Facebook,
		DingTalk: user.DingTalk,
		Weibo:    user.Weibo,
		Gitee:    user.Gitee,
		LinkedIn: user.LinkedIn,
		Wecom:    user.Wecom,
		Lark:     user.Lark,
		Gitlab:   user.Gitlab,

		CreatedIp:      user.CreatedIp,
		LastSigninTime: user.LastSigninTime,
		LastSigninIp:   user.LastSigninIp,

		PreferredMfaType: user.PreferredMfaType,
		RecoveryCodes:    user.RecoveryCodes,
		TotpSecret:       user.TotpSecret,
		MfaPhoneEnabled:  user.MfaPhoneEnabled,
		MfaEmailEnabled:  user.MfaEmailEnabled,

		Ldap:       user.Ldap,
		Properties: user.Properties,

		Roles:       user.Roles,
		Permissions: user.Permissions,
		Groups:      user.Groups,

		LastSigninWrongTime: user.LastSigninWrongTime,
		SigninWrongTimes:    user.SigninWrongTimes,

		ManagedAccounts: user.ManagedAccounts,
	}

	return res
}

func getShortClaims(claims Claims) ClaimsShort {
	res := ClaimsShort{
		UserShort:        getShortUser(claims.User),
		TokenType:        claims.TokenType,
		Nonce:            claims.Nonce,
		Scope:            claims.Scope,
		RegisteredClaims: claims.RegisteredClaims,
		Azp:              claims.Azp,
	}
	return res
}

func getClaimsWithoutThirdIdp(claims Claims) ClaimsWithoutThirdIdp {
	res := ClaimsWithoutThirdIdp{
		UserWithoutThirdIdp: getUserWithoutThirdIdp(claims.User),
		TokenType:           claims.TokenType,
		Nonce:               claims.Nonce,
		Tag:                 claims.Tag,
		Scope:               claims.Scope,
		RegisteredClaims:    claims.RegisteredClaims,
		Azp:                 claims.Azp,
	}
	return res
}

func getClaimsCustom(claims Claims, tokenField []string) jwt.MapClaims {
	res := make(jwt.MapClaims)

	userValue := reflect.ValueOf(claims.User).Elem()

	res["iss"] = claims.RegisteredClaims.Issuer
	res["sub"] = claims.RegisteredClaims.Subject
	res["aud"] = claims.RegisteredClaims.Audience
	res["exp"] = claims.RegisteredClaims.ExpiresAt
	res["nbf"] = claims.RegisteredClaims.NotBefore
	res["iat"] = claims.RegisteredClaims.IssuedAt
	res["jti"] = claims.RegisteredClaims.ID
	res["tokenType"] = claims.TokenType
	res["nonce"] = claims.Nonce
	res["tag"] = claims.Tag
	res["scope"] = claims.Scope
	res["azp"] = claims.Azp

	for _, field := range tokenField {
		userField := userValue.FieldByName(field)
		if userField.IsValid() {
			newfield := util.SnakeToCamel(util.CamelToSnakeCase(field))
			res[newfield] = userField.Interface()
		}
	}

	return res
}

func refineUser(user *User) *User {
	user.Password = ""

	if user.Address == nil {
		user.Address = []string{}
	}
	if user.Properties == nil {
		user.Properties = map[string]string{}
	}
	if user.Roles == nil {
		user.Roles = []*Role{}
	}
	if user.Permissions == nil {
		user.Permissions = []*Permission{}
	}
	if user.Groups == nil {
		user.Groups = []string{}
	}

	return user
}

func generateJwtToken(application *Application, user *User, nonce string, scope string, host string) (string, string, string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(application.ExpireInHours) * time.Hour)
	refreshExpireTime := nowTime.Add(time.Duration(application.RefreshExpireInHours) * time.Hour)
	if application.RefreshExpireInHours == 0 {
		refreshExpireTime = expireTime
	}

	user = refineUser(user)

	_, originBackend := getOriginFromHost(host)

	name := util.GenerateId()
	jti := util.GetId(application.Owner, name)

	claims := Claims{
		User:      user,
		TokenType: "access-token",
		Nonce:     nonce,
		// FIXME: A workaround for custom claim by reusing `tag` in user info
		Tag:   user.Tag,
		Scope: scope,
		Azp:   application.ClientId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    originBackend,
			Subject:   user.Id,
			Audience:  []string{application.ClientId},
			ExpiresAt: jwt.NewNumericDate(expireTime),
			NotBefore: jwt.NewNumericDate(nowTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			ID:        jti,
		},
	}

	if application.IsShared {
		claims.Audience = []string{application.ClientId + "-org-" + user.Owner}
	}

	var token *jwt.Token
	var refreshToken *jwt.Token

	if application.TokenFormat == "" {
		application.TokenFormat = "JWT"
	}

	var jwtMethod jwt.SigningMethod

	if application.TokenSigningMethod == "RS256" {
		jwtMethod = jwt.SigningMethodRS256
	} else if application.TokenSigningMethod == "RS512" {
		jwtMethod = jwt.SigningMethodRS512
	} else if application.TokenSigningMethod == "ES256" {
		jwtMethod = jwt.SigningMethodES256
	} else if application.TokenSigningMethod == "ES512" {
		jwtMethod = jwt.SigningMethodES512
	} else if application.TokenSigningMethod == "ES384" {
		jwtMethod = jwt.SigningMethodES384
	} else {
		jwtMethod = jwt.SigningMethodRS256
	}

	// the JWT token length in "JWT-Empty" mode will be very short, as User object only has two properties: owner and name
	if application.TokenFormat == "JWT" {
		claimsWithoutThirdIdp := getClaimsWithoutThirdIdp(claims)

		token = jwt.NewWithClaims(jwtMethod, claimsWithoutThirdIdp)
		claimsWithoutThirdIdp.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
		claimsWithoutThirdIdp.TokenType = "refresh-token"
		refreshToken = jwt.NewWithClaims(jwtMethod, claimsWithoutThirdIdp)
	} else if application.TokenFormat == "JWT-Empty" {
		claimsShort := getShortClaims(claims)

		token = jwt.NewWithClaims(jwtMethod, claimsShort)
		claimsShort.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
		claimsShort.TokenType = "refresh-token"
		refreshToken = jwt.NewWithClaims(jwtMethod, claimsShort)
	} else if application.TokenFormat == "JWT-Custom" {
		claimsCustom := getClaimsCustom(claims, application.TokenFields)

		token = jwt.NewWithClaims(jwtMethod, claimsCustom)
		refreshClaims := getClaimsCustom(claims, application.TokenFields)
		refreshClaims["exp"] = jwt.NewNumericDate(refreshExpireTime)
		refreshClaims["TokenType"] = "refresh-token"
		refreshToken = jwt.NewWithClaims(jwtMethod, refreshClaims)
	} else if application.TokenFormat == "JWT-Standard" {
		claimsStandard := getStandardClaims(claims)

		token = jwt.NewWithClaims(jwtMethod, claimsStandard)
		claimsStandard.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
		claimsStandard.TokenType = "refresh-token"
		refreshToken = jwt.NewWithClaims(jwtMethod, claimsStandard)
	} else {
		return "", "", "", fmt.Errorf("unknown application TokenFormat: %s", application.TokenFormat)
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return "", "", "", err
	}

	if cert == nil {
		if application.Cert == "" {
			return "", "", "", fmt.Errorf("The cert field of the application \"%s\" should not be empty", application.GetId())
		} else {
			return "", "", "", fmt.Errorf("The cert \"%s\" does not exist", application.Cert)
		}
	}

	var (
		tokenString        string
		refreshTokenString string
		key                interface{}
	)

	if strings.Contains(application.TokenSigningMethod, "RS") || application.TokenSigningMethod == "" {
		// RSA private key
		key, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(cert.PrivateKey))
	} else if strings.Contains(application.TokenSigningMethod, "ES") {
		// ES private key
		key, err = jwt.ParseECPrivateKeyFromPEM([]byte(cert.PrivateKey))
	} else if strings.Contains(application.TokenSigningMethod, "Ed") {
		// Ed private key
		key, err = jwt.ParseEdPrivateKeyFromPEM([]byte(cert.PrivateKey))
	}
	if err != nil {
		return "", "", "", err
	}

	token.Header["kid"] = cert.Name
	tokenString, err = token.SignedString(key)
	if err != nil {
		return "", "", "", err
	}
	refreshTokenString, err = refreshToken.SignedString(key)

	return tokenString, refreshTokenString, name, err
}

func ParseJwtToken(token string, cert *Cert) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		var (
			certificate interface{}
			err         error
		)

		if cert.Certificate == "" {
			return nil, fmt.Errorf("the certificate field should not be empty for the cert: %v", cert)
		}

		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			// RSA certificate
			certificate, err = jwt.ParseRSAPublicKeyFromPEM([]byte(cert.Certificate))
		} else if _, ok := token.Method.(*jwt.SigningMethodECDSA); ok {
			// ES certificate
			certificate, err = jwt.ParseECPublicKeyFromPEM([]byte(cert.Certificate))
		} else {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if err != nil {
			return nil, err
		}

		return certificate, nil
	})

	if t != nil {
		if claims, ok := t.Claims.(*Claims); ok && t.Valid {
			return claims, nil
		}
	}

	return nil, err
}

func ParseJwtTokenByApplication(token string, application *Application) (*Claims, error) {
	cert, err := getCertByApplication(application)
	if err != nil {
		return nil, err
	}

	return ParseJwtToken(token, cert)
}
