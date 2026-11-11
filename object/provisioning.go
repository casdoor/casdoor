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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	ProvisioningTypeScim     = "SCIM"
	ProvisioningTypeDatabase = "Database"

	ProvisioningDeprovisionDisable = "Disable"
	ProvisioningDeprovisionDelete  = "Delete"

	ProvisionedAccountActive   = "Active"
	ProvisionedAccountDisabled = "Disabled"
	ProvisionedAccountDeleted  = "Deleted"

	ProvisioningTaskPending  = "Pending"
	ProvisioningTaskRetrying = "Retrying"
	ProvisioningTaskDone     = "Done"
	ProvisioningTaskFailed   = "Failed"

	provisioningMaxAttempts = 8
)

// ProvisioningConfig pushes the accounts of an application's users to the system behind it: created when a user can
// use the application, updated when the user changes, and disabled or deleted when the user leaves, is forbidden or no
// longer belongs to the chosen groups. A downstream system with a SCIM 2.0 API is the target itself, a legacy one is
// reached through a database syncer whose connection and column mapping write its user table.
type ProvisioningConfig struct {
	IsEnabled bool   `json:"isEnabled"`
	Type      string `json:"type"`

	// SCIM: the base URL ending before "/Users" and the bearer token of the SCIM API
	Url   string `json:"url"`
	Token string `json:"token"`
	// SCIM: the user field that becomes "userName": "Name", "Email", "Phone" or "Id"
	UserNameField string `json:"userNameField"`

	// Database: the database syncer ("<organization>/<name>") whose connection and table columns are used
	Syncer string `json:"syncer"`

	// only the users in these groups (or the groups under them), every user of the organization when empty
	Groups []string `json:"groups"`
	// "Disable" or "Delete" the account of a user who leaves
	Deprovision string `json:"deprovision"`
	// reconcile every user every so many minutes, 0 for only on changes and on demand
	SyncIntervalMinutes int    `json:"syncIntervalMinutes"`
	LastSyncTime        string `json:"lastSyncTime"`
}

// ProvisionedAccount is the account of a user in the system behind an application
type ProvisionedAccount struct {
	Application  string `xorm:"varchar(100) notnull pk" json:"application"`
	UserId       string `xorm:"varchar(100) notnull pk" json:"userId"`
	Organization string `xorm:"varchar(100) index" json:"organization"`
	User         string `xorm:"varchar(200)" json:"user"`
	CreatedTime  string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime  string `xorm:"varchar(100)" json:"updatedTime"`

	RemoteId  string `xorm:"varchar(200)" json:"remoteId"`
	State     string `xorm:"varchar(100) index" json:"state"`
	Hash      string `xorm:"varchar(100)" json:"hash"`
	LastError string `xorm:"mediumtext" json:"lastError"`
}

// ProvisioningTask is a user whose account in an application's system has to be brought up to date
type ProvisioningTask struct {
	Id           int64  `xorm:"pk autoincr" json:"id"`
	Application  string `xorm:"varchar(100) index" json:"application"`
	Organization string `xorm:"varchar(100)" json:"organization"`
	UserId       string `xorm:"varchar(100) index" json:"userId"`
	User         string `xorm:"varchar(200)" json:"user"`
	CreatedTime  string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime  string `xorm:"varchar(100)" json:"updatedTime"`

	State     string `xorm:"varchar(100) index" json:"state"`
	Attempts  int    `json:"attempts"`
	NextTime  string `xorm:"varchar(100) index" json:"nextTime"`
	Action    string `xorm:"varchar(100)" json:"action"`
	LastError string `xorm:"mediumtext" json:"lastError"`
}

func (config *ProvisioningConfig) isActive() bool {
	return config != nil && config.IsEnabled
}

func (config *ProvisioningConfig) getFeature() (string, string) {
	if config.Type == ProvisioningTypeDatabase {
		return edition.Enterprise, edition.FeatureDbProvisioning
	}
	return edition.Commercial, edition.FeatureScimProvisioning
}

// CheckProvisioningConfig validates the provisioning of an application before it is saved
func CheckProvisioningConfig(application *Application) error {
	config := application.Provisioning
	if !config.isActive() {
		return nil
	}

	required, feature := config.getFeature()
	if err := edition.Require(required, feature); err != nil {
		return err
	}

	switch config.Type {
	case ProvisioningTypeScim:
		if config.Url == "" {
			return fmt.Errorf("the SCIM URL of the provisioning is empty")
		}
	case ProvisioningTypeDatabase:
		syncer, err := getProvisioningSyncer(application, config)
		if err != nil {
			return err
		}
		if syncer.getKeyColumnForProvisioning() == nil {
			return fmt.Errorf("the syncer: %s has no table columns", syncer.GetId())
		}
		if config.Deprovision != ProvisioningDeprovisionDelete && syncer.getForbiddenColumn() == nil {
			return fmt.Errorf("to disable the accounts, map a column of the syncer: %s to IsForbidden, or delete the accounts instead", syncer.GetId())
		}
	default:
		return fmt.Errorf("unknown provisioning type: %s", config.Type)
	}
	return nil
}

var (
	provisioningApplicationCache   sync.Map
	provisioningApplicationTimeout = 30 * time.Second
)

type cachedProvisioningApplications struct {
	applications []*Application
	fetchTime    time.Time
}

func ClearProvisioningApplicationCache() {
	provisioningApplicationCache = sync.Map{}
}

// getProvisioningApplications returns the applications of an organization that provision accounts, cached because
// every change of a user asks for them
func getProvisioningApplications(organization string) ([]*Application, error) {
	if value, ok := provisioningApplicationCache.Load(organization); ok {
		cached := value.(*cachedProvisioningApplications)
		if time.Since(cached.fetchTime) < provisioningApplicationTimeout {
			return cached.applications, nil
		}
	}

	applications := []*Application{}
	err := ormer.Engine.Where("organization = ?", organization).Find(&applications)
	if err != nil {
		return nil, err
	}

	res := []*Application{}
	for _, application := range applications {
		if application.Provisioning.isActive() {
			res = append(res, application)
		}
	}
	provisioningApplicationCache.Store(organization, &cachedProvisioningApplications{applications: res, fetchTime: time.Now()})
	return res, nil
}

func addProvisioningTask(application *Application, organization string, userId string, user string) error {
	now := util.GetCurrentTime()
	existing := &ProvisioningTask{}
	has, err := ormer.Engine.Where("application = ? and user_id = ? and state in (?, ?)", application.Name, userId, ProvisioningTaskPending, ProvisioningTaskRetrying).Get(existing)
	if err != nil {
		return err
	}
	if has {
		if existing.State == ProvisioningTaskPending {
			_, err = ormer.Engine.ID(existing.Id).Cols("next_time", "updated_time", "user").Update(&ProvisioningTask{NextTime: now, UpdatedTime: now, User: user})
		}
		return err
	}

	_, err = ormer.Engine.Insert(&ProvisioningTask{
		Application:  application.Name,
		Organization: organization,
		UserId:       userId,
		User:         user,
		CreatedTime:  now,
		UpdatedTime:  now,
		State:        ProvisioningTaskPending,
		NextTime:     now,
	})
	return err
}

// the columns that every sign-in writes, an update of only these changes nothing that is provisioned
var signinBookkeepingColumns = []string{"last_signin_time", "last_signin_ip", "last_signin_wrong_time", "signin_wrong_times", "is_online", "hash", "mfa_remember_deadline", "updated_time"}

func isProvisionedUserUpdate(columns []string) bool {
	if len(columns) == 0 {
		return true
	}
	for _, column := range columns {
		if !util.InSlice(signinBookkeepingColumns, column) {
			return true
		}
	}
	return false
}

// provisionUser queues the user for every application of its organization that provisions accounts
func provisionUser(user *User) {
	if user == nil || user.Id == "" {
		return
	}
	applications, err := getProvisioningApplications(user.Owner)
	if err != nil {
		logs.Warning("provisionUser() error: %s", err.Error())
		return
	}
	for _, application := range applications {
		if err = addProvisioningTask(application, user.Owner, user.Id, user.GetId()); err != nil {
			logs.Warning("provisionUser() error: %s", err.Error())
		}
	}
}

// requestOrganizationProvisioning reconciles the applications of an organization soon, for the changes of many users
// at once such as a syncer run
func requestOrganizationProvisioning(organization string) {
	applications, err := getProvisioningApplications(organization)
	if err != nil {
		logs.Warning("requestOrganizationProvisioning() error: %s", err.Error())
		return
	}
	for _, application := range applications {
		reconcileRequests.Store(application.Name, true)
	}
}

var reconcileRequests sync.Map

// ReconcileProvisioning queues every user of the application and every account it has provisioned, so that the
// accounts of the users who are in scope are created or updated and the others are disabled or deleted
func ReconcileProvisioning(application *Application) (int, error) {
	config := application.Provisioning
	if !config.isActive() {
		return 0, fmt.Errorf("the provisioning of the application: %s is not enabled", application.Name)
	}
	required, feature := config.getFeature()
	if err := edition.Require(required, feature); err != nil {
		return 0, err
	}

	queued := map[string]bool{}
	users := []*User{}
	err := ormer.Engine.Cols("owner", "name", "id").Where("owner = ?", application.Organization).Find(&users)
	if err != nil {
		return 0, err
	}
	for _, user := range users {
		if user.Id == "" || queued[user.Id] {
			continue
		}
		queued[user.Id] = true
		if err = addProvisioningTask(application, user.Owner, user.Id, user.GetId()); err != nil {
			return 0, err
		}
	}

	accounts := []*ProvisionedAccount{}
	err = ormer.Engine.Where("application = ? and state <> ?", application.Name, ProvisionedAccountDeleted).Find(&accounts)
	if err != nil {
		return 0, err
	}
	for _, account := range accounts {
		if queued[account.UserId] {
			continue
		}
		queued[account.UserId] = true
		if err = addProvisioningTask(application, account.Organization, account.UserId, account.User); err != nil {
			return 0, err
		}
	}

	application.Provisioning.LastSyncTime = util.GetCurrentTime()
	_, err = ormer.Engine.ID(core.PK{application.Owner, application.Name}).Cols("provisioning").Update(application)
	if err != nil {
		return 0, err
	}
	ClearProvisioningApplicationCache()
	return len(queued), nil
}

// isUserInProvisioningScope tells whether the user should have an account in the application's system
func isUserInProvisioningScope(user *User, application *Application) bool {
	if user == nil || user.IsDeleted || user.IsForbidden || user.Owner != application.Organization {
		return false
	}
	groups := application.Provisioning.Groups
	if len(groups) == 0 {
		return true
	}
	for _, group := range getUserGroupsWithAncestors(user) {
		if util.InSlice(groups, group) {
			return true
		}
	}
	return false
}

func getProvisionedAccount(application string, userId string) (*ProvisionedAccount, error) {
	account := &ProvisionedAccount{Application: application, UserId: userId}
	has, err := ormer.Engine.Get(account)
	if err != nil || !has {
		return nil, err
	}
	return account, nil
}

func saveProvisionedAccount(account *ProvisionedAccount, isNew bool) error {
	account.UpdatedTime = util.GetCurrentTime()
	if isNew {
		account.CreatedTime = account.UpdatedTime
		_, err := ormer.Engine.Insert(account)
		return err
	}
	_, err := ormer.Engine.ID(core.PK{account.Application, account.UserId}).AllCols().Update(account)
	return err
}

func getProvisioningHash(data interface{}) string {
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// provisioningTarget is the system behind an application that accounts are pushed to
type provisioningTarget interface {
	// payload is what the account looks like, its hash tells whether the account has to be written again
	payload(user *User) (interface{}, error)
	upsert(user *User, account *ProvisionedAccount) (string, error)
	disable(user *User, account *ProvisionedAccount) error
	remove(account *ProvisionedAccount) error
	test() error
}

func getProvisioningTarget(application *Application) (provisioningTarget, error) {
	config := application.Provisioning
	switch config.Type {
	case ProvisioningTypeScim:
		return newScimProvisioningTarget(application), nil
	case ProvisioningTypeDatabase:
		return newDatabaseProvisioningTarget(application)
	default:
		return nil, fmt.Errorf("unknown provisioning type: %s", config.Type)
	}
}

// TestProvisioning checks that the application's system can be reached with the provisioning settings
func TestProvisioning(application *Application) error {
	if application.Provisioning == nil {
		return fmt.Errorf("the provisioning of the application: %s is not set", application.Name)
	}
	required, feature := application.Provisioning.getFeature()
	if err := edition.Require(required, feature); err != nil {
		return err
	}
	target, err := getProvisioningTarget(application)
	if err != nil {
		return err
	}
	return target.test()
}

// processProvisioningTask brings one account up to date and returns what it did
func processProvisioningTask(task *ProvisioningTask) (string, error) {
	application, err := getApplication("admin", task.Application)
	if err != nil {
		return "", err
	}
	if application == nil || !application.Provisioning.isActive() {
		return "Skip", nil
	}
	required, feature := application.Provisioning.getFeature()
	if err = edition.Require(required, feature); err != nil {
		return "", err
	}

	user, err := GetUserByUserId(task.Organization, task.UserId)
	if err != nil {
		return "", err
	}
	account, err := getProvisionedAccount(application.Name, task.UserId)
	if err != nil {
		return "", err
	}

	target, err := getProvisioningTarget(application)
	if err != nil {
		return "", err
	}

	if isUserInProvisioningScope(user, application) {
		payload, err := target.payload(user)
		if err != nil {
			return "", err
		}
		hash := getProvisioningHash(payload)
		if account != nil && account.State == ProvisionedAccountActive && account.Hash == hash {
			return "Skip", nil
		}

		isNew := account == nil
		if isNew {
			account = &ProvisionedAccount{Application: application.Name, UserId: task.UserId, Organization: task.Organization}
		}
		account.User = user.GetId()
		remoteId, err := target.upsert(user, account)
		if err != nil {
			account.LastError = err.Error()
			_ = saveProvisionedAccount(account, isNew)
			return "", err
		}

		action := "Update"
		if isNew || account.State != ProvisionedAccountActive {
			action = "Create"
		}
		account.RemoteId, account.State, account.Hash, account.LastError = remoteId, ProvisionedAccountActive, hash, ""
		return action, saveProvisionedAccount(account, isNew)
	}

	if account == nil || account.State == ProvisionedAccountDeleted {
		return "Skip", nil
	}
	if application.Provisioning.Deprovision == ProvisioningDeprovisionDelete {
		if err = target.remove(account); err != nil {
			account.LastError = err.Error()
			_ = saveProvisionedAccount(account, false)
			return "", err
		}
		account.State, account.Hash, account.LastError = ProvisionedAccountDeleted, "", ""
		return "Delete", saveProvisionedAccount(account, false)
	}
	if account.State == ProvisionedAccountDisabled {
		return "Skip", nil
	}
	if err = target.disable(user, account); err != nil {
		account.LastError = err.Error()
		_ = saveProvisionedAccount(account, false)
		return "", err
	}
	account.State, account.Hash, account.LastError = ProvisionedAccountDisabled, "", ""
	return "Disable", saveProvisionedAccount(account, false)
}

func getProvisioningRetryTime(attempts int) string {
	delay := 30 * (1 << uint(min(attempts-1, 7)))
	if delay > 3600 {
		delay = 3600
	}
	return time.Now().Add(time.Duration(delay) * time.Second).Format(time.RFC3339)
}

func runProvisioningTask(task *ProvisioningTask) {
	action, err := processProvisioningTask(task)
	task.Attempts++
	task.UpdatedTime = util.GetCurrentTime()
	if err == nil {
		task.State, task.Action, task.LastError = ProvisioningTaskDone, action, ""
	} else {
		task.LastError = err.Error()
		if task.Attempts >= provisioningMaxAttempts {
			task.State = ProvisioningTaskFailed
		} else {
			task.State, task.NextTime = ProvisioningTaskRetrying, getProvisioningRetryTime(task.Attempts)
		}
	}
	_, err = ormer.Engine.ID(task.Id).Cols("state", "attempts", "next_time", "action", "last_error", "updated_time").Update(task)
	if err != nil {
		logs.Warning("runProvisioningTask() error: %s", err.Error())
	}
}

func runDueProvisioningTasks() {
	tasks := []*ProvisioningTask{}
	err := ormer.Engine.Where("state in (?, ?) and next_time <= ?", ProvisioningTaskPending, ProvisioningTaskRetrying, util.GetCurrentTime()).
		Asc("id").Limit(100).Find(&tasks)
	if err != nil {
		logs.Warning("runDueProvisioningTasks() error: %s", err.Error())
		return
	}
	for _, task := range tasks {
		runProvisioningTask(task)
	}
}

// runScheduledReconciliations reconciles the applications asked for by requestOrganizationProvisioning() and the
// ones whose sync interval has passed
func runScheduledReconciliations() {
	applications := []*Application{}
	err := ormer.Engine.Find(&applications)
	if err != nil {
		logs.Warning("runScheduledReconciliations() error: %s", err.Error())
		return
	}
	for _, application := range applications {
		config := application.Provisioning
		if !config.isActive() {
			continue
		}
		_, isRequested := reconcileRequests.LoadAndDelete(application.Name)
		isDue := false
		if config.SyncIntervalMinutes > 0 {
			lastSyncTime, err := time.Parse(time.RFC3339, config.LastSyncTime)
			isDue = err != nil || time.Since(lastSyncTime) >= time.Duration(config.SyncIntervalMinutes)*time.Minute
		}
		if !isRequested && !isDue {
			continue
		}
		required, feature := config.getFeature()
		if !edition.Allows(required, feature) {
			continue
		}
		if _, err = ReconcileProvisioning(application); err != nil {
			logs.Warning("ReconcileProvisioning() error: %s", err.Error())
		}
	}
}

var provisioningWorkerOnce sync.Once

// StartProvisioningWorker pushes the queued accounts every few seconds and runs the scheduled reconciliations
func StartProvisioningWorker() {
	provisioningWorkerOnce.Do(func() {
		util.SafeGoroutine(func() {
			tick := 0
			for {
				time.Sleep(5 * time.Second)
				if tick%12 == 0 {
					runScheduledReconciliations()
				}
				runDueProvisioningTasks()
				tick++
			}
		})
	})
}

func GetProvisionedAccounts(application string, state string, offset int, limit int) ([]*ProvisionedAccount, int64, error) {
	session := ormer.Engine.Where("application = ?", application)
	if state != "" {
		session = session.And("state = ?", state)
	}
	count, err := session.Count(&ProvisionedAccount{})
	if err != nil {
		return nil, 0, err
	}

	accounts := []*ProvisionedAccount{}
	session = ormer.Engine.Where("application = ?", application)
	if state != "" {
		session = session.And("state = ?", state)
	}
	err = session.Desc("updated_time").Limit(limit, offset).Find(&accounts)
	return accounts, count, err
}

func GetProvisioningTasks(application string, offset int, limit int) ([]*ProvisioningTask, error) {
	tasks := []*ProvisioningTask{}
	err := ormer.Engine.Where("application = ?", application).Desc("id").Limit(limit, offset).Find(&tasks)
	return tasks, err
}

type ProvisioningSummary struct {
	Accounts map[string]int64 `json:"accounts"`
	Tasks    map[string]int64 `json:"tasks"`
}

func GetProvisioningSummary(application string) (*ProvisioningSummary, error) {
	res := &ProvisioningSummary{Accounts: map[string]int64{}, Tasks: map[string]int64{}}
	for _, state := range []string{ProvisionedAccountActive, ProvisionedAccountDisabled, ProvisionedAccountDeleted} {
		count, err := ormer.Engine.Where("application = ? and state = ?", application, state).Count(&ProvisionedAccount{})
		if err != nil {
			return nil, err
		}
		res.Accounts[state] = count
	}
	count, err := ormer.Engine.Where("application = ? and last_error <> ?", application, "").Count(&ProvisionedAccount{})
	if err != nil {
		return nil, err
	}
	res.Accounts["Error"] = count
	for _, state := range []string{ProvisioningTaskPending, ProvisioningTaskRetrying, ProvisioningTaskFailed} {
		count, err = ormer.Engine.Where("application = ? and state = ?", application, state).Count(&ProvisioningTask{})
		if err != nil {
			return nil, err
		}
		res.Tasks[state] = count
	}
	return res, nil
}

// RetryFailedProvisioningTasks queues the failed tasks of the application again
func RetryFailedProvisioningTasks(application string) (int64, error) {
	now := util.GetCurrentTime()
	return ormer.Engine.Where("application = ? and state = ?", application, ProvisioningTaskFailed).
		Cols("state", "attempts", "next_time", "updated_time").
		Update(&ProvisioningTask{State: ProvisioningTaskPending, Attempts: 0, NextTime: now, UpdatedTime: now})
}
