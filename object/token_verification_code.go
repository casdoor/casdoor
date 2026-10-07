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
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/util"
)

// VerificationCodeGrantType lets a native app sign a user in with the code sent to the
// phone or email, and sign the address up first when the application allows it
const VerificationCodeGrantType = "urn:casdoor:params:oauth:grant-type:verification-code"

// IsVerificationCodeSignupEnabled tells whether the verification code grant may create the
// account of an email or phone that has none, the code proves nothing but the address, so
// a signup item that asks for more than that keeps the grant to existing users
func (application *Application) IsVerificationCodeSignupEnabled(verifyType string) bool {
	if !util.InSlice(application.GrantTypes, VerificationCodeGrantType) {
		return false
	}

	if !application.EnableSignUp || !application.IsSignupAllowedFor(application.Organization) {
		return false
	}

	for _, signupItem := range application.SignupItems {
		if signupItem == nil || !signupItem.Required {
			continue
		}

		switch signupItem.Name {
		case "ID", "Username", "Display name", "Password", "Confirm password", "Agreement", "Signup button", "Providers":
			continue
		case "Email":
			if verifyType != VerifyTypeEmail {
				return false
			}
		case "Phone":
			if verifyType != VerifyTypePhone {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func getUserByEmailIgnoreCase(owner string, email string) (*User, error) {
	user, err := GetUserByEmail(owner, email)
	if err != nil || user != nil {
		return user, err
	}

	lowered := strings.ToLower(email)
	if lowered == email {
		return nil, nil
	}

	return GetUserByEmail(owner, lowered)
}

// GetVerificationCodeToken handles the verification code grant: "username" is the email or
// phone the code was sent to by /api/send-verification-code with the "login" method
func GetVerificationCodeToken(application *Application, username string, countryCode string, code string, scope string, host string, clientIp string, lang string) (*Token, *TokenError, error) {
	expandedScope, ok := IsScopeValidAndExpand(scope, application)
	if !ok {
		return nil, &TokenError{
			Error:            InvalidScope,
			ErrorDescription: "the requested scope is invalid or not defined in the application",
		}, nil
	}
	scope = expandedScope

	username = strings.TrimSpace(username)
	if username == "" || code == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "username and code are required",
		}, nil
	}

	var user *User
	var dest string
	var err error
	verifyType := GetVerifyType(username)
	if verifyType == VerifyTypeEmail {
		if !util.IsEmailValid(username) {
			return nil, &TokenError{
				Error:            InvalidRequest,
				ErrorDescription: "the email is invalid",
			}, nil
		}

		dest = username
		user, err = getUserByEmailIgnoreCase(application.Organization, username)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// a national number without a country code is the one of the user who has it, the
		// same as /api/send-verification-code resolves it
		if countryCode == "" && !strings.HasPrefix(username, "+") {
			user, err = GetUserByPhone(application.Organization, username)
			if err != nil {
				return nil, nil, err
			}
			if user != nil {
				countryCode = user.GetCountryCode("")
			} else if application.OrganizationObj != nil && len(application.OrganizationObj.CountryCodes) > 0 {
				countryCode = application.OrganizationObj.CountryCodes[0]
			}
		}

		var ok bool
		dest, ok = util.GetE164Number(username, countryCode)
		if !ok {
			return nil, &TokenError{
				Error:            InvalidRequest,
				ErrorDescription: "the phone number is invalid",
			}, nil
		}

		if user == nil {
			user, err = GetUserByPhone(application.Organization, dest)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	if user != nil {
		err = CheckSigninCode(user, dest, code, lang)
	} else {
		err = CheckVerifyCodeWithLimitAndIp(nil, clientIp, dest, code, lang)
	}
	if err != nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: err.Error(),
		}, nil
	}

	if user == nil {
		if !application.IsVerificationCodeSignupEnabled(verifyType) {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "the user does not exist, and the application does not allow a verification code to sign up new account: enable signup and require no signup item that the code can't fill in",
			}, nil
		}

		user, err = addVerificationCodeUser(application, verifyType, dest, clientIp, lang)
		if err != nil {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: err.Error(),
			}, nil
		}
	} else if verifyType == VerifyTypeEmail && !user.EmailVerified {
		user.EmailVerified = true
		_, err = UpdateUser(user.GetId(), user, []string{"email_verified"}, false)
		if err != nil {
			return nil, nil, err
		}
	}

	err = DisableVerificationCode(dest)
	if err != nil {
		return nil, nil, err
	}

	if user.IsMfaEnabled() {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user has MFA enabled and cannot sign in with a verification code grant, please use the authorization code flow",
		}, nil
	}

	if tokenError := getSigninPolicyTokenError(user, lang); tokenError != nil {
		return nil, tokenError, nil
	}

	if tokenError := checkGrantUserSignin(application, user, clientIp, lang); tokenError != nil {
		return nil, tokenError, nil
	}

	return getUserGrantToken(application, user, scope, host)
}

func addVerificationCodeUser(application *Application, verifyType string, dest string, clientIp string, lang string) (*User, error) {
	organization, err := GetOrganization(util.GetId("admin", application.Organization))
	if err != nil {
		return nil, err
	}
	if organization == nil {
		return nil, fmt.Errorf("the organization: %s does not exist", application.Organization)
	}

	err = CheckEntryIp(clientIp, nil, application, organization, lang)
	if err != nil {
		return nil, err
	}

	id, err := GenerateIdForNewUser(application)
	if err != nil {
		return nil, err
	}

	initScore, err := organization.GetInitScore()
	if err != nil {
		return nil, err
	}

	user := &User{
		Owner:             application.Organization,
		Name:              id,
		CreatedTime:       util.GetCurrentTime(),
		Id:                id,
		Type:              "normal-user",
		Avatar:            organization.DefaultAvatar,
		Address:           []string{},
		Score:             initScore,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
		RegisterType:      "Application Signup",
		RegisterSource:    fmt.Sprintf("%s/%s", application.Organization, application.Name),
	}

	if verifyType == VerifyTypeEmail {
		email := strings.ToLower(dest)
		user.Email = email
		user.EmailVerified = true
		user.DisplayName = email
		if organization.UseEmailAsUsername {
			existedUser, err := GetUser(util.GetId(application.Organization, email))
			if err != nil {
				return nil, err
			}
			if existedUser == nil {
				user.Name = email
			}
		}
	} else {
		user.Phone, user.CountryCode = util.ParseE164Phone(dest)
		if !util.IsPhoneAllowInRegin(user.CountryCode, organization.CountryCodes) {
			return nil, fmt.Errorf("the phone number: %s is not in the allowed regions of the organization", dest)
		}
	}

	if application.DefaultGroup != "" {
		user.Groups = []string{application.DefaultGroup}
	}
	if application.DefaultTag != "" {
		user.Tag = application.DefaultTag
	}

	affected, err := AddUser(user, lang)
	if err != nil {
		return nil, err
	}
	if !affected {
		return nil, fmt.Errorf("failed to add user: %s", user.GetId())
	}

	err = AddUserToOriginalDatabase(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
