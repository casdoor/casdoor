// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package mcptools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/ee"
	"github.com/casdoor/casdoor/mcpself"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// BuiltinServer is the server name that the records and approvals of Casdoor's own MCP server are kept under
const BuiltinServer = "casdoor"

const traceIdHeader = "X-Casdoor-Trace-Id"

// toolDef is a tool of the commercial edition
type toolDef struct {
	Name        string
	Description string
	Scope       string
	// more scopes that the tool needs, e.g. one that changes both a provider and an application
	MoreScopes []string
	Write      bool
	Delete     bool
	Properties map[string]interface{}
	Required   []string
	// the organization that a call acts on, for its approval and records
	Organization func(t *toolCall) (string, error)
	Run          func(t *toolCall) (string, error)
}

// builtinWrite tells what a write tool of the MCP server itself changes: the ID of the object before the call and the
// object that the call writes
type builtinWrite struct {
	Kind   *kind
	Action string
	Parse  func(args map[string]json.RawMessage) (string, map[string]interface{})
}

var (
	toolDefs      []*toolDef
	builtinWrites map[string]*builtinWrite
)

func init() {
	object.RegisterTables(new(McpChange))
	toolDefs = append(getKindTools(), getSetupTools()...)
	builtinWrites = getBuiltinWrites()
	mcpself.McpExtension = &extension{}
}

func findToolDef(name string) *toolDef {
	for _, def := range toolDefs {
		if def.Name == name {
			return def
		}
	}
	return nil
}

type extension struct{}

func (e *extension) GetTools() []*mcpself.ExtraTool {
	tools := []*mcpself.ExtraTool{}
	for _, def := range toolDefs {
		tools = append(tools, &mcpself.ExtraTool{Tool: def.toMcpTool(), Scope: def.Scope})
	}
	return tools
}

var dryRunProperty = map[string]interface{}{
	"type":        "boolean",
	"description": "Only tell what the call would change, without changing anything",
}

// ListTools adds the tools of the commercial edition, and the dry run of the write tools, when the license allows
func (e *extension) ListTools(tools []mcpself.McpTool) []mcpself.McpTool {
	if !edition.Has(edition.Commercial) {
		return tools
	}

	result := []mcpself.McpTool{}
	for _, tool := range tools {
		if write, ok := builtinWrites[tool.Name]; ok {
			tool = withDryRun(tool)
			if write.Action == actionUpdate {
				tool.Description += ". Only the fields given change, the others keep their values."
			}
		}
		result = append(result, tool)
	}
	for _, def := range toolDefs {
		result = append(result, def.toMcpTool())
	}
	return result
}

func withDryRun(tool mcpself.McpTool) mcpself.McpTool {
	schema := map[string]interface{}{}
	for key, value := range tool.InputSchema {
		schema[key] = value
	}
	properties := map[string]interface{}{}
	if old, ok := schema["properties"].(map[string]interface{}); ok {
		for key, value := range old {
			properties[key] = value
		}
	}
	properties["dryRun"] = dryRunProperty
	schema["properties"] = properties
	tool.InputSchema = schema
	return tool
}

func (def *toolDef) toMcpTool() mcpself.McpTool {
	properties := map[string]interface{}{}
	for key, value := range def.Properties {
		properties[key] = value
	}
	if def.Write {
		properties["dryRun"] = dryRunProperty
	}
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(def.Required) > 0 {
		schema["required"] = def.Required
	}
	return mcpself.McpTool{Name: def.Name, Description: def.Description, InputSchema: schema}
}

// toolCall is a call of a tool of Casdoor's MCP server
type toolCall struct {
	c         *mcpself.McpController
	id        interface{}
	name      string
	arguments json.RawMessage
	args      map[string]json.RawMessage
	def       *toolDef
	builtin   *builtinWrite

	client       *client
	caller       *ee.McpCaller
	traceId      string
	startTime    time.Time
	dryRun       bool
	organization string
	approval     string

	changes []*McpChange
	planned []string
}

func (t *toolCall) isWrite() bool {
	return t.builtin != nil || (t.def != nil && t.def.Write)
}

func (t *toolCall) isDelete() bool {
	return (t.builtin != nil && t.builtin.Action == actionDelete) || (t.def != nil && t.def.Delete) || t.name == "undo_migration"
}

// isMigrationChange tells whether a migration tool of the server adds or deletes the objects of an organization
func isMigrationChange(name string) bool {
	return name == "apply_migration" || name == "undo_migration"
}

// CallTool runs a tool of the commercial edition, and wraps the tools of the server itself with the dry run, the
// approval of the enterprise edition and the records of who changed what
func (e *extension) CallTool(c *mcpself.McpController, id interface{}, name string, arguments json.RawMessage, next func()) {
	def := findToolDef(name)
	if def == nil && !edition.Has(edition.Commercial) {
		// an agent that asks for a dry run must not get the change made instead
		if builtinWrites[name] != nil && strings.Contains(string(arguments), `"dryRun"`) {
			c.SendToolErrorResult(id, edition.Require(edition.Commercial, edition.FeatureMcpManagement).Error())
			return
		}
		next()
		return
	}

	t := &toolCall{c: c, id: id, name: name, arguments: arguments, def: def, builtin: builtinWrites[name], startTime: time.Now()}
	t.args = map[string]json.RawMessage{}
	if len(arguments) > 0 && json.Unmarshal(arguments, &t.args) != nil {
		c.SendToolErrorResult(id, "the arguments must be a JSON object")
		return
	}
	t.traceId, _ = ee.GetMcpTraceContext(c.Ctx.Request.Header.Get("traceparent"), c.Ctx.Request.Header.Get(traceIdHeader))
	c.Ctx.Output.Header(traceIdHeader, t.traceId)

	userId := c.GetSessionUsername()
	t.client = &client{ctx: c.Ctx, userId: userId}
	caller, err := getCaller(c, userId)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}
	t.caller = caller

	if def != nil {
		if err = edition.Require(edition.Commercial, edition.FeatureMcpManagement); err != nil {
			t.finish("", err)
			return
		}
		for _, scope := range def.MoreScopes {
			if !t.hasScope(scope) {
				t.finish("", fmt.Errorf("the tool %s also needs the scope %s, sign in again and grant it", name, scope))
				return
			}
		}
	}

	t.dryRun = t.isWrite() && t.getBool("dryRun")
	t.organization, err = t.getOrganization()
	if err != nil {
		t.finish("", err)
		return
	}

	if (t.isWrite() && !t.dryRun || isMigrationChange(name)) && t.holdForApproval() {
		return
	}

	switch {
	case def != nil:
		text, err := def.Run(t)
		t.finish(text, err)
	case t.builtin != nil:
		t.runBuiltinWrite(next)
	default:
		next()
		text, isError := getResult(c)
		t.record(resultOf(isError), text)
	}
}

func (t *toolCall) hasScope(scope string) bool {
	grantedScopes, isScoped := t.c.GetGrantedScopes()
	if !isScoped {
		return true
	}
	return mcpself.GetToolsForScopes(grantedScopes, []*object.ScopeItem{{Name: scope, Tools: []string{scope}}})[scope]
}

// getOrganization is the organization that the call acts on, the user's own one when the call doesn't tell
func (t *toolCall) getOrganization() (string, error) {
	organization := ""
	var err error
	switch {
	case t.def != nil && t.def.Organization != nil:
		organization, err = t.def.Organization(t)
	case t.builtin != nil:
		organization, err = t.getBuiltinOrganization()
	default:
		for _, key := range []string{"owner", "organization"} {
			if value := t.getString(key); value != "" && value != "admin" {
				organization = value
			}
		}
	}
	if err != nil {
		return "", err
	}
	if organization == "" {
		organization, _ = util.GetOwnerAndNameFromIdNoCheck(t.caller.User)
		if organization == "app" {
			organization = "built-in"
		}
	}
	return organization, nil
}

// holdForApproval keeps a change of an agent waiting when the organization's enterprise policy asks an admin to
// approve it first, a person working in the console isn't held
func (t *toolCall) holdForApproval() bool {
	if !edition.Has(edition.Enterprise) || t.caller.AuthMethod == "Session" {
		return false
	}
	policy := ee.GetCachedEnterprisePolicy(t.organization)
	if policy == nil || policy.McpApproval == "" || (policy.McpApproval == "Delete" && !t.isDelete()) {
		return false
	}

	server := &object.Server{Owner: t.organization, Name: BuiltinServer}
	approval, isApproved, err := ee.CheckMcpApproval(server, t.caller, t.name, maskArguments(string(t.arguments)), t.traceId, util.GetClientIpFromRequest(t.c.Ctx.Request))
	if err != nil {
		t.finish("", err)
		return true
	}
	if isApproved {
		t.approval = approval.GetId()
		return false
	}

	t.approval = approval.GetId()
	approvalUrl := ee.GetMcpApprovalUrl(t.c.Ctx.Request.Host, approval)
	text := fmt.Sprintf("The change made by the tool \"%s\" needs approval before it runs. An administrator of the organization %s can approve it at %s, then call the tool again with the same arguments within an hour of the approval. Approval: %s, trace ID: %s",
		t.name, t.organization, approvalUrl, approval.GetId(), t.traceId)
	t.record("Pending approval", approvalUrl)
	t.c.SendToolErrorResult(t.id, text)
	return true
}

// finish answers the call with the text of the tool, or what a dry run would change, and records it
func (t *toolCall) finish(text string, err error) {
	if err != nil {
		t.record("Error", err.Error())
		t.c.SendToolErrorResult(t.id, err.Error())
		return
	}

	if t.dryRun {
		lines := []string{"Dry run, nothing was changed. The call would:"}
		if len(t.planned) == 0 {
			lines = append(lines, "- change nothing")
		}
		lines = append(lines, t.planned...)
		if text != "" {
			lines = append(lines, "", text)
		}
		text = strings.Join(lines, "\n")
		t.record("Dry run", text)
		t.c.SendToolResult(t.id, text)
		return
	}

	if len(t.changes) > 0 {
		lines := []string{text, "", fmt.Sprintf("Changes (trace ID %s):", t.traceId)}
		for _, change := range t.changes {
			lines = append(lines, fmt.Sprintf("- %s %s %s", strings.ToLower(change.Action), change.Kind, change.ObjectId))
		}
		text = strings.TrimSpace(strings.Join(lines, "\n"))
	}
	t.record("Success", text)
	t.c.SendToolResult(t.id, text)
}

// apply adds, updates or deletes an object for the call, or notes what it would do in a dry run. id is the object's
// ID before an update.
func (t *toolCall) apply(k *kind, action string, id string, before map[string]interface{}, after map[string]interface{}) error {
	objectId := ""
	var diff []string
	switch action {
	case actionAdd:
		objectId = getObjectId(after)
		if existing, err := t.client.get(k, objectId); err == nil && existing != nil {
			return fmt.Errorf("the %s %s already exists", k.Name, objectId)
		}
		diff = diffObjects(map[string]interface{}{}, after, k.Secrets)
	case actionUpdate:
		objectId = getObjectId(after)
		diff = diffObjects(before, after, k.Secrets)
	case actionDelete:
		objectId = getObjectId(before)
	}

	if t.dryRun {
		t.planned = append(t.planned, fmt.Sprintf("- %s the %s %s", strings.ToLower(action), k.Name, objectId))
		for _, line := range diff {
			t.planned = append(t.planned, "    "+line)
		}
		return nil
	}

	body := after
	if action == actionDelete {
		body = before
	}
	affected, err := t.client.save(k, action, id, body)
	if err != nil {
		return err
	}
	if !affected && action != actionUpdate {
		return fmt.Errorf("the %s: %s was not %s", k.Name, objectId, map[string]string{actionAdd: "added, it may already exist", actionDelete: "deleted, it may not exist"}[action])
	}
	if action != actionDelete {
		if saved, err := t.client.get(k, objectId); err == nil && saved != nil {
			after = saved
		}
	}
	return t.addChange(k, action, objectId, before, after, diff)
}

func (t *toolCall) addChange(k *kind, action string, objectId string, before map[string]interface{}, after map[string]interface{}, diff []string) error {
	if action == actionDelete {
		after = nil
	}
	change := &McpChange{
		Owner:       k.organizationOf(before, after, t.organization),
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		User:        t.caller.User,
		Application: t.caller.Application,
		Agent:       t.caller.GetAgentId(),
		AuthMethod:  t.caller.AuthMethod,
		TraceId:     t.traceId,
		Approval:    t.approval,
		Tool:        t.name,
		Sequence:    len(t.changes) + 1,
		Kind:        k.Name,
		ObjectId:    objectId,
		Action:      action,
		Before:      marshalObject(before),
		After:       marshalObject(after),
		Diff:        strings.Join(diff, "\n"),
		State:       ChangeApplied,
	}
	t.changes = append(t.changes, change)
	return addMcpChange(change)
}

func (k *kind) organizationOf(before map[string]interface{}, after map[string]interface{}, fallback string) string {
	for _, obj := range []map[string]interface{}{after, before} {
		if obj != nil {
			if organization := k.OrganizationOf(obj); organization != "" {
				return organization
			}
		}
	}
	return fallback
}

// runBuiltinWrite runs a write tool of the server itself, after telling what it changes in a dry run, and records
// the change
func (t *toolCall) runBuiltinWrite(next func()) {
	k := t.builtin.Kind
	id, obj := t.builtin.Parse(t.args)

	var before map[string]interface{}
	if t.builtin.Action != actionAdd && id != "" {
		var err error
		before, err = t.client.get(k, id)
		if err != nil && t.dryRun {
			t.finish("", err)
			return
		}
	}

	// an update keeps the fields that the call leaves out, instead of clearing them as the tool itself does
	if t.builtin.Action == actionUpdate {
		if before == nil {
			t.finish("", fmt.Errorf("the %s: %s does not exist", k.Name, id))
			return
		}
		after := map[string]interface{}{}
		for key, value := range before {
			after[key] = value
		}
		for key, value := range obj {
			after[key] = value
		}
		err := t.apply(k, actionUpdate, id, before, after)
		if err == nil && !t.dryRun {
			t.finish(mcpself.FormatOperationResult("update", k.Name, true), nil)
			return
		}
		t.finish("", err)
		return
	}

	if t.dryRun {
		if t.builtin.Action != actionAdd && before == nil {
			t.finish("", fmt.Errorf("the %s: %s does not exist", k.Name, id))
			return
		}
		after := obj
		if t.builtin.Action == actionDelete {
			after = nil
		}
		t.finish("", t.apply(k, t.builtin.Action, id, before, after))
		return
	}

	next()
	text, isError := getResult(t.c)
	if !isError && strings.HasPrefix(text, "Successfully") {
		objectId := id
		var after map[string]interface{}
		if t.builtin.Action != actionDelete {
			objectId = getObjectId(obj)
			after, _ = t.client.get(k, objectId)
		}
		var diff []string
		if t.builtin.Action == actionAdd {
			diff = diffObjects(map[string]interface{}{}, after, k.Secrets)
		} else if t.builtin.Action == actionUpdate {
			diff = diffObjects(before, after, k.Secrets)
		}
		_ = t.addChange(k, t.builtin.Action, objectId, before, after, diff)
	}
	t.record(resultOf(isError), text)
}

func (t *toolCall) getBuiltinOrganization() (string, error) {
	id, obj := t.builtin.Parse(t.args)
	if organization := t.builtin.Kind.OrganizationOf(obj); organization != "" && organization != "admin" {
		return organization, nil
	}
	if id == "" || t.builtin.Action == actionAdd {
		return "", nil
	}
	before, err := t.client.get(t.builtin.Kind, id)
	if err != nil || before == nil {
		return "", nil
	}
	return t.builtin.Kind.OrganizationOf(before), nil
}

func getBuiltinWrites() map[string]*builtinWrite {
	objectOf := func(args map[string]json.RawMessage, key string) map[string]interface{} {
		obj := map[string]interface{}{}
		_ = json.Unmarshal(args[key], &obj)
		return obj
	}
	idOf := func(args map[string]json.RawMessage) string {
		id := ""
		_ = json.Unmarshal(args["id"], &id)
		return id
	}

	writes := map[string]*builtinWrite{}
	for _, name := range []string{"application", "user"} {
		k := kinds[name]
		key := name
		writes["add_"+name] = &builtinWrite{Kind: k, Action: actionAdd, Parse: func(args map[string]json.RawMessage) (string, map[string]interface{}) {
			obj := objectOf(args, key)
			if key == "application" && getString(obj, "owner") == "" {
				obj["owner"] = "admin"
			}
			return "", obj
		}}
		writes["update_"+name] = &builtinWrite{Kind: k, Action: actionUpdate, Parse: func(args map[string]json.RawMessage) (string, map[string]interface{}) {
			// "fields" is how the update tools of the other kinds take the changes
			if _, ok := args[key]; !ok {
				return idOf(args), objectOf(args, "fields")
			}
			return idOf(args), objectOf(args, key)
		}}
		writes["delete_"+name] = &builtinWrite{Kind: k, Action: actionDelete, Parse: func(args map[string]json.RawMessage) (string, map[string]interface{}) {
			obj := objectOf(args, key)
			if key == "application" && getString(obj, "owner") == "" {
				obj["owner"] = "admin"
			}
			return getObjectId(obj), obj
		}}
	}
	return writes
}

func getResult(c *mcpself.McpController) (string, bool) {
	resp, ok := c.Data["json"].(mcpself.McpResponse)
	if !ok {
		return "", true
	}
	if resp.Error != nil {
		return resp.Error.Message, true
	}
	result, ok := resp.Result.(mcpself.McpCallToolResult)
	if !ok {
		return "", false
	}
	texts := []string{}
	for _, content := range result.Content {
		texts = append(texts, content.Text)
	}
	return strings.Join(texts, "\n"), result.IsError
}

func resultOf(isError bool) string {
	if isError {
		return "Error"
	}
	return "Success"
}

var secretArgumentRegex = regexp.MustCompile(`"(clientSecret|clientSecret2|password|privateKey|accessSecret|masterPassword|defaultPassword)"\s*:\s*"[^"]*"`)

func maskArguments(arguments string) string {
	return secretArgumentRegex.ReplaceAllString(arguments, `"$1":"***"`)
}

// record keeps who called which tool: in the MCP tool call records (the changes always, the reads when the
// organization's security policy asks so) and in the request's audit record
func (t *toolCall) record(result string, response string) {
	arguments := maskArguments(string(t.arguments))
	summary := map[string]interface{}{
		"server":      BuiltinServer,
		"tool":        t.name,
		"arguments":   json.RawMessage(nonEmptyJson(arguments)),
		"user":        t.caller.User,
		"application": t.caller.Application,
		"agent":       t.caller.GetAgentId(),
		"authMethod":  t.caller.AuthMethod,
		"traceId":     t.traceId,
		"result":      result,
		"approval":    t.approval,
	}
	if data, err := json.Marshal(summary); err == nil {
		status := "ok"
		if result == "Error" {
			status = "error"
		}
		t.c.Ctx.Input.SetParam("recordMcpTraceId", t.traceId)
		t.c.Ctx.Input.SetParam("recordMcpObject", string(data))
		t.c.Ctx.Input.SetParam("recordMcpResponse", fmt.Sprintf("{status:\"%s\", msg:\"%s\"}", status, strings.ReplaceAll(truncate(result+" "+response, 500), "\"", "'")))
	}

	if !edition.Has(edition.Commercial) || (!t.isWrite() && !object.IsMcpToolCallRecorded(t.organization)) {
		return
	}
	limit := object.GetMcpRecordBodyLimit()
	object.AddMcpRecord(&object.McpRecord{
		Owner:       t.organization,
		Server:      BuiltinServer,
		UserName:    t.caller.User,
		ClientIp:    util.GetClientIpFromRequest(t.c.Ctx.Request),
		UserAgent:   t.c.Ctx.Request.UserAgent(),
		Tool:        t.name,
		Arguments:   object.TruncateString(arguments, limit),
		Result:      result,
		StatusCode:  http.StatusOK,
		Response:    object.TruncateString(response, limit),
		DurationMs:  time.Since(t.startTime).Milliseconds(),
		Application: t.caller.Application,
		Agent:       t.caller.GetAgentId(),
		Actor:       t.caller.GetActorJson(),
		AuthMethod:  t.caller.AuthMethod,
		TokenId:     t.caller.TokenId,
		TraceId:     t.traceId,
		Approval:    t.approval,
	})
}

func nonEmptyJson(text string) []byte {
	if text == "" || !json.Valid([]byte(text)) {
		return []byte("null")
	}
	return []byte(text)
}

// getCaller tells who calls the tool from the credential that the AutoSigninFilter has verified: the user and, for a
// token or key of an OAuth client, the client and the agent it is
func getCaller(c *mcpself.McpController, userId string) (*ee.McpCaller, error) {
	accessToken := c.GetTokenFromRequest()
	accessKey := c.Ctx.Request.Header.Get("X-Casdoor-Access-Key")
	if key, _, ok := strings.Cut(accessToken, ":"); ok {
		accessKey, accessToken = key, ""
	}

	if accessToken != "" {
		token, err := object.GetTokenByAccessToken(accessToken)
		if err != nil {
			return nil, err
		}
		if token != nil {
			return ee.NewMcpCallerByToken(userId, token)
		}
	}
	if accessKey != "" {
		key, err := object.GetKeyByAccessKey(accessKey)
		if err != nil {
			return nil, err
		}
		if key != nil {
			return ee.NewMcpCallerByKey(userId, key)
		}
	}
	return &ee.McpCaller{User: userId, AuthMethod: "Session"}, nil
}

func (t *toolCall) getString(key string) string {
	value := ""
	_ = json.Unmarshal(t.args[key], &value)
	return strings.TrimSpace(value)
}

func (t *toolCall) getBool(key string) bool {
	value := false
	_ = json.Unmarshal(t.args[key], &value)
	return value
}

func (t *toolCall) getInt(key string, defaultValue int, max int) int {
	value := 0
	if json.Unmarshal(t.args[key], &value) != nil || value <= 0 {
		return defaultValue
	}
	if value > max {
		return max
	}
	return value
}

func (t *toolCall) getStrings(key string) []string {
	values := []string{}
	if json.Unmarshal(t.args[key], &values) != nil {
		value := t.getString(key)
		if value != "" {
			values = []string{value}
		}
	}
	return values
}

func (t *toolCall) getObject(key string) (map[string]interface{}, error) {
	obj := map[string]interface{}{}
	if len(t.args[key]) == 0 {
		return nil, fmt.Errorf("the argument %s is required", key)
	}
	if err := json.Unmarshal(t.args[key], &obj); err != nil {
		return nil, fmt.Errorf("the argument %s must be a JSON object", key)
	}
	return obj, nil
}

func (t *toolCall) requireString(key string) (string, error) {
	value := t.getString(key)
	if value == "" {
		return "", fmt.Errorf("the argument %s is required", key)
	}
	return value, nil
}

func toJson(value interface{}) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}
