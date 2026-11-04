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

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

// adapterStderr is the "adapter" of logConfig that writes to stderr, for Docker and systemd.
// For example: logConfig = {"adapter":"stderr", "format":"json"}
const adapterStderr = "stderr"

// beego's router puts color codes in its messages even when the console colors are off
var ansiColorRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func init() {
	logs.Register(adapterStderr, func() logs.Logger {
		return &stderrLogger{writer: os.Stderr}
	})
}

func InitLogger(logConfig string) error {
	config := struct {
		Adapter string `json:"adapter"`
	}{Adapter: logs.AdapterFile}
	err := json.Unmarshal([]byte(logConfig), &config)
	if err != nil {
		return err
	}

	if config.Adapter == logs.AdapterConsole || config.Adapter == adapterStderr {
		logs.Reset()
	}
	err = logs.SetLogger(config.Adapter, logConfig)
	if err != nil {
		return err
	}
	logs.SetLogFuncCall(false)

	log.SetFlags(0)
	log.SetOutput(stdLogWriter{})
	return nil
}

type stdLogWriter struct{}

func (stdLogWriter) Write(p []byte) (int, error) {
	logs.Info(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

type stderrLogger struct {
	writer  io.Writer
	handler slog.Handler
	level   int
}

func (l *stderrLogger) Init(config string) error {
	settings := struct {
		Format string `json:"format"`
		Level  int    `json:"level"`
	}{Format: "text", Level: logs.LevelDebug}
	err := json.Unmarshal([]byte(config), &settings)
	if err != nil {
		return err
	}

	options := &slog.HandlerOptions{Level: slog.LevelDebug}
	switch settings.Format {
	case "text":
		l.handler = slog.NewTextHandler(l.writer, options)
	case "json":
		l.handler = slog.NewJSONHandler(l.writer, options)
	default:
		return fmt.Errorf("unsupported log format: %s", settings.Format)
	}
	l.level = settings.Level
	return nil
}

func (l *stderrLogger) WriteMsg(lm *logs.LogMsg) error {
	if lm.Level > l.level {
		return nil
	}

	msg := lm.Msg
	if len(lm.Args) > 0 {
		msg = fmt.Sprintf(msg, lm.Args...)
	}
	msg = ansiColorRegex.ReplaceAllString(msg, "")

	record := slog.NewRecord(lm.When, getSlogLevel(lm.Level), strings.TrimSpace(msg), 0)
	return l.handler.Handle(context.Background(), record)
}

func (l *stderrLogger) Destroy() {}

func (l *stderrLogger) Flush() {}

func (l *stderrLogger) SetFormatter(logs.LogFormatter) {}

func getSlogLevel(level int) slog.Level {
	switch {
	case level <= logs.LevelError:
		return slog.LevelError
	case level == logs.LevelWarn:
		return slog.LevelWarn
	case level == logs.LevelDebug:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}
