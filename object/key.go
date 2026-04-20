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

// ValidateKeyByType checks whether the given accessKey/accessSecret pair refers to a
// valid, active, non-expired Key record of the specified type. It returns the matched
// Key on success and an error describing the validation failure otherwise.
func ValidateKeyByType(accessKey, accessSecret, keyType string) (*Key, error) {
	if accessKey == "" || accessSecret == "" {
		return nil, fmt.Errorf("accessKey and accessSecret must not be empty")
	}

	key, err := GetKeyByAccessKey(accessKey)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, fmt.Errorf("access key not found: %s", accessKey)
	}

	if key.Type != keyType {
		return nil, fmt.Errorf("access key type is not %s: %s", keyType, key.Name)
	}

	if key.AccessSecret != accessSecret {
		return nil, fmt.Errorf("incorrect access secret for key: %s", key.Name)
	}

	if key.State != "Active" {
		return nil, fmt.Errorf("access key is not active: %s", key.Name)
	}

	if key.ExpireTime != "" {
		expireTime, err := time.Parse(time.RFC3339, key.ExpireTime)
		if err != nil {
			return nil, fmt.Errorf("invalid expire time format for key: %s", key.Name)
		}
		if time.Now().After(expireTime) {
			return nil, fmt.Errorf("access key has expired, expireTime = %s", key.ExpireTime)
		}
	}

	return key, nil
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
