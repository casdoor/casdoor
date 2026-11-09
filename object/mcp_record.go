// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/util"
)

// McpRecord is one tool call through the MCP gateway, kept when the server owner's security policy asks so
type McpRecord struct {
	Id          int    `xorm:"int notnull pk autoincr" json:"id"`
	Owner       string `xorm:"varchar(100) index" json:"owner"`
	Server      string `xorm:"varchar(100) index" json:"server"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	UserName   string `xorm:"varchar(200)" json:"user"`
	ClientIp   string `xorm:"varchar(100)" json:"clientIp"`
	UserAgent  string `xorm:"varchar(500)" json:"userAgent"`
	Tool       string `xorm:"varchar(200) index" json:"tool"`
	Arguments  string `xorm:"mediumtext" json:"arguments"`
	Result     string `xorm:"varchar(20)" json:"result"`
	StatusCode int    `json:"statusCode"`
	Response   string `xorm:"mediumtext" json:"response"`
	DurationMs int64  `json:"durationMs"`
}

func IsMcpToolCallRecorded(owner string) bool {
	policy := GetCachedSecurityPolicy(owner)
	return policy != nil && policy.RecordMcpToolCalls
}

func AddMcpRecord(record *McpRecord) {
	record.CreatedTime = util.GetCurrentTime()
	record.Arguments = truncateString(record.Arguments, 4000)
	record.Response = truncateString(record.Response, 4000)
	record.UserAgent = truncateString(record.UserAgent, 500)
	if _, err := ormer.Engine.Insert(record); err != nil {
		logs.Warn("the MCP tool call record is not saved: %s", err.Error())
	}
}

func GetMcpRecordCount(owner string, server string, tool string) (int64, error) {
	session := ormer.Engine.Where("owner = ?", owner)
	if server != "" {
		session = session.And("server = ?", server)
	}
	if tool != "" {
		session = session.And("tool = ?", tool)
	}
	return session.Count(&McpRecord{})
}

func GetPaginationMcpRecords(owner string, server string, tool string, offset int, limit int) ([]*McpRecord, error) {
	records := []*McpRecord{}
	session := ormer.Engine.Where("owner = ?", owner)
	if server != "" {
		session = session.And("server = ?", server)
	}
	if tool != "" {
		session = session.And("tool = ?", tool)
	}
	err := session.Desc("id").Limit(limit, offset).Find(&records)
	return records, err
}
