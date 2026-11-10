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
func AddConsentLog(user *User, application string, action string, grantedScopes []string, deniedScopes []string, clientIp string, userAgent string) error {
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
	_, err := ormer.Engine.Insert(consentLog)
	return err
}

func GetConsentLogCount(owner string, field string, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&ConsentLog{})
}

func GetPaginationConsentLogs(owner string, offset int, limit int, field string, value string, sortField string, sortOrder string) ([]*ConsentLog, error) {
	consentLogs := []*ConsentLog{}
	if sortField == "" || sortOrder == "" {
		sortField, sortOrder = "createdTime", "descend"
	}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&consentLogs)
	return consentLogs, err
}
