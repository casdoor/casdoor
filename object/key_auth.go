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

package object

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/casdoor/casdoor/util"
)

func GetUsernameByKey(accessKey string, accessSecret string) (string, error) {
	if accessKey == "" || accessSecret == "" {
		return "", nil
	}

	key, err := GetKeyByAccessKey(accessKey)
	if err != nil {
		return "", err
	}
	if key == nil {
		return "", nil
	}

	if subtle.ConstantTimeCompare([]byte(key.AccessSecret), []byte(accessSecret)) != 1 {
		return "", fmt.Errorf("incorrect access secret for key: %s", key.Name)
	}
	if !key.IsActive() {
		return "", fmt.Errorf("key: %s is inactive", key.GetId())
	}

	expired, err := key.IsExpired()
	if err != nil {
		return "", err
	}
	if expired {
		return "", fmt.Errorf("key: %s is expired", key.GetId())
	}

	organization, err := key.getBoundOrganization()
	if err != nil {
		return "", err
	}

	return getUsernameFromKey(key, organization)
}

func (key *Key) IsActive() bool {
	return key != nil && (key.State == "" || key.State == KeyStateActive)
}

func (key *Key) IsExpired() (bool, error) {
	if key == nil || key.ExpireTime == "" {
		return false, nil
	}

	expireTime, err := parseKeyExpireTime(key.ExpireTime)
	if err != nil {
		return false, err
	}
	return time.Now().After(expireTime), nil
}

func parseKeyExpireTime(expireTime string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, expireTime); err == nil {
		return parsed, nil
	}

	localLayouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, layout := range localLayouts {
		if parsed, err := time.ParseInLocation(layout, expireTime, time.Local); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid expire time format: %s", expireTime)
}

func getUsernameFromKey(key *Key, organization string) (string, error) {
	switch key.Type {
	case KeyTypeUser:
		if key.User == "" {
			return "", fmt.Errorf("user key: %s is not bound to a user", key.GetId())
		}

		user, err := getUser(organization, key.User)
		if err != nil {
			return "", err
		}
		if user == nil {
			return "", fmt.Errorf("the user: %s does not exist", util.GetId(organization, key.User))
		}
		if user.IsForbidden {
			return "", fmt.Errorf("the user: %s is forbidden", user.GetId())
		}
		// User keys return the normal "owner/name" user id.
		return user.GetId(), nil
	case KeyTypeApplication:
		if key.Application == "" {
			return "", fmt.Errorf("application key: %s is not bound to an application", key.GetId())
		}

		application, err := GetApplication(util.GetId("admin", key.Application))
		if err != nil {
			return "", err
		}
		if application == nil {
			return "", fmt.Errorf("the application: %s does not exist", key.Application)
		}
		if application.Organization != organization {
			return "", fmt.Errorf("application: %s does not belong to organization: %s", application.Name, organization)
		}
		// Application keys reuse the existing "app/<application>" format.
		return fmt.Sprintf("app/%s", application.Name), nil
	case KeyTypeOrganization, KeyTypeGeneral:
		return "", fmt.Errorf("key type: %s is not supported for direct authentication yet", key.Type)
	default:
		return "", fmt.Errorf("unsupported key type: %s", key.Type)
	}
}

func (key *Key) getBoundOrganization() (string, error) {
	if key == nil {
		return "", fmt.Errorf("key is nil")
	}
	if key.Owner == "" {
		return "", fmt.Errorf("key: %s has empty owner", key.Name)
	}
	if key.Organization != "" && key.Organization != key.Owner {
		return "", fmt.Errorf("key: %s organization: %s does not match owner: %s", key.GetId(), key.Organization, key.Owner)
	}
	return key.Owner, nil
}
