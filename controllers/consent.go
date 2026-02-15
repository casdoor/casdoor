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

package controllers

import (
	"encoding/json"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetConsents returns consent records for the current user
// @Title GetConsents
// @Tag Consent API
// @Description get all consent records for the current user
// @Param owner query string true "The owner"
// @Success 200 {array} object.ConsentRecord The Response object
// @router /get-consents [get]
func (c *ApiController) GetConsents() {
	owner := c.Ctx.Input.Query("owner")
	user := c.GetSessionUsername()

	if user == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	consents, err := object.GetConsentRecords(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Filter consents for the current user
	var userConsents []*object.ConsentRecord
	for _, consent := range consents {
		if consent.User == user {
			userConsents = append(userConsents, consent)
		}
	}

	c.ResponseOk(userConsents)
}

// RevokeConsent revokes a consent record
// @Title RevokeConsent
// @Tag Consent API
// @Description revoke a consent record
// @Param body body object.ConsentRecord true "The consent object"
// @Success 200 {object} controllers.Response The Response object
// @router /revoke-consent [post]
func (c *ApiController) RevokeConsent() {
	user := c.GetSessionUsername()
	if user == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	var consent object.ConsentRecord
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &consent)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Verify that the consent belongs to the current user
	existingConsent, err := object.GetConsentRecord(util.GetId(consent.Owner, consent.Name))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if existingConsent == nil {
		c.ResponseError(c.T("general:The consent does not exist"))
		return
	}

	if existingConsent.User != user {
		c.ResponseError(c.T("general:Unauthorized operation"))
		return
	}

	success, err := object.DeleteConsentRecord(&consent)
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
	user := c.GetSessionUsername()
	if user == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	var request struct {
		Owner        string   `json:"owner"`
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

	// Create and save consent record
	consent := object.ConsentRecord{
		Owner:          request.Owner,
		Name:           util.GenerateId(),
		CreatedTime:    util.GetCurrentTime(),
		User:           user,
		Application:    request.Application,
		GrantedScopes:  request.Scopes,
		ConsentTime:    util.GetCurrentTime(),
		ExpirationTime: "",
	}

	// Check if consent already exists for this user and application
	existingConsent, err := object.GetUserConsentForApplication(user, request.Application)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if existingConsent != nil {
		// Update existing consent
		existingConsent.GrantedScopes = request.Scopes
		existingConsent.ConsentTime = util.GetCurrentTime()
		_, err := object.UpdateConsentRecord(util.GetId(existingConsent.Owner, existingConsent.Name), existingConsent)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else {
		// Add new consent
		_, err = object.AddConsentRecord(&consent)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	// Now get the OAuth code
	// Get the user object to build proper userId
	username := c.GetSessionUsername()
	// The username from session is already in "owner/name" format
	userObj, err := object.GetUser(username)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if userObj == nil {
		c.ResponseError(c.T("general:User not found"))
		return
	}

	userId := userObj.GetId()
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

// CheckConsentRequired checks if consent is required for the OAuth flow
// @Title CheckConsentRequired
// @Tag Consent API
// @Description check if consent is required for the OAuth flow
// @Param clientId query string true "The client ID"
// @Param scope query string true "The requested scopes"
// @Success 200 {object} controllers.Response The Response object
// @router /check-consent-required [get]
func (c *ApiController) CheckConsentRequired() {
	clientId := c.Ctx.Input.Query("clientId")
	scopeStr := c.Ctx.Input.Query("scope")
	user := c.GetSessionUsername()

	if user == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if application == nil {
		c.ResponseError(c.T("general:Invalid client ID"))
		return
	}

	// Parse requested scopes
	requestedScopes := object.ParseScopes(scopeStr)

	// Check consent policy
	consentPolicy := application.ConsentPolicy
	if consentPolicy == "" || consentPolicy == "skip" {
		// No consent required
		c.ResponseOk(map[string]interface{}{
			"required": false,
		})
		return
	}

	if consentPolicy == "always" {
		// Always require consent
		c.ResponseOk(map[string]interface{}{
			"required":        true,
			"application":     application,
			"requestedScopes": object.GetScopeDescriptions(requestedScopes),
		})
		return
	}

	// Policy is "once" - check if consent already granted
	existingConsent, err := object.GetUserConsentForApplication(user, application.Name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if existingConsent != nil && object.ContainsAllScopes(existingConsent.GrantedScopes, requestedScopes) {
		// Consent already granted for all requested scopes
		c.ResponseOk(map[string]interface{}{
			"required": false,
		})
		return
	}

	// Consent required
	c.ResponseOk(map[string]interface{}{
		"required":        true,
		"application":     application,
		"requestedScopes": object.GetScopeDescriptions(requestedScopes),
	})
}
