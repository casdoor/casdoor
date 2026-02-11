// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package object

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/nyaruka/phonenumbers"
	"github.com/thanhpk/randstr"
	"golang.org/x/text/encoding/unicode"
)

const (
	LdapGroupType = "ldap-group"
)

// formatUserPhone processes phone number for a user based on their CountryCode
func formatUserPhone(u *User) {
	if u.Phone == "" {
		return
	}

	// 1. Normalize hint (e.g., "China" -> "CN") for the parser
	countryHint := u.CountryCode
	if strings.EqualFold(countryHint, "China") {
		countryHint = "CN"
	}
	if len(countryHint) != 2 {
		countryHint = "" // Only 2-letter codes are valid hints
	}

	// 2. Try parsing (Strictly using countryHint from LDAP)
	num, err := phonenumbers.Parse(u.Phone, countryHint)

	if err == nil && num != nil && phonenumbers.IsValidNumber(num) {
		// Store a clean national number (digits only, without country prefix)
		u.Phone = fmt.Sprint(num.GetNationalNumber())
	}
}

type LdapConn struct {
	Conn *goldap.Conn
	IsAD bool
}

//type ldapGroup struct {
//	GidNumber string
//	Cn        string
//}

type LdapUser struct {
	UidNumber string `json:"uidNumber"`
	Uid       string `json:"uid"`
	Cn        string `json:"cn"`
	GidNumber string `json:"gidNumber"`
	// Gcn                   string
	Uuid                  string `json:"uuid"`
	UserPrincipalName     string `json:"userPrincipalName"`
	DisplayName           string `json:"displayName"`
	Mail                  string
	Email                 string `json:"email"`
	EmailAddress          string
	TelephoneNumber       string
	Mobile                string `json:"mobile"`
	MobileTelephoneNumber string
	RegisteredAddress     string
	PostalAddress         string
	Country               string `json:"country"`
	CountryName           string `json:"countryName"`

	GroupId    string            `json:"groupId"`
	Address    string            `json:"address"`
	MemberOf   string            `json:"memberOf"`
	MemberOfs  []string          `json:"memberOfs"`
	Attributes map[string]string `json:"attributes"`
}

func (ldap *Ldap) GetLdapConn() (c *LdapConn, err error) {
	var conn *goldap.Conn
	tlsConfig := tls.Config{
		InsecureSkipVerify: ldap.AllowSelfSignedCert,
	}
	if ldap.EnableSsl {
		conn, err = goldap.DialTLS("tcp", fmt.Sprintf("%s:%d", ldap.Host, ldap.Port), &tlsConfig)
	} else {
		conn, err = goldap.Dial("tcp", fmt.Sprintf("%s:%d", ldap.Host, ldap.Port))
	}

	if err != nil {
		return nil, err
	}

	err = conn.Bind(ldap.Username, ldap.Password)
	if err != nil {
		return nil, err
	}

	isAD, err := isMicrosoftAD(conn)
	if err != nil {
		return nil, err
	}
	return &LdapConn{Conn: conn, IsAD: isAD}, nil
}

func (l *LdapConn) Close() {
	if l.Conn == nil {
		return
	}

	err := l.Conn.Unbind()
	if err != nil {
		panic(err)
	}
}

func isMicrosoftAD(Conn *goldap.Conn) (bool, error) {
	SearchFilter := "(objectClass=*)"
	SearchAttributes := []string{"vendorname", "vendorversion", "isGlobalCatalogReady", "forestFunctionality"}

	searchReq := goldap.NewSearchRequest("",
		goldap.ScopeBaseObject, goldap.NeverDerefAliases, 0, 0, false,
		SearchFilter, SearchAttributes, nil)
	searchResult, err := Conn.Search(searchReq)
	if err != nil {
		return false, err
	}
	if len(searchResult.Entries) == 0 {
		return false, nil
	}
	isMicrosoft := false

	type ldapServerType struct {
		Vendorname           string
		Vendorversion        string
		IsGlobalCatalogReady string
		ForestFunctionality  string
	}
	var ldapServerTypes ldapServerType
	for _, entry := range searchResult.Entries {
		for _, attribute := range entry.Attributes {
			switch attribute.Name {
			case "vendorname":
				ldapServerTypes.Vendorname = attribute.Values[0]
			case "vendorversion":
				ldapServerTypes.Vendorversion = attribute.Values[0]
			case "isGlobalCatalogReady":
				ldapServerTypes.IsGlobalCatalogReady = attribute.Values[0]
			case "forestFunctionality":
				ldapServerTypes.ForestFunctionality = attribute.Values[0]
			}
		}
	}
	if ldapServerTypes.Vendorname == "" &&
		ldapServerTypes.Vendorversion == "" &&
		ldapServerTypes.IsGlobalCatalogReady == "TRUE" &&
		ldapServerTypes.ForestFunctionality != "" {
		isMicrosoft = true
	}
	return isMicrosoft, err
}

func (l *LdapConn) GetLdapUsers(ldapServer *Ldap) ([]LdapUser, error) {
	SearchAttributes := []string{
		"uidNumber", "cn", "sn", "gidNumber", "entryUUID", "displayName", "mail", "email",
		"emailAddress", "telephoneNumber", "mobile", "mobileTelephoneNumber", "registeredAddress", "postalAddress",
		"c", "co", "memberOf",
	}
	if l.IsAD {
		SearchAttributes = append(SearchAttributes, "sAMAccountName")
	} else {
		SearchAttributes = append(SearchAttributes, "uid")
	}

	for attribute := range ldapServer.CustomAttributes {
		SearchAttributes = append(SearchAttributes, attribute)
	}

	searchReq := goldap.NewSearchRequest(ldapServer.BaseDn, goldap.ScopeWholeSubtree, goldap.NeverDerefAliases,
		0, 0, false,
		ldapServer.Filter, SearchAttributes, nil)
	searchResult, err := l.Conn.SearchWithPaging(searchReq, 100)
	if err != nil {
		return nil, err
	}

	if len(searchResult.Entries) == 0 {
		return nil, errors.New("no result")
	}

	var ldapUsers []LdapUser
	for _, entry := range searchResult.Entries {
		var user LdapUser
		for _, attribute := range entry.Attributes {
			switch attribute.Name {
			case "uidNumber":
				user.UidNumber = attribute.Values[0]
			case "uid":
				user.Uid = attribute.Values[0]
			case "sAMAccountName":
				user.Uid = attribute.Values[0]
			case "cn":
				user.Cn = attribute.Values[0]
			case "gidNumber":
				user.GidNumber = attribute.Values[0]
			case "entryUUID":
				user.Uuid = attribute.Values[0]
			case "objectGUID":
				user.Uuid = attribute.Values[0]
			case "userPrincipalName":
				user.UserPrincipalName = attribute.Values[0]
			case "displayName":
				user.DisplayName = attribute.Values[0]
			case "mail":
				user.Mail = attribute.Values[0]
			case "email":
				user.Email = attribute.Values[0]
			case "emailAddress":
				user.EmailAddress = attribute.Values[0]
			case "telephoneNumber":
				user.TelephoneNumber = attribute.Values[0]
			case "mobile":
				user.Mobile = attribute.Values[0]
			case "mobileTelephoneNumber":
				user.MobileTelephoneNumber = attribute.Values[0]
			case "registeredAddress":
				user.RegisteredAddress = attribute.Values[0]
			case "postalAddress":
				user.PostalAddress = attribute.Values[0]
			case "c":
				user.Country = attribute.Values[0]
			case "co":
				user.CountryName = attribute.Values[0]
			case "memberOf":
				user.MemberOf = attribute.Values[0]
				user.MemberOfs = attribute.Values
			default:
				if propName, ok := ldapServer.CustomAttributes[attribute.Name]; ok {
					if user.Attributes == nil {
						user.Attributes = make(map[string]string)
					}
					user.Attributes[propName] = attribute.Values[0]
				}
			}
		}
		ldapUsers = append(ldapUsers, user)
	}

	return ldapUsers, nil
}

// FIXME: The Base DN does not necessarily contain the Group
//
//	func (l *ldapConn) GetLdapGroups(baseDn string) (map[string]ldapGroup, error) {
//		SearchFilter := "(objectClass=posixGroup)"
//		SearchAttributes := []string{"cn", "gidNumber"}
//		groupMap := make(map[string]ldapGroup)
//
//		searchReq := goldap.NewSearchRequest(baseDn,
//			goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
//			SearchFilter, SearchAttributes, nil)
//		searchResult, err := l.Conn.Search(searchReq)
//		if err != nil {
//			return nil, err
//		}
//
//		if len(searchResult.Entries) == 0 {
//			return nil, errors.New("no result")
//		}
//
//		for _, entry := range searchResult.Entries {
//			var ldapGroupItem ldapGroup
//			for _, attribute := range entry.Attributes {
//				switch attribute.Name {
//				case "gidNumber":
//					ldapGroupItem.GidNumber = attribute.Values[0]
//					break
//				case "cn":
//					ldapGroupItem.Cn = attribute.Values[0]
//					break
//				}
//			}
//			groupMap[ldapGroupItem.GidNumber] = ldapGroupItem
//		}
//
//		return groupMap, nil
//	}

func AutoAdjustLdapUser(users []LdapUser) []LdapUser {
	res := make([]LdapUser, len(users))
	for i, user := range users {
		res[i] = LdapUser{
			UidNumber:   user.UidNumber,
			Uid:         user.Uid,
			Cn:          user.Cn,
			GroupId:     user.GidNumber,
			Uuid:        user.GetLdapUuid(),
			DisplayName: user.DisplayName,
			Email:       util.ReturnAnyNotEmpty(user.Email, user.EmailAddress, user.Mail),
			Mobile:      util.ReturnAnyNotEmpty(user.Mobile, user.MobileTelephoneNumber, user.TelephoneNumber),
			Address:     util.ReturnAnyNotEmpty(user.Address, user.PostalAddress, user.RegisteredAddress),
			Country:     util.ReturnAnyNotEmpty(user.Country, user.CountryName),
			CountryName: user.CountryName,
			MemberOf:    user.MemberOf,
			MemberOfs:   user.MemberOfs,
			Attributes:  user.Attributes,
		}
	}
	return res
}

// parseGroupNameFromDN extracts the CN (Common Name) from an LDAP DN
// e.g., "CN=GroupName,OU=Groups,DC=example,DC=com" -> "GroupName"
func parseGroupNameFromDN(dn string) string {
	if dn == "" {
		return ""
	}

	// Split by comma and find the CN component
	parts := strings.Split(dn, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "cn=") {
			return part[3:] // Return everything after "cn="
		}
	}
	return ""
}

// extractGroupNamesFromMemberOf extracts group names from memberOf DNs
func extractGroupNamesFromMemberOf(memberOfs []string) []string {
	var groupNames []string
	for _, dn := range memberOfs {
		groupName := parseGroupNameFromDN(dn)
		if groupName != "" {
			groupNames = append(groupNames, groupName)
		}
	}
	return groupNames
}

// ensureGroupExists creates a group if it doesn't exist
func ensureGroupExists(owner, groupName string) error {
	if groupName == "" {
		return nil
	}

	existingGroup, err := getGroup(owner, groupName)
	if err != nil {
		return err
	}

	if existingGroup != nil {
		return nil // Group already exists
	}

	// Create the group
	newGroup := &Group{
		Owner:       owner,
		Name:        groupName,
		CreatedTime: util.GetCurrentTime(),
		UpdatedTime: util.GetCurrentTime(),
		DisplayName: groupName,
		Type:        LdapGroupType,
		IsEnabled:   true,
		IsTopGroup:  true,
	}

	_, err = AddGroup(newGroup)
	return err
}

// updateUserGroups updates an existing user's group memberships from LDAP
func updateUserGroups(owner string, syncUser LdapUser, ldapGroupNames []string, defaultGroup string) error {
	// Find the user by LDAP UUID
	user := &User{}
	has, err := ormer.Engine.Where("owner = ? AND ldap = ?", owner, syncUser.Uuid).Get(user)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("user with LDAP UUID %s not found", syncUser.Uuid)
	}

	// Prepare new group list
	newGroups := []string{}
	if defaultGroup != "" {
		newGroups = append(newGroups, defaultGroup)
	}
	newGroups = append(newGroups, ldapGroupNames...)

	// Update user groups
	user.Groups = newGroups
	_, err = UpdateUser(user.GetId(), user, []string{"groups"}, false)
	return err
}

func SyncLdapUsers(owner string, syncUsers []LdapUser, ldapId string) (existUsers []LdapUser, failedUsers []LdapUser, err error) {
	var uuids []string
	for _, user := range syncUsers {
		uuids = append(uuids, user.Uuid)
	}

	organization, err := getOrganization("admin", owner)
	if err != nil {
		panic(err)
	}

	ldap, err := GetLdap(ldapId)

	var dc []string
	for _, basedn := range strings.Split(ldap.BaseDn, ",") {
		if strings.Contains(basedn, "dc=") {
			dc = append(dc, basedn[3:])
		}
	}
	affiliation := strings.Join(dc, ".")

	var ou []string
	for _, admin := range strings.Split(ldap.Username, ",") {
		if strings.Contains(admin, "ou=") {
			ou = append(ou, admin[3:])
		}
	}
	tag := strings.Join(ou, ".")

	for _, syncUser := range syncUsers {
		existUuids, err := GetExistUuids(owner, uuids)
		if err != nil {
			return nil, nil, err
		}

		// Extract group names from LDAP memberOf attributes
		ldapGroupNames := extractGroupNamesFromMemberOf(syncUser.MemberOfs)

		// Ensure all LDAP groups exist in Casdoor
		for _, groupName := range ldapGroupNames {
			err := ensureGroupExists(owner, groupName)
			if err != nil {
				// Log warning but continue processing
				logs.Warning("Failed to create LDAP group %s: %v", groupName, err)
			}
		}

		found := false
		if len(existUuids) > 0 {
			for _, existUuid := range existUuids {
				if syncUser.Uuid == existUuid {
					existUsers = append(existUsers, syncUser)
					found = true

					// Update existing user's group memberships
					if len(ldapGroupNames) > 0 {
						err := updateUserGroups(owner, syncUser, ldapGroupNames, ldap.DefaultGroup)
						if err != nil {
							logs.Warning("Failed to update groups for user %s: %v", syncUser.Uuid, err)
						}
					}
				}
			}
		}

		if !found {
			score, err := organization.GetInitScore()
			if err != nil {
				return nil, nil, err
			}

			name, err := syncUser.buildLdapUserName(owner)
			if err != nil {
				return nil, nil, err
			}

			// Prepare group assignments for new user
			userGroups := []string{}
			if ldap.DefaultGroup != "" {
				userGroups = append(userGroups, ldap.DefaultGroup)
			}
			// Add LDAP groups
			userGroups = append(userGroups, ldapGroupNames...)

			newUser := &User{
				Owner:             owner,
				Name:              name,
				CreatedTime:       util.GetCurrentTime(),
				DisplayName:       syncUser.buildLdapDisplayName(),
				SignupApplication: organization.DefaultApplication,
				Type:              "normal-user",
				Avatar:            organization.DefaultAvatar,
				Email:             syncUser.Email,
				Phone:             syncUser.Mobile,
				CountryCode:       syncUser.Country,
				Address:           []string{syncUser.Address},
				Region:            util.ReturnAnyNotEmpty(syncUser.Country, syncUser.CountryName),
				Affiliation:       affiliation,
				Tag:               tag,
				Score:             score,
				Ldap:              syncUser.Uuid,
				Properties:        syncUser.Attributes,
				Groups:            userGroups,
			}
			formatUserPhone(newUser)

			affected, err := AddUser(newUser, "en")
			if err != nil {
				return nil, nil, err
			}

			if !affected {
				failedUsers = append(failedUsers, syncUser)
				continue
			}

			// Trigger webhook for LDAP user sync
			TriggerWebhookForUser("new-user-ldap", newUser)
		}
	}

	return existUsers, failedUsers, err
}

func GetExistUuids(owner string, uuids []string) ([]string, error) {
	var existUuids []string

	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	err := ormer.Engine.Table(tableNamePrefix+"user").Where("owner = ?", owner).Cols("ldap").
		In("ldap", uuids).Select("DISTINCT ldap").Find(&existUuids)
	if err != nil {
		return existUuids, err
	}

	return existUuids, nil
}

func ResetLdapPassword(user *User, oldPassword string, newPassword string, lang string) error {
	ldaps, err := GetLdaps(user.Owner)
	if err != nil {
		return err
	}

	for _, ldapServer := range ldaps {
		conn, err := ldapServer.GetLdapConn()
		if err != nil {
			continue
		}

		searchReq := goldap.NewSearchRequest(ldapServer.BaseDn, goldap.ScopeWholeSubtree, goldap.NeverDerefAliases,
			0, 0, false, ldapServer.buildAuthFilterString(user), []string{}, nil)

		searchResult, err := conn.Conn.Search(searchReq)
		if err != nil {
			conn.Close()
			return err
		}

		if len(searchResult.Entries) == 0 {
			conn.Close()
			continue
		}
		if len(searchResult.Entries) > 1 {
			conn.Close()
			return fmt.Errorf(i18n.Translate(lang, "check:Multiple accounts with same uid, please check your ldap server"))
		}

		userDn := searchResult.Entries[0].DN

		var pwdEncoded string
		modifyPasswordRequest := goldap.NewModifyRequest(userDn, nil)
		if conn.IsAD {
			utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
			pwdEncoded, err := utf16.NewEncoder().String("\"" + newPassword + "\"")
			if err != nil {
				conn.Close()
				return err
			}
			modifyPasswordRequest.Replace("unicodePwd", []string{pwdEncoded})
			modifyPasswordRequest.Replace("userAccountControl", []string{"512"})
		} else if oldPassword != "" {
			modifyPasswordRequestWithOldPassword := goldap.NewPasswordModifyRequest(userDn, oldPassword, newPassword)
			_, err = conn.Conn.PasswordModify(modifyPasswordRequestWithOldPassword)
			if err != nil {
				conn.Close()
				return err
			}
			conn.Close()
			return nil
		} else {
			switch ldapServer.PasswordType {
			case "SSHA":
				pwdEncoded, err = generateSSHA(newPassword)
				break
			case "MD5":
				md5Byte := md5.Sum([]byte(newPassword))
				md5Password := base64.StdEncoding.EncodeToString(md5Byte[:])
				pwdEncoded = "{MD5}" + md5Password
				break
			case "Plain":
				pwdEncoded = newPassword
				break
			default:
				pwdEncoded = newPassword
				break
			}
			modifyPasswordRequest.Replace("userPassword", []string{pwdEncoded})
		}

		err = conn.Conn.Modify(modifyPasswordRequest)
		if err != nil {
			conn.Close()
			return err
		}
		conn.Close()
	}
	return nil
}

func (ldapUser *LdapUser) buildLdapUserName(owner string) (string, error) {
	user := User{}
	uidWithNumber := fmt.Sprintf("%s_%s", ldapUser.Uid, ldapUser.UidNumber)
	has, err := ormer.Engine.Where("owner = ? and (name = ? or name = ?)", owner, ldapUser.Uid, uidWithNumber).Get(&user)
	if err != nil {
		return "", err
	}

	if has {
		if user.Name == ldapUser.Uid {
			return uidWithNumber, nil
		}
		return fmt.Sprintf("%s_%s", uidWithNumber, randstr.Hex(6)), nil
	}

	if ldapUser.Uid != "" {
		return ldapUser.Uid, nil
	}

	return ldapUser.Cn, nil
}

func (ldapUser *LdapUser) buildLdapDisplayName() string {
	if ldapUser.DisplayName != "" {
		return ldapUser.DisplayName
	}

	return ldapUser.Cn
}

func (ldapUser *LdapUser) GetLdapUuid() string {
	if ldapUser.Uuid != "" {
		return ldapUser.Uuid
	}
	if ldapUser.Uid != "" {
		return ldapUser.Uid
	}

	return ldapUser.Cn
}

func (ldap *Ldap) buildAuthFilterString(user *User) string {
	if len(ldap.FilterFields) == 0 {
		return fmt.Sprintf("(&%s(uid=%s))", ldap.Filter, user.Name)
	}

	filter := fmt.Sprintf("(&%s(|", ldap.Filter)
	for _, field := range ldap.FilterFields {
		filter = fmt.Sprintf("%s(%s=%s)", filter, field, user.getFieldFromLdapAttribute(field))
	}
	filter = fmt.Sprintf("%s))", filter)

	return filter
}

func (user *User) getFieldFromLdapAttribute(attribute string) string {
	switch attribute {
	case "uid":
		return user.Name
	case "sAMAccountName":
		return user.Name
	case "mail":
		return user.Email
	case "mobile":
		return user.Phone
	case "c", "co":
		return user.Region
	default:
		return ""
	}
}
