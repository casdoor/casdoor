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

//go:build !windows

package log

import (
	"fmt"
	"log/syslog"
)

type unixLogger struct {
	writer *syslog.Writer
}

func newPlatformLogger(tag string) (platformLogger, error) {
	w, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, tag)
	if err != nil {
		return nil, fmt.Errorf("SystemLogProvider: failed to open syslog: %w", err)
	}
	return &unixLogger{writer: w}, nil
}

func (u *unixLogger) log(severity, message string) error {
	switch severity {
	case "emerg":
		return u.writer.Emerg(message)
	case "alert":
		return u.writer.Alert(message)
	case "crit":
		return u.writer.Crit(message)
	case "err", "error":
		return u.writer.Err(message)
	case "warning", "warn":
		return u.writer.Warning(message)
	case "notice":
		return u.writer.Notice(message)
	case "debug":
		return u.writer.Debug(message)
	default: // "info" and anything else
		return u.writer.Info(message)
	}
}

func (u *unixLogger) close() error {
	return u.writer.Close()
}
