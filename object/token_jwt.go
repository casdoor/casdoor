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

type Claims struct {
	*User
	Name  string `json:"name,omitempty"`
	Owner string `json:"owner,omitempty"`
	Nonce string `json:"nonce,omitempty"`
	jwt.RegisteredClaims
}

func generateJwtToken(application *Application, user *User, nonce string) (string, string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(application.ExpireInHours) * time.Hour)
	refreshExpireTime := nowTime.Add(time.Duration(application.RefreshExpireInHours) * time.Hour)

	user.Password = ""

	claims := Claims{
		User:  user,
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
	//all fields of the User struct are not added in "JWT-Empty" format
	if application.TokenFormat == "JWT-Empty" {
		claims.User = nil
	}
	claims.Name = user.Name
	claims.Owner = user.Owner

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	claims.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	cert := getCertByApplication(application)

	// RSA private key
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cert.PrivateKey))
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

func ParseJwtToken(token string, cert *Cert) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// RSA public key
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cert.PublicKey))
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
