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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/casdoor/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
)

const (
	adUserSearchFilter = "(&(objectClass=user)(objectCategory=person))"
	adPagingSize       = 100
)

// convertGUIDToString converts a binary GUID byte array to a standard UUID string format
// Active Directory GUIDs are 16 bytes in a specific byte order
func convertGUIDToString(guidBytes []byte) string {
	if len(guidBytes) != 16 {
		return ""
	}

	// Active Directory GUID format is:
	// Data1 (4 bytes, little-endian) - Data2 (2 bytes, little-endian) - Data3 (2 bytes, little-endian) - Data4 (2 bytes, big-endian) - Data5 (6 bytes, big-endian)
	// Convert to standard UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		guidBytes[3], guidBytes[2], guidBytes[1], guidBytes[0], // Data1 (reverse byte order)
		guidBytes[5], guidBytes[4], // Data2 (reverse byte order)
		guidBytes[7], guidBytes[6], // Data3 (reverse byte order)
		guidBytes[8], guidBytes[9], // Data4 (big-endian)
		guidBytes[10], guidBytes[11], guidBytes[12], guidBytes[13], guidBytes[14], guidBytes[15]) // Data5 (big-endian)
}

// sanitizeUTF8 drops invalid UTF-8 sequences from s.
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	var builder strings.Builder
	builder.Grow(len(s))

	for _, r := range s {
		if r == utf8.RuneError {
			continue
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

type adTargetKind int

const (
	adKindString adTargetKind = iota
	adKindBool
	adKindInt
)

type adTargetSpec struct {
	Kind adTargetKind
}

// adSettableTargets mirrors the fields handled by Syncer.setUserByKeyValue().
// Keep it in sync with object/syncer_util.go when new user fields become settable.
var adSettableTargets = []string{
	"Name", "CreatedTime", "UpdatedTime", "DeletedTime", "Id", "Type",
	"Password", "PasswordSalt", "PasswordType",
	"DisplayName", "FirstName", "LastName", "Avatar", "PermanentAvatar", "AvatarType",
	"Email", "EmailVerified", "Phone", "CountryCode", "Region", "Location", "Address",
	"Affiliation", "Title", "IdCardType", "IdCard", "RealName", "IsVerified",
	"Homepage", "Bio", "Tag", "Language", "Gender", "Birthday", "Education",
	"UidNumber", "Score", "Ranking", "IsDefaultAvatar", "IsOnline", "IsAdmin",
	"IsForbidden", "IsDeleted", "CreatedIp", "PreferredMfaType", "TotpSecret",
	"SignupApplication", "MfaPhoneEnabled", "MfaEmailEnabled", "RecoveryCodes",
	"ExternalId", "MfaRadiusEnabled", "MfaRadiusUsername", "MfaRadiusProvider",
	"MfaPushEnabled", "MfaPushReceiver", "MfaPushProvider", "Invitation",
	"InvitationCode", "Ldap", "LastSigninTime", "LastSigninIp", "LastChangePasswordTime",
	"LastSigninWrongTime", "SigninWrongTimes", "NeedUpdatePassword", "IpWhitelist",
	"MfaRememberDeadline", "WebauthnCredentials", "FaceIds", "ManagedAccounts",
	"MfaAccounts", "MfaItems",
}

var adCanonicalTargets = func() map[string]string {
	targets := make(map[string]string, len(adSettableTargets))
	for _, target := range adSettableTargets {
		targets[strings.ToLower(target)] = target
	}
	return targets
}()

// adRejectedTargets are settable fields that must not be driven by the directory.
var adRejectedTargets = map[string]string{
	"Password":            "credentials cannot be synced from Active Directory",
	"PasswordSalt":        "credentials cannot be synced from Active Directory",
	"PasswordType":        "credentials cannot be synced from Active Directory",
	"IsAdmin":             "directory-driven administrator privileges are not allowed",
	"IsDeleted":           "directory-driven deletion is not allowed",
	"TotpSecret":          "MFA secrets cannot be synced from Active Directory",
	"RecoveryCodes":       "MFA secrets cannot be synced from Active Directory",
	"WebauthnCredentials": "MFA secrets cannot be synced from Active Directory",
	"FaceIds":             "MFA secrets cannot be synced from Active Directory",
	"ManagedAccounts":     "managed accounts cannot be synced from Active Directory",
	"MfaAccounts":         "MFA secrets cannot be synced from Active Directory",
	"MfaItems":            "MFA secrets cannot be synced from Active Directory",
	"Address":             "Address is not a persisted scalar user column",
	"IpWhitelist":         "IP allowlists must not be driven by the directory",
}

// adKeyTargets are the only targets stable and unique enough to match users by.
var adKeyTargets = map[string]bool{
	"Id":    true,
	"Name":  true,
	"Email": true,
}

type adColumnPlan struct {
	Column     *TableColumn
	Sources    []string
	SourceAttr string
	Target     string
	Spec       adTargetSpec
}

type adSearchPlan struct {
	Columns    []*adColumnPlan
	Attributes []string
	Key        *adColumnPlan
}

func adTargetSpecOf(target string) (adTargetSpec, bool) {
	field, ok := reflect.TypeOf(User{}).FieldByName(target)
	if !ok {
		return adTargetSpec{}, false
	}

	switch field.Type.Kind() {
	case reflect.String:
		return adTargetSpec{Kind: adKindString}, true
	case reflect.Bool:
		return adTargetSpec{Kind: adKindBool}, true
	case reflect.Int:
		return adTargetSpec{Kind: adKindInt}, true
	default:
		return adTargetSpec{}, false
	}
}

func adLookupTarget(raw string) (string, adTargetSpec, error) {
	trimmed := strings.TrimSpace(raw)
	canonical, ok := adCanonicalTargets[strings.ToLower(trimmed)]
	if !ok {
		return "", adTargetSpec{}, fmt.Errorf("Casdoor column %q is not supported by the Active Directory syncer", trimmed)
	}
	if reason, rejected := adRejectedTargets[canonical]; rejected {
		return "", adTargetSpec{}, fmt.Errorf("Casdoor column %q cannot be synced: %s", canonical, reason)
	}

	spec, ok := adTargetSpecOf(canonical)
	if !ok {
		return "", adTargetSpec{}, fmt.Errorf("Casdoor column %q is not a supported scalar field", canonical)
	}
	return canonical, spec, nil
}

func parseADValueSources(name string) ([]string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}

	parts := strings.Split(name, "+")
	sources := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty attribute in %q", name)
		}
		for _, r := range part {
			if unicode.IsSpace(r) || unicode.IsControl(r) {
				return nil, fmt.Errorf("attribute %q contains whitespace or control characters", part)
			}
		}
		sources = append(sources, part)
	}
	return sources, nil
}

func isVirtualADAttribute(name string) bool {
	return strings.EqualFold(name, "dn") || strings.EqualFold(name, "distinguishedName")
}

func adTargetMapped(columns []*adColumnPlan, target string) bool {
	for _, column := range columns {
		if column.Target == target {
			return true
		}
	}
	return false
}

func validateADKey(column *adColumnPlan) error {
	if column == nil || column.Column == nil {
		return fmt.Errorf("the key column is missing")
	}
	if column.Spec.Kind != adKindString {
		return fmt.Errorf("the key column %q must map to a string field", column.Target)
	}
	if !adKeyTargets[column.Target] {
		return fmt.Errorf("the key column %q cannot be used as the primary key; allowed targets are Id, Name and Email", column.Target)
	}
	return nil
}

func resolveADKey(columns []*adColumnPlan, rawColumns []*TableColumn) (*adColumnPlan, bool, error) {
	planByColumn := make(map[*TableColumn]*adColumnPlan, len(columns))
	for _, column := range columns {
		planByColumn[column.Column] = column
	}

	explicitKeys := []*adColumnPlan{}
	for i, rawColumn := range rawColumns {
		if rawColumn == nil || !rawColumn.IsKey {
			continue
		}
		column, ok := planByColumn[rawColumn]
		if !ok {
			return nil, true, fmt.Errorf("table column #%d (%q) is marked as the key but has no usable attribute name or Casdoor column", i+1, rawColumn.Name)
		}
		explicitKeys = append(explicitKeys, column)
	}

	if len(explicitKeys) > 1 {
		return nil, true, fmt.Errorf("only one table column can be the key, found %d", len(explicitKeys))
	}
	if len(explicitKeys) == 1 {
		if err := validateADKey(explicitKeys[0]); err != nil {
			return nil, true, err
		}
		return explicitKeys[0], true, nil
	}

	for _, column := range columns {
		if column.Target == "Id" {
			if err := validateADKey(column); err != nil {
				return nil, false, err
			}
			return column, false, nil
		}
	}

	return nil, false, fmt.Errorf("no key column: enable 'Is key' on one row, or map an attribute to Id")
}

func normalizeADKeyFlags(columns []*TableColumn, key *TableColumn) {
	for _, column := range columns {
		if column == nil {
			continue
		}
		column.IsKey = column == key
	}
}

func (p *ActiveDirectorySyncerProvider) buildSearchPlan() (*adSearchPlan, error) {
	if len(p.Syncer.TableColumns) == 0 {
		return nil, fmt.Errorf("the Active Directory syncer has no table columns")
	}

	plan := &adSearchPlan{}
	for i, tableColumn := range p.Syncer.TableColumns {
		if tableColumn == nil {
			return nil, fmt.Errorf("table column #%d is null", i+1)
		}

		sources, err := parseADValueSources(tableColumn.Name)
		if err != nil {
			return nil, fmt.Errorf("table column #%d: %w", i+1, err)
		}
		if len(sources) == 0 {
			fmt.Printf("ActiveDirectorySyncer: ignoring table column #%d without an attribute name\n", i+1)
			continue
		}

		if strings.TrimSpace(tableColumn.CasdoorName) == "" {
			fmt.Printf("ActiveDirectorySyncer: ignoring table column #%d (%q) without a Casdoor column\n", i+1, tableColumn.Name)
			continue
		}
		target, spec, err := adLookupTarget(tableColumn.CasdoorName)
		if err != nil {
			return nil, fmt.Errorf("table column #%d: %w", i+1, err)
		}
		tableColumn.CasdoorName = target

		plan.Columns = append(plan.Columns, &adColumnPlan{
			Column:     tableColumn,
			Sources:    sources,
			SourceAttr: strings.Join(sources, "+"),
			Target:     target,
			Spec:       spec,
		})
	}
	if len(plan.Columns) == 0 {
		return nil, fmt.Errorf("the Active Directory syncer has no usable table columns")
	}

	if adTargetMapped(plan.Columns, "Ldap") {
		fmt.Print("ActiveDirectorySyncer: a column maps to Ldap; these users will sign in through LDAP instead of a local password\n")
	}

	seenAttributes := map[string]bool{}
	addAttribute := func(name string) {
		key := strings.ToLower(strings.TrimSpace(name))
		if key == "" || seenAttributes[key] {
			return
		}
		seenAttributes[key] = true
		plan.Attributes = append(plan.Attributes, strings.TrimSpace(name))
	}
	for _, column := range plan.Columns {
		for _, source := range column.Sources {
			if isVirtualADAttribute(source) {
				continue
			}
			addAttribute(source)
		}
	}
	if !adTargetMapped(plan.Columns, "Id") {
		addAttribute("objectGUID")
	}
	if !adTargetMapped(plan.Columns, "Name") {
		addAttribute("sAMAccountName")
	}
	if !adTargetMapped(plan.Columns, "IsForbidden") {
		addAttribute("userAccountControl")
	}
	if !adTargetMapped(plan.Columns, "CreatedTime") {
		addAttribute("whenCreated")
	}

	key, explicit, err := resolveADKey(plan.Columns, p.Syncer.TableColumns)
	if err != nil {
		return nil, err
	}
	if !explicit {
		normalizeADKeyFlags(p.Syncer.TableColumns, key.Column)
	}
	plan.Key = key

	if key.Target != "Id" {
		fmt.Printf("ActiveDirectorySyncer: the key column maps to %s; make sure it is unique and stable in Active Directory, or map objectGUID to Id\n", key.Target)
	}
	if !adTargetMapped(plan.Columns, "IsForbidden") {
		fmt.Print("ActiveDirectorySyncer: no column maps to IsForbidden; an account disabled in Active Directory is only marked as forbidden when the user is first created\n")
	}

	return plan, nil
}

type adEntryView struct {
	entry *goldap.Entry
	attrs map[string]*goldap.EntryAttribute
}

func newADEntryView(entry *goldap.Entry) *adEntryView {
	view := &adEntryView{
		entry: entry,
		attrs: make(map[string]*goldap.EntryAttribute, len(entry.Attributes)),
	}
	for _, attribute := range entry.Attributes {
		name := strings.ToLower(attribute.Name)
		if _, ok := view.attrs[name]; !ok {
			view.attrs[name] = attribute
		}
	}
	return view
}

func (v *adEntryView) first(name string) (raw []byte, text string, ok bool) {
	if isVirtualADAttribute(name) {
		if v.entry.DN == "" {
			return nil, "", false
		}
		return []byte(v.entry.DN), v.entry.DN, true
	}

	attribute, found := v.attrs[strings.ToLower(name)]
	if !found {
		return nil, "", false
	}
	if len(attribute.Values) > 0 {
		text = attribute.Values[0]
	}
	raw = v.entry.GetRawAttributeValue(attribute.Name)
	if raw == nil && text != "" {
		raw = []byte(text)
	}
	if text == "" && raw != nil {
		text = string(raw)
	}
	return raw, text, true
}

func convertADGeneralizedTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	for _, layout := range []string{"20060102150405.999Z", "20060102150405Z"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC().Format(time.RFC3339)
		}
	}
	return ""
}

func canonicalizeADValue(source string, raw []byte, text string, target string) string {
	switch strings.ToLower(source) {
	case "objectguid":
		if len(raw) == 16 {
			if guid := convertGUIDToString(raw); guid != "" {
				return guid
			}
		}
		return sanitizeUTF8(text)
	case "useraccountcontrol":
		if target == "IsForbidden" {
			value, err := strconv.Atoi(strings.TrimSpace(text))
			if err != nil {
				return ""
			}
			if value&0x02 != 0 {
				return "true"
			}
			return "false"
		}
	case "whencreated", "whenchanged":
		if converted := convertADGeneralizedTime(text); converted != "" {
			return converted
		}
		return sanitizeUTF8(text)
	case "objectsid":
		fmt.Printf("ActiveDirectorySyncer: skipping binary attribute %s (not supported yet)\n", source)
		return ""
	}

	if !utf8.ValidString(text) {
		fmt.Printf("ActiveDirectorySyncer: skipping attribute %s with a non-UTF-8 value\n", source)
		return ""
	}
	return sanitizeUTF8(text)
}

func normalizeADBool(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on", "y":
		return "true", true
	case "false", "0", "no", "off", "n":
		return "false", true
	default:
		return "", false
	}
}

func coerceADValue(value string, tableColumn *adColumnPlan) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	switch tableColumn.Spec.Kind {
	case adKindBool:
		return normalizeADBool(value)
	case adKindInt:
		number, err := strconv.Atoi(value)
		if err != nil {
			return "", false
		}
		return strconv.Itoa(number), true
	}

	switch strings.ToLower(strings.TrimSpace(tableColumn.Column.Type)) {
	case "integer", "int":
		if _, err := strconv.Atoi(value); err != nil {
			return "", false
		}
		return value, true
	case "boolean", "bool":
		return normalizeADBool(value)
	default:
		return value, true
	}
}

func applyADFallbacks(user *OriginalUser, view *adEntryView, assigned map[string]bool) {
	if user.Id == "" {
		if raw, text, ok := view.first("objectGUID"); ok {
			if len(raw) == 16 {
				user.Id = convertGUIDToString(raw)
			}
			if user.Id == "" {
				user.Id = sanitizeUTF8(text)
			}
		}
	}
	if user.Id == "" {
		if _, text, ok := view.first("sAMAccountName"); ok {
			user.Id = sanitizeUTF8(text)
		}
	}
	if user.Id == "" {
		user.Id = strings.TrimSpace(view.entry.DN)
	}

	if user.Name == "" {
		if _, text, ok := view.first("sAMAccountName"); ok {
			user.Name = sanitizeUTF8(text)
		}
		if user.Name == "" {
			user.Name = user.Id
		}
	}

	if !assigned["CreatedTime"] && user.CreatedTime == "" {
		if _, text, ok := view.first("whenCreated"); ok {
			user.CreatedTime = convertADGeneralizedTime(text)
			if user.CreatedTime == "" {
				user.CreatedTime = sanitizeUTF8(text)
			}
		}
		if user.CreatedTime == "" {
			user.CreatedTime = util.GetCurrentTime()
		}
	}

	if !assigned["IsForbidden"] {
		if _, text, ok := view.first("userAccountControl"); ok {
			if value, err := strconv.Atoi(strings.TrimSpace(text)); err == nil {
				user.IsForbidden = value&0x02 != 0
			}
		}
	}
}

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
	plan, err := p.buildSearchPlan()
	if err != nil {
		return nil, err
	}
	return p.getActiveDirectoryUsers(plan)
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

// formatADIssues renders at most limit issues and summarizes the rest.
func formatADIssues(issues []string, limit int) string {
	shown := issues
	if len(shown) > limit {
		shown = shown[:limit]
	}
	result := "  " + strings.Join(shown, "\n  ")
	if len(issues) > limit {
		result += fmt.Sprintf("\n  ... and %d more", len(issues)-limit)
	}
	return result
}

// getActiveDirectoryUsers retrieves all users from Active Directory
func (p *ActiveDirectorySyncerProvider) getActiveDirectoryUsers(plan *adSearchPlan) ([]*OriginalUser, error) {
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

	fmt.Printf("ActiveDirectorySyncer: searching attributes %v\n", plan.Attributes)

	searchRequest := goldap.NewSearchRequest(
		baseDN,
		goldap.ScopeWholeSubtree,
		goldap.NeverDerefAliases,
		0,     // No size limit
		0,     // No time limit
		false, // Types only = false
		adUserSearchFilter,
		plan.Attributes,
		nil,
	)

	sr, err := conn.SearchWithPaging(searchRequest, adPagingSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search Active Directory: %w", err)
	}

	keyField := util.CamelToSnakeCase(plan.Key.Target)
	originalUsers := []*OriginalUser{}
	seenKeys := map[string]string{}
	issues := []string{}
	for _, entry := range sr.Entries {
		originalUser, err := p.adEntryToOriginalUser(entry, plan)
		if err != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", entry.DN, err))
			continue
		}
		key := p.Syncer.getUserValue(originalUser, keyField)
		if key == "" {
			issues = append(issues, fmt.Sprintf("%s: no value for key %s", entry.DN, plan.Key.Target))
			continue
		}
		if previous, ok := seenKeys[key]; ok {
			issues = append(issues, fmt.Sprintf("%s: duplicate key %q, already used by %s", entry.DN, key, previous))
			continue
		}
		seenKeys[key] = entry.DN
		originalUsers = append(originalUsers, originalUser)
	}

	if len(issues) > 0 {
		return nil, fmt.Errorf("the Active Directory sync is aborted because %d of %d entries could not be mapped:\n%s", len(issues), len(sr.Entries), formatADIssues(issues, 5))
	}

	return originalUsers, nil
}

// adEntryToOriginalUser converts an Active Directory LDAP entry to Casdoor OriginalUser
func (p *ActiveDirectorySyncerProvider) adEntryToOriginalUser(entry *goldap.Entry, plan *adSearchPlan) (*OriginalUser, error) {
	view := newADEntryView(entry)
	user := &OriginalUser{
		Address:    []string{},
		Properties: map[string]string{},
		Groups:     []string{},
	}

	assigned := map[string]bool{}
	invalid := []string{}
	for _, tableColumn := range plan.Columns {
		values := []string{}
		for _, source := range tableColumn.Sources {
			raw, text, ok := view.first(source)
			if !ok {
				continue
			}
			value := canonicalizeADValue(source, raw, text, tableColumn.Target)
			if strings.TrimSpace(value) == "" {
				continue
			}
			values = append(values, value)
		}
		if len(values) == 0 {
			continue
		}

		value, ok := coerceADValue(strings.Join(values, " "), tableColumn)
		if !ok {
			invalid = append(invalid, fmt.Sprintf("invalid value %q for %s -> %s", strings.Join(values, " "), tableColumn.SourceAttr, tableColumn.Target))
			continue
		}
		if assigned[tableColumn.Target] {
			continue
		}
		assigned[tableColumn.Target] = true
		p.Syncer.setUserByKeyValue(user, tableColumn.Target, value)
	}

	applyADFallbacks(user, view, assigned)
	if len(invalid) > 0 {
		return user, errors.New(strings.Join(invalid, "; "))
	}
	return user, nil
}

// GetOriginalGroups retrieves all groups from Active Directory (not implemented yet)
func (p *ActiveDirectorySyncerProvider) GetOriginalGroups() ([]*OriginalGroup, error) {
	// TODO: Implement Active Directory group sync
	return []*OriginalGroup{}, nil
}

// GetOriginalUserGroups retrieves the group IDs that a user belongs to (not implemented yet)
func (p *ActiveDirectorySyncerProvider) GetOriginalUserGroups(userId string) ([]string, error) {
	// TODO: Implement Active Directory user group membership sync
	return []string{}, nil
}
