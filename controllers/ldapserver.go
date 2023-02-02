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
	"log"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/forestmgy/ldapserver"
	"github.com/lor00x/goldap/message"
)

func StartLdapServer() {
	server := ldapserver.NewServer()
	routes := ldapserver.NewRouteMux()

	routes.Bind(handleBind)
	routes.Search(handleSearch).Label(" SEARCH****")

	server.Handle(routes)
	server.ListenAndServe("0.0.0.0:" + conf.GetConfigString("ldapServerPort"))
}

func handleBind(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	r := m.GetBindRequest()
	res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)

	if r.AuthenticationChoice() == "simple" {
		bindusername, bindorg, err := object.GetNameAndOrgFromDN(string(r.Name()))
		if err != "" {
			log.Printf("Bind failed ,ErrMsg=%s", err)
			res.SetResultCode(ldapserver.LDAPResultInvalidDNSyntax)
			res.SetDiagnosticMessage("bind failed ErrMsg: " + err)
			w.Write(res)
			return
		}
		bindpassword := string(r.AuthenticationSimple())
		binduser, err := object.CheckUserPassword(bindorg, bindusername, bindpassword, "en")
		if err != "" {
			log.Printf("Bind failed User=%s, Pass=%#v, ErrMsg=%s", string(r.Name()), r.Authentication(), err)
			res.SetResultCode(ldapserver.LDAPResultInvalidCredentials)
			res.SetDiagnosticMessage("invalid credentials ErrMsg: " + err)
			w.Write(res)
			return
		}
		if bindorg == "built-in" {
			m.Client.IsGlobalAdmin, m.Client.IsOrgAdmin = true, true
		} else if binduser.IsAdmin {
			m.Client.IsOrgAdmin = true
		}
		m.Client.IsAuthenticated = true
		m.Client.UserName = bindusername
		m.Client.OrgName = bindorg
	} else {
		res.SetResultCode(ldapserver.LDAPResultAuthMethodNotSupported)
		res.SetDiagnosticMessage("Authentication method not supported,Please use Simple Authentication")
	}
	w.Write(res)
}

func handleSearch(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
	if !m.Client.IsAuthenticated {
		res.SetResultCode(ldapserver.LDAPResultUnwillingToPerform)
		w.Write(res)
		return
	}
	r := m.GetSearchRequest()
	if r.FilterString() == "(objectClass=*)" {
		w.Write(res)
		return
	}
	name, org, errCode := object.GetUserNameAndOrgFromBaseDnAndFilter(string(r.BaseObject()), r.FilterString())
	if errCode != ldapserver.LDAPResultSuccess {
		res.SetResultCode(errCode)
		w.Write(res)
		return
	}
	// Handle Stop Signal (server stop / client disconnected / Abandoned request....)
	select {
	case <-m.Done:
		log.Print("Leaving handleSearch...")
		return
	default:
	}
	users, errCode := object.GetFilteredUsers(m, name, org)
	if errCode != ldapserver.LDAPResultSuccess {
		res.SetResultCode(errCode)
		w.Write(res)
		return
	}
	for i := 0; i < len(users); i++ {
		user := users[i]
		dn := fmt.Sprintf("cn=%s,%s", user.Name, string(r.BaseObject()))
		e := ldapserver.NewSearchResultEntry(dn)
		e.AddAttribute("cn", message.AttributeValue(user.Name))
		e.AddAttribute("uid", message.AttributeValue(user.Name))
		e.AddAttribute("email", message.AttributeValue(user.Email))
		e.AddAttribute("mobile", message.AttributeValue(user.Phone))
		e.AddAttribute("userPassword", message.AttributeValue(getUserPasswordWithType(user)))
		// e.AddAttribute("postalAddress", message.AttributeValue(user.Address[0]))
		w.Write(e)
	}
	w.Write(res)
}

// get user password with hash type prefix
// TODO not handle salt yet
// @return {md5}5f4dcc3b5aa765d61d8327deb882cf99
func getUserPasswordWithType(user *object.User) string {
	org := object.GetOrganizationByUser(user)
	if org.PasswordType == "" || org.PasswordType == "plain" {
		return user.Password
	}
	prefix := org.PasswordType
	if prefix == "salt" {
		prefix = "sha256"
	} else if prefix == "md5-salt" {
		prefix = "md5"
	} else if prefix == "pbkdf2-salt" {
		prefix = "pbkdf2"
	}
	return fmt.Sprintf("{%s}%s", prefix, user.Password)
}
