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

package routers

import (
	"fmt"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
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

	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		panic(err)
	}

	if application == nil || application.ClientSecret != clientSecret {
		return ""
	}

	return util.GetId(application.Organization, application.Name)
}

func RecordMessage(ctx *context.Context) {
	if ctx.Request.URL.Path == "/api/login" || ctx.Request.URL.Path == "/api/signup" {
		return
	}

	userId := getUser(ctx)

	// Special handling for set-password endpoint to capture target user
	if ctx.Request.URL.Path == "/api/set-password" {
		// Parse form if not already parsed
		if err := ctx.Request.ParseForm(); err != nil {
			fmt.Printf("RecordMessage() error parsing form: %s\n", err.Error())
		} else {
			userOwner := ctx.Request.Form.Get("userOwner")
			userName := ctx.Request.Form.Get("userName")

			if userOwner != "" && userName != "" {
				targetUserId := util.GetId(userOwner, userName)
				ctx.Input.SetParam("recordTargetUserId", targetUserId)
			}
		}
	}

	ctx.Input.SetParam("recordUserId", userId)
}

func AfterRecordMessage(ctx *context.Context) {
	record, err := object.NewRecord(ctx)
	if err != nil {
		fmt.Printf("AfterRecordMessage() error: %s\n", err.Error())
		return
	}

	userId := ctx.Input.Params()["recordUserId"]
	targetUserId := ctx.Input.Params()["recordTargetUserId"]

	// For set-password endpoint, use target user if available
	// We use defensive error handling here (log instead of panic) because target user
	// parsing is a new feature. If it fails, we gracefully fall back to the regular
	// userId flow or empty user/org fields, maintaining backward compatibility.
	if record.Action == "set-password" && targetUserId != "" {
		owner, user, err := util.GetOwnerAndNameFromIdWithError(targetUserId)
		if err != nil {
			fmt.Printf("AfterRecordMessage() error parsing target user %s: %s\n", targetUserId, err.Error())
		} else {
			record.Organization, record.User = owner, user
		}
	} else if userId != "" {
		owner, user, err := util.GetOwnerAndNameFromIdWithError(userId)
		if err != nil {
			panic(err)
		}
		record.Organization, record.User = owner, user
	}

	var record2 *casvisorsdk.Record
	recordSignup := ctx.Input.Params()["recordSignup"]
	if recordSignup == "true" {
		record2 = object.CopyRecord(record)
		record2.Action = "new-user"

		var user *object.User
		user, err = object.GetUser(userId)
		if err != nil {
			fmt.Printf("AfterRecordMessage() error: %s\n", err.Error())
			return
		}
		if user == nil {
			err = fmt.Errorf("the user: %s is not found", userId)
			fmt.Printf("AfterRecordMessage() error: %s\n", err.Error())
			return
		}

		record2.Object = util.StructToJson(user)
	}

	util.SafeGoroutine(func() {
		object.AddRecord(record)

		if record2 != nil {
			object.AddRecord(record2)
		}
	})
}
