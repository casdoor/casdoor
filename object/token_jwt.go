// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	_ "embed"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/golang-jwt/jwt/v4"
)

//go:embed token_jwt_key.pem
var tokenJwtPublicKey string

//go:embed token_jwt_key.key
var tokenJwtPrivateKey string

type Claims struct {
	User
	Nonce string `json:"nonce,omitempty"`
	jwt.RegisteredClaims
}

func generateJwtToken(application *Application, user *User, nonce string) (string, string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(application.ExpireInHours) * time.Hour)
	refreshExpireTime := nowTime.Add(time.Duration(application.RefreshExpireInHours) * time.Hour)

	user.Password = ""

	claims := Claims{
		User:  *user,
		Nonce: nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    beego.AppConfig.String("origin"),
			Subject:   user.Id,
			Audience:  []string{application.ClientId},
			ExpiresAt: jwt.NewNumericDate(expireTime),
			NotBefore: jwt.NewNumericDate(nowTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			ID:        "",
		},
	}

	if application.TokenFormat == "JWT-Empty" {
		claims.User = User{}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	claims.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Use "token_jwt_key.key" as RSA private key
	privateKey := tokenJwtPrivateKey
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return "", "", err
	}

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", "", err
	}
	refreshTokenString, err := refreshToken.SignedString(key)

	return tokenString, refreshTokenString, err
}

func ParseJwtToken(token string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Use "token_jwt_key.pem" as RSA public key
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(tokenJwtPublicKey))
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if t != nil {
		if claims, ok := t.Claims.(*Claims); ok && t.Valid {
			return claims, nil
		}
	}

	return nil, err
}
