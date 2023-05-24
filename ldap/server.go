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

package ldap

import (
	"fmt"
	"log"

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
	err := server.ListenAndServe("0.0.0.0:" + conf.GetConfigString("ldapServerPort"))
	if err != nil {
		return
	}
}

func handleBind(w ldap.ResponseWriter, m *ldap.Message) {
	r := m.GetBindRequest()
	res := ldap.NewBindResponse(ldap.LDAPResultSuccess)

	if r.AuthenticationChoice() == "simple" {
		bindUsername, bindOrg, err := getNameAndOrgFromDN(string(r.Name()))
		if err != "" {
			log.Printf("Bind failed ,ErrMsg=%s", err)
			res.SetResultCode(ldap.LDAPResultInvalidDNSyntax)
			res.SetDiagnosticMessage("bind failed ErrMsg: " + err)
			w.Write(res)
			return
		}

		bindPassword := string(r.AuthenticationSimple())
		bindUser, err := object.CheckUserPassword(bindOrg, bindUsername, bindPassword, "en")
		if err != "" {
			log.Printf("Bind failed User=%s, Pass=%#v, ErrMsg=%s", string(r.Name()), r.Authentication(), err)
			res.SetResultCode(ldap.LDAPResultInvalidCredentials)
			res.SetDiagnosticMessage("invalid credentials ErrMsg: " + err)
			w.Write(res)
			return
		}

		if bindOrg == "built-in" || bindUser.IsGlobalAdmin {
			m.Client.IsGlobalAdmin, m.Client.IsOrgAdmin = true, true
		} else if bindUser.IsAdmin {
			m.Client.IsOrgAdmin = true
		}

		m.Client.IsAuthenticated = true
		m.Client.UserName = bindUsername
		m.Client.OrgName = bindOrg
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
	if r.FilterString() == "(objectClass=*)" {
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

	users, code := GetFilteredUsers(m)
	if code != ldap.LDAPResultSuccess {
		res.SetResultCode(code)
		w.Write(res)
		return
	}

	for _, user := range users {
		dn := fmt.Sprintf("cn=%s,%s", user.Name, string(r.BaseObject()))
		e := ldap.NewSearchResultEntry(dn)

		for _, attr := range r.Attributes() {
			e.AddAttribute(message.AttributeDescription(attr), getAttribute(string(attr), user))
			if string(attr) == "cn" {
				e.AddAttribute(message.AttributeDescription(attr), getAttribute("title", user))
			}
		}

		w.Write(e)
	}
	w.Write(res)
}
