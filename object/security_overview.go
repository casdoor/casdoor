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
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type SecurityCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type SecurityDay struct {
	Date    string `json:"date"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
	Signups int    `json:"signups"`
}

// SecurityOverview sums up the sign-ins, failures, admin operations and new places and devices of an organization
// over a period, for the security dashboard and the monthly report
type SecurityOverview struct {
	Organization string `json:"organization"`
	From         string `json:"from"`
	To           string `json:"to"`

	Requests        int     `json:"requests"`
	SigninSuccess   int     `json:"signinSuccess"`
	SigninFailure   int     `json:"signinFailure"`
	FailureRate     float64 `json:"failureRate"`
	Signups         int     `json:"signups"`
	AdminOperations int     `json:"adminOperations"`
	Denied          int     `json:"denied"`
	ActiveUsers     int     `json:"activeUsers"`
	NewLocations    int     `json:"newLocations"`
	NewDevices      int     `json:"newDevices"`
	Truncated       bool    `json:"truncated"`

	Days              []*SecurityDay   `json:"days"`
	TopFailedIps      []*SecurityCount `json:"topFailedIps"`
	TopFailedUsers    []*SecurityCount `json:"topFailedUsers"`
	TopActions        []*SecurityCount `json:"topActions"`
	TopAdmins         []*SecurityCount `json:"topAdmins"`
	RecentSigninRisks []*SigninTrace   `json:"recentSigninRisks"`

	// the tool calls of AI agents through the MCP gateway
	McpCalls           int              `json:"mcpCalls"`
	McpBlocked         int              `json:"mcpBlocked"`
	McpPendingApproval int              `json:"mcpPendingApproval"`
	McpAgents          int              `json:"mcpAgents"`
	TopMcpTools        []*SecurityCount `json:"topMcpTools"`
	TopMcpAgents       []*SecurityCount `json:"topMcpAgents"`
	TopMcpBlockedTools []*SecurityCount `json:"topMcpBlockedTools"`
}

type mcpAuditObject struct {
	Server string `json:"server"`
	Tool   string `json:"tool"`
	Agent  string `json:"agent"`
	Result string `json:"result"`
}

// addMcpOverview counts the tool calls of the period from the audit records, which are kept whether or not the
// organization records the details of the calls
func addMcpOverview(res *SecurityOverview, owner string) error {
	tools, agents, blockedTools := map[string]int{}, map[string]int{}, map[string]int{}
	lastId := 0
	for res.McpCalls < securityOverviewMaxRecords {
		records := []*Record{}
		err := ormer.Engine.Cols("id", "object").
			Where("owner = ? and action = ? and created_time >= ? and created_time < ? and id > ?", owner, "mcp-tool-call", res.From, res.To, lastId).
			Asc("id").Limit(5000).Find(&records)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			lastId = record.Id
			call := mcpAuditObject{}
			if json.Unmarshal([]byte(record.Object), &call) != nil {
				continue
			}
			res.McpCalls++
			tool := call.Server + " / " + call.Tool
			tools[tool]++
			if call.Agent != "" {
				agents[call.Agent]++
			}
			switch call.Result {
			case "Blocked":
				res.McpBlocked++
				blockedTools[tool]++
			case "Pending approval":
				res.McpPendingApproval++
			}
		}
	}
	res.McpAgents = len(agents)
	res.TopMcpTools = topCounts(tools, 10)
	res.TopMcpAgents = topCounts(agents, 10)
	res.TopMcpBlockedTools = topCounts(blockedTools, 10)
	return nil
}

const securityOverviewMaxRecords = 300000

var signinActions = map[string]bool{"login": true, "signin": true}

var nonAdminActions = map[string]bool{
	"login": true, "signin": true, "logout": true, "signout": true, "sso-logout": true, "signup": true,
	"send-verification-code": true, "verify-code": true, "verify-captcha": true, "set-password": true,
}

func topCounts(counts map[string]int, n int) []*SecurityCount {
	res := []*SecurityCount{}
	for name, count := range counts {
		if name == "" {
			continue
		}
		res = append(res, &SecurityCount{Name: name, Count: count})
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Count != res[j].Count {
			return res[i].Count > res[j].Count
		}
		return res[i].Name < res[j].Name
	})
	if len(res) > n {
		res = res[:n]
	}
	return res
}

// GetSecurityOverview reads the organization's records between from and to
func GetSecurityOverview(owner string, from time.Time, to time.Time) (*SecurityOverview, error) {
	res := &SecurityOverview{
		Organization:      owner,
		From:              from.Format(time.RFC3339),
		To:                to.Format(time.RFC3339),
		Days:              []*SecurityDay{},
		RecentSigninRisks: []*SigninTrace{},
	}

	dayMap := map[string]*SecurityDay{}
	for day := from; day.Before(to); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		dayMap[date] = &SecurityDay{Date: date}
		res.Days = append(res.Days, dayMap[date])
	}

	failedIps, failedUsers, actions, admins, activeUsers := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}, map[string]bool{}
	lastId := 0
	for res.Requests < securityOverviewMaxRecords {
		records := []*Record{}
		err := ormer.Engine.Cols("id", "created_time", "client_ip", "user", "method", "action", "response", "status_code").
			Where("owner = ? and created_time >= ? and created_time < ? and id > ?", owner, res.From, res.To, lastId).
			Asc("id").Limit(5000).Find(&records)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			lastId = record.Id
			res.Requests++
			isError := record.StatusCode >= 400 || strings.HasPrefix(record.Response, `{status:"error"`)
			if record.StatusCode == 403 || strings.Contains(record.Response, "Unauthorized operation") {
				res.Denied++
			}

			date := ""
			if t, err := time.Parse(time.RFC3339, record.CreatedTime); err == nil {
				date = t.In(from.Location()).Format("2006-01-02")
			}
			day := dayMap[date]

			if signinActions[record.Action] {
				if isError {
					res.SigninFailure++
					failedIps[record.ClientIp]++
					failedUsers[record.User]++
					if day != nil {
						day.Failure++
					}
				} else {
					res.SigninSuccess++
					activeUsers[record.User] = true
					if day != nil {
						day.Success++
					}
				}
				continue
			}
			if record.Action == "signup" && !isError {
				res.Signups++
				if day != nil {
					day.Signups++
				}
				continue
			}
			// the tool calls through the MCP gateway are counted on their own, see addMcpOverview()
			isMcpCall := record.Action == "mcp-tool-call" || strings.HasPrefix(record.Action, "server/")
			if record.Method != "GET" && !nonAdminActions[record.Action] && !isMcpCall && !isError && record.User != "" {
				res.AdminOperations++
				actions[record.Action]++
				admins[record.User]++
			}
		}
	}
	res.Truncated = res.Requests >= securityOverviewMaxRecords
	delete(activeUsers, "")
	res.ActiveUsers = len(activeUsers)
	if total := res.SigninSuccess + res.SigninFailure; total > 0 {
		res.FailureRate = float64(res.SigninFailure) / float64(total)
	}
	res.TopFailedIps = topCounts(failedIps, 10)
	res.TopFailedUsers = topCounts(failedUsers, 10)
	res.TopActions = topCounts(actions, 10)
	res.TopAdmins = topCounts(admins, 10)

	if err := addMcpOverview(res, owner); err != nil {
		return nil, err
	}

	traces := []*SigninTrace{}
	err := ormer.Engine.Where("owner = ?", owner).Asc("created_time").Find(&traces)
	if err != nil {
		return nil, err
	}
	// a user's first place and first device are where they always sign in, every later one is new
	seen := map[string]bool{}
	risks := []*SigninTrace{}
	for _, trace := range traces {
		key := trace.UserName + "/" + trace.Kind
		isFirst := !seen[key]
		seen[key] = true
		if isFirst || trace.CreatedTime < res.From || trace.CreatedTime >= res.To {
			continue
		}
		if trace.Kind == "location" {
			res.NewLocations++
		} else {
			res.NewDevices++
		}
		risks = append(risks, trace)
	}
	for i := len(risks) - 1; i >= 0 && len(res.RecentSigninRisks) < 20; i-- {
		res.RecentSigninRisks = append(res.RecentSigninRisks, risks[i])
	}
	return res, nil
}
