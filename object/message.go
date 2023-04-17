// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Message struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	Chat         string `xorm:"varchar(100) index" json:"chat"`
	Author       string `xorm:"varchar(100)" json:"author"`
	Text         string `xorm:"mediumtext" json:"text"`
}

func GetMaskedMessage(message *Message) *Message {
	if message == nil {
		return nil
	}

	return message
}

func GetMaskedMessages(messages []*Message) []*Message {
	for _, message := range messages {
		message = GetMaskedMessage(message)
	}
	return messages
}

func GetMessageCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Message{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetMessages(owner string) []*Message {
	messages := []*Message{}
	err := adapter.Engine.Desc("created_time").Find(&messages, &Message{Owner: owner})
	if err != nil {
		panic(err)
	}

	return messages
}

func GetChatMessages(chat string) []*Message {
	messages := []*Message{}
	err := adapter.Engine.Asc("created_time").Find(&messages, &Message{Chat: chat})
	if err != nil {
		panic(err)
	}

	return messages
}

func GetPaginationMessages(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Message {
	messages := []*Message{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&messages)
	if err != nil {
		panic(err)
	}

	return messages
}

func getMessage(owner string, name string) *Message {
	if owner == "" || name == "" {
		return nil
	}

	message := Message{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&message)
	if err != nil {
		panic(err)
	}

	if existed {
		return &message
	} else {
		return nil
	}
}

func GetMessage(id string) *Message {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getMessage(owner, name)
}

func UpdateMessage(id string, message *Message) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getMessage(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(message)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddMessage(message *Message) bool {
	affected, err := adapter.Engine.Insert(message)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteMessage(message *Message) bool {
	affected, err := adapter.Engine.ID(core.PK{message.Owner, message.Name}).Delete(&Message{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (p *Message) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}
