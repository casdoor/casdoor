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
	"sync"

	"github.com/casdoor/casdoor/util"
)

// databaseProvisioningTarget writes the accounts into the user table of a legacy system that has no provisioning API,
// through the connection and the column mapping of a database syncer. The syncer may stay disabled, it is then only
// the description of the table and is never run.
type databaseProvisioningTarget struct {
	application *Application
	syncer      *Syncer
}

var (
	provisioningSyncers   = map[string]*Syncer{}
	provisioningSyncersMu sync.Mutex
)

func getProvisioningSyncer(application *Application, config *ProvisioningConfig) (*Syncer, error) {
	syncer, err := GetSyncer(config.Syncer)
	if err != nil {
		return nil, err
	}
	if syncer == nil {
		return nil, fmt.Errorf("the syncer: %s does not exist", config.Syncer)
	}
	if syncer.Organization != application.Organization {
		return nil, fmt.Errorf("the syncer: %s is not of the organization: %s", syncer.GetId(), application.Organization)
	}
	if !syncer.isDatabaseSyncer() {
		return nil, fmt.Errorf("the syncer: %s is not a database syncer", syncer.GetId())
	}
	if len(syncer.TableColumns) == 0 {
		return nil, fmt.Errorf("the syncer: %s has no table columns", syncer.GetId())
	}
	return syncer, nil
}

// getConnectedSyncer keeps one connection per syncer and opens a new one when the syncer is changed
func getConnectedSyncer(syncer *Syncer) (*Syncer, error) {
	provisioningSyncersMu.Lock()
	defer provisioningSyncersMu.Unlock()

	key := syncer.GetId()
	settings := getProvisioningHash(syncer)
	if cached, ok := provisioningSyncers[key+"|"+settings]; ok {
		return cached, nil
	}
	for cachedKey, cached := range provisioningSyncers {
		if len(cachedKey) > len(key) && cachedKey[:len(key)+1] == key+"|" {
			_ = cached.Close()
			delete(provisioningSyncers, cachedKey)
		}
	}

	err := GetSyncerProvider(syncer).InitAdapter()
	if err != nil {
		return nil, err
	}
	provisioningSyncers[key+"|"+settings] = syncer
	return syncer, nil
}

func newDatabaseProvisioningTarget(application *Application) (*databaseProvisioningTarget, error) {
	syncer, err := getProvisioningSyncer(application, application.Provisioning)
	if err != nil {
		return nil, err
	}
	return &databaseProvisioningTarget{application: application, syncer: syncer}, nil
}

func (syncer *Syncer) getKeyColumnForProvisioning() *TableColumn {
	if len(syncer.TableColumns) == 0 {
		return nil
	}
	return syncer.getKeyColumn()
}

func (syncer *Syncer) getForbiddenColumn() *TableColumn {
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.CasdoorName == "IsForbidden" {
			return tableColumn
		}
	}
	return nil
}

func (t *databaseProvisioningTarget) connect() (*Syncer, string, error) {
	syncer, err := getConnectedSyncer(t.syncer)
	if err != nil {
		return nil, "", err
	}
	key := syncer.getTargetTablePrimaryKey()
	if !util.FilterSQLIdentifier(key) {
		return nil, "", fmt.Errorf("invalid key column name: %s", key)
	}
	if !util.FilterSQLIdentifier(syncer.Table) {
		return nil, "", fmt.Errorf("invalid table name: %s", syncer.Table)
	}
	return syncer, key, nil
}

func (t *databaseProvisioningTarget) payload(user *User) (interface{}, error) {
	return t.syncer.getMapFromOriginalUser(user), nil
}

func (t *databaseProvisioningTarget) upsert(user *User, account *ProvisionedAccount) (string, error) {
	syncer, key, err := t.connect()
	if err != nil {
		return "", err
	}

	row := syncer.getMapFromOriginalUser(user)
	keyValue := row[key]
	if keyValue == "" {
		return "", fmt.Errorf("the key column: %s of the user: %s is empty", key, user.GetId())
	}

	// a renamed user keeps its row, found by the key it was written with
	lookupValue := keyValue
	if account.RemoteId != "" {
		lookupValue = account.RemoteId
	}
	count, err := syncer.Ormer.Engine.Table(syncer.getTable()).Where(fmt.Sprintf("%s = ?", key), lookupValue).Count()
	if err != nil {
		return "", err
	}
	if count == 0 && lookupValue != keyValue {
		lookupValue = keyValue
		count, err = syncer.Ormer.Engine.Table(syncer.getTable()).Where(fmt.Sprintf("%s = ?", key), lookupValue).Count()
		if err != nil {
			return "", err
		}
	}

	if count == 0 {
		_, err = syncer.Ormer.Engine.Table(syncer.getTable()).Insert(row)
	} else {
		_, err = syncer.Ormer.Engine.Table(syncer.getTable()).Where(fmt.Sprintf("%s = ?", key), lookupValue).Update(&row)
	}
	if err != nil {
		return "", err
	}
	return keyValue, nil
}

func (t *databaseProvisioningTarget) disable(user *User, account *ProvisionedAccount) error {
	syncer, key, err := t.connect()
	if err != nil {
		return err
	}
	column := syncer.getForbiddenColumn()
	if column == nil {
		return fmt.Errorf("no column of the syncer: %s is mapped to IsForbidden", syncer.GetId())
	}
	if !util.FilterSQLIdentifier(column.Name) {
		return fmt.Errorf("invalid column name: %s", column.Name)
	}

	row := syncer.getMapFromOriginalUser(&User{IsForbidden: true})
	_, err = syncer.Ormer.Engine.Table(syncer.getTable()).Where(fmt.Sprintf("%s = ?", key), account.RemoteId).
		Update(map[string]interface{}{column.Name: row[column.Name]})
	return err
}

func (t *databaseProvisioningTarget) remove(account *ProvisionedAccount) error {
	syncer, key, err := t.connect()
	if err != nil {
		return err
	}
	_, err = syncer.Ormer.Engine.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = ?", syncer.getTable(), key), account.RemoteId)
	return err
}

func (t *databaseProvisioningTarget) test() error {
	syncer, _, err := t.connect()
	if err != nil {
		return err
	}
	_, err = syncer.Ormer.Engine.Table(syncer.getTable()).Count()
	return err
}
