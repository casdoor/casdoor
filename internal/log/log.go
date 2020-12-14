// Copyright 2020 The casbin Authors. All Rights Reserved.
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

import (
	"go.uber.org/zap"
)

var logger = zap.NewExample()

func LogError(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func LogDebug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func LogInfo(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func LogWarn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}
