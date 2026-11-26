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

package i18n

import (
	"regexp"
	"strings"
	"sync"
)

// errorCodes gives common business errors a stable code in the "code" field of the
// API response, so that clients can tell them apart without matching translated text
var errorCodes = map[string]string{
	"check:Username already exists":                              "username-exists",
	"check:Email already exists":                                 "email-exists",
	"check:Phone already exists":                                 "phone-exists",
	"general:The users already exist: %s":                        "user-exists",
	"general:The user: %s in line %d is duplicated with line %d": "user-exists",
	"check:The account has already linked another %s account":    "account-already-linked",
	"auth:The account for provider: %s and username: %s (%s) is already linked to another account: %s (%s)": "account-already-linked",

	"check:Empty username.":                                   "username-required",
	"check:Username cannot be an email address":               "username-invalid",
	"check:Username cannot contain white spaces":              "username-invalid",
	"check:Username cannot start with a digit":                "username-invalid",
	"check:Username is too long (maximum is 255 characters).": "username-invalid",
	"check:Username must have at least 2 characters":          "username-invalid",
	"check:The username may only contain alphanumeric characters, underlines or hyphens, cannot have consecutive hyphens or underlines, and cannot begin or end with a hyphen or underline.":                                                                              "username-invalid",
	"check:Username supports email format. Also The username may only contain alphanumeric characters, underlines or hyphens, cannot have consecutive hyphens or underlines, and cannot begin or end with a hyphen or underline. Also pay attention to the email format.": "username-invalid",
	"check:Email cannot be empty":                                    "email-required",
	"check:Email is invalid":                                         "email-invalid",
	"check:The email domain: %s is not allowed to sign up":           "email-domain-not-allowed",
	"check:Phone cannot be empty":                                    "phone-required",
	"check:Phone number is invalid":                                  "phone-invalid",
	"verification:Phone number is invalid in your region %s":         "phone-invalid",
	"check:Your region is not allow to signup by phone":              "phone-region-not-allowed",
	"check:Password cannot be empty":                                 "password-required",
	"check:The password is not allowed: %s":                          "password-invalid",
	"check:The password must contain at least one special character": "password-invalid",
	"check:The password must contain at least one uppercase letter, one lowercase letter and one digit":            "password-invalid",
	"check:The password must have at least 6 characters":                                                           "password-invalid",
	"check:The password must have at least 8 characters":                                                           "password-invalid",
	"check:The password must not contain any repeated characters":                                                  "password-invalid",
	"user:New password cannot contain blank space.":                                                                "password-invalid",
	"user:The new password must be different from your current password":                                           "password-reused",
	"user:The new password must be different from your last %d passwords":                                          "password-reused",
	"check:DisplayName cannot be blank":                                                                            "field-required",
	"check:FirstName cannot be blank":                                                                              "field-required",
	"check:LastName cannot be blank":                                                                               "field-required",
	"check:Affiliation cannot be blank":                                                                            "field-required",
	"user:Display name cannot be empty":                                                                            "field-required",
	"check:DisplayName is not valid real name":                                                                     "field-invalid",
	"check:The value \"%s\" for account field \"%s\" doesn't match the account item regex":                         "field-invalid",
	"check:The value \"%s\" for signup field \"%s\" doesn't match the signup item regex of the application \"%s\"": "field-invalid",
	"general:Missing parameter":                                                                                    "parameter-required",

	"check:Invitation code cannot be blank":                                         "invitation-code-required",
	"check:Invitation code is invalid":                                              "invitation-code-invalid",
	"check:Invitation code expired":                                                 "invitation-code-expired",
	"check:Invitation code exhausted":                                               "invitation-code-used",
	"check:The invitation code has already been used":                               "invitation-code-used",
	"check:Invitation code suspended":                                               "invitation-code-suspended",
	"check:Please register using the email corresponding to the invitation code":    "invitation-code-mismatch",
	"check:Please register using the phone  corresponding to the invitation code":   "invitation-code-mismatch",
	"check:Please register using the username corresponding to the invitation code": "invitation-code-mismatch",
	"account:The application does not allow to sign up new account":                 "signup-disabled",

	"check:password or code is incorrect":                                                                        "wrong-password",
	"check:password or code is incorrect, you have %s remaining chances":                                         "wrong-password",
	"check:You have entered the wrong password or code too many times, please wait for %d minutes and try again": "account-frozen",
	"check:Too many accounts failed to sign in from your IP address, please wait for %d minutes and try again":   "ip-frozen",
	"check:Your IP address: %s has been banned according to the configuration of: ":                              "ip-banned",
	"check:The user is forbidden to sign in, please contact the administrator":                                   "account-disabled",
	"check:The user has been deleted and cannot be used to sign in, please contact the administrator":            "account-deleted",
	"check:Your account is waiting for the administrator's approval":                                             "account-pending-approval",
	"check:Your password has expired. Please reset your password by clicking \"Forgot password\"":                "password-expired",
	"general:The user: %s doesn't exist":                                                                         "user-not-found",
	"general:The user doesn't exist":                                                                             "user-not-found",
	"general:The user: %s is not found":                                                                          "user-not-found",
	"general:User not found":                                                                                     "user-not-found",
	"verification:the user does not exist, please sign up first":                                                 "user-not-found",
	"auth:The organization: %s has disabled users to signin":                                                     "signin-disabled",
	"auth:The application: %s has disabled users to signin":                                                      "signin-disabled",
	"auth:The login method: login with password is not enabled for the application":                              "signin-method-disabled",
	"auth:The login method: login with SMS is not enabled for the application":                                   "signin-method-disabled",
	"auth:The login method: login with email is not enabled for the application":                                 "signin-method-disabled",
	"auth:The login method: login with LDAP is not enabled for the application":                                  "signin-method-disabled",
	"auth:The login method: login with WebAuthn is not enabled for the application":                              "signin-method-disabled",
	"auth:The login method: login with face is not enabled for the application":                                  "signin-method-disabled",
	"auth:The login method: login with magic link is not enabled for the application":                            "signin-method-disabled",

	"general:Please login first":                               "login-required",
	"general:Please sign in first":                             "login-required",
	"auth:Please sign in again":                                "login-required",
	"check:Session outdated, please login again":               "login-required",
	"auth:Invalid token":                                       "invalid-token",
	"auth:Unauthorized operation":                              "permission-denied",
	"general:this operation requires administrator to perform": "permission-denied",
	"general:Only admin user can specify user":                 "permission-denied",
	"general:this operation is not allowed in demo mode":       "demo-mode-forbidden",

	"auth:The application: %s does not exist":  "application-not-found",
	"check:Application does not exist":         "application-not-found",
	"auth:The organization: %s does not exist": "organization-not-found",
	"auth:the organization: %s is not found":   "organization-not-found",
	"check:Organization does not exist":        "organization-not-found",
	"auth:The provider: %s does not exist":     "provider-not-found",
	"general:The object does not exist":        "object-not-found",

	"verification:Wrong verification code!":                                                          "verification-code-wrong",
	"verification:You should verify your code in %d min!":                                            "verification-code-expired",
	"verification:The verification code has already been used!":                                      "verification-code-used",
	"verification:The verification code has not been sent yet!":                                      "verification-code-not-sent",
	"verification:you can only send one code in %ds":                                                 "verification-code-too-frequent",
	"verification:Too many SMS codes have been sent to this number, please try again later":          "verification-code-limit",
	"verification:The SMS codes of today have run out, please try again tomorrow or use another way": "verification-code-limit",
	"verification:Turing test failed.":                                                               "captcha-failed",
	"verification:please add an Email provider to the \"Providers\" list for the application: %s":    "email-provider-missing",
	"verification:please add a SMS provider to the \"Providers\" list for the application: %s":       "sms-provider-missing",
	"verification:The magic link is invalid":                                                         "magic-link-invalid",
	"verification:The magic link has expired":                                                        "magic-link-expired",
	"verification:The magic link has already been used":                                              "magic-link-used",

	"general:Insufficient balance: new balance %v would be below credit limit %v":              "insufficient-balance",
	"general:Insufficient balance: new organization balance %v would be below credit limit %v": "insufficient-balance",
}

type errorCodePattern struct {
	regex *regexp.Regexp
	code  string
}

var (
	errorCodeTexts    = map[string]string{}
	errorCodePatterns = map[string][]errorCodePattern{}
	errorCodeOnce     sync.Once
	formatVerbRegex   = regexp.MustCompile(`%[-+# 0-9.]*[sdvqwf]`)
)

func loadErrorCodes() {
	for language, data := range langMap {
		for key, code := range errorCodes {
			tokens := strings.SplitN(key, ":", 2)
			text := data[tokens[0]][tokens[1]]
			if text == "" {
				text = tokens[1]
			}
			addErrorCodeText(language, text, code)
			addErrorCodeText(language, tokens[1], code)
		}
	}
}

func addErrorCodeText(language string, text string, code string) {
	text = strings.TrimSpace(text)
	// a message ending with a colon gets a name appended by its caller
	openEnded := strings.HasSuffix(text, ":") || strings.HasSuffix(text, "：")
	if !openEnded && !formatVerbRegex.MatchString(text) {
		errorCodeTexts[text] = code
		return
	}

	parts := formatVerbRegex.Split(text, -1)
	for i, part := range parts {
		parts[i] = regexp.QuoteMeta(part)
	}
	pattern := strings.Join(parts, "(?s:.*)")
	if openEnded {
		pattern += "(?s:.*)"
	}
	regex := regexp.MustCompile("^" + pattern + "$")
	errorCodePatterns[language] = append(errorCodePatterns[language], errorCodePattern{regex: regex, code: code})
}

// GetErrorCode returns the stable code of a translated error message, or "" when it is not a known business error
func GetErrorCode(language string, msg string) string {
	errorCodeOnce.Do(loadErrorCodes)

	msg = strings.TrimSpace(msg)
	if code, ok := errorCodeTexts[msg]; ok {
		return code
	}

	if _, ok := errorCodePatterns[language]; !ok {
		language = "en"
	}
	for _, lang := range []string{language, "en"} {
		for _, pattern := range errorCodePatterns[lang] {
			if pattern.regex.MatchString(msg) {
				return pattern.code
			}
		}
		if lang == "en" {
			break
		}
	}
	return ""
}
