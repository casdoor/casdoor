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
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/casbin/casdoor/util"
	"io/ioutil"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var tokenJwtPublicKey string

var tokenJwtPrivateKey string

type Claims struct {
	User
	jwt.RegisteredClaims
}

func generateJwtToken(application *Application, user *User) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(application.ExpireInHours) * time.Hour)

	user.Password = ""

	claims := Claims{
		User: *user,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Casdoor",
			Subject:   user.Id,
			Audience:  []string{application.ClientId},
			ExpiresAt: jwt.NewNumericDate(expireTime),
			NotBefore: jwt.NewNumericDate(nowTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			ID:        "",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Use "token_jwt_key.key" as RSA private key
	privateKey := tokenJwtPrivateKey
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString(key)

	return tokenString, err
}

func InitJwtToken() {
	publicKeyFile := beego.AppConfig.String("jwtPublicKeyFile")
	privateKeyFile := beego.AppConfig.String("jwtPrivateKeyFile")
	if !util.FileExist(publicKeyFile) && !util.FileExist(privateKeyFile) {
		// generate new certs
		logs.Info("Generating jwt tokens")
		generateRsaKeys(privateKeyFile, publicKeyFile)
		logs.Info("Generated jwt token: %s and %s", publicKeyFile, privateKeyFile)
	}

	buffer, err := ioutil.ReadFile(publicKeyFile)
	if err != nil {
		panic(err)
	}
	tokenJwtPublicKey = string(buffer)

	buffer, err = ioutil.ReadFile(privateKeyFile)
	if err != nil {
		panic(err)
	}
	tokenJwtPrivateKey = string(buffer)
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
