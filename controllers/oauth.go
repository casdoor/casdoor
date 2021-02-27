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
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) RegisterOAuthApp() {
	var client object.Client
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &client)
	if err != nil {
		panic(err)
	}

	clientID, clientSecret := generateClientInfo()
	client.ID = clientID
	client.Secret = clientSecret
	c.Data["json"] = object.AddClient(&client)
	c.ServeJSON()
}

func generateClientInfo() (string, string) {
	clientID := util.GenerateRandStr(20)
	clientSecret := util.GenerateRandStr(40)
	return clientID, clientSecret
}

func (c *ApiController) GetOAuthApps() {
	userID := c.Input().Get("userId")
	clients := object.GetClientByUserID(userID)
	c.Data["json"] = clients
	c.ServeJSON()
}

func (c *ApiController) GetOAuthApp() {
	userID := c.Input().Get("userId")
	name := c.Input().Get("name")
	client := object.GetClientByUserIDAndName(userID, name)
	c.Data["json"] = client
	c.ServeJSON()
}

func (c *ApiController) DeleteOAuthApp() {
	var client object.Client
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &client)
	if err != nil {
		panic(err)
	}
	c.Data["json"] = object.DeleteClient(&client)
	c.ServeJSON()
}

func (c *ApiController) UpdateOAuthApp() {
	clientID := c.Input().Get("clientId")
	var client object.Client
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &client)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.UpdateClient(clientID, &client)
	c.ServeJSON()
}
