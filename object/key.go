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
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
)

const (
	KeyTypeOrganization = "organization"
	KeyTypeApplication  = "application"
	KeyTypeUser         = "user"
	KeyTypeGeneral      = "general"
)

type Key struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	DisplayName  string   `xorm:"varchar(100)" json:"displayName"`
	Description  string   `xorm:"varchar(100)" json:"description"`
	Type         string   `xorm:"varchar(100) index" json:"type"`
	Organization string   `xorm:"varchar(100) index" json:"organization"`
	Application  string   `xorm:"varchar(100) index" json:"application"`
	User         string   `xorm:"varchar(100) index" json:"user"`
	Scopes       []string `xorm:"mediumtext" json:"scopes"`

	IsEnabled    bool   `json:"isEnabled"`
	ExpiresTime  string `xorm:"varchar(100)" json:"expiresTime"`
	LastUsedTime string `xorm:"varchar(100)" json:"lastUsedTime"`

	SecretPreview string `xorm:"varchar(100)" json:"secretPreview"`
	SecretHash    string `xorm:"varchar(100) index" json:"secretHash"`
}

func (key *Key) GetId() string {
	return fmt.Sprintf("%s/%s", key.Owner, key.Name)
}

func getKeyHash(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	res := hex.EncodeToString(hash[:])
	if len(res) > 64 {
		return res[:64]
	}
	return res
}

func GetKeyCount(owner, keyType, organization, application, user, field, value string) (int64, error) {
	session := getKeySession(owner, keyType, organization, application, user, -1, -1, field, value, "", "")
	return session.Count(&Key{})
}

func GetKeys(owner, keyType, organization, application, user string) ([]*Key, error) {
	keys := []*Key{}
	session := getKeySession(owner, keyType, organization, application, user, -1, -1, "", "", "", "")
	err := session.Find(&keys)
	if err != nil {
		return keys, err
	}

	return keys, nil
}

func GetPaginationKeys(owner, keyType, organization, application, user string, offset, limit int, field, value, sortField, sortOrder string) ([]*Key, error) {
	keys := []*Key{}
	session := getKeySession(owner, keyType, organization, application, user, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&keys)
	if err != nil {
		return keys, err
	}

	return keys, nil
}

func getKeySession(owner, keyType, organization, application, user string, offset, limit int, field, value, sortField, sortOrder string) *xorm.Session {
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	if keyType != "" {
		session = session.And("type = ?", keyType)
	}
	if organization != "" {
		session = session.And("organization = ?", organization)
	}
	if application != "" {
		session = session.And("application = ?", application)
	}
	if user != "" {
		session = session.And("user = ?", user)
	}
	return session
}

func getKey(owner, name string) (*Key, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	key := Key{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&key)
	if err != nil {
		return nil, err
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

func GetKeyBySecret(secret string) (*Key, error) {
	if secret == "" {
		return nil, nil
	}

	key := Key{SecretHash: getKeyHash(secret)}
	existed, err := ormer.Engine.Get(&key)
	if err != nil {
		return nil, err
	}

	if existed {
		return &key, nil
	}

	return nil, nil
}

func GetMaskedKey(key *Key, errs ...error) (*Key, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	if key == nil {
		return nil, nil
	}

	key.SecretHash = ""
	return key, nil
}

func GetMaskedKeys(keys []*Key, errs ...error) ([]*Key, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	var err error
	for _, key := range keys {
		key, err = GetMaskedKey(key)
		if err != nil {
			return nil, err
		}
	}

	return keys, nil
}

func AddKey(key *Key) (bool, error) {
	if key.CreatedTime == "" {
		key.CreatedTime = util.GetCurrentTime()
	}
	key.UpdatedTime = util.GetCurrentTime()

	affected, err := ormer.Engine.Insert(key)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func UpdateKey(id string, key *Key) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}

	existed, err := getKey(owner, name)
	if err != nil {
		return false, err
	}
	if existed == nil {
		return false, nil
	}

	key.UpdatedTime = util.GetCurrentTime()

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(key)
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
