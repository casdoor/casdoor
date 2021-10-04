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
    "fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("CasdoorSecret")

type Claims struct {
	User
	jwt.StandardClaims
}

func generateJwtToken(application *Application, user *User) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(application.ExpireInHours) * time.Hour)

	user.Password = ""

	claims := Claims{
		User: *user,
		StandardClaims: jwt.StandardClaims{
			Audience:  application.ClientId,
			ExpiresAt: expireTime.Unix(),
			Id:        "",
			IssuedAt:  nowTime.Unix(),
			Issuer:    "casdoor",
			NotBefore: nowTime.Unix(),
			Subject:   user.Id,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString([]byte(application.ClientSecret))

	return token, err
}

func ParseJwtToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
	    // TODO: 获取 token 关联的 application.ClientSecret， 并替代 jwtSecret
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}
