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
	"time"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Session struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	SessionId []string `json:"sessionId"`
}

func SetSession(id string, sessionId string) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	session := &Session{Owner: owner, Name: name}
	get, err := adapter.Engine.Get(session)
	if err != nil {
		panic(err)
	}

	session.SessionId = append(session.SessionId, sessionId)
	if get {
		_, err = adapter.Engine.ID(core.PK{owner, name}).Update(session)
	} else {
		session.CreatedTime = time.Now().Format(time.RFC3339)
		_, err = adapter.Engine.Insert(session)
	}
	if err != nil {
		panic(err)
	}
}

func DeleteSession(id string) bool {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)

	session := &Session{Owner: owner, Name: name}
	_, err := adapter.Engine.ID(core.PK{owner, name}).Get(session)
	if err != nil {
		return false
	}

	DeleteBeegoSession(session.SessionId)

	affected, err := adapter.Engine.ID(core.PK{owner, name}).Delete(session)
	return affected != 0
}

func DeleteSessionId(id string, sessionId string) bool {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)

	session := &Session{Owner: owner, Name: name}
	_, err := adapter.Engine.ID(core.PK{owner, name}).Get(session)
	if err != nil {
		return false
	}

	DeleteBeegoSession([]string{sessionId})
	session.SessionId = util.DeleteVal(session.SessionId, sessionId)

	if len(session.SessionId) < 1 {
		affected, _ := adapter.Engine.ID(core.PK{owner, name}).Delete(session)
		return affected != 0
	} else {
		affected, _ := adapter.Engine.ID(core.PK{owner, name}).Update(session)
		return affected != 0
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

func GetSessions(owner string) []*Session {
	sessions := []*Session{}
	var err error
	if owner != "" {
		err = adapter.Engine.Desc("created_time").Where("owner = ?", owner).Find(&sessions)
	} else {
		err = adapter.Engine.Desc("created_time").Find(&sessions)
	}
	if err != nil {
		panic(err)
	}

	return sessions
}

func GetPaginationSessions(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Session {
	sessions := []*Session{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&sessions)
	if err != nil {
		panic(err)
	}

	return sessions
}

func GetSessionCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Session{})
	if err != nil {
		panic(err)
	}

	return int(count)
}
