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

	"github.com/astaxie/beego"
	"xorm.io/core"
)

type VerificationRecord struct {
	RemoteAddr string `xorm:"varchar(100) notnull pk"`
	Type       string `xorm:"varchar(10) notnull pk"`
	Provider   string `xorm:"varchar(100) notnull"`
	Receiver   string `xorm:"varchar(100) notnull"`
	Code       string `xorm:"varchar(10) notnull"`
	Time       int64  `xorm:"notnull"`
	IsUsed     bool
}

func SendVerificationCodeToEmail(provider *Provider, remoteAddr string, dest string) string {
	if provider == nil {
		return "Please set an Email provider first"
	}

	title := "Casdoor Verification Code"
	sender := "Casdoor"
	code := getRandomCode(5)
	content := fmt.Sprintf("You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes.", code)

	if result := AddToVerificationRecord(provider.Name, remoteAddr, "Email", dest, code); len(result) != 0 {
		return result
	}

	msg, err := SendEmail(provider, title, content, dest, sender)
	if msg != "" {
		return msg
	}
	if err != nil {
		panic(err)
	}

	return ""
}

func SendVerificationCodeToPhone(provider *Provider, remoteAddr string, dest string) string {
	if provider == nil {
		return "Please set a SMS provider first"
	}

	code := getRandomCode(5)
	if result := AddToVerificationRecord(provider.Name, remoteAddr, "SMS", dest, code); len(result) != 0 {
		return result
	}

	return SendCodeToPhone(provider, dest, code)
}

func AddToVerificationRecord(providerName, remoteAddr, recordType, dest, code string) string {
	var record VerificationRecord
	record.RemoteAddr = remoteAddr
	record.Type = recordType
	record.Provider = providerName
	has, err := adapter.Engine.Get(&record)
	if err != nil {
		panic(err)
	}

	now := time.Now().Unix()

	if has && now-record.Time < 60 {
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

func getVerificationRecord(dest string) *VerificationRecord {
	var record VerificationRecord
	record.Receiver = dest
	has, err := adapter.Engine.Desc("time").Where("is_used = 0").Get(&record)
	if err != nil {
		panic(err)
	}
	if !has {
		return nil
	}
	return &record
}

func CheckVerificationCode(dest, code string) string {
	record := getVerificationRecord(dest)

	if record == nil {
		return "Code has not been sent yet!"
	}

	timeout, err := beego.AppConfig.Int64("verificationCodeTimeout")
	if err != nil {
		panic(err)
	}

	now := time.Now().Unix()
	if now-record.Time > timeout*60 {
		return fmt.Sprintf("You should verify your code in %d min!", timeout)
	}

	if record.Code != code {
		return "Wrong code!"
	}

	return ""
}

func DisableVerificationCode(dest string) {
	record := getVerificationRecord(dest)
	if record == nil {
		return
	}

	record.IsUsed = true
	_, err := adapter.Engine.ID(core.PK{record.RemoteAddr, record.Type}).AllCols().Update(record)
	if err != nil {
		panic(err)
	}
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
