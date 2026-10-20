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

	"github.com/casdoor/casdoor/util"
)

// MobileOneClickGrantType lets a native app sign a user in with the token that the carrier
// one-click login SDK of Alibaba Cloud PNVS got on the phone, and sign the number up first
// when the application allows it
const MobileOneClickGrantType = "urn:casdoor:params:oauth:grant-type:mobile-one-click"

const pnvsProviderType = "Alibaba Cloud PNVS SMS"

func getMobileOneClickProvider(application *Application) (*Provider, error) {
	providerItem := application.GetProviderItemByType(pnvsProviderType)
	if providerItem == nil {
		return nil, nil
	}

	return GetProvider(util.GetId(providerItem.Provider.Owner, providerItem.Provider.Name))
}

func getMobileByOneClickToken(provider *Provider, token string) (string, error) {
	client, err := newPnvsSmsClient(provider.ClientId, provider.ClientSecret, provider.SignName, provider.TemplateCode, provider.RegionId)
	if err != nil {
		return "", err
	}

	return client.GetMobile(token)
}

// getMobileOneClickUser also takes a user whose phone was stored without a country code
func getMobileOneClickUser(owner string, dest string) (*User, error) {
	user, err := GetUserByPhone(owner, dest)
	if user != nil || err != nil {
		return user, err
	}

	phone, _ := util.ParseE164Phone(dest)
	user = &User{}
	existed, err := ormer.Engine.Where("owner = ? and phone = ? and country_code = ?", owner, phone, "").Get(user)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}

	return user, nil
}

// GetMobileOneClickToken handles the mobile one-click grant: "code" is the token that the
// one-click login SDK returned on the phone
func GetMobileOneClickToken(application *Application, code string, scope string, host string, clientIp string, lang string) (*Token, *TokenError, error) {
	expandedScope, ok := IsScopeValidAndExpand(scope, application)
	if !ok {
		return nil, &TokenError{
			Error:            InvalidScope,
			ErrorDescription: "the requested scope is invalid or not defined in the application",
		}, nil
	}
	scope = expandedScope

	if code == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "code is required",
		}, nil
	}

	provider, err := getMobileOneClickProvider(application)
	if err != nil {
		return nil, nil, err
	}
	if provider == nil {
		return nil, &TokenError{
			Error:            UnauthorizedClient,
			ErrorDescription: fmt.Sprintf("the application has no %s provider for mobile one-click login", pnvsProviderType),
		}, nil
	}

	mobile, err := getMobileByOneClickToken(provider, code)
	if err != nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: err.Error(),
		}, nil
	}

	dest, ok := util.GetE164Number(mobile, "CN")
	if !ok {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the phone number: %s is invalid", mobile),
		}, nil
	}

	user, err := getMobileOneClickUser(application.Organization, dest)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		if !application.isAddressGrantSignupEnabled(MobileOneClickGrantType, VerifyTypePhone) {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "the user does not exist, and the application does not allow mobile one-click login to sign up new account: enable signup and require no signup item that the phone number can't fill in",
			}, nil
		}

		user, err = addVerificationCodeUser(application, VerifyTypePhone, dest, clientIp, lang)
		if err != nil {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: err.Error(),
			}, nil
		}
	}

	if user.IsMfaEnabled() {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user has MFA enabled and cannot sign in with a mobile one-click grant, please use the authorization code flow",
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
