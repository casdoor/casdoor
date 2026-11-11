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

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// getProvisioningApplication returns the application of the "id" query when the current user administers its
// organization
func (c *ApiController) getProvisioningApplication() (*object.Application, bool) {
	application, err := object.GetApplication(c.Ctx.Input.Query("id"))
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if application == nil {
		c.ResponseError(c.T("check:Application does not exist"))
		return nil, false
	}

	user := c.getCurrentUser()
	if user == nil || !(c.IsGlobalAdmin() || (user.IsAdmin && user.Owner == application.Organization)) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	return application, true
}

// TestProvisioning
// @Title TestProvisioning
// @Tag Application API
// @Description check that the system behind the application can be reached with the provisioning settings in the body
// @Param   id   query string                    true "The id ( owner/name ) of the application"
// @Param   body body  object.ProvisioningConfig false "The provisioning settings to test, the saved ones when empty"
// @Success 200 {object} controllers.Response The Response object
// @router /test-provisioning [post]
func (c *ApiController) TestProvisioning() {
	application, ok := c.getProvisioningApplication()
	if !ok {
		return
	}

	if len(c.Ctx.Input.RequestBody) > 0 {
		config := &object.ProvisioningConfig{}
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, config); err != nil {
			c.ResponseError(err.Error())
			return
		}
		if config.Token == "" && application.Provisioning != nil {
			config.Token = application.Provisioning.Token
		}
		application.Provisioning = config
	}

	if err := object.TestProvisioning(application); err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk()
}

// SyncProvisioning
// @Title SyncProvisioning
// @Tag Application API
// @Description queue every user of the application's organization and every provisioned account to be brought up to date
// @Param   id query string true "The id ( owner/name ) of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /sync-provisioning [post]
func (c *ApiController) SyncProvisioning() {
	application, ok := c.getProvisioningApplication()
	if !ok {
		return
	}

	count, err := object.ReconcileProvisioning(application)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(count)
}

// RetryProvisioning
// @Title RetryProvisioning
// @Tag Application API
// @Description queue the failed provisioning tasks of the application again
// @Param   id query string true "The id ( owner/name ) of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /retry-provisioning [post]
func (c *ApiController) RetryProvisioning() {
	application, ok := c.getProvisioningApplication()
	if !ok {
		return
	}

	count, err := object.RetryFailedProvisioningTasks(application.Name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(count)
}

// GetProvisionedAccounts
// @Title GetProvisionedAccounts
// @Tag Application API
// @Description get the accounts provisioned to the system behind the application
// @Param   id       query string true  "The id ( owner/name ) of the application"
// @Param   state    query string false "Active, Disabled or Deleted"
// @Param   p        query string false "The page"
// @Param   pageSize query string false "The page size"
// @Success 200 {array} object.ProvisionedAccount The Response object
// @router /get-provisioned-accounts [get]
func (c *ApiController) GetProvisionedAccounts() {
	application, ok := c.getProvisioningApplication()
	if !ok {
		return
	}

	limit := util.ParseInt(c.Ctx.Input.Query("pageSize"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	page := util.ParseInt(c.Ctx.Input.Query("p"))
	if page <= 0 {
		page = 1
	}

	accounts, count, err := object.GetProvisionedAccounts(application.Name, c.Ctx.Input.Query("state"), (page-1)*limit, limit)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	c.ResponseOk(accounts, paginator.Nums())
}

// GetProvisioningTasks
// @Title GetProvisioningTasks
// @Tag Application API
// @Description get the recent provisioning tasks of the application
// @Param   id query string true "The id ( owner/name ) of the application"
// @Success 200 {array} object.ProvisioningTask The Response object
// @router /get-provisioning-tasks [get]
func (c *ApiController) GetProvisioningTasks() {
	application, ok := c.getProvisioningApplication()
	if !ok {
		return
	}

	tasks, err := object.GetProvisioningTasks(application.Name, 0, 50)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	summary, err := object.GetProvisioningSummary(application.Name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(tasks, summary)
}
