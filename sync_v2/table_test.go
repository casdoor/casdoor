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

package sync_v2

import (
	"log"
	"math/rand"
	"testing"

	"github.com/casdoor/casdoor/util"
)

type TestUser struct {
	Id       int64  `xorm:"pk autoincr"`
	Username string `xorm:"varchar(50)"`
	Address  string `xorm:"varchar(50)"`
	Card     string `xorm:"varchar(50)"`
	Age      int
}

func TestCreateUserTable(t *testing.T) {
	db := newDatabase(&Configs[0])
	err := db.engine.Sync2(new(TestUser))
	if err != nil {
		log.Fatalln(err)
	}
}

func TestInsertUser(t *testing.T) {
	db := newDatabase(&Configs[0])
	// random generate user
	user := &TestUser{
		Username: util.GetRandomName(),
		Age:      rand.Intn(100) + 10,
	}
	_, err := db.engine.Insert(user)
	if err != nil {
		log.Fatalln(err)
	}
}

func TestDeleteUser(t *testing.T) {
	db := newDatabase(&Configs[0])
	user := &TestUser{
		Id: 10,
	}
	_, err := db.engine.Delete(user)
	if err != nil {
		log.Fatalln(err)
	}
}
