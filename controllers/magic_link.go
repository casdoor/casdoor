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
	"errors"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const magicLinkSessionKey = "magicLinkSecret"

// newMagicLinkSessionHash ties the links issued to this browser to its session, so a
// link that leaves the mailbox is of no use in any other browser.
func (c *ApiController) newMagicLinkSessionHash() string {
	secret, ok := c.GetSession(magicLinkSessionKey).(string)
	if !ok || secret == "" {
		secret = util.GenerateId()
		c.SetSession(magicLinkSessionKey, secret)
	}

	return object.HashMagicLinkSecret(secret)
}

func (c *ApiController) getMagicLinkSessionHash() string {
	secret, ok := c.GetSession(magicLinkSessionKey).(string)
	if !ok || secret == "" {
		return ""
	}

	return object.HashMagicLinkSecret(secret)
}

// checkMagicLinkSignin claims the one-time link the browser came back with and answers
// with the user to sign in, creating that user when the application's signin method
// lets a magic link sign up.
func (c *ApiController) checkMagicLinkSignin(authForm *form.AuthForm) (*object.User, error) {
	application, err := object.GetApplication(fmt.Sprintf("admin/%s", authForm.Application))
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, fmt.Errorf(c.T("auth:The application: %s does not exist"), authForm.Application)
	}
	if !application.IsMagicLinkEnabled() {
		return nil, errors.New(c.T("auth:The login method: login with magic link is not enabled for the application"))
	}

	magicLink, err := object.ConsumeMagicLink(authForm.Code, c.getMagicLinkSessionHash(), application, c.GetAcceptLanguage())
	if err != nil {
		return nil, err
	}

	authForm.Organization = magicLink.Owner
	authForm.Username = magicLink.Email

	user, err := getUserByEmail(magicLink.Owner, magicLink.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return c.addMagicLinkUser(application, magicLink)
	}

	if user.IsDeleted || user.IsForbidden {
		return nil, errors.New(c.T("check:The user is forbidden to sign in, please contact the administrator"))
	}

	if !user.EmailVerified {
		user.EmailVerified = true
		_, err = object.UpdateUser(user.GetId(), user, []string{"email_verified"}, false)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (c *ApiController) addMagicLinkUser(application *object.Application, magicLink *object.MagicLink) (*object.User, error) {
	if !application.IsMagicLinkSignupEnabled() {
		return nil, errors.New(c.T("verification:the user does not exist, please sign up first"))
	}

	err := object.CheckMagicLinkSignup(application, c.GetAcceptLanguage())
	if err != nil {
		return nil, err
	}

	organization, err := object.GetOrganization(util.GetId("admin", application.Organization))
	if err != nil {
		return nil, err
	}
	if organization == nil {
		return nil, fmt.Errorf(c.T("auth:The organization: %s does not exist"), application.Organization)
	}

	clientIp := util.GetClientIpFromRequest(c.Ctx.Request)
	err = object.CheckEntryIp(clientIp, nil, application, organization, c.GetAcceptLanguage())
	if err != nil {
		return nil, err
	}

	// an application that asks the signup page for an invitation code is not signed up
	// to without one, whichever way the signup is started
	authForm := &form.AuthForm{Application: application.Name, Organization: application.Organization, Email: magicLink.Email}
	_, msg := object.CheckInvitationCode(application, organization, authForm, c.GetAcceptLanguage())
	if msg != "" {
		return nil, errors.New(msg)
	}

	id, err := object.GenerateIdForNewUser(application)
	if err != nil {
		return nil, err
	}

	initScore, err := organization.GetInitScore()
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(magicLink.Email)
	username := id
	if organization.UseEmailAsUsername {
		existedUser, err := object.GetUser(util.GetId(application.Organization, email))
		if err != nil {
			return nil, err
		}
		if existedUser == nil {
			username = email
		}
	}

	user := &object.User{
		Owner:             application.Organization,
		Name:              username,
		CreatedTime:       util.GetCurrentTime(),
		Id:                id,
		Type:              "normal-user",
		DisplayName:       email,
		Avatar:            organization.DefaultAvatar,
		Email:             email,
		Address:           []string{},
		Score:             initScore,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
		EmailVerified:     true,
		RegisterType:      "Application Signup",
		RegisterSource:    fmt.Sprintf("%s/%s", application.Organization, application.Name),
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
		return nil, errors.New(c.T("account:Failed to add user"))
	}

	err = object.AddUserToOriginalDatabase(user)
	if err != nil {
		return nil, err
	}

	c.Ctx.Input.SetParam("recordSignup", "true")
	util.LogInfo(c.Ctx, "API: [%s] is signed up as new user", user.GetId())

	return user, nil
}
