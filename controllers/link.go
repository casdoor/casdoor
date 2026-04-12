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
)

type LinkForm struct {
	ProviderType string      `json:"providerType"`
	ProviderName string      `json:"providerName"`
	User         object.User `json:"user"`
}

// Unlink ...
// @Tag Login API
// @Title Unlink
// @router /unlink [post]
// @Success 200 {object} object.Userinfo The Response object
func (c *ApiController) Unlink() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}

	var form LinkForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	providerType := form.ProviderType
	providerName := form.ProviderName

	// the user will be unlinked from the provider
	unlinkedUser := form.User

	if user.Id != unlinkedUser.Id && !user.IsGlobalAdmin() {
		// if the user is not the same as the one we are unlinking, we need to make sure the user is the global admin.
		c.ResponseError(c.T("link:You are not the global admin, you can't unlink other users"))
		return
	}

	if user.Id == unlinkedUser.Id && !user.IsGlobalAdmin() {
		// if the user is unlinking themselves, should check the provider can be unlinked, if not, we should return an error.
		application, err := object.GetApplicationByUser(user)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if application == nil {
			c.ResponseError(c.T("link:You can't unlink yourself, you are not a member of any application"))
			return
		}

		if len(application.Providers) == 0 {
			c.ResponseError(c.T("link:This application has no providers"))
			return
		}

		provider := application.GetProviderItem(providerName)
		if provider == nil && !object.IsFlexibleCustomProvider(providerType) {
			provider = application.GetProviderItemByType(providerType)
		}
		if provider == nil {
			if providerName != "" {
				c.ResponseError(c.T("link:This application has no providers") + ": " + providerName)
			} else {
				c.ResponseError(c.T("link:This application has no providers of type") + providerType)
			}
			return
		}

		if provider.Provider == nil {
			c.ResponseError(c.T("link:This application has no providers"))
			return
		}

		providerType = provider.Provider.Type
		providerName = provider.Name

		if !provider.CanUnlink {
			c.ResponseError(c.T("link:This provider can't be unlinked"))
			return
		}

	}

	// only two situations can happen here
	// 1. the user is the global admin
	// 2. the user is unlinking themselves and provider can be unlinked

	if object.IsFlexibleCustomProvider(providerType) {
		if providerName == "" {
			c.ResponseError(c.T("link:Provider name is required for Custom Flexible providers"))
			return
		}

		link, err := object.GetThirdPartyLink(unlinkedUser.Owner, unlinkedUser.Name, providerName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if link == nil {
			c.ResponseError(c.T("link:Please link first"))
			return
		}

		_, err = object.DeleteThirdPartyLink(unlinkedUser.Owner, unlinkedUser.Name, providerName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk()
		return
	}

	value := object.GetUserField(&unlinkedUser, providerType)

	if value == "" {
		c.ResponseError(c.T("link:Please link first"), value)
		return
	}

	_, err = object.ClearUserOAuthProperties(&unlinkedUser, providerType)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	_, err = object.LinkUserAccount(&unlinkedUser, providerType, "")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk()
}
