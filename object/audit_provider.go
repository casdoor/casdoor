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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/audit"
)

var (
	runningAuditProviders   = map[string]audit.AuditProvider{}
	runningAuditProvidersMu sync.Mutex
)

func getAuditConfig(provider *Provider) *audit.Config {
	return &audit.Config{
		Type:     provider.Type,
		Host:     provider.Host,
		Port:     provider.Port,
		Protocol: provider.Method,
		Facility: provider.Title,
		Format:   provider.TemplateCode,
	}
}

func validateAuditProvider(provider *Provider) error {
	if provider.Category != "Audit" {
		return nil
	}
	return audit.ValidateConfig(getAuditConfig(provider))
}

func getAuditProvider(provider *Provider) (audit.AuditProvider, error) {
	runningAuditProvidersMu.Lock()
	defer runningAuditProvidersMu.Unlock()

	id := provider.GetId()
	if auditProvider, ok := runningAuditProviders[id]; ok {
		return auditProvider, nil
	}

	auditProvider, err := audit.GetAuditProvider(getAuditConfig(provider))
	if err != nil {
		return nil, err
	}
	runningAuditProviders[id] = auditProvider
	return auditProvider, nil
}

func stopAuditProvider(id string) {
	runningAuditProvidersMu.Lock()
	defer runningAuditProvidersMu.Unlock()

	if auditProvider, ok := runningAuditProviders[id]; ok {
		_ = auditProvider.Close()
		delete(runningAuditProviders, id)
	}
}

func getAuditSeverity(record *Record) int {
	if record.StatusCode >= 400 || strings.HasPrefix(record.Response, `{status:"error"`) {
		return audit.SeverityWarning
	}
	return audit.SeverityInfo
}

func sendAuditRecord(record *Record) {
	providers, err := GetProvidersByCategory(record.Organization, "Audit")
	if err != nil {
		fmt.Printf("sendAuditRecord() error: %v\n", err)
		return
	}
	if len(providers) == 0 {
		return
	}

	message, err := json.Marshal(record)
	if err != nil {
		fmt.Printf("sendAuditRecord() error: %v\n", err)
		return
	}

	eventTime, err := time.Parse(time.RFC3339, record.CreatedTime)
	if err != nil {
		eventTime = time.Now()
	}

	event := &audit.Event{
		Time:     eventTime,
		Severity: getAuditSeverity(record),
		Action:   record.Action,
		Message:  string(message),
	}

	for _, provider := range providers {
		if provider.State == "Disabled" {
			continue
		}

		auditProvider, err := getAuditProvider(provider)
		if err != nil {
			fmt.Printf("sendAuditRecord() error: provider %s: %v\n", provider.GetId(), err)
			continue
		}

		err = auditProvider.Send(event)
		if err != nil {
			fmt.Printf("sendAuditRecord() error: provider %s: %v\n", provider.GetId(), err)
		}
	}
}
