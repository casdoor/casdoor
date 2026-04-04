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

//go:build windows

package log

import (
	"fmt"

	"golang.org/x/sys/windows/svc/eventlog"
)

type windowsLogger struct {
	writer *eventlog.Log
}

func newPlatformLogger(tag string) (platformLogger, error) {
	// Ensure the event source exists; ignore the error if it already does.
	_ = eventlog.InstallAsEventCreate(tag, eventlog.Info|eventlog.Warning|eventlog.Error)

	w, err := eventlog.Open(tag)
	if err != nil {
		return nil, fmt.Errorf("SystemLogProvider: failed to open Windows Event Log: %w", err)
	}
	return &windowsLogger{writer: w}, nil
}

func (w *windowsLogger) log(severity, message string) error {
	switch severity {
	case "emerg", "alert", "crit", "err", "error":
		return w.writer.Error(1, message)
	case "warning", "warn":
		return w.writer.Warning(2, message)
	default: // "notice", "info", "debug" and anything else
		return w.writer.Info(3, message)
	}
}

func (w *windowsLogger) close() error {
	return w.writer.Close()
}
