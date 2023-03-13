// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

//go:build !skipCi
// +build !skipCi

package sync

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestStartSyncJob(t *testing.T) {
	db1 := newDatabase("127.0.0.1", 3306, "casdoor", "root", "123456")
	db2 := newDatabase("127.0.0.1", 3306, "casdoor2", "root", "123456")
	startSyncJob(db1, db2)
}
