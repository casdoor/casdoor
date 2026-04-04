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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type unixCollector struct {
	tag string
}

func newPlatformCollector(tag string) platformCollector {
	return &unixCollector{tag: tag}
}

// collect streams new journald records to addEntry until ctx is cancelled or
// a fatal error occurs. It runs `journalctl -n 0 -f --output=json` so only
// records that arrive after Start is called are collected (no backfill).
// Returns nil when ctx is cancelled normally; returns a non-nil error if the
// process could not be started or the output pipe broke unexpectedly.
func (u *unixCollector) collect(ctx context.Context, addEntry EntryAdder) error {
	cmd := exec.CommandContext(ctx, "journalctl", "-n", "0", "-f", "--output=json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("SystemLogProvider: failed to open journalctl stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("SystemLogProvider: failed to start journalctl: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	// journald JSON lines can be large; use a 1 MB buffer.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var fields map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &fields); err != nil {
			continue
		}

		severity := journalSeverity(fields)
		message := journalMessage(fields)
		createdTime := journalTimestamp(fields)
		name := fmt.Sprintf("%x", time.Now().UnixNano())

		if err := addEntry("built-in", name, createdTime, u.tag,
			fmt.Sprintf("[%s] %s", severity, message)); err != nil {
			return fmt.Errorf("SystemLogProvider: failed to persist journal entry: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		// A cancelled context causes the pipe to close; treat that as normal exit.
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("SystemLogProvider: journalctl output error: %w", err)
	}
	return nil
}

// journalSeverity maps the journald PRIORITY field to a syslog severity name.
// PRIORITY values: 0=emerg 1=alert 2=crit 3=err 4=warning 5=notice 6=info 7=debug
func journalSeverity(fields map[string]interface{}) string {
	mapping := map[string]string{
		"0": "emerg", "1": "alert", "2": "crit", "3": "err",
		"4": "warning", "5": "notice", "6": "info", "7": "debug",
	}
	if p, ok := fields["PRIORITY"].(string); ok {
		if s, ok2 := mapping[p]; ok2 {
			return s
		}
	}
	return "info"
}

// journalMessage extracts the human-readable message from journald JSON.
func journalMessage(fields map[string]interface{}) string {
	if msg, ok := fields["MESSAGE"].(string); ok {
		return msg
	}
	return ""
}

// journalTimestamp converts the journald __REALTIME_TIMESTAMP (microseconds
// since Unix epoch) to an RFC3339 string.
func journalTimestamp(fields map[string]interface{}) string {
	if ts, ok := fields["__REALTIME_TIMESTAMP"].(string); ok {
		usec, err := strconv.ParseInt(ts, 10, 64)
		if err == nil {
			t := time.Unix(usec/1_000_000, (usec%1_000_000)*1_000).UTC()
			return t.Format(time.RFC3339)
		}
	}
	return time.Now().UTC().Format(time.RFC3339)
}
