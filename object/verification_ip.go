// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"crypto/subtle"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/casdoor/casdoor/i18n"
)

// Hard-coded thresholds for OTP / verification-code brute force protection (per IP + dest).
// These can be made configurable later if needed.
const (
	defaultVerifyCodeIpLimit        = 5
	defaultVerifyCodeIpFrozenMinute = 10
)

var (
	verifyCodeIpErrorMap     = map[string]*verifyCodeErrorInfo{}
	verifyCodeIpErrorMapLock sync.Mutex
)

func getVerifyCodeIpErrorKey(remoteAddr, dest string) string {
	return fmt.Sprintf("%s:%s", remoteAddr, dest)
}

func checkVerifyCodeIpErrorTimes(remoteAddr, dest, lang string) error {
	if remoteAddr == "" {
		return nil
	}

	key := getVerifyCodeIpErrorKey(remoteAddr, dest)

	verifyCodeIpErrorMapLock.Lock()
	defer verifyCodeIpErrorMapLock.Unlock()

	errorInfo, ok := verifyCodeIpErrorMap[key]
	if !ok || errorInfo == nil {
		return nil
	}

	if errorInfo.wrongTimes < defaultVerifyCodeIpLimit {
		return nil
	}

	minutesLeft := int64(defaultVerifyCodeIpFrozenMinute) - int64(time.Now().UTC().Sub(errorInfo.lastWrongTime).Minutes())
	if minutesLeft > 0 {
		return fmt.Errorf(i18n.Translate(lang, "check:You have entered the wrong password or code too many times, please wait for %d minutes and try again"), minutesLeft)
	}

	delete(verifyCodeIpErrorMap, key)
	return nil
}

func recordVerifyCodeIpErrorInfo(remoteAddr, dest, lang string) error {
	// If remoteAddr is missing, still return a normal "wrong code" error.
	if remoteAddr == "" {
		return errors.New(i18n.Translate(lang, "verification:Wrong verification code!"))
	}

	key := getVerifyCodeIpErrorKey(remoteAddr, dest)

	verifyCodeIpErrorMapLock.Lock()
	defer verifyCodeIpErrorMapLock.Unlock()

	errorInfo, ok := verifyCodeIpErrorMap[key]
	if !ok || errorInfo == nil {
		errorInfo = &verifyCodeErrorInfo{}
		verifyCodeIpErrorMap[key] = errorInfo
	}

	if errorInfo.wrongTimes < defaultVerifyCodeIpLimit {
		errorInfo.wrongTimes++
	}

	if errorInfo.wrongTimes >= defaultVerifyCodeIpLimit {
		errorInfo.lastWrongTime = time.Now().UTC()
	}

	leftChances := defaultVerifyCodeIpLimit - errorInfo.wrongTimes
	if leftChances >= 0 {
		return fmt.Errorf(i18n.Translate(lang, "check:password or code is incorrect, you have %s remaining chances"), strconv.Itoa(leftChances))
	}

	return fmt.Errorf(i18n.Translate(lang, "check:You have entered the wrong password or code too many times, please wait for %d minutes and try again"), defaultVerifyCodeIpFrozenMinute)
}

func resetVerifyCodeIpErrorTimes(remoteAddr, dest string) {
	if remoteAddr == "" {
		return
	}

	key := getVerifyCodeIpErrorKey(remoteAddr, dest)

	verifyCodeIpErrorMapLock.Lock()
	defer verifyCodeIpErrorMapLock.Unlock()

	delete(verifyCodeIpErrorMap, key)
}

// CheckVerifyCodeWithLimitAndIp enforces both per-user and per-IP attempt limits for verification codes.
// It is intended for security-sensitive flows like password reset.
func CheckVerifyCodeWithLimitAndIp(user *User, remoteAddr, dest, code, lang string) error {
	_, err := CheckVerifyCodeOrMasterCodeWithLimitAndIp(user, "", remoteAddr, dest, code, lang)
	return err
}

func CheckVerifyCodeOrMasterCodeWithLimitAndIp(user *User, masterCode, remoteAddr, dest, code, lang string) (bool, error) {
	if err := checkVerifyCodeIpErrorTimes(remoteAddr, dest, lang); err != nil {
		return false, err
	}

	if user != nil {
		if err := checkVerifyCodeErrorTimes(user, dest, lang); err != nil {
			return false, err
		}
	}

	if masterCode != "" && subtle.ConstantTimeCompare([]byte(code), []byte(masterCode)) == 1 {
		resetVerifyCodeLimit(user, remoteAddr, dest)
		return true, nil
	}

	result, err := CheckVerificationCode(dest, code, lang)
	if err != nil {
		return false, err
	}

	isMasterCodeMiss := masterCode != "" && result.Code != VerificationSuccess
	switch {
	case result.Code == VerificationSuccess:
		resetVerifyCodeLimit(user, remoteAddr, dest)
		return false, nil
	case result.Code == wrongCodeError || isMasterCodeMiss:
		ipErr := recordVerifyCodeIpErrorInfo(remoteAddr, dest, lang)
		if user != nil {
			// Keep existing user-level error semantics when user is known.
			return false, recordVerifyCodeErrorInfo(user, dest, lang)
		}
		return false, ipErr
	default:
		return false, errors.New(result.Msg)
	}
}

func resetVerifyCodeLimit(user *User, remoteAddr, dest string) {
	resetVerifyCodeIpErrorTimes(remoteAddr, dest)
	if user != nil {
		resetVerifyCodeErrorTimes(user, dest)
	}
}
