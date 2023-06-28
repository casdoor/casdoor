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
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	*User
	TokenType string `json:"tokenType,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Scope     string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type UserShort struct {
	Owner string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name  string `xorm:"varchar(100) notnull pk" json:"name"`
}

type UserWithoutThirdIdp struct {
	Owner               string            `xorm:"varchar(100) notnull pk" json:"owner"`
	Name                string            `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime         string            `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime         string            `xorm:"varchar(100)" json:"updatedTime"`
	Id                  string            `xorm:"varchar(100) index" json:"id"`
	Type                string            `xorm:"varchar(100)" json:"type"`
	Password            string            `xorm:"varchar(100)" json:"password"`
	PasswordSalt        string            `xorm:"varchar(100)" json:"passwordSalt"`
	DisplayName         string            `xorm:"varchar(100)" json:"displayName"`
	FirstName           string            `xorm:"varchar(100)" json:"firstName"`
	LastName            string            `xorm:"varchar(100)" json:"lastName"`
	Avatar              string            `xorm:"varchar(500)" json:"avatar"`
	PermanentAvatar     string            `xorm:"varchar(500)" json:"permanentAvatar"`
	Email               string            `xorm:"varchar(100) index" json:"email"`
	EmailVerified       bool              `json:"emailVerified"`
	Phone               string            `xorm:"varchar(100) index" json:"phone"`
	Location            string            `xorm:"varchar(100)" json:"location"`
	Address             []string          `json:"address"`
	Affiliation         string            `xorm:"varchar(100)" json:"affiliation"`
	Title               string            `xorm:"varchar(100)" json:"title"`
	IdCardType          string            `xorm:"varchar(100)" json:"idCardType"`
	IdCard              string            `xorm:"varchar(100) index" json:"idCard"`
	Homepage            string            `xorm:"varchar(100)" json:"homepage"`
	Bio                 string            `xorm:"varchar(100)" json:"bio"`
	Tag                 string            `xorm:"varchar(100)" json:"tag"`
	Region              string            `xorm:"varchar(100)" json:"region"`
	Language            string            `xorm:"varchar(100)" json:"language"`
	Gender              string            `xorm:"varchar(100)" json:"gender"`
	Birthday            string            `xorm:"varchar(100)" json:"birthday"`
	Education           string            `xorm:"varchar(100)" json:"education"`
	Score               int               `json:"score"`
	Karma               int               `json:"karma"`
	Ranking             int               `json:"ranking"`
	IsDefaultAvatar     bool              `json:"isDefaultAvatar"`
	IsOnline            bool              `json:"isOnline"`
	IsAdmin             bool              `json:"isAdmin"`
	IsGlobalAdmin       bool              `json:"isGlobalAdmin"`
	IsForbidden         bool              `json:"isForbidden"`
	IsDeleted           bool              `json:"isDeleted"`
	SignupApplication   string            `xorm:"varchar(100)" json:"signupApplication"`
	Hash                string            `xorm:"varchar(100)" json:"hash"`
	PreHash             string            `xorm:"varchar(100)" json:"preHash"`
	CreatedIp           string            `xorm:"varchar(100)" json:"createdIp"`
	LastSigninTime      string            `xorm:"varchar(100)" json:"lastSigninTime"`
	LastSigninIp        string            `xorm:"varchar(100)" json:"lastSigninIp"`
	Ldap                string            `xorm:"ldap varchar(100)" json:"ldap"`
	Properties          map[string]string `json:"properties"`
	Roles               []*Role           `xorm:"-" json:"roles"`
	Permissions         []*Permission     `xorm:"-" json:"permissions"`
	LastSigninWrongTime string            `xorm:"varchar(100)" json:"lastSigninWrongTime"`
	SigninWrongTimes    int               `json:"signinWrongTimes"`
}

type ClaimsShort struct {
	*UserShort
	TokenType string `json:"tokenType,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
	Scope     string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type ClaimsWithoutThirdIdp struct {
	*UserWithoutThirdIdp
	TokenType string `json:"tokenType,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Scope     string `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

func getShortUser(user *User) *UserShort {
	res := &UserShort{
		Owner: user.Owner,
		Name:  user.Name,
	}
	return res
}

func getUserWithoutThirdIdp(user *User) *UserWithoutThirdIdp {
	res := &UserWithoutThirdIdp{
		Owner:       user.Owner,
		Name:        user.Name,
		CreatedTime: user.CreatedTime,
		UpdatedTime: user.UpdatedTime,

		Id:                user.Id,
		Type:              user.Type,
		Password:          user.Password,
		PasswordSalt:      user.PasswordSalt,
		DisplayName:       user.DisplayName,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Avatar:            user.Avatar,
		PermanentAvatar:   user.PermanentAvatar,
		Email:             user.Email,
		EmailVerified:     user.EmailVerified,
		Phone:             user.Phone,
		Location:          user.Location,
		Address:           user.Address,
		Affiliation:       user.Affiliation,
		Title:             user.Title,
		IdCardType:        user.IdCardType,
		IdCard:            user.IdCard,
		Homepage:          user.Homepage,
		Bio:               user.Bio,
		Tag:               user.Tag,
		Region:            user.Region,
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
		IsGlobalAdmin:     user.IsGlobalAdmin,
		IsForbidden:       user.IsForbidden,
		IsDeleted:         user.IsDeleted,
		SignupApplication: user.SignupApplication,
		Hash:              user.Hash,
		PreHash:           user.PreHash,

		CreatedIp:      user.CreatedIp,
		LastSigninTime: user.LastSigninTime,
		LastSigninIp:   user.LastSigninIp,

		Ldap:       user.Ldap,
		Properties: user.Properties,

		Roles:       user.Roles,
		Permissions: user.Permissions,

		LastSigninWrongTime: user.LastSigninWrongTime,
		SigninWrongTimes:    user.SigninWrongTimes,
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
	if user.PasswordChangeRequired {
		user.Permissions = []*Permission{}
	}

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

	var token *jwt.Token
	var refreshToken *jwt.Token

	// the JWT token length in "JWT-Empty" mode will be very short, as User object only has two properties: owner and name
	if application.TokenFormat == "JWT-Empty" {
		claimsShort := getShortClaims(claims)

		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claimsShort)
		claimsShort.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
		claimsShort.TokenType = "refresh-token"
		refreshToken = jwt.NewWithClaims(jwt.SigningMethodRS256, claimsShort)
	} else {
		claimsWithoutThirdIdp := getClaimsWithoutThirdIdp(claims)

		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claimsWithoutThirdIdp)
		claimsWithoutThirdIdp.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
		claimsWithoutThirdIdp.TokenType = "refresh-token"
		refreshToken = jwt.NewWithClaims(jwt.SigningMethodRS256, claimsWithoutThirdIdp)
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return "", "", "", err
	}

	// RSA private key
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cert.PrivateKey))
	if err != nil {
		return "", "", "", err
	}

	token.Header["kid"] = cert.Name
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", "", "", err
	}
	refreshTokenString, err := refreshToken.SignedString(key)

	return tokenString, refreshTokenString, name, err
}

func ParseJwtToken(token string, cert *Cert) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// RSA certificate
		certificate, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cert.Certificate))
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
