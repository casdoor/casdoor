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
	"encoding/json"
	"fmt"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func getUser(ctx *context.Context) string {
	if username, ok := ctx.Input.Session("username").(string); ok && username != "" {
		return username
	}

	username, err := getUsernameByClientIdSecret(ctx)
	if err != nil {
		return ""
	}

	return username
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

// getOrganizationFromRequest derives the organization of a request that has no authenticated
// subject, from the organization or application that the request names.
func getOrganizationFromRequest(ctx *context.Context) string {
	var body struct {
		Organization string `json:"organization"`
		Application  string `json:"application"`
		ClientId     string `json:"clientId"`
	}
	if len(ctx.Input.RequestBody) != 0 {
		_ = json.Unmarshal(ctx.Input.RequestBody, &body)
	}

	// the name is caller-controlled, so a made-up one must not become an owner
	for _, name := range []string{body.Organization, ctx.Input.Query("organization")} {
		if name == "" {
			continue
		}

		organization, err := object.GetOrganization(util.GetId("admin", name))
		if err == nil && organization != nil {
			return organization.Name
		}
	}

	applicationId := ctx.Input.Query("applicationId")
	if body.Application != "" {
		applicationId = util.GetId("admin", body.Application)
	}
	if applicationId != "" {
		application, err := object.GetApplication(applicationId)
		if err == nil && application != nil {
			return application.Organization
		}
	}

	clientId := body.ClientId
	if clientId == "" {
		clientId = ctx.Input.Query("clientId")
	}
	if clientId != "" {
		application, err := object.GetApplicationByClientId(clientId)
		if err == nil && application != nil {
			return application.Organization
		}
	}

	return ""
}

func AfterRecordMessage(ctx *context.Context) {
	record, err := object.NewRecord(ctx)
	if err != nil {
		fmt.Printf("AfterRecordMessage() error: %s\n", err.Error())
		return
	}

	userId := ctx.Input.Params()["recordUserId"]
	targetUserId := ctx.Input.Params()["recordTargetUserId"]
	detail := ctx.Input.Params()["recordFailureReason"]
	if detail != "" {
		record.Detail = detail
	}

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
		err = record.SetUser(userId)
		if err != nil {
			panic(err)
		}
	}

	if record.Organization == "" {
		record.Organization = getOrganizationFromRequest(ctx)
	}

	var record2 *object.Record
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
