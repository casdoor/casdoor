// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"strconv"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/cred"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/builder"
	"github.com/xorm-io/core"
)

type AccountItem struct {
	Name       string `json:"name"`
	Visible    bool   `json:"visible"`
	ViewRule   string `json:"viewRule"`
	ModifyRule string `json:"modifyRule"`
	Regex      string `json:"regex"`
}

type ThemeData struct {
	ThemeType    string `xorm:"varchar(30)" json:"themeType"`
	ColorPrimary string `xorm:"varchar(10)" json:"colorPrimary"`
	BorderRadius int    `xorm:"int" json:"borderRadius"`
	IsCompact    bool   `xorm:"bool" json:"isCompact"`
	IsEnabled    bool   `xorm:"bool" json:"isEnabled"`
}

type MfaItem struct {
	Name string `json:"name"`
	Rule string `json:"rule"`
}

type Organization struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName            string     `xorm:"varchar(100)" json:"displayName"`
	WebsiteUrl             string     `xorm:"varchar(100)" json:"websiteUrl"`
	Logo                   string     `xorm:"varchar(200)" json:"logo"`
	LogoDark               string     `xorm:"varchar(200)" json:"logoDark"`
	Favicon                string     `xorm:"varchar(200)" json:"favicon"`
	PasswordType           string     `xorm:"varchar(100)" json:"passwordType"`
	PasswordSalt           string     `xorm:"varchar(100)" json:"passwordSalt"`
	PasswordOptions        []string   `xorm:"varchar(100)" json:"passwordOptions"`
	PasswordObfuscatorType string     `xorm:"varchar(100)" json:"passwordObfuscatorType"`
	PasswordObfuscatorKey  string     `xorm:"varchar(100)" json:"passwordObfuscatorKey"`
	PasswordExpireDays     int        `json:"passwordExpireDays"`
	CountryCodes           []string   `xorm:"mediumtext"  json:"countryCodes"`
	DefaultAvatar          string     `xorm:"varchar(200)" json:"defaultAvatar"`
	DefaultApplication     string     `xorm:"varchar(100)" json:"defaultApplication"`
	UserTypes              []string   `xorm:"mediumtext" json:"userTypes"`
	Tags                   []string   `xorm:"mediumtext" json:"tags"`
	Languages              []string   `xorm:"varchar(255)" json:"languages"`
	ThemeData              *ThemeData `xorm:"json" json:"themeData"`
	MasterPassword         string     `xorm:"varchar(200)" json:"masterPassword"`
	DefaultPassword        string     `xorm:"varchar(200)" json:"defaultPassword"`
	MasterVerificationCode string     `xorm:"varchar(100)" json:"masterVerificationCode"`
	IpWhitelist            string     `xorm:"varchar(200)" json:"ipWhitelist"`
	InitScore              int        `json:"initScore"`
	EnableSoftDeletion     bool       `json:"enableSoftDeletion"`
	IsProfilePublic        bool       `json:"isProfilePublic"`
	UseEmailAsUsername     bool       `json:"useEmailAsUsername"`
	EnableTour             bool       `json:"enableTour"`
	IpRestriction          string     `json:"ipRestriction"`
	NavItems               []string   `xorm:"varchar(1000)" json:"navItems"`
	WidgetItems            []string   `xorm:"varchar(1000)" json:"widgetItems"`

	MfaItems     []*MfaItem     `xorm:"varchar(300)" json:"mfaItems"`
	AccountItems []*AccountItem `xorm:"varchar(5000)" json:"accountItems"`
}

func GetOrganizationCount(owner, name, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Organization{Name: name})
}

func GetOrganizations(owner string, name ...string) ([]*Organization, error) {
	organizations := []*Organization{}
	if name != nil && len(name) > 0 {
		err := ormer.Engine.Desc("created_time").Where(builder.In("name", name)).Find(&organizations)
		if err != nil {
			return nil, err
		}
	} else {
		err := ormer.Engine.Desc("created_time").Find(&organizations, &Organization{Owner: owner})
		if err != nil {
			return nil, err
		}
	}

	return organizations, nil
}

func GetOrganizationsByFields(owner string, fields ...string) ([]*Organization, error) {
	organizations := []*Organization{}
	err := ormer.Engine.Desc("created_time").Cols(fields...).Find(&organizations, &Organization{Owner: owner})
	if err != nil {
		return nil, err
	}

	return organizations, nil
}

func GetPaginationOrganizations(owner string, name string, offset, limit int, field, value, sortField, sortOrder string) ([]*Organization, error) {
	organizations := []*Organization{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	var err error
	if name != "" {
		err = session.Find(&organizations, &Organization{Name: name})
	} else {
		err = session.Find(&organizations)
	}
	if err != nil {
		return nil, err
	}

	return organizations, nil
}

func getOrganization(owner string, name string) (*Organization, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	organization := Organization{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&organization)
	if err != nil {
		return nil, err
	}

	if existed {
		return &organization, nil
	}

	return nil, nil
}

func GetOrganization(id string) (*Organization, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getOrganization(owner, name)
}

func GetMaskedOrganization(organization *Organization, errs ...error) (*Organization, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	if organization == nil {
		return nil, nil
	}

	if organization.MasterPassword != "" {
		organization.MasterPassword = "***"
	}
	if organization.DefaultPassword != "" {
		organization.DefaultPassword = "***"
	}
	if organization.MasterVerificationCode != "" {
		organization.MasterVerificationCode = "***"
	}
	return organization, nil
}

func GetMaskedOrganizations(organizations []*Organization, errs ...error) ([]*Organization, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	var err error
	for _, organization := range organizations {
		organization, err = GetMaskedOrganization(organization)
		if err != nil {
			return nil, err
		}
	}

	return organizations, nil
}

func UpdateOrganization(id string, organization *Organization, isGlobalAdmin bool) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	org, err := getOrganization(owner, name)
	if err != nil {
		return false, err
	} else if org == nil {
		return false, nil
	}

	if name == "built-in" {
		organization.Name = name
	}

	if name != organization.Name {
		err := organizationChangeTrigger(name, organization.Name)
		if err != nil {
			return false, err
		}
	}

	if organization.MasterPassword != "" && organization.MasterPassword != "***" {
		credManager := cred.GetCredManager(organization.PasswordType)
		if credManager != nil {
			hashedPassword := credManager.GetHashedPassword(organization.MasterPassword, "", organization.PasswordSalt)
			organization.MasterPassword = hashedPassword
		}
	}

	if !isGlobalAdmin {
		organization.NavItems = org.NavItems
		organization.WidgetItems = org.WidgetItems
	}

	session := ormer.Engine.ID(core.PK{owner, name}).AllCols()

	if organization.MasterPassword == "***" {
		session.Omit("master_password")
	}
	if organization.DefaultPassword == "***" {
		session.Omit("default_password")
	}
	if organization.MasterVerificationCode == "***" {
		session.Omit("master_verification_code")
	}

	affected, err := session.Update(organization)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddOrganization(organization *Organization) (bool, error) {
	affected, err := ormer.Engine.Insert(organization)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func deleteOrganization(organization *Organization) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{organization.Owner, organization.Name}).Delete(&Organization{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteOrganization(organization *Organization) (bool, error) {
	if organization.Name == "built-in" {
		return false, nil
	}

	return deleteOrganization(organization)
}

func GetOrganizationByUser(user *User) (*Organization, error) {
	if user == nil {
		return nil, nil
	}

	return getOrganization("admin", user.Owner)
}

func GetAccountItemByName(name string, organization *Organization) *AccountItem {
	if organization == nil {
		return nil
	}
	for _, accountItem := range organization.AccountItems {
		if accountItem.Name == name {
			return accountItem
		}
	}
	return nil
}

func CheckAccountItemModifyRule(accountItem *AccountItem, isAdmin bool, lang string) (bool, string) {
	if accountItem == nil {
		return true, ""
	}

	switch accountItem.ModifyRule {
	case "Admin":
		if !isAdmin {
			return false, fmt.Sprintf(i18n.Translate(lang, "organization:Only admin can modify the %s."), accountItem.Name)
		}
	case "Immutable":
		return false, fmt.Sprintf(i18n.Translate(lang, "organization:The %s is immutable."), accountItem.Name)
	case "Self":
		break
	default:
		return false, fmt.Sprintf(i18n.Translate(lang, "organization:Unknown modify rule %s."), accountItem.ModifyRule)
	}
	return true, ""
}

func GetDefaultApplication(id string) (*Application, error) {
	organization, err := GetOrganization(id)
	if err != nil {
		return nil, err
	}

	if organization == nil {
		return nil, fmt.Errorf("The organization: %s does not exist", id)
	}

	if organization.DefaultApplication != "" {
		defaultApplication, err := getApplication("admin", organization.DefaultApplication)
		if err != nil {
			return nil, err
		}

		if defaultApplication == nil {
			return nil, fmt.Errorf("The default application: %s does not exist", organization.DefaultApplication)
		} else {
			defaultApplication.Organization = organization.Name
			return defaultApplication, nil
		}
	}

	applications := []*Application{}
	err = ormer.Engine.Asc("created_time").Find(&applications, &Application{Organization: organization.Name})
	if err != nil {
		return nil, err
	}

	if len(applications) == 0 {
		return nil, fmt.Errorf("The application does not exist")
	}

	defaultApplication := applications[0]
	for _, application := range applications {
		if application.EnableSignUp {
			defaultApplication = application
			break
		}
	}

	err = extendApplicationWithProviders(defaultApplication)
	if err != nil {
		return nil, err
	}

	err = extendApplicationWithOrg(defaultApplication)
	if err != nil {
		return nil, err
	}

	err = extendApplicationWithSigninItems(defaultApplication)
	if err != nil {
		return nil, err
	}

	err = extendApplicationWithSigninMethods(defaultApplication)
	if err != nil {
		return nil, err
	}

	return defaultApplication, nil
}

func organizationChangeTrigger(oldName string, newName string) error {
	session := ormer.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	application := new(Application)
	application.Organization = newName
	_, err = session.Where("organization=?", oldName).Update(application)
	if err != nil {
		return err
	}

	user := new(User)
	user.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(user)
	if err != nil {
		return err
	}

	group := new(Group)
	group.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(group)
	if err != nil {
		return err
	}

	role := new(Role)
	_, err = ormer.Engine.Where("owner=?", oldName).Get(role)
	if err != nil {
		return err
	}
	for i, u := range role.Users {
		// u = organization/username
		owner, name := util.GetOwnerAndNameFromId(u)
		if name == oldName {
			role.Users[i] = util.GetId(owner, newName)
		}
	}
	for i, u := range role.Roles {
		// u = organization/username
		owner, name := util.GetOwnerAndNameFromId(u)
		if name == oldName {
			role.Roles[i] = util.GetId(owner, newName)
		}
	}
	role.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(role)
	if err != nil {
		return err
	}

	permission := new(Permission)
	_, err = ormer.Engine.Where("owner=?", oldName).Get(permission)
	if err != nil {
		return err
	}
	for i, u := range permission.Users {
		// u = organization/username
		owner, name := util.GetOwnerAndNameFromId(u)
		if name == oldName {
			permission.Users[i] = util.GetId(owner, newName)
		}
	}
	for i, u := range permission.Roles {
		// u = organization/username
		owner, name := util.GetOwnerAndNameFromId(u)
		if name == oldName {
			permission.Roles[i] = util.GetId(owner, newName)
		}
	}
	permission.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(permission)
	if err != nil {
		return err
	}

	adapter := new(Adapter)
	adapter.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(adapter)
	if err != nil {
		return err
	}

	ldap := new(Ldap)
	ldap.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(ldap)
	if err != nil {
		return err
	}

	model := new(Model)
	model.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(model)
	if err != nil {
		return err
	}

	payment := new(Payment)
	payment.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(payment)
	if err != nil {
		return err
	}

	record := new(Record)
	record.Owner = newName
	record.Organization = newName
	_, err = session.Where("organization=?", oldName).Update(record)
	if err != nil {
		if err.Error() != "no columns found to be updated" {
			return err
		}
	}

	resource := new(Resource)
	resource.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(resource)
	if err != nil {
		return err
	}

	syncer := new(Syncer)
	syncer.Organization = newName
	_, err = session.Where("organization=?", oldName).Update(syncer)
	if err != nil {
		return err
	}

	token := new(Token)
	token.Organization = newName
	_, err = session.Where("organization=?", oldName).Update(token)
	if err != nil {
		return err
	}

	webhook := new(Webhook)
	webhook.Organization = newName
	_, err = session.Where("organization=?", oldName).Update(webhook)
	if err != nil {
		return err
	}

	return session.Commit()
}

func IsNeedPromptMfa(org *Organization, user *User) bool {
	if org == nil || user == nil {
		return false
	}
	for _, item := range org.MfaItems {
		if item.Rule == "Required" {
			if item.Name == EmailType && !user.MfaEmailEnabled {
				return true
			}
			if item.Name == SmsType && !user.MfaPhoneEnabled {
				return true
			}
			if item.Name == TotpType && user.TotpSecret == "" {
				return true
			}
		}
	}
	return false
}

func (org *Organization) GetInitScore() (int, error) {
	if org != nil {
		return org.InitScore, nil
	} else {
		return strconv.Atoi(conf.GetConfigString("initScore"))
	}
}
