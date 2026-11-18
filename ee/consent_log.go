// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// ConsentLog is one consent or revocation of a user to an application, kept for organizations whose enterprise
// policy records consents, so that the consent can be proven later (PIPL article 14)
type ConsentLog struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100) index" json:"createdTime"`

	User          string   `xorm:"varchar(100) index" json:"user"`
	Application   string   `xorm:"varchar(100) index" json:"application"`
	Action        string   `xorm:"varchar(20)" json:"action"`
	GrantedScopes []string `xorm:"mediumtext" json:"grantedScopes"`
	DeniedScopes  []string `xorm:"mediumtext" json:"deniedScopes"`
	ClientIp      string   `xorm:"varchar(100)" json:"clientIp"`
	UserAgent     string   `xorm:"varchar(500)" json:"userAgent"`
}

// AddConsentLog records a "grant" or "revoke" when the user's organization records consents
func AddConsentLog(user *object.User, application string, action string, grantedScopes []string, deniedScopes []string, clientIp string, userAgent string) error {
	policy := GetCachedEnterprisePolicy(user.Owner)
	if policy == nil || !policy.RecordConsents {
		return nil
	}
	if len(userAgent) > 500 {
		userAgent = userAgent[:500]
	}
	consentLog := &ConsentLog{
		Owner:         user.Owner,
		Name:          util.GenerateId(),
		CreatedTime:   util.GetCurrentTime(),
		User:          user.GetId(),
		Application:   application,
		Action:        action,
		GrantedScopes: grantedScopes,
		DeniedScopes:  deniedScopes,
		ClientIp:      clientIp,
		UserAgent:     userAgent,
	}
	_, err := object.GetOrmer().Engine.Insert(consentLog)
	return err
}

func GetConsentLogCount(owner string, field string, value string) (int64, error) {
	session := object.GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&ConsentLog{})
}

func GetPaginationConsentLogs(owner string, offset int, limit int, field string, value string, sortField string, sortOrder string) ([]*ConsentLog, error) {
	consentLogs := []*ConsentLog{}
	if sortField == "" || sortOrder == "" {
		sortField, sortOrder = "createdTime", "descend"
	}
	session := object.GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&consentLogs)
	return consentLogs, err
}
