// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"fmt"
	"strings"
	"time"

	"github.com/casdoor/casdoor/audit"
)

const siemProductVersion = "1.0"

var cefHeaderEscaper = strings.NewReplacer(`\`, `\\`, `|`, `\|`, "\r", " ", "\n", " ")

var cefValueEscaper = strings.NewReplacer(`\`, `\\`, `=`, `\=`, "\r", `\r`, "\n", `\n`)

var leefValueEscaper = strings.NewReplacer("\t", " ", "\r", " ", "\n", " ")

func getSiemFields(event *audit.Event) *audit.EventFields {
	if event.Fields == nil {
		return &audit.EventFields{Success: event.Severity != audit.SeverityWarning}
	}
	return event.Fields
}

func getOutcome(fields *audit.EventFields) string {
	if fields.Success {
		return "success"
	}
	return "failure"
}

// formatCef writes the event in ArcSight Common Event Format: CEF:Version|Vendor|Product|Version|EventClassID|Name|
// Severity|Extension, the severity from 0 to 10
func formatCef(event *audit.Event, eventTime time.Time) string {
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
func formatLeef(event *audit.Event, eventTime time.Time) string {
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
