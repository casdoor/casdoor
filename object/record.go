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
	"fmt"
	"strings"

	"github.com/astaxie/beego/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
)

var logPostOnly bool

func init() {
	var err error
	logPostOnly, err = conf.GetConfigBool("logPostOnly")
	if err != nil {
		// panic(err)
	}
}

type Record struct {
	Id int `xorm:"int notnull pk autoincr" json:"id"`

	Owner       string `xorm:"varchar(100) index" json:"owner"`
	Name        string `xorm:"varchar(100) index" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	ClientIp     string `xorm:"varchar(100)" json:"clientIp"`
	User         string `xorm:"varchar(100)" json:"user"`
	Method       string `xorm:"varchar(100)" json:"method"`
	RequestUri   string `xorm:"varchar(1000)" json:"requestUri"`
	Action       string `xorm:"varchar(1000)" json:"action"`

	ExtendedUser *User `xorm:"-" json:"extendedUser"`

	IsTriggered bool `json:"isTriggered"`
}

func NewRecord(ctx *context.Context) *Record {
	ip := strings.Replace(util.GetIPFromRequest(ctx.Request), ": ", "", -1)
	action := strings.Replace(ctx.Request.URL.Path, "/api/", "", -1)
	requestUri := util.FilterQuery(ctx.Request.RequestURI, []string{"accessToken"})
	if len(requestUri) > 1000 {
		requestUri = requestUri[0:1000]
	}

	record := Record{
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		ClientIp:    ip,
		User:        "",
		Method:      ctx.Request.Method,
		RequestUri:  requestUri,
		Action:      action,
		IsTriggered: false,
	}
	return &record
}

func AddRecord(record *Record) bool {
	if logPostOnly {
		if record.Method == "GET" {
			return false
		}
	}

	if record.Organization == "app" {
		return false
	}

	record.Owner = record.Organization

	errWebhook := SendWebhooks(record)
	if errWebhook == nil {
		record.IsTriggered = true
	} else {
		fmt.Println(errWebhook)
	}

	affected, err := adapter.Engine.Insert(record)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetRecordCount(field, value string, filterRecord *Record) int {
	session := GetSession("", -1, -1, field, value, "", "")
	count, err := session.Count(filterRecord)
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetRecords() []*Record {
	records := []*Record{}
	err := adapter.Engine.Desc("id").Find(&records)
	if err != nil {
		panic(err)
	}

	return records
}

func GetPaginationRecords(offset, limit int, field, value, sortField, sortOrder string, filterRecord *Record) []*Record {
	records := []*Record{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&records, filterRecord)
	if err != nil {
		panic(err)
	}

	return records
}

func GetRecordsByField(record *Record) []*Record {
	records := []*Record{}
	err := adapter.Engine.Find(&records, record)
	if err != nil {
		panic(err)
	}

	return records
}

func SendWebhooks(record *Record) error {
	webhooks := getWebhooksByOrganization(record.Organization)
	for _, webhook := range webhooks {
		if !webhook.IsEnabled {
			continue
		}

		matched := false
		for _, event := range webhook.Events {
			if record.Action == event {
				matched = true
				break
			}
		}

		if matched {
			if webhook.IsUserExtended {
				user := getUser(record.Organization, record.User)
				record.ExtendedUser = user
			}

			err := sendWebhook(webhook, record)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
