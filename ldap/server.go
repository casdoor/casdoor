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

	orgCache := make(map[string]*object.Organization)

	for _, user := range users {
		if _, ok := orgCache[user.Owner]; !ok {
			org, err := object.GetOrganizationByUser(user)
			if err != nil {
				log.Printf("handleSearch: failed to get organization for user %s: %v", user.Name, err)
			}
			orgCache[user.Owner] = org
		}
		org := orgCache[user.Owner]

		e := buildUserSearchEntry(user, string(r.BaseObject()), resolveRequestAttributes(r.Attributes()), org)
		w.Write(e)
	}
	w.Write(res)
}

// resolveRequestAttributes expands the "*" wildcard to the full list of additional LDAP attributes.
func resolveRequestAttributes(attrs message.AttributeSelection) []string {
	result := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		if string(attr) == "*" {
			result = make([]string, 0, len(AdditionalLdapAttributes))
			for _, a := range AdditionalLdapAttributes {
				result = append(result, string(a))
			}
			return result
		}
		result = append(result, string(attr))
	}
	return result
}

// buildUserSearchEntry constructs an LDAP search result entry for the given user,
// respecting the organization's LdapAttributes filter.
func buildUserSearchEntry(user *object.User, baseDN string, attrs []string, org *object.Organization) message.SearchResultEntry {
	dn := fmt.Sprintf("uid=%s,cn=%s,%s", user.Id, user.Name, baseDN)
	e := ldap.NewSearchResultEntry(dn)
	uidNumberStr := fmt.Sprintf("%v", hash(user.Name))
	if IsLdapAttrAllowed(org, "uidNumber") {
		e.AddAttribute("uidNumber", message.AttributeValue(uidNumberStr))
	}
	if IsLdapAttrAllowed(org, "gidNumber") {
		e.AddAttribute("gidNumber", message.AttributeValue(uidNumberStr))
	}
	if IsLdapAttrAllowed(org, "homeDirectory") {
		e.AddAttribute("homeDirectory", message.AttributeValue("/home/"+user.Name))
	}
	if IsLdapAttrAllowed(org, "cn") {
		e.AddAttribute("cn", message.AttributeValue(user.Name))
	}
	if IsLdapAttrAllowed(org, "uid") {
		e.AddAttribute("uid", message.AttributeValue(user.Id))
	}
	if IsLdapAttrAllowed(org, "mail") {
		e.AddAttribute("mail", message.AttributeValue(user.Email))
	}
	if IsLdapAttrAllowed(org, "mobile") {
		e.AddAttribute("mobile", message.AttributeValue(user.Phone))
	}
	if IsLdapAttrAllowed(org, "sn") {
		e.AddAttribute("sn", message.AttributeValue(user.LastName))
	}
	if IsLdapAttrAllowed(org, "givenName") {
		e.AddAttribute("givenName", message.AttributeValue(user.FirstName))
	}
	// Add POSIX attributes for Linux machine login support
	if IsLdapAttrAllowed(org, "loginShell") {
		e.AddAttribute("loginShell", getAttribute("loginShell", user))
	}
	if IsLdapAttrAllowed(org, "gecos") {
		e.AddAttribute("gecos", getAttribute("gecos", user))
	}
	// Add SSH public key if available
	if IsLdapAttrAllowed(org, "sshPublicKey") {
		sshKey := getAttribute("sshPublicKey", user)
		if sshKey != "" {
			e.AddAttribute("sshPublicKey", sshKey)
		}
	}
	// Add objectClass for posixAccount
	e.AddAttribute("objectClass", "posixAccount")
	if IsLdapAttrAllowed(org, ldapMemberOfAttr) {
		for _, group := range user.Groups {
			e.AddAttribute(ldapMemberOfAttr, message.AttributeValue(group))
		}
	}
	for _, attr := range attrs {
		if !IsLdapAttrAllowed(org, attr) {
			continue
		}
		e.AddAttribute(message.AttributeDescription(attr), getAttribute(attr, user))
	}
	return e
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
