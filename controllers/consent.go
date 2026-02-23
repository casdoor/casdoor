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
	"encoding/json"

	"github.com/casdoor/casdoor/object"
)

// RevokeConsent revokes a consent record
// @Title RevokeConsent
// @Tag Consent API
// @Description revoke a consent record
// @Param body body object.ConsentRecord true "The consent object"
// @Success 200 {object} controllers.Response The Response object
// @router /revoke-consent [post]
func (c *ApiController) RevokeConsent() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	var consent object.ConsentRecord
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &consent)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Validate that consent.Application is not empty
	if consent.Application == "" {
		c.ResponseError(c.T("general:Application cannot be empty"))
		return
	}

	// Validate that GrantedScopes is not empty when scope-specific revoke is requested
	if len(consent.GrantedScopes) == 0 {
		c.ResponseError(c.T("general:Granted scopes cannot be empty"))
		return
	}

	userObj, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if userObj == nil {
		c.ResponseError(c.T("general:The user doesn't exist"))
		return
	}

	newScopes := []object.ConsentRecord{}
	for _, record := range userObj.ApplicationScopes {
		if record.Application != consent.Application {
			// skip other applications
			newScopes = append(newScopes, record)
			continue
		}
		// revoke specified scopes
		revokeSet := make(map[string]bool)
		for _, s := range consent.GrantedScopes {
			revokeSet[s] = true
		}
		remaining := []string{}
		for _, s := range record.GrantedScopes {
			if !revokeSet[s] {
				remaining = append(remaining, s)
			}
		}
		if len(remaining) > 0 {
			// still have remaining scopes, keep the record and update
			record.GrantedScopes = remaining
			newScopes = append(newScopes, record)
		}
		// otherwise the application authorization is revoked, delete the whole record
	}
	userObj.ApplicationScopes = newScopes
	success, err := object.UpdateUser(userObj.GetId(), userObj, nil, false)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(success)
}

// GrantConsent grants consent for an OAuth application and returns authorization code
// @Title GrantConsent
// @Tag Consent API
// @Description grant consent for an OAuth application and get authorization code
// @Param body body object.ConsentRecord true "The consent object with OAuth parameters"
// @Success 200 {object} controllers.Response The Response object
// @router /grant-consent [post]
func (c *ApiController) GrantConsent() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	var request struct {
		Application  string   `json:"application"`
		Scopes       []string `json:"grantedScopes"`
		ClientId     string   `json:"clientId"`
		Provider     string   `json:"provider"`
		SigninMethod string   `json:"signinMethod"`
		ResponseType string   `json:"responseType"`
		RedirectUri  string   `json:"redirectUri"`
		Scope        string   `json:"scope"`
		State        string   `json:"state"`
		Nonce        string   `json:"nonce"`
		Challenge    string   `json:"challenge"`
		Resource     string   `json:"resource"`
	}

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Validate application by clientId
	application, err := object.GetApplicationByClientId(request.ClientId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(c.T("general:Invalid client_id"))
		return
	}

	// Verify that request.Application matches the application's actual ID
	if request.Application != application.GetId() {
		c.ResponseError(c.T("general:Invalid application"))
		return
	}

	// Update user's ApplicationScopes
	userObj, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if userObj == nil {
		c.ResponseError(c.T("general:User not found"))
		return
	}

	appId := application.GetId()
	found := false
	// Insert new scope into existing applicationScopes
	for i, record := range userObj.ApplicationScopes {
		if record.Application == appId {
			existing := make(map[string]bool)
			for _, s := range userObj.ApplicationScopes[i].GrantedScopes {
				existing[s] = true
			}
			for _, s := range request.Scopes {
				if !existing[s] {
					userObj.ApplicationScopes[i].GrantedScopes = append(userObj.ApplicationScopes[i].GrantedScopes, s)
					existing[s] = true
				}
			}
			found = true
			break
		}
	}
	// create a new applicationScopes if not found
	if !found {
		uniqueScopes := []string{}
		existing := make(map[string]bool)
		for _, s := range request.Scopes {
			if !existing[s] {
				uniqueScopes = append(uniqueScopes, s)
				existing[s] = true
			}
		}
		userObj.ApplicationScopes = append(userObj.ApplicationScopes, object.ConsentRecord{
			Application:   appId,
			GrantedScopes: uniqueScopes,
		})
	}

	_, err = object.UpdateUser(userObj.GetId(), userObj, []string{"application_scopes"}, false)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Now get the OAuth code
	code, err := object.GetOAuthCode(
		userId,
		request.ClientId,
		request.Provider,
		request.SigninMethod,
		request.ResponseType,
		request.RedirectUri,
		request.Scope,
		request.State,
		request.Nonce,
		request.Challenge,
		request.Resource,
		c.Ctx.Request.Host,
		c.GetAcceptLanguage(),
	)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(code.Code)
}
