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
	"github.com/xorm-io/core"
)

// SamlSession records a SAML assertion issued to an SP, so that SAML Single Logout can later
// address the SP with the same NameID and SessionIndex
type SamlSession struct {
	Owner        string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name         string `xorm:"varchar(100) notnull pk" json:"name"`
	SessionIndex string `xorm:"varchar(100) notnull pk" json:"sessionIndex"`
	CreatedTime  string `xorm:"varchar(100)" json:"createdTime"`

	Application  string `xorm:"varchar(200)" json:"application"`
	SessionId    string `xorm:"varchar(100) index" json:"sessionId"`
	NameId       string `xorm:"varchar(200)" json:"nameId"`
	NameIdFormat string `xorm:"varchar(200)" json:"nameIdFormat"`
	SpEntityId   string `xorm:"varchar(500)" json:"spEntityId"`
}

func addSamlSession(samlSession *SamlSession) error {
	samlSession.CreatedTime = util.GetCurrentTime()

	// A session that just expires is never logged out, so its rows are dropped here instead
	sessions, err := GetUserSessions(samlSession.Owner, samlSession.Name)
	if err != nil {
		return err
	}

	liveIds := []string{samlSession.SessionId}
	for _, session := range sessions {
		liveIds = append(liveIds, session.SessionId...)
	}

	_, err = ormer.Engine.Where("owner = ? and name = ?", samlSession.Owner, samlSession.Name).NotIn("session_id", liveIds).Delete(&SamlSession{})
	if err != nil {
		return err
	}

	_, err = ormer.Engine.Insert(samlSession)
	return err
}

func getUserSamlSessions(owner string, name string, sessionIds []string) ([]*SamlSession, error) {
	samlSessions := []*SamlSession{}
	session := ormer.Engine.Where("owner = ? and name = ?", owner, name)
	if sessionIds != nil {
		session = session.In("session_id", sessionIds)
	}

	err := session.Find(&samlSessions)
	return samlSessions, err
}

func deleteSamlSession(samlSession *SamlSession) error {
	_, err := ormer.Engine.ID(core.PK{samlSession.Owner, samlSession.Name, samlSession.SessionIndex}).Delete(&SamlSession{})
	return err
}
