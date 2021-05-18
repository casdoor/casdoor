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

package object

import (
	"fmt"
	"math/rand"
	"time"

	"xorm.io/core"
)

type VerificationRecord struct {
	RemoteAddr string `xorm:"varchar(100) notnull pk"`
	Type       string `xorm:"varchar(10) notnull pk"`
	Receiver   string `xorm:"varchar(100) notnull"`
	Code       string `xorm:"varchar(10) notnull"`
	Time       int64  `xorm:"notnull"`
	IsUsed     bool
}

func SendVerificationCodeToEmail(remoteAddr, dest string) string {
	title := "Casdoor Verification Code"
	sender := "Casdoor"
	code := getRandomCode(5)
	content := fmt.Sprintf("You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes.", code)

	if result := AddToVerificationRecord(remoteAddr, "email", dest, code); len(result) != 0 {
		return result
	}

	msg, err := SendEmail(title, content, dest, sender)
	if msg != "" {
		return msg
	}
	if err != nil {
		panic(err)
	}

	return ""
}

func SendVerificationCodeToPhone(remoteAddr, dest string) string {
	code := getRandomCode(5)
	if result := AddToVerificationRecord(remoteAddr, "phone", dest, code); len(result) != 0 {
		return result
	}

	return SendCodeToPhone(dest, code)
}

func AddToVerificationRecord(remoteAddr, recordType, dest, code string) string {
	var record VerificationRecord
	record.RemoteAddr = remoteAddr
	record.Type = recordType
	has, err := adapter.Engine.Get(&record)
	if err != nil {
		panic(err)
	}

	now := time.Now().Unix()

	if has && now - record.Time < 60 {
		return "You can only send one code in 60s."
	}

	record.Receiver = dest
	record.Code = code
	record.Time = now
	record.IsUsed = false

	if has {
		_, err = adapter.Engine.ID(core.PK{remoteAddr, recordType}).AllCols().Update(record)
	} else {
		_, err = adapter.Engine.Insert(record)
	}

	if err != nil {
		panic(err)
	}

	return ""
}

func CheckVerificationCode(dest, code string) string {
	var record VerificationRecord
	record.Receiver = dest
	has, err := adapter.Engine.Desc("time").Where("is_used = 0").Get(&record)
	if err != nil {
		panic(err)
	}

	if !has {
		return "Code has not been sent yet!"
	}

	now := time.Now().Unix()
	if now-record.Time > 5*60 {
		return "You should verify your code in 5 min!"
	}

	if record.Code != code {
		return "Wrong code!"
	}

	record.IsUsed = true
	_, err = adapter.Engine.ID(core.PK{record.RemoteAddr, record.Type}).AllCols().Update(record)
	if err != nil {
		panic(err)
	}

	return ""
}

// from Casnode/object/validateCode.go line 116
var stdNums = []byte("0123456789")

func getRandomCode(length int) string {
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, stdNums[r.Intn(len(stdNums))])
	}
	return string(result)
}
