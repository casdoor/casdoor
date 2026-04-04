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
	"time"

	"github.com/casdoor/casdoor/log"
)

var (
	runningCollectors   = map[string]log.LogProvider{} // providerGetId() -> LogProvider
	runningCollectorsMu sync.Mutex
)

// InitLogProviders scans all globally-configured Log providers and starts
// background collection for pull-based providers (e.g. System Log).
// It is called once from main() after the database is ready.
func InitLogProviders() {
	providers, err := GetGlobalProviders()
	if err != nil {
		return
	}
	for _, p := range providers {
		if p.Category == "Log" && p.Type == "System Log" {
			startLogCollector(p)
		}
	}
}

// startLogCollector starts a System Log collector for the given provider.
// If a collector for the same provider is already running it is stopped first.
func startLogCollector(provider *Provider) {
	runningCollectorsMu.Lock()
	defer runningCollectorsMu.Unlock()

	id := provider.GetId()

	// Stop the existing collector for this provider if any.
	if existing, ok := runningCollectors[id]; ok {
		_ = existing.Stop()
		delete(runningCollectors, id)
	}

	tag := provider.Title
	if tag == "" {
		tag = "casdoor"
	}

	lp, err := log.NewSystemLogProvider(tag)
	if err != nil {
		return
	}

	providerName := provider.Name
	addEntry := func(owner, name, createdTime, pName, message string) error {
		entry := &Entry{
			Owner:       owner,
			Name:        name,
			CreatedTime: createdTime,
			UpdatedTime: createdTime,
			DisplayName: name,
			Provider:    pName,
			Message:     message,
		}
		_, err := AddEntry(entry)
		return err
	}

	onError := func(err error) {
		fmt.Printf("InitLogProviders: collector for provider %s stopped with error: %v\n", providerName, err)
	}
	if err := lp.Start(addEntry, onError); err != nil {
		fmt.Printf("InitLogProviders: failed to start collector for provider %s: %v\n", providerName, err)
		return
	}

	runningCollectors[id] = lp
}

// GetOpenClawProviderByIP returns the first Log/Agent/OpenClaw provider that
// allows the given client IP. A provider with an empty Host field allows any IP.
func GetOpenClawProviderByIP(clientIP string) (*Provider, error) {
	providers := []*Provider{}
	err := ormer.Engine.Where("category = ? AND type = ? AND sub_type = ?", "Log", "Agent", "OpenClaw").Find(&providers)
	if err != nil {
		return nil, err
	}
	for _, p := range providers {
		if p.Host == "" || p.Host == clientIP {
			return p, nil
		}
	}
	return nil, nil
}

// makeEntryName returns a hex-encoded unique name for an Entry row.
func makeEntryName() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
