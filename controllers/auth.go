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

	"github.com/casdoor/casdoor/captcha"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/form"
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
	return &Response{Status: "ok", Msg: "", Data: token.AccessToken, Data2: token.RefreshToken}
}

// HandleLoggedIn ...
func (c *ApiController) HandleLoggedIn(application *object.Application, user *object.User, form *form.AuthForm) (resp *Response) {
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

	if form.Password != "" && user.IsMfaEnabled() {
		c.setMfaSessionData(&object.MfaSessionData{UserId: userId})
		resp = &Response{Status: object.NextMfa, Data: user.GetPreferredMfaProps(true)}
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
		code, err := object.GetOAuthCode(userId, clientId, responseType, redirectUri, scope, state, nonce, codeChallenge, c.Ctx.Request.Host, c.GetAcceptLanguage())
		if err != nil {
			c.ResponseError(err.Error(), nil)
			return
		}

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
		res, redirectUrl, method, err := object.GetSamlResponse(application, user, form.SamlRequest, c.Ctx.Request.Host)
		if err != nil {
			c.ResponseError(err.Error(), nil)
			return
		}
		resp = &Response{Status: "ok", Msg: "", Data: res, Data2: map[string]string{"redirectUrl": redirectUrl, "method": method}}
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
		c.setExpireForSession()
	}

	if resp.Status == "ok" {
		_, err = object.AddSession(&object.Session{
			Owner:       user.Owner,
			Name:        user.Name,
			Application: application.Name,
			SessionId:   []string{c.Ctx.Input.CruSession.SessionID()},
		})
		if err != nil {
			c.ResponseError(err.Error(), nil)
			return
		}
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

	msg, application, err := object.CheckOAuthLogin(clientId, responseType, redirectUri, scope, state, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	application = object.GetMaskedApplication(application, "")
	if msg != "" {
		c.ResponseError(msg, application)
	} else {
		c.ResponseOk(application)
	}
}

func setHttpClient(idProvider idp.IdProvider, providerType string) {
	if isProxyProviderType(providerType) {
		idProvider.SetHttpClient(proxy.ProxyHttpClient)
	} else {
		idProvider.SetHttpClient(proxy.DefaultHttpClient)
	}
}

func isProxyProviderType(providerType string) bool {
	providerTypes := []string{
		"GitHub",
		"Google",
		"Facebook",
		"LinkedIn",
		"Steam",
		"Line",
		"Amazon",
		"Instagram",
		"TikTok",
		"Twitter",
		"Uber",
		"Yahoo",
	}
	for _, v := range providerTypes {
		if strings.EqualFold(v, providerType) {
			return true
		}
	}
	return false
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
// @Param   form   body   controllers.AuthForm  true        "Login information"
// @Success 200 {object} Response The Response object
// @router /login [post]
func (c *ApiController) Login() {
	resp := &Response{}

	var authForm form.AuthForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &authForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if authForm.Username != "" {
		if authForm.Type == ResponseTypeLogin {
			if c.GetSessionUsername() != "" {
				c.ResponseError(c.T("account:Please sign out first"), c.GetSessionUsername())
				return
			}
		}

		var user *object.User
		var msg string

		if authForm.Password == "" {
			if user, err = object.GetUserByFields(authForm.Organization, authForm.Username); err != nil {
				c.ResponseError(err.Error(), nil)
				return
			} else if user == nil {
				c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(authForm.Organization, authForm.Username)))
				return
			}

			verificationCodeType := object.GetVerifyType(authForm.Username)
			var checkDest string
			if verificationCodeType == object.VerifyTypePhone {
				authForm.CountryCode = user.GetCountryCode(authForm.CountryCode)
				var ok bool
				if checkDest, ok = util.GetE164Number(authForm.Username, authForm.CountryCode); !ok {
					c.ResponseError(fmt.Sprintf(c.T("verification:Phone number is invalid in your region %s"), authForm.CountryCode))
					return
				}
			}

			// check result through Email or Phone
			checkResult := object.CheckSigninCode(user, checkDest, authForm.Code, c.GetAcceptLanguage())
			if len(checkResult) != 0 {
				c.ResponseError(fmt.Sprintf("%s - %s", verificationCodeType, checkResult))
				return
			}

			// disable the verification code
			err := object.DisableVerificationCode(checkDest)
			if err != nil {
				c.ResponseError(err.Error(), nil)
				return
			}
		} else {
			application, err := object.GetApplication(fmt.Sprintf("admin/%s", authForm.Application))
			if err != nil {
				c.ResponseError(err.Error(), nil)
				return
			}

			if application == nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), authForm.Application))
				return
			}
			if !application.EnablePassword {
				c.ResponseError(c.T("auth:The login method: login with password is not enabled for the application"))
				return
			}
			var enableCaptcha bool
			if enableCaptcha, err = object.CheckToEnableCaptcha(application, authForm.Organization, authForm.Username); err != nil {
				c.ResponseError(err.Error())
				return
			} else if enableCaptcha {
				isHuman, err := captcha.VerifyCaptchaByCaptchaType(authForm.CaptchaType, authForm.CaptchaToken, authForm.ClientSecret)
				if err != nil {
					c.ResponseError(err.Error())
					return
				}

				if !isHuman {
					c.ResponseError(c.T("verification:Turing test failed."))
					return
				}
			}

			password := authForm.Password
			user, msg = object.CheckUserPassword(authForm.Organization, authForm.Username, password, c.GetAcceptLanguage(), enableCaptcha)
		}

		if msg != "" {
			resp = &Response{Status: "error", Msg: msg}
		} else {
			application, err := object.GetApplication(fmt.Sprintf("admin/%s", authForm.Application))
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			if application == nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), authForm.Application))
				return
			}

			resp = c.HandleLoggedIn(application, user, &authForm)

			organization, err := object.GetOrganizationByUser(user)
			if err != nil {
				c.ResponseError(err.Error())
			}

			if user != nil && organization.HasRequiredMfa() && !user.IsMfaEnabled() {
				resp.Msg = object.RequiredMfa
			}

			record := object.NewRecord(c.Ctx)
			record.Organization = application.Organization
			record.User = user.Name
			util.SafeGoroutine(func() { object.AddRecord(record) })
		}
	} else if authForm.Provider != "" {
		var application *object.Application
		if authForm.ClientId != "" {
			application, err = object.GetApplicationByClientId(authForm.ClientId)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		} else {
			application, err = object.GetApplication(fmt.Sprintf("admin/%s", authForm.Application))
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}

		if application == nil {
			c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), authForm.Application))
			return
		}
		organization, err := object.GetOrganization(util.GetId("admin", application.Organization))
		if err != nil {
			c.ResponseError(c.T(err.Error()))
		}

		provider, err := object.GetProvider(util.GetId("admin", authForm.Provider))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		providerItem := application.GetProviderItem(provider.Name)
		if !providerItem.IsProviderVisible() {
			c.ResponseError(fmt.Sprintf(c.T("auth:The provider: %s is not enabled for the application"), provider.Name))
			return
		}

		userInfo := &idp.UserInfo{}
		if provider.Category == "SAML" {
			// SAML
			userInfo.Id, err = object.ParseSamlResponse(authForm.SamlResponse, provider, c.Ctx.Request.Host)
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

			idProvider := idp.GetIdProvider(provider.Type, provider.SubType, clientId, clientSecret, provider.AppId, authForm.RedirectUri, provider.Domain, provider.CustomAuthUrl, provider.CustomTokenUrl, provider.CustomUserInfoUrl)
			if idProvider == nil {
				c.ResponseError(fmt.Sprintf(c.T("storage:The provider type: %s is not supported"), provider.Type))
				return
			}

			setHttpClient(idProvider, provider.Type)

			if authForm.State != conf.GetConfigString("authState") && authForm.State != application.Name {
				c.ResponseError(fmt.Sprintf(c.T("auth:State expected: %s, but got: %s"), conf.GetConfigString("authState"), authForm.State))
				return
			}

			// https://github.com/golang/oauth2/issues/123#issuecomment-103715338
			token, err := idProvider.GetToken(authForm.Code)
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

		if authForm.Method == "signup" {
			user := &object.User{}
			if provider.Category == "SAML" {
				user, err = object.GetUser(util.GetId(application.Organization, userInfo.Id))
				if err != nil {
					c.ResponseError(err.Error())
					return
				}
			} else if provider.Category == "OAuth" {
				user, err = object.GetUserByField(application.Organization, provider.Type, userInfo.Id)
				if err != nil {
					c.ResponseError(err.Error())
					return
				}
			}

			if user != nil && !user.IsDeleted {
				// Sign in via OAuth (want to sign up but already have account)

				if user.IsForbidden {
					c.ResponseError(c.T("check:The user is forbidden to sign in, please contact the administrator"))
				}

				resp = c.HandleLoggedIn(application, user, &authForm)

				record := object.NewRecord(c.Ctx)
				record.Organization = application.Organization
				record.User = user.Name
				util.SafeGoroutine(func() { object.AddRecord(record) })
			} else if provider.Category == "OAuth" {
				// Sign up via OAuth
				if application.EnableLinkWithEmail {
					if userInfo.Email != "" {
						// Find existing user with Email
						user, err = object.GetUserByField(application.Organization, "email", userInfo.Email)
						if err != nil {
							c.ResponseError(err.Error())
							return
						}
					}

					if user == nil && userInfo.Phone != "" {
						// Find existing user with phone number
						user, err = object.GetUserByField(application.Organization, "phone", userInfo.Phone)
						if err != nil {
							c.ResponseError(err.Error())
							return
						}
					}
				}

				if user == nil || user.IsDeleted {
					if !application.EnableSignUp {
						c.ResponseError(fmt.Sprintf(c.T("auth:The account for provider: %s and username: %s (%s) does not exist and is not allowed to sign up as new account, please contact your IT support"), provider.Type, userInfo.Username, userInfo.DisplayName))
						return
					}

					if !providerItem.CanSignUp {
						c.ResponseError(fmt.Sprintf(c.T("auth:The account for provider: %s and username: %s (%s) does not exist and is not allowed to sign up as new account via %%s, please use another way to sign up"), provider.Type, userInfo.Username, userInfo.DisplayName, provider.Type))
						return
					}

					// Handle username conflicts
					tmpUser, err := object.GetUser(util.GetId(application.Organization, userInfo.Username))
					if err != nil {
						c.ResponseError(err.Error())
						return
					}

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
					count, err := object.GetUserCount(application.Organization, "", "", "")
					if err != nil {
						c.ResponseError(err.Error())
						return
					}

					properties["no"] = strconv.Itoa(int(count + 2))
					initScore, err := organization.GetInitScore()
					if err != nil {
						c.ResponseError(fmt.Errorf(c.T("account:Get init score failed, error: %w"), err).Error())
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
						Phone:             userInfo.Phone,
						CountryCode:       userInfo.CountryCode,
						Region:            userInfo.CountryCode,
						Score:             initScore,
						IsAdmin:           false,
						IsGlobalAdmin:     false,
						IsForbidden:       false,
						IsDeleted:         false,
						SignupApplication: application.Name,
						Properties:        properties,
					}

					affected, err := object.AddUser(user)
					if err != nil {
						c.ResponseError(err.Error())
						return
					}

					if !affected {
						c.ResponseError(fmt.Sprintf(c.T("auth:Failed to create user, user information is invalid: %s"), util.StructToJson(user)))
						return
					}
				}

				// sync info from 3rd-party if possible
				_, err := object.SetUserOAuthProperties(organization, user, provider.Type, userInfo)
				if err != nil {
					c.ResponseError(err.Error())
					return
				}

				_, err = object.LinkUserAccount(user, provider.Type, userInfo.Id)
				if err != nil {
					c.ResponseError(err.Error())
					return
				}

				resp = c.HandleLoggedIn(application, user, &authForm)

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
				resp = &Response{Status: "error", Msg: fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(application.Organization, userInfo.Id))}
			}
			// resp = &Response{Status: "ok", Msg: "", Data: res}
		} else { // authForm.Method != "signup"
			userId := c.GetSessionUsername()
			if userId == "" {
				c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(application.Organization, userInfo.Id)), userInfo)
				return
			}

			oldUser, err := object.GetUserByField(application.Organization, provider.Type, userInfo.Id)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			if oldUser != nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The account for provider: %s and username: %s (%s) is already linked to another account: %s (%s)"), provider.Type, userInfo.Username, userInfo.DisplayName, oldUser.Name, oldUser.DisplayName))
				return
			}

			user, err := object.GetUser(userId)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			// sync info from 3rd-party if possible
			_, err = object.SetUserOAuthProperties(organization, user, provider.Type, userInfo)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			isLinked, err := object.LinkUserAccount(user, provider.Type, userInfo.Id)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			if isLinked {
				resp = &Response{Status: "ok", Msg: "", Data: isLinked}
			} else {
				resp = &Response{Status: "error", Msg: "Failed to link user account", Data: isLinked}
			}
		}
	} else if c.getMfaSessionData() != nil {
		mfaSession := c.getMfaSessionData()
		user, err := object.GetUser(mfaSession.UserId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if authForm.Passcode != "" {
			mfaUtil := object.GetMfaUtil(authForm.MfaType, user.GetPreferredMfaProps(false))
			if mfaUtil == nil {
				c.ResponseError("Invalid multi-factor authentication type")
				return
			}

			err = mfaUtil.Verify(authForm.Passcode)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}
		if authForm.RecoveryCode != "" {
			err = object.MfaRecover(user, authForm.RecoveryCode)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}

		application, err := object.GetApplication(fmt.Sprintf("admin/%s", authForm.Application))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if application == nil {
			c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), authForm.Application))
			return
		}

		resp = c.HandleLoggedIn(application, user, &authForm)

		record := object.NewRecord(c.Ctx)
		record.Organization = application.Organization
		record.User = user.Name
		util.SafeGoroutine(func() { object.AddRecord(record) })
	} else {
		if c.GetSessionUsername() != "" {
			// user already signed in to Casdoor, so let the user click the avatar button to do the quick sign-in
			application, err := object.GetApplication(fmt.Sprintf("admin/%s", authForm.Application))
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			if application == nil {
				c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), authForm.Application))
				return
			}

			user := c.getCurrentUser()
			resp = c.HandleLoggedIn(application, user, &authForm)

			record := object.NewRecord(c.Ctx)
			record.Organization = application.Organization
			record.User = user.Name
			util.SafeGoroutine(func() { object.AddRecord(record) })
		} else {
			c.ResponseError(fmt.Sprintf(c.T("auth:Unknown authentication type (not password or provider), form = %s"), util.StructToJson(authForm)))
			return
		}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) GetSamlLogin() {
	providerId := c.Input().Get("id")
	relayState := c.Input().Get("relayState")
	authURL, method, err := object.GenerateSamlRequest(providerId, relayState, c.Ctx.Request.Host, c.GetAcceptLanguage())
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
		return
	}

	var data struct {
		MsgType  string `xml:"MsgType"`
		Event    string `xml:"Event"`
		EventKey string `xml:"EventKey"`
	}
	err = xml.Unmarshal(respBytes, &data)
	if err != nil {
		c.ResponseError(err.Error())
		return
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

// GetCaptchaStatus
// @Title GetCaptchaStatus
// @Tag Token API
// @Description Get Login Error Counts
// @Param   id     query    string  true        "The id ( owner/name ) of user"
// @Success 200 {object} controllers.Response The Response object
// @router /api/get-captcha-status [get]
func (c *ApiController) GetCaptchaStatus() {
	organization := c.Input().Get("organization")
	userId := c.Input().Get("user_id")
	user, err := object.GetUserByFields(organization, userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	var captchaEnabled bool
	if user != nil && user.SigninWrongTimes >= object.SigninWrongTimesLimit {
		captchaEnabled = true
	}
	c.ResponseOk(captchaEnabled)
}
