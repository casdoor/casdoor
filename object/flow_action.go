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
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

var flowOperators = []string{
	"equals", "notEquals", "contains", "notContains", "startsWith", "endsWith", "in", "notIn",
	"matches", "greaterThan", "lessThan", "exists", "notExists", "inNetwork", "notInNetwork",
}

// flowUserFields are the user fields that a flow can set, by their JSON names
var flowUserFields = map[string]string{
	"displayName": "display_name",
	"firstName":   "first_name",
	"lastName":    "last_name",
	"avatar":      "avatar",
	"tag":         "tag",
	"region":      "region",
	"location":    "location",
	"affiliation": "affiliation",
	"title":       "title",
	"homepage":    "homepage",
	"bio":         "bio",
	"language":    "language",
	"gender":      "gender",
	"education":   "education",
}

var reservedClaims = []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti", "nonce", "azp", "at_hash", "c_hash", "auth_time", "acr", "amr", "scope", "tokenType", "owner", "name", "id"}

func isReservedClaim(name string) bool {
	return util.InSlice(reservedClaims, name)
}

func isFlowUserFieldWritable(field string) bool {
	if strings.HasPrefix(field, "properties.") {
		return len(field) > len("properties.")
	}
	_, ok := flowUserFields[field]
	return ok
}

// getFlowFieldValue reads a field of the flow's environment: a string, a number, a bool or a list of strings
func getFlowFieldValue(env *FlowEnv, field string) interface{} {
	user := env.User
	switch {
	case strings.HasPrefix(field, "user.properties."):
		return user.Properties[strings.TrimPrefix(field, "user.properties.")]
	case strings.HasPrefix(field, "claims."):
		return env.Claims[strings.TrimPrefix(field, "claims.")]
	}

	switch field {
	case "user.name":
		return user.Name
	case "user.displayName":
		return user.DisplayName
	case "user.email":
		return user.Email
	case "user.emailDomain":
		if i := strings.LastIndex(user.Email, "@"); i >= 0 {
			return strings.ToLower(user.Email[i+1:])
		}
		return ""
	case "user.emailVerified":
		return user.EmailVerified
	case "user.phone":
		return user.Phone
	case "user.countryCode":
		return user.CountryCode
	case "user.type":
		return user.Type
	case "user.tag":
		return user.Tag
	case "user.region":
		return user.Region
	case "user.affiliation":
		return user.Affiliation
	case "user.title":
		return user.Title
	case "user.groups":
		return getShortNames(user.Groups)
	case "user.roles":
		return getFlowUserRoles(user)
	case "user.isAdmin":
		return user.IsAdmin
	case "user.isMfaEnabled":
		return user.IsMfaEnabled()
	case "user.signupApplication":
		return user.SignupApplication
	case "user.createdTime":
		return user.CreatedTime
	case "user.accountAgeDays":
		if user.CreatedTime == "" {
			return 0
		}
		return int(time.Since(util.String2Time(user.CreatedTime)).Hours() / 24)
	case "application.name":
		return env.Application
	case "request.ip":
		return env.Ip
	case "request.userAgent":
		return env.UserAgent
	case "request.country":
		return env.Country
	case "request.hour":
		return time.Now().Hour()
	case "request.weekday":
		return strings.ToLower(time.Now().Weekday().String())
	case "risk.score":
		if env.Risk == nil {
			return 0
		}
		return env.Risk.Score
	case "risk.signals":
		if env.Risk == nil {
			return []string{}
		}
		return env.Risk.Signals
	case "risk.decision":
		if env.Risk == nil {
			return "Allow"
		}
		return env.Risk.Decision
	case "risk.location":
		if env.Risk == nil {
			return ""
		}
		return env.Risk.Location
	}
	return nil
}

func getShortNames(ids []string) []string {
	res := []string{}
	for _, id := range ids {
		if i := strings.Index(id, "/"); i >= 0 {
			id = id[i+1:]
		}
		res = append(res, id)
	}
	return res
}

func getFlowUserRoles(user *User) []string {
	roles, err := GetRolesByUser(user.GetId())
	if err != nil {
		return []string{}
	}
	res := []string{}
	for _, role := range roles {
		res = append(res, role.Name)
	}
	return res
}

func formatFlowValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case []string:
		return strings.Join(v, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func splitFlowList(value string) []string {
	res := []string{}
	for _, item := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '\n' }) {
		if item = strings.TrimSpace(item); item != "" {
			res = append(res, item)
		}
	}
	return res
}

func evaluateFlowCondition(actual interface{}, operator string, expected string) bool {
	if list, ok := actual.([]string); ok {
		return evaluateFlowListCondition(list, operator, expected)
	}

	text := formatFlowValue(actual)
	lowerText, lowerExpected := strings.ToLower(text), strings.ToLower(expected)
	switch operator {
	case "equals":
		return lowerText == lowerExpected
	case "notEquals":
		return lowerText != lowerExpected
	case "contains":
		return strings.Contains(lowerText, lowerExpected)
	case "notContains":
		return !strings.Contains(lowerText, lowerExpected)
	case "startsWith":
		return strings.HasPrefix(lowerText, lowerExpected)
	case "endsWith":
		return strings.HasSuffix(lowerText, lowerExpected)
	case "in", "notIn":
		found := false
		for _, item := range splitFlowList(expected) {
			if strings.EqualFold(item, text) {
				found = true
				break
			}
		}
		return found == (operator == "in")
	case "matches":
		re, err := regexp.Compile(expected)
		return err == nil && re.MatchString(text)
	case "greaterThan", "lessThan":
		a, errA := strconv.ParseFloat(text, 64)
		b, errB := strconv.ParseFloat(expected, 64)
		if errA != nil || errB != nil {
			return false
		}
		if operator == "greaterThan" {
			return a > b
		}
		return a < b
	case "exists":
		return text != "" && text != "false" && text != "0"
	case "notExists":
		return text == "" || text == "false" || text == "0"
	case "inNetwork":
		return IsIpInNetworks(text, splitFlowList(expected))
	case "notInNetwork":
		return !IsIpInNetworks(text, splitFlowList(expected))
	}
	return false
}

func evaluateFlowListCondition(list []string, operator string, expected string) bool {
	has := func(item string) bool {
		for _, v := range list {
			if strings.EqualFold(v, item) {
				return true
			}
		}
		return false
	}
	switch operator {
	case "contains", "equals":
		return has(strings.TrimSpace(expected))
	case "notContains", "notEquals":
		return !has(strings.TrimSpace(expected))
	case "in":
		for _, item := range splitFlowList(expected) {
			if has(item) {
				return true
			}
		}
		return false
	case "notIn":
		for _, item := range splitFlowList(expected) {
			if has(item) {
				return false
			}
		}
		return true
	case "exists":
		return len(list) > 0
	case "notExists":
		return len(list) == 0
	}
	return false
}

var flowTemplateRegex = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_.\-]+)\s*\}\}`)

// renderFlowTemplate replaces the {{field}} in a parameter with the fields of the environment
func renderFlowTemplate(text string, env *FlowEnv) string {
	if !strings.Contains(text, "{{") {
		return text
	}
	return flowTemplateRegex.ReplaceAllStringFunc(text, func(match string) string {
		field := flowTemplateRegex.FindStringSubmatch(match)[1]
		return formatFlowValue(getFlowFieldValue(env, field))
	})
}

func runFlowAction(step *FlowStep, env *FlowEnv, result *FlowResult) (string, error) {
	param := func(key string) string {
		return renderFlowTemplate(step.Params[key], env)
	}
	user := env.User

	switch step.Action {
	case FlowActionDeny:
		message := param("message")
		if message == "" {
			message = "Access denied"
		}
		result.Deny = message
		return message, nil
	case FlowActionRequireMfa:
		result.RequireMfa = true
		return "MFA required", nil
	case FlowActionStop:
		return "", nil
	case FlowActionAddClaim:
		name, value := step.Params["name"], param("value")
		var parsed interface{} = value
		if step.Params["valueType"] == "json" {
			if err := json.Unmarshal([]byte(value), &parsed); err != nil {
				return "", fmt.Errorf("the value of the claim: %s is not JSON: %s", name, err.Error())
			}
		}
		result.Claims[name] = parsed
		if env.Claims != nil {
			env.Claims[name] = parsed
		}
		return fmt.Sprintf("%s = %s", name, value), nil
	case FlowActionSetUserField:
		return setFlowUserField(env, step.Params["field"], param("value"))
	case FlowActionAddToGroup, FlowActionRemoveFromGroup:
		return changeFlowUserGroup(env, param("group"), step.Action == FlowActionAddToGroup)
	case FlowActionAddToRole, FlowActionRemoveFromRole:
		return changeFlowUserRole(env, param("role"), step.Action == FlowActionAddToRole)
	case FlowActionHttp:
		return callFlowHttp(step, env, result)
	case FlowActionSendEmail:
		to := param("to")
		if to == "" {
			to = user.Email
		}
		dest := splitFlowList(to)
		if len(dest) == 0 {
			return "no recipient", nil
		}
		if env.DryRun {
			return fmt.Sprintf("would email %s", strings.Join(dest, ", ")), nil
		}
		organization, err := LookupOrganization("admin", user.Owner)
		if err != nil {
			return "", err
		}
		if organization == nil {
			return "", fmt.Errorf("the organization: %s doesn't exist", user.Owner)
		}
		if err = SendOrganizationEmail(organization, param("subject"), param("content"), dest); err != nil {
			return "", err
		}
		return fmt.Sprintf("emailed %s", strings.Join(dest, ", ")), nil
	}
	return "", fmt.Errorf("unknown action: %s", step.Action)
}

func setFlowUserField(env *FlowEnv, field string, value string) (string, error) {
	user := env.User
	column := ""
	if strings.HasPrefix(field, "properties.") {
		if user.Properties == nil {
			user.Properties = map[string]string{}
		}
		user.Properties[strings.TrimPrefix(field, "properties.")] = value
		column = "properties"
	} else {
		var ok bool
		column, ok = flowUserFields[field]
		if !ok {
			return "", fmt.Errorf("the user field: %s can't be set by a flow", field)
		}
		data, err := json.Marshal(user)
		if err != nil {
			return "", err
		}
		fields := map[string]interface{}{}
		if err = json.Unmarshal(data, &fields); err != nil {
			return "", err
		}
		fields[field] = value
		if data, err = json.Marshal(fields); err != nil {
			return "", err
		}
		if err = json.Unmarshal(data, user); err != nil {
			return "", err
		}
	}

	message := fmt.Sprintf("user.%s = %s", field, value)
	if env.DryRun || env.Unsaved {
		return message, nil
	}
	_, err := UpdateUser(user.GetId(), user, []string{column}, true)
	return message, err
}

func changeFlowUserGroup(env *FlowEnv, group string, add bool) (string, error) {
	user := env.User
	groupId := group
	if !strings.Contains(groupId, "/") {
		groupId = util.GetId(user.Owner, group)
	}
	if !add && !util.InSlice(user.Groups, groupId) {
		return fmt.Sprintf("not in %s", group), nil
	}
	if add && util.InSlice(user.Groups, groupId) {
		return fmt.Sprintf("already in %s", group), nil
	}

	if !env.Unsaved {
		existing, err := GetGroup(groupId)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return "", fmt.Errorf("the group: %s doesn't exist", groupId)
		}
	}

	if add {
		user.Groups = append(user.Groups, groupId)
	} else {
		groups := []string{}
		for _, item := range user.Groups {
			if item != groupId {
				groups = append(groups, item)
			}
		}
		user.Groups = groups
	}

	verb := map[bool]string{true: "added to", false: "removed from"}[add]
	if env.DryRun || env.Unsaved {
		return fmt.Sprintf("%s %s", verb, group), nil
	}
	_, err := UpdateUser(user.GetId(), user, []string{"groups"}, true)
	return fmt.Sprintf("%s %s", verb, group), err
}

func changeFlowUserRole(env *FlowEnv, roleName string, add bool) (string, error) {
	user := env.User
	roleId := roleName
	if !strings.Contains(roleId, "/") {
		roleId = util.GetId(user.Owner, roleName)
	}
	role, err := GetRole(roleId)
	if err != nil {
		return "", err
	}
	if role == nil {
		return "", fmt.Errorf("the role: %s doesn't exist", roleId)
	}

	userId := user.GetId()
	has := util.InSlice(role.Users, userId)
	if add == has {
		return fmt.Sprintf("%s already %s", userId, map[bool]string{true: "has the role", false: "doesn't have the role"}[add]), nil
	}
	if add {
		role.Users = append(role.Users, userId)
	} else {
		users := []string{}
		for _, item := range role.Users {
			if item != userId {
				users = append(users, item)
			}
		}
		role.Users = users
	}

	verb := map[bool]string{true: "granted", false: "revoked"}[add]
	if env.DryRun {
		return fmt.Sprintf("would be %s %s", verb, roleName), nil
	}
	_, err = UpdateRole(roleId, role, true, "en")
	return fmt.Sprintf("%s %s", verb, roleName), err
}

// flowHttpResponse is what the service that an HTTP step calls can answer to change the flow
type flowHttpResponse struct {
	Deny       string                 `json:"deny"`
	RequireMfa bool                   `json:"requireMfa"`
	Claims     map[string]interface{} `json:"claims"`
	User       map[string]string      `json:"user"`
	Stop       bool                   `json:"stop"`
}

func getFlowPayload(env *FlowEnv) map[string]interface{} {
	user := env.User
	payload := map[string]interface{}{
		"trigger":     env.Trigger,
		"application": env.Application,
		"time":        util.GetCurrentTime(),
		"user": map[string]interface{}{
			"owner":       user.Owner,
			"name":        user.Name,
			"id":          user.Id,
			"displayName": user.DisplayName,
			"email":       user.Email,
			"phone":       user.Phone,
			"type":        user.Type,
			"tag":         user.Tag,
			"groups":      user.Groups,
			"properties":  user.Properties,
			"createdTime": user.CreatedTime,
		},
		"request": map[string]interface{}{
			"ip":        env.Ip,
			"userAgent": env.UserAgent,
			"country":   env.Country,
		},
	}
	if env.Risk != nil {
		payload["risk"] = env.Risk
	}
	if env.Claims != nil {
		payload["claims"] = env.Claims
	}
	return payload
}

func callFlowHttp(step *FlowStep, env *FlowEnv, result *FlowResult) (string, error) {
	url := renderFlowTemplate(step.Params["url"], env)
	method := strings.ToUpper(step.Params["method"])
	if method == "" {
		method = http.MethodPost
	}
	if env.DryRun && step.Params["callInTest"] != "true" {
		return fmt.Sprintf("would call %s %s", method, url), nil
	}

	body, err := json.Marshal(getFlowPayload(env))
	if err != nil {
		return "", err
	}
	timeout := 5 * time.Second
	if seconds, err := strconv.Atoi(step.Params["timeout"]); err == nil && seconds > 0 && seconds <= 30 {
		timeout = time.Duration(seconds) * time.Second
	}

	var reader io.Reader
	if method != http.MethodGet {
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Casdoor-Flow")
	if secret := step.Params["secret"]; secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		request.Header.Set("X-Casdoor-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}
	for _, line := range strings.Split(step.Params["headers"], "\n") {
		if key, value, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(key) != "" {
			request.Header.Set(strings.TrimSpace(key), renderFlowTemplate(strings.TrimSpace(value), env))
		}
	}

	client := &http.Client{Timeout: timeout}
	if IsTenantOwner(env.User.Owner) {
		client = util.NewInternetOnlyHttpClient(timeout)
	}
	onError := func(err error) (string, error) {
		if step.Params["onError"] == "deny" {
			result.Deny = "The sign-in can't be checked right now, please try again later"
			return err.Error(), nil
		}
		if step.Params["onError"] == "continue" {
			return "ignored: " + err.Error(), nil
		}
		return "", err
	}

	resp, err := client.Do(request)
	if err != nil {
		return onError(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return onError(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return onError(fmt.Errorf("%s %s returned %d: %s", method, url, resp.StatusCode, TruncateString(string(data), 200)))
	}

	answer := flowHttpResponse{}
	if len(bytes.TrimSpace(data)) > 0 {
		if err = json.Unmarshal(data, &answer); err != nil {
			return fmt.Sprintf("%d, the response is not JSON and is ignored", resp.StatusCode), nil
		}
	}

	allowed := flowTriggerActions[env.Trigger]
	applied := []string{}
	if answer.Deny != "" && util.InSlice(allowed, FlowActionDeny) {
		result.Deny = answer.Deny
		applied = append(applied, "deny")
	}
	if answer.RequireMfa && util.InSlice(allowed, FlowActionRequireMfa) {
		result.RequireMfa = true
		applied = append(applied, "requireMfa")
	}
	if len(answer.Claims) > 0 && util.InSlice(allowed, FlowActionAddClaim) {
		for name, value := range answer.Claims {
			if isReservedClaim(name) {
				continue
			}
			result.Claims[name] = value
			if env.Claims != nil {
				env.Claims[name] = value
			}
		}
		applied = append(applied, "claims")
	}
	if len(answer.User) > 0 && util.InSlice(allowed, FlowActionSetUserField) {
		for field, value := range answer.User {
			if !isFlowUserFieldWritable(field) {
				continue
			}
			if _, err = setFlowUserField(env, field, value); err != nil {
				return "", err
			}
		}
		applied = append(applied, "user")
	}
	if answer.Stop {
		result.stop = true
		applied = append(applied, "stop")
	}

	message := fmt.Sprintf("%d", resp.StatusCode)
	if len(applied) > 0 {
		message += ", applied " + strings.Join(applied, ", ")
	}
	return message, nil
}
