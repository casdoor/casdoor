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
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/casdoor/casdoor/i18n"
)

var reRealName *regexp.Regexp

func init() {
	var err error
	reRealName, err = regexp.Compile("^[\u4E00-\u9FA5]{2,3}(?:·[\u4E00-\u9FA5]{2,3})*$")
	if err != nil {
		panic(err)
	}
}

func isValidRealName(s string) bool {
	return reRealName.MatchString(s)
}

func resetUserSigninErrorTimes(user *User) error {
	// if the password is correct and wrong times is not zero, reset the error times
	if user.SigninWrongTimes == 0 {
		return nil
	}

	user.SigninWrongTimes = 0
	_, err := UpdateUser(user.GetId(), user, []string{"signin_wrong_times", "last_signin_wrong_time"}, false)
	return err
}

func GetFailedSigninConfigByUser(user *User) (int, int, error) {
	application, err := GetApplicationByUser(user)
	if err != nil {
		return 0, 0, err
	}
	if application == nil {
		return 0, 0, fmt.Errorf("the application for user %s is not found", user.GetId())
	}

	failedSigninLimit := application.FailedSigninLimit
	if failedSigninLimit == 0 {
		failedSigninLimit = DefaultFailedSigninLimit
	}

	failedSigninFrozenTime := application.FailedSigninFrozenTime
	if failedSigninFrozenTime == 0 {
		failedSigninFrozenTime = DefaultFailedSigninFrozenTime
	}

	return failedSigninLimit, failedSigninFrozenTime, nil
}

func recordSigninErrorInfo(user *User, lang string, options ...bool) error {
	enableCaptcha := false
	if len(options) > 0 {
		enableCaptcha = options[0]
	}

	failedSigninLimit, failedSigninFrozenTime, errSignin := GetFailedSigninConfigByUser(user)
	if errSignin != nil {
		return errSignin
	}

	// increase failed login count
	if user.SigninWrongTimes < failedSigninLimit {
		user.SigninWrongTimes++
	}

	if user.SigninWrongTimes >= failedSigninLimit {
		// record the latest failed login time
		user.LastSigninWrongTime = time.Now().UTC().Format(time.RFC3339)
	}

	// update user
	_, err := UpdateUser(user.GetId(), user, []string{"signin_wrong_times", "last_signin_wrong_time"}, false)
	if err != nil {
		return err
	}

	leftChances := failedSigninLimit - user.SigninWrongTimes
	if leftChances == 0 && enableCaptcha {
		return fmt.Errorf(i18n.Translate(lang, "check:password or code is incorrect"))
	} else if leftChances >= 0 {
		return fmt.Errorf(i18n.Translate(lang, "check:password or code is incorrect, you have %s remaining chances"), strconv.Itoa(leftChances))
	}

	// don't show the chance error message if the user has no chance left
	return fmt.Errorf(i18n.Translate(lang, "check:You have entered the wrong password or code too many times, please wait for %d minutes and try again"), failedSigninFrozenTime)
}
