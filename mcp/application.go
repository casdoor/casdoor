// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"encoding/json"
	"fmt"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
)

// HandleGetApplicationsTool handles the get_applications MCP tool
func (c *MCPController) HandleGetApplicationsTool(id interface{}, args map[string]interface{}) {
	userId := c.GetSessionUsername()
	owner, ok := args["owner"].(string)
	if !ok {
		c.SendMCPError(id, -32602, "Invalid params", "Missing or invalid 'owner' parameter")
		return
	}

	applications, err := object.GetApplications(owner)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	maskedApps := object.GetMaskedApplications(applications, userId)
	jsonData, err := json.MarshalIndent(maskedApps, "", "  ")
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, string(jsonData))
}

// HandleGetApplicationTool handles the get_application MCP tool
func (c *MCPController) HandleGetApplicationTool(id interface{}, args map[string]interface{}) {
	userId := c.GetSessionUsername()
	appId, ok := args["id"].(string)
	if !ok {
		c.SendMCPError(id, -32602, "Invalid params", "Missing or invalid 'id' parameter")
		return
	}

	application, err := object.GetApplication(appId)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	maskedApp := object.GetMaskedApplication(application, userId)
	jsonData, err := json.MarshalIndent(maskedApp, "", "  ")
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, string(jsonData))
}

// HandleAddApplicationTool handles the add_application MCP tool
func (c *MCPController) HandleAddApplicationTool(id interface{}, args map[string]interface{}) {
	appData, ok := args["application"].(map[string]interface{})
	if !ok {
		c.SendMCPError(id, -32602, "Invalid params", "Missing or invalid 'application' parameter")
		return
	}

	jsonBytes, err := json.Marshal(appData)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	var application object.Application
	err = json.Unmarshal(jsonBytes, &application)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	count, err := object.GetApplicationCount("", "", "")
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	if err := checkQuotaForApplication(int(count)); err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	if err = object.CheckIpWhitelist(application.IpWhitelist, c.GetAcceptLanguage()); err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	affected, err := object.AddApplication(&application)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, FormatOperationResult("add", "application", affected))
}

// HandleUpdateApplicationTool handles the update_application MCP tool
func (c *MCPController) HandleUpdateApplicationTool(id interface{}, args map[string]interface{}) {
	appId, ok := args["id"].(string)
	if !ok {
		c.SendMCPError(id, -32602, "Invalid params", "Missing or invalid 'id' parameter")
		return
	}

	appData, ok := args["application"].(map[string]interface{})
	if !ok {
		c.SendMCPError(id, -32602, "Invalid params", "Missing or invalid 'application' parameter")
		return
	}

	jsonBytes, err := json.Marshal(appData)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	var application object.Application
	err = json.Unmarshal(jsonBytes, &application)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	if err = object.CheckIpWhitelist(application.IpWhitelist, c.GetAcceptLanguage()); err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	affected, err := object.UpdateApplication(appId, &application, c.IsGlobalAdmin(), c.GetAcceptLanguage())
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	c.SendToolResult(id, FormatOperationResult("update", "application", affected))
}

// HandleDeleteApplicationTool handles the delete_application MCP tool
func (c *MCPController) HandleDeleteApplicationTool(id interface{}, args map[string]interface{}) {
	appData, ok := args["application"].(map[string]interface{})
	if !ok {
		c.SendMCPError(id, -32602, "Invalid params", "Missing or invalid 'application' parameter")
		return
	}

	jsonBytes, err := json.Marshal(appData)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	var application object.Application
	err = json.Unmarshal(jsonBytes, &application)
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}

	affected, err := object.DeleteApplication(&application)
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
