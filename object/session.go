// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

var (
	CasdoorApplication  = "app-built-in"
	CasdoorOrganization = "built-in"
)

type Session struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	Application string `xorm:"varchar(100) notnull pk" json:"application"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	SessionId []string `json:"sessionId"`
}

func GetSessions(owner string) ([]*Session, error) {
	sessions := []*Session{}
	var err error
	if owner != "" {
		err = adapter.Engine.Desc("created_time").Where("owner = ?", owner).Find(&sessions)
	} else {
		err = adapter.Engine.Desc("created_time").Find(&sessions)
	}
	if err != nil {
		return sessions, err
	}

	return sessions, nil
}

func GetPaginationSessions(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Session, error) {
	sessions := []*Session{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&sessions)
	if err != nil {
		return sessions, err
	}

	return sessions, nil
}

func GetSessionCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Session{})
}

func GetSingleSession(id string) (*Session, error) {
	owner, name, application := util.GetOwnerAndNameAndOtherFromId(id)
	session := Session{Owner: owner, Name: name, Application: application}
	get, err := adapter.Engine.Get(&session)
	if err != nil {
		return &session, err
	}

	if !get {
		return nil, nil
	}

	return &session, nil
}

func UpdateSession(id string, session *Session) (bool, error) {
	owner, name, application := util.GetOwnerAndNameAndOtherFromId(id)

	if ss, err := GetSingleSession(id); err != nil {
		return false, err
	} else if ss == nil {
		return false, nil
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name, application}).Update(session)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func removeExtraSessionIds(session *Session) {
	if len(session.SessionId) > 100 {
		session.SessionId = session.SessionId[(len(session.SessionId) - 100):]
	}
}

func AddSession(session *Session) (bool, error) {
	dbSession, err := GetSingleSession(session.GetId())
	if err != nil {
		return false, err
	}

	if dbSession == nil {
		session.CreatedTime = util.GetCurrentTime()

		affected, err := adapter.Engine.Insert(session)
		if err != nil {
			return false, err
		}

		return affected != 0, nil
	} else {
		m := make(map[string]struct{})
		for _, v := range dbSession.SessionId {
			m[v] = struct{}{}
		}
		for _, v := range session.SessionId {
			if _, exists := m[v]; !exists {
				dbSession.SessionId = append(dbSession.SessionId, v)
			}
		}

		removeExtraSessionIds(dbSession)

		return UpdateSession(dbSession.GetId(), dbSession)
	}
}

func DeleteSession(id string) (bool, error) {
	owner, name, application := util.GetOwnerAndNameAndOtherFromId(id)
	if owner == CasdoorOrganization && application == CasdoorApplication {
		session, err := GetSingleSession(id)
		if err != nil {
			return false, err
		}

		if session != nil {
			DeleteBeegoSession(session.SessionId)
		}
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name, application}).Delete(&Session{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteSessionId(id string, sessionId string) (bool, error) {
	session, err := GetSingleSession(id)
	if err != nil {
		return false, err
	}
	if session == nil {
		return false, nil
	}

	owner, _, application := util.GetOwnerAndNameAndOtherFromId(id)
	if owner == CasdoorOrganization && application == CasdoorApplication {
		DeleteBeegoSession([]string{sessionId})
	}

	session.SessionId = util.DeleteVal(session.SessionId, sessionId)
	if len(session.SessionId) == 0 {
		return DeleteSession(id)
	} else {
		return UpdateSession(id, session)
	}
}

func DeleteBeegoSession(sessionIds []string) {
	for _, sessionId := range sessionIds {
		err := beego.GlobalSessions.GetProvider().SessionDestroy(sessionId)
		if err != nil {
			return
		}
	}
}

func (session *Session) GetId() string {
	return fmt.Sprintf("%s/%s/%s", session.Owner, session.Name, session.Application)
}

func IsSessionDuplicated(id string, sessionId string) (bool, error) {
	session, err := GetSingleSession(id)
	if err != nil {
		return false, err
	}

	if session == nil {
		return false, nil
	} else {
		if len(session.SessionId) > 1 {
			return true, nil
		} else if len(session.SessionId) < 1 {
			return false, nil
		} else {
			return session.SessionId[0] != sessionId, nil
		}
	}
}
