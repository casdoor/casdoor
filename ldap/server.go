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
	"hash/fnv"
	"log"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	ldap "github.com/forestmgy/ldapserver"
	"github.com/lor00x/goldap/message"
)

func StartLdapServer() {
	ldapServerPort := conf.GetConfigString("ldapServerPort")
	if ldapServerPort == "" || ldapServerPort == "0" {
		return
	}

	server := ldap.NewServer()
	routes := ldap.NewRouteMux()

	routes.Bind(handleBind)
	routes.Search(handleSearch).Label(" SEARCH****")

	server.Handle(routes)
	err := server.ListenAndServe("0.0.0.0:" + ldapServerPort)
	if err != nil {
		log.Printf("StartLdapServer() failed, err = %s", err.Error())
	}
}

func handleBind(w ldap.ResponseWriter, m *ldap.Message) {
	r := m.GetBindRequest()
	res := ldap.NewBindResponse(ldap.LDAPResultSuccess)

	if r.AuthenticationChoice() == "simple" {
		bindDN := string(r.Name())
		bindPassword := string(r.AuthenticationSimple())

		if bindDN == "" && bindPassword == "" {
			res.SetResultCode(ldap.LDAPResultInappropriateAuthentication)
			res.SetDiagnosticMessage("Anonymous bind disallowed")
			w.Write(res)
			return
		}

		bindUsername, bindOrg, err := getNameAndOrgFromDN(bindDN)
		if err != nil {
			log.Printf("getNameAndOrgFromDN() error: %s", err.Error())
			res.SetResultCode(ldap.LDAPResultInvalidDNSyntax)
			res.SetDiagnosticMessage(fmt.Sprintf("getNameAndOrgFromDN() error: %s", err.Error()))
			w.Write(res)
			return
		}

		bindUser, err := object.CheckUserPassword(bindOrg, bindUsername, bindPassword, "en")
		if err != nil {
			log.Printf("Bind failed User=%s, Pass=%#v, ErrMsg=%s", string(r.Name()), r.Authentication(), err)
			res.SetResultCode(ldap.LDAPResultInvalidCredentials)
			res.SetDiagnosticMessage("invalid credentials ErrMsg: " + err.Error())
			w.Write(res)
			return
		}

		if bindOrg == "built-in" || bindUser.IsGlobalAdmin() {
			m.Client.IsGlobalAdmin, m.Client.IsOrgAdmin = true, true
		} else if bindUser.IsAdmin {
			m.Client.IsOrgAdmin = true
		}

		m.Client.IsAuthenticated = true
		m.Client.UserName = bindUsername
		m.Client.OrgName = bindOrg
	} else {
		res.SetResultCode(ldap.LDAPResultAuthMethodNotSupported)
		res.SetDiagnosticMessage("Authentication method not supported, please use Simple Authentication")
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

	// case insensitive match
	if strings.EqualFold(r.FilterString(), "(objectClass=*)") {
		if len(r.Attributes()) == 0 {
			w.Write(res)
			return
		}
		first_attr := string(r.Attributes()[0])

		if string(r.BaseObject()) == "" {
			// handle special search requests

			if first_attr == "namingContexts" {
				orgs, code := GetFilteredOrganizations(m)
				if code != ldap.LDAPResultSuccess {
					res.SetResultCode(code)
					w.Write(res)
					return
				}
				e := ldap.NewSearchResultEntry(string(r.BaseObject()))
				dnlist := make([]message.AttributeValue, len(orgs))
				for i, org := range orgs {
					dnlist[i] = message.AttributeValue(fmt.Sprintf("ou=%s", org.Name))
				}
				e.AddAttribute("namingContexts", dnlist...)
				w.Write(e)
			} else if first_attr == "subschemaSubentry" {
				e := ldap.NewSearchResultEntry(string(r.BaseObject()))
				e.AddAttribute("subschemaSubentry", message.AttributeValue("cn=Subschema"))
				w.Write(e)
			}
		} else if strings.EqualFold(first_attr, "objectclasses") && string(r.BaseObject()) == "cn=Subschema" {
			e := ldap.NewSearchResultEntry(string(r.BaseObject()))
			e.AddAttribute("objectClasses", []message.AttributeValue{
				"( 1.3.6.1.1.1.2.0 NAME 'posixAccount' DESC 'Abstraction of an account with POSIX attributes' SUP top AUXILIARY MUST ( cn $ uid $ uidNumber $ gidNumber $ homeDirectory ) MAY ( userPassword $ loginShell $ gecos $ description ) )",
				"( 1.3.6.1.1.1.2.2 NAME 'posixGroup' DESC 'Abstraction of a group of accounts' SUP top STRUCTURAL MUST ( cn $ gidNumber ) MAY ( userPassword $ memberUid $ description ) )",
			}...)
			w.Write(e)
		}

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

	objectClass := searchFilterForEquality(r.Filter(), "objectClass", "posixAccount", "posixGroup")
	switch objectClass {
	case "posixAccount":
		users, code := GetFilteredUsers(m)
		if code != ldap.LDAPResultSuccess {
			res.SetResultCode(code)
			w.Write(res)
			return
		}

		// log.Printf("Handling posixAccount filter=%s", r.FilterString())
		for _, user := range users {
			dn := fmt.Sprintf("uid=%s,cn=users,%s", user.Name, string(r.BaseObject()))
			e := ldap.NewSearchResultEntry(dn)
			attrs := r.Attributes()
			for _, attr := range attrs {
				if string(attr) == "*" {
					attrs = AdditionalLdapUserAttributes
					break
				}
			}
			for _, attr := range attrs {
				if strings.HasSuffix(string(attr), ";binary") {
					// unsupported: userCertificate;binary
					continue
				}
				field, ok := ldapUserAttributesMapping.CaseInsensitiveGet(string(attr))
				if ok {
					e.AddAttribute(message.AttributeDescription(attr), field.GetAttributeValues(user)...)
				}
			}
			w.Write(e)
		}

	case "posixGroup":
		// log.Printf("Handling posixGroup filter=%s", r.FilterString())
		groups, code := GetFilteredGroups(m)
		if code != ldap.LDAPResultSuccess {
			res.SetResultCode(code)
			w.Write(res)
			return
		}

		for _, group := range groups {
			dn := fmt.Sprintf("cn=%s,cn=groups,%s", group.Name, string(r.BaseObject()))
			e := ldap.NewSearchResultEntry(dn)
			attrs := r.Attributes()
			for _, attr := range attrs {
				if string(attr) == "*" {
					attrs = AdditionalLdapGroupAttributes
					break
				}
			}
			for _, attr := range attrs {
				field, ok := ldapGroupAttributesMapping.CaseInsensitiveGet(string(attr))
				if ok {
					e.AddAttribute(message.AttributeDescription(attr), field.GetAttributeValues(group)...)
				}
			}
			w.Write(e)
		}

	case "":
		log.Printf("Unmatched search request. filter=%s", r.FilterString())
	}

	w.Write(res)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
