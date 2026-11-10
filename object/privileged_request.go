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
	"errors"
	"fmt"
	"html"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	PamStatePending   = "Pending"
	PamStateApproved  = "Approved"
	PamStateRejected  = "Rejected"
	PamStateActive    = "Active"
	PamStateExpired   = "Expired"
	PamStateReturned  = "Returned"
	PamStateCancelled = "Cancelled"
)

// an approved request of an account has this long to be checked out, a pending one this long to be decided
const (
	pamCheckoutWindow = 24 * time.Hour
	pamPendingWindow  = 7 * 24 * time.Hour
)

// PrivilegedRequest is a user's request for a privileged account or a just-in-time role. An approved account request
// becomes active when the requester first reveals the password, a role request when it is approved, and either ends
// when it's returned or expires.
type PrivilegedRequest struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	Requester          string `xorm:"varchar(100) index" json:"requester"`
	Account            string `xorm:"varchar(100) index" json:"account"`
	AccountDisplayName string `xorm:"varchar(100)" json:"accountDisplayName"`
	Type               string `xorm:"varchar(100)" json:"type"`
	Role               string `xorm:"varchar(100)" json:"role"`
	// the user was put into the role by this request, and is taken out of it when the request ends
	RoleAdded       bool   `json:"roleAdded"`
	Reason          string `xorm:"varchar(1000)" json:"reason"`
	DurationMinutes int    `json:"durationMinutes"`

	State        string `xorm:"varchar(100) index" json:"state"`
	Approver     string `xorm:"varchar(100)" json:"approver"`
	Comment      string `xorm:"varchar(1000)" json:"comment"`
	DecisionTime string `xorm:"varchar(100)" json:"decisionTime"`
	// the deadline of the current state: to be decided, to be checked out, or of the access
	ExpireTime     string `xorm:"varchar(100)" json:"expireTime"`
	StartTime      string `xorm:"varchar(100)" json:"startTime"`
	EndTime        string `xorm:"varchar(100)" json:"endTime"`
	RevealCount    int    `json:"revealCount"`
	LastRevealTime string `xorm:"varchar(100)" json:"lastRevealTime"`
}

func (request *PrivilegedRequest) GetId() string {
	return fmt.Sprintf("%s/%s", request.Owner, request.Name)
}

func formatPamTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func parsePamTime(value string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, value)
	return t, err == nil
}

func GetPrivilegedRequestCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&PrivilegedRequest{})
}

func GetPaginationPrivilegedRequests(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*PrivilegedRequest, error) {
	requests := []*PrivilegedRequest{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&requests)
	return requests, err
}

func GetPrivilegedRequests(owner string) ([]*PrivilegedRequest, error) {
	requests := []*PrivilegedRequest{}
	err := ormer.Engine.Desc("created_time").Find(&requests, &PrivilegedRequest{Owner: owner})
	return requests, err
}

func getPrivilegedRequest(owner string, name string) (*PrivilegedRequest, error) {
	if owner == "" || name == "" {
		return nil, nil
	}
	request := PrivilegedRequest{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&request)
	if err != nil {
		return nil, err
	}
	if existed {
		return &request, nil
	}
	return nil, nil
}

func GetPrivilegedRequest(id string) (*PrivilegedRequest, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getPrivilegedRequest(owner, name)
}

// updateRequestState writes the request when it's still in the given state, so that two approvers or the job and the
// requester can't both move it
func updateRequestState(request *PrivilegedRequest, fromState string, columns ...string) (bool, error) {
	request.UpdatedTime = util.GetCurrentTime()
	columns = append(columns, "state", "updated_time")
	affected, err := ormer.Engine.ID(core.PK{request.Owner, request.Name}).Where("state = ?", fromState).Cols(columns...).Update(request)
	return affected != 0, err
}

func addPamRecord(owner string, userName string, action string, object interface{}, failure string) {
	record := &Record{
		Name:         util.GenerateId(),
		CreatedTime:  util.GetCurrentTime(),
		Organization: owner,
		User:         userName,
		Method:       "POST",
		RequestUri:   "/api/" + action,
		Action:       action,
		Object:       util.StructToJson(object),
		StatusCode:   200,
	}
	if failure != "" {
		record.StatusCode = 500
		record.Response = fmt.Sprintf("{status:\"error\", msg:\"%s\"}", failure)
		if len(failure) > 100 {
			failure = failure[:100]
		}
		record.Detail = failure
	} else {
		record.Response = "{status:\"ok\", msg:\"\"}"
	}
	util.SafeGoroutine(func() {
		AddRecord(record)
	})
}

// PamAccess is what the "Privileged access" page shows the current user
type PamAccess struct {
	Accounts  []*PamAccountView    `json:"accounts"`
	Requests  []*PrivilegedRequest `json:"requests"`
	Approvals []*PrivilegedRequest `json:"approvals"`
}

// GetPamAccess returns the accounts the user may request, the user's requests and the requests the user decides
func GetPamAccess(user *User) (*PamAccess, error) {
	accounts, err := GetPrivilegedAccounts(user.Owner)
	if err != nil {
		return nil, err
	}

	res := &PamAccess{Accounts: []*PamAccountView{}, Requests: []*PrivilegedRequest{}, Approvals: []*PrivilegedRequest{}}
	approverOf := map[string]bool{}
	for _, account := range accounts {
		if account.IsEnabled && account.isRequester(user.Name) {
			res.Accounts = append(res.Accounts, newPamAccountView(account))
		}
		if account.isApprover(user) {
			approverOf[account.Name] = true
		}
	}

	err = ormer.Engine.Where("owner = ? and requester = ?", user.Owner, user.Name).Desc("created_time").Limit(100).Find(&res.Requests)
	if err != nil {
		return nil, err
	}

	if len(approverOf) > 0 {
		names := []string{}
		for name := range approverOf {
			names = append(names, name)
		}
		approvals := []*PrivilegedRequest{}
		err = ormer.Engine.Where("owner = ? and requester <> ?", user.Owner, user.Name).In("account", names).
			And("(state = ? or approver = ?)", PamStatePending, user.Name).Desc("created_time").Limit(200).Find(&approvals)
		if err != nil {
			return nil, err
		}
		res.Approvals = approvals
	}
	return res, nil
}

func getRequestAccount(request *PrivilegedRequest, lang string) (*PrivilegedAccount, error) {
	account, err := getPrivilegedAccount(request.Owner, request.Account)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf(i18n.Translate(lang, "pam:The privileged account: %s does not exist"), request.Account)
	}
	return account, nil
}

// CreatePrivilegedRequest files a user's request, which is active right away when the account needs no approval
func CreatePrivilegedRequest(user *User, accountName string, reason string, minutes int, origin string, lang string) (*PrivilegedRequest, error) {
	account, err := getPrivilegedAccount(user.Owner, accountName)
	if err != nil {
		return nil, err
	}
	if account == nil || !account.IsEnabled || !account.isRequester(user.Name) {
		return nil, fmt.Errorf(i18n.Translate(lang, "pam:You cannot request the privileged account: %s"), accountName)
	}
	if reason == "" {
		return nil, errors.New(i18n.Translate(lang, "pam:Please give a reason for the request"))
	}
	if minutes <= 0 || minutes > account.getMaxCheckoutMinutes() {
		return nil, fmt.Errorf(i18n.Translate(lang, "pam:The duration must be between 1 and %d minutes"), account.getMaxCheckoutMinutes())
	}
	count, err := ormer.Engine.Where("owner = ? and requester = ? and account = ?", user.Owner, user.Name, accountName).
		In("state", PamStatePending, PamStateApproved, PamStateActive).Count(&PrivilegedRequest{})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New(i18n.Translate(lang, "pam:You already have an open request for this account"))
	}

	now := time.Now()
	request := &PrivilegedRequest{
		Owner:              user.Owner,
		Name:               util.GenerateId(),
		CreatedTime:        util.GetCurrentTime(),
		UpdatedTime:        util.GetCurrentTime(),
		Requester:          user.Name,
		Account:            account.Name,
		AccountDisplayName: account.DisplayName,
		Type:               account.Type,
		Role:               account.Role,
		Reason:             reason,
		DurationMinutes:    minutes,
		State:              PamStatePending,
		ExpireTime:         formatPamTime(now.Add(pamPendingWindow)),
	}

	if !account.RequireApproval {
		if err = activatePrivilegedRequest(request, account, lang); err != nil {
			return nil, err
		}
	}

	_, err = ormer.Engine.Insert(request)
	if err != nil {
		if request.RoleAdded {
			removeUserFromPamRole(request)
		}
		return nil, err
	}

	if request.State == PamStatePending {
		util.SafeGoroutine(func() {
			notifyPamApprovers(account, request, origin, lang)
		})
	}
	return request, nil
}

// activatePrivilegedRequest starts the access window, putting the user into the role of a role request
func activatePrivilegedRequest(request *PrivilegedRequest, account *PrivilegedAccount, lang string) error {
	if account.isRole() {
		if err := addUserToPamRole(request, lang); err != nil {
			return err
		}
	}
	now := time.Now()
	request.State = PamStateActive
	request.StartTime = formatPamTime(now)
	request.ExpireTime = formatPamTime(now.Add(time.Duration(request.DurationMinutes) * time.Minute))
	return nil
}

func addUserToPamRole(request *PrivilegedRequest, lang string) error {
	role, err := getRole(request.Owner, request.Role)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf(i18n.Translate(lang, "pam:The role: %s does not exist in the organization"), request.Role)
	}
	userId := util.GetId(request.Owner, request.Requester)
	if util.InSlice(role.Users, userId) {
		request.RoleAdded = false
		return nil
	}

	role.Users = append(role.Users, userId)
	if err = CheckRoleSod(role, lang); err != nil {
		return err
	}
	_, err = ormer.Engine.ID(core.PK{role.Owner, role.Name}).Cols("users").Update(role)
	if err != nil {
		return err
	}
	request.RoleAdded = true
	addPamRecord(request.Owner, request.Requester, "add-privileged-role", map[string]string{"request": request.GetId(), "role": role.GetId(), "user": userId, "expireTime": formatPamTime(time.Now().Add(time.Duration(request.DurationMinutes) * time.Minute))}, "")
	return nil
}

func removeUserFromPamRole(request *PrivilegedRequest) {
	role, err := getRole(request.Owner, request.Role)
	failure := ""
	if err != nil {
		failure = err.Error()
	} else if role != nil {
		userId := util.GetId(request.Owner, request.Requester)
		users := []string{}
		for _, user := range role.Users {
			if user != userId {
				users = append(users, user)
			}
		}
		role.Users = users
		_, err = ormer.Engine.ID(core.PK{role.Owner, role.Name}).Cols("users").Update(role)
		if err != nil {
			failure = err.Error()
		}
	}
	if failure != "" {
		logs.Error("failed to take the user: %s out of the role: %s of the privileged request: %s: %s", request.Requester, request.Role, request.GetId(), failure)
	}
	addPamRecord(request.Owner, request.Requester, "remove-privileged-role", map[string]string{"request": request.GetId(), "role": util.GetId(request.Owner, request.Role), "user": util.GetId(request.Owner, request.Requester)}, failure)
}

// DecidePrivilegedRequest approves or rejects a pending request, by one of its approvers other than the requester
func DecidePrivilegedRequest(approver *User, id string, approve bool, comment string, origin string, lang string) (*PrivilegedRequest, error) {
	request, err := GetPrivilegedRequest(id)
	if err != nil {
		return nil, err
	}
	if request == nil || request.Owner != approver.Owner {
		return nil, errors.New(i18n.Translate(lang, "pam:The request does not exist"))
	}
	if request.Requester == approver.Name {
		return nil, errors.New(i18n.Translate(lang, "pam:You cannot approve your own request"))
	}
	account, err := getRequestAccount(request, lang)
	if err != nil {
		return nil, err
	}
	if !account.isApprover(approver) {
		return nil, errors.New(i18n.Translate(lang, "pam:You are not an approver of this account"))
	}
	if request.State != PamStatePending {
		return nil, fmt.Errorf(i18n.Translate(lang, "pam:The request is %s, not pending"), request.State)
	}

	now := time.Now()
	request.Approver = approver.Name
	request.Comment = comment
	request.DecisionTime = formatPamTime(now)
	columns := []string{"approver", "comment", "decision_time", "expire_time"}
	if !approve {
		request.State = PamStateRejected
		request.EndTime = formatPamTime(now)
		columns = append(columns, "end_time")
	} else if account.isRole() {
		if err = activatePrivilegedRequest(request, account, lang); err != nil {
			return nil, err
		}
		columns = append(columns, "start_time", "role_added")
	} else {
		request.State = PamStateApproved
		request.ExpireTime = formatPamTime(now.Add(pamCheckoutWindow))
	}

	ok, err := updateRequestState(request, PamStatePending, columns...)
	if err != nil || !ok {
		if request.RoleAdded {
			removeUserFromPamRole(request)
		}
		if err == nil {
			err = errors.New(i18n.Translate(lang, "pam:The request has just been changed, please refresh"))
		}
		return nil, err
	}

	util.SafeGoroutine(func() {
		notifyPamRequester(request, origin, lang)
	})
	return request, nil
}

func getOwnRequest(user *User, id string, lang string) (*PrivilegedRequest, error) {
	request, err := GetPrivilegedRequest(id)
	if err != nil {
		return nil, err
	}
	if request == nil || request.Owner != user.Owner || request.Requester != user.Name {
		return nil, errors.New(i18n.Translate(lang, "pam:The request does not exist"))
	}
	return request, nil
}

// RevealPrivilegedPassword hands the account's password to its requester while the request is approved or active,
// an approved request becomes active with the first reveal
func RevealPrivilegedPassword(user *User, id string, lang string) (string, *PrivilegedRequest, error) {
	request, err := getOwnRequest(user, id, lang)
	if err != nil {
		return "", nil, err
	}
	account, err := getRequestAccount(request, lang)
	if err != nil {
		return "", nil, err
	}
	if account.isRole() {
		return "", nil, errors.New(i18n.Translate(lang, "pam:A role has no password"))
	}
	if !account.IsEnabled {
		return "", nil, fmt.Errorf(i18n.Translate(lang, "pam:You cannot request the privileged account: %s"), account.Name)
	}

	now := time.Now()
	expireTime, _ := parsePamTime(request.ExpireTime)
	switch request.State {
	case PamStateApproved:
		if now.After(expireTime) {
			return "", nil, errors.New(i18n.Translate(lang, "pam:The access has expired"))
		}
		if err = activatePrivilegedRequest(request, account, lang); err != nil {
			return "", nil, err
		}
		ok, err := updateRequestState(request, PamStateApproved, "start_time", "expire_time")
		if err != nil {
			return "", nil, err
		}
		if !ok {
			return "", nil, errors.New(i18n.Translate(lang, "pam:The request has just been changed, please refresh"))
		}
	case PamStateActive:
		if now.After(expireTime) {
			return "", nil, errors.New(i18n.Translate(lang, "pam:The access has expired"))
		}
	default:
		return "", nil, fmt.Errorf(i18n.Translate(lang, "pam:The request is %s, not active"), request.State)
	}

	secret, err := account.getSecret()
	if err != nil {
		return "", nil, err
	}
	if secret == "" {
		return "", nil, errors.New(i18n.Translate(lang, "pam:The account has no password in the vault"))
	}

	request.RevealCount++
	request.LastRevealTime = formatPamTime(now)
	_, err = ormer.Engine.ID(core.PK{request.Owner, request.Name}).Cols("reveal_count", "last_reveal_time").Update(request)
	if err != nil {
		return "", nil, err
	}
	return secret, request, nil
}

// ReturnPrivilegedRequest ends the requester's access before it expires
func ReturnPrivilegedRequest(user *User, id string, lang string) (*PrivilegedRequest, error) {
	request, err := getOwnRequest(user, id, lang)
	if err != nil {
		return nil, err
	}
	if request.State != PamStateActive && request.State != PamStateApproved {
		return nil, fmt.Errorf(i18n.Translate(lang, "pam:The request is %s, not active"), request.State)
	}
	ok, err := endPrivilegedRequest(request, PamStateReturned)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(i18n.Translate(lang, "pam:The request has just been changed, please refresh"))
	}
	return request, nil
}

// CancelPrivilegedRequest withdraws a request that is still pending
func CancelPrivilegedRequest(user *User, id string, lang string) (*PrivilegedRequest, error) {
	request, err := getOwnRequest(user, id, lang)
	if err != nil {
		return nil, err
	}
	if request.State != PamStatePending {
		return nil, fmt.Errorf(i18n.Translate(lang, "pam:The request is %s, not pending"), request.State)
	}
	request.State = PamStateCancelled
	request.EndTime = formatPamTime(time.Now())
	ok, err := updateRequestState(request, PamStatePending, "end_time")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(i18n.Translate(lang, "pam:The request has just been changed, please refresh"))
	}
	return request, nil
}

// endPrivilegedRequest returns or expires an approved or active request: the user leaves the role, and the account's
// password is due for rotation if it was revealed and the account rotates after check-in
func endPrivilegedRequest(request *PrivilegedRequest, state string) (bool, error) {
	fromState := request.State
	request.State = state
	request.EndTime = formatPamTime(time.Now())
	ok, err := updateRequestState(request, fromState, "end_time")
	if err != nil || !ok {
		return ok, err
	}

	if request.RoleAdded {
		removeUserFromPamRole(request)
	}
	if request.RevealCount > 0 {
		account, err := getPrivilegedAccount(request.Owner, request.Account)
		if err != nil {
			return true, err
		}
		if account != nil && account.RotateAfterCheckin && account.canRotate() {
			account.NeedsRotation = true
			_, err = ormer.Engine.ID(core.PK{account.Owner, account.Name}).Cols("needs_rotation").Update(account)
			if err != nil {
				return true, err
			}
		}
	}
	return true, nil
}

// ExpirePrivilegedRequests ends the requests whose deadline has passed
func ExpirePrivilegedRequests() error {
	requests := []*PrivilegedRequest{}
	err := ormer.Engine.In("state", PamStatePending, PamStateApproved, PamStateActive).Find(&requests)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, request := range requests {
		expireTime, ok := parsePamTime(request.ExpireTime)
		if !ok || now.Before(expireTime) {
			continue
		}
		if request.State == PamStatePending {
			request.State = PamStateExpired
			request.EndTime = formatPamTime(now)
			ok, err = updateRequestState(request, PamStatePending, "end_time")
		} else {
			ok, err = endPrivilegedRequest(request, PamStateExpired)
		}
		if err != nil {
			logs.Error("failed to expire the privileged request: %s: %s", request.GetId(), err.Error())
			continue
		}
		if ok {
			addPamRecord(request.Owner, request.Requester, "expire-privileged-request", map[string]string{"request": request.GetId(), "account": request.Account, "role": request.Role}, "")
		}
	}
	return nil
}

func getPamUserEmails(owner string, names []string) []string {
	emails := []string{}
	for _, name := range names {
		user, err := getUser(owner, name)
		if err == nil && user != nil && user.Email != "" && !user.IsForbidden {
			emails = append(emails, user.Email)
		}
	}
	return emails
}

func getPamApproverNames(account *PrivilegedAccount) []string {
	if len(account.Approvers) > 0 {
		return account.Approvers
	}
	names := []string{}
	admins := []*User{}
	err := ormer.Engine.Cols("name").Where("owner = ? and is_admin = ?", account.Owner, true).Find(&admins)
	if err == nil {
		for _, admin := range admins {
			names = append(names, admin.Name)
		}
	}
	return names
}

func sendPamEmail(owner string, names []string, title string, content string) {
	organization, err := getOrganization("admin", owner)
	if err != nil || organization == nil {
		return
	}
	dest := getPamUserEmails(owner, names)
	if len(dest) == 0 {
		return
	}
	err = sendOrganizationEmail(organization, title, content, dest)
	if err != nil {
		logs.Warn("the privileged access email of the organization: %s is not sent: %s", owner, err.Error())
	}
}

func notifyPamApprovers(account *PrivilegedAccount, request *PrivilegedRequest, origin string, lang string) {
	target := account.DisplayName
	if target == "" {
		target = account.Name
	}
	title := fmt.Sprintf(i18n.Translate(lang, "pam:%s requests access to %s"), request.Requester, target)
	content := fmt.Sprintf("<p>%s</p><p>%s: %s</p><p>%s: %d</p><p><a href=\"%s/privileged-access\">%s</a></p>",
		html.EscapeString(title), html.EscapeString(i18n.Translate(lang, "pam:Reason")), html.EscapeString(request.Reason),
		html.EscapeString(i18n.Translate(lang, "pam:Duration in minutes")), request.DurationMinutes, origin,
		html.EscapeString(i18n.Translate(lang, "pam:Open the privileged access page to approve or reject it")))
	sendPamEmail(request.Owner, getPamApproverNames(account), title, content)
}

func notifyPamRequester(request *PrivilegedRequest, origin string, lang string) {
	target := request.AccountDisplayName
	if target == "" {
		target = request.Account
	}
	title := fmt.Sprintf(i18n.Translate(lang, "pam:Your request for %s is %s"), target, request.State)
	content := fmt.Sprintf("<p>%s</p><p>%s</p><p><a href=\"%s/privileged-access\">%s</a></p>", html.EscapeString(title),
		html.EscapeString(request.Comment), origin, html.EscapeString(i18n.Translate(lang, "pam:Open the privileged access page")))
	sendPamEmail(request.Owner, []string{request.Requester}, title, content)
}

// GetPamOrigin returns the console address for the links in the emails
func GetPamOrigin(host string) string {
	origin, _ := getOriginFromHost(host)
	return origin
}
