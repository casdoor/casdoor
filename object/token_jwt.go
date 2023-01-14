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

type ClaimsShort struct {
	*UserShort
	TokenType string `json:"tokenType,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
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

func generateJwtToken(application *Application, user *User, nonce string, scope string, host string) (string, string, string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(application.ExpireInHours) * time.Hour)
	refreshExpireTime := nowTime.Add(time.Duration(application.RefreshExpireInHours) * time.Hour)

	user.Password = ""
	_, originBackend := getOriginFromHost(host)

	name := util.GenerateId()
	jti := fmt.Sprintf("%s/%s", application.Owner, name)

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
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		claims.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
		claims.TokenType = "refresh-token"
		refreshToken = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	}

	cert := getCertByApplication(application)

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
	return ParseJwtToken(token, getCertByApplication(application))
}
