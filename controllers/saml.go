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
	"net/http"

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

	enablePostBinding, err := c.GetBool("enablePostBinding", false)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	metadata, err := object.GetSamlMeta(application, host, enablePostBinding)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["xml"] = metadata
	c.ServeXML()
}

func (c *ApiController) HandleSamlRedirect() {
	host := c.Ctx.Request.Host

	owner := c.Ctx.Input.Param(":owner")
	application := c.Ctx.Input.Param(":application")

	relayState := c.Input().Get("RelayState")
	samlRequest := c.Input().Get("SAMLRequest")

	targetURL := object.GetSamlRedirectAddress(owner, application, relayState, samlRequest, host)

	c.Redirect(targetURL, http.StatusSeeOther)
}
