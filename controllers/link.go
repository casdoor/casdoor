// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	ProviderType string `json:"providerType"`
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

	user := object.GetUser(userId)
	value := object.GetUserField(user, providerType)

	if value == "" {
		c.ResponseError("Please link first", value)
		return
	}

	object.ClearUserOAuthProperties(user, providerType)

	object.LinkUserAccount(user, providerType, "")
	c.ResponseOk()
}
