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
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
	"github.com/go-webauthn/webauthn/webauthn"
	jsoniter "github.com/json-iterator/go"
	"github.com/xorm-io/core"
	"golang.org/x/oauth2"
)

func GetUserByField(organizationName string, field string, value string) (*User, error) {
	if field == "" || value == "" {
		return nil, nil
	}

	user := User{Owner: organizationName}
	existed, err := ormer.Engine.Where(fmt.Sprintf("%s=?", strings.ToLower(field)), value).Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	} else {
		return nil, nil
	}
}

func HasUserByField(organizationName string, field string, value string) bool {
	user, err := GetUserByField(organizationName, field, value)
	if err != nil {
		panic(err)
	}
	return user != nil
}

func GetUserByFields(organization string, field string) (*User, error) {
	isUsernameLowered := conf.GetConfigBool("isUsernameLowered")
	if isUsernameLowered {
		field = strings.ToLower(field)
	}

	field = strings.TrimSpace(field)

	// check username
	user, err := GetUserByField(organization, "name", field)
	if err != nil || user != nil {
		return user, err
	}

	// check email
	if strings.Contains(field, "@") {
		normalizedEmail := strings.ToLower(field)
		user, err = GetUserByField(organization, "email", normalizedEmail)
		if user != nil || err != nil {
			return user, err
		}
	}

	// check phone
	phone := util.GetSeperatedPhone(field)
	user, err = GetUserByField(organization, "phone", phone)
	if user != nil || err != nil {
		return user, err
	}

	// check user ID
	user, err = GetUserByField(organization, "id", field)
	if user != nil || err != nil {
		return user, err
	}

	// check ID card
	user, err = GetUserByField(organization, "id_card", field)
	if user != nil || err != nil {
		return user, err
	}

	return nil, nil
}

func SetUserField(user *User, field string, value string) (bool, error) {
	bean := make(map[string]interface{})
	if field == "password" {
		organization, err := GetOrganizationByUser(user)
		if err != nil {
			return false, err
		}

		user.UpdateUserPassword(organization)
		bean[strings.ToLower(field)] = user.Password
		bean["password_type"] = user.PasswordType
	} else {
		bean[strings.ToLower(field)] = value
	}

	affected, err := ormer.Engine.Table(user).ID(core.PK{user.Owner, user.Name}).Update(bean)
	if err != nil {
		return false, err
	}

	user, err = getUser(user.Owner, user.Name)
	if err != nil {
		return false, err
	}

	err = user.UpdateUserHash()
	if err != nil {
		return false, err
	}

	if user != nil {
		user.UpdatedTime = util.GetCurrentTime()
	}

	_, err = ormer.Engine.ID(core.PK{user.Owner, user.Name}).Cols("hash").Update(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func GetUserField(user *User, field string) string {
	// https://socketloop.com/tutorials/golang-how-to-get-struct-field-and-value-by-name
	u := reflect.ValueOf(user)
	f := reflect.Indirect(u).FieldByName(field)
	return f.String()
}

func setUserProperty(user *User, field string, value string) {
	if value == "" {
		delete(user.Properties, field)
	} else {
		if user.Properties == nil {
			user.Properties = make(map[string]string)
		}

		user.Properties[field] = value
	}
}

func getUserProperty(user *User, field string) string {
	if user.Properties == nil {
		return ""
	}
	return user.Properties[field]
}

func getUserExtraProperty(user *User, providerType, key string) (string, error) {
	extraJson := getUserProperty(user, fmt.Sprintf("oauth_%s_extra", providerType))
	if extraJson == "" {
		return "", nil
	}
	extra := make(map[string]string)
	if err := jsoniter.Unmarshal([]byte(extraJson), &extra); err != nil {
		return "", err
	}
	return extra[key], nil
}

// GetUserOAuthAccessToken retrieves the OAuth access token for a specific provider
func GetUserOAuthAccessToken(user *User, providerType string) string {
	accessTokenKey := fmt.Sprintf("oauth_%s_accessToken", providerType)
	return getUserProperty(user, accessTokenKey)
}

// GetUserOAuthRefreshToken retrieves the OAuth refresh token for a specific provider
func GetUserOAuthRefreshToken(user *User, providerType string) string {
	refreshTokenKey := fmt.Sprintf("oauth_%s_refreshToken", providerType)
	return getUserProperty(user, refreshTokenKey)
}

func SetUserOAuthProperties(organization *Organization, user *User, providerType string, userInfo *idp.UserInfo, token *oauth2.Token, userMapping ...map[string]string) (bool, error) {
	// Store the original OAuth provider token if available
	if token != nil && token.AccessToken != "" {
		// Store tokens per provider in Properties map
		accessTokenKey := fmt.Sprintf("oauth_%s_accessToken", providerType)
		setUserProperty(user, accessTokenKey, token.AccessToken)
		
		if token.RefreshToken != "" {
			refreshTokenKey := fmt.Sprintf("oauth_%s_refreshToken", providerType)
			setUserProperty(user, refreshTokenKey, token.RefreshToken)
		}
		
		// Also update the legacy fields for backward compatibility
		user.OriginalToken = token.AccessToken
		user.OriginalRefreshToken = token.RefreshToken
	}

	if userInfo.Id != "" {
		propertyName := fmt.Sprintf("oauth_%s_id", providerType)
		setUserProperty(user, propertyName, userInfo.Id)
	}
	if userInfo.Username != "" {
		propertyName := fmt.Sprintf("oauth_%s_username", providerType)
		setUserProperty(user, propertyName, userInfo.Username)
	}
	if userInfo.DisplayName != "" {
		propertyName := fmt.Sprintf("oauth_%s_displayName", providerType)
		setUserProperty(user, propertyName, userInfo.DisplayName)
		if user.DisplayName == "" {
			user.DisplayName = userInfo.DisplayName
		}
	} else if user.DisplayName == "" {
		if userInfo.Username != "" {
			user.DisplayName = userInfo.Username
		} else {
			user.DisplayName = userInfo.Id
		}
	}
	if userInfo.Email != "" {
		propertyName := fmt.Sprintf("oauth_%s_email", providerType)
		setUserProperty(user, propertyName, userInfo.Email)
		if user.Email == "" {
			user.Email = userInfo.Email
		}
	}

	if userInfo.UnionId != "" {
		propertyName := fmt.Sprintf("oauth_%s_unionId", providerType)
		setUserProperty(user, propertyName, userInfo.UnionId)
	}

	if userInfo.AvatarUrl != "" {
		propertyName := fmt.Sprintf("oauth_%s_avatarUrl", providerType)
		setUserProperty(user, propertyName, userInfo.AvatarUrl)
		if user.Avatar == "" || user.Avatar == organization.DefaultAvatar {
			user.Avatar = userInfo.AvatarUrl
		}
	}

	// Apply custom user mapping from provider configuration
	if len(userMapping) > 0 && userMapping[0] != nil && len(userMapping[0]) > 0 && userInfo.Extra != nil {
		applyUserMapping(user, userInfo.Extra, userMapping[0])
	}

	if userInfo.Extra != nil {
		// Save extra info as json string
		propertyName := fmt.Sprintf("oauth_%s_extra", providerType)
		oldExtraJson := getUserProperty(user, propertyName)
		extra := make(map[string]string)
		if oldExtraJson != "" {
			if err := jsoniter.Unmarshal([]byte(oldExtraJson), &extra); err != nil {
				return false, err
			}
		}
		for k, v := range userInfo.Extra {
			extra[k] = v
		}

		newExtraJson, err := jsoniter.Marshal(extra)
		if err != nil {
			return false, err
		}
		setUserProperty(user, propertyName, string(newExtraJson))
	}

	return UpdateUserForAllFields(user.GetId(), user)
}

func applyUserMapping(user *User, extraClaims map[string]string, userMapping map[string]string) {
	// Map of user fields that can be set from IDP claims
	for userField, claimName := range userMapping {
		// Skip standard fields that are already handled
		if userField == "id" || userField == "username" || userField == "displayName" || userField == "email" || userField == "avatarUrl" {
			continue
		}

		// Get value from extra claims
		claimValue, exists := extraClaims[claimName]
		if !exists || claimValue == "" {
			continue
		}

		// Map to user fields based on field name
		switch strings.ToLower(userField) {
		case "phone":
			if user.Phone == "" {
				user.Phone = claimValue
			}
		case "countrycode":
			if user.CountryCode == "" {
				user.CountryCode = claimValue
			}
		case "firstname":
			if user.FirstName == "" {
				user.FirstName = claimValue
			}
		case "lastname":
			if user.LastName == "" {
				user.LastName = claimValue
			}
		case "region":
			if user.Region == "" {
				user.Region = claimValue
			}
		case "location":
			if user.Location == "" {
				user.Location = claimValue
			}
		case "affiliation":
			if user.Affiliation == "" {
				user.Affiliation = claimValue
			}
		case "title":
			if user.Title == "" {
				user.Title = claimValue
			}
		case "homepage":
			if user.Homepage == "" {
				user.Homepage = claimValue
			}
		case "bio":
			if user.Bio == "" {
				user.Bio = claimValue
			}
		case "tag":
			if user.Tag == "" {
				user.Tag = claimValue
			}
		case "language":
			if user.Language == "" {
				user.Language = claimValue
			}
		case "gender":
			if user.Gender == "" {
				user.Gender = claimValue
			}
		case "birthday":
			if user.Birthday == "" {
				user.Birthday = claimValue
			}
		case "education":
			if user.Education == "" {
				user.Education = claimValue
			}
		case "idcard":
			if user.IdCard == "" {
				user.IdCard = claimValue
			}
		case "idcardtype":
			if user.IdCardType == "" {
				user.IdCardType = claimValue
			}
		}
	}
}

func getUserRoleNames(user *User) (res []string) {
	for _, role := range user.Roles {
		res = append(res, role.Name)
	}
	return res
}

func getUserPermissionNames(user *User) (res []string) {
	for _, permission := range user.Permissions {
		res = append(res, permission.Name)
	}
	return res
}

func ClearUserOAuthProperties(user *User, providerType string) (bool, error) {
	for k := range user.Properties {
		prefix := fmt.Sprintf("oauth_%s_", providerType)
		if strings.HasPrefix(k, prefix) {
			delete(user.Properties, k)
		}
	}

	affected, err := ormer.Engine.ID(core.PK{user.Owner, user.Name}).Cols("properties").Update(user)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func userVisible(isAdmin bool, item *AccountItem) bool {
	if item == nil {
		return false
	}

	if item.ViewRule == "Admin" && !isAdmin {
		return false
	}

	return true
}

func CheckPermissionForUpdateUser(oldUser, newUser *User, isAdmin bool, allowDisplayNameEmpty bool, lang string) (bool, string) {
	organization, err := GetOrganizationByUser(oldUser)
	if err != nil {
		return false, err.Error()
	}

	var itemsChanged []*AccountItem

	if oldUser.Owner != newUser.Owner {
		item := GetAccountItemByName("Organization", organization)
		if !userVisible(isAdmin, item) {
			newUser.Owner = oldUser.Owner
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Name != newUser.Name {
		item := GetAccountItemByName("Name", organization)
		if !userVisible(isAdmin, item) {
			newUser.Name = oldUser.Name
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Id != newUser.Id {
		item := GetAccountItemByName("ID", organization)
		if !userVisible(isAdmin, item) {
			newUser.Id = oldUser.Id
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.DisplayName != newUser.DisplayName {
		item := GetAccountItemByName("Display name", organization)
		if !userVisible(isAdmin, item) {
			newUser.DisplayName = oldUser.DisplayName
		} else {
			if !allowDisplayNameEmpty && newUser.DisplayName == "" {
				return false, i18n.Translate(lang, "user:Display name cannot be empty")
			}

			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Avatar != newUser.Avatar {
		item := GetAccountItemByName("Avatar", organization)
		if !userVisible(isAdmin, item) {
			newUser.Avatar = oldUser.Avatar
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Type != newUser.Type {
		item := GetAccountItemByName("User type", organization)
		if !userVisible(isAdmin, item) {
			newUser.Type = oldUser.Type
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	// The password is *** when not modified
	if oldUser.Password != newUser.Password && newUser.Password != "***" {
		item := GetAccountItemByName("Password", organization)
		if !userVisible(isAdmin, item) {
			newUser.Password = oldUser.Password
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Email != newUser.Email {
		item := GetAccountItemByName("Email", organization)
		if !userVisible(isAdmin, item) {
			newUser.Email = oldUser.Email
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Phone != newUser.Phone {
		item := GetAccountItemByName("Phone", organization)
		if !userVisible(isAdmin, item) {
			newUser.Phone = oldUser.Phone
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.CountryCode != newUser.CountryCode {
		item := GetAccountItemByName("Country code", organization)
		if !userVisible(isAdmin, item) {
			newUser.CountryCode = oldUser.CountryCode
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Region != newUser.Region {
		item := GetAccountItemByName("Country/Region", organization)
		if !userVisible(isAdmin, item) {
			newUser.Region = oldUser.Region
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Location != newUser.Location {
		item := GetAccountItemByName("Location", organization)
		if !userVisible(isAdmin, item) {
			newUser.Location = oldUser.Location
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Affiliation != newUser.Affiliation {
		item := GetAccountItemByName("Affiliation", organization)
		if !userVisible(isAdmin, item) {
			newUser.Affiliation = oldUser.Affiliation
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Title != newUser.Title {
		item := GetAccountItemByName("Title", organization)
		if !userVisible(isAdmin, item) {
			newUser.Title = oldUser.Title
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Homepage != newUser.Homepage {
		item := GetAccountItemByName("Homepage", organization)
		if !userVisible(isAdmin, item) {
			newUser.Homepage = oldUser.Homepage
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Bio != newUser.Bio {
		item := GetAccountItemByName("Bio", organization)
		if !userVisible(isAdmin, item) {
			newUser.Bio = oldUser.Bio
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.Tag != newUser.Tag {
		item := GetAccountItemByName("Tag", organization)
		if !userVisible(isAdmin, item) {
			newUser.Tag = oldUser.Tag
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.SignupApplication != newUser.SignupApplication {
		item := GetAccountItemByName("Signup application", organization)
		if !userVisible(isAdmin, item) {
			newUser.SignupApplication = oldUser.SignupApplication
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Gender != newUser.Gender {
		item := GetAccountItemByName("Gender", organization)
		if !userVisible(isAdmin, item) {
			newUser.Gender = oldUser.Gender
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Birthday != newUser.Birthday {
		item := GetAccountItemByName("Birthday", organization)
		if !userVisible(isAdmin, item) {
			newUser.Birthday = oldUser.Birthday
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Education != newUser.Education {
		item := GetAccountItemByName("Education", organization)
		if !userVisible(isAdmin, item) {
			newUser.Education = oldUser.Education
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.IdCard != newUser.IdCard {
		item := GetAccountItemByName("ID card", organization)
		if !userVisible(isAdmin, item) {
			newUser.IdCard = oldUser.IdCard
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.IdCardType != newUser.IdCardType {
		item := GetAccountItemByName("ID card type", organization)
		if !userVisible(isAdmin, item) {
			newUser.IdCardType = oldUser.IdCardType
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	oldUserPropertiesJson, _ := json.Marshal(oldUser.Properties)
	if newUser.Properties == nil {
		newUser.Properties = make(map[string]string)
	}
	newUserPropertiesJson, _ := json.Marshal(newUser.Properties)
	if string(oldUserPropertiesJson) != string(newUserPropertiesJson) {
		item := GetAccountItemByName("Properties", organization)
		if !userVisible(isAdmin, item) {
			newUser.Properties = oldUser.Properties
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.PreferredMfaType != newUser.PreferredMfaType {
		item := GetAccountItemByName("Multi-factor authentication", organization)
		if !userVisible(isAdmin, item) {
			newUser.PreferredMfaType = oldUser.PreferredMfaType
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Groups == nil {
		oldUser.Groups = []string{}
	}
	oldUserGroupsJson, _ := json.Marshal(oldUser.Groups)

	if newUser.Groups == nil {
		newUser.Groups = []string{}
	}
	newUserGroupsJson, _ := json.Marshal(newUser.Groups)
	if string(oldUserGroupsJson) != string(newUserGroupsJson) {
		item := GetAccountItemByName("Groups", organization)
		if !userVisible(isAdmin, item) {
			newUser.Groups = oldUser.Groups
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Address == nil {
		oldUser.Address = []string{}
	}
	oldUserAddressJson, _ := json.Marshal(oldUser.Address)

	if newUser.Address == nil {
		newUser.Address = []string{}
	}
	newUserAddressJson, _ := json.Marshal(newUser.Address)
	if string(oldUserAddressJson) != string(newUserAddressJson) {
		item := GetAccountItemByName("Address", organization)
		if !userVisible(isAdmin, item) {
			newUser.Address = oldUser.Address
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if newUser.FaceIds != nil {
		item := GetAccountItemByName("Face ID", organization)
		if !userVisible(isAdmin, item) {
			newUser.FaceIds = oldUser.FaceIds
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.IsAdmin != newUser.IsAdmin {
		item := GetAccountItemByName("Is admin", organization)
		if !userVisible(isAdmin, item) {
			newUser.IsAdmin = oldUser.IsAdmin
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.IsForbidden != newUser.IsForbidden {
		item := GetAccountItemByName("Is forbidden", organization)
		if !userVisible(isAdmin, item) {
			newUser.IsForbidden = oldUser.IsForbidden
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.IsDeleted != newUser.IsDeleted {
		item := GetAccountItemByName("Is deleted", organization)
		if !userVisible(isAdmin, item) {
			newUser.IsDeleted = oldUser.IsDeleted
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.NeedUpdatePassword != newUser.NeedUpdatePassword {
		item := GetAccountItemByName("Need update password", organization)
		if !userVisible(isAdmin, item) {
			newUser.NeedUpdatePassword = oldUser.NeedUpdatePassword
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}
	if oldUser.IpWhitelist != newUser.IpWhitelist {
		item := GetAccountItemByName("IP whitelist", organization)
		if !userVisible(isAdmin, item) {
			newUser.IpWhitelist = oldUser.IpWhitelist
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Balance != newUser.Balance {
		item := GetAccountItemByName("Balance", organization)
		if !userVisible(isAdmin, item) {
			newUser.Balance = oldUser.Balance
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Score != newUser.Score {
		item := GetAccountItemByName("Score", organization)
		if !userVisible(isAdmin, item) {
			newUser.Score = oldUser.Score
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Karma != newUser.Karma {
		item := GetAccountItemByName("Karma", organization)
		if !userVisible(isAdmin, item) {
			newUser.Karma = oldUser.Karma
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Language != newUser.Language {
		item := GetAccountItemByName("Language", organization)
		if !userVisible(isAdmin, item) {
			newUser.Language = oldUser.Language
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Ranking != newUser.Ranking {
		item := GetAccountItemByName("Ranking", organization)
		if !userVisible(isAdmin, item) {
			newUser.Ranking = oldUser.Ranking
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Currency != newUser.Currency {
		item := GetAccountItemByName("Currency", organization)
		if !userVisible(isAdmin, item) {
			newUser.Currency = oldUser.Currency
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	if oldUser.Hash != newUser.Hash {
		item := GetAccountItemByName("Hash", organization)
		if !userVisible(isAdmin, item) {
			newUser.Hash = oldUser.Hash
		} else {
			itemsChanged = append(itemsChanged, item)
		}
	}

	for _, accountItem := range itemsChanged {

		if pass, err := CheckAccountItemModifyRule(accountItem, isAdmin, lang); !pass {
			return pass, err
		}

		exist, userValue, err := GetUserFieldStringValue(newUser, util.SpaceToCamel(accountItem.Name))
		if err != nil {
			return false, err.Error()
		}

		if !exist {
			continue
		}

		if accountItem.Regex == "" {
			continue
		}
		regexSignupItem, err := regexp.Compile(accountItem.Regex)
		if err != nil {
			return false, err.Error()
		}

		matched := regexSignupItem.MatchString(userValue)
		if !matched {
			return false, fmt.Sprintf(i18n.Translate(lang, "check:The value \"%s\" for account field \"%s\" doesn't match the account item regex"), userValue, accountItem.Name)
		}
	}
	return true, ""
}

func (user *User) GetCountryCode(countryCode string) string {
	if countryCode != "" {
		return countryCode
	}

	if user != nil && user.CountryCode != "" {
		return user.CountryCode
	}

	if org, _ := GetOrganizationByUser(user); org != nil && len(org.CountryCodes) > 0 {
		return org.CountryCodes[0]
	}
	return ""
}

func (user *User) IsAdminUser() bool {
	if user == nil {
		return false
	}

	return user.IsAdmin || user.IsGlobalAdmin()
}

func IsAppUser(userId string) bool {
	if strings.HasPrefix(userId, "app/") {
		return true
	}
	return false
}

func setReflectAttr[T any](fieldValue *reflect.Value, fieldString string) error {
	unmarshalValue := new(T)
	err := json.Unmarshal([]byte(fieldString), unmarshalValue)
	if err != nil {
		return err
	}

	fvElem := fieldValue
	fvElem.Set(reflect.ValueOf(*unmarshalValue))
	return nil
}

func StringArrayToStruct[T any](stringArray [][]string) ([]*T, error) {
	fieldNames := stringArray[0]
	excelMap := []map[string]string{}
	structFieldMap := map[string]int{}

	reflectedStruct := reflect.TypeOf(*new(T))
	for i := 0; i < reflectedStruct.NumField(); i++ {
		structFieldMap[strings.ToLower(reflectedStruct.Field(i).Name)] = i
	}

	for idx, field := range stringArray {
		if idx == 0 {
			continue
		}

		tempMap := map[string]string{}
		for idx, val := range field {
			tempMap[fieldNames[idx]] = val
		}
		excelMap = append(excelMap, tempMap)
	}

	instances := []*T{}
	var err error

	for idx, m := range excelMap {
		instance := new(T)
		reflectedInstance := reflect.ValueOf(instance).Elem()

		for k, v := range m {
			if v == "" || v == "null" || v == "[]" || v == "{}" {
				continue
			}
			fName := strings.ToLower(strings.ReplaceAll(k, "_", ""))
			fieldIdx, ok := structFieldMap[fName]
			if !ok {
				continue
			}
			fv := reflectedInstance.Field(fieldIdx)
			if !fv.IsValid() {
				continue
			}
			switch fv.Kind() {
			case reflect.String:
				fv.SetString(v)
				continue
			case reflect.Bool:
				fv.SetBool(v == "1")
				continue
			case reflect.Int:
				intVal, err := strconv.Atoi(v)
				if err != nil {
					return nil, fmt.Errorf("line %d - column %s: %s", idx+1, fName, err.Error())
				}
				fv.SetInt(int64(intVal))
				continue
			}

			switch fv.Type() {
			case reflect.TypeOf([]string{}):
				err = setReflectAttr[[]string](&fv, v)
			case reflect.TypeOf([]*string{}):
				err = setReflectAttr[[]*string](&fv, v)
			case reflect.TypeOf([]*FaceId{}):
				err = setReflectAttr[[]*FaceId](&fv, v)
			case reflect.TypeOf([]*MfaProps{}):
				err = setReflectAttr[[]*MfaProps](&fv, v)
			case reflect.TypeOf([]*Role{}):
				err = setReflectAttr[[]*Role](&fv, v)
			case reflect.TypeOf([]*Permission{}):
				err = setReflectAttr[[]*Permission](&fv, v)
			case reflect.TypeOf([]ManagedAccount{}):
				err = setReflectAttr[[]ManagedAccount](&fv, v)
			case reflect.TypeOf([]MfaAccount{}):
				err = setReflectAttr[[]MfaAccount](&fv, v)
			case reflect.TypeOf([]webauthn.Credential{}):
				err = setReflectAttr[[]webauthn.Credential](&fv, v)
			case reflect.TypeOf(map[string]string{}):
				err = setReflectAttr[map[string]string](&fv, v)
			}

			if err != nil {
				return nil, fmt.Errorf("line %d: %s", idx, err.Error())
			}
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

func replaceAttributeValue(user *User, value string) []string {
	if user == nil {
		return nil
	}
	valueList := []string{value}
	if strings.Contains(value, "$user.roles") {
		valueList = replaceAttributeValuesWithList("$user.roles", getUserRoleNames(user), valueList)
	}

	if strings.Contains(value, "$user.permissions") {
		valueList = replaceAttributeValuesWithList("$user.permissions", getUserPermissionNames(user), valueList)
	}

	if strings.Contains(value, "$user.groups") {
		valueList = replaceAttributeValuesWithList("$user.groups", user.Groups, valueList)
	}

	valueList = replaceAttributeValues("$user.owner", user.Owner, valueList)
	valueList = replaceAttributeValues("$user.name", user.Name, valueList)
	valueList = replaceAttributeValues("$user.email", user.Email, valueList)
	valueList = replaceAttributeValues("$user.id", user.Id, valueList)
	valueList = replaceAttributeValues("$user.phone", user.Phone, valueList)

	return valueList
}

func replaceAttributeValues(val string, replaceVal string, values []string) []string {
	var newValues []string
	for _, value := range values {
		newValues = append(newValues, strings.ReplaceAll(value, val, replaceVal))
	}

	return newValues
}

func replaceAttributeValuesWithList(val string, replaceVals []string, values []string) []string {
	var newValues []string
	for _, value := range values {
		for _, rVal := range replaceVals {
			newValues = append(newValues, strings.ReplaceAll(value, val, rVal))
		}
	}

	return newValues
}

// TriggerWebhookForUser triggers a webhook for user operations (add, update, delete)
// action: the action type, e.g., "new-user", "update-user", "delete-user"
// user: the user object
func TriggerWebhookForUser(action string, user *User) {
	if user == nil {
		return
	}

	record := &casvisorsdk.Record{
		Name:         util.GenerateId(),
		CreatedTime:  util.GetCurrentTime(),
		Organization: user.Owner,
		User:         user.Name,
		Method:       "POST",
		RequestUri:   "/api/" + action,
		Action:       action,
		Object:       util.StructToJson(user),
		StatusCode:   200,
		IsTriggered:  false,
	}

	util.SafeGoroutine(func() {
		AddRecord(record)
	})
}
