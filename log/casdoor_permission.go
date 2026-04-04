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

// PermissionLogProvider records Casbin authorization decisions as Entry rows.
// It is push-based: callers supply log lines via Write, which are immediately
// persisted through the injected EntryAdder. Start and Stop are no-ops.
type PermissionLogProvider struct {
	providerName string
	addEntry     EntryAdder
}

// NewPermissionLogProvider creates a PermissionLogProvider backed by addEntry.
func NewPermissionLogProvider(providerName string, addEntry EntryAdder) *PermissionLogProvider {
	return &PermissionLogProvider{providerName: providerName, addEntry: addEntry}
}

// Write stores one permission-log entry.
// severity follows syslog conventions (e.g. info, warning, err).
func (p *PermissionLogProvider) Write(severity string, message string) error {
	createdTime := time.Now().UTC().Format(time.RFC3339)
	return p.addEntry("built-in", createdTime, p.providerName, fmt.Sprintf("[%s] %s", severity, message))
}

// Start is a no-op for PermissionLogProvider; it received its EntryAdder at
// construction time and does not require background collection.
func (p *PermissionLogProvider) Start(_ EntryAdder, _ func(error)) error { return nil }

// Stop is a no-op for PermissionLogProvider.
func (p *PermissionLogProvider) Stop() error { return nil }
