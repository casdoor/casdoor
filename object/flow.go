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
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	FlowTriggerPreSignin  = "pre-signin"
	FlowTriggerPostSignin = "post-signin"
	FlowTriggerPreSignup  = "pre-signup"
	FlowTriggerPostSignup = "post-signup"
	FlowTriggerPreToken   = "pre-token"

	FlowStepCondition = "condition"
	FlowStepAction    = "action"

	FlowActionDeny            = "deny"
	FlowActionRequireMfa      = "require-mfa"
	FlowActionAddClaim        = "add-claim"
	FlowActionSetUserField    = "set-user-field"
	FlowActionAddToGroup      = "add-to-group"
	FlowActionRemoveFromGroup = "remove-from-group"
	FlowActionAddToRole       = "add-to-role"
	FlowActionRemoveFromRole  = "remove-from-role"
	FlowActionHttp            = "http"
	FlowActionSendEmail       = "send-email"
	FlowActionStop            = "stop"

	FlowRunCompleted = "Completed"
	FlowRunStopped   = "Stopped"
	FlowRunDenied    = "Denied"
	FlowRunFailed    = "Failed"
)

// FlowTriggers are the points of the sign-in, the sign-up and the token issuance that a flow runs at, in this order
var FlowTriggers = []string{FlowTriggerPreSignup, FlowTriggerPostSignup, FlowTriggerPreSignin, FlowTriggerPostSignin, FlowTriggerPreToken}

// flowTriggerActions are the actions that make sense at each trigger, a flow can't deny a sign-in that is done
var flowTriggerActions = map[string][]string{
	FlowTriggerPreSignup:  {FlowActionDeny, FlowActionSetUserField, FlowActionAddToGroup, FlowActionHttp, FlowActionStop},
	FlowTriggerPostSignup: {FlowActionSetUserField, FlowActionAddToGroup, FlowActionRemoveFromGroup, FlowActionAddToRole, FlowActionRemoveFromRole, FlowActionHttp, FlowActionSendEmail, FlowActionStop},
	FlowTriggerPreSignin:  {FlowActionDeny, FlowActionRequireMfa, FlowActionSetUserField, FlowActionAddToGroup, FlowActionRemoveFromGroup, FlowActionAddToRole, FlowActionRemoveFromRole, FlowActionHttp, FlowActionSendEmail, FlowActionStop},
	FlowTriggerPostSignin: {FlowActionSetUserField, FlowActionAddToGroup, FlowActionRemoveFromGroup, FlowActionAddToRole, FlowActionRemoveFromRole, FlowActionHttp, FlowActionSendEmail, FlowActionStop},
	FlowTriggerPreToken:   {FlowActionDeny, FlowActionAddClaim, FlowActionHttp, FlowActionStop},
}

// FlowStep is a condition that the steps after it need, or an action
type FlowStep struct {
	Id       string            `json:"id"`
	Type     string            `json:"type"`
	Disabled bool              `json:"disabled"`
	Field    string            `json:"field"`
	Operator string            `json:"operator"`
	Value    string            `json:"value"`
	Action   string            `json:"action"`
	Params   map[string]string `json:"params"`
}

// Flow is a no-code automation of an organization: at a trigger, its steps run one by one, a condition that doesn't
// hold ends the flow
type Flow struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"mediumtext" json:"description"`

	Trigger      string      `xorm:"varchar(100)" json:"trigger"`
	Applications []string    `xorm:"mediumtext" json:"applications"`
	Order        int         `json:"order"`
	IsEnabled    bool        `json:"isEnabled"`
	Steps        []*FlowStep `xorm:"mediumtext" json:"steps"`

	RunCount    int    `json:"runCount"`
	LastRunTime string `xorm:"varchar(100)" json:"lastRunTime"`
}

func (flow *Flow) GetId() string {
	return fmt.Sprintf("%s/%s", flow.Owner, flow.Name)
}

// FlowStepRun is how a step went in a run
type FlowStepRun struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
	Result  string `json:"result"`
	Message string `json:"message"`
}

// FlowRun is a run of a flow, kept for the admins to see what it did
type FlowRun struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100) index" json:"createdTime"`

	Flow        string         `xorm:"varchar(100) index" json:"flow"`
	Trigger     string         `xorm:"varchar(100)" json:"trigger"`
	User        string         `xorm:"varchar(100)" json:"user"`
	Application string         `xorm:"varchar(100)" json:"application"`
	Status      string         `xorm:"varchar(20)" json:"status"`
	DurationMs  int            `json:"durationMs"`
	IsTest      bool           `json:"isTest"`
	Steps       []*FlowStepRun `xorm:"mediumtext" json:"steps"`
	Error       string         `xorm:"mediumtext" json:"error"`
}

// FlowResult is what the flows of a trigger decided together
type FlowResult struct {
	Deny       string                 `json:"deny"`
	RequireMfa bool                   `json:"requireMfa"`
	Claims     map[string]interface{} `json:"claims"`

	stop bool
}

// FlowEnv is what a flow runs on
type FlowEnv struct {
	Trigger     string
	User        *User
	Application string
	Ip          string
	UserAgent   string
	Country     string
	Risk        *SigninRisk
	Claims      map[string]interface{}
	// a test run doesn't change anything, its actions only say what they would do
	DryRun bool
	// a pre-sign-up flow changes the user that is about to be added, nothing is saved
	Unsaved bool
}

// SigninRisk is what adaptive risk found in a sign-in, for the flows to look at
type SigninRisk struct {
	Score    int      `json:"score"`
	Signals  []string `json:"signals"`
	Decision string   `json:"decision"`
	Location string   `json:"location"`
}

func GetFlowCount(owner string, field string, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Flow{})
}

func GetPaginationFlows(owner string, offset int, limit int, field string, value string, sortField string, sortOrder string) ([]*Flow, error) {
	flows := []*Flow{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&flows)
	return flows, err
}

func GetFlow(id string) (*Flow, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	flow := Flow{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&flow)
	if err != nil || !existed {
		return nil, err
	}
	return &flow, nil
}

// CheckFlow returns what is wrong with a flow that is about to be saved, "" when nothing
func CheckFlow(flow *Flow) string {
	allowed, ok := flowTriggerActions[flow.Trigger]
	if !ok {
		return fmt.Sprintf("unknown trigger: %s", flow.Trigger)
	}
	for i, step := range flow.Steps {
		switch step.Type {
		case FlowStepCondition:
			if step.Field == "" || step.Operator == "" {
				return fmt.Sprintf("step %d: a condition needs a field and an operator", i+1)
			}
			if !util.InSlice(flowOperators, step.Operator) {
				return fmt.Sprintf("step %d: unknown operator: %s", i+1, step.Operator)
			}
		case FlowStepAction:
			if !util.InSlice(allowed, step.Action) {
				return fmt.Sprintf("step %d: the action: %s can't run at the trigger: %s", i+1, step.Action, flow.Trigger)
			}
			if message := checkFlowActionParams(step); message != "" {
				return fmt.Sprintf("step %d: %s", i+1, message)
			}
		default:
			return fmt.Sprintf("step %d: unknown type: %s", i+1, step.Type)
		}
	}
	return ""
}

func checkFlowActionParams(step *FlowStep) string {
	need := func(keys ...string) string {
		for _, key := range keys {
			if strings.TrimSpace(step.Params[key]) == "" {
				return fmt.Sprintf("the action: %s needs %s", step.Action, key)
			}
		}
		return ""
	}
	switch step.Action {
	case FlowActionAddClaim:
		if message := need("name"); message != "" {
			return message
		}
		if isReservedClaim(step.Params["name"]) {
			return fmt.Sprintf("the claim: %s is set by Casdoor and can't be changed", step.Params["name"])
		}
	case FlowActionSetUserField:
		if message := need("field"); message != "" {
			return message
		}
		if !isFlowUserFieldWritable(step.Params["field"]) {
			return fmt.Sprintf("the user field: %s can't be set by a flow", step.Params["field"])
		}
	case FlowActionAddToGroup, FlowActionRemoveFromGroup:
		return need("group")
	case FlowActionAddToRole, FlowActionRemoveFromRole:
		return need("role")
	case FlowActionHttp:
		if message := need("url"); message != "" {
			return message
		}
		if !strings.HasPrefix(step.Params["url"], "https://") && !strings.HasPrefix(step.Params["url"], "http://") {
			return "the URL must start with http:// or https://"
		}
	case FlowActionSendEmail:
		return need("subject", "content")
	}
	return ""
}

func AddFlow(flow *Flow) (bool, error) {
	now := util.GetCurrentTime()
	if flow.CreatedTime == "" {
		flow.CreatedTime = now
	}
	flow.UpdatedTime = now
	affected, err := ormer.Engine.Insert(flow)
	clearFlowCache(flow.Owner)
	return affected != 0, err
}

func UpdateFlow(id string, flow *Flow) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	flow.UpdatedTime = util.GetCurrentTime()
	affected, err := ormer.Engine.ID(core.PK{owner, name}).
		Cols("name", "display_name", "description", "trigger", "applications", "order", "is_enabled", "steps", "updated_time").
		Update(flow)
	clearFlowCache(owner)
	clearFlowCache(flow.Owner)
	return affected != 0, err
}

func DeleteFlow(flow *Flow) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{flow.Owner, flow.Name}).Delete(&Flow{})
	if err != nil {
		return false, err
	}
	_, err = ormer.Engine.Where("owner = ? and flow = ?", flow.Owner, flow.Name).Delete(&FlowRun{})
	clearFlowCache(flow.Owner)
	return affected != 0, err
}

func GetFlowRunCount(owner string, flow string, field string, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	if flow != "" {
		session = session.And("flow = ?", flow)
	}
	return session.Count(&FlowRun{})
}

func GetPaginationFlowRuns(owner string, flow string, offset int, limit int, field string, value string, sortField string, sortOrder string) ([]*FlowRun, error) {
	runs := []*FlowRun{}
	if sortField == "" {
		sortField, sortOrder = "created_time", "descend"
	}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	if flow != "" {
		session = session.And("flow = ?", flow)
	}
	err := session.Find(&runs)
	return runs, err
}

type cachedFlows struct {
	flows     []*Flow
	fetchTime time.Time
}

var flowCache sync.Map

func clearFlowCache(owner string) {
	flowCache.Delete(owner)
}

// getEnabledFlows returns the enabled flows of the organization at the trigger for the application, in their order
func getEnabledFlows(owner string, trigger string, application string) []*Flow {
	var flows []*Flow
	if value, ok := flowCache.Load(owner); ok && time.Since(value.(*cachedFlows).fetchTime) < time.Minute {
		flows = value.(*cachedFlows).flows
	} else {
		flows = []*Flow{}
		if err := ormer.Engine.Where("owner = ? and is_enabled = ?", owner, true).Find(&flows); err != nil {
			logs.Warn("the flows of the organization: %s are not read: %s", owner, err.Error())
			return nil
		}
		flowCache.Store(owner, &cachedFlows{flows: flows, fetchTime: time.Now()})
	}

	res := []*Flow{}
	for _, flow := range flows {
		if flow.Trigger != trigger {
			continue
		}
		if len(flow.Applications) > 0 && application != "" && !util.InSlice(flow.Applications, application) {
			continue
		}
		res = append(res, flow)
	}
	sort.SliceStable(res, func(i, j int) bool {
		if res[i].Order != res[j].Order {
			return res[i].Order < res[j].Order
		}
		return res[i].Name < res[j].Name
	})
	return res
}

// RunFlows runs the enabled flows of the organization at the trigger and merges what they decide, a denial stops the
// flows after it
func RunFlows(env *FlowEnv) *FlowResult {
	if env.User == nil {
		return nil
	}
	flows := getEnabledFlows(env.User.Owner, env.Trigger, env.Application)
	if len(flows) == 0 {
		return nil
	}

	result := &FlowResult{Claims: map[string]interface{}{}}
	for _, flow := range flows {
		run := runFlow(flow, env, result)
		saveFlowRun(flow, run)
		if result.Deny != "" {
			break
		}
	}
	return result
}

// RunSigninFlows runs the flows of a sign-in trigger for the request, the post-sign-in ones in the background
func RunSigninFlows(user *User, application string, clientIp string, request *http.Request, trigger string, risk *SigninRisk) *FlowResult {
	if user == nil {
		return nil
	}
	env := &FlowEnv{
		Trigger:     trigger,
		User:        user,
		Application: application,
		Ip:          clientIp,
		Risk:        risk,
	}
	if request != nil {
		env.UserAgent = request.UserAgent()
		env.Country = GetIpDbCountry(clientIp)
	}
	if trigger == FlowTriggerPostSignin || trigger == FlowTriggerPostSignup {
		go RunFlows(env)
		return nil
	}
	return RunFlows(env)
}

func saveFlowRun(flow *Flow, run *FlowRun) {
	if _, err := ormer.Engine.Insert(run); err != nil {
		logs.Warn("the run of the flow: %s is not saved: %s", flow.GetId(), err.Error())
	}
	_, err := ormer.Engine.Exec("update flow set run_count = run_count + 1, last_run_time = ? where owner = ? and name = ?", run.CreatedTime, flow.Owner, flow.Name)
	if err != nil {
		logs.Warn("the run count of the flow: %s is not updated: %s", flow.GetId(), err.Error())
	}
}

func runFlow(flow *Flow, env *FlowEnv, result *FlowResult) *FlowRun {
	start := time.Now()
	run := &FlowRun{
		Owner:       flow.Owner,
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		Flow:        flow.Name,
		Trigger:     env.Trigger,
		User:        env.User.Name,
		Application: env.Application,
		Status:      FlowRunCompleted,
		IsTest:      env.DryRun,
		Steps:       []*FlowStepRun{},
	}
	defer func() {
		run.DurationMs = int(time.Since(start).Milliseconds())
	}()

	stopped := false
	for _, step := range flow.Steps {
		stepRun := &FlowStepRun{Id: step.Id, Type: step.Type, Summary: summarizeFlowStep(step)}
		run.Steps = append(run.Steps, stepRun)
		if stopped || step.Disabled {
			stepRun.Result = "skipped"
			continue
		}

		if step.Type == FlowStepCondition {
			actual := getFlowFieldValue(env, step.Field)
			ok := evaluateFlowCondition(actual, step.Operator, renderFlowTemplate(step.Value, env))
			stepRun.Message = fmt.Sprintf("%s = %s", step.Field, formatFlowValue(actual))
			if ok {
				stepRun.Result = "passed"
			} else {
				stepRun.Result = "failed"
				run.Status = FlowRunStopped
				stopped = true
			}
			continue
		}

		message, err := runFlowAction(step, env, result)
		stepRun.Message = message
		if err != nil {
			stepRun.Result = "error"
			stepRun.Message = err.Error()
			run.Status = FlowRunFailed
			run.Error = err.Error()
			stopped = true
			continue
		}
		stepRun.Result = "done"
		if step.Action == FlowActionStop || result.stop {
			result.stop = false
			run.Status = FlowRunStopped
			stopped = true
		}
		if result.Deny != "" {
			run.Status = FlowRunDenied
			stopped = true
		}
	}
	return run
}

func summarizeFlowStep(step *FlowStep) string {
	if step.Type == FlowStepCondition {
		return strings.TrimSpace(fmt.Sprintf("%s %s %s", step.Field, step.Operator, step.Value))
	}
	switch step.Action {
	case FlowActionDeny:
		return fmt.Sprintf("deny: %s", step.Params["message"])
	case FlowActionAddClaim:
		return fmt.Sprintf("claim %s = %s", step.Params["name"], step.Params["value"])
	case FlowActionSetUserField:
		return fmt.Sprintf("user.%s = %s", step.Params["field"], step.Params["value"])
	case FlowActionAddToGroup, FlowActionRemoveFromGroup:
		return fmt.Sprintf("%s %s", step.Action, step.Params["group"])
	case FlowActionAddToRole, FlowActionRemoveFromRole:
		return fmt.Sprintf("%s %s", step.Action, step.Params["role"])
	case FlowActionHttp:
		method := step.Params["method"]
		if method == "" {
			method = http.MethodPost
		}
		return fmt.Sprintf("%s %s", method, step.Params["url"])
	case FlowActionSendEmail:
		return fmt.Sprintf("email: %s", step.Params["subject"])
	}
	return step.Action
}

// getFlowTokenClaims runs the pre-token flows of the user's organization for the token, a denial refuses the token
func getFlowTokenClaims(application *Application, user *User, scope string) (map[string]interface{}, error) {
	if user == nil || application == nil {
		return nil, nil
	}
	claims := map[string]interface{}{"scope": scope}
	result := RunFlows(&FlowEnv{Trigger: FlowTriggerPreToken, User: user, Application: application.Name, Claims: claims})
	if result == nil {
		return nil, nil
	}
	if result.Deny != "" {
		return nil, fmt.Errorf("%s", result.Deny)
	}
	return result.Claims, nil
}

// RunSignupFlows runs the pre-sign-up flows on the user that is about to be added, or the post-sign-up ones in
// the background, and returns the denial of a pre-sign-up flow
func RunSignupFlows(user *User, application string, clientIp string, request *http.Request, trigger string) string {
	if user == nil {
		return ""
	}
	env := &FlowEnv{Trigger: trigger, User: user, Application: application, Ip: clientIp, Unsaved: trigger == FlowTriggerPreSignup}
	if request != nil {
		env.UserAgent = request.UserAgent()
		env.Country = GetIpDbCountry(clientIp)
	}
	if trigger == FlowTriggerPostSignup {
		go RunFlows(env)
		return ""
	}
	result := RunFlows(env)
	if result == nil {
		return ""
	}
	return result.Deny
}

// TestFlow runs a flow on a user without changing anything, unless its HTTP steps are told to call in tests
func TestFlow(flow *Flow, user *User, application string, clientIp string, risk *SigninRisk) (*FlowRun, *FlowResult) {
	env := &FlowEnv{
		Trigger:     flow.Trigger,
		User:        user,
		Application: application,
		Ip:          clientIp,
		Risk:        risk,
		DryRun:      true,
		Unsaved:     true,
	}
	if flow.Trigger == FlowTriggerPreToken {
		env.Claims = map[string]interface{}{"scope": "openid profile"}
	}
	result := &FlowResult{Claims: map[string]interface{}{}}
	run := runFlow(flow, env, result)
	return run, result
}
