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

package controllers

import (
	"net/http"

	"github.com/casdoor/casdoor/object"
)

// GetOauthProtectedResourceMetadata
// @Title GetOauthProtectedResourceMetadata
// @Tag OAuth 2.0 API
// @Description Get OAuth 2.0 Protected Resource Metadata (RFC 9728)
// @Success 200 {object} object.OauthProtectedResourceMetadata
// @router /.well-known/oauth-protected-resource [get]
func (c *RootController) GetOauthProtectedResourceMetadata() {
	host := c.Ctx.Request.Host
	c.Data["json"] = object.GetOauthProtectedResourceMetadata(host)
	c.ServeJSON()
}

// GetOauthProtectedResourceMetadataByApplication
// @Title GetOauthProtectedResourceMetadataByApplication
// @Tag OAuth 2.0 API
// @Description Get OAuth 2.0 Protected Resource Metadata for specific application (RFC 9728)
// @Param application path string true "application name"
// @Success 200 {object} object.OauthProtectedResourceMetadata
// @router /.well-known/:application/oauth-protected-resource [get]
func (c *RootController) GetOauthProtectedResourceMetadataByApplication() {
	application := c.Ctx.Input.Param(":application")
	host := c.Ctx.Request.Host
	c.Data["json"] = object.GetOauthProtectedResourceMetadataByApplication(host, application)
	c.ServeJSON()
}

// GetOauthProtectedResourceMetadataByServer
// @Title GetOauthProtectedResourceMetadataByServer
// @Tag OAuth 2.0 API
// @Description Get OAuth 2.0 Protected Resource Metadata for an MCP server behind the proxy (RFC 9728)
// @Param owner path string true "The owner name of the server"
// @Param name path string true "The name of the server"
// @Success 200 {object} object.OauthProtectedResourceMetadata
// @router /.well-known/oauth-protected-resource/api/server/:owner/:name [get]
func (c *RootController) GetOauthProtectedResourceMetadataByServer() {
	metadata, err := object.GetOauthProtectedResourceMetadataByServer(c.Ctx.Request.Host, c.Ctx.Input.Param(":owner"), c.Ctx.Input.Param(":name"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if metadata == nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.ResponseError("server not found")
		return
	}

	c.Data["json"] = metadata
	c.ServeJSON()
}
