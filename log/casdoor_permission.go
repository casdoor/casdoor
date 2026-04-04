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

package log

import (
	"fmt"
	"time"
)

// EntryAdder persists a log entry to the backing store.
// Parameters map to the Entry table columns: owner, name (unique ID),
// createdTime (RFC3339), and message (the log content).
// This indirection keeps the log package free of import-cycle dependencies
// on the object package.
type EntryAdder func(owner, name, createdTime, message string) error

// PermissionLogProvider records Casbin authorization decisions as Entry rows.
// It implements LogProvider; actual storage is delegated to an EntryAdder so
// that the log package remains free of import-cycle dependencies on object.
type PermissionLogProvider struct {
	addEntry EntryAdder
}

// NewPermissionLogProvider creates a PermissionLogProvider backed by addEntry.
func NewPermissionLogProvider(addEntry EntryAdder) *PermissionLogProvider {
	return &PermissionLogProvider{addEntry: addEntry}
}

// Write stores one permission-log entry.
// severity follows syslog conventions (e.g. info, err).
func (p *PermissionLogProvider) Write(severity string, message string) error {
	name := fmt.Sprintf("perm-%d", time.Now().UnixNano())
	createdTime := time.Now().UTC().Format(time.RFC3339)
	return p.addEntry("built-in", name, createdTime, fmt.Sprintf("[%s] %s", severity, message))
}
