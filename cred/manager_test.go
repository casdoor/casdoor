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

package cred

import (
	"fmt"
	"strconv"
	"testing"
)

func TestGetCredManager(t *testing.T) {
	password := "123456"
	salt := "123"
	var cm CredManager
	methods := []string{"plain", "sha256-salt", "md5-salt"}
	for _, method := range methods {
		cm = GetCredManager(method)
		passwordHash := cm.GetSealedPassword(password, salt)
		fmt.Printf("[%s] %s -> %s\n", method, password, passwordHash)
		fmt.Printf("[%s] Test CheckPassword: %s\n", method, strconv.FormatBool(cm.CheckSealedPassword(password, passwordHash)))
	}
	cm = GetCredManager("plain")
	fmt.Printf("Test CheckPassword for plain $plain$123456: %s\n", strconv.FormatBool(cm.CheckSealedPassword(password, "$plain$123456")))
	fmt.Printf("Test CheckPassword for plain 123456: %s\n", strconv.FormatBool(cm.CheckSealedPassword(password, "123456")))
}
