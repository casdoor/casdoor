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

// LogProvider sends application log lines to an external logging backend.
// Severity uses syslog-style names (e.g. emerg, alert, crit, err, warning, notice, info, debug).
type LogProvider interface {
	Write(severity string, message string) error
}

// GetLogProvider returns a concrete log provider for the given type and connection settings.
// The title parameter is used as the syslog/event-log tag for System Log.
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
