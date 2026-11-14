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
	"fmt"
	"slices"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/mcp"
	"github.com/casdoor/casdoor/util"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xorm-io/core"
)

type Tool struct {
	mcpsdk.Tool
	IsAllowed       bool `json:"isAllowed"`
	RequireApproval bool `json:"requireApproval"`
}

type Server struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Url         string  `xorm:"varchar(500)" json:"url"`
	Token       string  `xorm:"mediumtext" json:"token"`
	TokenHeader string  `xorm:"varchar(100)" json:"tokenHeader"`
	Application string  `xorm:"varchar(100)" json:"application"`
	Tools       []*Tool `xorm:"mediumtext" json:"tools"`

	// IdentityHeader is the request header that carries the caller's identity token to the upstream MCP server
	IdentityHeader string `xorm:"varchar(100)" json:"identityHeader"`
	// EnableToolPermissions makes every tool call need an "MCP tool" permission of the user and of the agent
	EnableToolPermissions bool `json:"enableToolPermissions"`
}

func (server *Server) GetTool(name string) *Tool {
	for _, tool := range server.Tools {
		if tool.Name == name {
			return tool
		}
	}
	return nil
}

func checkServer(server *Server) error {
	if server.IdentityHeader != "" && server.Application == "" {
		return fmt.Errorf("the identity token is signed with the certificate of the application, please select an application for the MCP server: %s", server.GetId())
	}
	for _, tool := range server.Tools {
		if tool.RequireApproval {
			if err := edition.Require(edition.Enterprise, edition.FeatureMcpApproval); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetMaskedServer clears the token used to access the upstream MCP server, so that it is
// never returned to the frontend. An empty token means "keep the old one" in UpdateServer().
func GetMaskedServer(server *Server) *Server {
	if server == nil {
		return nil
	}

	server.Token = ""
	return server
}

func GetMaskedServers(servers []*Server) []*Server {
	for _, server := range servers {
		server = GetMaskedServer(server)
	}
	return servers
}

func GetServers(owner string) ([]*Server, error) {
	servers := []*Server{}
	err := ormer.Engine.Desc("created_time").Find(&servers, &Server{Owner: owner})
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func getServer(owner string, name string) (*Server, error) {
	server := Server{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&server)
	if err != nil {
		return nil, err
	}

	if existed {
		return &server, nil
	}
	return nil, nil
}

func GetServer(id string) (*Server, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getServer(owner, name)
}

func UpdateServer(id string, server *Server) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	oldServer, err := getServer(owner, name)
	if err != nil {
		return false, err
	}
	if oldServer == nil {
		return false, nil
	}

	if server.Token == "" {
		server.Token = oldServer.Token
	}

	server.UpdatedTime = util.GetCurrentTime()

	_ = syncServerTools(server)

	if err = checkServer(server); err != nil {
		return false, err
	}

	_, err = ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(server)
	if err != nil {
		return false, err
	}

	return true, nil
}

func SyncMcpTool(id string, server *Server, isCleared bool) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)

	if isCleared {
		server.Tools = nil
		server.UpdatedTime = util.GetCurrentTime()
		_, err := ormer.Engine.ID(core.PK{owner, name}).Cols("tools", "updated_time").Update(server)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	oldServer, err := getServer(owner, name)
	if err != nil {
		return false, err
	}
	if oldServer == nil {
		return false, nil
	}

	if server.Token == "" {
		server.Token = oldServer.Token
	}

	server.UpdatedTime = util.GetCurrentTime()

	err = syncServerTools(server)
	if err != nil {
		return false, err
	}

	_, err = ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(server)
	if err != nil {
		return false, err
	}

	return true, nil
}

func syncServerTools(server *Server) error {
	oldTools := server.Tools
	if oldTools == nil {
		oldTools = []*Tool{}
	}

	tools, err := mcp.GetServerTools(server.Owner, server.Name, server.Url, server.Token)
	if err != nil {
		return err
	}

	var newTools []*Tool
	for _, tool := range tools {
		oldToolIndex := slices.IndexFunc(oldTools, func(oldTool *Tool) bool {
			return oldTool.Name == tool.Name
		})

		isAllowed := true
		requireApproval := false
		if oldToolIndex != -1 {
			isAllowed = oldTools[oldToolIndex].IsAllowed
			requireApproval = oldTools[oldToolIndex].RequireApproval
		}

		newTool := Tool{
			Tool:            *tool,
			IsAllowed:       isAllowed,
			RequireApproval: requireApproval,
		}
		newTools = append(newTools, &newTool)
	}

	server.Tools = newTools
	return nil
}

func AddServer(server *Server) (bool, error) {
	if err := checkServer(server); err != nil {
		return false, err
	}

	affected, err := ormer.Engine.Insert(server)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteServer(server *Server) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{server.Owner, server.Name}).Delete(&Server{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (server *Server) GetId() string {
	return fmt.Sprintf("%s/%s", server.Owner, server.Name)
}

func GetServerCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Server{})
}

func GetPaginationServers(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Server, error) {
	servers := []*Server{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&servers)
	if err != nil {
		return servers, err
	}

	return servers, nil
}

// IsServerAllowed reports whether a user of the server's organization may use an MCP server through the proxy:
// the users who may sign in to the application of the server
func IsServerAllowed(user *User, owner string, name string) (bool, error) {
	server, err := GetServer(util.GetId(owner, name))
	if err != nil || server == nil || server.Application == "" || user.Owner != owner {
		return false, err
	}

	application, err := getApplication("admin", server.Application)
	if err != nil || application == nil {
		return false, err
	}
	if application.Organization != owner && !application.IsShared {
		return false, nil
	}

	return CheckLoginPermission(user.GetId(), application)
}

// UserMcpServer is an MCP server as the agents of a user see it: only its gateway address and its allowed tools
type UserMcpServer struct {
	Owner                 string         `json:"owner"`
	Name                  string         `json:"name"`
	DisplayName           string         `json:"displayName"`
	Url                   string         `json:"url"`
	EnableToolPermissions bool           `json:"enableToolPermissions"`
	Tools                 []*UserMcpTool `json:"tools"`
}

type UserMcpTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetUserMcpServers lists the MCP servers of the user's organization with their gateway addresses, so an agent signed
// in as the user can connect to them, the upstream addresses and tokens stay hidden
func GetUserMcpServers(user *User, host string) ([]*UserMcpServer, error) {
	servers, err := GetServers(user.Owner)
	if err != nil {
		return nil, err
	}

	res := []*UserMcpServer{}
	for _, server := range servers {
		tools := []*UserMcpTool{}
		for _, tool := range server.Tools {
			if tool.IsAllowed {
				tools = append(tools, &UserMcpTool{Name: tool.Name, Description: tool.Description})
			}
		}
		res = append(res, &UserMcpServer{
			Owner:                 server.Owner,
			Name:                  server.Name,
			DisplayName:           server.DisplayName,
			Url:                   GetServerProxyUrl(host, server.Owner, server.Name),
			EnableToolPermissions: server.EnableToolPermissions,
			Tools:                 tools,
		})
	}
	return res, nil
}
