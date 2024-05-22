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

package idp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/nyaruka/phonenumbers"
	"golang.org/x/oauth2"
)

const GoogleIdTokenKey = "GoogleIdToken"

type GoogleIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

// https://developers.google.com/identity/sign-in/web/backend-auth#calling-the-tokeninfo-endpoint
type GoogleIdToken struct {
	// These six fields are included in all Google ID Tokens.
	Iss string `json:"iss"` // The issuer, or signer, of the token. For Google-signed ID tokens, this value is https://accounts.google.com.
	Sub string `json:"sub"` // The subject: the ID that represents the principal making the request.
	Azp string `json:"azp"` // Optional. Who the token was issued to. Here is the ClientID
	Aud string `json:"aud"` // The audience of the token. Here is the ClientID
	Iat string `json:"iat"` // 	Unix epoch time when the token was issued.
	Exp string `json:"exp"` // 	Unix epoch time when the token expires.
	// These seven fields are only included when the user has granted the "profile" and "email" OAuth scopes to the application.
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Locale        string `json:"locale"`
}

func NewGoogleIdProvider(clientId string, clientSecret string, redirectUrl string) *GoogleIdProvider {
	idp := &GoogleIdProvider{}

	config := idp.getConfig()
	config.ClientID = clientId
	config.ClientSecret = clientSecret
	config.RedirectURL = redirectUrl
	idp.Config = config

	return idp
}

func (idp *GoogleIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *GoogleIdProvider) getConfig() *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://accounts.google.com/o/oauth2/token",
	}

	config := &oauth2.Config{
		Scopes:   []string{"profile", "email"},
		Endpoint: endpoint,
	}

	return config
}

func (idp *GoogleIdProvider) GetToken(code string) (*oauth2.Token, error) {
	// Obtained the GoogleIdToken through Google OneTap authorization.
	if strings.HasPrefix(code, GoogleIdTokenKey) {
		code = strings.TrimPrefix(code, GoogleIdTokenKey+"-")
		var googleIdToken GoogleIdToken
		if err := json.Unmarshal([]byte(code), &googleIdToken); err != nil {
			return nil, err
		}
		expiry := int64(util.ParseInt(googleIdToken.Exp))
		token := &oauth2.Token{
			AccessToken: fmt.Sprintf("%v-%v", GoogleIdTokenKey, googleIdToken.Sub),
			TokenType:   "Bearer",
			Expiry:      time.Unix(expiry, 0),
		}
		token = token.WithExtra(map[string]interface{}{
			GoogleIdTokenKey: googleIdToken,
		})
		return token, nil
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

//{
//	"id": "110613473084924141234",
//	"email": "jimgreen@gmail.com",
//	"verified_email": true,
//	"name": "Jim Green",
//	"given_name": "Jim",
//	"family_name": "Green",
//	"picture": "https://lh3.googleusercontent.com/-XdUIqdMkCWA/AAAAAAAAAAI/AAAAAAAAAAA/4252rscbv5M/photo.jpg",
//	"locale": "en"
//}

type GoogleUserInfo struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type GooglePeopleApiPhoneNumberMetaData struct {
	Primary bool `json:"primary"`
}

type GooglePeopleApiPhoneNumber struct {
	CanonicalForm string                             `json:"canonicalForm"`
	MetaData      GooglePeopleApiPhoneNumberMetaData `json:"metadata"`
	Value         string                             `json:"value"`
	Type          string                             `json:"type"`
}

type GooglePeopleApiResult struct {
	PhoneNumbers []GooglePeopleApiPhoneNumber `json:"phoneNumbers"`
	Etag         string                       `json:"etag"`
	ResourceName string                       `json:"resourceName"`
}

func (idp *GoogleIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	if strings.HasPrefix(token.AccessToken, GoogleIdTokenKey) {
		googleIdToken, ok := token.Extra(GoogleIdTokenKey).(GoogleIdToken)
		if !ok {
			return nil, errors.New("invalid googleIdToken")
		}
		userInfo := UserInfo{
			Id:          googleIdToken.Sub,
			Username:    googleIdToken.Email,
			DisplayName: googleIdToken.Name,
			Email:       googleIdToken.Email,
			AvatarUrl:   googleIdToken.Picture,
		}
		return &userInfo, nil
	}
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?alt=json&access_token=%s", token.AccessToken)
	resp, err := idp.Client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUserInfo GoogleUserInfo
	err = json.Unmarshal(body, &googleUserInfo)
	if err != nil {
		return nil, err
	}

	if googleUserInfo.Email == "" {
		return nil, errors.New("google email is empty")
	}

	url = fmt.Sprintf("https://people.googleapis.com/v1/people/me?personFields=phoneNumbers&access_token=%s", token.AccessToken)
	resp, err = idp.Client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googlePeopleResult GooglePeopleApiResult
	err = json.Unmarshal(body, &googlePeopleResult)
	if err != nil {
		return nil, err
	}

	var phoneNumber string
	var countryCode string
	if len(googlePeopleResult.PhoneNumbers) != 0 {
		for _, phoneData := range googlePeopleResult.PhoneNumbers {
			if phoneData.MetaData.Primary {
				phoneNumber = phoneData.CanonicalForm
				break
			}
		}
		phoneNumberParsed, err := phonenumbers.Parse(phoneNumber, "")
		if err != nil {
			return nil, err
		}
		countryCode = phonenumbers.GetRegionCodeForNumber(phoneNumberParsed)
		phoneNumber = fmt.Sprintf("%d", phoneNumberParsed.GetNationalNumber())
	}

	userInfo := UserInfo{
		Id:          googleUserInfo.Id,
		Username:    googleUserInfo.Email,
		DisplayName: googleUserInfo.Name,
		Email:       googleUserInfo.Email,
		AvatarUrl:   googleUserInfo.Picture,
		Phone:       phoneNumber,
		CountryCode: countryCode,
	}
	return &userInfo, nil
}
