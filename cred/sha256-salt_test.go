// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"testing"
)

func TestGetSaltedPassword(t *testing.T) {
	password := "123456"
	salt := "123"
	cm := NewSha256SaltCredManager()
	fmt.Printf("%s -> %s\n", password, cm.GetHashedPassword(password, salt))
}

func TestGetPassword(t *testing.T) {
	password := "123456"
	cm := NewSha256SaltCredManager()
	// https://passwordsgenerator.net/sha256-hash-generator/
	fmt.Printf("%s -> %s\n", "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", cm.GetHashedPassword(password, ""))
}
