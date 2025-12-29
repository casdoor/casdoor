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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	mathrand "math/rand"
	"net/url"
	"regexp"
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

var ResetLinkReg *regexp.Regexp

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

func init() {
	ResetLinkReg = regexp.MustCompile("(?s)<reset-link>(.*?)</reset-link>")
}

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

func IsAllowSend(user *User, remoteAddr, recordType string, application *Application) error {
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

	// Get timeout from application, or use default
	resendTimeoutInSeconds := int64(60)
	if application != nil && application.CodeResendTimeout > 0 {
		resendTimeoutInSeconds = int64(application.CodeResendTimeout)
	}

	now := time.Now().Unix()
	if has && now-record.Time < resendTimeoutInSeconds {
		return fmt.Errorf("you can only send one code in %ds", resendTimeoutInSeconds)
	}

	return nil
}

func SendVerificationCodeToEmail(organization *Organization, user *User, provider *Provider, remoteAddr string, dest string, method string, host string, applicationName string, application *Application) error {
	sender := organization.DisplayName
	title := provider.Title

	code := getRandomCode(6)
	// if organization.MasterVerificationCode != "" {
	//	code = organization.MasterVerificationCode
	// }

	// "You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes."
	content := strings.Replace(provider.Content, "%s", code, 1)

	if method == "forget" {
		originFrontend, _ := getOriginFromHost(host)

		// Check if the email template contains <reset-link> tags for magic link
		matchContent := ResetLinkReg.Find([]byte(provider.Content))
		if matchContent != nil {
			// Generate a secure token for magic link
			token, err := generateSecureToken(32)
			if err != nil {
				return err
			}

			// Create magic link URL with token
			query := url.Values{}
			query.Add("token", token)
			query.Add("username", user.Name)
			query.Add("dest", util.GetMaskedEmail(dest))
			magicLinkURL := originFrontend + "/forget/" + applicationName + "?" + query.Encode()

			// Replace %link with the magic link URL
			content = strings.Replace(provider.Content, "%link", magicLinkURL, -1)
			// Extract and replace the content within <reset-link> tags
			submatches := ResetLinkReg.FindSubmatch([]byte(provider.Content))
			if submatches != nil && len(submatches) > 1 {
				linkContent := string(submatches[1])
				linkHTML := fmt.Sprintf("<a href=\"%s\">%s</a>", magicLinkURL, linkContent)
				content = ResetLinkReg.ReplaceAllString(content, linkHTML)
			}
			// Replace %s with the code in case it's also in the template
			content = strings.Replace(content, "%s", code, 1)

			// Store the token as the code in the verification record
			code = token
		} else {
			// Fallback to original verification code flow
			query := url.Values{}
			query.Add("code", code)
			query.Add("username", user.Name)
			query.Add("dest", util.GetMaskedEmail(dest))
			forgetURL := originFrontend + "/forget/" + applicationName + "?" + query.Encode()

			content = strings.Replace(content, "%link", forgetURL, -1)
			content = strings.Replace(content, "<reset-link>", "", -1)
			content = strings.Replace(content, "</reset-link>", "", -1)
		}
	} else {
		matchContent := ResetLinkReg.Find([]byte(content))
		content = strings.Replace(content, string(matchContent), "", -1)
	}

	userString := "Hi"
	if user != nil {
		userString = user.GetFriendlyName()
	}
	content = strings.Replace(content, "%{user.friendlyName}", userString, 1)

	err := IsAllowSend(user, remoteAddr, provider.Category, application)
	if err != nil {
		return err
	}

	err = SendEmail(provider, title, content, []string{dest}, sender)
	if err != nil {
		return err
	}

	err = AddToVerificationRecord(user, provider, organization, remoteAddr, provider.Category, dest, code)
	if err != nil {
		return err
	}

	return nil
}

func SendVerificationCodeToPhone(organization *Organization, user *User, provider *Provider, remoteAddr string, dest string, application *Application) error {
	err := IsAllowSend(user, remoteAddr, provider.Category, application)
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

	err = AddToVerificationRecord(user, provider, organization, remoteAddr, provider.Category, dest, code)
	if err != nil {
		return err
	}

	return nil
}

func AddToVerificationRecord(user *User, provider *Provider, organization *Organization, remoteAddr, recordType, dest, code string) error {
	var record VerificationRecord
	record.RemoteAddr = remoteAddr
	record.Type = recordType
	if user != nil {
		record.User = user.GetId()
	}
	record.Owner = organization.Name
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

// CheckVerificationToken validates a magic link token
func CheckVerificationToken(token string, lang string) (*VerifyResult, *VerificationRecord, error) {
	record := &VerificationRecord{}
	record.Code = token

	has, err := ormer.Engine.Desc("time").Where("is_used = false").Get(record)
	if err != nil {
		return nil, nil, err
	}

	record = filterRecordIn24Hours(record)
	if record == nil {
		has = false
	}

	if !has {
		return &VerifyResult{noRecordError, i18n.Translate(lang, "verification:The verification link is invalid or has expired!")}, nil, nil
	}

	if record.IsUsed {
		return &VerifyResult{noRecordError, i18n.Translate(lang, "verification:The verification link has already been used!")}, nil, nil
	}

	// Token-based verification uses 24-hour expiration (already filtered above)
	// This provides a longer validity period than the default code timeout

	return &VerifyResult{VerificationSuccess, ""}, record, nil
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

// DisableVerificationToken marks a verification token as used
func DisableVerificationToken(token string) error {
	record := &VerificationRecord{}
	record.Code = token

	has, err := ormer.Engine.Desc("time").Where("is_used = false").Get(record)
	if err != nil {
		return err
	}

	if !has {
		// Record not found, which is acceptable - may have already been disabled
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
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, stdNums[r.Intn(len(stdNums))])
	}
	return string(result)
}

// generateSecureToken generates a cryptographically secure random token for magic links
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Use URL-safe base64 encoding and remove padding
	return base64.URLEncoding.EncodeToString(bytes), nil
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
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getVerification(owner, name)
}
