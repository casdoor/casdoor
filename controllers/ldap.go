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
	"github.com/casdoor/casdoor/util"
)

type LdapResp struct {
	// Groups []LdapRespGroup `json:"groups"`
	Users      []object.LdapUser `json:"users"`
	ExistUuids []string          `json:"existUuids"`
}

//type LdapRespGroup struct {
//	GroupId   string
//	GroupName string
//}

type LdapSyncResp struct {
	Exist  []object.LdapUser `json:"exist"`
	Failed []object.LdapUser `json:"failed"`
}

// GetLdapUsers
// @Title GetLdapser
// @Tag Account API
// @Description get ldap users
// Param	id	string	true	"id"
// @Success 200 {object} LdapResp The Response object
// @router /get-ldap-users [get]
func (c *ApiController) GetLdapUsers() {
	id := c.Input().Get("id")

	_, ldapId := util.GetOwnerAndNameFromId(id)
	ldapServer, err := object.GetLdap(ldapId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	conn, err := ldapServer.GetLdapConn()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	//groupsMap, err := conn.GetLdapGroups(ldapServer.BaseDn)
	//if err != nil {
	//  c.ResponseError(err.Error())
	//	return
	//}

	//for _, group := range groupsMap {
	//	resp.Groups = append(resp.Groups, LdapRespGroup{
	//		GroupId:   group.GidNumber,
	//		GroupName: group.Cn,
	//	})
	//}

	users, err := conn.GetLdapUsers(ldapServer)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	uuids := make([]string, len(users))
	for i, user := range users {
		uuids[i] = user.GetLdapUuid()
	}
	existUuids, err := object.GetExistUuids(ldapServer.Owner, uuids)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := LdapResp{
		Users:      object.AutoAdjustLdapUser(users),
		ExistUuids: existUuids,
	}
	c.ResponseOk(resp)
}

// GetLdaps
// @Title GetLdaps
// @Tag Account API
// @Description get ldaps
// @Param	owner	query	string	false	"owner"
// @Success 200 {array} object.Ldap The Response object
// @router /get-ldaps [get]
func (c *ApiController) GetLdaps() {
	owner := c.Input().Get("owner")

	c.ResponseOk(object.GetMaskedLdaps(object.GetLdaps(owner)))
}

// GetLdap
// @Title GetLdap
// @Tag Account API
// @Description get ldap
// @Param	id	query	string	true	"id"
// @Success 200 {object} object.Ldap The Response object
// @router /get-ldap [get]
func (c *ApiController) GetLdap() {
	id := c.Input().Get("id")

	if util.IsStringsEmpty(id) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	_, name := util.GetOwnerAndNameFromId(id)
	c.ResponseOk(object.GetMaskedLdap(object.GetLdap(name)))
}

// AddLdap
// @Title AddLdap
// @Tag Account API
// @Description add ldap
// @Param	body	body	object.Ldap		true	"The details of the ldap"
// @Success 200 {object} controllers.Response The Response object
// @router /add-ldap [post]
func (c *ApiController) AddLdap() {
	var ldap object.Ldap
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldap)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if util.IsStringsEmpty(ldap.Owner, ldap.ServerName, ldap.Host, ldap.Username, ldap.Password, ldap.BaseDn) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	if ok, err := object.CheckLdapExist(&ldap); err != nil {
		c.ResponseError(err.Error())
		return
	} else if ok {
		c.ResponseError(c.T("ldap:Ldap server exist"))
		return
	}

	resp := wrapActionResponse(object.AddLdap(&ldap))
	resp.Data2 = ldap

	if ldap.AutoSync != 0 {
		err = object.GetLdapAutoSynchronizer().StartAutoSync(ldap.Id)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

// UpdateLdap
// @Title UpdateLdap
// @Tag Account API
// @Description update ldap
// @Param	body	body	object.Ldap		true	"The details of the ldap"
// @Success 200 {object} controllers.Response The Response object
// @router /update-ldap [post]
func (c *ApiController) UpdateLdap() {
	var ldap object.Ldap
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldap)
	if err != nil || util.IsStringsEmpty(ldap.Owner, ldap.ServerName, ldap.Host, ldap.Username, ldap.Password, ldap.BaseDn) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	prevLdap, err := object.GetLdap(ldap.Id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.UpdateLdap(&ldap)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if ldap.AutoSync != 0 {
		err := object.GetLdapAutoSynchronizer().StartAutoSync(ldap.Id)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else if ldap.AutoSync == 0 && prevLdap.AutoSync != 0 {
		object.GetLdapAutoSynchronizer().StopAutoSync(ldap.Id)
	}

	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// DeleteLdap
// @Title DeleteLdap
// @Tag Account API
// @Description delete ldap
// @Param	body	body	object.Ldap		true	"The details of the ldap"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-ldap [post]
func (c *ApiController) DeleteLdap() {
	var ldap object.Ldap
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldap)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.DeleteLdap(&ldap)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	object.GetLdapAutoSynchronizer().StopAutoSync(ldap.Id)

	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// SyncLdapUsers
// @Title SyncLdapUsers
// @Tag Account API
// @Description sync ldap users
// @Param	id	query	string		true	"id"
// @Success 200 {object} LdapSyncResp The Response object
// @router /sync-ldap-users [post]
func (c *ApiController) SyncLdapUsers() {
	id := c.Input().Get("id")

	owner, ldapId := util.GetOwnerAndNameFromId(id)
	var users []object.LdapUser
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &users)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	err = object.UpdateLdapSyncTime(ldapId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	exist, failed, _ := object.SyncLdapUsers(owner, users, ldapId)

	c.ResponseOk(&LdapSyncResp{
		Exist:  exist,
		Failed: failed,
	})
}
