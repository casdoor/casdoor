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

// GetBackups
// @Title GetBackups
// @Tag Backup API
// @Description get backups
// @Param   owner     query    string  true        "The owner of backups"
// @Success 200 {array} object.Backup The Response object
// @router /get-backups [get]
func (c *ApiController) GetBackups() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		backups, err := object.GetMaskedBackups(object.GetBackups(owner))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(backups)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetBackupCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		backups, err := object.GetMaskedBackups(object.GetPaginationBackups(owner, paginator.Offset(), limit, field, value, sortField, sortOrder))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(backups, paginator.Nums())
	}
}

// GetGlobalBackups
// @Title GetGlobalBackups
// @Tag Backup API
// @Description get global backups
// @Success 200 {array} object.Backup The Response object
// @router /get-global-backups [get]
func (c *ApiController) GetGlobalBackups() {
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		backups, err := object.GetMaskedBackups(object.GetGlobalBackups())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(backups)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetGlobalBackupsCount(field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		backups, err := object.GetMaskedBackups(object.GetPaginationGlobalBackups(paginator.Offset(), limit, field, value, sortField, sortOrder))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(backups, paginator.Nums())
	}
}

// GetBackup
// @Title GetBackup
// @Tag Backup API
// @Description get backup
// @Param   id     query    string  true        "The id ( owner/name ) of the backup"
// @Success 200 {object} object.Backup The Response object
// @router /get-backup [get]
func (c *ApiController) GetBackup() {
	id := c.Ctx.Input.Query("id")
	backup, err := object.GetBackup(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(object.GetMaskedBackup(backup))
}

// UpdateBackup
// @Title UpdateBackup
// @Tag Backup API
// @Description update backup
// @Param   id     query    string  true        "The id ( owner/name ) of the backup"
// @Param   body    body   object.Backup  true        "The details of the backup"
// @Success 200 {object} controllers.Response The Response object
// @router /update-backup [post]
func (c *ApiController) UpdateBackup() {
	id := c.Ctx.Input.Query("id")

	var backup object.Backup
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &backup)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateBackup(id, &backup))
	c.ServeJSON()
}

// AddBackup
// @Title AddBackup
// @Tag Backup API
// @Description add backup
// @Param   body    body   object.Backup  true        "The details of the backup"
// @Success 200 {object} controllers.Response The Response object
// @router /add-backup [post]
func (c *ApiController) AddBackup() {
	var backup object.Backup
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &backup)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddBackup(&backup))
	c.ServeJSON()
}

// DeleteBackup
// @Title DeleteBackup
// @Tag Backup API
// @Description delete backup
// @Param   body    body   object.Backup  true        "The details of the backup"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-backup [post]
func (c *ApiController) DeleteBackup() {
	var backup object.Backup
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &backup)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteBackup(&backup))
	c.ServeJSON()
}

// ExecuteBackup
// @Title ExecuteBackup
// @Tag Backup API
// @Description execute a database backup
// @Param   id     query    string  true        "The id ( owner/name ) of the backup"
// @Success 200 {object} controllers.Response The Response object
// @router /execute-backup [post]
func (c *ApiController) ExecuteBackup() {
	id := c.Ctx.Input.Query("id")
	backup, err := object.GetBackup(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if backup == nil {
		c.ResponseError("Backup not found")
		return
	}

	// Execute backup in background
	go func() {
		_ = backup.ExecuteBackup()
	}()

	c.Data["json"] = Response{Status: "ok", Msg: "Backup started", Data: "Backup process started"}
	c.ServeJSON()
}

// RestoreBackup
// @Title RestoreBackup
// @Tag Backup API
// @Description restore a database from backup
// @Param   id     query    string  true        "The id ( owner/name ) of the backup"
// @Success 200 {object} controllers.Response The Response object
// @router /restore-backup [post]
func (c *ApiController) RestoreBackup() {
	id := c.Ctx.Input.Query("id")
	backup, err := object.GetBackup(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if backup == nil {
		c.ResponseError("Backup not found")
		return
	}

	err = backup.RestoreBackup()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = Response{Status: "ok", Msg: "Backup restored successfully", Data: "Backup restored"}
	c.ServeJSON()
}
