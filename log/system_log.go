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

// SystemLogProvider writes log lines to the operating-system's native logging
// facility (syslog on Linux/Unix, Event Log on Windows).
// It implements LogProvider; the platform-specific open/write/close logic is
// provided by system_log_unix.go and system_log_windows.go via the
// platformLogger interface.
type SystemLogProvider struct {
	tag    string
	logger platformLogger
}

// platformLogger abstracts the OS-level logging primitives so that
// SystemLogProvider stays free of build-tag clutter.
type platformLogger interface {
	// log writes a single line at the given syslog-style severity.
	log(severity, message string) error
	// close releases any OS resource held by the logger.
	close() error
}

// NewSystemLogProvider opens the OS logging facility with the supplied tag and
// returns a ready-to-use SystemLogProvider.
func NewSystemLogProvider(tag string) (*SystemLogProvider, error) {
	pl, err := newPlatformLogger(tag)
	if err != nil {
		return nil, err
	}
	return &SystemLogProvider{tag: tag, logger: pl}, nil
}

// Write implements LogProvider.
func (s *SystemLogProvider) Write(severity string, message string) error {
	return s.logger.log(severity, message)
}
