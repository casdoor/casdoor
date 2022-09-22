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
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	ldap "github.com/forestmgy/ldapserver"
	"github.com/lor00x/goldap/message"
)

func StartLdapServer() {
	server := ldap.NewServer()
	routes := ldap.NewRouteMux()

	routes.Bind(handleBind)
	routes.Search(handleSearch).Label(" SEARCH****")

	server.Handle(routes)

	go server.ListenAndServe("127.0.0.1:" + conf.GetConfigString("ldapServerPort"))

	// When CTRL+C, SIGINT and SIGTERM signal occurs
	// Then stop server gracefully
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	close(ch)

	server.Stop()
}

func handleBind(w ldap.ResponseWriter, m *ldap.Message) {
	r := m.GetBindRequest()
	res := ldap.NewBindResponse(ldap.LDAPResultSuccess)

	if r.AuthenticationChoice() == "simple" {
		bindusername, bindorg, err := object.GetNameAndOrgFromDN(string(r.Name()))
		if err != "" {
			log.Printf("Bind failed ,ErrMsg=%s", err)
			res.SetResultCode(ldap.LDAPResultInvalidDNSyntax)
			res.SetDiagnosticMessage("bind failed ErrMsg: " + err)
			w.Write(res)
			return
		}
		bindpassword := string(r.AuthenticationSimple())
		binduser, err := object.CheckUserPassword(bindorg, bindusername, bindpassword)
		if err != "" {
			log.Printf("Bind failed User=%s, Pass=%#v, ErrMsg=%s", string(r.Name()), r.Authentication(), err)
			res.SetResultCode(ldap.LDAPResultInvalidCredentials)
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
		res.SetResultCode(ldap.LDAPResultAuthMethodNotSupported)
		res.SetDiagnosticMessage("Authentication method not supported,Please use Simple Authentication")
	}
	w.Write(res)
}

func handleSearch(w ldap.ResponseWriter, m *ldap.Message) {
	res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
	if !m.Client.IsAuthenticated {
		res.SetResultCode(ldap.LDAPResultUnwillingToPerform)
		w.Write(res)
		return
	}
	r := m.GetSearchRequest()
	object.PrintSearchInfo(r)
	if r.FilterString() == "(objectClass=*)" {
		w.Write(res)
		return
	}
	name, org, errCode := object.GetUserNameAndOrgFromBaseDnAndFilter(string(r.BaseObject()), r.FilterString())
	if errCode != ldap.LDAPResultSuccess {
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
	if errCode != ldap.LDAPResultSuccess {
		res.SetResultCode(errCode)
		w.Write(res)
		return
	}
	for i := 0; i < len(users); i++ {
		user := users[i]
		dn := fmt.Sprintf("cn=%s,%s", user.DisplayName, string(r.BaseObject()))
		e := ldap.NewSearchResultEntry(dn)
		e.AddAttribute("cn", message.AttributeValue(user.Name))
		e.AddAttribute("uid", message.AttributeValue(user.Name))
		e.AddAttribute("email", message.AttributeValue(user.Email))
		e.AddAttribute("mobile", message.AttributeValue(user.Phone))
		// e.AddAttribute("postalAddress", message.AttributeValue(user.Address[0]))
		w.Write(e)
	}
	w.Write(res)
}
