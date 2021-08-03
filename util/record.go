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

package util

import (
	"strings"

	"github.com/astaxie/beego/context"
)

type Record struct {
	ClientIp		string	`xorm:"varchar(100)" json:"clientIp"`
	Timestamp		string	`xorm:"varchar(100)" json:"timestamp"`
	Organization		string	`xorm:"varchar(100)" json:"organization"`
	Username		string	`xorm:"varchar(100)" json:"username"`
	RequestUri		string	`xorm:"varchar(1000)" json:"requestUri"`
	Action			string	`xorm:"varchar(1000)" json:"action"`
}

func Records(ctx *context.Context) *Record {
	ip := strings.Replace(GetIPFromRequest(ctx.Request), ": ", "", -1)
	currenttime := GetCurrentTime()
	requesturi := ctx.Request.RequestURI
	action := strings.Replace(ctx.Request.URL.Path, "/api/", "", -1)

	record := Record{
		ClientIp:     ip,
		Timestamp:    currenttime,
		RequestUri:   requesturi,
		Username:     "",
		Organization: "",
		Action:       action,
	}
	return &record
}
