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
	"errors"
	"fmt"
	"regexp"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// The kinds of privileged account. "Role" is a role of the organization that users are put into for a while (just-in-time
// elevation), the others are shared admin accounts of target systems.
const (
	PamTypeLinux      = "Linux"
	PamTypeWindows    = "Windows"
	PamTypeMysql      = "MySQL"
	PamTypePostgresql = "PostgreSQL"
	PamTypeLdap       = "LDAP"
	PamTypeAd         = "Active Directory"
	PamTypeOther      = "Other"
	PamTypeRole       = "Role"
)

var pamTypes = []string{PamTypeLinux, PamTypeWindows, PamTypeMysql, PamTypePostgresql, PamTypeLdap, PamTypeAd, PamTypeOther, PamTypeRole}

const maskedSecret = "***"

// PrivilegedAccount is a shared admin account of a target system kept in the vault, or a role that users can be
// elevated into for a while. The secrets are kept SM4-encrypted and are only handed out to the requester of an
// active access request.
type PrivilegedAccount struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(1000)" json:"description"`

	Type          string `xorm:"varchar(100)" json:"type"`
	Host          string `xorm:"varchar(200)" json:"host"`
	Port          int    `json:"port"`
	Database      string `xorm:"varchar(100)" json:"database"`
	Username      string `xorm:"varchar(200)" json:"username"`
	Secret        string `xorm:"mediumtext" json:"secret"`
	AdminUsername string `xorm:"varchar(200)" json:"adminUsername"`
	AdminSecret   string `xorm:"mediumtext" json:"adminSecret"`
	// the new password of a rotation that may or may not have reached the target system
	PendingSecret string `xorm:"mediumtext" json:"-"`
	// the SSH host key fingerprint, pinned on the first connection when empty
	HostKey       string `xorm:"varchar(200)" json:"hostKey"`
	SkipTlsVerify bool   `json:"skipTlsVerify"`
	Role          string `xorm:"varchar(100)" json:"role"`

	// the users who may request access, everyone of the organization when empty
	Users              []string `xorm:"mediumtext" json:"users"`
	Approvers          []string `xorm:"mediumtext" json:"approvers"`
	RequireApproval    bool     `json:"requireApproval"`
	MaxCheckoutMinutes int      `json:"maxCheckoutMinutes"`
	RotateAfterCheckin bool     `json:"rotateAfterCheckin"`
	RotateIntervalDays int      `json:"rotateIntervalDays"`
	PasswordLength     int      `json:"passwordLength"`
	PasswordCharset    string   `xorm:"varchar(100)" json:"passwordCharset"`

	NeedsRotation   bool   `json:"needsRotation"`
	LastRotatedTime string `xorm:"varchar(100)" json:"lastRotatedTime"`
	RotationTime    string `xorm:"varchar(100)" json:"rotationTime"`
	RotationState   string `xorm:"varchar(100)" json:"rotationState"`
	RotationMessage string `xorm:"varchar(1000)" json:"rotationMessage"`

	IsEnabled bool `json:"isEnabled"`
}

// the columns an admin edits, the rotation state and the pending secret are only written by the rotation
var privilegedAccountEditColumns = []string{
	"name", "updated_time", "display_name", "description", "type", "host", "port", "database", "username", "secret",
	"admin_username", "admin_secret", "host_key", "skip_tls_verify", "role", "users", "approvers", "require_approval",
	"max_checkout_minutes", "rotate_after_checkin", "rotate_interval_days", "password_length", "password_charset", "is_enabled",
}

func (account *PrivilegedAccount) GetId() string {
	return fmt.Sprintf("%s/%s", account.Owner, account.Name)
}

func (account *PrivilegedAccount) isRole() bool {
	return account.Type == PamTypeRole
}

// canRotate tells whether Casdoor can change the account's password on the target system
func (account *PrivilegedAccount) canRotate() bool {
	switch account.Type {
	case PamTypeLinux, PamTypeMysql, PamTypePostgresql, PamTypeLdap, PamTypeAd:
		return true
	}
	return false
}

func (account *PrivilegedAccount) getMaxCheckoutMinutes() int {
	if account.MaxCheckoutMinutes <= 0 {
		return 60
	}
	return account.MaxCheckoutMinutes
}

func (account *PrivilegedAccount) isRequester(userName string) bool {
	return len(account.Users) == 0 || util.InSlice(account.Users, userName)
}

// isApprover tells whether the user decides the account's requests: a listed approver, or an admin of the
// organization when no approver is listed
func (account *PrivilegedAccount) isApprover(user *User) bool {
	if user == nil || user.Owner != account.Owner {
		return false
	}
	if len(account.Approvers) > 0 {
		return util.InSlice(account.Approvers, user.Name)
	}
	return user.IsAdmin
}

func GetPrivilegedAccountCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&PrivilegedAccount{})
}

func GetPrivilegedAccounts(owner string) ([]*PrivilegedAccount, error) {
	accounts := []*PrivilegedAccount{}
	err := ormer.Engine.Desc("created_time").Find(&accounts, &PrivilegedAccount{Owner: owner})
	return accounts, err
}

func GetPaginationPrivilegedAccounts(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*PrivilegedAccount, error) {
	accounts := []*PrivilegedAccount{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&accounts)
	return accounts, err
}

func getPrivilegedAccount(owner string, name string) (*PrivilegedAccount, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	account := PrivilegedAccount{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&account)
	if err != nil {
		return nil, err
	}
	if existed {
		return &account, nil
	}
	return nil, nil
}

func GetPrivilegedAccount(id string) (*PrivilegedAccount, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getPrivilegedAccount(owner, name)
}

// GetMaskedPrivilegedAccount hides the secrets, no API hands them out except the reveal of an active request
func GetMaskedPrivilegedAccount(account *PrivilegedAccount) *PrivilegedAccount {
	if account == nil {
		return nil
	}
	if account.Secret != "" {
		account.Secret = maskedSecret
	}
	if account.AdminSecret != "" {
		account.AdminSecret = maskedSecret
	}
	account.PendingSecret = ""
	return account
}

func GetMaskedPrivilegedAccounts(accounts []*PrivilegedAccount) []*PrivilegedAccount {
	for _, account := range accounts {
		GetMaskedPrivilegedAccount(account)
	}
	return accounts
}

var pamLinuxUsernameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.-]*\$?$`)

// ValidatePrivilegedAccount checks an account an admin saves, before its secrets get encrypted
func ValidatePrivilegedAccount(account *PrivilegedAccount, lang string) error {
	if account.Name == "" {
		return errors.New(i18n.Translate(lang, "pam:The name must not be empty"))
	}
	organization, err := getOrganization("admin", account.Owner)
	if err != nil {
		return err
	}
	if organization == nil {
		return fmt.Errorf(i18n.Translate(lang, "auth:The organization: %s does not exist"), account.Owner)
	}
	if !util.InSlice(pamTypes, account.Type) {
		return fmt.Errorf(i18n.Translate(lang, "pam:Unknown privileged account type: %s"), account.Type)
	}
	if account.MaxCheckoutMinutes < 0 || account.RotateIntervalDays < 0 || account.Port < 0 || account.Port > 65535 {
		return errors.New(i18n.Translate(lang, "pam:The numbers of the privileged account must not be negative"))
	}
	if account.PasswordLength != 0 && (account.PasswordLength < 12 || account.PasswordLength > 128) {
		return errors.New(i18n.Translate(lang, "pam:The generated password length must be between 12 and 128"))
	}

	if account.isRole() {
		role, err := getRole(account.Owner, account.Role)
		if err != nil {
			return err
		}
		if role == nil {
			return fmt.Errorf(i18n.Translate(lang, "pam:The role: %s does not exist in the organization"), account.Role)
		}
		account.Host, account.Port, account.Username, account.Secret, account.AdminUsername, account.AdminSecret = "", 0, "", "", "", ""
		account.RotateAfterCheckin, account.RotateIntervalDays = false, 0
	} else {
		account.Role = ""
		if account.Username == "" {
			return errors.New(i18n.Translate(lang, "pam:The account username must not be empty"))
		}
	}

	if (account.RotateAfterCheckin || account.RotateIntervalDays > 0) && !account.canRotate() {
		return fmt.Errorf(i18n.Translate(lang, "pam:Passwords of the type: %s cannot be rotated by Casdoor"), account.Type)
	}
	if account.canRotate() && (account.RotateAfterCheckin || account.RotateIntervalDays > 0) && account.Host == "" {
		return errors.New(i18n.Translate(lang, "pam:The host must not be empty to rotate the password"))
	}
	if account.Type == PamTypeLinux {
		if !pamLinuxUsernameRegex.MatchString(account.Username) || (account.AdminUsername != "" && !pamLinuxUsernameRegex.MatchString(account.AdminUsername)) {
			return errors.New(i18n.Translate(lang, "pam:The Linux username is invalid"))
		}
	}

	for _, name := range append(append([]string{}, account.Approvers...), account.Users...) {
		user, err := getUser(account.Owner, name)
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf(i18n.Translate(lang, "general:The user: %s doesn't exist"), util.GetId(account.Owner, name))
		}
	}

	if (account.Secret != "" && account.Secret != maskedSecret) || (account.AdminSecret != "" && account.AdminSecret != maskedSecret) {
		if !IsFieldEncryptionKeyConfigured() {
			return errors.New(i18n.Translate(lang, "pam:Set the configuration fieldEncryptionKey before saving privileged account passwords"))
		}
	}
	return nil
}

// encryptPrivilegedSecrets encrypts the secrets of an account going into the database, a masked one keeps the old value
func encryptPrivilegedSecrets(account *PrivilegedAccount, oldAccount *PrivilegedAccount) error {
	var err error
	if account.Secret == maskedSecret {
		account.Secret = ""
		if oldAccount != nil {
			account.Secret = oldAccount.Secret
		}
	}
	if account.AdminSecret == maskedSecret {
		account.AdminSecret = ""
		if oldAccount != nil {
			account.AdminSecret = oldAccount.AdminSecret
		}
	}
	account.Secret, err = encryptField(account.Owner, "privilegedAccount.secret", account.Secret)
	if err != nil {
		return err
	}
	account.AdminSecret, err = encryptField(account.Owner, "privilegedAccount.adminSecret", account.AdminSecret)
	return err
}

func (account *PrivilegedAccount) getSecret() (string, error) {
	return decryptField(account.Owner, "privilegedAccount.secret", account.Secret)
}

func (account *PrivilegedAccount) getAdminSecret() (string, error) {
	return decryptField(account.Owner, "privilegedAccount.adminSecret", account.AdminSecret)
}

func (account *PrivilegedAccount) getPendingSecret() (string, error) {
	return decryptField(account.Owner, "privilegedAccount.secret", account.PendingSecret)
}

func AddPrivilegedAccount(account *PrivilegedAccount) (bool, error) {
	if err := encryptPrivilegedSecrets(account, nil); err != nil {
		return false, err
	}
	account.NeedsRotation, account.LastRotatedTime, account.RotationTime, account.RotationState, account.RotationMessage = false, "", "", "", ""
	affected, err := ormer.Engine.Insert(account)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func UpdatePrivilegedAccount(id string, account *PrivilegedAccount) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	oldAccount, err := getPrivilegedAccount(owner, name)
	if err != nil {
		return false, err
	}
	if oldAccount == nil {
		return false, nil
	}

	account.Owner = owner
	if err = encryptPrivilegedSecrets(account, oldAccount); err != nil {
		return false, err
	}
	account.UpdatedTime = util.GetCurrentTime()
	affected, err := ormer.Engine.ID(core.PK{owner, name}).Cols(privilegedAccountEditColumns...).Update(account)
	if err != nil {
		return false, err
	}
	if affected != 0 && name != account.Name {
		_, err = ormer.Engine.Where("owner = ? and account = ?", owner, name).Cols("account").Update(&PrivilegedRequest{Account: account.Name})
		if err != nil {
			return false, err
		}
	}
	return affected != 0, nil
}

func DeletePrivilegedAccount(account *PrivilegedAccount) (bool, error) {
	count, err := ormer.Engine.Where("owner = ? and account = ? and (state = ? or state = ?)", account.Owner, account.Name, PamStateActive, PamStateApproved).Count(&PrivilegedRequest{})
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, fmt.Errorf("the privileged account: %s has active requests, end them first", account.GetId())
	}
	affected, err := ormer.Engine.ID(core.PK{account.Owner, account.Name}).Delete(&PrivilegedAccount{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// PamAccountView is what a user who may request an account sees of it
type PamAccountView struct {
	Name               string   `json:"name"`
	DisplayName        string   `json:"displayName"`
	Description        string   `json:"description"`
	Type               string   `json:"type"`
	Host               string   `json:"host"`
	Port               int      `json:"port"`
	Username           string   `json:"username"`
	Role               string   `json:"role"`
	RequireApproval    bool     `json:"requireApproval"`
	MaxCheckoutMinutes int      `json:"maxCheckoutMinutes"`
	Approvers          []string `json:"approvers"`
}

func newPamAccountView(account *PrivilegedAccount) *PamAccountView {
	return &PamAccountView{
		Name:               account.Name,
		DisplayName:        account.DisplayName,
		Description:        account.Description,
		Type:               account.Type,
		Host:               account.Host,
		Port:               account.Port,
		Username:           account.Username,
		Role:               account.Role,
		RequireApproval:    account.RequireApproval,
		MaxCheckoutMinutes: account.getMaxCheckoutMinutes(),
		Approvers:          account.Approvers,
	}
}
