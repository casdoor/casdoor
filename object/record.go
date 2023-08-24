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

	record := casvisorsdk.Record{
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		ClientIp:    ip,
		User:        "",
		Method:      ctx.Request.Method,
		RequestUri:  requestUri,
		Action:      action,
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
		return false
	}

	affected, err := casvisorsdk.AddRecord(record)
	if err != nil {
		panic(err)
	}

	return affected
}

func SendWebhooks(record *casvisorsdk.Record) error {
	webhooks, err := getWebhooksByOrganization(record.Organization)
	if err != nil {
		return err
	}

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
			var user *User
			if webhook.IsUserExtended {
				user, err = getUser(record.Organization, record.User)
				user, err = GetMaskedUser(user, false, err)
				if err != nil {
					return err
				}
			}

			err = sendWebhook(webhook, record, user)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
