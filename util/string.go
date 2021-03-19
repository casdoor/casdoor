// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"strings"

	"github.com/google/uuid"
)

func GetOwnerAndNameFromId(id string) (string, string) {
	tokens := strings.Split(id, "/")
	if len(tokens) == 2 {
		return tokens[0], tokens[1]
		//panic(errors.New("GetOwnerAndNameFromId() error, wrong token count for ID: " + id))
	} else if len(tokens) == 1 {
		return tokens[0], ""
	} else {
		return "", ""
	}

}

func GenerateId() string {
	return uuid.NewString()
}
