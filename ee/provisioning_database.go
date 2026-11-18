// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"fmt"
	"sync"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// databaseProvisioningTarget writes the accounts into the user table of a legacy system that has no provisioning API,
// through the connection and the column mapping of a database syncer. The syncer may stay disabled, it is then only
// the description of the table and is never run.
type databaseProvisioningTarget struct {
	application *object.Application
	syncer      *object.Syncer
}

var (
	provisioningSyncers   = map[string]*object.Syncer{}
	provisioningSyncersMu sync.Mutex
)

func getProvisioningSyncer(application *object.Application, config *object.ProvisioningConfig) (*object.Syncer, error) {
	syncer, err := object.GetSyncer(config.Syncer)
	if err != nil {
		return nil, err
	}
	if syncer == nil {
		return nil, fmt.Errorf("the syncer: %s does not exist", config.Syncer)
	}
	if syncer.Organization != application.Organization {
		return nil, fmt.Errorf("the syncer: %s is not of the organization: %s", syncer.GetId(), application.Organization)
	}
	if !syncer.IsDatabaseSyncer() {
		return nil, fmt.Errorf("the syncer: %s is not a database syncer", syncer.GetId())
	}
	if len(syncer.TableColumns) == 0 {
		return nil, fmt.Errorf("the syncer: %s has no table columns", syncer.GetId())
	}
	return syncer, nil
}

// getConnectedSyncer keeps one connection per syncer and opens a new one when the syncer is changed
func getConnectedSyncer(syncer *object.Syncer) (*object.Syncer, error) {
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

	err := object.GetSyncerProvider(syncer).InitAdapter()
	if err != nil {
		return nil, err
	}
	provisioningSyncers[key+"|"+settings] = syncer
	return syncer, nil
}

func newDatabaseProvisioningTarget(application *object.Application) (*databaseProvisioningTarget, error) {
	syncer, err := getProvisioningSyncer(application, application.Provisioning)
	if err != nil {
		return nil, err
	}
	return &databaseProvisioningTarget{application: application, syncer: syncer}, nil
}

func getKeyColumnForProvisioning(syncer *object.Syncer) *object.TableColumn {
	if len(syncer.TableColumns) == 0 {
		return nil
	}
	return syncer.GetKeyColumn()
}

func getForbiddenColumn(syncer *object.Syncer) *object.TableColumn {
	for _, tableColumn := range syncer.TableColumns {
		if tableColumn.CasdoorName == "IsForbidden" {
			return tableColumn
		}
	}
	return nil
}

func (t *databaseProvisioningTarget) connect() (*object.Syncer, string, error) {
	syncer, err := getConnectedSyncer(t.syncer)
	if err != nil {
		return nil, "", err
	}
	key := syncer.GetTargetTablePrimaryKey()
	if !util.FilterSQLIdentifier(key) {
		return nil, "", fmt.Errorf("invalid key column name: %s", key)
	}
	if !util.FilterSQLIdentifier(syncer.Table) {
		return nil, "", fmt.Errorf("invalid table name: %s", syncer.Table)
	}
	return syncer, key, nil
}

func (t *databaseProvisioningTarget) payload(user *object.User) (interface{}, error) {
	return t.syncer.GetMapFromOriginalUser(user), nil
}

func (t *databaseProvisioningTarget) upsert(user *object.User, account *ProvisionedAccount) (string, error) {
	syncer, key, err := t.connect()
	if err != nil {
		return "", err
	}

	row := syncer.GetMapFromOriginalUser(user)
	keyValue := row[key]
	if keyValue == "" {
		return "", fmt.Errorf("the key column: %s of the user: %s is empty", key, user.GetId())
	}

	// a renamed user keeps its row, found by the key it was written with
	lookupValue := keyValue
	if account.RemoteId != "" {
		lookupValue = account.RemoteId
	}
	count, err := syncer.Ormer.Engine.Table(syncer.GetTable()).Where(fmt.Sprintf("%s = ?", key), lookupValue).Count()
	if err != nil {
		return "", err
	}
	if count == 0 && lookupValue != keyValue {
		lookupValue = keyValue
		count, err = syncer.Ormer.Engine.Table(syncer.GetTable()).Where(fmt.Sprintf("%s = ?", key), lookupValue).Count()
		if err != nil {
			return "", err
		}
	}

	if count == 0 {
		_, err = syncer.Ormer.Engine.Table(syncer.GetTable()).Insert(row)
	} else {
		_, err = syncer.Ormer.Engine.Table(syncer.GetTable()).Where(fmt.Sprintf("%s = ?", key), lookupValue).Update(&row)
	}
	if err != nil {
		return "", err
	}
	return keyValue, nil
}

func (t *databaseProvisioningTarget) disable(user *object.User, account *ProvisionedAccount) error {
	syncer, key, err := t.connect()
	if err != nil {
		return err
	}
	column := getForbiddenColumn(syncer)
	if column == nil {
		return fmt.Errorf("no column of the syncer: %s is mapped to IsForbidden", syncer.GetId())
	}
	if !util.FilterSQLIdentifier(column.Name) {
		return fmt.Errorf("invalid column name: %s", column.Name)
	}

	row := syncer.GetMapFromOriginalUser(&object.User{IsForbidden: true})
	_, err = syncer.Ormer.Engine.Table(syncer.GetTable()).Where(fmt.Sprintf("%s = ?", key), account.RemoteId).
		Update(map[string]interface{}{column.Name: row[column.Name]})
	return err
}

func (t *databaseProvisioningTarget) remove(account *ProvisionedAccount) error {
	syncer, key, err := t.connect()
	if err != nil {
		return err
	}
	_, err = syncer.Ormer.Engine.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = ?", syncer.GetTable(), key), account.RemoteId)
	return err
}

func (t *databaseProvisioningTarget) test() error {
	syncer, _, err := t.connect()
	if err != nil {
		return err
	}
	_, err = syncer.Ormer.Engine.Table(syncer.GetTable()).Count()
	return err
}
