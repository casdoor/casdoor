// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package mcptools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
)

const (
	actionAdd    = "Add"
	actionUpdate = "Update"
	actionDelete = "Delete"

	ChangeApplied    = "Applied"
	ChangeRolledBack = "Rolled back"
)

// McpChange is a change that an agent made through the tools of Casdoor's MCP server: who, which tool, the object
// and its state before and after, so that an admin sees what the agent did and can roll it back
type McpChange struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100) index" json:"createdTime"`

	User        string `xorm:"varchar(200) index" json:"user"`
	Application string `xorm:"varchar(100)" json:"application"`
	Agent       string `xorm:"varchar(200) index" json:"agent"`
	AuthMethod  string `xorm:"varchar(100)" json:"authMethod"`
	TraceId     string `xorm:"varchar(100) index" json:"traceId"`
	Approval    string `xorm:"varchar(200)" json:"approval"`
	Tool        string `xorm:"varchar(200)" json:"tool"`
	// the order of the change among those of the same tool call
	Sequence int `json:"sequence"`

	Kind     string `xorm:"varchar(100)" json:"kind"`
	ObjectId string `xorm:"varchar(300) index" json:"objectId"`
	Action   string `xorm:"varchar(100)" json:"action"`
	// the whole objects stay on the server, the API hands out only the masked diff
	Before string `xorm:"mediumtext" json:"-"`
	After  string `xorm:"mediumtext" json:"-"`
	Diff   string `xorm:"mediumtext" json:"diff"`

	State        string `xorm:"varchar(100) index" json:"state"`
	RollbackUser string `xorm:"varchar(200)" json:"rollbackUser"`
	RollbackTime string `xorm:"varchar(100)" json:"rollbackTime"`
}

func (change *McpChange) GetId() string {
	return util.GetId(change.Owner, change.Name)
}

func (change *McpChange) getObjects() (map[string]interface{}, map[string]interface{}) {
	var before, after map[string]interface{}
	if change.Before != "" {
		_ = json.Unmarshal([]byte(change.Before), &before)
	}
	if change.After != "" {
		_ = json.Unmarshal([]byte(change.After), &after)
	}
	return before, after
}

func marshalObject(obj map[string]interface{}) string {
	if obj == nil {
		return ""
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(data)
}

func addMcpChange(change *McpChange) error {
	_, err := object.GetOrmer().Engine.Insert(change)
	return err
}

type McpChangeFilter struct {
	Agent    string
	User     string
	Kind     string
	ObjectId string
	State    string
	TraceId  string
}

func getMcpChangeSession(owner string, filter *McpChangeFilter) *xorm.Session {
	session := object.GetOrmer().Engine.Where("owner = ?", owner)
	if filter.Agent != "" {
		session = session.And("agent = ?", filter.Agent)
	}
	if filter.User != "" {
		session = session.And("user = ?", filter.User)
	}
	if filter.Kind != "" {
		session = session.And("kind = ?", filter.Kind)
	}
	if filter.ObjectId != "" {
		session = session.And("object_id = ?", filter.ObjectId)
	}
	if filter.State != "" {
		session = session.And("state = ?", filter.State)
	}
	if filter.TraceId != "" {
		session = session.And("trace_id = ?", filter.TraceId)
	}
	return session
}

func GetMcpChangeCount(owner string, filter *McpChangeFilter) (int64, error) {
	if err := edition.Require(edition.Commercial, edition.FeatureMcpManagement); err != nil {
		return 0, err
	}
	return getMcpChangeSession(owner, filter).Count(&McpChange{})
}

func GetPaginationMcpChanges(owner string, filter *McpChangeFilter, offset int, limit int) ([]*McpChange, error) {
	if err := edition.Require(edition.Commercial, edition.FeatureMcpManagement); err != nil {
		return nil, err
	}
	changes := []*McpChange{}
	err := getMcpChangeSession(owner, filter).Desc("created_time", "sequence").Limit(limit, offset).Find(&changes)
	return changes, err
}

func GetMcpChange(id string) (*McpChange, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	change := McpChange{Owner: owner, Name: name}
	existed, err := object.GetOrmer().Engine.Get(&change)
	if err != nil || !existed {
		return nil, err
	}
	return &change, nil
}

// getRelatedIds are the IDs the object had in the change, two when it was renamed
func (change *McpChange) getRelatedIds() []string {
	ids := []string{change.ObjectId}
	before, after := change.getObjects()
	for _, obj := range []map[string]interface{}{before, after} {
		if obj != nil {
			if id := getObjectId(obj); !util.InSlice(ids, id) {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// RollbackMcpChange undoes a change of an agent for the admin: an added object is deleted, an updated one gets its
// fields back and a deleted one is added again. The newer changes of the object have to be rolled back first. The
// request is authorized for the admin as the same change in the console would be.
func RollbackMcpChange(ctx *context.Context, userId string, id string) error {
	if err := edition.Require(edition.Enterprise, edition.FeatureMcpRollback); err != nil {
		return err
	}

	change, err := GetMcpChange(id)
	if err != nil {
		return err
	}
	if change == nil {
		return fmt.Errorf("the change: %s does not exist", id)
	}
	if err = change.checkRollback(false); err != nil {
		return err
	}
	return change.rollback(ctx, userId)
}

// RollbackMcpCall undoes all the changes of one tool call, the last one first, after checking that none of them
// has a newer change from another call
func RollbackMcpCall(ctx *context.Context, userId string, owner string, traceId string) error {
	if err := edition.Require(edition.Enterprise, edition.FeatureMcpRollback); err != nil {
		return err
	}
	if traceId == "" {
		return fmt.Errorf("the trace ID is required")
	}

	changes := []*McpChange{}
	err := object.GetOrmer().Engine.Where("owner = ? and trace_id = ? and state = ?", owner, traceId, ChangeApplied).Desc("sequence").Find(&changes)
	if err != nil {
		return err
	}
	if len(changes) == 0 {
		return fmt.Errorf("the call: %s has no change to roll back", traceId)
	}
	for _, change := range changes {
		if err = change.checkRollback(true); err != nil {
			return err
		}
	}
	for _, change := range changes {
		if err = change.rollback(ctx, userId); err != nil {
			return err
		}
	}
	return nil
}

// checkRollback tells why the change can't be rolled back now: the newer changes of the object come first, except
// those of the same call when the whole call is rolled back
func (change *McpChange) checkRollback(isWholeCall bool) error {
	if change.State != ChangeApplied {
		return fmt.Errorf("the change: %s has been rolled back", change.GetId())
	}
	if kinds[change.Kind] == nil {
		return fmt.Errorf("the change: %s can't be rolled back", change.GetId())
	}

	newer := []*McpChange{}
	session := object.GetOrmer().Engine.Where("owner = ? and kind = ? and state = ?", change.Owner, change.Kind, ChangeApplied).In("object_id", change.getRelatedIds())
	if isWholeCall {
		session = session.And("created_time >= ? and trace_id <> ?", change.CreatedTime, change.TraceId)
	} else {
		session = session.And("(created_time > ? or (trace_id = ? and sequence > ?))", change.CreatedTime, change.TraceId, change.Sequence)
	}
	if err := session.Find(&newer); err != nil {
		return err
	}
	if len(newer) > 0 {
		return fmt.Errorf("roll back the newer change of %s first: %s by the tool %s (%s)", change.ObjectId, newer[0].Name, newer[0].Tool, newer[0].CreatedTime)
	}
	return nil
}

func (change *McpChange) rollback(ctx *context.Context, userId string) error {
	k := kinds[change.Kind]
	cl := &client{ctx: ctx, userId: userId}
	before, after := change.getObjects()
	var err error
	switch change.Action {
	case actionAdd:
		var current map[string]interface{}
		current, err = cl.mustGet(k, getObjectId(after))
		if err == nil {
			_, err = cl.save(k, actionDelete, "", current)
		}
	case actionUpdate:
		if before == nil || after == nil {
			return fmt.Errorf("the change: %s has no saved state to roll back to", change.GetId())
		}
		_, err = cl.save(k, actionUpdate, getObjectId(after), before)
	case actionDelete:
		if before == nil {
			return fmt.Errorf("the change: %s has no saved state to roll back to", change.GetId())
		}
		_, err = cl.save(k, actionAdd, "", withoutMaskedValues(before))
	default:
		return fmt.Errorf("the change: %s can't be rolled back", change.GetId())
	}
	if err != nil {
		return err
	}

	_, err = object.GetOrmer().Engine.ID(core.PK{change.Owner, change.Name}).Where("state = ?", ChangeApplied).
		Cols("state", "rollback_user", "rollback_time").Update(&McpChange{State: ChangeRolledBack, RollbackUser: userId, RollbackTime: util.GetCurrentTime()})
	return err
}

// withoutMaskedValues clears the values that the API masked as "***": adding them back would store "***" itself
func withoutMaskedValues(obj map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range obj {
		if text, ok := value.(string); ok && strings.Trim(text, "*") == "" && text != "" {
			result[key] = ""
			continue
		}
		result[key] = value
	}
	return result
}
