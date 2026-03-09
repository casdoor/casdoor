// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

// AdapterStderr is the name of the stderr log adapter.
const AdapterStderr = "stderr"

// logLevelNames maps beego log levels to human-readable names for JSON output.
// This array is read-only after init and does not require synchronization.
var logLevelNames = [logs.LevelDebug + 1]string{
	"emergency", "alert", "critical", "error", "warning", "notice", "info", "debug",
}

var ansiEscapeRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// defaultTimeFormat is the timestamp format used in plain-text log output.
const defaultTimeFormat = "2006/01/02 15:04:05.000 "

// JSONFormatter formats log messages as JSON for use with log aggregation systems.
// To use it, set the log formatter to "json" in logConfig, e.g.:
//
//	logConfig = {"adapter":"stderr","formatter":"json"}
//	logConfig = {"adapter":"file","filename":"logs/casdoor.log","formatter":"json"}
type JSONFormatter struct{}

// Format renders a log message as a JSON string.
func (j *JSONFormatter) Format(lm *logs.LogMsg) string {
	msg := lm.Msg
	if len(lm.Args) > 0 {
		msg = fmt.Sprintf(lm.Msg, lm.Args...)
	}
	msg = ansiEscapeRegexp.ReplaceAllString(msg, "")

	level := "unknown"
	if lm.Level >= 0 && lm.Level <= logs.LevelDebug {
		level = logLevelNames[lm.Level]
	}

	record := map[string]interface{}{
		"level": level,
		"ts":    lm.When.UTC().Format(time.RFC3339Nano),
		"msg":   msg,
	}

	if lm.FilePath != "" {
		record["caller"] = fmt.Sprintf("%s:%d", lm.FilePath, lm.LineNumber)
	}

	b, err := json.Marshal(record)
	if err != nil {
		return fmt.Sprintf(`{"level":"error","ts":%q,"msg":"failed to marshal log record: %v"}`,
			lm.When.UTC().Format(time.RFC3339Nano), err)
	}
	return string(b)
}

// stderrWriter implements logs.Logger and writes log messages to os.Stderr.
// It supports the same "formatter" and "level" config options as the console adapter.
// To use it, set the log adapter to "stderr" in logConfig, e.g.:
//
//	logConfig = {"adapter":"stderr"}
//	logConfig = {"adapter":"stderr","formatter":"json"}
//	logConfig = {"adapter":"stderr","level":4}
type stderrWriter struct {
	mu        sync.Mutex
	writer    io.Writer
	formatter logs.LogFormatter
	Level     int    `json:"level"`
	Formatter string `json:"formatter"`
}

// Format implements logs.LogFormatter. It provides plain-text output without color escapes.
func (w *stderrWriter) Format(lm *logs.LogMsg) string {
	msg := ansiEscapeRegexp.ReplaceAllString(lm.OldStyleFormat(), "")
	t := lm.When.Format(defaultTimeFormat)
	return t + msg
}

// SetFormatter sets the formatter for this writer.
func (w *stderrWriter) SetFormatter(f logs.LogFormatter) {
	w.mu.Lock()
	w.formatter = f
	w.mu.Unlock()
}

// Init initializes the stderrWriter from a JSON config string.
func (w *stderrWriter) Init(config string) error {
	if len(config) == 0 {
		return nil
	}
	err := json.Unmarshal([]byte(config), w)
	if err != nil {
		return err
	}
	if len(w.Formatter) > 0 {
		fmtr, ok := logs.GetFormatter(w.Formatter)
		if !ok {
			return fmt.Errorf("the formatter with name: %s not found", w.Formatter)
		}
		w.mu.Lock()
		w.formatter = fmtr
		w.mu.Unlock()
	}
	return nil
}

// WriteMsg writes a log message to stderr.
func (w *stderrWriter) WriteMsg(lm *logs.LogMsg) error {
	if lm.Level > w.Level {
		return nil
	}
	w.mu.Lock()
	formatter := w.formatter
	w.mu.Unlock()
	msg := formatter.Format(lm)
	w.mu.Lock()
	_, err := fmt.Fprintln(w.writer, msg)
	w.mu.Unlock()
	return err
}

// Destroy is a no-op for stderr.
func (w *stderrWriter) Destroy() {}

// Flush is a no-op for stderr.
func (w *stderrWriter) Flush() {}

func newStderrWriter() logs.Logger {
	sw := &stderrWriter{
		writer: os.Stderr,
		Level:  logs.LevelDebug,
	}
	sw.formatter = sw
	return sw
}

func init() {
	logs.RegisterFormatter("json", &JSONFormatter{})
	logs.Register(AdapterStderr, newStderrWriter)
}
