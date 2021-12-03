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

import "github.com/casbin/casdoor/object"

// @Title GetOidcDiscovery
// @Tag OIDC API
// @router /.well-known/openid-configuration [get]
func (c *RootController) GetOidcDiscovery() {
	c.Data["json"] = object.GetOidcDiscovery()
	c.ServeJSON()
}

// @Title GetOidcCert
// @Tag OIDC API
// @router /api/certs [get]
func (c *RootController) GetOidcCert() {
	jwks, err := object.GetJSONWebKeySet()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = jwks
	c.ServeJSON()
}
