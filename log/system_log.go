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
	"context"
	"fmt"
)

// platformCollector is an OS-specific log reader.
// Implementations are in system_log_unix.go and system_log_windows.go.
type platformCollector interface {
	// collect blocks and streams new OS log records to addEntry until ctx is
	// cancelled or a fatal error occurs. It must return promptly when
	// ctx.Done() is closed. A non-nil error means collection stopped
	// unexpectedly and should be reported to the operator.
	collect(ctx context.Context, addEntry EntryAdder) error
}

// SystemLogProvider collects log records from the operating-system's native
// logging facility (journald/syslog on Linux/Unix, Event Log on Windows) and
// stores each record as an Entry row via the EntryAdder supplied to Start.
//
// It is pull-based: Write is not applicable and returns an error.
// Start launches the background collector; Stop cancels it.
type SystemLogProvider struct {
	tag    string
	cancel context.CancelFunc
}

// NewSystemLogProvider creates a SystemLogProvider that will identify itself
// with the given tag when collecting OS log records.
// Call Start to begin collection.
func NewSystemLogProvider(tag string) (*SystemLogProvider, error) {
	return &SystemLogProvider{tag: tag}, nil
}

// Write is not applicable for a pull-based collector and always returns an
// error. Callers in the permission-log path should skip System Log providers.
func (s *SystemLogProvider) Write(severity string, message string) error {
	return fmt.Errorf("SystemLogProvider is a log collector and does not accept Write calls")
}

// Start launches a background goroutine that reads new OS log records and
// persists each one by calling addEntry. It returns immediately; collection
// runs until Stop is called. If the goroutine encounters a fatal error,
// onError is called with that error (onError may be nil).
func (s *SystemLogProvider) Start(addEntry EntryAdder, onError func(error)) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	collector := newPlatformCollector(s.tag)
	go func() {
		if err := collector.collect(ctx, addEntry); err != nil && onError != nil {
			onError(err)
		}
	}()
	return nil
}

// Stop cancels background collection. It is safe to call multiple times.
func (s *SystemLogProvider) Stop() error {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	return nil
}
