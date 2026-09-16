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
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type MagicLinkForm struct {
	Email        string `json:"email"`
	Organization string `json:"organization"`
	Application  string `json:"application"`
}

// SendMagicLink
// @Title SendMagicLink
// @Tag Login API
// @Description send a one-time sign-in link to an email address
// @Param clientId query string false "OAuth client ID"
// @Param responseType query string false "OAuth response type"
// @Param redirectUri query string false "OAuth redirect URI"
// @Param scope query string false "OAuth scope"
// @Param state query string false "OAuth state"
// @Param nonce query string false "OAuth nonce"
// @Param code_challenge_method query string false "OAuth PKCE code challenge method"
// @Param code_challenge query string false "OAuth PKCE code challenge"
// @Param body body controllers.MagicLinkForm true "The magic link request"
// @Success 200 {object} controllers.Response The Response object
// @router /send-magic-link [post]
func (c *ApiController) SendMagicLink() {
	var magicLinkForm MagicLinkForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &magicLinkForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	magicLinkForm.Email = strings.ToLower(strings.TrimSpace(magicLinkForm.Email))
	magicLinkForm.Organization = strings.TrimSpace(magicLinkForm.Organization)
	magicLinkForm.Application = strings.TrimSpace(magicLinkForm.Application)

	if magicLinkForm.Email == "" || magicLinkForm.Organization == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}
	if !util.IsEmailValid(magicLinkForm.Email) {
		c.ResponseError(c.T("check:Email is invalid"))
		return
	}

	application, err := c.getMagicLinkApplication(&magicLinkForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application.Organization != magicLinkForm.Organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	if !application.IsMagicLinkEnabled() {
		c.ResponseError(c.T("auth:The login method: magic link is not enabled for the application"))
		return
	}

	organization, err := object.GetOrganization(util.GetId(application.Owner, application.Organization))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if organization == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The organization: %s does not exist"), application.Organization))
		return
	}

	oauth := c.getMagicLinkOAuthParams()
	if oauth["responseType"] != "" && oauth["responseType"] != ResponseTypeLogin {
		msg, oauthApplication, err := object.CheckOAuthLogin(oauth["clientId"], oauth["responseType"], oauth["redirectUri"], oauth["scope"], oauth["state"], c.GetAcceptLanguage())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if msg != "" {
			c.ResponseError(msg)
			return
		}
		if oauthApplication == nil || oauthApplication.GetId() != application.GetId() {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
	}

	clientIp := util.GetClientIpFromRequest(c.Ctx.Request)
	err = object.CheckMagicLinkRateLimit(application, magicLinkForm.Email, clientIp)
	if err != nil {
		util.LogWarning(c.Ctx, "Magic link rate limited, organization = %s, application = %s, email = %s, error = %s", application.Organization, application.Name, magicLinkForm.Email, err.Error())
		c.ResponseError(err.Error())
		return
	}

	expireAt := object.ResolveMagicLinkExpireTime(application, time.Now())

	// Whether the address has an account, whether the application takes new users
	// and whether the mail could be sent are all answered the same way, so the
	// endpoint cannot be used to find out who has an account.
	provider, err := application.GetEmailProvider(object.MagicLinkProviderRule)
	if err != nil || provider == nil {
		errText := fmt.Sprintf(c.T("verification:please add an Email provider to the \"Providers\" list for the application: %s"), application.Name)
		if err != nil {
			errText = err.Error()
		}
		c.saveFailedMagicLink(application, &magicLinkForm, clientIp, oauth, expireAt, errText)
		util.LogWarning(c.Ctx, "Magic link email provider is not available, organization = %s, application = %s, email = %s, error = %s", application.Organization, application.Name, magicLinkForm.Email, errText)
		c.responseMagicLinkAccepted(expireAt)
		return
	}

	user, err := object.GetUserByEmail(application.Organization, magicLinkForm.Email)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user != nil && (user.IsDeleted || user.IsForbidden) {
		user = nil
	}
	if user == nil && !application.IsMagicLinkSignupEnabled() {
		errText := "the magic link user does not exist"
		c.saveFailedMagicLink(application, &magicLinkForm, clientIp, oauth, expireAt, errText)
		util.LogInfo(c.Ctx, "Magic link request accepted without sending, organization = %s, application = %s, email = %s, reason = %s", application.Organization, application.Name, magicLinkForm.Email, errText)
		c.responseMagicLinkAccepted(expireAt)
		return
	}

	token, err := object.GenerateMagicLinkToken()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	magicLink := object.NewMagicLink(application, magicLinkForm.Email, clientIp, c.GetSessionUsername(), token, oauth, expireAt)
	magicLink.AuthAction = getMagicLinkAuthAction(user == nil)
	_, err = object.AddMagicLink(magicLink)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	magicLinkUrl := object.BuildMagicLinkUrl(magicLink, token, c.Ctx.Request.Host)
	err = object.SendMagicLinkEmail(organization, provider, magicLinkForm.Email, magicLinkUrl, magicLink.ExpireTime)
	if err != nil {
		_ = object.UpdateMagicLinkStatus(magicLink, object.MagicLinkStatusFailed, err.Error())
		util.LogWarning(c.Ctx, "Magic link email sending failed, organization = %s, application = %s, email = %s, magicLink = %s, error = %s", application.Organization, application.Name, magicLinkForm.Email, util.GetId(magicLink.Owner, magicLink.Name), err.Error())
		c.responseMagicLinkAccepted(expireAt)
		return
	}

	err = object.UpdateMagicLinkStatus(magicLink, object.MagicLinkStatusSent, "")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.responseMagicLinkAccepted(expireAt)
}

// VerifyMagicLink
// @Title VerifyMagicLink
// @Tag Login API
// @Description exchange a one-time sign-in link for a session, a code or a token
// @Param token query string true "The magic link token"
// @Param clientId query string false "OAuth client ID"
// @Param responseType query string false "OAuth response type"
// @Param redirectUri query string false "OAuth redirect URI"
// @Param scope query string false "OAuth scope"
// @Param state query string false "OAuth state"
// @Param nonce query string false "OAuth nonce"
// @Param code_challenge_method query string false "OAuth PKCE code challenge method"
// @Param code_challenge query string false "OAuth PKCE code challenge"
// @Success 200 {object} controllers.Response The Response object
// @router /verify-magic-link [get]
func (c *ApiController) VerifyMagicLink() {
	token := c.Ctx.Input.Query("token")
	if token == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	magicLink, err := object.ConsumeMagicLink(token)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	application, err := object.GetApplication(magicLink.Application)
	if err != nil {
		c.failMagicLink(magicLink, err.Error())
		c.ResponseError(err.Error())
		return
	}
	if application == nil || application.Organization != magicLink.Owner {
		c.failMagicLink(magicLink, "the magic link application does not match")
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	if !application.IsMagicLinkEnabled() {
		c.failMagicLink(magicLink, "magic link sign-in is disabled for the application")
		c.ResponseError(c.T("auth:The login method: magic link is not enabled for the application"))
		return
	}

	err = c.checkMagicLinkOAuthParams(magicLink)
	if err != nil {
		c.failMagicLink(magicLink, err.Error())
		c.ResponseError(err.Error())
		return
	}

	isNewUser := false
	user, err := object.GetUserByEmail(magicLink.Owner, magicLink.Email)
	if err != nil {
		c.failMagicLink(magicLink, err.Error())
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		if !application.IsMagicLinkSignupEnabled() {
			c.failMagicLink(magicLink, "the magic link user does not exist")
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}

		user, err = c.addMagicLinkUser(application, magicLink)
		if err != nil {
			c.failMagicLink(magicLink, err.Error())
			c.ResponseError(err.Error())
			return
		}

		isNewUser = true
		c.Ctx.Input.SetParam("recordUserId", user.GetId())
		c.Ctx.Input.SetParam("recordSignup", "true")
	}
	if user.IsDeleted || user.IsForbidden {
		c.failMagicLink(magicLink, "the magic link user is not available")
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	if !isNewUser {
		ok, err := object.CheckLoginPermission(user.GetId(), application)
		if err != nil {
			c.failMagicLink(magicLink, err.Error())
			c.ResponseError(err.Error())
			return
		}
		if !ok {
			c.failMagicLink(magicLink, "the magic link user is not allowed to sign in to the application")
			c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s has disabled users to signin"), application.Name))
			return
		}
	}

	authForm := form.AuthForm{
		Type:         magicLink.ResponseType,
		SigninMethod: "Magic link",
		Application:  application.Name,
		Organization: magicLink.Owner,
		AutoSignin:   true,
	}
	if authForm.Type == "" {
		authForm.Type = ResponseTypeLogin
	}

	resp := c.HandleLoggedIn(application, user, &authForm)
	if resp == nil {
		// HandleLoggedIn() has already written its own response (MFA, for one)
		_ = object.UpdateMagicLinkStatus(magicLink, object.MagicLinkStatusFailed, "magic link sign-in did not complete")
		return
	}
	if resp.Status != "ok" {
		_ = object.UpdateMagicLinkStatus(magicLink, object.MagicLinkStatusFailed, resp.Msg)
	} else {
		_ = object.UpdateMagicLinkStatus(magicLink, object.MagicLinkStatusUsed, "")
	}

	// the consent page needs to know which application it is consenting to
	if _, ok := resp.Data.(map[string]bool); ok {
		resp.Data2 = application.Name
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

// GetMagicLinks
// @Title GetMagicLinks
// @Tag Magic Link API
// @Description get magic links
// @Param owner query string false "The owner of the magic links"
// @Param p query string false "The page number"
// @Param pageSize query string false "The page size"
// @Param field query string false "The search field"
// @Param value query string false "The search value"
// @Param sortField query string false "The sort field"
// @Param sortOrder query string false "The sort order"
// @Success 200 {array} object.MagicLink The Response object
// @router /get-magic-links [get]
func (c *ApiController) GetMagicLinks() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")
	owner := c.Ctx.Input.Query("owner")
	if c.IsGlobalAdmin() && owner != "" {
		organization = owner
	}

	if limit == "" || page == "" {
		magicLinks, err := object.GetMagicLinks(organization)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(magicLinks)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetMagicLinkCount(organization, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		magicLinks, err := object.GetPaginationMagicLinks(organization, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(magicLinks, paginator.Nums())
	}
}

// RevokeMagicLink
// @Title RevokeMagicLink
// @Tag Magic Link API
// @Description revoke a magic link that has not been used yet
// @Param id query string true "The id ( owner/name ) of the magic link"
// @Success 200 {object} controllers.Response The Response object
// @router /revoke-magic-link [post]
func (c *ApiController) RevokeMagicLink() {
	id, ok := c.getMagicLinkIdForAdmin()
	if !ok {
		return
	}

	c.Data["json"] = wrapActionResponse(object.RevokeMagicLink(id))
	c.ServeJSON()
}

// DeleteMagicLink
// @Title DeleteMagicLink
// @Tag Magic Link API
// @Description delete a magic link
// @Param id query string true "The id ( owner/name ) of the magic link"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-magic-link [post]
func (c *ApiController) DeleteMagicLink() {
	id, ok := c.getMagicLinkIdForAdmin()
	if !ok {
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteMagicLink(id))
	c.ServeJSON()
}

func (c *ApiController) getMagicLinkIdForAdmin() (string, bool) {
	organization, ok := c.RequireAdmin()
	if !ok {
		return "", false
	}

	id := c.Ctx.Input.Query("id")
	if id == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return "", false
	}

	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if !c.IsGlobalAdmin() && owner != organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return "", false
	}

	return id, true
}

func (c *ApiController) getMagicLinkApplication(magicLinkForm *MagicLinkForm) (*object.Application, error) {
	if magicLinkForm.Application == "" {
		application, err := object.GetApplicationByOrganizationName(magicLinkForm.Organization)
		if err != nil {
			return nil, err
		}
		if application == nil {
			return nil, fmt.Errorf(c.T("auth:The application: %s does not exist"), magicLinkForm.Organization)
		}

		return application, nil
	}

	applicationId := magicLinkForm.Application
	if !strings.Contains(applicationId, "/") {
		applicationId = util.GetId("admin", applicationId)
	}

	application, err := object.GetApplication(applicationId)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, fmt.Errorf(c.T("auth:The application: %s does not exist"), magicLinkForm.Application)
	}

	return application, nil
}

func (c *ApiController) getMagicLinkOAuthParams() map[string]string {
	return map[string]string{
		"clientId":            c.Ctx.Input.Query("clientId"),
		"responseType":        c.Ctx.Input.Query("responseType"),
		"redirectUri":         c.Ctx.Input.Query("redirectUri"),
		"scope":               c.Ctx.Input.Query("scope"),
		"state":               c.Ctx.Input.Query("state"),
		"nonce":               c.Ctx.Input.Query("nonce"),
		"codeChallengeMethod": c.Ctx.Input.Query("code_challenge_method"),
		"codeChallenge":       c.Ctx.Input.Query("code_challenge"),
	}
}

// checkMagicLinkOAuthParams makes sure the OAuth request the link finishes is
// the one it was issued for, so a link cannot be replayed against another
// redirect URI.
func (c *ApiController) checkMagicLinkOAuthParams(magicLink *object.MagicLink) error {
	if magicLink.ResponseType == "" || magicLink.ResponseType == ResponseTypeLogin {
		return nil
	}

	expected := map[string]string{
		"clientId":              magicLink.ClientId,
		"responseType":          magicLink.ResponseType,
		"redirectUri":           magicLink.RedirectUri,
		"scope":                 magicLink.Scope,
		"state":                 magicLink.State,
		"nonce":                 magicLink.Nonce,
		"code_challenge_method": magicLink.CodeChallengeMethod,
		"code_challenge":        magicLink.CodeChallenge,
	}

	actual := map[string]string{}
	for key := range expected {
		actual[key] = c.Ctx.Input.Query(key)
	}

	return checkMagicLinkOAuthPayload(expected, actual)
}

func checkMagicLinkOAuthPayload(expected map[string]string, actual map[string]string) error {
	for key, value := range expected {
		if actual[key] != value {
			return fmt.Errorf("the magic link OAuth parameters do not match")
		}
	}

	return nil
}

func (c *ApiController) addMagicLinkUser(application *object.Application, magicLink *object.MagicLink) (*object.User, error) {
	organization, err := object.GetOrganization(util.GetId(application.Owner, application.Organization))
	if err != nil {
		return nil, err
	}
	if organization == nil {
		return nil, fmt.Errorf(c.T("auth:The organization: %s does not exist"), application.Organization)
	}

	id, err := object.GenerateIdForNewUser(application)
	if err != nil {
		return nil, err
	}

	initScore, err := organization.GetInitScore()
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(magicLink.Email))
	username := id
	if organization.UseEmailAsUsername {
		username = email
	}

	user := &object.User{
		Owner:             application.Organization,
		Name:              username,
		CreatedTime:       util.GetCurrentTime(),
		Id:                id,
		Type:              "normal-user",
		Password:          util.GenerateId(),
		DisplayName:       email,
		Avatar:            organization.DefaultAvatar,
		Email:             email,
		Address:           []string{},
		Score:             initScore,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
		EmailVerified:     true,
		RegisterType:      "Magic link",
		RegisterSource:    util.GetId(application.Organization, application.Name),
	}
	if application.DefaultGroup != "" {
		user.Groups = []string{application.DefaultGroup}
	}
	if application.DefaultTag != "" {
		user.Tag = application.DefaultTag
	}

	affected, err := object.AddUser(user, c.GetAcceptLanguage())
	if err != nil {
		return nil, err
	}
	if !affected {
		return nil, fmt.Errorf(c.T("auth:Failed to create user, user information is invalid: %s"), user.GetId())
	}

	err = object.AddUserToOriginalDatabase(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *ApiController) saveFailedMagicLink(application *object.Application, magicLinkForm *MagicLinkForm, clientIp string, oauth map[string]string, expireAt time.Time, lastError string) {
	magicLink := object.NewMagicLink(application, magicLinkForm.Email, clientIp, c.GetSessionUsername(), "", oauth, expireAt)
	magicLink.Status = object.MagicLinkStatusFailed
	magicLink.LastError = lastError
	_, _ = object.AddMagicLink(magicLink)
}

func (c *ApiController) failMagicLink(magicLink *object.MagicLink, lastError string) {
	util.LogWarning(c.Ctx, "Magic link verification failed, organization = %s, magicLink = %s, error = %s", magicLink.Owner, util.GetId(magicLink.Owner, magicLink.Name), lastError)
	_ = object.UpdateMagicLinkStatus(magicLink, object.MagicLinkStatusFailed, lastError)
}

// responseMagicLinkAccepted is the one answer every accepted request gets, it
// tells nothing about whether the address has an account.
func (c *ApiController) responseMagicLinkAccepted(expireAt time.Time) {
	c.ResponseOk(map[string]string{"expireTime": util.Time2String(expireAt)})
}

func getMagicLinkAuthAction(isNewUser bool) string {
	if isNewUser {
		return object.MagicLinkAuthActionSignupNewUser
	}

	return object.MagicLinkAuthActionSigninExistingUser
}
