// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// PersonalInfoFields are the fields masked in the console when the organization's policy asks so
var PersonalInfoFields = []string{"phone", "email", "idCard"}

func getMaskedIdCard(idCard string) string {
	runes := []rune(idCard)
	if len(runes) <= 7 {
		return strings.Repeat("*", len(runes))
	}
	return string(runes[:3]) + strings.Repeat("*", len(runes)-7) + string(runes[len(runes)-4:])
}

func getMaskedPersonalValue(field string, value string) string {
	if value == "" {
		return ""
	}
	switch field {
	case "phone":
		masked := util.GetMaskedPhone(value)
		if masked == value && len(value) > 4 {
			masked = strings.Repeat("*", len(value)-4) + value[len(value)-4:]
		}
		return masked
	case "email":
		return util.GetMaskedEmail(value)
	case "idCard":
		return getMaskedIdCard(value)
	}
	return value
}

func getPersonalValue(user *object.User, field string) string {
	switch field {
	case "phone":
		return user.Phone
	case "email":
		return user.Email
	case "idCard":
		return user.IdCard
	}
	return ""
}

func setPersonalValue(user *object.User, field string, value string) {
	switch field {
	case "phone":
		user.Phone = value
	case "email":
		user.Email = value
	case "idCard":
		user.IdCard = value
	}
}

// IsPersonalInfoMasked tells whether the console shows the phone, email and ID card of the organization's users masked
func IsPersonalInfoMasked(owner string) bool {
	policy := GetCachedSecurityPolicy(owner)
	return policy != nil && policy.MaskSensitiveData
}

// MaskPersonalInfo masks the user's phone, email and ID card for the console. The user's own requests and the
// applications calling the API aren't masked.
func MaskPersonalInfo(user *object.User) {
	if user == nil || !IsPersonalInfoMasked(user.Owner) {
		return
	}
	for _, field := range PersonalInfoFields {
		setPersonalValue(user, field, getMaskedPersonalValue(field, getPersonalValue(user, field)))
	}
}

// RestoreMaskedPersonalInfo keeps the stored values when a console form sends the masked ones back
func RestoreMaskedPersonalInfo(newUser *object.User, oldUser *object.User) {
	if newUser == nil || oldUser == nil || !IsPersonalInfoMasked(oldUser.Owner) {
		return
	}
	for _, field := range PersonalInfoFields {
		oldValue := getPersonalValue(oldUser, field)
		if oldValue != "" && getPersonalValue(newUser, field) == getMaskedPersonalValue(field, oldValue) {
			setPersonalValue(newUser, field, oldValue)
		}
	}
}

// GetPersonalInfoValue returns one unmasked field, for an administrator who clicks to show it
func GetPersonalInfoValue(user *object.User, field string) (string, error) {
	if !util.InSlice(PersonalInfoFields, field) {
		return "", fmt.Errorf("unknown field: %s", field)
	}
	return getPersonalValue(user, field), nil
}

// ExportUserData collects what Casdoor keeps about the user, for the user to download
func ExportUserData(user *object.User) (map[string]interface{}, error) {
	exported, err := object.GetMaskedUser(user, true)
	if err != nil {
		return nil, err
	}
	exported.Password = ""
	exported.PasswordSalt = ""
	exported.TotpSecret = ""
	exported.RecoveryCodes = nil
	exported.OriginalToken = ""
	exported.OriginalRefreshToken = ""

	sessions := []*object.Session{}
	err = object.GetOrmer().Engine.Find(&sessions, &object.Session{Owner: user.Owner, Name: user.Name})
	if err != nil {
		return nil, err
	}

	traces, err := getSigninTraces(user.Owner, user.Name)
	if err != nil {
		return nil, err
	}

	records := []*object.Record{}
	err = object.GetOrmer().Engine.Desc("id").Limit(1000).Cols("created_time", "client_ip", "action", "request_uri", "status_code").
		Find(&records, &object.Record{Organization: user.Owner, User: user.Name})
	if err != nil {
		return nil, err
	}

	tokens := []*object.Token{}
	err = object.GetOrmer().Engine.Desc("created_time").Limit(1000).Cols("created_time", "application", "scope", "expires_in").
		Find(&tokens, &object.Token{Organization: user.Owner, User: user.Name})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"exportedTime":   util.GetCurrentTime(),
		"user":           exported,
		"sessions":       sessions,
		"signinTraces":   traces,
		"recentActivity": records,
		"tokens":         tokens,
	}, nil
}
