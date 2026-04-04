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

// SELinuxLogProvider collects SELinux audit events (AVC denials and related
// records) from the local system and stores each record as an Entry row via
// the EntryAdder supplied to Start.
//
// It is pull-based: Write is not applicable and returns an error.
// Start launches the background collector; Stop cancels it.
// On platforms where SELinux is not supported, Start returns an error.
type SELinuxLogProvider struct {
	cancel context.CancelFunc
}

// NewSELinuxLogProvider creates a SELinuxLogProvider.
// Call Start to begin collection.
func NewSELinuxLogProvider() (*SELinuxLogProvider, error) {
	return &SELinuxLogProvider{}, nil
}

// Write is not applicable for a pull-based collector and always returns an error.
func (s *SELinuxLogProvider) Write(severity string, message string) error {
	return fmt.Errorf("SELinuxLogProvider is a log collector and does not accept Write calls")
}

// Start launches a background goroutine that reads new SELinux audit records
// and persists each one by calling addEntry. Returns immediately; collection
// runs until Stop is called. If the goroutine encounters a fatal error,
// onError is called with that error (onError may be nil).
func (s *SELinuxLogProvider) Start(addEntry EntryAdder, onError func(error)) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go func() {
		if err := collectSELinuxLogs(ctx, addEntry); err != nil && onError != nil {
			onError(err)
		}
	}()
	return nil
}

// Stop cancels background collection. It is safe to call multiple times.
func (s *SELinuxLogProvider) Stop() error {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	return nil
}
