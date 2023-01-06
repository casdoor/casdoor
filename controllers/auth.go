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
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/captcha"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/casdoor/util"
	"github.com/google/uuid"
)

var (
	wechatScanType string
	lock           sync.RWMutex
)

func codeToResponse(code *object.Code) *Response {
	if code.Code == "" {
		return &Response{Status: "error", Msg: code.Message, Data: code.Code}
	}

	return &Response{Status: "ok", Msg: "", Data: code.Code}
}

func tokenToResponse(token *object.Token) *Response {
	if token.AccessToken == "" {
		return &Response{Status: "error", Msg: "fail to get accessToken", Data: token.AccessToken}
	}
	return &Response{Status: "ok", Msg: "", Data: token.AccessToken}
}

// HandleLoggedIn ...
func (c *ApiController) HandleLoggedIn(application *object.Application, user *object.User, form *RequestForm) (resp *Response) {
	userId := user.GetId()

	allowed, err := object.CheckAccessPermission(userId, application)
	if err != nil {
		c.ResponseError(err.Error(), nil)
		return
	}
	if !allowed {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	if form.Type == ResponseTypeLogin {
		c.SetSessionUsername(userId)
		util.LogInfo(c.Ctx, "API: [%s] signed in", userId)
		resp = &Response{Status: "ok", Msg: "", Data: userId}
	} else if form.Type == ResponseTypeCode {
		clientId := c.Input().Get("clientId")
		responseType := c.Input().Get("responseType")
		redirectUri := c.Input().Get("redirectUri")
		scope := c.Input().Get("scope")
		state := c.Input().Get("state")
		nonce := c.Input().Get("nonce")
		challengeMethod := c.Input().Get("code_challenge_method")
		codeChallenge := c.Input().Get("code_challenge")

		if challengeMethod != "S256" && challengeMethod != "null" && challengeMethod != "" {
			c.ResponseError(c.T("auth:Challenge method should be S256"))
			return
		}
		code := object.GetOAuthCode(userId, clientId, responseType, redirectUri, scope, state, nonce, codeChallenge, c.Ctx.Request.Host, c.GetAcceptLanguage())
		resp = codeToResponse(code)

		if application.EnableSigninSession || application.HasPromptPage() {
			// The prompt page needs the user to be signed in
			c.SetSessionUsername(userId)
		}
	} else if form.Type == ResponseTypeToken || form.Type == ResponseTypeIdToken { // implicit flow
		if !object.IsGrantTypeValid(form.Type, application.GrantTypes) {
			resp = &Response{Status: "error", Msg: fmt.Sprintf("error: grant_type: %s is not supported in this application", form.Type), Data: ""}
		} else {
			scope := c.Input().Get("scope")
			token, _ := object.GetTokenByUser(application, user, scope, c.Ctx.Request.Host)
			resp = tokenToResponse(token)
		}
	} else if form.Type == ResponseTypeSaml { // saml flow
		res, redirectUrl, err := object.GetSamlResponse(application, user, form.SamlRequest, c.Ctx.Request.Host)
		if err != nil {
			c.ResponseError(err.Error(), nil)
			return
		}
		resp = &Response{Status: "ok", Msg: "", Data: res, Data2: redirectUrl}
	} else if form.Type == ResponseTypeCas {
		// not oauth but CAS SSO protocol
		service := c.Input().Get("service")
		resp = wrapErrorResponse(nil)
		if service != "" {
			st, err := object.GenerateCasToken(userId, service)
			if err != nil {
				resp = wrapErrorResponse(err)
			} else {
				resp.Data = st
			}
		}
		if application.EnableSigninSession || application.HasPromptPage() {
			// The prompt page needs the user to be signed in
			c.SetSessionUsername(userId)
		}

	} else {
		resp = wrapErrorResponse(fmt.Errorf("unknown response type: %s", form.Type))
	}

	// if user did not check auto signin
	if resp.Status == "ok" && !form.AutoSignin {
		timestamp := time.Now().Unix()
		timestamp += 3600 * 24
		c.SetSessionData(&SessionData{
			ExpireTime: timestamp,
		})
	}

	return resp
}

// GetApplicationLogin ...
// @Title GetApplicationLogin
// @Tag Login API
// @Description get application login
// @Param   clientId    query    string  true        "client id"
// @Param   responseType    query    string  true        "response type"
// @Param   redirectUri    query    string  true        "redirect uri"
// @Param   scope    query    string  true        "scope"
// @Param   state    query    string  true        "state"
// @Success 200 {object}  Response The Response object
// @router /get-app-login [get]
func (c *ApiController) GetApplicationLogin() {
	clientId := c.Input().Get("clientId")
	responseType := c.Input().Get("responseType")
	redirectUri := c.Input().Get("redirectUri")
	scope := c.Input().Get("scope")
	state := c.Input().Get("state")

	msg, application := object.CheckOAuthLogin(clientId, responseType, redirectUri, scope, state, c.GetAcceptLanguage())
	application = object.GetMaskedApplication(application, "")
	if msg != "" {
		c.ResponseError(msg, application)
	} else {
		c.ResponseOk(application)
	}
}

func setHttpClient(idProvider idp.IdProvider, providerType string) {
	if providerType == "GitHub" || providerType == "Google" || providerType == "Facebook" || providerType == "LinkedIn" || providerType == "Steam" || providerType == "Line" {
		idProvider.SetHttpClient(proxy.ProxyHttpClient)
	} else {
		idProvider.SetHttpClient(proxy.DefaultHttpClient)
	}
}

// Login ...
// @Title Login
// @Tag Login API
// @Description login
// @Param clientId        query    string  true clientId
// @Param responseType    query    string  true responseType
// @Param redirectUri     query    string  true redirectUri
// @Param scope     query    string  false  scope
// @Param state     query    string  false  state
// @Param nonce     query    string  false nonce
// @Param code_challenge_method   query    string  false code_challenge_method
// @Param code_challenge          query    string  false code_challenge
// @Param   form   body   controllers.RequestForm  true        "Login information"
// @Success 200 {object} Response The Response object
// @router /login [post]
func (c *ApiController) Login() {
	resp := &Response{}

	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if form.Username != "" {
		if form.Type == ResponseTypeLogin {
			if c.GetSessionUsername() != "" {
				c.ResponseError(c.T("auth:Please sign out first before signing in"), c.GetSessionUsername())
				return
			}
		}

		var user *object.User
		var msg string

		if form.Password == "" {
			var verificationCodeType string
			var checkResult string

			if form.Name != "" {
				user = object.GetUserByFields(form.Organization, form.Name)
			}

			// check result through Email or Phone
			var checkDest string
			if strings.Contains(form.Username, "@") {
				verificationCodeType = "email"
				if user != nil && util.GetMaskedEmail(user.Email) == form.Username {
					form.Username = user.Email
				}
				checkDest = form.Username
			} else {
				verificationCodeType = "phone"
				if len(form.PhonePrefix) == 0 {
					responseText := fmt.Sprintf(c.T("auth:%s No phone prefix"), verificationCodeType)
					c.ResponseError(responseText)
					return
				}
				if user != nil && util.GetMaskedPhone(user.Phone) == form.Username {
					form.Username = user.Phone
				}
				checkDest = fmt.Sprintf("+%s%s", form.PhonePrefix, form.Username)
			}
			user = object.GetUserByFields(form.Organization, form.Username)
			if user == nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The user: %s/%s doesn't exist"), form.Organization, form.Username))
				return
			}
			checkResult = object.CheckSigninCode(user, checkDest, form.Code, c.GetAcceptLanguage())
			if len(checkResult) != 0 {
				responseText := fmt.Sprintf("%s - %s", verificationCodeType, checkResult)
				c.ResponseError(responseText)
				return
			}

			// disable the verification code
			if strings.Contains(form.Username, "@") {
				object.DisableVerificationCode(form.Username)
			} else {
				object.DisableVerificationCode(fmt.Sprintf("+%s%s", form.PhonePrefix, form.Username))
			}
		} else {
			application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
			if application == nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), form.Application))
				return
			}
			if !application.EnablePassword {
				c.ResponseError(c.T("auth:The login method: login with password is not enabled for the application"))
				return
			}

			if object.CheckToEnableCaptcha(application) {
				isHuman, err := captcha.VerifyCaptchaByCaptchaType(form.CaptchaType, form.CaptchaToken, form.ClientSecret)
				if err != nil {
					c.ResponseError(err.Error())
					return
				}

				if !isHuman {
					c.ResponseError(c.T("auth:Turing test failed."))
					return
				}
			}

			password := form.Password
			user, msg = object.CheckUserPassword(form.Organization, form.Username, password, c.GetAcceptLanguage())
		}

		if msg != "" {
			resp = &Response{Status: "error", Msg: msg}
		} else {
			application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
			if application == nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), form.Application))
				return
			}

			resp = c.HandleLoggedIn(application, user, &form)

			record := object.NewRecord(c.Ctx)
			record.Organization = application.Organization
			record.User = user.Name
			util.SafeGoroutine(func() { object.AddRecord(record) })
		}
	} else if form.Provider != "" {
		application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
		if application == nil {
			c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), form.Application))
			return
		}

		organization := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", application.Organization))
		provider := object.GetProvider(util.GetId("admin", form.Provider))
		providerItem := application.GetProviderItem(provider.Name)
		if !providerItem.IsProviderVisible() {
			c.ResponseError(fmt.Sprintf(c.T("auth:The provider: %s is not enabled for the application"), provider.Name))
			return
		}

		userInfo := &idp.UserInfo{}
		if provider.Category == "SAML" {
			// SAML
			userInfo.Id, err = object.ParseSamlResponse(form.SamlResponse, provider.Type)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		} else if provider.Category == "OAuth" {
			// OAuth

			clientId := provider.ClientId
			clientSecret := provider.ClientSecret
			if provider.Type == "WeChat" && strings.Contains(c.Ctx.Request.UserAgent(), "MicroMessenger") {
				clientId = provider.ClientId2
				clientSecret = provider.ClientSecret2
			}

			idProvider := idp.GetIdProvider(provider.Type, provider.SubType, clientId, clientSecret, provider.AppId, form.RedirectUri, provider.Domain, provider.CustomAuthUrl, provider.CustomTokenUrl, provider.CustomUserInfoUrl)
			if idProvider == nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The provider type: %s is not supported"), provider.Type))
				return
			}

			setHttpClient(idProvider, provider.Type)

			if form.State != conf.GetConfigString("authState") && form.State != application.Name {
				c.ResponseError(fmt.Sprintf(c.T("auth:State expected: %s, but got: %s"), conf.GetConfigString("authState"), form.State))
				return
			}

			// https://github.com/golang/oauth2/issues/123#issuecomment-103715338
			token, err := idProvider.GetToken(form.Code)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			if !token.Valid() {
				c.ResponseError(c.T("auth:Invalid token"))
				return
			}

			userInfo, err = idProvider.GetUserInfo(token)
			if err != nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:Failed to login in: %s"), err.Error()))
				return
			}
		}

		if form.Method == "signup" {
			user := &object.User{}
			if provider.Category == "SAML" {
				user = object.GetUser(fmt.Sprintf("%s/%s", application.Organization, userInfo.Id))
			} else if provider.Category == "OAuth" {
				user = object.GetUserByField(application.Organization, provider.Type, userInfo.Id)
			}

			if user != nil && !user.IsDeleted {
				// Sign in via OAuth (want to sign up but already have account)

				if user.IsForbidden {
					c.ResponseError(c.T("auth:The user is forbidden to sign in, please contact the administrator"))
				}

				resp = c.HandleLoggedIn(application, user, &form)

				record := object.NewRecord(c.Ctx)
				record.Organization = application.Organization
				record.User = user.Name
				util.SafeGoroutine(func() { object.AddRecord(record) })
			} else if provider.Category == "OAuth" {
				// Sign up via OAuth
				if !application.EnableSignUp {
					c.ResponseError(fmt.Sprintf(c.T("auth:The account for provider: %s and username: %s (%s) does not exist and is not allowed to sign up as new account, please contact your IT support"), provider.Type, userInfo.Username, userInfo.DisplayName))
					return
				}

				if !providerItem.CanSignUp {
					c.ResponseError(fmt.Sprintf(c.T("auth:The account for provider: %s and username: %s (%s) does not exist and is not allowed to sign up as new account via %%s, please use another way to sign up"), provider.Type, userInfo.Username, userInfo.DisplayName, provider.Type))
					return
				}

				// Handle username conflicts
				tmpUser := object.GetUser(fmt.Sprintf("%s/%s", application.Organization, userInfo.Username))
				if tmpUser != nil {
					uid, err := uuid.NewRandom()
					if err != nil {
						c.ResponseError(err.Error())
						return
					}

					uidStr := strings.Split(uid.String(), "-")
					userInfo.Username = fmt.Sprintf("%s_%s", userInfo.Username, uidStr[1])
				}

				properties := map[string]string{}
				properties["no"] = strconv.Itoa(len(object.GetUsers(application.Organization)) + 2)
				initScore, err := getInitScore(organization)
				if err != nil {
					c.ResponseError(fmt.Errorf(c.T("auth:Get init score failed, error: %w"), err).Error())
					return
				}

				user = &object.User{
					Owner:             application.Organization,
					Name:              userInfo.Username,
					CreatedTime:       util.GetCurrentTime(),
					Id:                util.GenerateId(),
					Type:              "normal-user",
					DisplayName:       userInfo.DisplayName,
					Avatar:            userInfo.AvatarUrl,
					Address:           []string{},
					Email:             userInfo.Email,
					Score:             initScore,
					IsAdmin:           false,
					IsGlobalAdmin:     false,
					IsForbidden:       false,
					IsDeleted:         false,
					SignupApplication: application.Name,
					Properties:        properties,
				}
				// sync info from 3rd-party if possible
				object.SetUserOAuthProperties(organization, user, provider.Type, userInfo)

				affected := object.AddUser(user)
				if !affected {
					c.ResponseError(fmt.Sprintf(c.T("auth:Failed to create user, user information is invalid: %s"), util.StructToJson(user)))
					return
				}

				object.LinkUserAccount(user, provider.Type, userInfo.Id)

				resp = c.HandleLoggedIn(application, user, &form)

				record := object.NewRecord(c.Ctx)
				record.Organization = application.Organization
				record.User = user.Name
				util.SafeGoroutine(func() { object.AddRecord(record) })

				record2 := object.NewRecord(c.Ctx)
				record2.Action = "signup"
				record2.Organization = application.Organization
				record2.User = user.Name
				util.SafeGoroutine(func() { object.AddRecord(record2) })
			} else if provider.Category == "SAML" {
				resp = &Response{Status: "error", Msg: "The account does not exist"}
			}
			// resp = &Response{Status: "ok", Msg: "", Data: res}
		} else { // form.Method != "signup"
			userId := c.GetSessionUsername()
			if userId == "" {
				c.ResponseError(c.T("auth:The account does not exist"), userInfo)
				return
			}

			oldUser := object.GetUserByField(application.Organization, provider.Type, userInfo.Id)
			if oldUser != nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The account for provider: %s and username: %s (%s) is already linked to another account: %s (%s)"), provider.Type, userInfo.Username, userInfo.DisplayName, oldUser.Name, oldUser.DisplayName))
				return
			}

			user := object.GetUser(userId)

			// sync info from 3rd-party if possible
			object.SetUserOAuthProperties(organization, user, provider.Type, userInfo)

			isLinked := object.LinkUserAccount(user, provider.Type, userInfo.Id)
			if isLinked {
				resp = &Response{Status: "ok", Msg: "", Data: isLinked}
			} else {
				resp = &Response{Status: "error", Msg: "Failed to link user account", Data: isLinked}
			}
		}
	} else {
		if c.GetSessionUsername() != "" {
			// user already signed in to Casdoor, so let the user click the avatar button to do the quick sign-in
			application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
			if application == nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), form.Application))
				return
			}

			user := c.getCurrentUser()
			resp = c.HandleLoggedIn(application, user, &form)

			record := object.NewRecord(c.Ctx)
			record.Organization = application.Organization
			record.User = user.Name
			util.SafeGoroutine(func() { object.AddRecord(record) })
		} else {
			c.ResponseError(fmt.Sprintf(c.T("auth:Unknown authentication type (not password or provider), form = %s"), util.StructToJson(form)))
			return
		}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) GetSamlLogin() {
	providerId := c.Input().Get("id")
	relayState := c.Input().Get("relayState")
	authURL, method, err := object.GenerateSamlLoginUrl(providerId, relayState, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
	}
	c.ResponseOk(authURL, method)
}

func (c *ApiController) HandleSamlLogin() {
	relayState := c.Input().Get("RelayState")
	samlResponse := c.Input().Get("SAMLResponse")
	decode, err := base64.StdEncoding.DecodeString(relayState)
	if err != nil {
		c.ResponseError(err.Error())
	}
	slice := strings.Split(string(decode), "&")
	relayState = url.QueryEscape(relayState)
	samlResponse = url.QueryEscape(samlResponse)
	targetUrl := fmt.Sprintf("%s?relayState=%s&samlResponse=%s",
		slice[4], relayState, samlResponse)
	c.Redirect(targetUrl, 303)
}

// HandleOfficialAccountEvent ...
// @Tag HandleOfficialAccountEvent API
// @Title HandleOfficialAccountEvent
// @router /api/webhook [POST]
func (c *ApiController) HandleOfficialAccountEvent() {
	respBytes, err := ioutil.ReadAll(c.Ctx.Request.Body)
	if err != nil {
		c.ResponseError(err.Error())
	}
	var data struct {
		MsgType  string `xml:"MsgType"`
		Event    string `xml:"Event"`
		EventKey string `xml:"EventKey"`
	}
	err = xml.Unmarshal(respBytes, &data)
	if err != nil {
		c.ResponseError(err.Error())
	}
	lock.Lock()
	defer lock.Unlock()
	if data.EventKey != "" {
		wechatScanType = data.Event
		c.Ctx.WriteString("")
	}
}

// GetWebhookEventType ...
// @Tag GetWebhookEventType API
// @Title GetWebhookEventType
// @router /api/get-webhook-event [GET]
func (c *ApiController) GetWebhookEventType() {
	lock.Lock()
	defer lock.Unlock()
	resp := &Response{
		Status: "ok",
		Msg:    "",
		Data:   wechatScanType,
	}
	c.Data["json"] = resp
	wechatScanType = ""
	c.ServeJSON()
}
