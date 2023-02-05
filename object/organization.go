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
	"strings"

	"github.com/casdoor/casdoor/cred"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type AccountItem struct {
	Name       string `json:"name"`
	Visible    bool   `json:"visible"`
	ViewRule   string `json:"viewRule"`
	ModifyRule string `json:"modifyRule"`
}

type Organization struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName        string   `xorm:"varchar(100)" json:"displayName"`
	WebsiteUrl         string   `xorm:"varchar(100)" json:"websiteUrl"`
	Favicon            string   `xorm:"varchar(100)" json:"favicon"`
	PasswordType       string   `xorm:"varchar(100)" json:"passwordType"`
	PasswordSalt       string   `xorm:"varchar(100)" json:"passwordSalt"`
	PhonePrefix        []string `xorm:"varchar(10)"  json:"phonePrefix"`
	DefaultAvatar      string   `xorm:"varchar(100)" json:"defaultAvatar"`
	DefaultApplication string   `xorm:"varchar(100)" json:"defaultApplication"`
	Tags               []string `xorm:"mediumtext" json:"tags"`
	Languages          []string `xorm:"varchar(255)" json:"languages"`
	MasterPassword     string   `xorm:"varchar(100)" json:"masterPassword"`
	InitScore          int      `json:"initScore"`
	EnableSoftDeletion bool     `json:"enableSoftDeletion"`
	IsProfilePublic    bool     `json:"isProfilePublic"`

	AccountItems []*AccountItem `xorm:"varchar(3000)" json:"accountItems"`
}

func GetOrganizationCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Organization{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetOrganizations(owner string) []*Organization {
	organizations := []*Organization{}
	err := adapter.Engine.Desc("created_time").Find(&organizations, &Organization{Owner: owner})
	if err != nil {
		panic(err)
	}

	return organizations
}

func GetPaginationOrganizations(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Organization {
	organizations := []*Organization{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&organizations)
	if err != nil {
		panic(err)
	}

	return organizations
}

func getOrganization(owner string, name string) *Organization {
	if owner == "" || name == "" {
		return nil
	}

	organization := Organization{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&organization)
	if err != nil {
		panic(err)
	}

	if existed {
		return &organization
	}

	return nil
}

func GetOrganization(id string) *Organization {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getOrganization(owner, name)
}

func GetMaskedOrganization(organization *Organization) *Organization {
	if organization == nil {
		return nil
	}

	if organization.MasterPassword != "" {
		organization.MasterPassword = "***"
	}
	return organization
}

func GetMaskedOrganizations(organizations []*Organization) []*Organization {
	for _, organization := range organizations {
		organization = GetMaskedOrganization(organization)
	}
	return organizations
}

func UpdateOrganization(id string, organization *Organization) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getOrganization(owner, name) == nil {
		return false
	}

	if name == "built-in" {
		organization.Name = name
	}

	if name != organization.Name {
		err := organizationChangeTrigger(name, organization.Name)
		if err != nil {
			return false
		}
	}

	if organization.MasterPassword != "" && organization.MasterPassword != "***" {
		credManager := cred.GetCredManager(organization.PasswordType)
		if credManager != nil {
			hashedPassword := credManager.GetHashedPassword(organization.MasterPassword, "", organization.PasswordSalt)
			organization.MasterPassword = hashedPassword
		}
	}

	session := adapter.Engine.ID(core.PK{owner, name}).AllCols()
	if organization.MasterPassword == "***" {
		session.Omit("master_password")
	}
	affected, err := session.Update(organization)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddOrganization(organization *Organization) bool {
	affected, err := adapter.Engine.Insert(organization)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteOrganization(organization *Organization) bool {
	if organization.Name == "built-in" {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{organization.Owner, organization.Name}).Delete(&Organization{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetOrganizationByUser(user *User) *Organization {
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

func CheckAccountItemModifyRule(accountItem *AccountItem, user *User, lang string) (bool, string) {
	switch accountItem.ModifyRule {
	case "Admin":
		if user == nil || !user.IsAdmin && !user.IsGlobalAdmin {
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
	organization := GetOrganization(id)
	if organization == nil {
		return nil, fmt.Errorf("The organization: %s does not exist", id)
	}

	if organization.DefaultApplication != "" {
		defaultApplication := getApplication("admin", organization.DefaultApplication)
		if defaultApplication == nil {
			return nil, fmt.Errorf("The default application: %s does not exist", organization.DefaultApplication)
		} else {
			return defaultApplication, nil
		}
	}

	applications := []*Application{}
	err := adapter.Engine.Asc("created_time").Find(&applications, &Application{Organization: organization.Name})
	if err != nil {
		panic(err)
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

	extendApplicationWithProviders(defaultApplication)
	extendApplicationWithOrg(defaultApplication)

	return defaultApplication, nil
}

func organizationChangeTrigger(oldName string, newName string) error {
	session := adapter.Engine.NewSession()
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

	role := new(Role)
	_, err = adapter.Engine.Where("owner=?", oldName).Get(role)
	if err != nil {
		return err
	}
	for i, u := range role.Users {
		// u = organization/username
		split := strings.Split(u, "/")
		if split[0] == oldName {
			split[0] = newName
			role.Users[i] = split[0] + "/" + split[1]
		}
	}
	for i, u := range role.Roles {
		// u = organization/username
		split := strings.Split(u, "/")
		if split[0] == oldName {
			split[0] = newName
			role.Roles[i] = split[0] + "/" + split[1]
		}
	}
	role.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(role)
	if err != nil {
		return err
	}

	permission := new(Permission)
	_, err = adapter.Engine.Where("owner=?", oldName).Get(permission)
	if err != nil {
		return err
	}
	for i, u := range permission.Users {
		// u = organization/username
		split := strings.Split(u, "/")
		if split[0] == oldName {
			split[0] = newName
			permission.Users[i] = split[0] + "/" + split[1]
		}
	}
	for i, u := range permission.Roles {
		// u = organization/username
		split := strings.Split(u, "/")
		if split[0] == oldName {
			split[0] = newName
			permission.Roles[i] = split[0] + "/" + split[1]
		}
	}
	permission.Owner = newName
	_, err = session.Where("owner=?", oldName).Update(permission)
	if err != nil {
		return err
	}

	casbinAdapter := new(CasbinAdapter)
	casbinAdapter.Owner = newName
	casbinAdapter.Organization = newName
	_, err = session.Where("owner=?", oldName).Update(casbinAdapter)
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
	payment.Organization = newName
	_, err = session.Where("organization=?", oldName).Update(payment)
	if err != nil {
		return err
	}

	record := new(Record)
	record.Owner = newName
	record.Organization = newName
	_, err = session.Where("organization=?", oldName).Update(record)
	if err != nil {
		return err
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
