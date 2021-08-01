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

	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

type LdapServer struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Admin  string `json:"admin"`
	Passwd string `json:"passwd"`
	BaseDn string `json:"baseDn"`
}

type LdapResp struct {
	//Groups []LdapRespGroup `json:"groups"`
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

func (c *ApiController) GetLdapUser() {
	ldapServer := LdapServer{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldapServer)
	if err != nil || util.IsStrsEmpty(ldapServer.Host, ldapServer.Admin, ldapServer.Passwd, ldapServer.BaseDn) {
		c.ResponseError("Missing parameter")
		return
	}

	var resp LdapResp

	conn, err := object.GetLdapConn(ldapServer.Host, ldapServer.Port, ldapServer.Admin, ldapServer.Passwd)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	//groupsMap, err := conn.GetLdapGroups(ldapServer.BaseDn)
	//if err != nil {
	//	c.ResponseError(err.Error())
	//	return
	//}

	//for _, group := range groupsMap {
	//	resp.Groups = append(resp.Groups, LdapRespGroup{
	//		GroupId:   group.GidNumber,
	//		GroupName: group.Cn,
	//	})
	//}

	users, err := conn.GetLdapUsers(ldapServer.BaseDn)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	for _, user := range users {
		resp.Users = append(resp.Users, object.LdapRespUser{
			UidNumber: user.UidNumber,
			Uid:       user.Uid,
			Cn:        user.Cn,
			GroupId:   user.GidNumber,
			//GroupName: groupsMap[user.GidNumber].Cn,
			Uuid:    user.Uuid,
			Email:   util.GetMaxLenStr(user.Mail, user.Email, user.EmailAddress),
			Phone:   util.GetMaxLenStr(user.TelephoneNumber, user.Mobile, user.MobileTelephoneNumber),
			Address: util.GetMaxLenStr(user.RegisteredAddress, user.PostalAddress),
		})
	}

	c.Data["json"] = Response{Status: "ok", Data: resp}
	c.ServeJSON()
	return
}

func (c *ApiController) GetLdaps() {
	owner := c.Input().Get("owner")

	c.Data["json"] = Response{Status: "ok", Data: object.GetLdaps(owner)}
	c.ServeJSON()
}

func (c *ApiController) GetLdap() {
	id := c.Input().Get("id")

	if util.IsStrsEmpty(id) {
		c.ResponseError("Missing parameter")
		return
	}

	c.Data["json"] = Response{Status: "ok", Data: object.GetLdap(id)}
	c.ServeJSON()
}

func (c *ApiController) AddLdap() {
	var ldap object.Ldap
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldap)
	if err != nil {
		c.ResponseError("Missing parameter")
		return
	}

	if util.IsStrsEmpty(ldap.Owner, ldap.ServerName, ldap.Host, ldap.Admin, ldap.Passwd, ldap.BaseDn) {
		c.ResponseError("Missing parameter")
		return
	}

	if object.CheckLdapExist(&ldap) {
		c.ResponseError("Ldap server exist")
		return
	}

	affected := object.AddLdap(&ldap)
	resp := wrapActionResponse(affected)
	if affected {
		resp.Data2 = ldap
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) UpdateLdap() {
	var ldap object.Ldap
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldap)
	if err != nil || util.IsStrsEmpty(ldap.Owner, ldap.ServerName, ldap.Host, ldap.Admin, ldap.Passwd, ldap.BaseDn) {
		c.ResponseError("Missing parameter")
		return
	}

	affected := object.UpdateLdap(&ldap)
	resp := wrapActionResponse(affected)
	if affected {
		resp.Data2 = ldap
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) DeleteLdap() {
	var ldap object.Ldap
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ldap)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteLdap(&ldap))
	c.ServeJSON()
}

func (c *ApiController) SyncLdapUsers() {
	owner := c.Input().Get("owner")
	ldapId := c.Input().Get("ldapId")
	var users []object.LdapRespUser
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &users)
	if err != nil {
		panic(err)
	}

	object.UpdateLdapSyncTime(ldapId)

	exist, failed := object.SyncLdapUsers(owner, users)
	c.Data["json"] = &Response{Status: "ok", Data: &LdapSyncResp{
		Exist:  *exist,
		Failed: *failed,
	}}
	c.ServeJSON()
}

func (c *ApiController) CheckLdapUsersExist() {
	owner := c.Input().Get("owner")
	var uuids []string
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &uuids)
	if err != nil {
		panic(err)
	}

	exist := object.CheckLdapUuidExist(owner, uuids)
	c.Data["json"] = &Response{Status: "ok", Data: exist}
	c.ServeJSON()
}
