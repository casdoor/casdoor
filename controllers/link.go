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
	User         object.User `json:"user"`
}

// Unlink ...
// @router /unlink [post]
// @Tag Login API
func (c *ApiController) Unlink() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	var form LinkForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		panic(err)
	}
	providerType := form.ProviderType

	// the user will be unlinked from the provider
	unlinkedUser := form.User
	user := object.GetUser(userId)

	if user.Id != unlinkedUser.Id && !user.IsGlobalAdmin {
		// if the user is not the same as the one we are unlinking
		// we need to make sure the user is the global admin
		c.ResponseError("You are not the global admin, you can't unlink other users")
		return
	}

	if user.Id == unlinkedUser.Id && !user.IsGlobalAdmin {
		// if the user is unlinking themselves
		// should check the provider can be unlinked
		// if not, we should return an error

		app := object.GetApplicationByUser(user)
		if app == nil {
			c.ResponseError("You can't unlink yourself, you are not a member of any application")
			return
		}

		if len(app.Providers) == 0 {
			c.ResponseError("This application has no providers")
			return
		}

		for _, item := range app.Providers {
			if item.Provider.Type == providerType && !item.CanUnlink {
				c.ResponseError("You can't unlink yourself, cause you are not allowed to unlink this provider")
				return
			}
		}

	}

	// only two situations can happen here
	// 1. the user is the global admin
	// 2. the user is unlinking themselves and provider can be unlinked

	value := object.GetUserField(&unlinkedUser, providerType)

	if value == "" {
		c.ResponseError("Please link first", value)
		return
	}

	object.ClearUserOAuthProperties(&unlinkedUser, providerType)

	object.LinkUserAccount(&unlinkedUser, providerType, "")
	c.ResponseOk()
}
