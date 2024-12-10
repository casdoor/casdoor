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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
)

var (
	logPostOnly   bool
	passwordRegex *regexp.Regexp
)

func init() {
	logPostOnly = conf.GetConfigBool("logPostOnly")
	passwordRegex = regexp.MustCompile("\"password\":\"([^\"]*?)\"")
}

type Record struct {
	casvisorsdk.Record
}

type Response struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func maskPassword(recordString string) string {
	return passwordRegex.ReplaceAllString(recordString, "\"password\":\"***\"")
}

func NewRecord(ctx *context.Context) (*casvisorsdk.Record, error) {
	clientIp := strings.Replace(util.GetClientIpFromRequest(ctx.Request), ": ", "", -1)
	action := strings.Replace(ctx.Request.URL.Path, "/api/", "", -1)
	requestUri := util.FilterQuery(ctx.Request.RequestURI, []string{"accessToken"})
	if len(requestUri) > 1000 {
		requestUri = requestUri[0:1000]
	}

	object := ""
	if ctx.Input.RequestBody != nil && len(ctx.Input.RequestBody) != 0 {
		object = string(ctx.Input.RequestBody)
		object = maskPassword(object)
	}

	respBytes, err := json.Marshal(ctx.Input.Data()["json"])
	if err != nil {
		return nil, err
	}

	var resp Response
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		return nil, err
	}

	language := ctx.Request.Header.Get("Accept-Language")
	if len(language) > 2 {
		language = language[0:2]
	}
	languageCode := conf.GetLanguage(language)

	record := casvisorsdk.Record{
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		ClientIp:    clientIp,
		User:        "",
		Method:      ctx.Request.Method,
		RequestUri:  requestUri,
		Action:      action,
		Language:    languageCode,
		Object:      object,
		StatusCode:  200,
		Response:    fmt.Sprintf("{status:\"%s\", msg:\"%s\"}", resp.Status, resp.Msg),
		IsTriggered: false,
	}
	return &record, nil
}

func addRecord(record *casvisorsdk.Record) (int64, error) {
	affected, err := ormer.Engine.Insert(record)
	return affected, err
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
	record.Object = maskPassword(record.Object)

	errWebhook := SendWebhooks(record)
	if errWebhook == nil {
		record.IsTriggered = true
	} else {
		fmt.Println(errWebhook)
	}

	if casvisorsdk.GetClient() == nil {
		affected, err := addRecord(record)
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

	if sortField == "" || sortOrder == "" {
		sortField = "id"
		sortOrder = "descend"
	}

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

func CopyRecord(record *casvisorsdk.Record) *casvisorsdk.Record {
	res := &casvisorsdk.Record{
		Owner:        record.Owner,
		Name:         record.Name,
		CreatedTime:  record.CreatedTime,
		Organization: record.Organization,
		ClientIp:     record.ClientIp,
		User:         record.User,
		Method:       record.Method,
		RequestUri:   record.RequestUri,
		Action:       record.Action,
		Language:     record.Language,
		Object:       record.Object,
		Response:     record.Response,
		IsTriggered:  record.IsTriggered,
	}
	return res
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

func addWebhookRecord(webhook *Webhook, record *casvisorsdk.Record, statusCode int, respBody string, sendError error) error {
	if statusCode == 200 {
		return nil
	}

	if len(respBody) > 300 {
		respBody = respBody[0:300]
	}

	webhookRecord := &casvisorsdk.Record{
		Owner:        record.Owner,
		Name:         util.GenerateId(),
		CreatedTime:  util.GetCurrentTime(),
		Organization: record.Organization,
		User:         record.User,

		Method:      webhook.Method,
		Action:      "send-webhook",
		RequestUri:  webhook.Url,
		StatusCode:  statusCode,
		Response:    respBody,
		Language:    record.Language,
		IsTriggered: false,
	}

	if sendError != nil {
		webhookRecord.Response = sendError.Error()
	}

	_, err := addRecord(webhookRecord)

	return err
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

		statusCode, respBody, err := sendWebhook(webhook, record, user)
		if err != nil {
			errs = append(errs, err)
		}

		err = addWebhookRecord(webhook, record, statusCode, respBody, err)
		if err != nil {
			errs = append(errs, err)
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
