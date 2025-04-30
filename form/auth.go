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

package form

import "reflect"

type AuthForm struct {
	Type         string `json:"type"`
	SigninMethod string `json:"signinMethod"`

	Organization   string `json:"organization"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Name           string `json:"name"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Gender         string `json:"gender"`
	Bio            string `json:"bio"`
	Tag            string `json:"tag"`
	Education      string `json:"education"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Affiliation    string `json:"affiliation"`
	IdCard         string `json:"idCard"`
	Language       string `json:"language"`
	Region         string `json:"region"`
	InvitationCode string `json:"invitationCode"`

	Application  string `json:"application"`
	ClientId     string `json:"clientId"`
	Provider     string `json:"provider"`
	ProviderBack string `json:"providerBack"`
	Code         string `json:"code"`
	State        string `json:"state"`
	RedirectUri  string `json:"redirectUri"`
	Method       string `json:"method"`

	EmailCode   string `json:"emailCode"`
	PhoneCode   string `json:"phoneCode"`
	CountryCode string `json:"countryCode"`

	AutoSignin bool `json:"autoSignin"`

	RelayState   string `json:"relayState"`
	SamlRequest  string `json:"samlRequest"`
	SamlResponse string `json:"samlResponse"`

	CaptchaType  string `json:"captchaType"`
	CaptchaToken string `json:"captchaToken"`
	ClientSecret string `json:"clientSecret"`

	MfaType      string `json:"mfaType"`
	Passcode     string `json:"passcode"`
	RecoveryCode string `json:"recoveryCode"`

	Plan    string `json:"plan"`
	Pricing string `json:"pricing"`

	FaceId      []float64 `json:"faceId"`
	FaceIdImage []string  `json:"faceIdImage"`
	UserCode    string    `json:"userCode"`
}

func GetAuthFormFieldValue(form *AuthForm, fieldName string) (bool, string) {
	val := reflect.ValueOf(*form)
	fieldValue := val.FieldByName(fieldName)

	if fieldValue.IsValid() && fieldValue.Kind() == reflect.String {
		return true, fieldValue.String()
	}
	return false, ""
}
