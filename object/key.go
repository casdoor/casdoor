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
	"errors"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/i18n"
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

func GenerateKeySecret() string {
	return fmt.Sprintf("casdoor_key_%s%s", util.GenerateClientId(), util.GenerateClientSecret())
}

func getKeyPreview(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 12 {
		return secret
	}
	return fmt.Sprintf("%s...%s", secret[:8], secret[len(secret)-4:])
}

func getKeyHash(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	res := hex.EncodeToString(hash[:])
	if len(res) > 64 {
		return res[:64]
	}
	return res
}

func (key *Key) SetSecret(secret string) {
	key.SecretPreview = getKeyPreview(secret)
	key.SecretHash = getKeyHash(secret)
}

func CheckKey(key *Key, lang string) error {
	if key == nil {
		return errors.New(i18n.Translate(lang, "general:Missing parameter"))
	}

	key.Owner = strings.TrimSpace(key.Owner)
	key.Name = strings.TrimSpace(key.Name)
	key.Type = strings.TrimSpace(key.Type)
	key.Organization = strings.TrimSpace(key.Organization)
	key.Application = strings.TrimSpace(key.Application)
	key.User = strings.TrimSpace(key.User)

	if key.Owner == "" {
		return errors.New("key owner cannot be empty")
	}
	if key.Name == "" {
		return errors.New("key name cannot be empty")
	}
	if key.Application == "" {
		return errors.New("key application cannot be empty")
	}

	application, err := GetApplication(util.GetId("admin", key.Application))
	if err != nil {
		return err
	}
	if application == nil {
		return fmt.Errorf("the application: %s does not exist", key.Application)
	}

	switch key.Type {
	case KeyTypeOrganization:
		if key.Organization == "" {
			return errors.New("key organization cannot be empty")
		}
		organization, err := GetOrganization(util.GetId("admin", key.Organization))
		if err != nil {
			return err
		}
		if organization == nil {
			return fmt.Errorf("the organization: %s does not exist", key.Organization)
		}
		if application.Organization != key.Organization {
			return fmt.Errorf("the application: %s does not belong to organization: %s", key.Application, key.Organization)
		}
	case KeyTypeApplication:
		if key.Application == "" {
			return errors.New("key application cannot be empty")
		}
	case KeyTypeUser:
		if key.Organization == "" {
			return errors.New("user key organization cannot be empty")
		}
		if key.User == "" {
			return errors.New("user key user cannot be empty")
		}
		user, err := GetUser(util.GetId(key.Organization, key.User))
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("the user: %s does not exist", util.GetId(key.Organization, key.User))
		}
		if user.Owner != key.Organization {
			return fmt.Errorf("the user: %s does not belong to organization: %s", util.GetId(key.Organization, key.User), key.Organization)
		}
		if application.Organization != key.Organization {
			return fmt.Errorf("the application: %s does not belong to organization: %s", key.Application, key.Organization)
		}
	case KeyTypeGeneral:
	default:
		return fmt.Errorf("unsupported key type: %s", key.Type)
	}

	return nil
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
