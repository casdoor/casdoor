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
	"fmt"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	KeyTypeOrganization = "Organization"
	KeyTypeApplication  = "Application"
	KeyTypeUser         = "User"
	KeyTypeGeneral      = "General"

	KeyStateActive   = "Active"
	KeyStateInactive = "Inactive"
)

type Key struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	// Type indicates the scope this key belongs to: "Organization", "Application", "User", or "General"
	Type         string `xorm:"varchar(100)" json:"type"`
	Organization string `xorm:"varchar(100)" json:"organization"`
	Application  string `xorm:"varchar(100)" json:"application"`
	User         string `xorm:"varchar(100)" json:"user"`

	AccessKey    string `xorm:"varchar(100) index" json:"accessKey"`
	AccessSecret string `xorm:"varchar(100)" json:"accessSecret"`

	ExpireTime string `xorm:"varchar(100)" json:"expireTime"`
	State      string `xorm:"varchar(100)" json:"state"`
}

func GetKeyCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Key{})
}

func GetGlobalKeyCount(field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Count(&Key{})
}

func GetKeys(owner string) ([]*Key, error) {
	keys := []*Key{}
	err := ormer.Engine.Desc("created_time").Find(&keys, &Key{Owner: owner})
	if err != nil {
		return keys, err
	}
	return keys, nil
}

func GetPaginationKeys(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Key, error) {
	keys := []*Key{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}

func GetGlobalKeys() ([]*Key, error) {
	keys := []*Key{}
	err := ormer.Engine.Desc("created_time").Find(&keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}

func GetPaginationGlobalKeys(offset, limit int, field, value, sortField, sortOrder string) ([]*Key, error) {
	keys := []*Key{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}

func getKey(owner, name string) (*Key, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	key := Key{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&key)
	if err != nil {
		return &key, err
	}

	if existed {
		return &key, nil
	}
	return nil, nil
}

func GetKey(id string) (*Key, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getKey(owner, name)
}

func GetMaskedKey(key *Key, isMaskEnabled bool) *Key {
	if !isMaskEnabled {
		return key
	}

	if key == nil {
		return nil
	}

	if key.AccessSecret != "" {
		key.AccessSecret = "***"
	}
	return key
}

func GetMaskedKeys(keys []*Key, isMaskEnabled bool, err error) ([]*Key, error) {
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		GetMaskedKey(key, isMaskEnabled)
	}
	return keys, nil
}

func UpdateKey(id string, key *Key) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if k, err := getKey(owner, name); err != nil {
		return false, err
	} else if k == nil {
		return false, nil
	}

	key.UpdatedTime = util.GetCurrentTime()

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(key)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddKey(key *Key) (bool, error) {
	if key.AccessKey == "" {
		key.AccessKey = util.GenerateId()
	}
	if key.AccessSecret == "" {
		key.AccessSecret = util.GenerateId()
	}

	affected, err := ormer.Engine.Insert(key)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteKey(key *Key) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{key.Owner, key.Name}).Delete(&Key{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (key *Key) GetId() string {
	return fmt.Sprintf("%s/%s", key.Owner, key.Name)
}

func GetKeyByAccessKey(accessKey string) (*Key, error) {
	if accessKey == "" {
		return nil, nil
	}

	key := Key{AccessKey: accessKey}
	existed, err := ormer.Engine.Get(&key)
	if err != nil {
		return nil, err
	}

	if existed {
		return &key, nil
	}
	return nil, nil
}

func GetKeyBySecret(accessSecret string) (*Key, error) {
	if accessSecret == "" {
		return nil, nil
	}

	key := Key{AccessSecret: accessSecret}
	existed, err := ormer.Engine.Get(&key)
	if err != nil {
		return nil, err
	}

	if existed {
		return &key, nil
	}
	return nil, nil
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
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, expireTime); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid expire time format: %s", expireTime)
}

func ResolveSubjectByKey(accessKey string, accessSecret string) (string, error) {
	if accessKey == "" || accessSecret == "" {
		return "", nil
	}

	key, err := GetKeyByAccessKey(accessKey)
	if err != nil {
		return "", err
	}
	if key == nil {
		return "", fmt.Errorf("key not found for access key: %s", accessKey)
	}

	if key.AccessSecret != accessSecret {
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
