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

package object

import (
	"regexp"
)

type ValidatorFunc func(password string) string

var (
	regexLowerCase = regexp.MustCompile(`[a-z]`)
	regexUpperCase = regexp.MustCompile(`[A-Z]`)
	regexDigit     = regexp.MustCompile(`\d`)
	regexSpecial   = regexp.MustCompile(`[!@#$%^&*]`)
)

func isValidOption_AtLeast6(password string) string {
	if len(password) < 6 {
		return "The password must have at least 6 characters"
	}
	return ""
}

func isValidOption_AtLeast8(password string) string {
	if len(password) < 8 {
		return "The password must have at least 8 characters"
	}
	return ""
}

func isValidOption_Aa123(password string) string {
	hasLowerCase := regexLowerCase.MatchString(password)
	hasUpperCase := regexUpperCase.MatchString(password)
	hasDigit := regexDigit.MatchString(password)

	if !hasLowerCase || !hasUpperCase || !hasDigit {
		return "The password must contain at least one uppercase letter, one lowercase letter and one digit"
	}
	return ""
}

func isValidOption_SpecialChar(password string) string {
	if !regexSpecial.MatchString(password) {
		return "The password must contain at least one special character"
	}
	return ""
}

func isValidOption_NoRepeat(password string) string {
	for i := 0; i < len(password)-1; i++ {
		if password[i] == password[i+1] {
			return "The password must not contain any repeated characters"
		}
	}
	return ""
}

func checkPasswordComplexity(password string, options []string) string {
	if len(password) == 0 {
		return "Please input your password!"
	}

	if len(options) == 0 {
		options = []string{"AtLeast6"}
	}

	checkers := map[string]ValidatorFunc{
		"AtLeast6":    isValidOption_AtLeast6,
		"AtLeast8":    isValidOption_AtLeast8,
		"Aa123":       isValidOption_Aa123,
		"SpecialChar": isValidOption_SpecialChar,
		"NoRepeat":    isValidOption_NoRepeat,
	}

	for _, option := range options {
		checkerFunc, ok := checkers[option]
		if ok {
			errorMsg := checkerFunc(password)
			if errorMsg != "" {
				return errorMsg
			}
		}
	}
	return ""
}
