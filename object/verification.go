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

package object

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type VerifyResult struct {
	Code int
	Msg  string
}

const (
	VerificationSuccess int = 0
	wrongCodeError          = 1
	noRecordError           = 2
	timeoutError            = 3
)

type VerificationRecord struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	RemoteAddr string `xorm:"varchar(100)"`
	Type       string `xorm:"varchar(10)"`
	User       string `xorm:"varchar(100) notnull"`
	Provider   string `xorm:"varchar(100) notnull"`
	Receiver   string `xorm:"varchar(100) notnull"`
	Code       string `xorm:"varchar(10) notnull"`
	Time       int64  `xorm:"notnull"`
	IsUsed     bool
}

func IsAllowSend(user *User, remoteAddr, recordType string) error {
	var record VerificationRecord
	record.RemoteAddr = remoteAddr
	record.Type = recordType
	if user != nil {
		record.User = user.GetId()
	}
	has, err := adapter.Engine.Desc("created_time").Get(&record)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	if has && now-record.Time < 60 {
		return errors.New("you can only send one code in 60s")
	}

	return nil
}

func SendVerificationCodeToEmail(organization *Organization, user *User, provider *Provider, remoteAddr string, dest string) error {
	if provider == nil {
		return fmt.Errorf("please set an Email provider first")
	}

	sender := organization.DisplayName
	title := provider.Title
	code := getRandomCode(6)
	// "You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes."
	content := fmt.Sprintf(provider.Content, code)

	if err := IsAllowSend(user, remoteAddr, provider.Category); err != nil {
		return err
	}

	if err := SendEmail(provider, title, content, dest, sender); err != nil {
		return err
	}

	if err := AddToVerificationRecord(user, provider, remoteAddr, provider.Category, dest, code); err != nil {
		return err
	}

	return nil
}

func SendVerificationCodeToPhone(organization *Organization, user *User, provider *Provider, remoteAddr string, dest string) error {
	if provider == nil {
		return errors.New("please set a SMS provider first")
	}

	if err := IsAllowSend(user, remoteAddr, provider.Category); err != nil {
		return err
	}

	code := getRandomCode(6)
	if err := SendSms(provider, code, dest); err != nil {
		return err
	}

	if err := AddToVerificationRecord(user, provider, remoteAddr, provider.Category, dest, code); err != nil {
		return err
	}

	return nil
}

func AddToVerificationRecord(user *User, provider *Provider, remoteAddr, recordType, dest, code string) error {
	var record VerificationRecord
	record.RemoteAddr = remoteAddr
	record.Type = recordType
	if user != nil {
		record.User = user.GetId()
	}
	record.Owner = provider.Owner
	record.Name = util.GenerateId()
	record.CreatedTime = util.GetCurrentTime()

	record.Provider = provider.Name
	record.Receiver = dest
	record.Code = code
	record.Time = time.Now().Unix()
	record.IsUsed = false

	_, err := adapter.Engine.Insert(record)
	if err != nil {
		return err
	}

	return nil
}

func getVerificationRecord(dest string) *VerificationRecord {
	var record VerificationRecord
	record.Receiver = dest
	has, err := adapter.Engine.Desc("time").Where("is_used = false").Get(&record)
	if err != nil {
		panic(err)
	}
	if !has {
		return nil
	}
	return &record
}

func CheckVerificationCode(dest, code, lang string) *VerifyResult {
	record := getVerificationRecord(dest)

	if record == nil {
		return &VerifyResult{noRecordError, i18n.Translate(lang, "verification:Code has not been sent yet!")}
	}

	timeout, err := conf.GetConfigInt64("verificationCodeTimeout")
	if err != nil {
		panic(err)
	}

	now := time.Now().Unix()
	if now-record.Time > timeout*60 {
		return &VerifyResult{timeoutError, fmt.Sprintf(i18n.Translate(lang, "verification:You should verify your code in %d min!"), timeout)}
	}

	if record.Code != code {
		return &VerifyResult{wrongCodeError, i18n.Translate(lang, "verification:Wrong verification code!")}
	}

	return &VerifyResult{VerificationSuccess, ""}
}

func DisableVerificationCode(dest string) {
	record := getVerificationRecord(dest)
	if record == nil {
		return
	}

	record.IsUsed = true
	_, err := adapter.Engine.ID(core.PK{record.Owner, record.Name}).AllCols().Update(record)
	if err != nil {
		panic(err)
	}
}

func CheckSigninCode(user *User, dest, code, lang string) string {
	// check the login error times
	if msg := checkSigninErrorTimes(user, lang); msg != "" {
		return msg
	}

	result := CheckVerificationCode(dest, code, lang)
	switch result.Code {
	case VerificationSuccess:
		resetUserSigninErrorTimes(user)
		return ""
	case wrongCodeError:
		return recordSigninErrorInfo(user, lang)
	default:
		return result.Msg
	}
}

// From Casnode/object/validateCode.go line 116
var stdNums = []byte("0123456789")

func getRandomCode(length int) string {
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, stdNums[r.Intn(len(stdNums))])
	}
	return string(result)
}
