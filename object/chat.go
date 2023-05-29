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

func GetMaskedChat(chat *Chat, err ...error) (*Chat, error) {
	if len(err) > 0 && err[0] != nil {
		return nil, err[0]
	}

	if chat == nil {
		return nil, nil
	}

	return chat, nil
}

func GetMaskedChats(chats []*Chat, errs ...error) ([]*Chat, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}
	var err error
	for _, chat := range chats {
		chat, err = GetMaskedChat(chat)
		if err != nil {
			return nil, err
		}
	}
	return chats, nil
}

func GetChatCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Chat{})
}

func GetChats(owner string) ([]*Chat, error) {
	chats := []*Chat{}
	err := adapter.Engine.Desc("created_time").Find(&chats, &Chat{Owner: owner})
	if err != nil {
		return chats, err
	}

	return chats, nil
}

func GetPaginationChats(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Chat, error) {
	chats := []*Chat{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&chats)
	if err != nil {
		return chats, err
	}

	return chats, nil
}

func getChat(owner string, name string) (*Chat, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	chat := Chat{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&chat)
	if err != nil {
		return &chat, err
	}

	if existed {
		return &chat, nil
	} else {
		return nil, nil
	}
}

func GetChat(id string) (*Chat, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getChat(owner, name)
}

func UpdateChat(id string, chat *Chat) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if c, err := getChat(owner, name); err != nil {
		return false, err
	} else if c == nil {
		return false, nil
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(chat)
	if err != nil {
		return false, nil
	}

	return affected != 0, nil
}

func AddChat(chat *Chat) (bool, error) {
	if chat.Type == "AI" && chat.User2 == "" {
		provider, err := getDefaultAiProvider()
		if err != nil {
			return false, err
		}

		if provider != nil {
			chat.User2 = provider.Name
		}
	}

	affected, err := adapter.Engine.Insert(chat)
	if err != nil {
		return false, nil
	}

	return affected != 0, nil
}

func DeleteChat(chat *Chat) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{chat.Owner, chat.Name}).Delete(&Chat{})
	if err != nil {
		return false, err
	}

	if affected != 0 {
		return DeleteChatMessages(chat.Name)
	}

	return affected != 0, nil
}

func (p *Chat) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}
