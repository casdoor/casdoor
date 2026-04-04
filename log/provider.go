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

import "fmt"

// EntryAdder persists a collected log entry into the backing store.
// Parameters map to the Entry table columns: owner, name (unique ID),
// createdTime (RFC3339), provider (the log provider name), and message.
// Defined here so it is shared by all LogProvider implementations without
// creating import cycles with the object package.
type EntryAdder func(owner, name, createdTime, provider, message string) error

// LogProvider is the common interface for all log providers.
//
// Push-based providers (e.g. PermissionLogProvider) receive individual log
// lines through Write and persist them immediately. Start and Stop are no-ops
// for these providers.
//
// Pull-based providers (e.g. SystemLogProvider) actively collect logs from an
// external source. Start begins a background collection goroutine that calls
// addEntry for every new record; Stop halts collection. Write returns an error
// for these providers as they are not designed to accept external input.
type LogProvider interface {
	// Write records a single log line. Used by push-based providers.
	Write(severity string, message string) error
	// Start begins background log collection with the given EntryAdder.
	// For push-based providers this is a no-op (they received addEntry at
	// construction time). onError is called from the background goroutine
	// when collection stops with a fatal error; it may be nil.
	Start(addEntry EntryAdder, onError func(error)) error
	// Stop halts background collection and releases any OS resources.
	Stop() error
}

// GetLogProvider returns a concrete log provider for the given type and connection settings.
// The title parameter is used as the OS log tag for System Log.
// Types that are not yet implemented return a non-nil error.
func GetLogProvider(typ string, _ string, _ int, title string) (LogProvider, error) {
	switch typ {
	case "System Log":
		tag := title
		if tag == "" {
			tag = "casdoor"
		}
		return NewSystemLogProvider(tag)
	default:
		return nil, fmt.Errorf("unsupported log provider type: %s", typ)
	}
}
