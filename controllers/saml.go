// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"fmt"

	"github.com/casdoor/casdoor/object"
)

func (c *ApiController) GetSamlMeta() {
	host := c.Ctx.Request.Host
	paramApp := c.Input().Get("application")
	application, err := object.GetApplication(paramApp)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("saml:Application %s not found"), paramApp))
		return
	}
	metadata, _ := object.GetSamlMeta(application, host)
	c.Data["xml"] = metadata
	c.ServeXML()
}
