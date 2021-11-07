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

package routers

import (
	"fmt"

	"github.com/astaxie/beego/context"
	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

func getUser(ctx *context.Context) (username string) {
	defer func() {
		if r := recover(); r != nil {
			username = getUserByClientIdSecret(ctx)
		}
	}()

	username = ctx.Input.Session("username").(string)

	if username == "" {
		username = getUserByClientIdSecret(ctx)
	}

	return
}

func getUserByClientIdSecret(ctx *context.Context) string {
	clientId := ctx.Input.Query("clientId")
	clientSecret := ctx.Input.Query("clientSecret")
	if clientId == "" || clientSecret == "" {
		return ""
	}

	application := object.GetApplicationByClientId(clientId)
	if application == nil || application.ClientSecret != clientSecret {
		return ""
	}

	return fmt.Sprintf("%s/%s", application.Organization, application.Name)
}

func RecordMessage(ctx *context.Context) {
	if ctx.Request.URL.Path == "/api/login" {
		return
	}

	record := object.NewRecord(ctx)

	userId := getUser(ctx)
	if userId != "" {
		record.Organization, record.Username = util.GetOwnerAndNameFromId(userId)
	}

	object.AddRecord(record)
}
