// Copyright 2021 The casbin Authors. All Rights Reserved.
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

import "xorm.io/core"

type Client struct {
	Name     string `xorm:"varchar(128) not null pk" json:"name"`
	ID       string `xorm:"'id' varchar(128) notnull pk" json:"clientId"`
	Callback string `xorm:"varchar(128) not null" json:"callback"`
	Secret   string `xorm:"varchar(128) notnull" json:"clientSecret"`
	Domain   string `xorm:"varchar(128) notnull" json:"domain"`
	UserID   string `xorm:"'user_id' varchar(128) notnull" json:"userId"`
}

func GetClientByID(id string) *Client {
	client := Client{ID: id}
	existed, err := adapter.engine.Get(&client)
	if err != nil {
		panic(err)
	}

	if existed {
		return &client
	} 
	return nil
}

func GetClientByUserID(userID string) []*Client {
	var clients []*Client
	err := adapter.engine.Find(&clients, &Client{UserID: userID})
	if err != nil {
		panic(err)
	}

	return clients
}

func AddClient(client *Client) bool {
	affected, err := adapter.engine.Insert(client)
	if err != nil {
		panic(err)
	}
	return affected != 0
}

func DeleteClient(client *Client) bool {
	affected, err := adapter.engine.Id(core.PK{client.Name, client.ID}).Delete(&Client{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetClientByUserIDAndName(userID, name string) *Client {
	client := Client{UserID: userID, Name: name}
	existed, err := adapter.engine.Get(&client)
	if err != nil {
		panic(err)
	}

	if existed {
		return &client
	} 
	return nil
}

func UpdateClient(id string, client *Client) bool {
	affected, err := adapter.engine.Update(client)
	if err != nil {
		panic(err)
	}
	return affected != 0
}