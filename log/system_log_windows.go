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
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Windows Event Log channels to collect from.
var eventLogChannels = []string{"System", "Application"}

type windowsCollector struct {
	tag string
}

func newPlatformCollector(tag string) platformCollector {
	return &windowsCollector{tag: tag}
}

// collect polls Windows Event Log channels every 5 seconds via wevtutil.exe
// and persists new records to addEntry. Only events that arrive after Start
// is called are collected; historical events are not backfilled.
// Returns nil when ctx is cancelled normally.
func (w *windowsCollector) collect(ctx context.Context, addEntry EntryAdder) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastCheck := time.Now().UTC()

	for {
		select {
		case <-ctx.Done():
			return nil
		case tick := <-ticker.C:
			for _, channel := range eventLogChannels {
				if err := w.queryChannel(ctx, channel, lastCheck, addEntry); err != nil {
					return fmt.Errorf("SystemLogProvider: error querying channel %s: %w", channel, err)
				}
			}
			lastCheck = tick.UTC()
		}
	}
}

// queryChannel runs wevtutil.exe to fetch events from channel that were
// created after since, then stores each event via addEntry.
// Returns a non-nil error if the wevtutil command fails or XML parsing fails.
func (w *windowsCollector) queryChannel(ctx context.Context, channel string, since time.Time, addEntry EntryAdder) error {
	sinceStr := since.Format("2006-01-02T15:04:05.000Z")
	query := fmt.Sprintf("*[System[TimeCreated[@SystemTime>='%s']]]", sinceStr)

	cmd := exec.CommandContext(ctx, "wevtutil.exe", "qe", channel,
		"/f:RenderedXml", "/rd:false",
		fmt.Sprintf("/q:%s", query),
	)
	out, err := cmd.Output()
	if err != nil {
		// A cancelled context is a normal shutdown, not an error.
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("wevtutil.exe failed for channel %s: %w", channel, err)
	}
	if len(out) == 0 {
		return nil
	}

	// wevtutil outputs one <Event> element per record; wrap in a root element
	// so the standard XML decoder can handle multiple records at once.
	wrapped := "<Events>" + string(out) + "</Events>"
	decoder := xml.NewDecoder(strings.NewReader(wrapped))

	for {
		var event winEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("SystemLogProvider: failed to decode event XML (channel=%s): %w", channel, err)
		}
		if event.XMLName.Local != "Event" {
			continue
		}

		severity := winEventSeverity(event.System.Level)
		message := strings.TrimSpace(event.RenderingInfo.Message)
		if message == "" {
			message = fmt.Sprintf("EventID=%d Source=%s", event.System.EventID, event.System.Provider.Name)
		}
		createdTime := winEventTimestamp(event.System.TimeCreated.SystemTime)
		name := fmt.Sprintf("%x", time.Now().UnixNano())

		if err := addEntry("built-in", name, createdTime, w.tag,
			fmt.Sprintf("[%s] [%s] %s", severity, channel, message)); err != nil {
			return fmt.Errorf("SystemLogProvider: failed to persist event (channel=%s EventID=%d): %w",
				channel, event.System.EventID, err)
		}
	}
	return nil
}

// winEvent represents the subset of the Windows Event XML schema that we need.
type winEvent struct {
	XMLName xml.Name `xml:"Event"`
	System  struct {
		Provider struct {
			Name string `xml:"Name,attr"`
		} `xml:"Provider"`
		EventID     int `xml:"EventID"`
		Level       int `xml:"Level"`
		TimeCreated struct {
			SystemTime string `xml:"SystemTime,attr"`
		} `xml:"TimeCreated"`
	} `xml:"System"`
	RenderingInfo struct {
		Message string `xml:"Message"`
	} `xml:"RenderingInfo"`
}

// winEventSeverity maps Windows Event Log Level values to syslog severity names.
// Level: 1=Critical 2=Error 3=Warning 4=Information 5=Verbose
func winEventSeverity(level int) string {
	switch level {
	case 1:
		return "crit"
	case 2:
		return "err"
	case 3:
		return "warning"
	case 5:
		return "debug"
	default: // 4=Information and anything else
		return "info"
	}
}

// winEventTimestamp parses a Windows Event SystemTime attribute string to RFC3339.
func winEventTimestamp(s string) string {
	// SystemTime is in the form "2024-01-15T10:30:00.000000000Z"
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		// Try without nanoseconds
		t, err = time.Parse("2006-01-02T15:04:05.000000000Z", s)
		if err != nil {
			return time.Now().UTC().Format(time.RFC3339)
		}
	}
	return t.UTC().Format(time.RFC3339)
}
