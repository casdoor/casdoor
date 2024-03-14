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

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
)

var logPostOnly bool

func init() {
	logPostOnly = conf.GetConfigBool("logPostOnly")
}

type Record struct {
	casvisorsdk.Record
}

func NewRecord(ctx *context.Context) *casvisorsdk.Record {
	ip := strings.Replace(util.GetIPFromRequest(ctx.Request), ": ", "", -1)
	action := strings.Replace(ctx.Request.URL.Path, "/api/", "", -1)
	requestUri := util.FilterQuery(ctx.Request.RequestURI, []string{"accessToken"})
	if len(requestUri) > 1000 {
		requestUri = requestUri[0:1000]
	}

	object := ""
	if ctx.Input.RequestBody != nil && len(ctx.Input.RequestBody) != 0 {
		object = string(ctx.Input.RequestBody)
	}

	language := ctx.Request.Header.Get("Accept-Language")
	if len(language) > 2 {
		language = language[0:2]
	}
	languageCode := conf.GetLanguage(language)

	record := casvisorsdk.Record{
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		ClientIp:    ip,
		User:        "",
		Method:      ctx.Request.Method,
		RequestUri:  requestUri,
		Action:      action,
		Language:    languageCode,
		Object:      object,
		IsTriggered: false,
	}
	return &record
}

func AddRecord(record *casvisorsdk.Record) bool {
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

	if casvisorsdk.GetClient() == nil {
		affected, err := ormer.Engine.Insert(record)
		if err != nil {
			panic(err)
		}

		return affected != 0
	}

	affected, err := casvisorsdk.AddRecord(record)
	if err != nil {
		fmt.Printf("AddRecord() error: %s", err.Error())
	}

	return affected
}

func GetRecordCount(field, value string, filterRecord *casvisorsdk.Record) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Count(filterRecord)
}

func GetRecords() ([]*casvisorsdk.Record, error) {
	records := []*casvisorsdk.Record{}
	err := ormer.Engine.Desc("id").Find(&records)
	if err != nil {
		return records, err
	}

	return records, nil
}

func GetPaginationRecords(offset, limit int, field, value, sortField, sortOrder string, filterRecord *casvisorsdk.Record) ([]*casvisorsdk.Record, error) {
	records := []*casvisorsdk.Record{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&records, filterRecord)
	if err != nil {
		return records, err
	}

	return records, nil
}

func GetRecordsByField(record *casvisorsdk.Record) ([]*casvisorsdk.Record, error) {
	records := []*casvisorsdk.Record{}
	err := ormer.Engine.Find(&records, record)
	if err != nil {
		return records, err
	}

	return records, nil
}

func getFilteredWebhooks(webhooks []*Webhook, organization string, action string) []*Webhook {
	res := []*Webhook{}
	for _, webhook := range webhooks {
		if !webhook.IsEnabled {
			continue
		}

		if webhook.SingleOrgOnly {
			if webhook.Organization != organization {
				continue
			}
		}

		matched := false
		for _, event := range webhook.Events {
			if action == event {
				matched = true
				break
			}
		}

		if matched {
			res = append(res, webhook)
		}
	}
	return res
}

func SendWebhooks(record *casvisorsdk.Record) error {
	webhooks, err := getWebhooksByOrganization("")
	if err != nil {
		return err
	}

	errs := []error{}
	webhooks = getFilteredWebhooks(webhooks, record.Organization, record.Action)
	for _, webhook := range webhooks {
		var user *User
		if webhook.IsUserExtended {
			user, err = getUser(record.Organization, record.User)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			user, err = GetMaskedUser(user, false, err)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}

		err = sendWebhook(webhook, record, user)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) > 0 {
		errStrings := []string{}
		for _, err := range errs {
			errStrings = append(errStrings, err.Error())
		}
		return fmt.Errorf(strings.Join(errStrings, " | "))
	}
	return nil
}
