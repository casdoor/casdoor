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
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/gm"
	"github.com/emmansun/gmsm/sm3"
	"github.com/xorm-io/core"
)

// The personal fields of users that an organization can keep encrypted with SM4 in the database. The phone and the
// email are not among them, as users sign in and are looked up by them. "managedAccounts" stands for the passwords of
// the managed accounts that form autofill signs users in with.
var EncryptableUserFields = []string{"idCard", "realName", "birthday", "location", "address", "managedAccounts"}

const encryptedFieldPrefix = "sm4:"

// getFieldEncryptionMasterKey reads the configuration "fieldEncryptionKey", 32 hex digits or any passphrase
func getFieldEncryptionMasterKey() []byte {
	key := strings.TrimSpace(conf.GetConfigString("fieldEncryptionKey"))
	if key == "" {
		return nil
	}
	if decoded, err := hex.DecodeString(key); err == nil && len(decoded) >= 16 {
		return decoded
	}
	sum := sm3.Sum([]byte(key))
	return sum[:]
}

func IsFieldEncryptionKeyConfigured() bool {
	return getFieldEncryptionMasterKey() != nil
}

var fieldCipherCache sync.Map

// getOrganizationFieldKey derives the SM4 key of an organization from the master key, so that each organization's
// fields are encrypted with its own key
func getOrganizationFieldKey(owner string) ([]byte, error) {
	masterKey := getFieldEncryptionMasterKey()
	if masterKey == nil {
		return nil, errors.New("the configuration fieldEncryptionKey is not set")
	}
	mac := hmac.New(sm3.New, masterKey)
	mac.Write([]byte("casdoor-field-encryption:" + owner))
	return mac.Sum(nil)[:16], nil
}

func encryptField(owner string, field string, value string) (string, error) {
	if value == "" || strings.HasPrefix(value, encryptedFieldPrefix) {
		return value, nil
	}
	key, err := getOrganizationFieldKey(owner)
	if err != nil {
		return "", err
	}
	aead, err := gm.NewSm4Gcm(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nonce, nonce, []byte(value), []byte(field))
	return encryptedFieldPrefix + base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func decryptField(owner string, field string, value string) (string, error) {
	if !strings.HasPrefix(value, encryptedFieldPrefix) {
		return value, nil
	}
	data, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, encryptedFieldPrefix))
	if err != nil {
		return "", err
	}
	key, err := getOrganizationFieldKey(owner)
	if err != nil {
		return "", err
	}
	aead, err := gm.NewSm4Gcm(key)
	if err != nil {
		return "", err
	}
	if len(data) < aead.NonceSize() {
		return "", fmt.Errorf("the encrypted %s is too short", field)
	}
	plaintext, err := aead.Open(nil, data[:aead.NonceSize()], data[aead.NonceSize():], []byte(field))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt %s: %w", field, err)
	}
	return string(plaintext), nil
}

// transformUserFields encrypts or decrypts the given personal fields of the user in place
func (user *User) transformUserFields(fields []string, transform func(owner string, field string, value string) (string, error)) error {
	for _, field := range fields {
		var err error
		switch field {
		case "idCard":
			user.IdCard, err = transform(user.Owner, field, user.IdCard)
		case "realName":
			user.RealName, err = transform(user.Owner, field, user.RealName)
		case "birthday":
			user.Birthday, err = transform(user.Owner, field, user.Birthday)
		case "location":
			user.Location, err = transform(user.Owner, field, user.Location)
		case "address":
			for i := range user.Address {
				user.Address[i], err = transform(user.Owner, field, user.Address[i])
				if err != nil {
					break
				}
			}
		case "managedAccounts":
			for i := range user.ManagedAccounts {
				user.ManagedAccounts[i].Password, err = transform(user.Owner, field, user.ManagedAccounts[i].Password)
				if err != nil {
					break
				}
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (user *User) hasEncryptedFields() bool {
	if strings.HasPrefix(user.IdCard, encryptedFieldPrefix) || strings.HasPrefix(user.RealName, encryptedFieldPrefix) ||
		strings.HasPrefix(user.Birthday, encryptedFieldPrefix) || strings.HasPrefix(user.Location, encryptedFieldPrefix) {
		return true
	}
	for _, address := range user.Address {
		if strings.HasPrefix(address, encryptedFieldPrefix) {
			return true
		}
	}
	for _, account := range user.ManagedAccounts {
		if strings.HasPrefix(account.Password, encryptedFieldPrefix) {
			return true
		}
	}
	return false
}

func (user *User) encryptFields() {
	policy := GetCachedEnterprisePolicy(user.Owner)
	if policy == nil || len(policy.EncryptedFields) == 0 {
		return
	}
	if err := user.transformUserFields(policy.EncryptedFields, encryptField); err != nil {
		logs.Error("failed to encrypt the personal fields of the user: %s: %v", user.GetId(), err)
	}
}

func (user *User) decryptFields() {
	if !user.hasEncryptedFields() {
		return
	}
	if err := user.transformUserFields(EncryptableUserFields, decryptField); err != nil {
		logs.Error("failed to decrypt the personal fields of the user: %s: %v", user.GetId(), err)
	}
}

// The xorm processors keep the fields encrypted in the database and in plaintext everywhere else

func (user *User) BeforeInsert() {
	user.encryptFields()
}

func (user *User) AfterInsert() {
	user.decryptFields()
}

func (user *User) BeforeUpdate() {
	user.encryptFields()
}

func (user *User) AfterUpdate() {
	user.decryptFields()
}

func (user *User) AfterLoad() {
	user.decryptFields()
}

// ReencryptOrganizationUsers rewrites the personal fields of all users of the organization after its encrypted fields
// changed: the fields now listed get encrypted, the others go back to plaintext
func ReencryptOrganizationUsers(owner string) (int, error) {
	clearEnterprisePolicyCache(owner)
	columns := []string{"id_card", "real_name", "birthday", "location", "address", "managedAccounts"}

	count := 0
	for offset := 0; ; offset += 500 {
		users := []*User{}
		err := ormer.Engine.Where("owner = ?", owner).Asc("name").Limit(500, offset).Find(&users)
		if err != nil {
			return count, err
		}
		if len(users) == 0 {
			return count, nil
		}
		for _, user := range users {
			// AfterLoad has decrypted every field, BeforeUpdate encrypts the listed ones
			_, err = ormer.Engine.ID(core.PK{user.Owner, user.Name}).Cols(columns...).Update(user)
			if err != nil {
				return count, err
			}
			count++
		}
	}
}
