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

package object

import (
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	xormlog "github.com/xorm-io/xorm/log"
)

type xormLogger struct {
	xormlog.DiscardLogger
	showSql bool
}

func (l *xormLogger) Info(v ...interface{})  { logs.Info(fmt.Sprint(v...)) }
func (l *xormLogger) Warn(v ...interface{})  { logs.Warning(fmt.Sprint(v...)) }
func (l *xormLogger) Error(v ...interface{}) { logs.Error(fmt.Sprint(v...)) }

func (l *xormLogger) Infof(format string, v ...interface{}) { logs.Info(fmt.Sprintf(format, v...)) }

func (l *xormLogger) Warnf(format string, v ...interface{}) { logs.Warning(fmt.Sprintf(format, v...)) }

func (l *xormLogger) Errorf(format string, v ...interface{}) { logs.Error(fmt.Sprintf(format, v...)) }

func (l *xormLogger) ShowSQL(show ...bool) {
	l.showSql = len(show) == 0 || show[0]
}

func (l *xormLogger) IsShowSQL() bool {
	return l.showSql
}
