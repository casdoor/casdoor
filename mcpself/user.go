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

package mcpself

import (
	"fmt"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// handleGetUsersTool handles the get_users MCP tool
func (c *McpController) handleGetUsersTool(id interface{}, args GetUsersArgs) {
	users, err := object.GetMaskedUsers(object.GetUsers(args.Owner))
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, util.StructToJsonFormatted(users))
}

// handleGetUserTool handles the get_user MCP tool
func (c *McpController) handleGetUserTool(id interface{}, args GetUserArgs) {
	var user *object.User
	var err error

	switch {
	case args.Id != "":
		user, err = object.GetUser(args.Id)
	case args.Email != "" && args.Owner != "":
		user, err = object.GetUserByEmail(args.Owner, args.Email)
	case args.Phone != "" && args.Owner != "":
		user, err = object.GetUserByPhone(args.Owner, args.Phone)
	default:
		c.SendToolErrorResult(id, "must provide id, or owner+email, or owner+phone")
		return
	}

	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	maskedUser, err := object.GetMaskedUser(user, true)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, util.StructToJsonFormatted(maskedUser))
}

// handleAddUserTool handles the add_user MCP tool
func (c *McpController) handleAddUserTool(id interface{}, args AddUserArgs) {
	if err := checkQuotaForUser(); err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	emptyUser := object.User{}
	if msg := object.CheckUpdateUser(&emptyUser, &args.User, c.GetAcceptLanguage()); msg != "" {
		c.SendToolErrorResult(id, msg)
		return
	}

	if args.User.RegisterType == "" {
		args.User.RegisterType = "Add User"
	}

	affected, err := object.AddUser(&args.User, c.GetAcceptLanguage())
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, FormatOperationResult("add", "user", affected))
}

// handleUpdateUserTool handles the update_user MCP tool
func (c *McpController) handleUpdateUserTool(id interface{}, args UpdateUserArgs) {
	oldUser, err := object.GetUser(args.Id)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}
	if oldUser == nil {
		c.SendToolErrorResult(id, fmt.Sprintf("user %s not found", args.Id))
		return
	}

	if oldUser.Owner == "built-in" && oldUser.Name == "admin" &&
		(args.User.Owner != "built-in" || args.User.Name != "admin") {
		c.SendToolErrorResult(id, "cannot modify the built-in admin user identity")
		return
	}

	if msg := object.CheckUpdateUser(oldUser, &args.User, c.GetAcceptLanguage()); msg != "" {
		c.SendToolErrorResult(id, msg)
		return
	}

	affected, err := object.UpdateUser(args.Id, &args.User, []string{}, c.IsGlobalAdmin())
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	if affected {
		if err = object.UpdateUserToOriginalDatabase(&args.User); err != nil {
			c.SendToolErrorResult(id, err.Error())
			return
		}
	}

	c.SendToolResult(id, FormatOperationResult("update", "user", affected))
}

// handleDeleteUserTool handles the delete_user MCP tool
func (c *McpController) handleDeleteUserTool(id interface{}, args DeleteUserArgs) {
	if args.User.Owner == "built-in" && args.User.Name == "admin" {
		c.SendToolErrorResult(id, "cannot delete the built-in admin user")
		return
	}

	affected, err := object.DeleteUser(&args.User)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, FormatOperationResult("delete", "user", affected))
}

// checkQuotaForUser checks if the user quota is exceeded
func checkQuotaForUser() error {
	quota := conf.GetConfigQuota().User
	if quota == -1 {
		return nil
	}

	count, err := object.GetGlobalUserCount("", "")
	if err != nil {
		return err
	}

	if int(count) >= quota {
		return fmt.Errorf("user quota is exceeded")
	}
	return nil
}
