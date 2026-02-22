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
	"strings"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// SessionData represents session metadata
type SessionData struct {
	ExpireTime int64
}

// GetSessionUsername returns the username from session or ctx
func (c *McpController) GetSessionUsername() string {
	// prefer username stored in Beego context by ApiFilter
	if ctxUser := c.Ctx.Input.GetData("currentUserId"); ctxUser != nil {
		if username, ok := ctxUser.(string); ok {
			return username
		}
	}

	// fallback to previous session-based logic with expiry check
	sessionData := c.GetSessionData()
	if sessionData != nil &&
		sessionData.ExpireTime != 0 &&
		sessionData.ExpireTime < time.Now().Unix() {
		c.ClearUserSession()
		return ""
	}

	user := c.GetSession("username")
	if user == nil {
		return ""
	}

	return user.(string)
}

// GetSessionData retrieves session data
func (c *McpController) GetSessionData() *SessionData {
	session := c.GetSession("SessionData")
	if session == nil {
		return nil
	}

	sessionData := &SessionData{}
	err := util.JsonToStruct(session.(string), sessionData)
	if err != nil {
		return nil
	}

	return sessionData
}

// ClearUserSession clears the user session
func (c *McpController) ClearUserSession() {
	c.SetSession("username", "")
	c.DelSession("SessionData")
	_ = c.SessionRegenerateID()
}

// IsGlobalAdmin checks if the current user is a global admin
func (c *McpController) IsGlobalAdmin() bool {
	isGlobalAdmin, _ := c.isGlobalAdmin()
	return isGlobalAdmin
}

func (c *McpController) isGlobalAdmin() (bool, *object.User) {
	username := c.GetSessionUsername()

	if object.IsAppUser(username) {
		// e.g., "app/app-casnode"
		return true, nil
	}

	user := c.getCurrentUser()
	if user == nil {
		return false, nil
	}

	return user.IsGlobalAdmin(), user
}

func (c *McpController) getCurrentUser() *object.User {
	var user *object.User
	var err error
	userId := c.GetSessionUsername()
	if userId == "" {
		user = nil
	} else {
		user, err = object.GetUser(userId)
		if err != nil {
			return nil
		}
	}
	return user
}

// GetAcceptLanguage returns the Accept-Language header value
func (c *McpController) GetAcceptLanguage() string {
	language := c.Ctx.Request.Header.Get("Accept-Language")
	if len(language) > 2 {
		language = language[0:2]
	}
	return language
}

// GetTokenFromRequest extracts the Bearer token from the Authorization header
func (c *McpController) GetTokenFromRequest() string {
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Extract Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return parts[1]
}

// GetClaimsFromToken parses and validates the JWT token and returns the claims
// Returns nil if no token is present or if token is invalid
func (c *McpController) GetClaimsFromToken() *object.Claims {
	tokenString := c.GetTokenFromRequest()
	if tokenString == "" {
		return nil
	}

	// Try to find the application for this token
	// For MCP, we'll try to parse using the first available application's certificate
	// In a production scenario, you might want to use a specific MCP application
	token, err := object.GetTokenByAccessToken(tokenString)
	if err != nil || token == nil {
		return nil
	}

	application, err := object.GetApplication(token.Application)
	if err != nil || application == nil {
		return nil
	}

	claims, err := object.ParseJwtTokenByApplication(tokenString, application)
	if err != nil {
		return nil
	}

	return claims
}

// GetScopesFromClaims extracts the scopes from JWT claims and returns them as a slice
func GetScopesFromClaims(claims *object.Claims) []string {
	if claims == nil || claims.Scope == "" {
		return []string{}
	}

	// Scopes are space-separated in OAuth 2.0
	return strings.Split(claims.Scope, " ")
}
