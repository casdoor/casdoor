// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/go-ldap/ldap/v3"
)

// ActiveDirectorySyncer implements SyncerProvider for Active Directory LDAP-based syncers
type ActiveDirectorySyncer struct {
	Syncer *Syncer
	conn   *ldap.Conn
}

// InitAdapter initializes the Active Directory LDAP connection
func (p *ActiveDirectorySyncer) InitAdapter() error {
	var err error
	ldapURL := fmt.Sprintf("%s:%d", p.Syncer.Host, p.Syncer.Port)

	// Use TLS if port is 636 (LDAPS)
	if p.Syncer.Port == 636 {
		p.conn, err = ldap.DialTLS("tcp", ldapURL, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		p.conn, err = ldap.Dial("tcp", ldapURL)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to Active Directory: %w", err)
	}

	// Bind with credentials
	err = p.conn.Bind(p.Syncer.User, p.Syncer.Password)
	if err != nil {
		p.conn.Close()
		return fmt.Errorf("failed to bind to Active Directory: %w", err)
	}

	return nil
}

// GetOriginalUsers retrieves all users from Active Directory
func (p *ActiveDirectorySyncer) GetOriginalUsers() ([]*OriginalUser, error) {
	if p.conn == nil {
		if err := p.InitAdapter(); err != nil {
			return nil, err
		}
	}

	// Use the database field as the base DN for searching
	baseDN := p.Syncer.Database
	if baseDN == "" {
		return nil, fmt.Errorf("database field must contain the base DN for Active Directory search")
	}

	// Search for all user objects
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=user)(objectCategory=person))", // Filter for user accounts
		[]string{
			"sAMAccountName",
			"userPrincipalName",
			"displayName",
			"givenName",
			"sn",
			"mail",
			"telephoneNumber",
			"mobile",
			"title",
			"physicalDeliveryOfficeName",
			"l",
			"streetAddress",
			"postalCode",
			"co",
			"preferredLanguage",
			"userAccountControl",
			"whenCreated",
		},
		nil,
	)

	sr, err := p.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search Active Directory: %w", err)
	}

	users := []*OriginalUser{}
	for _, entry := range sr.Entries {
		user := p.adEntryToOriginalUser(entry)
		users = append(users, user)
	}

	return users, nil
}

// AddUser adds a new user to Active Directory (not supported)
func (p *ActiveDirectorySyncer) AddUser(user *OriginalUser) (bool, error) {
	return false, fmt.Errorf("adding users to Active Directory is not supported")
}

// UpdateUser updates an existing user in Active Directory (not supported)
func (p *ActiveDirectorySyncer) UpdateUser(user *OriginalUser) (bool, error) {
	return false, fmt.Errorf("updating users in Active Directory is not supported")
}

// TestConnection tests the Active Directory connection
func (p *ActiveDirectorySyncer) TestConnection() error {
	if err := p.InitAdapter(); err != nil {
		return err
	}

	// Close connection after test
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}

	return nil
}

// adEntryToOriginalUser converts an Active Directory LDAP entry to Casdoor OriginalUser
func (p *ActiveDirectorySyncer) adEntryToOriginalUser(entry *ldap.Entry) *OriginalUser {
	sAMAccountName := entry.GetAttributeValue("sAMAccountName")
	userPrincipalName := entry.GetAttributeValue("userPrincipalName")
	displayName := entry.GetAttributeValue("displayName")
	givenName := entry.GetAttributeValue("givenName")
	surname := entry.GetAttributeValue("sn")
	mail := entry.GetAttributeValue("mail")
	phone := entry.GetAttributeValue("telephoneNumber")
	mobile := entry.GetAttributeValue("mobile")
	title := entry.GetAttributeValue("title")
	office := entry.GetAttributeValue("physicalDeliveryOfficeName")
	city := entry.GetAttributeValue("l")
	street := entry.GetAttributeValue("streetAddress")
	postalCode := entry.GetAttributeValue("postalCode")
	country := entry.GetAttributeValue("co")
	language := entry.GetAttributeValue("preferredLanguage")
	userAccountControl := entry.GetAttributeValue("userAccountControl")
	whenCreated := entry.GetAttributeValue("whenCreated")

	// Use sAMAccountName as the primary identifier
	userName := sAMAccountName
	if userName == "" {
		userName = userPrincipalName
	}

	user := &OriginalUser{
		Id:          sAMAccountName,
		Name:        userName,
		DisplayName: displayName,
		FirstName:   givenName,
		LastName:    surname,
		Email:       mail,
		Phone:       phone,
		Title:       title,
		Location:    office,
		Language:    language,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Set phone to mobile if primary phone is empty
	if user.Phone == "" && mobile != "" {
		user.Phone = mobile
	}

	// Build address from components
	if street != "" || city != "" || postalCode != "" || country != "" {
		addressParts := []string{}
		if street != "" {
			addressParts = append(addressParts, street)
		}
		if city != "" {
			addressParts = append(addressParts, city)
		}
		if postalCode != "" {
			addressParts = append(addressParts, postalCode)
		}
		if country != "" {
			addressParts = append(addressParts, country)
		}
		user.Address = []string{strings.Join(addressParts, ", ")}
	}

	// If display name is empty, construct from first and last name
	if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
		user.DisplayName = strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	}

	// If email is empty, use userPrincipalName as email
	if user.Email == "" && userPrincipalName != "" {
		user.Email = userPrincipalName
	}

	// Check if account is disabled
	// In Active Directory, userAccountControl bit 2 (0x0002) indicates disabled account
	if userAccountControl != "" {
		var uac int
		fmt.Sscanf(userAccountControl, "%d", &uac)
		user.IsForbidden = (uac & 0x0002) != 0
	}

	// Set CreatedTime from Active Directory whenCreated attribute
	if whenCreated != "" {
		user.CreatedTime = whenCreated
	} else {
		user.CreatedTime = util.GetCurrentTime()
	}

	// Store userPrincipalName in properties if different from email
	if userPrincipalName != "" && userPrincipalName != user.Email {
		user.Properties["userPrincipalName"] = userPrincipalName
	}

	return user
}

// Close closes the Active Directory connection
func (p *ActiveDirectorySyncer) Close() {
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
}
