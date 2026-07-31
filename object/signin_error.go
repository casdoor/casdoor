// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

import "github.com/casdoor/casdoor/i18n"

const (
	SigninReasonUserNotFound    = "user-not-found"
	SigninReasonAccountDisabled = "account-disabled"
	SigninReasonAccountFrozen   = "account-frozen"
	SigninReasonWrongPassword   = "wrong-password"
	SigninReasonPasswordExpired = "password-expired"
	SigninReasonMfaFailed       = "mfa-failed"
)

// SigninError is an error that carries a structured failure reason for audit logging.
type SigninError struct {
	Reason string
	msg    string
}

func (e *SigninError) Error() string { return e.msg }

func newSigninError(reason, msg string) *SigninError {
	return &SigninError{Reason: reason, msg: msg}
}

// InvalidSigninCredentialsMsg is the generic client-facing sign-in failure text (OWASP anti-enumeration).
func InvalidSigninCredentialsMsg(lang string) string {
	return i18n.Translate(lang, "check:password or code is incorrect")
}

// NewSigninUserNotFoundError records user-not-found for audit logs while returning a generic message to the client.
func NewSigninUserNotFoundError(lang string) *SigninError {
	return newSigninError(SigninReasonUserNotFound, InvalidSigninCredentialsMsg(lang))
}
