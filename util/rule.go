// Copyright 2024 The casbin Authors. All Rights Reserved.
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

import "fmt"

// GenerateVerboseReason creates a detailed reason message for verbose mode
func GenerateVerboseReason(ruleId string, expressionReason string, customReason string) string {
	verboseReason := fmt.Sprintf("Rule [%s] triggered", ruleId)
	if expressionReason != "" {
		verboseReason += fmt.Sprintf(" - %s", expressionReason)
	}
	if customReason != "" {
		verboseReason += fmt.Sprintf(" - Custom reason: %s", customReason)
	}
	return verboseReason
}
