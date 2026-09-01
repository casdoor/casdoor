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

package ldap

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/casdoor/casdoor/object"
	"golang.org/x/crypto/md4"
)

const (
	sambaAcctFlagsNormal            = "[U          ]"
	sambaAcctFlagsDisabled          = "[DU         ]"
	sambaLmPasswordPlaceholder      = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	sambaPasswordHistoryPlaceholder = "0000000000000000000000000000000000000000000000000000000000000000"
)

// getSambaDomainSid derives the domain SID from the organization name, so all
// users of an organization share one prefix without it having to be stored.
func getSambaDomainSid(orgName string) string {
	return fmt.Sprintf("S-1-5-21-%d-%d-%d", hash(orgName+"-samba-1"), hash(orgName+"-samba-2"), hash(orgName+"-samba-3"))
}

func getSambaSid(user *object.User) string {
	if sid := getUserProperty(user, "sambaSID"); sid != "" {
		return sid
	}
	return fmt.Sprintf("%s-%s", getSambaDomainSid(user.Owner), getUidNumber(user))
}

// getSambaPrimaryGroupSid uses the same RID, as the LDAP server publishes gidNumber = uidNumber.
func getSambaPrimaryGroupSid(user *object.User) string {
	if sid := getUserProperty(user, "sambaPrimaryGroupSID"); sid != "" {
		return sid
	}
	return fmt.Sprintf("%s-%s", getSambaDomainSid(user.Owner), getUidNumber(user))
}

func getSambaAcctFlags(user *object.User) string {
	if user.IsForbidden || user.IsDeleted {
		return sambaAcctFlagsDisabled
	}
	return sambaAcctFlagsNormal
}

// getSambaPwdLastSet never returns 0, which Samba reads as "must change password at next logon".
func getSambaPwdLastSet(user *object.User) string {
	timestamp := user.LastChangePasswordTime
	if timestamp == "" {
		timestamp = user.CreatedTime
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil || t.Unix() <= 0 {
		return "1"
	}
	return strconv.FormatInt(t.Unix(), 10)
}

// getSambaNtPassword can only derive the hash when the organization stores passwords
// in plain text, otherwise it has to come from the user's "sambaNTPassword" property.
func getSambaNtPassword(user *object.User, org *object.Organization) string {
	if ntPassword := getUserProperty(user, "sambaNTPassword"); ntPassword != "" {
		return strings.ToUpper(ntPassword)
	}
	if user.Password == "" || org == nil || (org.PasswordType != "" && org.PasswordType != "plain") {
		return ""
	}
	return getNtHash(user.Password)
}

// getNtHash is the MD4 digest of the UTF-16LE encoded password.
func getNtHash(password string) string {
	h := md4.New()
	for _, unit := range utf16.Encode([]rune(password)) {
		h.Write([]byte{byte(unit), byte(unit >> 8)})
	}
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func getUserProperty(user *object.User, key string) string {
	if user.Properties == nil {
		return ""
	}
	return user.Properties[key]
}
