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
	"fmt"
	"regexp"
	"strings"
)

var jsonTypeErrorRegex = regexp.MustCompile(`^json: cannot unmarshal (\w+)(?: \S+)? into Go (?:struct field \w+\.(\S+)|value) of type (\S+)$`)

// GetReadableError rewrites the raw errors of Go that reach the API caller into what the caller can act on
func GetReadableError(error string) string {
	if error == "unexpected end of JSON input" {
		return "invalid JSON: unexpected end of JSON input, the request body may be empty or cut off"
	}

	if match := jsonTypeErrorRegex.FindStringSubmatch(error); match != nil {
		if match[2] == "" {
			return fmt.Sprintf("invalid JSON: the request body should be %s, not %s", describeJsonType(match[3]), describeJsonType(match[1]))
		}
		return fmt.Sprintf("invalid JSON: the field \"%s\" should be %s, not %s", match[2], describeJsonType(match[3]), describeJsonType(match[1]))
	}

	if id, ok := strings.CutPrefix(error, "GetOwnerAndNameFromId() error, wrong token count for ID: "); ok {
		return fmt.Sprintf("the ID: \"%s\" is invalid, it should be in the form of \"owner/name\"", id)
	}

	return error
}

func describeJsonType(typ string) string {
	switch {
	case typ == "array" || strings.HasPrefix(typ, "[]"):
		return "an array"
	case typ == "string":
		return "a string"
	case typ == "bool":
		return "a boolean"
	case typ == "number" || strings.HasPrefix(typ, "int") || strings.HasPrefix(typ, "uint") || strings.HasPrefix(typ, "float"):
		return "a number"
	default:
		return "an object"
	}
}
