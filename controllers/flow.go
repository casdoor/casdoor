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
	"fmt"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// requireOrganizationAdmin checks that the current user administers the organization in the query parameter "owner",
// then returns that organization
func (c *ApiController) requireOrganizationAdmin() (string, bool) {
	owner := c.Ctx.Input.Query("owner")
	if owner == "" || !c.IsAdminOfOrganization(owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return "", false
	}
	return owner, true
}

func (c *ApiController) getListPageParams() (int, string, string, string, string) {
	limit := util.ParseInt(c.Ctx.Input.Query("pageSize"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return limit, c.Ctx.Input.Query("field"), c.Ctx.Input.Query("value"), c.Ctx.Input.Query("sortField"), c.Ctx.Input.Query("sortOrder")
}

// requireFlow checks the license and that the current user administers the organization of the flow in the query
// parameter "id", then returns the flow
func (c *ApiController) requireFlow() (*object.Flow, bool) {
	id := c.Ctx.Input.Query("id")
	owner, _, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if !c.IsAdminOfOrganization(owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	flow, err := object.GetFlow(id)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if flow == nil {
		c.ResponseError(fmt.Sprintf(c.T("enterprise:The flow: %s doesn't exist"), id))
		return nil, false
	}
	return flow, true
}

// readFlow reads the flow of the request body and checks that the current user administers its organization
func (c *ApiController) readFlow() (*object.Flow, bool) {
	var flow object.Flow
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &flow); err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if flow.Owner == "" || !c.IsAdminOfOrganization(flow.Owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	if flow.Name == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": name")
		return nil, false
	}
	if message := object.CheckFlow(&flow); message != "" {
		c.ResponseError(message)
		return nil, false
	}
	return &flow, true
}

// GetFlows
// @Title GetFlows
// @Tag Flow API
// @Description list the flows of an organization
// @Param   owner     query    string  true        "The organization"
// @Success 200 {array} object.Flow The Response object
// @router /get-flows [get]
func (c *ApiController) GetFlows() {
	owner, ok := c.requireOrganizationAdmin()
	if !ok {
		return
	}
	limit, field, value, sortField, sortOrder := c.getListPageParams()
	count, err := object.GetFlowCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	flows, err := object.GetPaginationFlows(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(flows, paginator.Nums())
}

// GetFlow
// @Title GetFlow
// @Tag Flow API
// @Description get a flow
// @Param   id     query    string  true        "The id ( owner/name ) of the flow"
// @Success 200 {object} object.Flow The Response object
// @router /get-flow [get]
func (c *ApiController) GetFlow() {
	flow, ok := c.requireFlow()
	if !ok {
		return
	}
	c.ResponseOk(flow)
}

// AddFlow
// @Title AddFlow
// @Tag Flow API
// @Description add a flow
// @Param   body    body   object.Flow  true        "The flow"
// @Success 200 {object} controllers.Response The Response object
// @router /add-flow [post]
func (c *ApiController) AddFlow() {
	flow, ok := c.readFlow()
	if !ok {
		return
	}
	c.Data["json"] = WrapActionResponse(object.AddFlow(flow))
	c.ServeJSON()
}

// UpdateFlow
// @Title UpdateFlow
// @Tag Flow API
// @Description update a flow
// @Param   id      query  string   true        "The id ( owner/name ) of the flow"
// @Param   body    body   object.Flow  true        "The flow"
// @Success 200 {object} controllers.Response The Response object
// @router /update-flow [post]
func (c *ApiController) UpdateFlow() {
	existing, ok := c.requireFlow()
	if !ok {
		return
	}
	flow, ok := c.readFlow()
	if !ok {
		return
	}
	if flow.Owner != existing.Owner {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.Data["json"] = WrapActionResponse(object.UpdateFlow(existing.GetId(), flow))
	c.ServeJSON()
}

// DeleteFlow
// @Title DeleteFlow
// @Tag Flow API
// @Description delete a flow and its runs
// @Param   id     query    string  true        "The id ( owner/name ) of the flow"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-flow [post]
func (c *ApiController) DeleteFlow() {
	flow, ok := c.requireFlow()
	if !ok {
		return
	}
	c.Data["json"] = WrapActionResponse(object.DeleteFlow(flow))
	c.ServeJSON()
}

// GetFlowRuns
// @Title GetFlowRuns
// @Tag Flow API
// @Description list the runs of the flows of an organization, or of one flow, the newest first
// @Param   owner     query    string  true        "The organization"
// @Param   flow      query    string  false       "The name of the flow"
// @Success 200 {array} object.FlowRun The Response object
// @router /get-flow-runs [get]
func (c *ApiController) GetFlowRuns() {
	owner, ok := c.requireOrganizationAdmin()
	if !ok {
		return
	}
	flow := c.Ctx.Input.Query("flow")
	limit, field, value, sortField, sortOrder := c.getListPageParams()
	count, err := object.GetFlowRunCount(owner, flow, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	runs, err := object.GetPaginationFlowRuns(owner, flow, paginator.Offset(), limit, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(runs, paginator.Nums())
}

type testFlowForm struct {
	Flow        *object.Flow `json:"flow"`
	User        string       `json:"user"`
	Application string       `json:"application"`
	Ip          string       `json:"ip"`
	RiskScore   int          `json:"riskScore"`
	RiskSignals []string     `json:"riskSignals"`
}

// TestFlow
// @Title TestFlow
// @Tag Flow API
// @Description run a flow, saved or not, on a user without changing anything
// @Param   body    body   controllers.testFlowForm  true        "The flow, the user and the request to test with"
// @Success 200 {object} object.FlowRun The Response object
// @router /test-flow [post]
func (c *ApiController) TestFlow() {
	var form testFlowForm
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &form); err != nil {
		c.ResponseError(err.Error())
		return
	}
	if form.Flow == nil || form.Flow.Owner == "" || !c.IsAdminOfOrganization(form.Flow.Owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	if message := object.CheckFlow(form.Flow); message != "" {
		c.ResponseError(message)
		return
	}
	user, err := object.GetUser(util.GetId(form.Flow.Owner, form.User))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(form.Flow.Owner, form.User)))
		return
	}
	ip := form.Ip
	if ip == "" {
		ip = util.GetClientIpFromRequest(c.Ctx.Request)
	}
	risk := &object.SigninRisk{Score: form.RiskScore, Signals: form.RiskSignals, Decision: "Allow"}
	if risk.Signals == nil {
		risk.Signals = []string{}
	}
	run, result := object.TestFlow(form.Flow, user, form.Application, ip, risk)
	c.ResponseOk(run, result)
}
