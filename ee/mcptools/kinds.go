// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package mcptools

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/util"
)

// kind is a kind of object that the tools manage through its REST APIs: get-<plural>, get-<name>, add-<name>,
// update-<name>?id= and delete-<name>
type kind struct {
	Name   string
	Plural string

	ReadScope  string
	WriteScope string

	List   apiAction
	Get    apiAction
	Add    apiAction
	Update apiAction
	Delete apiAction

	// the query of get-<plural> for the objects of an organization
	ListQuery func(organization string) url.Values
	// the organization an object belongs to
	OrganizationOf func(obj map[string]interface{}) string
	// the fields that the tools don't hand to agents
	Secrets []string
	// the fields that get-<plural> lists
	Summary []string
}

func ownerQuery(organization string) url.Values {
	return url.Values{"owner": {organization}}
}

func ownerOrganization(obj map[string]interface{}) string {
	owner := getString(obj, "owner")
	if owner == "admin" {
		return "built-in"
	}
	return owner
}

var kinds = map[string]*kind{
	"application": {
		Name: "application", Plural: "applications",
		ReadScope: "application:read", WriteScope: "application:write",
		Get: (*controllers.ApiController).GetApplication, Add: (*controllers.ApiController).AddApplication,
		Update: (*controllers.ApiController).UpdateApplication, Delete: (*controllers.ApiController).DeleteApplication,
		OrganizationOf: func(obj map[string]interface{}) string { return getString(obj, "organization") },
		Secrets:        []string{"clientSecret", "clientCert", "registrationAccessToken"},
	},
	"user": {
		Name: "user", Plural: "users",
		ReadScope: "user:read", WriteScope: "user:write",
		Get: (*controllers.ApiController).GetUser, Add: (*controllers.ApiController).AddUser,
		Update: (*controllers.ApiController).UpdateUser, Delete: (*controllers.ApiController).DeleteUser,
		OrganizationOf: func(obj map[string]interface{}) string { return getString(obj, "owner") },
		Secrets:        []string{"password", "passwordSalt", "accessSecret", "totpSecret", "recoveryCodes"},
	},
	"organization": {
		Name: "organization", Plural: "organizations",
		ReadScope: "organization:read", WriteScope: "organization:write",
		List: (*controllers.ApiController).GetOrganizations, Get: (*controllers.ApiController).GetOrganization,
		Add: (*controllers.ApiController).AddOrganization, Update: (*controllers.ApiController).UpdateOrganization,
		Delete:         (*controllers.ApiController).DeleteOrganization,
		ListQuery:      func(string) url.Values { return url.Values{"owner": {"admin"}} },
		OrganizationOf: func(obj map[string]interface{}) string { return getString(obj, "name") },
		Secrets:        []string{"masterPassword", "defaultPassword", "masterVerificationCode"},
		Summary:        []string{"name", "displayName", "websiteUrl", "passwordType", "defaultApplication"},
	},
	"provider": {
		Name: "provider", Plural: "providers",
		ReadScope: "provider:read", WriteScope: "provider:write",
		List: (*controllers.ApiController).GetProviders, Get: (*controllers.ApiController).GetProvider,
		Add: (*controllers.ApiController).AddProvider, Update: (*controllers.ApiController).UpdateProvider,
		Delete:         (*controllers.ApiController).DeleteProvider,
		ListQuery:      ownerQuery,
		OrganizationOf: ownerOrganization,
		Secrets:        []string{"clientSecret", "clientSecret2"},
		Summary:        []string{"owner", "name", "displayName", "category", "type", "clientId"},
	},
	"role": {
		Name: "role", Plural: "roles",
		ReadScope: "role:read", WriteScope: "role:write",
		List: (*controllers.ApiController).GetRoles, Get: (*controllers.ApiController).GetRole,
		Add: (*controllers.ApiController).AddRole, Update: (*controllers.ApiController).UpdateRole,
		Delete:         (*controllers.ApiController).DeleteRole,
		ListQuery:      ownerQuery,
		OrganizationOf: ownerOrganization,
		Summary:        []string{"owner", "name", "displayName", "users", "groups", "roles", "isEnabled"},
	},
	"permission": {
		Name: "permission", Plural: "permissions",
		ReadScope: "permission:read", WriteScope: "permission:write",
		List: (*controllers.ApiController).GetPermissions, Get: (*controllers.ApiController).GetPermission,
		Add: (*controllers.ApiController).AddPermission, Update: (*controllers.ApiController).UpdatePermission,
		Delete:         (*controllers.ApiController).DeletePermission,
		ListQuery:      ownerQuery,
		OrganizationOf: ownerOrganization,
		Summary:        []string{"owner", "name", "displayName", "users", "groups", "roles", "resourceType", "resources", "actions", "effect", "isEnabled"},
	},
	"group": {
		Name: "group", Plural: "groups",
		ReadScope: "group:read", WriteScope: "group:write",
		List: (*controllers.ApiController).GetGroups, Get: (*controllers.ApiController).GetGroup,
		Add: (*controllers.ApiController).AddGroup, Update: (*controllers.ApiController).UpdateGroup,
		Delete:         (*controllers.ApiController).DeleteGroup,
		ListQuery:      ownerQuery,
		OrganizationOf: ownerOrganization,
		Summary:        []string{"owner", "name", "displayName", "type", "parentId", "isEnabled"},
	},
	"cert": {
		Name: "cert", Plural: "certs",
		ReadScope: "cert:read", WriteScope: "cert:write",
		List: (*controllers.ApiController).GetCerts, Get: (*controllers.ApiController).GetCert,
		Add: (*controllers.ApiController).AddCert, Update: (*controllers.ApiController).UpdateCert,
		Delete:         (*controllers.ApiController).DeleteCert,
		ListQuery:      ownerQuery,
		OrganizationOf: ownerOrganization,
		Secrets:        []string{"privateKey"},
		Summary:        []string{"owner", "name", "displayName", "scope", "type", "cryptoAlgorithm", "expireInYears"},
	},
	"token": {
		Name: "token", Plural: "tokens",
		ReadScope: "token:read", WriteScope: "token:write",
		List: (*controllers.ApiController).GetTokens, Get: (*controllers.ApiController).GetToken,
		Delete: (*controllers.ApiController).DeleteToken,
		ListQuery: func(organization string) url.Values {
			return url.Values{"owner": {"admin"}, "organization": {organization}}
		},
		OrganizationOf: func(obj map[string]interface{}) string { return getString(obj, "organization") },
		Secrets:        []string{"code", "accessToken", "refreshToken", "idToken", "accessTokenHash", "refreshTokenHash"},
		Summary:        []string{"owner", "name", "application", "organization", "user", "createdTime", "expiresIn", "scope"},
	},
}

func getString(obj map[string]interface{}, key string) string {
	value, _ := obj[key].(string)
	return value
}

func getObjectId(obj map[string]interface{}) string {
	return util.GetId(getString(obj, "owner"), getString(obj, "name"))
}

// mask hides the secrets of an object before a tool hands it to an agent
func (k *kind) mask(obj map[string]interface{}) map[string]interface{} {
	masked := map[string]interface{}{}
	for key, value := range obj {
		masked[key] = value
	}
	for _, key := range k.Secrets {
		if text, ok := masked[key].(string); ok && text != "" {
			masked[key] = "***"
		}
	}
	return masked
}

func (k *kind) summarize(obj map[string]interface{}) map[string]interface{} {
	if len(k.Summary) == 0 {
		return k.mask(obj)
	}
	summary := map[string]interface{}{}
	for _, key := range k.Summary {
		if value, ok := obj[key]; ok && value != nil && value != "" {
			summary[key] = value
		}
	}
	return summary
}

// client makes the REST API requests of a tool call for its user
type client struct {
	ctx    *context.Context
	userId string
}

func (cl *client) call(method string, path string, query url.Values, body interface{}, action apiAction) (*apiResponse, error) {
	resp, err := callApi(cl.ctx, cl.userId, method, path, query, body, action)
	if err != nil {
		return nil, err
	}
	return resp, resp.err()
}

// get returns an object by its ID ("owner/name"), or nil when there is none
func (cl *client) get(k *kind, id string) (map[string]interface{}, error) {
	query := url.Values{"id": {id}}
	if k.Name == "application" {
		query.Set("withKey", "1")
	}
	resp, err := cl.call("GET", "/api/get-"+k.Name, query, nil, k.Get)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}
	if err = json.Unmarshal(resp.Data, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (cl *client) mustGet(k *kind, id string) (map[string]interface{}, error) {
	obj, err := cl.get(k, id)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, fmt.Errorf("the %s: %s does not exist", k.Name, id)
	}
	return obj, nil
}

func (cl *client) list(k *kind, organization string, page int, pageSize int) ([]map[string]interface{}, int, error) {
	query := k.ListQuery(organization)
	query.Set("p", fmt.Sprintf("%d", page))
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	resp, err := cl.call("GET", "/api/get-"+k.Plural, query, nil, k.List)
	if err != nil {
		return nil, 0, err
	}

	objs := []map[string]interface{}{}
	if err = json.Unmarshal(resp.Data, &objs); err != nil {
		return nil, 0, err
	}
	total := len(objs)
	_ = json.Unmarshal(resp.Data2, &total)
	return objs, total, nil
}

// save adds, updates or deletes an object, id is the object's ID before an update
func (cl *client) save(k *kind, action string, id string, obj map[string]interface{}) (bool, error) {
	var resp *apiResponse
	var err error
	switch action {
	case actionAdd:
		resp, err = cl.call("POST", "/api/add-"+k.Name, nil, obj, k.Add)
	case actionUpdate:
		resp, err = cl.call("POST", "/api/update-"+k.Name, url.Values{"id": {id}}, obj, k.Update)
	case actionDelete:
		resp, err = cl.call("POST", "/api/delete-"+k.Name, nil, obj, k.Delete)
	default:
		return false, fmt.Errorf("unknown action: %s", action)
	}
	if err != nil {
		return false, err
	}
	return resp.isAffected(), nil
}

// diffObjects lists the top-level fields that differ, e.g. "clientId: \"a\" -> \"b\""
func diffObjects(before map[string]interface{}, after map[string]interface{}, secrets []string) []string {
	keys := map[string]bool{}
	for key := range before {
		keys[key] = true
	}
	for key := range after {
		keys[key] = true
	}

	lines := []string{}
	for _, key := range sortedKeys(keys) {
		oldText, newText := formatValue(before[key]), formatValue(after[key])
		if oldText == newText || isIgnoredField(key) {
			continue
		}
		if util.InSlice(secrets, key) {
			oldText, newText = maskedText(oldText), maskedText(newText)
		}
		// a new object lists its fields that have a value
		if len(before) == 0 {
			if newText != `""` && newText != "[]" && newText != "null" && newText != "{}" {
				lines = append(lines, fmt.Sprintf("%s: %s", key, truncate(newText, 200)))
			}
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s -> %s", key, truncate(oldText, 200), truncate(newText, 200)))
	}
	return lines
}

func isIgnoredField(key string) bool {
	return key == "createdTime" || key == "updatedTime"
}

func maskedText(text string) string {
	if text == "" || text == `""` || text == "null" {
		return text
	}
	return `"***"`
}

func formatValue(value interface{}) string {
	if value == nil {
		return "null"
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func sortedKeys(keys map[string]bool) []string {
	list := make([]string, 0, len(keys))
	for key := range keys {
		list = append(list, key)
	}
	sort.Strings(list)
	return list
}
