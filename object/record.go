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
	"strings"

	"github.com/astaxie/beego/context"
	"github.com/casbin/casdoor/util"
)

type Record struct {
	Id int `xorm:"int notnull pk autoincr" json:"id"`

	Owner       string `xorm:"varchar(100) index" json:"owner"`
	Name        string `xorm:"varchar(100) index" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	ClientIp     string `xorm:"varchar(100)" json:"clientIp"`
	User         string `xorm:"varchar(100)" json:"user"`
	RequestUri   string `xorm:"varchar(1000)" json:"requestUri"`
	Action       string `xorm:"varchar(1000)" json:"action"`
}

func NewRecord(ctx *context.Context) *Record {
	ip := strings.Replace(util.GetIPFromRequest(ctx.Request), ": ", "", -1)
	action := strings.Replace(ctx.Request.URL.Path, "/api/", "", -1)

	record := Record{
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		ClientIp:    ip,
		RequestUri:  ctx.Request.RequestURI,
		User:        "",
		Action:      action,
	}
	return &record
}

func AddRecord(record *Record) bool {
	record.Owner = record.Organization

	affected, err := adapter.Engine.Insert(record)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetRecordCount() int {
	count, err := adapter.Engine.Count(&Record{})
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

func GetPaginationRecords(offset, limit int) []*Record {
	records := []*Record{}
	err := adapter.Engine.Desc("id").Limit(limit, offset).Find(&records)
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
