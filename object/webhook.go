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

import (
	"fmt"

	"github.com/casbin/casdoor/util"
	"xorm.io/core"
)

type Webhook struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Url         string   `xorm:"varchar(100)" json:"url"`
	ContentType string   `xorm:"varchar(100)" json:"contentType"`
	Events      []string `xorm:"varchar(100)" json:"events"`

	Organization string `xorm:"varchar(100)" json:"organization"`
}

func GetWebhookCount(owner string) int {
	count, err := adapter.Engine.Count(&Webhook{Owner: owner})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetWebhooks(owner string) []*Webhook {
	webhooks := []*Webhook{}
	err := adapter.Engine.Desc("created_time").Find(&webhooks, &Webhook{Owner: owner})
	if err != nil {
		panic(err)
	}

	return webhooks
}

func GetPaginationWebhooks(owner string, offset, limit int) []*Webhook {
	webhooks := []*Webhook{}
	err := adapter.Engine.Desc("created_time").Limit(limit, offset).Find(&webhooks, &Webhook{Owner: owner})
	if err != nil {
		panic(err)
	}

	return webhooks
}

func getWebhook(owner string, name string) *Webhook {
	if owner == "" || name == "" {
		return nil
	}

	webhook := Webhook{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&webhook)
	if err != nil {
		panic(err)
	}

	if existed {
		return &webhook
	} else {
		return nil
	}
}

func GetWebhook(id string) *Webhook {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getWebhook(owner, name)
}

func UpdateWebhook(id string, webhook *Webhook) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getWebhook(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(webhook)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddWebhook(webhook *Webhook) bool {
	affected, err := adapter.Engine.Insert(webhook)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteWebhook(webhook *Webhook) bool {
	affected, err := adapter.Engine.ID(core.PK{webhook.Owner, webhook.Name}).Delete(&Webhook{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (p *Webhook) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}
