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

package controllers

import (
	"encoding/json"

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetMigrations
// @Title GetMigrations
// @Tag Migration API
// @Description get migrations
// @Param   owner     query    string  true        "The owner of migrations"
// @Param   organization     query    string  true        "The organization of migrations"
// @Success 200 {array} object.Migration The Response object
// @router /get-migrations [get]
func (c *ApiController) GetMigrations() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")
	organization := c.Ctx.Input.Query("organization")

	if limit == "" || page == "" {
		migrations, err := object.GetMaskedMigrations(object.GetOrganizationMigrations(owner, organization))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(migrations)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetMigrationCount(owner, organization, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		migrations, err := object.GetMaskedMigrations(object.GetPaginationMigrations(owner, organization, paginator.Offset(), limit, field, value, sortField, sortOrder))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(migrations, paginator.Nums())
	}
}

func (c *ApiController) getMigration(id string, organization string) (*object.Migration, error) {
	isGlobalAdmin, _ := c.isGlobalAdmin()
	if isGlobalAdmin {
		return object.GetMigration(id)
	}
	return object.GetMigrationByOrganization(id, organization)
}

// GetMigration
// @Title GetMigration
// @Tag Migration API
// @Description get migration
// @Param   id     query    string  true        "The id ( owner/name ) of the migration"
// @Param   organization     query    string  true        "The organization of the migration"
// @Success 200 {object} object.Migration The Response object
// @router /get-migration [get]
func (c *ApiController) GetMigration() {
	id := c.Ctx.Input.Query("id")
	organization := c.Ctx.Input.Query("organization")

	migration, err := object.GetMaskedMigration(c.getMigration(id, organization))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(migration)
}

// UpdateMigration
// @Title UpdateMigration
// @Tag Migration API
// @Description update migration
// @Param   id     query    string  true        "The id ( owner/name ) of the migration"
// @Param   body    body   object.Migration  true        "The details of the migration"
// @Success 200 {object} controllers.Response The Response object
// @router /update-migration [post]
func (c *ApiController) UpdateMigration() {
	id := c.Ctx.Input.Query("id")

	var migration object.Migration
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &migration)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.requireOrganizationPermission(migration.Organization) {
		return
	}

	c.Data["json"] = WrapActionResponse(object.UpdateMigration(id, &migration, c.IsGlobalAdmin(), c.GetAcceptLanguage()))
	c.ServeJSON()
}

// AddMigration
// @Title AddMigration
// @Tag Migration API
// @Description add migration
// @Param   body    body   object.Migration  true        "The details of the migration"
// @Success 200 {object} controllers.Response The Response object
// @router /add-migration [post]
func (c *ApiController) AddMigration() {
	var migration object.Migration
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &migration)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.requireOrganizationPermission(migration.Organization) {
		return
	}

	migration.Owner = "admin"
	c.Data["json"] = WrapActionResponse(object.AddMigration(&migration))
	c.ServeJSON()
}

// DeleteMigration
// @Title DeleteMigration
// @Tag Migration API
// @Description delete migration, the objects it added stay
// @Param   body    body   object.Migration  true        "The details of the migration"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-migration [post]
func (c *ApiController) DeleteMigration() {
	var migration object.Migration
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &migration)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.requireOrganizationPermission(migration.Organization) {
		return
	}

	c.Data["json"] = WrapActionResponse(object.DeleteMigration(&migration))
	c.ServeJSON()
}

// RunMigration
// @Title RunMigration
// @Tag Migration API
// @Description test the connection, preview, apply or undo a migration in the background, the report of the migration shows the progress
// @Param   id     query    string  true        "The id ( owner/name ) of the migration"
// @Param   organization     query    string  true        "The organization of the migration"
// @Param   body    body   object.MigrationRunRequest  true        "The mode: test, plan, apply or undo, and the files to read"
// @Success 200 {object} controllers.Response The Response object
// @router /run-migration [post]
func (c *ApiController) RunMigration() {
	id := c.Ctx.Input.Query("id")
	organization := c.Ctx.Input.Query("organization")

	var req object.MigrationRunRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	migration, err := c.getMigration(id, organization)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if migration == nil {
		c.ResponseError(c.T("general:Missing parameter") + ": id")
		return
	}
	if !c.requireOrganizationPermission(migration.Organization) {
		return
	}

	err = object.StartMigration(migration, &req, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk()
}
