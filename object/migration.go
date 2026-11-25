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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// Migration moves the users, groups, applications and the rest of another identity system, such as Auth0 or
// Keycloak, into an organization through an installed action of kind "import" that reads the other system
type Migration struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	Plugin       string `xorm:"varchar(100)" json:"plugin"`
	Host         string `xorm:"varchar(200)" json:"host"`
	ClientId     string `xorm:"varchar(200)" json:"clientId"`
	ClientSecret string `xorm:"varchar(500)" json:"clientSecret"`
	Metadata     string `xorm:"mediumtext" json:"metadata"`

	// "", "Running", "Previewed", "Applied", "Undone" or "Failed"
	State  string           `xorm:"varchar(100)" json:"state"`
	Report *MigrationReport `xorm:"mediumtext json" json:"report"`
	// the objects the migration added, as "<kind>:<owner>/<name>", which undoing it deletes
	Objects []string `xorm:"longtext" json:"-"`
}

type MigrationKindReport struct {
	Kind    string `json:"kind"`
	Total   int    `json:"total"`
	New     int    `json:"new"`
	Exists  int    `json:"exists"`
	Invalid int    `json:"invalid"`
	Added   int    `json:"added"`
	Failed  int    `json:"failed"`
	Deleted int    `json:"deleted"`
}

type MigrationIssue struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type MigrationReport struct {
	// "test", "plan", "apply" or "undo"
	Mode         string                 `json:"mode"`
	StartedTime  string                 `json:"startedTime"`
	FinishedTime string                 `json:"finishedTime"`
	Error        string                 `json:"error"`
	Kinds        []*MigrationKindReport `json:"kinds"`
	Warnings     []string               `json:"warnings"`
	Issues       []*MigrationIssue      `json:"issues"`
}

func GetMigrationCount(owner, organization, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Migration{Organization: organization})
}

func GetOrganizationMigrations(owner, organization string) ([]*Migration, error) {
	migrations := []*Migration{}
	err := ormer.Engine.Desc("created_time").Omit("objects").Find(&migrations, &Migration{Owner: owner, Organization: organization})
	return migrations, err
}

func GetPaginationMigrations(owner, organization string, offset, limit int, field, value, sortField, sortOrder string) ([]*Migration, error) {
	migrations := []*Migration{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Omit("objects").Find(&migrations, &Migration{Organization: organization})
	return migrations, err
}

func getMigration(owner string, name string) (*Migration, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	migration := Migration{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&migration)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	return &migration, nil
}

func GetMigration(id string) (*Migration, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getMigration(owner, name)
}

func GetMigrationByOrganization(id string, organization string) (*Migration, error) {
	migration, err := GetMigration(id)
	if err != nil {
		return nil, err
	}
	if migration == nil || migration.Organization != organization {
		return nil, nil
	}
	return migration, nil
}

func GetMaskedMigration(migration *Migration, errs ...error) (*Migration, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}
	if migration == nil {
		return nil, nil
	}

	if migration.ClientSecret != "" {
		migration.ClientSecret = "***"
	}
	return migration, nil
}

func GetMaskedMigrations(migrations []*Migration, errs ...error) ([]*Migration, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}
	for _, migration := range migrations {
		_, _ = GetMaskedMigration(migration)
	}
	return migrations, nil
}

// the fields an admin edits, the state, report and added objects belong to the runs
var migrationSettingCols = []string{"name", "updated_time", "display_name", "organization", "plugin", "host", "client_id", "client_secret", "metadata"}

func UpdateMigration(id string, migration *Migration, isGlobalAdmin bool, lang string) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	m, err := getMigration(owner, name)
	if err != nil {
		return false, err
	}
	if m == nil {
		return false, nil
	}
	if !isGlobalAdmin && m.Organization != migration.Organization {
		return false, errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
	}
	if isMigrationRunning(m) {
		return false, fmt.Errorf("the migration: %s is running", m.Name)
	}
	// the added objects belong to the organization, moving the migration would undo them in the wrong one
	if len(m.Objects) > 0 && m.Organization != migration.Organization {
		return false, fmt.Errorf("undo the migration: %s before moving it to another organization", m.Name)
	}

	if migration.ClientSecret == "***" {
		migration.ClientSecret = m.ClientSecret
	}
	migration.UpdatedTime = util.GetCurrentTime()

	affected, err := ormer.Engine.ID(core.PK{owner, name}).Cols(migrationSettingCols...).Update(migration)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func AddMigration(migration *Migration) (bool, error) {
	if migration.CreatedTime == "" {
		migration.CreatedTime = util.GetCurrentTime()
	}
	migration.UpdatedTime = migration.CreatedTime
	migration.State = ""
	migration.Report = nil
	migration.Objects = nil

	affected, err := ormer.Engine.Insert(migration)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteMigration(migration *Migration) (bool, error) {
	m, err := getMigration(migration.Owner, migration.Name)
	if err != nil {
		return false, err
	}
	if m == nil {
		return false, nil
	}
	if isMigrationRunning(m) {
		return false, fmt.Errorf("the migration: %s is running", m.Name)
	}

	affected, err := ormer.Engine.ID(core.PK{migration.Owner, migration.Name}).Where("organization = ?", migration.Organization).Delete(&Migration{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (migration *Migration) GetId() string {
	return fmt.Sprintf("%s/%s", migration.Owner, migration.Name)
}

// GetImportActions returns the actions of kind "import" installed in the organization, which migrations read the
// other identity systems with
func GetImportActions(owner string) ([]*Integration, error) {
	integrations, err := GetIntegrations(owner)
	if err != nil {
		return nil, err
	}

	actions := []*Integration{}
	for _, integration := range integrations {
		if integration.Type != "action" {
			continue
		}
		kind := integration.Kind
		if kind == "" {
			bundle := &IntegrationBundle{}
			if json.Unmarshal([]byte(integration.Bundle), bundle) != nil {
				continue
			}
			file, err := getActionFile(bundle)
			if err != nil {
				continue
			}
			kind = file.Kind
		}
		if kind == "import" {
			actions = append(actions, integration)
		}
	}
	return actions, nil
}
