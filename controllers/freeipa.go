// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// FreeIPARequest represents a JSON-RPC request in FreeIPA format
type FreeIPARequest struct {
	Method string                 `json:"method"`
	Params []interface{}          `json:"params"`
	ID     int                    `json:"id"`
}

// FreeIPAResponse represents a JSON-RPC response in FreeIPA format
type FreeIPAResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	ID     int         `json:"id"`
}

// FreeIPAUserResult represents user information in FreeIPA format
type FreeIPAUserResult struct {
	UID          []string `json:"uid"`
	UIDNumber    []string `json:"uidnumber"`
	GIDNumber    []string `json:"gidnumber"`
	CN           []string `json:"cn"`
	DisplayName  []string `json:"displayname"`
	Mail         []string `json:"mail"`
	HomeDirectory []string `json:"homedirectory"`
	LoginShell   []string `json:"loginshell"`
	MemberOf     []string `json:"memberof,omitempty"`
}

// FreeIPAJsonRpc handles JSON-RPC requests for FreeIPA compatibility
// @Title FreeIPAJsonRpc
// @Tag FreeIPA API
// @Description Handle JSON-RPC requests compatible with FreeIPA
// @Param   body    body   FreeIPARequest  true   "JSON-RPC request"
// @Success 200 {object} FreeIPAResponse The Response object
// @router /ipa/json [post]
func (c *ApiController) FreeIPAJsonRpc() {
	var req FreeIPARequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.sendFreeIPAError(0, fmt.Sprintf("Invalid JSON-RPC request: %s", err.Error()))
		return
	}

	switch req.Method {
	case "user_show":
		c.handleUserShow(&req)
	case "user_find":
		c.handleUserFind(&req)
	case "ping":
		c.handlePing(&req)
	case "group_show":
		c.handleGroupShow(&req)
	default:
		c.sendFreeIPAError(req.ID, fmt.Sprintf("Unknown method: %s", req.Method))
	}
}

// handlePing handles ping requests
func (c *ApiController) handlePing(req *FreeIPARequest) {
	response := FreeIPAResponse{
		Result: map[string]interface{}{
			"summary": "IPA server version 4.9.0. API version 2.245",
		},
		Error: nil,
		ID:    req.ID,
	}
	c.Data["json"] = response
	c.ServeJSON()
}

// handleUserShow handles user_show method
func (c *ApiController) handleUserShow(req *FreeIPARequest) {
	if len(req.Params) == 0 {
		c.sendFreeIPAError(req.ID, "Missing parameters")
		return
	}

	// Extract username from params
	var username string
	params, ok := req.Params[0].([]interface{})
	if ok && len(params) > 0 {
		username, _ = params[0].(string)
	}

	if username == "" {
		c.sendFreeIPAError(req.ID, "Username is required")
		return
	}

	// Extract organization from params or use default
	organization := "built-in"
	if len(req.Params) > 1 {
		options, ok := req.Params[1].(map[string]interface{})
		if ok {
			if org, exists := options["organization"]; exists {
				organization, _ = org.(string)
			}
		}
	}

	user, err := object.GetUserByFields(organization, username)
	if err != nil {
		c.sendFreeIPAError(req.ID, err.Error())
		return
	}

	if user == nil {
		c.sendFreeIPAError(req.ID, fmt.Sprintf("User %s not found", username))
		return
	}

	userResult := c.convertUserToFreeIPAFormat(user)
	response := FreeIPAResponse{
		Result: map[string]interface{}{
			"result": userResult,
			"value":  username,
		},
		Error: nil,
		ID:    req.ID,
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// handleUserFind handles user_find method
func (c *ApiController) handleUserFind(req *FreeIPARequest) {
	// Extract search criteria
	organization := "built-in"
	var searchTerm string

	if len(req.Params) > 0 {
		params, ok := req.Params[0].([]interface{})
		if ok && len(params) > 0 {
			searchTerm, _ = params[0].(string)
		}
	}

	if len(req.Params) > 1 {
		options, ok := req.Params[1].(map[string]interface{})
		if ok {
			if org, exists := options["organization"]; exists {
				organization, _ = org.(string)
			}
		}
	}

	// Get users
	users, err := object.GetUsers(organization)
	if err != nil {
		c.sendFreeIPAError(req.ID, err.Error())
		return
	}

	// Filter users if search term provided
	var filteredUsers []*object.User
	for _, user := range users {
		if searchTerm == "" || strings.Contains(user.Name, searchTerm) || 
		   strings.Contains(user.DisplayName, searchTerm) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	// Convert users to FreeIPA format
	var results []FreeIPAUserResult
	for _, user := range filteredUsers {
		results = append(results, c.convertUserToFreeIPAFormat(user))
	}

	response := FreeIPAResponse{
		Result: map[string]interface{}{
			"result":   results,
			"count":    len(results),
			"truncated": false,
		},
		Error: nil,
		ID:    req.ID,
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// handleGroupShow handles group_show method
func (c *ApiController) handleGroupShow(req *FreeIPARequest) {
	if len(req.Params) == 0 {
		c.sendFreeIPAError(req.ID, "Missing parameters")
		return
	}

	// Extract group name from params
	var groupName string
	params, ok := req.Params[0].([]interface{})
	if ok && len(params) > 0 {
		groupName, _ = params[0].(string)
	}

	if groupName == "" {
		c.sendFreeIPAError(req.ID, "Group name is required")
		return
	}

	// Extract organization from params or use default
	organization := "built-in"
	if len(req.Params) > 1 {
		options, ok := req.Params[1].(map[string]interface{})
		if ok {
			if org, exists := options["organization"]; exists {
				organization, _ = org.(string)
			}
		}
	}

	groupId := util.GetId(organization, groupName)
	group, err := object.GetGroup(groupId)
	if err != nil {
		c.sendFreeIPAError(req.ID, err.Error())
		return
	}

	if group == nil {
		c.sendFreeIPAError(req.ID, fmt.Sprintf("Group %s not found", groupName))
		return
	}

	// Get group members
	members := object.GetGroupUsersWithoutError(groupId)
	var memberNames []string
	for _, member := range members {
		memberNames = append(memberNames, member.Name)
	}

	groupResult := map[string]interface{}{
		"cn":       []string{group.Name},
		"gidnumber": []string{fmt.Sprintf("%d", hashString(group.Name))},
		"member":   memberNames,
	}

	response := FreeIPAResponse{
		Result: map[string]interface{}{
			"result": groupResult,
			"value":  groupName,
		},
		Error: nil,
		ID:    req.ID,
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// convertUserToFreeIPAFormat converts a Casdoor user to FreeIPA format
func (c *ApiController) convertUserToFreeIPAFormat(user *object.User) FreeIPAUserResult {
	uidNumber := fmt.Sprintf("%d", hashString(user.Name))
	gidNumber := uidNumber // Using same as UID for simplicity
	
	homeDir := fmt.Sprintf("/home/%s", user.Name)
	if user.Homepage != "" {
		homeDir = user.Homepage
	}

	loginShell := "/bin/bash"

	var memberOf []string
	for _, group := range user.Groups {
		memberOf = append(memberOf, group)
	}

	return FreeIPAUserResult{
		UID:          []string{user.Name},
		UIDNumber:    []string{uidNumber},
		GIDNumber:    []string{gidNumber},
		CN:           []string{user.DisplayName},
		DisplayName:  []string{user.DisplayName},
		Mail:         []string{user.Email},
		HomeDirectory: []string{homeDir},
		LoginShell:   []string{loginShell},
		MemberOf:     memberOf,
	}
}

// sendFreeIPAError sends an error response in FreeIPA format
func (c *ApiController) sendFreeIPAError(id int, message string) {
	response := FreeIPAResponse{
		Result: nil,
		Error: map[string]interface{}{
			"code":    2001,
			"message": message,
			"name":    "ExecutionError",
		},
		ID: id,
	}
	c.Data["json"] = response
	c.ServeJSON()
}

// hashString generates a consistent hash for strings (for UID/GID numbers)
func hashString(s string) uint32 {
	h := uint32(0)
	for i := 0; i < len(s); i++ {
		h = 31*h + uint32(s[i])
	}
	// Ensure it's in a reasonable range for Unix UID/GID (1000-60000)
	return 1000 + (h % 59000)
}

// FreeIPASessionJson handles session-based authentication for FreeIPA
// @Title FreeIPASessionJson
// @Tag FreeIPA API
// @Description Handle authenticated JSON-RPC requests with session
// @Param   body    body   FreeIPARequest  true   "JSON-RPC request"
// @Success 200 {object} FreeIPAResponse The Response object
// @router /ipa/session/json [post]
func (c *ApiController) FreeIPASessionJson() {
	// Check authentication
	username := c.GetSessionUsername()
	if username == "" {
		c.sendFreeIPAError(0, "Authentication required")
		return
	}

	// Delegate to the regular JSON-RPC handler
	c.FreeIPAJsonRpc()
}

// FreeIPASessionLogin handles login for FreeIPA session-based authentication
// @Title FreeIPASessionLogin
// @Tag FreeIPA API
// @Description Handle session login for FreeIPA compatibility
// @Param   username    formData    string  true   "Username"
// @Param   password    formData    string  true   "Password"
// @Success 200 {string} string "OK"
// @router /ipa/session/login_password [post]
func (c *ApiController) FreeIPASessionLogin() {
	username := c.GetString("username")
	password := c.GetString("password")
	organization := c.GetString("organization")

	if organization == "" {
		organization = "built-in"
	}

	if username == "" || password == "" {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{
			"error": "Username and password are required",
		}
		c.ServeJSON()
		return
	}

	user, err := object.CheckUserPassword(organization, username, password, c.GetAcceptLanguage())
	if err != nil {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{
			"error": err.Error(),
		}
		c.ServeJSON()
		return
	}

	if user == nil {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{
			"error": "Invalid credentials",
		}
		c.ServeJSON()
		return
	}

	// Set session
	c.SetSessionUsername(user.GetId())

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = map[string]string{
		"status": "ok",
	}
	c.ServeJSON()
}

// FreeIPASessionLogout handles logout for FreeIPA session-based authentication
// @Title FreeIPASessionLogout
// @Tag FreeIPA API
// @Description Handle session logout for FreeIPA compatibility
// @Success 200 {string} string "OK"
// @router /ipa/session/logout [post]
func (c *ApiController) FreeIPASessionLogout() {
	c.SetSessionUsername("")
	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = map[string]string{
		"status": "ok",
	}
	c.ServeJSON()
}
