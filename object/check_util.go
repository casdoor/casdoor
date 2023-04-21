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
	"time"

	"github.com/casdoor/casdoor/i18n"
)

type CheckPassWordOptions struct {
	enableCaptcha bool
}

type CheckPassWordOptionFunc func(*CheckPassWordOptions)

func WithEnableCaptcha() CheckPassWordOptionFunc {
	return func(ops *CheckPassWordOptions) {
		ops.enableCaptcha = true
	}
}

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

func resetUserSigninErrorTimes(user *User) {
	// if the password is correct and wrong times is not zero, reset the error times
	if user.SigninWrongTimes == 0 {
		return
	}
	user.SigninWrongTimes = 0
	UpdateUser(user.GetId(), user, []string{"signin_wrong_times", "last_signin_wrong_time"}, user.IsGlobalAdmin)
}

func recordSigninErrorInfo(user *User, lang string, enableCaptcha bool) string {
	// increase failed login count
	if user.SigninWrongTimes < SigninWrongTimesLimit {
		user.SigninWrongTimes++
	}

	if user.SigninWrongTimes >= SigninWrongTimesLimit {
		// record the latest failed login time
		user.LastSigninWrongTime = time.Now().UTC().Format(time.RFC3339)
	}

	// update user
	UpdateUser(user.GetId(), user, []string{"signin_wrong_times", "last_signin_wrong_time"}, user.IsGlobalAdmin)
	leftChances := SigninWrongTimesLimit - user.SigninWrongTimes
	if leftChances > 0 {
		return fmt.Sprintf(i18n.Translate(lang, "check:password or code is incorrect, you have %d remaining chances"), leftChances)
	} else if enableCaptcha {
		return fmt.Sprint(i18n.Translate(lang, "check:password or code is incorrect"))
	}
	// don't show the chance error message if the user has no chance left
	return fmt.Sprintf(i18n.Translate(lang, "check:You have entered the wrong password or code too many times, please wait for %d minutes and try again"), int(LastSignWrongTimeDuration.Minutes()))
}
