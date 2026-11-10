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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	McpApprovalPending  = "Pending"
	McpApprovalApproved = "Approved"
	McpApprovalRejected = "Rejected"
	McpApprovalUsed     = "Used"

	// how long a request waits for an approver, and how long an approved call can be made
	mcpApprovalPendingHours  = 24
	mcpApprovalApprovedHours = 1
)

// McpApproval is a call of a tool that needs a person's approval: the agent's call is held, an admin of the
// organization approves or rejects it, and the same call (user, agent, tool and arguments) then goes through once
type McpApproval struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Server        string `xorm:"varchar(100) index" json:"server"`
	Tool          string `xorm:"varchar(200)" json:"tool"`
	User          string `xorm:"varchar(200)" json:"user"`
	Agent         string `xorm:"varchar(200)" json:"agent"`
	Application   string `xorm:"varchar(100)" json:"application"`
	Arguments     string `xorm:"mediumtext" json:"arguments"`
	ArgumentsHash string `xorm:"varchar(100) index" json:"argumentsHash"`
	TraceId       string `xorm:"varchar(100)" json:"traceId"`
	ClientIp      string `xorm:"varchar(100)" json:"clientIp"`

	State       string `xorm:"varchar(100) index" json:"state"`
	Approver    string `xorm:"varchar(200)" json:"approver"`
	ApproveTime string `xorm:"varchar(100)" json:"approveTime"`
	Comment     string `xorm:"varchar(1000)" json:"comment"`
	ExpireTime  string `xorm:"varchar(100)" json:"expireTime"`
	UsedTime    string `xorm:"varchar(100)" json:"usedTime"`
	UsedTraceId string `xorm:"varchar(100)" json:"usedTraceId"`
}

func (approval *McpApproval) GetId() string {
	return util.GetId(approval.Owner, approval.Name)
}

func (approval *McpApproval) isExpired() bool {
	expireTime, err := time.Parse(time.RFC3339, approval.ExpireTime)
	return err == nil && time.Now().After(expireTime)
}

// GetMcpApprovalUrl is the page where an admin approves the call
func GetMcpApprovalUrl(host string, approval *McpApproval) string {
	originFrontend, _ := getOriginFromHost(host)
	return fmt.Sprintf("%s/mcp-approvals/%s/%s", originFrontend, url.PathEscape(approval.Owner), url.PathEscape(approval.Name))
}

func getMcpArgumentsHash(arguments string) string {
	hash := sha256.Sum256([]byte(arguments))
	return hex.EncodeToString(hash[:])
}

func getMcpApprovals(owner string, state string) ([]*McpApproval, error) {
	approvals := []*McpApproval{}
	session := ormer.Engine.NewSession()
	defer session.Close()
	if owner != "" {
		session = session.Where("owner = ?", owner)
	}
	if state != "" {
		session = session.And("state = ?", state)
	}
	err := session.Desc("created_time").Limit(500).Find(&approvals)
	return approvals, err
}

func GetMcpApprovals(owner string, state string) ([]*McpApproval, error) {
	approvals, err := getMcpApprovals(owner, state)
	if err != nil {
		return nil, err
	}
	for _, approval := range approvals {
		if approval.State == McpApprovalPending && approval.isExpired() {
			approval.State = "Expired"
		}
	}
	return approvals, nil
}

func GetMcpApproval(id string) (*McpApproval, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	approval := McpApproval{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&approval)
	if err != nil || !existed {
		return nil, err
	}
	return &approval, nil
}

func GetPendingMcpApprovalCount(owner string) (int64, error) {
	return ormer.Engine.Where("owner = ? and state = ? and expire_time > ?", owner, McpApprovalPending, util.GetCurrentTime()).Count(&McpApproval{})
}

// CheckMcpApproval lets a call of a tool that needs approval through when an approved and unused approval matches it,
// and marks the approval used. Otherwise it asks for one, reusing a pending request of the same call.
func CheckMcpApproval(server *Server, caller *McpCaller, tool string, arguments string, traceId string, clientIp string) (*McpApproval, bool, error) {
	if err := edition.Require(edition.Enterprise, edition.FeatureMcpApproval); err != nil {
		return nil, false, err
	}

	argumentsHash := getMcpArgumentsHash(arguments)
	now := util.GetCurrentTime()
	approvals := []*McpApproval{}
	err := ormer.Engine.Where("owner = ? and server = ? and tool = ? and user = ? and agent = ? and arguments_hash = ? and state in (?, ?) and expire_time > ?",
		server.Owner, server.Name, tool, caller.User, caller.GetAgentId(), argumentsHash, McpApprovalPending, McpApprovalApproved, now).
		Asc("created_time").Find(&approvals)
	if err != nil {
		return nil, false, err
	}

	for _, approval := range approvals {
		if approval.State != McpApprovalApproved {
			continue
		}
		affected, err := ormer.Engine.ID(core.PK{approval.Owner, approval.Name}).Where("state = ?", McpApprovalApproved).
			Cols("state", "used_time", "used_trace_id").Update(&McpApproval{State: McpApprovalUsed, UsedTime: now, UsedTraceId: traceId})
		if err != nil {
			return nil, false, err
		}
		if affected != 0 {
			approval.State = McpApprovalUsed
			return approval, true, nil
		}
	}
	for _, approval := range approvals {
		if approval.State == McpApprovalPending {
			return approval, false, nil
		}
	}

	approval := &McpApproval{
		Owner:         server.Owner,
		Name:          util.GenerateId(),
		CreatedTime:   now,
		Server:        server.Name,
		Tool:          tool,
		User:          caller.User,
		Agent:         caller.GetAgentId(),
		Application:   caller.Application,
		Arguments:     truncateString(arguments, GetMcpRecordBodyLimit()),
		ArgumentsHash: argumentsHash,
		TraceId:       traceId,
		ClientIp:      clientIp,
		State:         McpApprovalPending,
		ExpireTime:    time.Now().Add(mcpApprovalPendingHours * time.Hour).Format(time.RFC3339),
	}
	_, err = ormer.Engine.Insert(approval)
	if err != nil {
		return nil, false, err
	}
	return approval, false, nil
}

// DecideMcpApproval approves or rejects a pending call. Whoever asked for the call may not approve it.
func DecideMcpApproval(id string, approver *User, isApproved bool, comment string) (bool, error) {
	if err := edition.Require(edition.Enterprise, edition.FeatureMcpApproval); err != nil {
		return false, err
	}

	approval, err := GetMcpApproval(id)
	if err != nil {
		return false, err
	}
	if approval == nil {
		return false, fmt.Errorf("the approval: %s does not exist", id)
	}
	if approval.State != McpApprovalPending || approval.isExpired() {
		return false, fmt.Errorf("the approval: %s is no longer pending", id)
	}
	if approver.GetId() == approval.User {
		return false, fmt.Errorf("the user: %s cannot approve their own tool call", approver.GetId())
	}

	update := &McpApproval{
		State:       McpApprovalRejected,
		Approver:    approver.GetId(),
		ApproveTime: util.GetCurrentTime(),
		Comment:     comment,
		ExpireTime:  approval.ExpireTime,
	}
	if isApproved {
		update.State = McpApprovalApproved
		update.ExpireTime = time.Now().Add(mcpApprovalApprovedHours * time.Hour).Format(time.RFC3339)
	}
	affected, err := ormer.Engine.ID(core.PK{approval.Owner, approval.Name}).Where("state = ?", McpApprovalPending).
		Cols("state", "approver", "approve_time", "comment", "expire_time").Update(update)
	return affected != 0, err
}

// checkAgentSod keeps an agent from calling a tool of a separation-of-duty rule while it is allowed two or more tools
// of the rule, such as creating a payment and approving it
func checkAgentSod(server *Server, subject *mcpToolSubject, tool string) (string, error) {
	if subject.agent == "" || !edition.Has(edition.Enterprise) {
		return "", nil
	}
	policy := GetCachedEnterprisePolicy(server.Owner)
	if policy == nil {
		return "", nil
	}

	for _, rule := range policy.SodRules {
		if rule == nil || !util.InSlice(rule.Tools, server.Name+"/"+tool) {
			continue
		}

		allowed := []string{}
		for _, ruleTool := range rule.Tools {
			serverName, toolName, ok := strings.Cut(ruleTool, "/")
			if !ok {
				continue
			}
			permissions, err := getMcpToolPermissions(&Server{Owner: server.Owner, Name: serverName}, toolName)
			if err != nil {
				return "", err
			}
			for _, permission := range permissions {
				if permission.Effect != "Deny" && subject.isAgentHit(permission) {
					allowed = append(allowed, ruleTool)
					break
				}
			}
		}
		if len(allowed) >= 2 {
			return fmt.Sprintf("separation of duties: the agent: %s is allowed the tools of the rule: %s at once (%s)", subject.agent, rule.Name, strings.Join(allowed, ", ")), nil
		}
	}
	return "", nil
}
