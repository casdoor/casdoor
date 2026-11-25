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

package mcpself

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// how long a tool waits for a migration run before it tells the agent to look again with get_migration
const migrationWaitTime = 50 * time.Second

// MigrationArgs are the arguments of the migration tools, with the names agents also use for them
type MigrationArgs struct {
	Organization string
	Name         string
	DisplayName  string
	Action       string
	Host         string
	ClientId     string
	ClientSecret string
	Metadata     string
}

func (args *MigrationArgs) UnmarshalJSON(data []byte) error {
	raw := map[string]json.RawMessage{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	get := func(keys ...string) string {
		for _, key := range keys {
			value, ok := raw[key]
			if !ok {
				continue
			}
			var text string
			if json.Unmarshal(value, &text) == nil {
				if text != "" {
					return strings.TrimSpace(text)
				}
				continue
			}
			// metadata given as a JSON object
			return string(value)
		}
		return ""
	}

	args.Organization = get("organization", "owner")
	args.Name = get("name", "migration")
	if id := get("id"); args.Name == "" && id != "" {
		_, args.Name = util.GetOwnerAndNameFromIdNoCheck(id)
	}
	args.DisplayName = get("displayName", "display_name")
	args.Action = get("action", "plugin", "importAction", "import_action")
	args.Host = get("host", "domain")
	args.ClientId = get("clientId", "client_id")
	args.ClientSecret = get("clientSecret", "client_secret")
	args.Metadata = get("metadata")
	return nil
}

func getMigrationTools() []McpTool {
	organization := map[string]interface{}{
		"type":        "string",
		"description": "The organization that the users go to, not built-in",
	}
	name := map[string]interface{}{
		"type":        "string",
		"description": "The name of the migration",
	}

	return []McpTool{
		{
			Name:        "get_migrations",
			Description: "List the migrations of an organization and the import actions installed in it. A migration moves the users, groups, roles, applications and signing keys of another identity system, such as Auth0 or Keycloak, into a Casdoor organization through an import action from the Marketplace, keeping the user IDs, password hashes and client IDs and secrets. Use it instead of adding the objects one by one.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"organization": organization},
				"required":   []string{"organization"},
			},
		},
		{
			Name:        "get_migration",
			Description: "Get the settings, state and report of a migration: for each kind of object how many the other system has, how many would be or were added, how many exist already and which can't be imported and why",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"organization": organization, "name": name},
				"required":   []string{"organization", "name"},
			},
		},
		{
			Name:        "plan_migration",
			Description: "Create or update a migration and preview it: the import action reads the other identity system and the report tells what would be added, without changing anything. Show the report to the user before applying it. Actions that read export files, such as Keycloak Migration, need the files uploaded on the migration page of the Casdoor admin console instead.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"organization": organization,
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The name of the migration, <organization>-<action> when empty",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "The import action installed in the organization, such as auth0-migration, see get_migrations. It can be left out when only one is installed",
					},
					"host": map[string]interface{}{
						"type":        "string",
						"description": "The host of the other system, such as the Auth0 tenant domain your-tenant.us.auth0.com",
					},
					"clientId": map[string]interface{}{
						"type":        "string",
						"description": "The client ID the action reads the other system with, such as the one of an Auth0 machine to machine application",
					},
					"clientSecret": map[string]interface{}{
						"type":        "string",
						"description": "The client secret that goes with the client ID, left out to keep the saved one",
					},
					"metadata": map[string]interface{}{
						"type":        "string",
						"description": "Settings of the action as JSON, such as {\"realm\": \"myrealm\"} for Keycloak",
					},
				},
				"required": []string{"organization"},
			},
		},
		{
			Name:        "apply_migration",
			Description: "Apply a previewed migration: add the objects of the other system that don't exist in the organization yet. Call it only after the user has seen the report of plan_migration and agreed. Applying again later only adds what is new.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"organization": organization, "name": name},
				"required":   []string{"organization", "name"},
			},
		},
		{
			Name:        "undo_migration",
			Description: "Undo a migration: delete every object it added, including the users who have signed in since. It can't be undone, call it only when the user asks for it.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"organization": organization, "name": name},
				"required":   []string{"organization", "name"},
			},
		},
	}
}

func (c *McpController) callMigrationTool(id interface{}, name string, args MigrationArgs) {
	if args.Organization == "" {
		c.SendToolErrorResult(id, "the organization is required")
		return
	}
	if !c.isOrganizationAllowed(args.Organization) {
		c.SendToolErrorResult(id, "unauthorized operation")
		return
	}

	var text string
	var err error
	switch name {
	case "get_migrations":
		text, err = getMigrationsText(args.Organization)
	case "get_migration":
		text, err = c.getMigrationText(args)
	case "plan_migration":
		text, err = c.planMigration(args)
	case "apply_migration":
		text, err = c.runMigration(args, "apply")
	case "undo_migration":
		text, err = c.runMigration(args, "undo")
	}
	if err != nil {
		c.SendToolErrorResult(id, err.Error())
		return
	}
	c.SendToolResult(id, text)
}

func getMigrationsText(organization string) (string, error) {
	migrations, err := object.GetMaskedMigrations(object.GetOrganizationMigrations("admin", organization))
	if err != nil {
		return "", err
	}
	actions, err := object.GetImportActions(organization)
	if err != nil {
		return "", err
	}

	type actionItem struct {
		Name          string `json:"name"`
		DisplayName   string `json:"displayName"`
		IntegrationId string `json:"integrationId"`
		Version       string `json:"version"`
	}
	type migrationItem struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Action      string `json:"action"`
		Host        string `json:"host"`
		State       string `json:"state"`
		UpdatedTime string `json:"updatedTime"`
	}
	result := struct {
		ImportActions []actionItem    `json:"importActions"`
		Migrations    []migrationItem `json:"migrations"`
		Note          string          `json:"note,omitempty"`
	}{ImportActions: []actionItem{}, Migrations: []migrationItem{}}
	for _, a := range actions {
		result.ImportActions = append(result.ImportActions, actionItem{Name: a.Name, DisplayName: a.DisplayName, IntegrationId: a.IntegrationId, Version: a.Version})
	}
	for _, m := range migrations {
		result.Migrations = append(result.Migrations, migrationItem{Name: m.Name, DisplayName: m.DisplayName, Action: m.Plugin, Host: m.Host, State: m.State, UpdatedTime: m.UpdatedTime})
	}
	if len(actions) == 0 {
		result.Note = fmt.Sprintf("No import action is installed in the organization: %s. An admin installs one in the Casdoor admin console under Identity > Marketplace, such as Auth0 Migration (auth0-migration) or Keycloak Migration (keycloak-migration).", organization)
	}
	return util.StructToJsonFormatted(result), nil
}

func getOrganizationMigration(organization string, name string) (*object.Migration, error) {
	if name == "" {
		return nil, fmt.Errorf("the name of the migration is required, see get_migrations")
	}
	migration, err := object.GetMigrationByOrganization(util.GetId("admin", name), organization)
	if err != nil {
		return nil, err
	}
	if migration == nil {
		return nil, fmt.Errorf("the migration: %s is not found in the organization: %s, see get_migrations", name, organization)
	}
	return migration, nil
}

func (c *McpController) getMigrationText(args MigrationArgs) (string, error) {
	migration, err := getOrganizationMigration(args.Organization, args.Name)
	if err != nil {
		return "", err
	}
	return formatMigration(migration), nil
}

// getImportAction is the import action of the migration, the only one installed when none is given
func getImportAction(organization string, name string) (string, error) {
	actions, err := object.GetImportActions(organization)
	if err != nil {
		return "", err
	}
	names := []string{}
	for _, a := range actions {
		if a.Name == name || (name != "" && a.IntegrationId == name) {
			return a.Name, nil
		}
		names = append(names, a.Name)
	}

	switch {
	case len(actions) == 0:
		return "", fmt.Errorf("no import action is installed in the organization: %s, an admin installs one in the Casdoor admin console under Identity > Marketplace, such as Auth0 Migration or Keycloak Migration", organization)
	case name != "":
		return "", fmt.Errorf("the import action: %s is not installed in the organization: %s, the installed ones are: %s", name, organization, strings.Join(names, ", "))
	case len(actions) > 1:
		return "", fmt.Errorf("choose the action, the import actions installed in the organization: %s are: %s", organization, strings.Join(names, ", "))
	}
	return actions[0].Name, nil
}

func (c *McpController) planMigration(args MigrationArgs) (string, error) {
	if args.Organization == "built-in" {
		return "", fmt.Errorf("choose an organization other than built-in, every user of built-in is a global admin")
	}
	actionName, err := getImportAction(args.Organization, args.Action)
	if err != nil {
		return "", err
	}
	name := args.Name
	if name == "" {
		name = fmt.Sprintf("%s-%s", args.Organization, actionName)
	}

	migration, err := object.GetMigration(util.GetId("admin", name))
	if err != nil {
		return "", err
	}
	if migration != nil && migration.Organization != args.Organization {
		return "", fmt.Errorf("the name: %s is taken by a migration of another organization, choose another name", name)
	}

	if migration == nil {
		migration = &object.Migration{Owner: "admin", Name: name, DisplayName: args.DisplayName, Organization: args.Organization}
		if migration.DisplayName == "" {
			migration.DisplayName = name
		}
		migration.Plugin, migration.Host, migration.ClientId, migration.ClientSecret, migration.Metadata = actionName, args.Host, args.ClientId, args.ClientSecret, args.Metadata
		_, err = object.AddMigration(migration)
		if err != nil {
			return "", err
		}
	} else {
		if args.DisplayName != "" {
			migration.DisplayName = args.DisplayName
		}
		migration.Plugin = actionName
		for _, field := range []struct {
			value  string
			target *string
		}{{args.Host, &migration.Host}, {args.ClientId, &migration.ClientId}, {args.ClientSecret, &migration.ClientSecret}, {args.Metadata, &migration.Metadata}} {
			if field.value != "" {
				*field.target = field.value
			}
		}
		_, err = object.UpdateMigration(migration.GetId(), migration, c.IsGlobalAdmin(), c.GetAcceptLanguage())
		if err != nil {
			return "", err
		}
	}

	return c.startMigration(migration.Organization, migration.Name, "plan")
}

func (c *McpController) runMigration(args MigrationArgs, mode string) (string, error) {
	migration, err := getOrganizationMigration(args.Organization, args.Name)
	if err != nil {
		return "", err
	}
	if mode == "apply" && migration.State != "Previewed" && migration.State != "Applied" {
		return "", fmt.Errorf("preview the migration: %s with plan_migration and show the report to the user before applying it", migration.Name)
	}
	return c.startMigration(migration.Organization, migration.Name, mode)
}

func (c *McpController) startMigration(organization string, name string, mode string) (string, error) {
	// the run reads the saved settings
	migration, err := getOrganizationMigration(organization, name)
	if err != nil {
		return "", err
	}
	err = object.StartMigration(migration, &object.MigrationRunRequest{Mode: mode}, c.GetAcceptLanguage())
	if err != nil {
		return "", err
	}

	deadline := time.Now().Add(migrationWaitTime)
	for {
		time.Sleep(time.Second)
		migration, err = getOrganizationMigration(organization, name)
		if err != nil {
			return "", err
		}
		if migration.State != "Running" {
			break
		}
		if time.Now().After(deadline) {
			return formatMigration(migration) + "\nThe migration is still running. Call get_migration later to see the result.", nil
		}
	}

	text := formatMigration(migration)
	if mode == "plan" && migration.State == "Previewed" {
		text += "\nNothing has been changed. Show this report to the user, and call apply_migration only after they agree."
	}
	return text, nil
}

func formatMigration(migration *object.Migration) string {
	_, _ = object.GetMaskedMigration(migration)

	var b strings.Builder
	fmt.Fprintf(&b, "Migration: %s (%s) into the organization: %s\n", migration.Name, migration.DisplayName, migration.Organization)
	fmt.Fprintf(&b, "Action: %s\n", migration.Plugin)
	if migration.Host != "" {
		fmt.Fprintf(&b, "Host: %s\n", migration.Host)
	}
	state := migration.State
	if state == "" {
		state = "Not run"
	}
	fmt.Fprintf(&b, "State: %s\n", state)

	report := migration.Report
	if report == nil {
		return b.String()
	}
	modes := map[string]string{"test": "Test connection", "plan": "Preview", "apply": "Apply", "undo": "Undo"}
	fmt.Fprintf(&b, "Last run: %s, started %s", modes[report.Mode], report.StartedTime)
	if report.FinishedTime != "" {
		fmt.Fprintf(&b, ", finished %s", report.FinishedTime)
	}
	b.WriteString("\n")
	if report.Error != "" {
		fmt.Fprintf(&b, "Error: %s\n", report.Error)
	}

	if len(report.Kinds) > 0 {
		b.WriteString("\n| Kind | Total | New | Exists | Invalid | Added | Failed | Deleted |\n|---|---|---|---|---|---|---|---|\n")
		for _, k := range report.Kinds {
			fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %d | %d | %d |\n", k.Kind, k.Total, k.New, k.Exists, k.Invalid, k.Added, k.Failed, k.Deleted)
		}
	}
	if len(report.Warnings) > 0 {
		b.WriteString("\nWarnings:\n")
		for _, warning := range report.Warnings {
			fmt.Fprintf(&b, "- %s\n", warning)
		}
	}
	if len(report.Issues) > 0 {
		b.WriteString("\nIssues:\n")
		for _, issue := range report.Issues {
			fmt.Fprintf(&b, "- %s %s: %s\n", issue.Kind, issue.Name, issue.Reason)
		}
	}
	return b.String()
}
