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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
)

const PermissionResourceTypeMcpTool = "MCP tool"

// PermissionCondition limits an "MCP tool" permission to the calls whose argument passes it, e.g. the "factory" argument
// must be one of the user's groups. The value is a comma-separated list that may refer to the calling user, such as
// ${user.name}, ${user.owner}, ${user.groups}, ${user.groupNames}, ${user.affiliation} and ${user.properties.<key>}.
type PermissionCondition struct {
	Argument string `json:"argument"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// McpCaller is who calls a tool through the MCP gateway: the user, the OAuth client and the agent acting for the user
type McpCaller struct {
	User        string
	Application string
	ClientId    string
	Agent       *Agent
	Actor       *ActClaim
	AuthMethod  string
	TokenId     string
}

func (caller *McpCaller) IsUser() bool {
	return caller.User != "" && !IsAppUser(caller.User)
}

func (caller *McpCaller) GetAgentId() string {
	if caller.Agent == nil {
		return ""
	}
	return caller.Agent.GetId()
}

func (caller *McpCaller) GetActorJson() string {
	return getActorJson(caller.Actor)
}

func (caller *McpCaller) setApplication(application *Application, actor *ActClaim) error {
	caller.Application = application.Name
	caller.ClientId = application.ClientId
	caller.Actor = actor

	var err error
	if actor != nil && actor.Agent != "" {
		caller.Agent, err = GetAgent(actor.Agent)
	} else {
		caller.Agent, err = GetAgentByApplication(application.Name)
	}
	if err != nil {
		return err
	}

	// a user who signed in through the agent's client lets the agent act for them
	if caller.Actor == nil && caller.Agent != nil && caller.IsUser() {
		caller.Actor = &ActClaim{Sub: GetAppUserId(application), ClientId: application.ClientId, Agent: caller.Agent.GetId()}
	}
	return nil
}

func NewMcpCallerByToken(userId string, token *Token) (*McpCaller, error) {
	caller := &McpCaller{User: userId, AuthMethod: "Access token", TokenId: util.GetId(token.Owner, token.Name)}
	application, err := getApplication(token.Owner, token.Application)
	if err != nil {
		return nil, err
	}
	if application != nil {
		err = caller.setApplication(application, token.GetActor())
		if err != nil {
			return nil, err
		}
	}
	return caller, nil
}

func NewMcpCallerByKey(userId string, key *Key) (*McpCaller, error) {
	caller := &McpCaller{User: userId, AuthMethod: "Access key", TokenId: key.GetId()}
	if key.Application == "" {
		return caller, nil
	}
	application, err := getApplication("admin", key.Application)
	if err != nil {
		return nil, err
	}
	if application != nil {
		err = caller.setApplication(application, nil)
		if err != nil {
			return nil, err
		}
	}
	return caller, nil
}

var (
	traceparentRegex   = regexp.MustCompile(`^([0-9a-f]{2})-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$`)
	customTraceIdRegex = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,64}$`)
)

func getRandomHex(n int) string {
	data := make([]byte, n)
	_, _ = rand.Read(data)
	return hex.EncodeToString(data)
}

// GetMcpTraceContext returns the trace ID of a tool call and the W3C traceparent for the upstream MCP server: the trace
// of the caller's traceparent when it has one, then the caller's own X-Casdoor-Trace-Id, otherwise a new trace
func GetMcpTraceContext(traceparent string, customTraceId string) (string, string) {
	spanId := getRandomHex(8)
	match := traceparentRegex.FindStringSubmatch(strings.ToLower(strings.TrimSpace(traceparent)))
	if match != nil && match[1] != "ff" && match[2] != strings.Repeat("0", 32) {
		return match[2], fmt.Sprintf("00-%s-%s-%s", match[2], spanId, match[4])
	}

	traceId := getRandomHex(16)
	upstreamTraceparent := fmt.Sprintf("00-%s-%s-01", traceId, spanId)
	if customTraceIdRegex.MatchString(customTraceId) {
		return customTraceId, upstreamTraceparent
	}
	return traceId, upstreamTraceparent
}

// GetMcpIdentityToken signs a short-lived JWT for the upstream MCP server with the certificate of the server's
// application: who the user is, their groups and roles, the agent acting for them and the trace of the call. The MCP
// server verifies it with the JWKS of Casdoor and can pass it on to the business systems behind it.
func GetMcpIdentityToken(server *Server, caller *McpCaller, traceId string, tool string, host string) (string, string, error) {
	application, err := getApplication("admin", server.Application)
	if err != nil {
		return "", "", err
	}
	if application == nil {
		return "", "", fmt.Errorf("the application: %s of the MCP server: %s does not exist", server.Application, server.GetId())
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return "", "", err
	}
	if cert == nil {
		return "", "", fmt.Errorf("the cert of the application: %s does not exist", application.Name)
	}

	issuer, err := GetOidcIssuerByApplication(host, application)
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	jti := util.GetId(server.Owner, util.GenerateId())
	claims := jwt.MapClaims{
		"iss":        issuer,
		"sub":        caller.User,
		"aud":        []string{server.Url},
		"iat":        jwt.NewNumericDate(now),
		"nbf":        jwt.NewNumericDate(now),
		"exp":        jwt.NewNumericDate(now.Add(5 * time.Minute)),
		"jti":        jti,
		"tokenType":  "mcp-identity",
		"mcp_server": server.GetId(),
		"trace_id":   traceId,
	}
	if tool != "" {
		claims["tool"] = tool
	}
	if caller.ClientId != "" {
		claims["azp"] = caller.ClientId
	}
	if caller.Agent != nil {
		claims["agent"] = caller.Agent.GetId()
	}
	if caller.Actor != nil {
		claims["act"] = caller.Actor
	}

	if caller.IsUser() {
		user, err := GetUser(caller.User)
		if err != nil {
			return "", "", err
		}
		if user != nil {
			claims["owner"] = user.Owner
			claims["name"] = user.Name
			claims["preferred_username"] = user.Name
			claims["displayName"] = user.DisplayName
			claims["email"] = user.Email
			claims["groups"] = getUserGroupsWithAncestors(user)
			roles, err := getRolesByUser(user.GetId())
			if err != nil {
				return "", "", err
			}
			roleIds := []string{}
			for _, role := range roles {
				roleIds = append(roleIds, role.GetId())
			}
			claims["roles"] = roleIds
		}
	}

	method := getJwtSigningMethod(application)
	key, err := getJwtSigningKey(method, cert)
	if err != nil {
		return "", "", err
	}
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = cert.GetKeyId()
	tokenString, err := signJwtToken(token, key)
	return tokenString, jti, err
}

// getUserGroupsWithAncestors returns the user's groups and the groups above them, so that a permission of a department
// covers the teams in it
func getUserGroupsWithAncestors(user *User) []string {
	res := []string{}
	if len(user.Groups) == 0 {
		return res
	}

	parents := map[string]string{}
	groups, err := GetGroups(user.Owner)
	if err == nil {
		for _, group := range groups {
			parents[group.GetId()] = util.GetId(group.Owner, group.ParentId)
		}
	}

	seen := map[string]bool{}
	for _, groupId := range user.Groups {
		for groupId != "" && !seen[groupId] {
			seen[groupId] = true
			res = append(res, groupId)
			parent, ok := parents[groupId]
			if !ok {
				break
			}
			groupId = parent
		}
	}
	return res
}

func isMcpToolResourceHit(permission *Permission, server *Server, tool string) bool {
	for _, resource := range permission.Resources {
		if resource == "*" || resource == server.Name+"/*" || resource == server.Name+"/"+tool {
			return true
		}
	}
	return false
}

type mcpToolSubject struct {
	user   *User
	groups []string
	roles  []string
	agent  string
}

func (subject *mcpToolSubject) isUserHit(permission *Permission) bool {
	if permission.isUserHit(subject.user.GetId()) {
		return true
	}
	for _, group := range permission.Groups {
		if group == "*" || group == util.GetId(permission.Owner, "*") {
			if len(subject.groups) > 0 {
				return true
			}
		} else if util.InSlice(subject.groups, group) {
			return true
		}
	}
	for _, role := range permission.Roles {
		if role == "*" || util.InSlice(subject.roles, role) {
			return true
		}
	}
	return false
}

func (subject *mcpToolSubject) isAgentHit(permission *Permission) bool {
	for _, agent := range permission.Agents {
		if agent == "*" || agent == subject.agent || agent == util.GetId(permission.Owner, "*") {
			return true
		}
	}
	return false
}

// decideMcpTool allows a call when an "Allow" permission of the subject holds for it and no "Deny" one does
func decideMcpTool(permissions []*Permission, isHit func(*Permission) bool, user *User, arguments map[string]interface{}) bool {
	isAllowed := false
	for _, permission := range permissions {
		if !isHit(permission) || !isConditionsHit(permission.Conditions, user, arguments) {
			continue
		}
		if permission.Effect == "Deny" {
			return false
		}
		isAllowed = true
	}
	return isAllowed
}

func getMcpToolPermissions(server *Server, tool string) ([]*Permission, error) {
	permissions, err := GetPermissions(server.Owner)
	if err != nil {
		return nil, err
	}

	res := []*Permission{}
	for _, permission := range permissions {
		if permission.IsEnabled && permission.State == "Approved" && permission.ResourceType == PermissionResourceTypeMcpTool && isMcpToolResourceHit(permission, server, tool) {
			res = append(res, permission)
		}
	}
	return res, nil
}

// CheckMcpToolPermission returns why the caller may not call the tool, or "" when it may. With tool permissions on,
// the user needs an "MCP tool" permission for it, and so does the agent acting for the user: what the agent can do is
// at most what the user can do and what the agent is allowed to do.
func CheckMcpToolPermission(server *Server, caller *McpCaller, tool string, arguments map[string]interface{}) (string, error) {
	if caller.Agent != nil && caller.Agent.IsForbidden {
		return fmt.Sprintf("the agent: %s is disabled", caller.Agent.GetId()), nil
	}
	if !server.EnableToolPermissions {
		return "", nil
	}
	if !caller.IsUser() && caller.Agent == nil {
		return "only users and agents can call the tools of this MCP server", nil
	}

	permissions, err := getMcpToolPermissions(server, tool)
	if err != nil {
		return "", err
	}

	subject := &mcpToolSubject{agent: caller.GetAgentId()}
	if caller.IsUser() {
		subject.user, err = GetUser(caller.User)
		if err != nil {
			return "", err
		}
		if subject.user == nil {
			return fmt.Sprintf("the user: %s does not exist", caller.User), nil
		}
		subject.groups = getUserGroupsWithAncestors(subject.user)
		roles, err := getRolesByUser(caller.User)
		if err != nil {
			return "", err
		}
		for _, role := range roles {
			subject.roles = append(subject.roles, role.GetId())
		}

		if !decideMcpTool(permissions, subject.isUserHit, subject.user, arguments) {
			return fmt.Sprintf("the user: %s has no permission to call the tool: %s", caller.User, tool), nil
		}
	}

	if caller.Agent != nil {
		if !decideMcpTool(permissions, subject.isAgentHit, subject.user, arguments) {
			return fmt.Sprintf("the agent: %s has no permission to call the tool: %s", caller.Agent.GetId(), tool), nil
		}

		reason, err := checkAgentSod(server, subject, tool)
		if err != nil || reason != "" {
			return reason, err
		}
	}
	return "", nil
}

func formatArgumentValue(value interface{}) []string {
	switch v := value.(type) {
	case nil:
		return []string{}
	case string:
		return []string{v}
	case float64:
		return []string{strconv.FormatFloat(v, 'f', -1, 64)}
	case bool:
		return []string{strconv.FormatBool(v)}
	case []interface{}:
		res := []string{}
		for _, item := range v {
			res = append(res, formatArgumentValue(item)...)
		}
		return res
	default:
		data, _ := json.Marshal(v)
		return []string{string(data)}
	}
}

// getArgumentValues reads an argument by its dotted path, e.g. "filter.factory", an array gives all its items
func getArgumentValues(arguments map[string]interface{}, path string) []string {
	var value interface{} = arguments
	for _, key := range strings.Split(path, ".") {
		m, ok := value.(map[string]interface{})
		if !ok {
			return []string{}
		}
		value = m[key]
	}
	return formatArgumentValue(value)
}

var userPlaceholderRegex = regexp.MustCompile(`\$\{user\.([A-Za-z0-9_.]+)\}`)

func getUserPlaceholderValues(user *User, field string) []string {
	if user == nil {
		return []string{}
	}
	switch field {
	case "name":
		return []string{user.Name}
	case "owner":
		return []string{user.Owner}
	case "id":
		return []string{user.GetId()}
	case "email":
		return []string{user.Email}
	case "phone":
		return []string{user.Phone}
	case "affiliation":
		return []string{user.Affiliation}
	case "tag":
		return []string{user.Tag}
	case "displayName":
		return []string{user.DisplayName}
	case "groups":
		return getUserGroupsWithAncestors(user)
	case "groupNames":
		res := []string{}
		for _, group := range getUserGroupsWithAncestors(user) {
			_, name := util.GetOwnerAndNameFromIdNoCheck(group)
			res = append(res, name)
		}
		return res
	}
	if strings.HasPrefix(field, "properties.") && user.Properties != nil {
		if value, ok := user.Properties[strings.TrimPrefix(field, "properties.")]; ok {
			return []string{value}
		}
	}
	return []string{}
}

// expandConditionValue turns the value of a condition into the strings to compare with, one for each item of the
// comma-separated list, and one for each value of a placeholder that has several, such as ${user.groups}
func expandConditionValue(value string, user *User) []string {
	res := []string{}
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		match := userPlaceholderRegex.FindStringSubmatchIndex(item)
		if match == nil {
			if item != "" {
				res = append(res, item)
			}
			continue
		}

		for _, v := range getUserPlaceholderValues(user, item[match[2]:match[3]]) {
			if v != "" {
				res = append(res, item[:match[0]]+v+item[match[1]:])
			}
		}
	}
	return res
}

func isConditionHit(condition *PermissionCondition, user *User, arguments map[string]interface{}) bool {
	values := getArgumentValues(arguments, condition.Argument)
	candidates := expandConditionValue(condition.Value, user)

	switch condition.Operator {
	case "Not equals", "Not in":
		for _, value := range values {
			if util.InSlice(candidates, value) {
				return false
			}
		}
		return true
	case "Starts with":
		if len(values) == 0 {
			return false
		}
		for _, value := range values {
			isHit := false
			for _, candidate := range candidates {
				if strings.HasPrefix(value, candidate) {
					isHit = true
					break
				}
			}
			if !isHit {
				return false
			}
		}
		return true
	default:
		// "Equals" and "In": every value of the argument is one of the candidates
		if len(values) == 0 {
			return false
		}
		for _, value := range values {
			if !util.InSlice(candidates, value) {
				return false
			}
		}
		return true
	}
}

func isConditionsHit(conditions []*PermissionCondition, user *User, arguments map[string]interface{}) bool {
	for _, condition := range conditions {
		if condition == nil || condition.Argument == "" {
			continue
		}
		if !isConditionHit(condition, user, arguments) {
			return false
		}
	}
	return true
}
