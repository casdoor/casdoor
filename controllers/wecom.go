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

	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type WeComResp struct {
	Users    []*idp.UserInfo `json:"users"`
	ExistIds []string        `json:"existIds"`
}

type WeComSyncResp struct {
	Exist  []string `json:"exist"`
	Failed []string `json:"failed"`
}

// GetWeComUsers
// @Title GetWeComUsers
// @Tag Account API
// @Description get WeCom users
// Param	id	string	true	"id"
// @Success 200 {object} controllers.WeComResp The Response object
// @router /get-wecom-users [get]
func (c *ApiController) GetWeComUsers() {
	id := c.Input().Get("id")

	_, wecomId := util.GetOwnerAndNameFromId(id)
	wecomServer, err := object.GetWeCom(wecomId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if wecomServer == nil {
		c.ResponseError("WeCom server not found")
		return
	}

	syncer := idp.NewWeComSyncer(wecomServer.CorpId, wecomServer.CorpSecret, wecomServer.DepartmentId)

	users, err := syncer.GetAllUsers()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	userIds := make([]string, len(users))
	for i, user := range users {
		userIds[i] = user.Id
	}
	existIds, err := object.GetExistIds(wecomServer.Owner, userIds)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := WeComResp{
		Users:    users,
		ExistIds: existIds,
	}
	c.ResponseOk(resp)
}

// GetWeComs
// @Title GetWeComs
// @Tag Account API
// @Description get WeComs
// @Param	owner	query	string	false	"owner"
// @Success 200 {array} object.WeCom The Response object
// @router /get-wecoms [get]
func (c *ApiController) GetWeComs() {
	owner := c.Input().Get("owner")

	c.ResponseOk(object.GetMaskedWeComs(object.GetWeComs(owner)))
}

// GetWeCom
// @Title GetWeCom
// @Tag Account API
// @Description get WeCom
// @Param	id	query	string	true	"id"
// @Success 200 {object} object.WeCom The Response object
// @router /get-wecom [get]
func (c *ApiController) GetWeCom() {
	id := c.Input().Get("id")

	if util.IsStringsEmpty(id) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	_, name := util.GetOwnerAndNameFromId(id)
	weCom, err := object.GetWeCom(name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(object.GetMaskedWeCom(weCom))
}

// AddWeCom
// @Title AddWeCom
// @Tag Account API
// @Description add WeCom
// @Param	body	body	object.WeCom		true	"The details of the WeCom"
// @Success 200 {object} controllers.Response The Response object
// @router /add-wecom [post]
func (c *ApiController) AddWeCom() {
	var weCom object.WeCom
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &weCom)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if util.IsStringsEmpty(weCom.Owner, weCom.ServerName, weCom.CorpId, weCom.CorpSecret) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	if ok, err := object.CheckWeComExist(&weCom); err != nil {
		c.ResponseError(err.Error())
		return
	} else if ok {
		c.ResponseError(c.T("wecom:WeCom server exists"))
		return
	}

	resp := wrapActionResponse(object.AddWeCom(&weCom))
	resp.Data2 = weCom

	if weCom.AutoSync != 0 {
		err = object.GetWeComAutoSynchronizer().StartAutoSync(weCom.Id)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

// UpdateWeCom
// @Title UpdateWeCom
// @Tag Account API
// @Description update WeCom
// @Param	body	body	object.WeCom		true	"The details of the WeCom"
// @Success 200 {object} controllers.Response The Response object
// @router /update-wecom [post]
func (c *ApiController) UpdateWeCom() {
	var weCom object.WeCom
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &weCom)
	if err != nil || util.IsStringsEmpty(weCom.Owner, weCom.ServerName, weCom.CorpId, weCom.CorpSecret) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	prevWeCom, err := object.GetWeCom(weCom.Id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.UpdateWeCom(&weCom)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if weCom.AutoSync != 0 {
		err := object.GetWeComAutoSynchronizer().StartAutoSync(weCom.Id)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else if weCom.AutoSync == 0 && prevWeCom.AutoSync != 0 {
		object.GetWeComAutoSynchronizer().StopAutoSync(weCom.Id)
	}

	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// DeleteWeCom
// @Title DeleteWeCom
// @Tag Account API
// @Description delete WeCom
// @Param	body	body	object.WeCom		true	"The details of the WeCom"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-wecom [post]
func (c *ApiController) DeleteWeCom() {
	var weCom object.WeCom
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &weCom)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.DeleteWeCom(&weCom)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	object.GetWeComAutoSynchronizer().StopAutoSync(weCom.Id)

	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// SyncWeComUsers
// @Title SyncWeComUsers
// @Tag Account API
// @Description sync WeCom users
// @Param	id	query	string		true	"id"
// @Success 200 {object} controllers.WeComSyncResp The Response object
// @router /sync-wecom-users [post]
func (c *ApiController) SyncWeComUsers() {
	id := c.Input().Get("id")

	owner, wecomId := util.GetOwnerAndNameFromId(id)
	var users []*idp.UserInfo
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &users)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	err = object.UpdateWeComSyncTime(wecomId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	exist, failed, err := object.SyncWeComUsers(owner, users, wecomId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(&WeComSyncResp{
		Exist:  exist,
		Failed: failed,
	})
}
