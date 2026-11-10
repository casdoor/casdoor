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
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/xorm"
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

	// who called: the OAuth client (application), the agent it is, the RFC 8693 actor chain and how it authenticated
	Application string `xorm:"varchar(100)" json:"application"`
	Agent       string `xorm:"varchar(200) index" json:"agent"`
	Actor       string `xorm:"varchar(1000)" json:"actor"`
	AuthMethod  string `xorm:"varchar(100)" json:"authMethod"`
	TokenId     string `xorm:"varchar(200)" json:"tokenId"`
	// the trace ID that the gateway passes to the upstream MCP server, which passes it on to the business systems
	TraceId string `xorm:"varchar(100) index" json:"traceId"`
	// the jti of the identity token sent to the upstream MCP server
	IdentityTokenId string `xorm:"varchar(200)" json:"identityTokenId"`
	Approval        string `xorm:"varchar(200)" json:"approval"`
}

// GetMcpRecordBodyLimit keeps the whole arguments and response of a tool call for the enterprise edition, as the
// evidence of an incident review, and the beginning of them otherwise
func GetMcpRecordBodyLimit() int {
	if edition.Has(edition.Enterprise) {
		return 1 << 20
	}
	return 4000
}

func IsMcpToolCallRecorded(owner string) bool {
	policy := GetCachedSecurityPolicy(owner)
	return policy != nil && policy.RecordMcpToolCalls
}

func AddMcpRecord(record *McpRecord) {
	record.CreatedTime = util.GetCurrentTime()
	record.Arguments = truncateString(record.Arguments, GetMcpRecordBodyLimit())
	record.Response = truncateString(record.Response, GetMcpRecordBodyLimit())
	record.UserAgent = truncateString(record.UserAgent, 500)
	if _, err := ormer.Engine.Insert(record); err != nil {
		logs.Warn("the MCP tool call record is not saved: %s", err.Error())
	}
}

type McpRecordFilter struct {
	Server  string
	Tool    string
	Agent   string
	User    string
	TraceId string
	Result  string
}

func getMcpRecordSession(owner string, filter *McpRecordFilter) *xorm.Session {
	session := ormer.Engine.Where("owner = ?", owner)
	if filter.Server != "" {
		session = session.And("server = ?", filter.Server)
	}
	if filter.Tool != "" {
		session = session.And("tool = ?", filter.Tool)
	}
	if filter.Agent != "" {
		session = session.And("agent = ?", filter.Agent)
	}
	if filter.User != "" {
		session = session.And("user_name = ?", filter.User)
	}
	if filter.TraceId != "" {
		session = session.And("trace_id = ?", filter.TraceId)
	}
	if filter.Result != "" {
		session = session.And("result = ?", filter.Result)
	}
	return session
}

func GetMcpRecordCount(owner string, filter *McpRecordFilter) (int64, error) {
	return getMcpRecordSession(owner, filter).Count(&McpRecord{})
}

func GetPaginationMcpRecords(owner string, filter *McpRecordFilter, offset int, limit int) ([]*McpRecord, error) {
	records := []*McpRecord{}
	err := getMcpRecordSession(owner, filter).Desc("id").Limit(limit, offset).Find(&records)
	return records, err
}
