// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/cred"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// the kinds in the order they are imported: the applications use the certs, the users join the groups, the roles
// and permissions name the users. Undoing deletes them in the opposite order.
var migrationKinds = []string{"certs", "groups", "applications", "users", "roles", "permissions"}

const (
	maxMigrationIssues   = 200
	maxMigrationWarnings = 50
	maxMigrationPages    = 100000
	// a user whose password hash only the import action can check has the password type "plugin:<action>"
	pluginPasswordTypePrefix = "plugin:"
)

var (
	migrationMutex    sync.Mutex
	runningMigrations = map[string]bool{}
)

type MigrationRunRequest struct {
	// "test", "plan", "apply" or "undo"
	Mode  string               `json:"mode"`
	Files []*action.ImportFile `json:"files"`
}

type migrationRun struct {
	migration    *Migration
	organization *Organization
	module       *action.Module
	config       *action.Config
	files        []*action.ImportFile
	report       *MigrationReport
	isApply      bool
	lang         string
	// the members a group lists by itself, for a system that keeps the memberships on the groups
	groupMembers map[string][]string
}

func isMigrationRunning(migration *Migration) bool {
	migrationMutex.Lock()
	defer migrationMutex.Unlock()
	return runningMigrations[migration.GetId()]
}

// StartMigration checks the migration and runs it in the background, the report shows its progress and result
func StartMigration(migration *Migration, req *MigrationRunRequest, lang string) error {
	if !util.InSlice([]string{"test", "plan", "apply", "undo"}, req.Mode) {
		return fmt.Errorf("unknown mode: %s", req.Mode)
	}
	if migration.Organization == "" || migration.Organization == "built-in" {
		return fmt.Errorf("choose an organization other than built-in, every user of built-in is a global admin")
	}
	if req.Mode != "undo" && migration.Plugin == "" {
		return fmt.Errorf("choose the action that reads the other system")
	}

	organization, err := getOrganization("admin", migration.Organization)
	if err != nil {
		return err
	}
	if organization == nil {
		return fmt.Errorf("the organization: %s is not found", migration.Organization)
	}

	run := &migrationRun{
		migration:    migration,
		organization: organization,
		config:       &action.Config{Host: migration.Host, ClientId: migration.ClientId, ClientSecret: migration.ClientSecret, Metadata: migration.Metadata},
		files:        req.Files,
		report:       &MigrationReport{Mode: req.Mode, StartedTime: util.GetCurrentTime(), Kinds: []*MigrationKindReport{}, Warnings: []string{}, Issues: []*MigrationIssue{}},
		isApply:      req.Mode == "apply",
		lang:         lang,
		groupMembers: map[string][]string{},
	}
	if req.Mode != "undo" {
		run.module, _, err = GetInstalledAction(migration.Organization, migration.Plugin, "import")
		if err != nil {
			return err
		}
	}

	migrationMutex.Lock()
	if runningMigrations[migration.GetId()] {
		migrationMutex.Unlock()
		return fmt.Errorf("the migration: %s is already running", migration.Name)
	}
	runningMigrations[migration.GetId()] = true
	migrationMutex.Unlock()

	previousState := migration.State
	migration.State = "Running"
	migration.Report = run.report
	err = run.save(false)
	if err != nil {
		run.finish()
		return err
	}

	go func() {
		defer run.finish()
		defer func() {
			if r := recover(); r != nil {
				run.fail(fmt.Errorf("%v", r))
			}
		}()

		var err error
		state := ""
		switch req.Mode {
		case "test":
			err = run.test()
			state = previousState
		case "plan":
			err = run.importAll()
			state = "Previewed"
		case "apply":
			err = run.importAll()
			state = "Applied"
		case "undo":
			err = run.undo()
			state = "Undone"
		}
		if err != nil {
			run.fail(err)
			return
		}

		run.report.FinishedTime = util.GetCurrentTime()
		run.migration.State = state
		err = run.save(req.Mode == "apply" || req.Mode == "undo")
		if err != nil {
			logs.Error("failed to save the migration: %s: %s", migration.GetId(), err.Error())
		}
	}()
	return nil
}

func (run *migrationRun) finish() {
	migrationMutex.Lock()
	delete(runningMigrations, run.migration.GetId())
	migrationMutex.Unlock()
}

func (run *migrationRun) fail(err error) {
	logs.Warn("the migration: %s failed: %s", run.migration.GetId(), err.Error())
	run.report.Error = err.Error()
	run.report.FinishedTime = util.GetCurrentTime()
	run.migration.State = "Failed"
	err = run.save(run.isApply || run.report.Mode == "undo")
	if err != nil {
		logs.Error("failed to save the migration: %s: %s", run.migration.GetId(), err.Error())
	}
}

func (run *migrationRun) save(withObjects bool) error {
	run.migration.UpdatedTime = util.GetCurrentTime()
	cols := []string{"state", "report", "updated_time"}
	if withObjects {
		cols = append(cols, "objects")
	}
	_, err := ormer.Engine.ID(core.PK{run.migration.Owner, run.migration.Name}).Cols(cols...).Update(run.migration)
	return err
}

func (run *migrationRun) test() error {
	err := action.TestImport(run.module, run.config, run.files)
	if action.IsNotImplemented(err) {
		// without a test of its own, reading the first page of users tells whether the action can reach the other system
		_, err = action.ListImport(run.module, run.config, &action.ImportListRequest{Kind: "users", Organization: run.migration.Organization, Files: run.files})
	}
	return err
}

func (run *migrationRun) addIssue(kind string, name string, reason string) {
	if len(run.report.Issues) < maxMigrationIssues {
		run.report.Issues = append(run.report.Issues, &MigrationIssue{Kind: kind, Name: name, Reason: reason})
	}
}

// splitWarningCount takes the count off a warning such as "3 users have no password", which every page repeats
func splitWarningCount(warning string) (int, string) {
	count, rest, ok := strings.Cut(warning, " ")
	n, err := strconv.Atoi(count)
	if !ok || err != nil || n < 0 {
		return -1, warning
	}
	return n, rest
}

// addWarnings adds up the counts of the same warning on several pages
func (run *migrationRun) addWarnings(warnings []string) {
	for _, warning := range warnings {
		count, rest := splitWarningCount(warning)
		merged := false
		for i, existing := range run.report.Warnings {
			existingCount, existingRest := splitWarningCount(existing)
			if existingRest != rest || (count < 0) != (existingCount < 0) {
				continue
			}
			if count >= 0 {
				run.report.Warnings[i] = fmt.Sprintf("%d %s", existingCount+count, rest)
			}
			merged = true
			break
		}
		if !merged && len(run.report.Warnings) < maxMigrationWarnings {
			run.report.Warnings = append(run.report.Warnings, warning)
		}
	}
}

func (run *migrationRun) importAll() error {
	for _, kind := range migrationKinds {
		err := run.importKind(kind)
		if err != nil {
			return fmt.Errorf("%s: %w", kind, err)
		}
		if kind == "users" && run.isApply {
			err = run.addGroupMembers()
			if err != nil {
				return fmt.Errorf("groups: %w", err)
			}
		}
	}
	return nil
}

func (run *migrationRun) addGroupMembers() error {
	owner := run.organization.Name
	for groupName, userNames := range run.groupMembers {
		groupId := util.GetId(owner, groupName)
		for _, userName := range userNames {
			if strings.HasPrefix(userName, owner+"/") {
				userName = strings.TrimPrefix(userName, owner+"/")
			}
			user, err := getUser(owner, userName)
			if err != nil {
				return err
			}
			if user == nil {
				run.addIssue("groups", groupName, fmt.Sprintf("the member: %s is not found", userName))
				continue
			}
			if util.InSlice(user.Groups, groupId) {
				continue
			}
			user.Groups = append(user.Groups, groupId)
			_, err = UpdateUser(user.GetId(), user, []string{"groups"}, false)
			if err != nil {
				run.addIssue("groups", groupName, fmt.Sprintf("the member: %s is not added: %s", userName, err.Error()))
			}
		}
	}
	return nil
}

func (run *migrationRun) importKind(kind string) error {
	kindReport := &MigrationKindReport{Kind: kind}
	run.report.Kinds = append(run.report.Kinds, kindReport)
	seen := map[string]bool{}

	cursor := ""
	for page := 0; page < maxMigrationPages; page++ {
		res, err := action.ListImport(run.module, run.config, &action.ImportListRequest{
			Kind:         kind,
			Cursor:       cursor,
			Organization: run.migration.Organization,
			Files:        run.files,
		})
		if err != nil {
			return err
		}

		run.addWarnings(res.Warnings)
		kindReport.Total += len(res.Items)
		err = run.importItems(kind, kindReport, res.Items, seen)
		if err != nil {
			return err
		}
		err = run.save(run.isApply)
		if err != nil {
			return err
		}

		if res.Cursor == "" || res.Cursor == cursor {
			break
		}
		cursor = res.Cursor
	}
	return nil
}

// checkItem counts an object that can't or won't be added, it returns true for a new one
func (run *migrationRun) checkItem(kind string, kindReport *MigrationKindReport, seen map[string]bool, name string, reason string, exists bool) bool {
	if reason == "" && name == "" {
		reason = "no name"
	}
	if reason == "" && seen[name] {
		reason = "the name appears more than once"
	}
	if reason != "" {
		kindReport.Invalid++
		run.addIssue(kind, name, reason)
		return false
	}
	seen[name] = true

	if exists {
		kindReport.Exists++
		return false
	}
	kindReport.New++
	return true
}

func (run *migrationRun) recordAdded(kind string, kindReport *MigrationKindReport, owner string, name string) {
	kindReport.Added++
	run.migration.Objects = append(run.migration.Objects, fmt.Sprintf("%s:%s/%s", kind, owner, name))
}

func (run *migrationRun) recordFailed(kind string, kindReport *MigrationKindReport, name string, err error) {
	kindReport.Failed++
	reason := "not added"
	if err != nil {
		reason = err.Error()
	}
	run.addIssue(kind, name, reason)
}

// hasField tells whether the action set a field, so that a missing bool keeps Casdoor's default rather than false
func hasField(raw json.RawMessage, field string) bool {
	fields := map[string]json.RawMessage{}
	err := json.Unmarshal(raw, &fields)
	if err != nil {
		return false
	}
	_, ok := fields[field]
	return ok
}

func (run *migrationRun) getOrganizationIds(ids []string) []string {
	res := []string{}
	for _, id := range ids {
		if id == "" {
			continue
		}
		if !strings.Contains(id, "/") {
			id = util.GetId(run.organization.Name, id)
		}
		res = append(res, id)
	}
	return res
}

func (run *migrationRun) importItems(kind string, kindReport *MigrationKindReport, items []json.RawMessage, seen map[string]bool) error {
	switch kind {
	case "certs":
		return run.importCerts(kindReport, items, seen)
	case "groups":
		return run.importGroups(kindReport, items, seen)
	case "applications":
		return run.importApplications(kindReport, items, seen)
	case "users":
		return run.importUsers(kindReport, items, seen)
	case "roles":
		return run.importRoles(kindReport, items, seen)
	case "permissions":
		return run.importPermissions(kindReport, items, seen)
	}
	return nil
}

func (run *migrationRun) importCerts(kindReport *MigrationKindReport, items []json.RawMessage, seen map[string]bool) error {
	owner := run.organization.Name
	for _, raw := range items {
		cert := &Cert{}
		reason := ""
		if err := json.Unmarshal(raw, cert); err != nil {
			reason = err.Error()
		} else if cert.Certificate == "" || cert.PrivateKey == "" {
			reason = "no certificate or private key"
		}

		existing, err := getCert(owner, cert.Name)
		if err != nil {
			return err
		}
		if !run.checkItem("certs", kindReport, seen, cert.Name, reason, existing != nil) || !run.isApply {
			continue
		}

		cert.Owner = owner
		if cert.CreatedTime == "" {
			cert.CreatedTime = util.GetCurrentTime()
		}
		if cert.DisplayName == "" {
			cert.DisplayName = cert.Name
		}
		if cert.Scope == "" {
			cert.Scope = "JWT"
		}
		if cert.Type == "" {
			cert.Type = "x509"
		}
		if cert.CryptoAlgorithm == "" {
			cert.CryptoAlgorithm = "RS256"
		}

		affected, err := AddCert(cert)
		if err != nil || !affected {
			run.recordFailed("certs", kindReport, cert.Name, err)
			continue
		}
		run.recordAdded("certs", kindReport, owner, cert.Name)
	}
	return nil
}

func (run *migrationRun) importGroups(kindReport *MigrationKindReport, items []json.RawMessage, seen map[string]bool) error {
	owner := run.organization.Name
	newGroups := []*Group{}
	for _, raw := range items {
		group := &Group{}
		reason := ""
		if err := json.Unmarshal(raw, group); err != nil {
			reason = err.Error()
		} else if err = checkGroupName(group.Name); err != nil && group.Name != "" {
			reason = err.Error()
		}

		existing, err := getGroup(owner, group.Name)
		if err != nil {
			return err
		}
		isNew := run.checkItem("groups", kindReport, seen, group.Name, reason, existing != nil)
		if reason == "" && group.Name != "" && len(group.Users) > 0 {
			run.groupMembers[group.Name] = append(run.groupMembers[group.Name], group.Users...)
		}
		if !isNew || !run.isApply {
			continue
		}

		group.Owner = owner
		if group.ParentId == "" || group.ParentId == owner {
			group.ParentId = owner
			group.IsTopGroup = true
		} else {
			group.IsTopGroup = false
		}
		if group.Type == "" {
			group.Type = "Virtual"
		}
		if !hasField(raw, "isEnabled") {
			group.IsEnabled = true
		}
		if group.CreatedTime == "" {
			group.CreatedTime = util.GetCurrentTime()
		}
		group.UpdatedTime = group.CreatedTime
		group.Users = nil
		group.Children = nil
		newGroups = append(newGroups, group)
	}

	if len(newGroups) == 0 {
		return nil
	}
	_, err := AddGroupsInBatch(newGroups)
	for _, group := range newGroups {
		if err != nil {
			run.recordFailed("groups", kindReport, group.Name, err)
		} else {
			run.recordAdded("groups", kindReport, owner, group.Name)
		}
	}
	return nil
}

func (run *migrationRun) importApplications(kindReport *MigrationKindReport, items []json.RawMessage, seen map[string]bool) error {
	for _, raw := range items {
		application := &Application{}
		reason := ""
		if err := json.Unmarshal(raw, application); err != nil {
			reason = err.Error()
		} else if application.ClientId != "" {
			other, err := GetApplicationByClientId(application.ClientId)
			if err != nil {
				return err
			}
			if other != nil && !(other.Owner == "admin" && other.Name == application.Name) {
				reason = fmt.Sprintf("the client ID is used by the application: %s", other.Name)
			}
		}

		existing, err := getApplication("admin", application.Name)
		if err != nil {
			return err
		}
		if existing != nil && existing.Organization != run.organization.Name && reason == "" {
			reason = fmt.Sprintf("an application of the organization: %s has this name", existing.Organization)
		}
		if !run.checkItem("applications", kindReport, seen, application.Name, reason, existing != nil) || !run.isApply {
			continue
		}

		application.Owner = "admin"
		application.Organization = run.organization.Name
		application.IsShared = false
		if !hasField(raw, "enablePassword") {
			application.EnablePassword = true
		}
		affected, err := AddApplication(application, run.lang)
		if err != nil || !affected {
			run.recordFailed("applications", kindReport, application.Name, err)
			continue
		}
		run.recordAdded("applications", kindReport, "admin", application.Name)
	}
	return nil
}

func (run *migrationRun) importUsers(kindReport *MigrationKindReport, items []json.RawMessage, seen map[string]bool) error {
	owner := run.organization.Name
	newUsers := []*User{}
	for _, raw := range items {
		user := &User{}
		reason := ""
		if err := json.Unmarshal(raw, user); err != nil {
			reason = err.Error()
		}
		user.Name = strings.TrimSpace(user.Name)

		passwordType := user.PasswordType
		if reason == "" && passwordType != "" && passwordType != "plugin" && cred.GetCredManager(passwordType) == nil {
			reason = fmt.Sprintf("Casdoor doesn't know the password type: %s, the action should return it as \"plugin\"", passwordType)
		}

		existing, err := getUser(owner, user.Name)
		if err != nil {
			return err
		}
		if existing == nil && reason == "" && user.Id != "" {
			other, err := GetUserByUserId(owner, user.Id)
			if err != nil {
				return err
			}
			if other != nil {
				reason = fmt.Sprintf("the ID is used by the user: %s", other.Name)
			}
		}
		if !run.checkItem("users", kindReport, seen, user.Name, reason, existing != nil) || !run.isApply {
			continue
		}

		user.Owner = owner
		if user.Id == "" {
			user.Id = util.GenerateId()
		}
		if user.Type == "" {
			user.Type = "normal-user"
		}
		if user.CreatedTime == "" {
			user.CreatedTime = util.GetCurrentTime()
		}
		if user.Avatar == "" {
			user.Avatar = run.organization.DefaultAvatar
		}
		if user.BalanceCurrency == "" {
			user.BalanceCurrency = run.organization.BalanceCurrency
			if user.BalanceCurrency == "" {
				user.BalanceCurrency = "USD"
			}
		}
		if passwordType == "plugin" {
			user.PasswordType = pluginPasswordTypePrefix + run.migration.Plugin
		} else if passwordType == "" || passwordType == "plain" {
			user.UpdateUserPassword(run.organization)
		}
		user.Groups = run.getOrganizationIds(user.Groups)
		newUsers = append(newUsers, user)
	}

	if len(newUsers) == 0 {
		return nil
	}
	_, err := AddUsersInBatch(newUsers)
	for _, user := range newUsers {
		if err != nil {
			run.recordFailed("users", kindReport, user.Name, err)
		} else {
			run.recordAdded("users", kindReport, owner, user.Name)
		}
	}
	return nil
}

func (run *migrationRun) importRoles(kindReport *MigrationKindReport, items []json.RawMessage, seen map[string]bool) error {
	owner := run.organization.Name
	for _, raw := range items {
		role := &Role{}
		reason := ""
		if err := json.Unmarshal(raw, role); err != nil {
			reason = err.Error()
		}

		existing, err := getRole(owner, role.Name)
		if err != nil {
			return err
		}
		if !run.checkItem("roles", kindReport, seen, role.Name, reason, existing != nil) || !run.isApply {
			continue
		}

		role.Owner = owner
		if role.CreatedTime == "" {
			role.CreatedTime = util.GetCurrentTime()
		}
		if !hasField(raw, "isEnabled") {
			role.IsEnabled = true
		}
		role.Users = run.getOrganizationIds(role.Users)
		role.Groups = run.getOrganizationIds(role.Groups)
		role.Roles = run.getOrganizationIds(role.Roles)

		affected, err := AddRole(role)
		if err != nil || !affected {
			run.recordFailed("roles", kindReport, role.Name, err)
			continue
		}
		run.recordAdded("roles", kindReport, owner, role.Name)
	}
	return nil
}

func (run *migrationRun) importPermissions(kindReport *MigrationKindReport, items []json.RawMessage, seen map[string]bool) error {
	owner := run.organization.Name
	for _, raw := range items {
		permission := &Permission{}
		reason := ""
		if err := json.Unmarshal(raw, permission); err != nil {
			reason = err.Error()
		}

		existing, err := getPermission(owner, permission.Name)
		if err != nil {
			return err
		}
		if !run.checkItem("permissions", kindReport, seen, permission.Name, reason, existing != nil) || !run.isApply {
			continue
		}

		permission.Owner = owner
		if permission.CreatedTime == "" {
			permission.CreatedTime = util.GetCurrentTime()
		}
		if !hasField(raw, "isEnabled") {
			permission.IsEnabled = true
		}
		if permission.Effect == "" {
			permission.Effect = "Allow"
		}
		if permission.ResourceType == "" {
			permission.ResourceType = "Application"
		}
		permission.Users = run.getOrganizationIds(permission.Users)
		permission.Groups = run.getOrganizationIds(permission.Groups)
		permission.Roles = run.getOrganizationIds(permission.Roles)

		affected, err := AddPermission(permission)
		if err != nil || !affected {
			run.recordFailed("permissions", kindReport, permission.Name, err)
			continue
		}
		run.recordAdded("permissions", kindReport, owner, permission.Name)
	}
	return nil
}

func (run *migrationRun) undo() error {
	objects := run.migration.Objects
	byKind := map[string][]string{}
	for _, object := range objects {
		kind, id, ok := strings.Cut(object, ":")
		if ok {
			byKind[kind] = append(byKind[kind], id)
		}
	}

	remaining := []string{}
	for i := len(migrationKinds) - 1; i >= 0; i-- {
		kind := migrationKinds[i]
		ids := byKind[kind]
		if len(ids) == 0 {
			continue
		}
		kindReport := &MigrationKindReport{Kind: kind, Total: len(ids)}
		run.report.Kinds = append(run.report.Kinds, kindReport)

		// a group with child groups can't be deleted, so the groups are deleted in rounds from the leaves up
		for round := 0; round < 100 && len(ids) > 0; round++ {
			failed := []string{}
			var lastErr error
			for _, id := range ids {
				err := deleteMigrationObject(kind, id)
				if err != nil {
					failed = append(failed, id)
					lastErr = err
					continue
				}
				kindReport.Deleted++
			}
			if kind != "groups" || len(failed) == len(ids) {
				for _, id := range failed {
					kindReport.Failed++
					reason := "not deleted"
					if lastErr != nil {
						reason = lastErr.Error()
					}
					run.addIssue(kind, id, reason)
					remaining = append(remaining, fmt.Sprintf("%s:%s", kind, id))
				}
				break
			}
			ids = failed
		}
	}

	run.migration.Objects = remaining
	if len(remaining) > 0 {
		return fmt.Errorf("%d objects could not be deleted, they are listed in the issues", len(remaining))
	}
	return nil
}

func deleteMigrationObject(kind string, id string) error {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return err
	}

	switch kind {
	case "certs":
		_, err = DeleteCert(&Cert{Owner: owner, Name: name})
	case "groups":
		group, err2 := getGroup(owner, name)
		if err2 != nil || group == nil {
			return err2
		}
		_, err = DeleteGroup(group)
	case "applications":
		application, err2 := getApplication(owner, name)
		if err2 != nil || application == nil {
			return err2
		}
		_, err = DeleteApplication(application)
	case "users":
		user, err2 := getUser(owner, name)
		if err2 != nil || user == nil {
			return err2
		}
		_, err = DeleteUser(user)
	case "roles":
		role, err2 := getRole(owner, name)
		if err2 != nil || role == nil {
			return err2
		}
		_, err = DeleteRole(role)
	case "permissions":
		permission, err2 := getPermission(owner, name)
		if err2 != nil || permission == nil {
			return err2
		}
		_, err = DeletePermission(permission)
	}
	return err
}

func isPluginPasswordType(passwordType string) bool {
	return strings.HasPrefix(passwordType, pluginPasswordTypePrefix)
}

// isImportedPasswordCorrect asks the import action that brought the user in to check a password hash in a format
// Casdoor doesn't know, such as one of Keycloak's
func isImportedPasswordCorrect(user *User, passwordType string, password string) bool {
	name := strings.TrimPrefix(passwordType, pluginPasswordTypePrefix)
	m, _, err := GetInstalledAction(user.Owner, name, "import")
	if err != nil {
		logs.Warn("the password of the user: %s can't be checked: %s", user.GetId(), err.Error())
		return false
	}

	valid, err := action.VerifyImportedPassword(m, password, user.Password, user.PasswordSalt)
	if err != nil {
		logs.Warn("the password of the user: %s can't be checked: %s", user.GetId(), err.Error())
		return false
	}
	return valid
}
