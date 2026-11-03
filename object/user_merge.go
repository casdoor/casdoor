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
	"reflect"
	"strings"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm/names"
)

// userProviderFields are the columns from GitHub to Custom10 that hold the ids of linked third-party accounts
var userProviderFields = getUserProviderFields()

// userMergeProfileFields are copied to the kept account only where it has nothing yet
var userMergeProfileFields = []string{"DisplayName", "FirstName", "LastName", "Affiliation", "Title", "Region", "Location", "Gender", "Birthday", "Bio", "Homepage", "Language", "Education"}

func getUserProviderFields() []string {
	t := reflect.TypeOf(User{})
	fields := []string{}
	inside := false
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if name == "GitHub" {
			inside = true
		}
		if inside && t.Field(i).Type.Kind() == reflect.String {
			fields = append(fields, name)
		}
		if name == "Custom10" {
			break
		}
	}
	return append(fields, "Ldap")
}

func getUserStringField(user *User, field string) string {
	return reflect.ValueOf(user).Elem().FieldByName(field).String()
}

func setUserStringField(user *User, field string, value string) {
	reflect.ValueOf(user).Elem().FieldByName(field).SetString(value)
}

// GetUserMergeError tells why the account "from" can't be merged into the account "into",
// an empty string means that "into" can take over everything "from" signs in with
func GetUserMergeError(from *User, into *User, lang string) string {
	if from == nil || into == nil || from.Owner != into.Owner || from.Name == into.Name {
		return i18n.Translate(lang, "auth:Unauthorized operation")
	}
	if from.IsAdminUser() || into.IsAdminUser() || from.IsDeleted || into.IsDeleted || from.IsForbidden || into.IsForbidden {
		return i18n.Translate(lang, "auth:Unauthorized operation")
	}
	if from.BalanceCredit != 0 || (from.Balance != 0 && getUserBalanceCurrency(from) != getUserBalanceCurrency(into)) {
		return i18n.Translate(lang, "auth:Unauthorized operation")
	}

	if from.Email != "" && into.Email != "" && !strings.EqualFold(from.Email, into.Email) {
		return i18n.Translate(lang, "check:Email already exists")
	}
	if from.Phone != "" && into.Phone != "" && (from.Phone != into.Phone || from.GetCountryCode("") != into.GetCountryCode("")) {
		return i18n.Translate(lang, "check:Phone already exists")
	}
	for _, field := range userProviderFields {
		fromValue := getUserStringField(from, field)
		intoValue := getUserStringField(into, field)
		if fromValue != "" && intoValue != "" && fromValue != intoValue {
			return fmt.Sprintf(i18n.Translate(lang, "check:The account has already linked another %s account"), field)
		}
	}

	fromLinks, err := GetThirdPartyLinksByUser(from.Owner, from.Name)
	if err != nil {
		return err.Error()
	}
	for _, link := range fromLinks {
		intoLink, err := GetThirdPartyLink(into.Owner, into.Name, link.ProviderName)
		if err != nil {
			return err.Error()
		}
		if intoLink != nil && intoLink.ProviderId != link.ProviderId {
			return fmt.Sprintf(i18n.Translate(lang, "check:The account has already linked another %s account"), link.ProviderName)
		}
	}

	return ""
}

func getUserBalanceCurrency(user *User) string {
	if user.BalanceCurrency != "" {
		return user.BalanceCurrency
	}
	organization, err := getOrganization("admin", user.Owner)
	if err == nil && organization != nil && organization.BalanceCurrency != "" {
		return organization.BalanceCurrency
	}
	return "USD"
}

// GetMergedUsers is the names of the accounts merged into the user, for the applications to move what they keep by
// user name
func GetMergedUsers(user *User) []string {
	res := []string{}
	for _, name := range strings.Split(getUserProperty(user, UserPropertiesMergedUsers), ",") {
		if name != "" {
			res = append(res, name)
		}
	}
	return res
}

// MergeUser moves the third-party accounts, email, phone and balance of "from" into "into", fills the
// profile fields "into" leaves empty, lists "from" in the mergedUsers property of "into", then deletes "from"
func MergeUser(from *User, into *User) error {
	columns := []string{}
	for _, field := range userProviderFields {
		value := getUserStringField(from, field)
		if value == "" || getUserStringField(into, field) != "" {
			continue
		}

		setUserStringField(into, field, value)
		columns = append(columns, strings.ToLower(field))

		prefix := strings.ToLower(fmt.Sprintf("oauth_%s_", field))
		for k, v := range from.Properties {
			if strings.HasPrefix(strings.ToLower(k), prefix) {
				setUserProperty(into, k, v)
			}
		}
	}
	merged := GetMergedUsers(into)
	for _, name := range append(GetMergedUsers(from), from.Name) {
		if name != into.Name && !util.InSlice(merged, name) {
			merged = append(merged, name)
		}
	}
	setUserProperty(into, UserPropertiesMergedUsers, strings.Join(merged, ","))
	columns = append(columns, "properties")

	if from.Balance != 0 {
		into.Balance = AddPrices(into.Balance, from.Balance)
		columns = append(columns, "balance")
	}

	if into.Email == "" && from.Email != "" {
		into.Email = from.Email
		into.EmailVerified = from.EmailVerified
		columns = append(columns, "email", "email_verified")
	}
	if into.Phone == "" && from.Phone != "" {
		into.Phone = from.Phone
		into.CountryCode = from.CountryCode
		columns = append(columns, "phone", "country_code")
	}
	if into.IsDefaultAvatar && !from.IsDefaultAvatar && from.Avatar != "" {
		into.Avatar = from.Avatar
		into.IsDefaultAvatar = false
		columns = append(columns, "avatar", "is_default_avatar")
	}
	for _, field := range userMergeProfileFields {
		value := getUserStringField(from, field)
		if value != "" && getUserStringField(into, field) == "" {
			setUserStringField(into, field, value)
			columns = append(columns, ormerColumnName(field))
		}
	}

	// the links have to leave "from" first, deleting it would delete them too
	fromLinks, err := GetThirdPartyLinksByUser(from.Owner, from.Name)
	if err != nil {
		return err
	}
	for _, link := range fromLinks {
		_, err = ormer.Engine.Where("owner = ? AND user_name = ? AND provider_name = ?", link.Owner, link.UserName, link.ProviderName).Cols("user_name").Update(&ThirdPartyLink{UserName: into.Name})
		if err != nil {
			return err
		}
	}

	_, err = DeleteUser(from)
	if err != nil {
		return err
	}

	_, err = UpdateUser(into.GetId(), into, columns, true)
	return err
}

func ormerColumnName(field string) string {
	return names.SnakeMapper{}.Obj2Table(field)
}

// OAuthCodeRequest is what an authorization code was requested with
type OAuthCodeRequest struct {
	Application  *Application
	Provider     string
	SigninMethod string
	Scope        string
	Nonce        string
	Challenge    string
	Resource     string
}

// TakeOAuthCodeRequest invalidates an unused authorization code of the user and returns what it
// was requested with, so that the same request can be answered for the account it is merged into
func TakeOAuthCodeRequest(code string, user *User) (*OAuthCodeRequest, error) {
	token, err := getTokenByCode(code)
	if err != nil {
		return nil, err
	}
	if token == nil || token.CodeIsUsed || token.Organization != user.Owner || token.User != user.Name {
		return nil, nil
	}

	application, err := getApplication(token.Owner, token.Application)
	if err != nil || application == nil {
		return nil, err
	}

	request := &OAuthCodeRequest{
		Application: application,
		Scope:       token.Scope,
		Challenge:   token.CodeChallenge,
		Resource:    token.Resource,
	}
	if claims, err := ParseJwtTokenByApplication(token.AccessToken, application); err == nil {
		request.Nonce, request.Provider, request.SigninMethod = claims.Nonce, claims.Provider, claims.SigninMethod
	}

	token.CodeIsUsed = true
	_, err = ormer.Engine.ID(core.PK{token.Owner, token.Name}).Cols("code_is_used").Update(token)
	if err != nil {
		return nil, err
	}
	return request, nil
}
