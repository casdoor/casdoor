// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package mcptools

import (
	"fmt"
	"sort"
	"strings"

	"github.com/casdoor/casdoor/util"
)

// the kinds that get the generic tools, applications and users have the tools of the server itself
var managedKinds = []string{"organization", "provider", "role", "permission", "group", "cert", "token"}

var idProperty = map[string]interface{}{
	"type":        "string",
	"description": "The ID of the object: \"<owner>/<name>\"",
}

func stringProperty(description string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": description}
}

func getKindTools() []*toolDef {
	defs := []*toolDef{}
	for _, name := range managedKinds {
		k := kinds[name]
		defs = append(defs, getListTool(k), getGetTool(k))
		if k.Add != nil {
			defs = append(defs, getAddTool(k))
		}
		if k.Update != nil {
			defs = append(defs, getUpdateTool(k))
		}
		if k.Delete != nil {
			defs = append(defs, getDeleteTool(k))
		}
	}
	return defs
}

func getListTool(k *kind) *toolDef {
	properties := map[string]interface{}{
		"page":     map[string]interface{}{"type": "integer", "description": "The page, from 1"},
		"pageSize": map[string]interface{}{"type": "integer", "description": "The number of objects in a page, 50 by default and 200 at most"},
	}
	required := []string{}
	description := fmt.Sprintf("List the %s of an organization, with their main fields. Use get_%s for all the fields of one.", k.Plural, k.Name)
	if k.Name == "organization" {
		description = "List the organizations that the user administers, with their main fields. Use get_organization for all the fields of one."
	} else {
		properties["organization"] = stringProperty("The name of the organization")
		required = append(required, "organization")
	}

	return &toolDef{
		Name:        "get_" + k.Plural,
		Description: description,
		Scope:       k.ReadScope,
		Properties:  properties,
		Required:    required,
		Organization: func(t *toolCall) (string, error) {
			return t.getListOrganization(), nil
		},
		Run: func(t *toolCall) (string, error) {
			organization := t.getListOrganization()
			if k.Name != "organization" && organization == "" {
				return "", fmt.Errorf("the argument organization is required")
			}
			page := t.getInt("page", 1, 100000)
			objs, total, err := t.client.list(k, organization, page, t.getInt("pageSize", 50, 200))
			if err != nil {
				return "", err
			}
			items := []map[string]interface{}{}
			for _, obj := range objs {
				items = append(items, k.summarize(obj))
			}
			return toJson(map[string]interface{}{"total": total, "page": page, k.Plural: items}), nil
		},
	}
}

func getGetTool(k *kind) *toolDef {
	return &toolDef{
		Name:        "get_" + k.Name,
		Description: fmt.Sprintf("Get all the fields of a %s, the secrets are masked as \"***\".", k.Name),
		Scope:       k.ReadScope,
		Properties:  map[string]interface{}{"id": idProperty},
		Required:    []string{"id"},
		Organization: func(t *toolCall) (string, error) {
			return t.getKindOrganization(k, t.getString("id")), nil
		},
		Run: func(t *toolCall) (string, error) {
			id, err := t.requireString("id")
			if err != nil {
				return "", err
			}
			obj, err := t.client.mustGet(k, id)
			if err != nil {
				return "", err
			}
			return toJson(k.mask(obj)), nil
		},
	}
}

func getAddTool(k *kind) *toolDef {
	description := fmt.Sprintf("Add a %s. Give the whole object as in the REST API add-%s, with \"owner\" (the organization) and \"name\"; get an existing %s with get_%s to see the fields.", k.Name, k.Name, k.Name, k.Name)
	if k.Name == "organization" {
		description = "Add an organization. Give the whole object as in the REST API add-organization, with \"owner\": \"admin\" and \"name\"; get an existing organization with get_organization to see the fields."
	}

	return &toolDef{
		Name:        "add_" + k.Name,
		Description: description,
		Scope:       k.WriteScope,
		Write:       true,
		Properties:  map[string]interface{}{k.Name: map[string]interface{}{"type": "object", "description": "The " + k.Name + " to add"}},
		Required:    []string{k.Name},
		Organization: func(t *toolCall) (string, error) {
			obj, err := t.getObject(k.Name)
			if err != nil {
				return "", nil
			}
			return k.organizationOf(nil, obj, ""), nil
		},
		Run: func(t *toolCall) (string, error) {
			obj, err := t.getObject(k.Name)
			if err != nil {
				return "", err
			}
			if getString(obj, "owner") == "" || getString(obj, "name") == "" {
				return "", fmt.Errorf("the %s needs \"owner\" and \"name\"", k.Name)
			}
			if getString(obj, "createdTime") == "" {
				obj["createdTime"] = util.GetCurrentTime()
			}
			if err = t.apply(k, actionAdd, "", nil, obj); err != nil {
				return "", err
			}
			if t.dryRun {
				return "", nil
			}
			return fmt.Sprintf("Added the %s %s", k.Name, getObjectId(obj)), nil
		},
	}
}

func getUpdateTool(k *kind) *toolDef {
	return &toolDef{
		Name:        "update_" + k.Name,
		Description: fmt.Sprintf("Change some fields of a %s and keep the others. A masked secret (\"***\") stays as it is.", k.Name),
		Scope:       k.WriteScope,
		Write:       true,
		Properties: map[string]interface{}{
			"id":     idProperty,
			"fields": map[string]interface{}{"type": "object", "description": fmt.Sprintf("The fields to change and their new values, named as in get_%s", k.Name)},
		},
		Required: []string{"id", "fields"},
		Organization: func(t *toolCall) (string, error) {
			return t.getKindOrganization(k, t.getString("id")), nil
		},
		Run: func(t *toolCall) (string, error) {
			id, err := t.requireString("id")
			if err != nil {
				return "", err
			}
			fields, err := t.getUpdateFields(k)
			if err != nil {
				return "", err
			}
			before, err := t.client.mustGet(k, id)
			if err != nil {
				return "", err
			}

			after := map[string]interface{}{}
			for key, value := range before {
				after[key] = value
			}
			unknown := []string{}
			for key, value := range fields {
				if _, ok := before[key]; !ok {
					unknown = append(unknown, key)
					continue
				}
				after[key] = value
			}
			if len(unknown) > 0 {
				sort.Strings(unknown)
				return "", fmt.Errorf("the %s has no field: %s, see get_%s for its fields", k.Name, strings.Join(unknown, ", "), k.Name)
			}

			if err = t.apply(k, actionUpdate, id, before, after); err != nil {
				return "", err
			}
			if t.dryRun {
				return "", nil
			}
			return fmt.Sprintf("Updated the %s %s", k.Name, getObjectId(after)), nil
		},
	}
}

func getDeleteTool(k *kind) *toolDef {
	return &toolDef{
		Name:        "delete_" + k.Name,
		Description: fmt.Sprintf("Delete a %s.", k.Name),
		Scope:       k.WriteScope,
		Write:       true,
		Delete:      true,
		Properties:  map[string]interface{}{"id": idProperty},
		Required:    []string{"id"},
		Organization: func(t *toolCall) (string, error) {
			return t.getKindOrganization(k, t.getString("id")), nil
		},
		Run: func(t *toolCall) (string, error) {
			id, err := t.requireString("id")
			if err != nil {
				return "", err
			}
			before, err := t.client.mustGet(k, id)
			if err != nil {
				return "", err
			}
			if err = t.apply(k, actionDelete, id, before, nil); err != nil {
				return "", err
			}
			if t.dryRun {
				return "", nil
			}
			return fmt.Sprintf("Deleted the %s %s", k.Name, id), nil
		},
	}
}

// getListOrganization is the organization to list the objects of, agents often call it "owner" as in the objects
func (t *toolCall) getListOrganization() string {
	if organization := t.getString("organization"); organization != "" {
		return organization
	}
	return t.getString("owner")
}

// getUpdateFields is the fields to change, also taken from the object named after the kind, as the update tools of
// applications and users take it
func (t *toolCall) getUpdateFields(k *kind) (map[string]interface{}, error) {
	if _, ok := t.args["fields"]; !ok {
		if _, ok = t.args[k.Name]; ok {
			return t.getObject(k.Name)
		}
	}
	return t.getObject("fields")
}

// getKindOrganization is the organization of the object with the ID, without loading it when the ID tells
func (t *toolCall) getKindOrganization(k *kind, id string) string {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	switch k.Name {
	case "organization":
		return name
	case "application", "token":
		obj, err := t.client.get(k, id)
		if err != nil || obj == nil {
			return ""
		}
		return k.OrganizationOf(obj)
	}
	if owner == "admin" {
		return "built-in"
	}
	return owner
}
