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

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	siteIdentityTtl      = time.Minute
	siteIdentityMaxCount = 10000
)

var siteUserHeaders = []string{"X-Forwarded-User", "X-Forwarded-Email", "X-Forwarded-Groups"}

type siteIdentity struct {
	name      string
	email     string
	groups    []string
	isActive  bool
	isAllowed bool
	expireAt  time.Time
}

var (
	siteIdentityMap   = map[string]*siteIdentity{}
	siteIdentityMutex sync.Mutex
)

func getSiteIdentityKey(site *object.Site, accessToken string) string {
	hash := sha256.Sum256([]byte(accessToken))
	return fmt.Sprintf("%s|%s", site.GetId(), hex.EncodeToString(hash[:]))
}

func loadSiteIdentity(site *object.Site, accessToken string) (*siteIdentity, error) {
	token, err := object.GetTokenByAccessToken(accessToken)
	if err != nil {
		return nil, err
	}
	if token == nil || token.User == "" {
		return nil, nil
	}

	userId := util.GetId(token.Organization, token.User)
	user, err := object.GetUser(userId)
	if err != nil {
		return nil, err
	}

	identity := &siteIdentity{}
	if user == nil || user.IsForbidden || user.IsDeleted {
		return identity, nil
	}

	identity.name = user.Name
	identity.email = user.Email
	for _, group := range user.Groups {
		identity.groups = append(identity.groups, group[strings.LastIndex(group, "/")+1:])
	}
	identity.isActive = true

	identity.isAllowed, err = object.CheckLoginPermission(userId, site.ApplicationObj)
	if err != nil {
		return nil, err
	}
	return identity, nil
}

// getSiteIdentity returns nil if the access token is no longer known to Casdoor.
func getSiteIdentity(site *object.Site, accessToken string) (*siteIdentity, error) {
	key := getSiteIdentityKey(site, accessToken)
	now := time.Now()

	siteIdentityMutex.Lock()
	identity, ok := siteIdentityMap[key]
	siteIdentityMutex.Unlock()
	if ok && now.Before(identity.expireAt) {
		return identity, nil
	}

	identity, err := loadSiteIdentity(site, accessToken)
	if err != nil || identity == nil {
		return nil, err
	}
	identity.expireAt = now.Add(siteIdentityTtl)

	siteIdentityMutex.Lock()
	if len(siteIdentityMap) >= siteIdentityMaxCount {
		siteIdentityMap = map[string]*siteIdentity{}
	}
	siteIdentityMap[key] = identity
	siteIdentityMutex.Unlock()
	return identity, nil
}

func clearUserHeaders(r *http.Request) {
	for _, header := range siteUserHeaders {
		r.Header.Del(header)
	}
}

func setUserHeaders(r *http.Request, identity *siteIdentity) {
	r.Header.Set("X-Forwarded-User", identity.name)
	if identity.email != "" {
		r.Header.Set("X-Forwarded-Email", identity.email)
	}
	if len(identity.groups) != 0 {
		r.Header.Set("X-Forwarded-Groups", strings.Join(identity.groups, ","))
	}
}

func clearAccessTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "casdoor_access_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

func responseForbidden(w http.ResponseWriter, format string, a ...interface{}) {
	w.WriteHeader(http.StatusForbidden)
	responseErrorWithoutCode(w, format, a...)
}
