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

package mcp

import (
	"fmt"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// handleGetApplicationsTool handles the get_applications MCP tool
func (c *McpController) handleGetApplicationsTool(id interface{}, args GetApplicationsArgs) {
	userId := c.GetSessionUsername()

	applications, err := object.GetApplications(args.Owner)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	maskedApps := object.GetMaskedApplications(applications, userId)
	c.SendToolResult(id, util.StructToJsonFormatted(maskedApps))
}

// handleGetApplicationTool handles the get_application MCP tool
func (c *McpController) handleGetApplicationTool(id interface{}, args GetApplicationArgs) {
	userId := c.GetSessionUsername()

	application, err := object.GetApplication(args.Id)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	maskedApp := object.GetMaskedApplication(application, userId)
	c.SendToolResult(id, util.StructToJsonFormatted(maskedApp))
}

// handleAddApplicationTool handles the add_application MCP tool
func (c *McpController) handleAddApplicationTool(id interface{}, args AddApplicationArgs) {
	count, err := object.GetApplicationCount("", "", "")
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	if err := checkQuotaForApplication(int(count)); err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	if err = object.CheckIpWhitelist(args.Application.IpWhitelist, c.GetAcceptLanguage()); err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	affected, err := object.AddApplication(&args.Application)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, FormatOperationResult("add", "application", affected))
}

// handleUpdateApplicationTool handles the update_application MCP tool
func (c *McpController) handleUpdateApplicationTool(id interface{}, args UpdateApplicationArgs) {
	if err := object.CheckIpWhitelist(args.Application.IpWhitelist, c.GetAcceptLanguage()); err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	affected, err := object.UpdateApplication(args.Id, &args.Application, c.IsGlobalAdmin(), c.GetAcceptLanguage())
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, FormatOperationResult("update", "application", affected))
}

// handleDeleteApplicationTool handles the delete_application MCP tool
func (c *McpController) handleDeleteApplicationTool(id interface{}, args DeleteApplicationArgs) {
	affected, err := object.DeleteApplication(&args.Application)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, FormatOperationResult("delete", "application", affected))
}

// checkQuotaForApplication checks if the application quota is exceeded
func checkQuotaForApplication(count int) error {
	quota := conf.GetConfigQuota().Application
	if quota == -1 {
		return nil
	}
	if count >= quota {
		return fmt.Errorf("application quota is exceeded")
	}
	return nil
}
