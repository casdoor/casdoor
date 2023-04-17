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
	Users []object.LdapRespUser `json:"users"`
}

//type LdapRespGroup struct {
//	GroupId   string
//	GroupName string
//}

type LdapSyncResp struct {
	Exist  []object.LdapRespUser `json:"exist"`
	Failed []object.LdapRespUser `json:"failed"`
}

// GetLdapUsers
// @Tag Account API
// @Title GetLdapser
// @router /get-ldap-users [get]
func (c *ApiController) GetLdapUsers() {
	id := c.Input().Get("id")

	_, ldapId := util.GetOwnerAndNameFromId(id)
	ldapServer := object.GetLdap(ldapId)

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

	var resp LdapResp
	uuids := make([]string, len(users))
	for _, user := range users {
		resp.Users = append(resp.Users, object.LdapRespUser{
			UidNumber: user.UidNumber,
			Uid:       user.Uid,
			Cn:        user.Cn,
			GroupId:   user.GidNumber,
			// GroupName: groupsMap[user.GidNumber].Cn,
			Uuid:        user.Uuid,
			DisplayName: user.DisplayName,
			Email:       util.GetMaxLenStr(user.Mail, user.Email, user.EmailAddress),
			Phone:       util.GetMaxLenStr(user.TelephoneNumber, user.Mobile, user.MobileTelephoneNumber),
			Address:     util.GetMaxLenStr(user.RegisteredAddress, user.PostalAddress),
		})
		uuids = append(uuids, user.Uuid)
	}

	existUuids := object.GetLdapUuid(ldapServer.Owner, uuids)

	c.ResponseOk(resp, existUuids)
}

// GetLdaps
// @Tag Account API
// @Title GetLdaps
// @router /get-ldaps [get]
func (c *ApiController) GetLdaps() {
	owner := c.Input().Get("owner")

	c.ResponseOk(object.GetLdaps(owner))
}

// GetLdap
// @Tag Account API
// @Title GetLdap
// @router /get-ldap [get]
func (c *ApiController) GetLdap() {
	id := c.Input().Get("id")

	if util.IsStringsEmpty(id) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	_, name := util.GetOwnerAndNameFromId(id)
	c.ResponseOk(object.GetLdap(name))
}

// AddLdap
// @Tag Account API
// @Title AddLdap
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

	if object.CheckLdapExist(&ldap) {
		c.ResponseError(c.T("ldap:Ldap server exist"))
		return
	}

	affected := object.AddLdap(&ldap)
	resp := wrapActionResponse(affected)
	resp.Data2 = ldap

	if ldap.AutoSync != 0 {
		object.GetLdapAutoSynchronizer().StartAutoSync(ldap.Id)
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

// UpdateLdap
// @Tag Account API
// @Title UpdateLdap
// @router /update-ldap [post]
func (c *ApiController) UpdateLdap() {
	var ldap object.Ldap
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldap)
	if err != nil || util.IsStringsEmpty(ldap.Owner, ldap.ServerName, ldap.Host, ldap.Username, ldap.Password, ldap.BaseDn) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	prevLdap := object.GetLdap(ldap.Id)
	affected := object.UpdateLdap(&ldap)

	if ldap.AutoSync != 0 {
		object.GetLdapAutoSynchronizer().StartAutoSync(ldap.Id)
	} else if ldap.AutoSync == 0 && prevLdap.AutoSync != 0 {
		object.GetLdapAutoSynchronizer().StopAutoSync(ldap.Id)
	}

	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// DeleteLdap
// @Tag Account API
// @Title DeleteLdap
// @router /delete-ldap [post]
func (c *ApiController) DeleteLdap() {
	var ldap object.Ldap
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldap)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected := object.DeleteLdap(&ldap)

	object.GetLdapAutoSynchronizer().StopAutoSync(ldap.Id)

	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// SyncLdapUsers
// @Tag Account API
// @Title SyncLdapUsers
// @router /sync-ldap-users [post]
func (c *ApiController) SyncLdapUsers() {
	owner := c.Input().Get("owner")
	ldapId := c.Input().Get("ldapId")
	var users []object.LdapRespUser
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &users)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	object.UpdateLdapSyncTime(ldapId)

	exist, failed := object.SyncLdapUsers(owner, users, ldapId)

	c.ResponseOk(&LdapSyncResp{
		Exist:  *exist,
		Failed: *failed,
	})
}
