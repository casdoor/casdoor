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
	"math"
	"math/rand"
	"strings"
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
	VerificationSuccess = iota
	wrongCodeError
	noRecordError
	timeoutError
)

const (
	VerifyTypePhone = "phone"
	VerifyTypeEmail = "email"
)

type VerificationRecord struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	RemoteAddr string `xorm:"varchar(100)" json:"remoteAddr"`
	Type       string `xorm:"varchar(10)" json:"type"`
	User       string `xorm:"varchar(100) notnull" json:"user"`
	Provider   string `xorm:"varchar(100) notnull" json:"provider"`
	Receiver   string `xorm:"varchar(100) index notnull" json:"receiver"`
	Code       string `xorm:"varchar(10) notnull" json:"code"`
	Time       int64  `xorm:"notnull" json:"time"`
	IsUsed     bool   `xorm:"notnull" json:"isUsed"`
}

func IsAllowSend(user *User, remoteAddr, recordType string) error {
	var record VerificationRecord
	record.RemoteAddr = remoteAddr
	record.Type = recordType
	if user != nil {
		record.User = user.GetId()
	}

	has, err := ormer.Engine.Desc("created_time").Get(&record)
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
	sender := organization.DisplayName
	title := provider.Title

	code := getRandomCode(6)
	// if organization.MasterVerificationCode != "" {
	//	code = organization.MasterVerificationCode
	// }

	// "You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes."
	content := strings.Replace(provider.Content, "%s", code, 1)

	userString := "Hi"
	if user != nil {
		userString = user.GetFriendlyName()
	}
	content = strings.Replace(content, "%{user.friendlyName}", userString, 1)

	err := IsAllowSend(user, remoteAddr, provider.Category)
	if err != nil {
		return err
	}

	err = SendEmail(provider, title, content, dest, sender)
	if err != nil {
		return err
	}

	err = AddToVerificationRecord(user, provider, remoteAddr, provider.Category, dest, code)
	if err != nil {
		return err
	}

	return nil
}

func SendVerificationCodeToPhone(organization *Organization, user *User, provider *Provider, remoteAddr string, dest string) error {
	err := IsAllowSend(user, remoteAddr, provider.Category)
	if err != nil {
		return err
	}

	code := getRandomCode(6)
	// if organization.MasterVerificationCode != "" {
	//	code = organization.MasterVerificationCode
	// }

	err = SendSms(provider, code, dest)
	if err != nil {
		return err
	}

	err = AddToVerificationRecord(user, provider, remoteAddr, provider.Category, dest, code)
	if err != nil {
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

	_, err := ormer.Engine.Insert(record)
	if err != nil {
		return err
	}

	return nil
}

func filterRecordIn24Hours(record *VerificationRecord) *VerificationRecord {
	if record == nil {
		return nil
	}

	now := time.Now().Unix()
	if now-record.Time > 60*60*24 {
		return nil
	}

	return record
}

func getVerificationRecord(dest string) (*VerificationRecord, error) {
	record := &VerificationRecord{}
	record.Receiver = dest

	has, err := ormer.Engine.Desc("time").Where("is_used = false").Get(record)
	if err != nil {
		return nil, err
	}

	record = filterRecordIn24Hours(record)
	if record == nil {
		has = false
	}

	if !has {
		record = &VerificationRecord{}
		record.Receiver = dest

		has, err = ormer.Engine.Desc("time").Get(record)
		if err != nil {
			return nil, err
		}

		record = filterRecordIn24Hours(record)
		if record == nil {
			has = false
		}

		if !has {
			return nil, nil
		}

		return record, nil
	}

	return record, nil
}

func getUnusedVerificationRecord(dest string) (*VerificationRecord, error) {
	record := &VerificationRecord{}
	record.Receiver = dest

	has, err := ormer.Engine.Desc("time").Where("is_used = false").Get(record)
	if err != nil {
		return nil, err
	}

	record = filterRecordIn24Hours(record)
	if record == nil {
		has = false
	}

	if !has {
		return nil, nil
	}

	return record, nil
}

func CheckVerificationCode(dest string, code string, lang string) (*VerifyResult, error) {
	record, err := getVerificationRecord(dest)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &VerifyResult{noRecordError, i18n.Translate(lang, "verification:The verification code has not been sent yet!")}, nil
	} else if record.IsUsed {
		return &VerifyResult{noRecordError, i18n.Translate(lang, "verification:The verification code has already been used!")}, nil
	}

	timeoutInMinutes, err := conf.GetConfigInt64("verificationCodeTimeout")
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	if now-record.Time > timeoutInMinutes*60 {
		return &VerifyResult{timeoutError, fmt.Sprintf(i18n.Translate(lang, "verification:You should verify your code in %d min!"), timeoutInMinutes)}, nil
	}

	if record.Code != code {
		return &VerifyResult{wrongCodeError, i18n.Translate(lang, "verification:Wrong verification code!")}, nil
	}

	return &VerifyResult{VerificationSuccess, ""}, nil
}

func DisableVerificationCode(dest string) error {
	record, err := getUnusedVerificationRecord(dest)
	if record == nil || err != nil {
		return nil
	}

	record.IsUsed = true
	_, err = ormer.Engine.ID(core.PK{record.Owner, record.Name}).AllCols().Update(record)
	return err
}

func CheckSigninCode(user *User, dest, code, lang string) error {
	// check the login error times
	err := checkSigninErrorTimes(user, lang)
	if err != nil {
		return err
	}

	result, err := CheckVerificationCode(dest, code, lang)
	if err != nil {
		return err
	}

	switch result.Code {
	case VerificationSuccess:
		return resetUserSigninErrorTimes(user)
	case wrongCodeError:
		return recordSigninErrorInfo(user, lang)
	default:
		return fmt.Errorf(result.Msg)
	}
}

func CheckFaceId(user *User, faceId []float64, lang string) error {
	if len(user.FaceIds) == 0 {
		return fmt.Errorf(i18n.Translate(lang, "check:Face data does not exist, cannot log in"))
	}

	for _, userFaceId := range user.FaceIds {
		if faceId == nil || len(userFaceId.FaceIdData) != len(faceId) {
			continue
		}
		var sumOfSquares float64
		for i := 0; i < len(userFaceId.FaceIdData); i++ {
			diff := userFaceId.FaceIdData[i] - faceId[i]
			sumOfSquares += diff * diff
		}
		if math.Sqrt(sumOfSquares) < 0.25 {
			return nil
		}
	}

	return fmt.Errorf(i18n.Translate(lang, "check:Face data mismatch"))
}

func GetVerifyType(username string) (verificationCodeType string) {
	if strings.Contains(username, "@") {
		return VerifyTypeEmail
	} else {
		return VerifyTypePhone
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

func GetVerificationCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&VerificationRecord{Owner: owner})
}

func GetVerifications(owner string) ([]*VerificationRecord, error) {
	verifications := []*VerificationRecord{}
	err := ormer.Engine.Desc("created_time").Find(&verifications, &VerificationRecord{Owner: owner})
	if err != nil {
		return nil, err
	}

	return verifications, nil
}

func GetUserVerifications(owner, user string) ([]*VerificationRecord, error) {
	verifications := []*VerificationRecord{}
	err := ormer.Engine.Desc("created_time").Find(&verifications, &VerificationRecord{Owner: owner, User: user})
	if err != nil {
		return nil, err
	}

	return verifications, nil
}

func GetPaginationVerifications(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*VerificationRecord, error) {
	verifications := []*VerificationRecord{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&verifications, &VerificationRecord{Owner: owner})
	if err != nil {
		return nil, err
	}

	return verifications, nil
}

func getVerification(owner string, name string) (*VerificationRecord, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	verification := VerificationRecord{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&verification)
	if err != nil {
		return nil, err
	}

	if existed {
		return &verification, nil
	} else {
		return nil, nil
	}
}

func GetVerification(id string) (*VerificationRecord, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getVerification(owner, name)
}
