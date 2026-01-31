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
	"crypto/tls"
	"fmt"
	"hash/fnv"
	"log"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	ldap "github.com/casdoor/ldapserver"
	"github.com/lor00x/goldap/message"
)

func StartLdapServer() {
	ldapServerPort := conf.GetConfigString("ldapServerPort")
	ldapsServerPort := conf.GetConfigString("ldapsServerPort")

	server := ldap.NewServer()
	serverSsl := ldap.NewServer()
	routes := ldap.NewRouteMux()

	routes.Bind(handleBind)
	routes.Search(handleSearch).Label(" SEARCH****")

	server.Handle(routes)
	serverSsl.Handle(routes)
	go func() {
		if ldapServerPort == "" || ldapServerPort == "0" {
			return
		}
		err := server.ListenAndServe("0.0.0.0:" + ldapServerPort)
		if err != nil {
			log.Printf("StartLdapServer() failed, err = %s", err.Error())
		}
	}()

	go func() {
		if ldapsServerPort == "" || ldapsServerPort == "0" {
			return
		}
		ldapsCertId := conf.GetConfigString("ldapsCertId")
		if ldapsCertId == "" {
			return
		}
		config, err := getTLSconfig(ldapsCertId)
		if err != nil {
			log.Printf("StartLdapsServer() failed, err = %s", err.Error())
			return
		}
		secureConn := func(s *ldap.Server) {
			s.Listener = tls.NewListener(s.Listener, config)
		}
		err = serverSsl.ListenAndServe("0.0.0.0:"+ldapsServerPort, secureConn)
		if err != nil {
			log.Printf("StartLdapsServer() failed, err = %s", err.Error())
		}
	}()
}

func getTLSconfig(ldapsCertId string) (*tls.Config, error) {
	rawCert, err := object.GetCert(ldapsCertId)
	if err != nil {
		return nil, err
	}
	if rawCert == nil {
		return nil, fmt.Errorf("cert is empty")
	}
	cert, err := tls.X509KeyPair([]byte(rawCert.Certificate), []byte(rawCert.PrivateKey))
	if err != nil {
		return &tls.Config{}, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS10,
		MaxVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}, nil
}

func handleBind(w ldap.ResponseWriter, m *ldap.Message) {
	r := m.GetBindRequest()
	res := ldap.NewBindResponse(ldap.LDAPResultSuccess)

	if r.AuthenticationChoice() == "simple" {
		bindUsername, bindOrg, err := getNameAndOrgFromDN(string(r.Name()))
		if err != nil {
			log.Printf("getNameAndOrgFromDN() error: %s", err.Error())
			res.SetResultCode(ldap.LDAPResultInvalidDNSyntax)
			res.SetDiagnosticMessage(fmt.Sprintf("getNameAndOrgFromDN() error: %s", err.Error()))
			w.Write(res)
			return
		}

		bindPassword := string(r.AuthenticationSimple())

		enableCaptcha := false
		isSigninViaLdap := false
		isPasswordWithLdapEnabled := false
		if bindPassword != "" {
			isPasswordWithLdapEnabled = true
		}

		bindUser, err := object.CheckUserPassword(bindOrg, bindUsername, bindPassword, "en", enableCaptcha, isSigninViaLdap, isPasswordWithLdapEnabled)
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

	// Handle Stop Signal (server stop / client disconnected / Abandoned request....)
	select {
	case <-m.Done:
		log.Print("Leaving handleSearch...")
		return
	default:
	}

	if strings.EqualFold(r.FilterString(), "(objectClass=*)") && (string(r.BaseObject()) == "" || strings.EqualFold(string(r.BaseObject()), "cn=Subschema")) {
		handleRootSearch(w, &r, &res, m)
		return
	}

	var isGroupSearch bool = false
	filter := r.Filter()
	if eq, ok := filter.(message.FilterEqualityMatch); ok && strings.EqualFold(string(eq.AttributeDesc()), "objectClass") && strings.EqualFold(string(eq.AssertionValue()), "posixGroup") {
		isGroupSearch = true
	}

	if isGroupSearch {
		groups, code := GetFilteredGroups(m, string(r.BaseObject()), r.FilterString())
		if code != ldap.LDAPResultSuccess {
			res.SetResultCode(code)
			w.Write(res)
			return
		}

		for _, group := range groups {
			dn := fmt.Sprintf("cn=%s,%s", group.Name, string(r.BaseObject()))
			e := ldap.NewSearchResultEntry(dn)
			e.AddAttribute("cn", message.AttributeValue(group.Name))
			gidNumberStr := fmt.Sprintf("%v", hash(group.Name))
			e.AddAttribute("gidNumber", message.AttributeValue(gidNumberStr))
			users := object.GetGroupUsersWithoutError(group.GetId())
			for _, user := range users {
				e.AddAttribute("memberUid", message.AttributeValue(user.Name))
			}
			e.AddAttribute("objectClass", "posixGroup")
			w.Write(e)
		}

		w.Write(res)
		return
	}

	users, code := GetFilteredUsers(m)
	if code != ldap.LDAPResultSuccess {
		res.SetResultCode(code)
		w.Write(res)
		return
	}

	for _, user := range users {
		dn := fmt.Sprintf("uid=%s,cn=%s,%s", user.Id, user.Name, string(r.BaseObject()))
		e := ldap.NewSearchResultEntry(dn)
		uidNumberStr := fmt.Sprintf("%v", hash(user.Name))
		e.AddAttribute("uidNumber", message.AttributeValue(uidNumberStr))
		e.AddAttribute("gidNumber", message.AttributeValue(uidNumberStr))
		e.AddAttribute("homeDirectory", message.AttributeValue("/home/"+user.Name))
		e.AddAttribute("cn", message.AttributeValue(user.Name))
		e.AddAttribute("uid", message.AttributeValue(user.Id))
		e.AddAttribute("mail", message.AttributeValue(user.Email))
		e.AddAttribute("mobile", message.AttributeValue(user.Phone))
		e.AddAttribute("sn", message.AttributeValue(user.LastName))
		e.AddAttribute("givenName", message.AttributeValue(user.FirstName))
		// Add POSIX attributes for Linux machine login support
		e.AddAttribute("loginShell", getAttribute("loginShell", user))
		e.AddAttribute("gecos", getAttribute("gecos", user))
		// Add SSH public key if available
		sshKey := getAttribute("sshPublicKey", user)
		if sshKey != "" {
			e.AddAttribute("sshPublicKey", sshKey)
		}
		// Add objectClass for posixAccount
		e.AddAttribute("objectClass", "posixAccount")
		for _, group := range user.Groups {
			e.AddAttribute(ldapMemberOfAttr, message.AttributeValue(group))
		}
		attrs := r.Attributes()
		for _, attr := range attrs {
			if string(attr) == "*" {
				attrs = AdditionalLdapAttributes
				break
			}
		}
		for _, attr := range attrs {
			e.AddAttribute(message.AttributeDescription(attr), getAttribute(string(attr), user))
			if string(attr) == "title" {
				e.AddAttribute(message.AttributeDescription(attr), getAttribute("title", user))
			}
		}

		w.Write(e)
	}
	w.Write(res)
}

func handleRootSearch(w ldap.ResponseWriter, r *message.SearchRequest, res *message.SearchResultDone, m *ldap.Message) {
	if len(r.Attributes()) == 0 {
		w.Write(res)
		return
	}
	firstAttr := string(r.Attributes()[0])

	if string(r.BaseObject()) == "" {
		// Handle special root DSE requests
		if strings.EqualFold(firstAttr, "namingContexts") {
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
		} else if strings.EqualFold(firstAttr, "subschemaSubentry") {
			e := ldap.NewSearchResultEntry(string(r.BaseObject()))
			e.AddAttribute("subschemaSubentry", message.AttributeValue("cn=Subschema"))
			w.Write(e)
		}
	} else if strings.EqualFold(firstAttr, "objectclasses") && strings.EqualFold(string(r.BaseObject()), "cn=Subschema") {
		e := ldap.NewSearchResultEntry(string(r.BaseObject()))
		e.AddAttribute("objectClasses", []message.AttributeValue{
			"( 1.3.6.1.1.1.2.0 NAME 'posixAccount' DESC 'Abstraction of an account with POSIX attributes' SUP top AUXILIARY MUST ( cn $ uid $ uidNumber $ gidNumber $ homeDirectory ) MAY ( userPassword $ loginShell $ gecos $ description ) )",
			"( 1.3.6.1.1.1.2.2 NAME 'posixGroup' DESC 'Abstraction of a group of accounts' SUP top STRUCTURAL MUST ( cn $ gidNumber ) MAY ( userPassword $ memberUid $ description ) )",
		}...)
		w.Write(e)
	}

	w.Write(res)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
