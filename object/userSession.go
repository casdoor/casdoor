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
	"github.com/beego/beego"
)

type userSession struct {
	Id        string   `xorm:"varchar(255) notnull pk" json:"Id"`
	SessionId []string `json:"sessionId"`
}

func SetUserSession(id string, session string) {
	userSession := &userSession{Id: id}
	exist, err := adapter.Engine.Exist(userSession)
	if err != nil {
		return
	}

	userSession.SessionId = append(userSession.SessionId, session)
	if exist {
		_, err = adapter.Engine.ID(id).Update(userSession)
	} else {
		_, err = adapter.Engine.Insert(userSession)
	}

	if err != nil {
		return
	}
}

func DeleteUserSession(id string) {
	userSession := &userSession{Id: id}
	adapter.Engine.ID(id).Get(userSession)
	DeleteSession(userSession.SessionId)

	_, err := adapter.Engine.ID(id).Delete(userSession)
	if err != nil {
		return
	}
}

func DeleteSession(sessionIds []string) {
	for _, sessionId := range sessionIds {
		err := beego.GlobalSessions.GetProvider().SessionDestroy(sessionId)
		if err != nil {
			return
		}
	}
}
