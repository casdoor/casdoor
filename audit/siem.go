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
	"strings"
	"time"
)

const siemProductVersion = "1.0"

var cefHeaderEscaper = strings.NewReplacer(`\`, `\\`, `|`, `\|`, "\r", " ", "\n", " ")

var cefValueEscaper = strings.NewReplacer(`\`, `\\`, `=`, `\=`, "\r", `\r`, "\n", `\n`)

var leefValueEscaper = strings.NewReplacer("\t", " ", "\r", " ", "\n", " ")

func getSiemFields(event *Event) *EventFields {
	if event.Fields == nil {
		return &EventFields{Success: event.Severity != SeverityWarning}
	}
	return event.Fields
}

func getOutcome(fields *EventFields) string {
	if fields.Success {
		return "success"
	}
	return "failure"
}

// formatCef writes the event in ArcSight Common Event Format: CEF:Version|Vendor|Product|Version|EventClassID|Name|
// Severity|Extension, the severity from 0 to 10
func formatCef(event *Event, eventTime time.Time) string {
	fields := getSiemFields(event)
	severity := 3
	if !fields.Success {
		severity = 6
	}
	action := cefHeaderEscaper.Replace(event.Action)

	pairs := [][2]string{
		{"rt", fmt.Sprintf("%d", eventTime.UnixMilli())},
		{"act", event.Action},
		{"outcome", getOutcome(fields)},
		{"src", fields.ClientIp},
		{"suser", fields.User},
		{"requestMethod", fields.Method},
		{"request", fields.RequestUri},
		{"cs1Label", "organization"},
		{"cs1", fields.Organization},
		{"cn1Label", "statusCode"},
		{"cn1", fmt.Sprintf("%d", fields.StatusCode)},
	}
	extension := []string{}
	for _, pair := range pairs {
		if pair[1] != "" {
			extension = append(extension, pair[0]+"="+cefValueEscaper.Replace(pair[1]))
		}
	}
	return fmt.Sprintf("CEF:0|Casdoor|Casdoor|%s|%s|%s|%d|%s", siemProductVersion, action, action, severity, strings.Join(extension, " "))
}

// formatLeef writes the event in IBM Log Event Extended Format 1.0, attributes separated by tabs
func formatLeef(event *Event, eventTime time.Time) string {
	fields := getSiemFields(event)
	pairs := [][2]string{
		{"devTime", eventTime.Format("Jan 02 2006 15:04:05")},
		{"devTimeFormat", "MMM dd yyyy HH:mm:ss"},
		{"cat", "authentication"},
		{"outcome", getOutcome(fields)},
		{"src", fields.ClientIp},
		{"usrName", fields.User},
		{"method", fields.Method},
		{"url", fields.RequestUri},
		{"organization", fields.Organization},
		{"statusCode", fmt.Sprintf("%d", fields.StatusCode)},
	}
	attributes := []string{}
	for _, pair := range pairs {
		if pair[1] != "" {
			attributes = append(attributes, pair[0]+"="+leefValueEscaper.Replace(pair[1]))
		}
	}
	return fmt.Sprintf("LEEF:1.0|Casdoor|Casdoor|%s|%s|%s", siemProductVersion, cefHeaderEscaper.Replace(event.Action), strings.Join(attributes, "\t"))
}
