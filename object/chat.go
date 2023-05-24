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

type Chat struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	Organization string   `xorm:"varchar(100)" json:"organization"`
	DisplayName  string   `xorm:"varchar(100)" json:"displayName"`
	Type         string   `xorm:"varchar(100)" json:"type"`
	Category     string   `xorm:"varchar(100)" json:"category"`
	User1        string   `xorm:"varchar(100)" json:"user1"`
	User2        string   `xorm:"varchar(100)" json:"user2"`
	Users        []string `xorm:"varchar(100)" json:"users"`
	MessageCount int      `json:"messageCount"`
}

func GetMaskedChat(chat *Chat) *Chat {
	if chat == nil {
		return nil
	}

	return chat
}

func GetMaskedChats(chats []*Chat) []*Chat {
	for _, chat := range chats {
		chat = GetMaskedChat(chat)
	}
	return chats
}

func GetChatCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Chat{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetChats(owner string) []*Chat {
	chats := []*Chat{}
	err := adapter.Engine.Desc("created_time").Find(&chats, &Chat{Owner: owner})
	if err != nil {
		panic(err)
	}

	return chats
}

func GetPaginationChats(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Chat {
	chats := []*Chat{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&chats)
	if err != nil {
		panic(err)
	}

	return chats
}

func getChat(owner string, name string) *Chat {
	if owner == "" || name == "" {
		return nil
	}

	chat := Chat{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&chat)
	if err != nil {
		panic(err)
	}

	if existed {
		return &chat
	} else {
		return nil
	}
}

func GetChat(id string) *Chat {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getChat(owner, name)
}

func UpdateChat(id string, chat *Chat) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getChat(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(chat)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddChat(chat *Chat) bool {
	if chat.Type == "AI" && chat.User2 == "" {
		provider := getDefaultAiProvider()
		if provider != nil {
			chat.User2 = provider.Name
		}
	}

	affected, err := adapter.Engine.Insert(chat)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteChat(chat *Chat) bool {
	affected, err := adapter.Engine.ID(core.PK{chat.Owner, chat.Name}).Delete(&Chat{})
	if err != nil {
		panic(err)
	}

	if affected != 0 {
		return DeleteChatMessages(chat.Name)
	}

	return affected != 0
}

func (p *Chat) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}
