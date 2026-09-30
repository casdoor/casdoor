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

package audit

import (
	"fmt"
	"time"
)

const (
	SeverityWarning = 4
	SeverityInfo    = 6
)

type Event struct {
	Time     time.Time
	Severity int
	Action   string
	Message  string
}

type AuditProvider interface {
	Send(event *Event) error
	Close() error
}

type Config struct {
	Type     string
	Host     string
	Port     int
	Protocol string
	Facility string
	Format   string
}

func GetAuditProvider(config *Config) (AuditProvider, error) {
	switch config.Type {
	case "Syslog":
		return NewSyslogProvider(config)
	default:
		return nil, fmt.Errorf("unsupported audit provider type: %s", config.Type)
	}
}

func ValidateConfig(config *Config) error {
	switch config.Type {
	case "Syslog":
		_, err := newSyslogSettings(config)
		return err
	default:
		return fmt.Errorf("unsupported audit provider type: %s", config.Type)
	}
}
