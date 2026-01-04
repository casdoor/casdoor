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
	goldap "github.com/go-ldap/ldap/v3"
)

// ActiveDirectorySyncerProvider implements SyncerProvider for Active Directory LDAP-based syncers
type ActiveDirectorySyncerProvider struct {
	Syncer *Syncer
}

// InitAdapter initializes the Active Directory syncer (no database adapter needed)
func (p *ActiveDirectorySyncerProvider) InitAdapter() error {
	// Active Directory syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from Active Directory via LDAP
func (p *ActiveDirectorySyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getActiveDirectoryUsers()
}

// AddUser adds a new user to Active Directory (not supported for read-only LDAP)
func (p *ActiveDirectorySyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// Active Directory syncer is typically read-only
	return false, fmt.Errorf("adding users to Active Directory is not supported")
}

// UpdateUser updates an existing user in Active Directory (not supported for read-only LDAP)
func (p *ActiveDirectorySyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// Active Directory syncer is typically read-only
	return false, fmt.Errorf("updating users in Active Directory is not supported")
}

// TestConnection tests the Active Directory LDAP connection
func (p *ActiveDirectorySyncerProvider) TestConnection() error {
	conn, err := p.getLdapConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

// Close closes any open connections (no-op for Active Directory LDAP-based syncer)
func (p *ActiveDirectorySyncerProvider) Close() error {
	// Active Directory syncer doesn't maintain persistent connections
	// LDAP connections are opened and closed per operation
	return nil
}

// getLdapConn establishes an LDAP connection to Active Directory
func (p *ActiveDirectorySyncerProvider) getLdapConn() (*goldap.Conn, error) {
	// syncer.Host should be the AD server hostname/IP
	// syncer.Port should be the LDAP port (usually 389 or 636 for LDAPS)
	// syncer.User should be the bind DN or username
	// syncer.Password should be the bind password

	host := p.Syncer.Host
	if host == "" {
		return nil, fmt.Errorf("host is required for Active Directory syncer")
	}

	port := p.Syncer.Port
	if port == 0 {
		port = 389 // Default LDAP port
	}

	user := p.Syncer.User
	if user == "" {
		return nil, fmt.Errorf("user (bind DN) is required for Active Directory syncer")
	}

	password := p.Syncer.Password
	if password == "" {
		return nil, fmt.Errorf("password is required for Active Directory syncer")
	}

	var conn *goldap.Conn
	var err error

	// Check if SSL is enabled (port 636 typically indicates LDAPS)
	if port == 636 {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // TODO: Make this configurable
		}
		conn, err = goldap.DialTLS("tcp", fmt.Sprintf("%s:%d", host, port), tlsConfig)
	} else {
		conn, err = goldap.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to Active Directory: %w", err)
	}

	// Bind with the provided credentials
	err = conn.Bind(user, password)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind to Active Directory: %w", err)
	}

	return conn, nil
}

// getActiveDirectoryUsers retrieves all users from Active Directory
func (p *ActiveDirectorySyncerProvider) getActiveDirectoryUsers() ([]*OriginalUser, error) {
	conn, err := p.getLdapConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Use the Database field to store the base DN for searching
	baseDN := p.Syncer.Database
	if baseDN == "" {
		return nil, fmt.Errorf("database field (base DN) is required for Active Directory syncer")
	}

	// Search filter for user objects in Active Directory
	// Filter for users: objectClass=user, objectCategory=person, and not disabled accounts
	searchFilter := "(&(objectClass=user)(objectCategory=person))"

	// Attributes to retrieve from Active Directory
	attributes := []string{
		"sAMAccountName",     // Username
		"userPrincipalName",  // UPN (email-like format)
		"displayName",        // Display name
		"givenName",          // First name
		"sn",                 // Last name (surname)
		"mail",               // Email address
		"telephoneNumber",    // Phone number
		"mobile",             // Mobile phone
		"title",              // Job title
		"department",         // Department
		"company",            // Company
		"streetAddress",      // Street address
		"l",                  // City/Locality
		"st",                 // State/Province
		"postalCode",         // Postal code
		"co",                 // Country
		"objectGUID",         // Unique identifier
		"whenCreated",        // Creation time
		"userAccountControl", // Account status
	}

	searchRequest := goldap.NewSearchRequest(
		baseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0,     // No size limit
		0,     // No time limit
		false, // Types only = false
		searchFilter,
		attributes,
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search Active Directory: %w", err)
	}

	originalUsers := []*OriginalUser{}
	for _, entry := range sr.Entries {
		originalUser := p.adEntryToOriginalUser(entry)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// adEntryToOriginalUser converts an Active Directory LDAP entry to Casdoor OriginalUser
func (p *ActiveDirectorySyncerProvider) adEntryToOriginalUser(entry *goldap.Entry) *OriginalUser {
	user := &OriginalUser{
		Address:    []string{},
		Properties: map[string]string{},
		Groups:     []string{},
	}

	// Helper function to get and sanitize text attributes
	getTextAttribute := func(attrName string) string {
		value := entry.GetAttributeValue(attrName)
		return util.SanitizeUTF8String(value)
	}

	// Get basic attributes with UTF-8 sanitization
	sAMAccountName := getTextAttribute("sAMAccountName")
	userPrincipalName := getTextAttribute("userPrincipalName")
	displayName := getTextAttribute("displayName")
	givenName := getTextAttribute("givenName")
	sn := getTextAttribute("sn")
	mail := getTextAttribute("mail")
	telephoneNumber := getTextAttribute("telephoneNumber")
	mobile := getTextAttribute("mobile")
	title := getTextAttribute("title")
	department := getTextAttribute("department")
	company := getTextAttribute("company")
	streetAddress := getTextAttribute("streetAddress")
	city := getTextAttribute("l")
	state := getTextAttribute("st")
	postalCode := getTextAttribute("postalCode")
	country := getTextAttribute("co")
	whenCreated := getTextAttribute("whenCreated")
	userAccountControlStr := getTextAttribute("userAccountControl")

	// Get objectGUID as binary data and convert to string UUID
	var objectGUID string
	guidBytes := entry.GetRawAttributeValue("objectGUID")
	if len(guidBytes) > 0 {
		objectGUID = util.FormatADObjectGUID(guidBytes)
	}

	// Set user fields
	// Use sAMAccountName as the primary username
	user.Name = sAMAccountName

	// Use objectGUID as the unique ID if available, otherwise use sAMAccountName
	if objectGUID != "" {
		user.Id = objectGUID
	} else {
		user.Id = sAMAccountName
	}

	user.DisplayName = displayName
	user.FirstName = givenName
	user.LastName = sn

	// If display name is empty, construct from first and last name
	if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
		user.DisplayName = strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	}

	// Set email - prefer mail attribute, fallback to userPrincipalName
	if mail != "" {
		user.Email = mail
	} else if userPrincipalName != "" {
		user.Email = userPrincipalName
	}

	// Set phone - prefer mobile, fallback to telephoneNumber
	if mobile != "" {
		user.Phone = mobile
	} else if telephoneNumber != "" {
		user.Phone = telephoneNumber
	}

	user.Title = title

	// Set affiliation/department
	if department != "" {
		user.Affiliation = department
	}

	// Construct location from city, state, country
	locationParts := []string{}
	if city != "" {
		locationParts = append(locationParts, city)
	}
	if state != "" {
		locationParts = append(locationParts, state)
	}
	if country != "" {
		locationParts = append(locationParts, country)
	}
	if len(locationParts) > 0 {
		user.Location = strings.Join(locationParts, ", ")
	}

	// Construct address
	if streetAddress != "" {
		addressParts := []string{streetAddress}
		if city != "" {
			addressParts = append(addressParts, city)
		}
		if state != "" {
			addressParts = append(addressParts, state)
		}
		if postalCode != "" {
			addressParts = append(addressParts, postalCode)
		}
		if country != "" {
			addressParts = append(addressParts, country)
		}
		user.Address = []string{strings.Join(addressParts, ", ")}
	}

	// Store additional properties
	if company != "" {
		user.Properties["company"] = company
	}
	if userPrincipalName != "" {
		user.Properties["userPrincipalName"] = userPrincipalName
	}

	// Set creation time
	if whenCreated != "" {
		user.CreatedTime = whenCreated
	} else {
		user.CreatedTime = util.GetCurrentTime()
	}

	// Parse userAccountControl to determine if account is disabled
	// Bit 2 (value 2) indicates the account is disabled
	if userAccountControlStr != "" {
		userAccountControl := util.ParseInt(userAccountControlStr)
		// Check if bit 2 is set (account disabled)
		user.IsForbidden = (userAccountControl & 0x02) != 0
	}

	return user
}
