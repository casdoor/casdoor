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

//go:build linux

package log

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const auditLogPath = "/var/log/audit/audit.log"

// selinuxAuditTypes is the set of audit record types that are SELinux-related.
var selinuxAuditTypes = map[string]bool{
	"AVC":             true,
	"USER_AVC":        true,
	"SELINUX_ERR":     true,
	"MAC_POLICY_LOAD": true,
	"MAC_STATUS":      true,
}

// auditTimestampRe matches the msg=audit(seconds.millis:serial) field.
var auditTimestampRe = regexp.MustCompile(`msg=audit\((\d+)\.\d+:\d+\)`)

// CheckSELinuxAvailable returns nil if SELinux is active and the audit log is
// readable on this system. Returns a descriptive error otherwise.
func CheckSELinuxAvailable() error {
	if _, err := os.Stat("/sys/fs/selinux/enforce"); os.IsNotExist(err) {
		return fmt.Errorf("SELinux is not available or not mounted on this system")
	}
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		return fmt.Errorf("SELinux audit log not found at %s (is auditd running?)", auditLogPath)
	}
	return nil
}

// collectSELinuxLogs tails /var/log/audit/audit.log and persists each
// SELinux-related audit record via addEntry until ctx is cancelled.
func collectSELinuxLogs(ctx context.Context, addEntry EntryAdder) error {
	if err := CheckSELinuxAvailable(); err != nil {
		return fmt.Errorf("SELinuxLogProvider: %w", err)
	}

	cmd := exec.CommandContext(ctx, "tail", "-f", "-n", "0", auditLogPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("SELinuxLogProvider: failed to open audit log pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("SELinuxLogProvider: failed to start tail: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line := scanner.Text()
		if !isSELinuxAuditLine(line) {
			continue
		}

		severity := selinuxSeverity(line)
		createdTime := parseAuditTimestamp(line)
		if err := addEntry("built-in", createdTime, "",
			fmt.Sprintf("[%s] %s", severity, line)); err != nil {
			return fmt.Errorf("SELinuxLogProvider: failed to persist audit entry: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("SELinuxLogProvider: audit log read error: %w", err)
	}
	return nil
}

// isSELinuxAuditLine reports whether the audit log line is an SELinux record.
func isSELinuxAuditLine(line string) bool {
	// Audit lines start with "type=<TYPE> "
	const prefix = "type="
	if !strings.HasPrefix(line, prefix) {
		return false
	}
	end := strings.IndexByte(line[len(prefix):], ' ')
	var typ string
	if end < 0 {
		typ = line[len(prefix):]
	} else {
		typ = line[len(prefix) : len(prefix)+end]
	}
	return selinuxAuditTypes[typ]
}

// selinuxSeverity maps SELinux audit record types to a syslog severity name.
func selinuxSeverity(line string) string {
	if strings.HasPrefix(line, "type=AVC") || strings.HasPrefix(line, "type=USER_AVC") || strings.HasPrefix(line, "type=SELINUX_ERR") {
		return "warning"
	}
	return "info"
}

// parseAuditTimestamp extracts the Unix timestamp from an audit log line and
// returns it as an RFC3339 string. Falls back to the current time on failure.
func parseAuditTimestamp(line string) string {
	m := auditTimestampRe.FindStringSubmatch(line)
	if m == nil {
		return time.Now().UTC().Format(time.RFC3339)
	}
	sec, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return time.Unix(sec, 0).UTC().Format(time.RFC3339)
}
