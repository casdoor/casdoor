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
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/robfig/cron/v3"
	"github.com/xorm-io/core"
)

// AccessReview is a campaign in which reviewers confirm or revoke the roles, groups and permissions the users of an
// organization hold, 定期权限复核. Each start of the campaign is a round with its own review items
type AccessReview struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"mediumtext" json:"description"`

	EntitlementTypes []string `xorm:"mediumtext" json:"entitlementTypes"`
	Roles            []string `xorm:"mediumtext" json:"roles"`
	Groups           []string `xorm:"mediumtext" json:"groups"`
	Permissions      []string `xorm:"mediumtext" json:"permissions"`

	ReviewerType    string   `xorm:"varchar(100)" json:"reviewerType"`
	Reviewers       []string `xorm:"mediumtext" json:"reviewers"`
	DurationDays    int      `json:"durationDays"`
	DefaultDecision string   `xorm:"varchar(20)" json:"defaultDecision"`
	AutoApply       bool     `json:"autoApply"`
	RecurrenceDays  int      `json:"recurrenceDays"`

	State         string `xorm:"varchar(20)" json:"state"`
	Round         int    `json:"round"`
	StartTime     string `xorm:"varchar(100)" json:"startTime"`
	EndTime       string `xorm:"varchar(100)" json:"endTime"`
	CompletedTime string `xorm:"varchar(100)" json:"completedTime"`
	NextStartTime string `xorm:"varchar(100)" json:"nextStartTime"`

	Progress *AccessReviewProgress `xorm:"-" json:"progress,omitempty"`
}

// AccessReviewProgress counts the items of the current round
type AccessReviewProgress struct {
	Total   int `json:"total"`
	Pending int `json:"pending"`
	Keep    int `json:"keep"`
	Revoke  int `json:"revoke"`
	Applied int `json:"applied"`
}

// AccessReviewItem is one entitlement of one user that a reviewer keeps or revokes in a round of a campaign
type AccessReviewItem struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Review string `xorm:"varchar(100) index" json:"review"`
	Round  int    `xorm:"index" json:"round"`

	User                   string `xorm:"varchar(100) index" json:"user"`
	UserDisplayName        string `xorm:"varchar(100)" json:"userDisplayName"`
	EntitlementType        string `xorm:"varchar(20)" json:"entitlementType"`
	Entitlement            string `xorm:"varchar(255)" json:"entitlement"`
	EntitlementDisplayName string `xorm:"varchar(255)" json:"entitlementDisplayName"`

	Reviewer    string `xorm:"varchar(100) index" json:"reviewer"`
	Decision    string `xorm:"varchar(20) index" json:"decision"`
	Comment     string `xorm:"mediumtext" json:"comment"`
	DecidedBy   string `xorm:"varchar(100)" json:"decidedBy"`
	DecidedTime string `xorm:"varchar(100)" json:"decidedTime"`

	Applied     bool   `json:"applied"`
	AppliedTime string `xorm:"varchar(100)" json:"appliedTime"`
	ApplyResult string `xorm:"mediumtext" json:"applyResult"`

	ReviewDisplayName string `xorm:"-" json:"reviewDisplayName,omitempty"`
	EndTime           string `xorm:"-" json:"endTime,omitempty"`
}

const (
	AccessReviewStateDraft     = "Draft"
	AccessReviewStateActive    = "Active"
	AccessReviewStateCompleted = "Completed"

	AccessReviewDecisionPending = "Pending"
	AccessReviewDecisionKeep    = "Keep"
	AccessReviewDecisionRevoke  = "Revoke"

	EntitlementTypeRole        = "Role"
	EntitlementTypeGroup       = "Group"
	EntitlementTypePermission  = "Permission"
	EntitlementTypeApplication = "Application"

	ReviewerTypeManager   = "Manager"
	ReviewerTypeReviewers = "Reviewers"

	// AccessReviewDeadlineDecider is who decided the items left undecided at the deadline
	AccessReviewDeadlineDecider = "(deadline)"
)

var entitlementTypes = []string{EntitlementTypeRole, EntitlementTypeGroup, EntitlementTypePermission, EntitlementTypeApplication}

func (review *AccessReview) GetId() string {
	return fmt.Sprintf("%s/%s", review.Owner, review.Name)
}

func GetAccessReviewCount(owner string, field string, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&AccessReview{})
}

func GetPaginationAccessReviews(owner string, offset int, limit int, field string, value string, sortField string, sortOrder string) ([]*AccessReview, error) {
	reviews := []*AccessReview{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&reviews)
	if err != nil {
		return nil, err
	}
	for _, review := range reviews {
		review.Progress, err = getAccessReviewProgress(review)
		if err != nil {
			return nil, err
		}
	}
	return reviews, nil
}

func getAccessReview(owner string, name string) (*AccessReview, error) {
	if owner == "" || name == "" {
		return nil, nil
	}
	review := AccessReview{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&review)
	if err != nil || !existed {
		return nil, err
	}
	return &review, nil
}

// GetAccessReview returns the campaign with the progress of its current round
func GetAccessReview(id string) (*AccessReview, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	review, err := getAccessReview(owner, name)
	if err != nil || review == nil {
		return review, err
	}
	review.Progress, err = getAccessReviewProgress(review)
	return review, err
}

func getAccessReviewProgress(review *AccessReview) (*AccessReviewProgress, error) {
	progress := &AccessReviewProgress{}
	if review.Round == 0 {
		return progress, nil
	}
	items := []*AccessReviewItem{}
	err := ormer.Engine.Cols("decision", "applied").Where("owner = ? and review = ? and round = ?", review.Owner, review.Name, review.Round).Find(&items)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		progress.Total++
		switch item.Decision {
		case AccessReviewDecisionKeep:
			progress.Keep++
		case AccessReviewDecisionRevoke:
			progress.Revoke++
		default:
			progress.Pending++
		}
		if item.Applied {
			progress.Applied++
		}
	}
	return progress, nil
}

// ValidateAccessReview checks the settings of a campaign an admin saves
func ValidateAccessReview(review *AccessReview, lang string) error {
	if review.Owner == "" || review.Name == "" || strings.Contains(review.Name, "/") {
		return errors.New(i18n.Translate(lang, "general:Missing parameter"))
	}
	if len(review.EntitlementTypes) == 0 {
		return errors.New(i18n.Translate(lang, "enterprise:Choose what the access review covers"))
	}
	for _, entitlementType := range review.EntitlementTypes {
		if !util.InSlice(entitlementTypes, entitlementType) {
			return fmt.Errorf("unknown entitlement type: %s", entitlementType)
		}
	}
	if review.ReviewerType != ReviewerTypeManager && review.ReviewerType != ReviewerTypeReviewers {
		return fmt.Errorf("unknown reviewer type: %s", review.ReviewerType)
	}
	if review.ReviewerType == ReviewerTypeReviewers && len(review.Reviewers) == 0 {
		return errors.New(i18n.Translate(lang, "enterprise:Choose at least one reviewer"))
	}
	for _, reviewer := range review.Reviewers {
		user, err := getUser(review.Owner, reviewer)
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf(i18n.Translate(lang, "general:The user: %s doesn't exist"), util.GetId(review.Owner, reviewer))
		}
	}
	if review.DurationDays <= 0 || review.DurationDays > 365 {
		return errors.New(i18n.Translate(lang, "enterprise:The review period must be between 1 and 365 days"))
	}
	if review.RecurrenceDays < 0 || (review.RecurrenceDays > 0 && review.RecurrenceDays < review.DurationDays) {
		return errors.New(i18n.Translate(lang, "enterprise:The recurrence must be longer than the review period"))
	}
	if review.DefaultDecision != AccessReviewDecisionKeep && review.DefaultDecision != AccessReviewDecisionRevoke {
		return fmt.Errorf("unknown default decision: %s", review.DefaultDecision)
	}
	return nil
}

func AddAccessReview(review *AccessReview, lang string) (bool, error) {
	if err := ValidateAccessReview(review, lang); err != nil {
		return false, err
	}
	review.CreatedTime = util.GetCurrentTime()
	review.UpdatedTime = review.CreatedTime
	review.State = AccessReviewStateDraft
	review.Round = 0
	review.StartTime, review.EndTime, review.CompletedTime, review.NextStartTime = "", "", "", ""
	affected, err := ormer.Engine.Insert(review)
	return affected != 0, err
}

// UpdateAccessReview changes the settings of a campaign that isn't running, its state and rounds stay as they are
func UpdateAccessReview(id string, review *AccessReview, lang string) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	oldReview, err := getAccessReview(owner, name)
	if err != nil || oldReview == nil {
		return false, err
	}
	if oldReview.State == AccessReviewStateActive {
		return false, errors.New(i18n.Translate(lang, "enterprise:Close the access review before changing it"))
	}
	if review.Owner != oldReview.Owner || (review.Name != oldReview.Name && oldReview.Round > 0) {
		return false, errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
	}
	if err = ValidateAccessReview(review, lang); err != nil {
		return false, err
	}

	review.CreatedTime = oldReview.CreatedTime
	review.UpdatedTime = util.GetCurrentTime()
	review.State, review.Round = oldReview.State, oldReview.Round
	review.StartTime, review.EndTime, review.CompletedTime = oldReview.StartTime, oldReview.EndTime, oldReview.CompletedTime
	review.NextStartTime = ""
	if review.State == AccessReviewStateCompleted && review.RecurrenceDays > 0 {
		review.NextStartTime = addDays(review.StartTime, review.RecurrenceDays)
	}
	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(review)
	return affected != 0, err
}

// DeleteAccessReview deletes a campaign that isn't running with all its items
func DeleteAccessReview(review *AccessReview, lang string) (bool, error) {
	oldReview, err := getAccessReview(review.Owner, review.Name)
	if err != nil || oldReview == nil {
		return false, err
	}
	if oldReview.State == AccessReviewStateActive {
		return false, errors.New(i18n.Translate(lang, "enterprise:Close the access review before changing it"))
	}
	_, err = ormer.Engine.Where("owner = ? and review = ?", oldReview.Owner, oldReview.Name).Delete(&AccessReviewItem{})
	if err != nil {
		return false, err
	}
	affected, err := ormer.Engine.ID(core.PK{oldReview.Owner, oldReview.Name}).Delete(&AccessReview{})
	return affected != 0, err
}

func addDays(start string, days int) string {
	t, ok := parseAnyTime(start)
	if !ok {
		t = time.Now()
	}
	return t.AddDate(0, 0, days).Format(time.RFC3339)
}

type reviewerResolver struct {
	review *AccessReview
	groups map[string]*Group
	next   int
}

func normalizeUserName(owner string, name string) string {
	if strings.Contains(name, "/") {
		userOwner, userName := util.GetOwnerAndNameFromIdNoCheck(name)
		if userOwner != owner {
			return ""
		}
		return userName
	}
	return name
}

func (r *reviewerResolver) getGroupManager(groupName string, user string) string {
	for i := 0; i < 100 && groupName != ""; i++ {
		group, ok := r.groups[groupName]
		if !ok {
			return ""
		}
		if manager := normalizeUserName(r.review.Owner, group.Manager); manager != "" && manager != user {
			return manager
		}
		if group.IsTopGroup || group.ParentId == group.Name {
			return ""
		}
		groupName = group.ParentId
	}
	return ""
}

func (r *reviewerResolver) nextReviewer(user string) string {
	for i := 0; i < len(r.review.Reviewers); i++ {
		reviewer := r.review.Reviewers[r.next%len(r.review.Reviewers)]
		r.next++
		if reviewer != user {
			return reviewer
		}
	}
	return ""
}

// resolve picks the reviewer of an item: the manager of the group, or of the user's groups, for the manager type,
// otherwise or when there is none the listed reviewers in turn. Nobody reviews their own access, an item without a
// reviewer is left to the organization's admins
func (r *reviewerResolver) resolve(user *User, userName string, entitlementType string, groupName string) string {
	if r.review.ReviewerType == ReviewerTypeManager {
		if entitlementType == EntitlementTypeGroup {
			if manager := r.getGroupManager(groupName, userName); manager != "" {
				return manager
			}
		} else if user != nil {
			for _, group := range user.Groups {
				groupOwner, name := util.GetOwnerAndNameFromIdNoCheck(group)
				if groupOwner != r.review.Owner {
					continue
				}
				if manager := r.getGroupManager(name, userName); manager != "" {
					return manager
				}
			}
		}
	}
	return r.nextReviewer(userName)
}

func isWildcardEntry(id string) bool {
	return strings.Contains(id, "*")
}

// buildAccessReviewItems lists the entitlements a round of the campaign reviews
func buildAccessReviewItems(review *AccessReview, round int) ([]*AccessReviewItem, error) {
	users, err := GetUsers(review.Owner)
	if err != nil {
		return nil, err
	}
	userMap := map[string]*User{}
	for _, user := range users {
		userMap[user.Name] = user
	}
	groups, err := GetGroups(review.Owner)
	if err != nil {
		return nil, err
	}
	groupMap := map[string]*Group{}
	for _, group := range groups {
		groupMap[group.Name] = group
	}
	resolver := &reviewerResolver{review: review, groups: groupMap}

	items := []*AccessReviewItem{}
	seen := map[string]bool{}
	now := util.GetCurrentTime()
	add := func(userName string, entitlementType string, entitlement string, displayName string, groupName string) {
		key := userName + "|" + entitlementType + "|" + entitlement
		if seen[key] {
			return
		}
		seen[key] = true
		user := userMap[userName]
		userDisplayName := ""
		if user != nil {
			userDisplayName = user.DisplayName
		}
		items = append(items, &AccessReviewItem{
			Owner:                  review.Owner,
			Name:                   util.GenerateId(),
			CreatedTime:            now,
			Review:                 review.Name,
			Round:                  round,
			User:                   userName,
			UserDisplayName:        userDisplayName,
			EntitlementType:        entitlementType,
			Entitlement:            entitlement,
			EntitlementDisplayName: displayName,
			Reviewer:               resolver.resolve(user, userName, entitlementType, groupName),
			Decision:               AccessReviewDecisionPending,
		})
	}
	addUserIds := func(userIds []string, entitlementType string, entitlement string, displayName string) {
		for _, userId := range userIds {
			owner, name := util.GetOwnerAndNameFromIdNoCheck(userId)
			if owner != review.Owner || name == "" || isWildcardEntry(userId) {
				continue
			}
			add(name, entitlementType, entitlement, displayName, "")
		}
	}

	if util.InSlice(review.EntitlementTypes, EntitlementTypeRole) {
		roles, err := GetRoles(review.Owner)
		if err != nil {
			return nil, err
		}
		for _, role := range roles {
			if len(review.Roles) == 0 || util.InSlice(review.Roles, role.Name) {
				addUserIds(role.Users, EntitlementTypeRole, role.GetId(), role.DisplayName)
			}
		}
	}

	if util.InSlice(review.EntitlementTypes, EntitlementTypeGroup) {
		for _, user := range users {
			for _, groupId := range user.Groups {
				groupOwner, groupName := util.GetOwnerAndNameFromIdNoCheck(groupId)
				if !strings.Contains(groupId, "/") {
					groupOwner, groupName = review.Owner, groupId
				}
				if groupOwner != review.Owner || (len(review.Groups) > 0 && !util.InSlice(review.Groups, groupName)) {
					continue
				}
				displayName := ""
				if group, ok := groupMap[groupName]; ok {
					displayName = group.DisplayName
				}
				add(user.Name, EntitlementTypeGroup, groupId, displayName, groupName)
			}
		}
	}

	if util.InSlice(review.EntitlementTypes, EntitlementTypePermission) || util.InSlice(review.EntitlementTypes, EntitlementTypeApplication) {
		permissions, err := GetPermissions(review.Owner)
		if err != nil {
			return nil, err
		}
		for _, permission := range permissions {
			entitlementType := EntitlementTypePermission
			if permission.ResourceType == "Application" {
				entitlementType = EntitlementTypeApplication
			}
			if !util.InSlice(review.EntitlementTypes, entitlementType) || (len(review.Permissions) > 0 && !util.InSlice(review.Permissions, permission.Name)) {
				continue
			}
			addUserIds(permission.Users, entitlementType, permission.GetId(), permission.DisplayName)
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].User != items[j].User {
			return items[i].User < items[j].User
		}
		return items[i].Entitlement < items[j].Entitlement
	})
	return items, nil
}

// StartAccessReview starts the next round of a campaign that isn't running and lists the items to review
func StartAccessReview(review *AccessReview, lang string) (int, error) {
	if review.State == AccessReviewStateActive {
		return 0, errors.New(i18n.Translate(lang, "enterprise:The access review is already running"))
	}
	if err := ValidateAccessReview(review, lang); err != nil {
		return 0, err
	}
	round := review.Round + 1
	items, err := buildAccessReviewItems(review, round)
	if err != nil {
		return 0, err
	}

	session := ormer.Engine.NewSession()
	defer session.Close()
	if err = session.Begin(); err != nil {
		return 0, err
	}
	for start := 0; start < len(items); start += 100 {
		end := start + 100
		if end > len(items) {
			end = len(items)
		}
		if _, err = session.Insert(items[start:end]); err != nil {
			_ = session.Rollback()
			return 0, err
		}
	}

	now := time.Now()
	review.State = AccessReviewStateActive
	review.Round = round
	review.StartTime = now.Format(time.RFC3339)
	review.EndTime = now.AddDate(0, 0, review.DurationDays).Format(time.RFC3339)
	review.CompletedTime = ""
	review.NextStartTime = ""
	review.UpdatedTime = util.GetCurrentTime()
	_, err = session.ID(core.PK{review.Owner, review.Name}).Cols("state", "round", "start_time", "end_time", "completed_time", "next_start_time", "updated_time").Update(review)
	if err != nil {
		_ = session.Rollback()
		return 0, err
	}
	return len(items), session.Commit()
}

// CompleteAccessReview closes the running round: the undecided items get the campaign's default decision, the
// revocations are applied when the campaign applies them on completion
func CompleteAccessReview(review *AccessReview, actor string, lang string) (*AccessReviewApplyResult, error) {
	if review.State != AccessReviewStateActive {
		return nil, errors.New(i18n.Translate(lang, "enterprise:The access review is not running"))
	}
	now := util.GetCurrentTime()
	_, err := ormer.Engine.Where("owner = ? and review = ? and round = ? and decision = ?", review.Owner, review.Name, review.Round, AccessReviewDecisionPending).
		Cols("decision", "decided_by", "decided_time").
		Update(&AccessReviewItem{Decision: review.DefaultDecision, DecidedBy: AccessReviewDeadlineDecider, DecidedTime: now})
	if err != nil {
		return nil, err
	}

	review.State = AccessReviewStateCompleted
	review.CompletedTime = now
	review.UpdatedTime = now
	review.NextStartTime = ""
	if review.RecurrenceDays > 0 {
		review.NextStartTime = addDays(review.StartTime, review.RecurrenceDays)
	}
	_, err = ormer.Engine.ID(core.PK{review.Owner, review.Name}).Cols("state", "completed_time", "next_start_time", "updated_time").Update(review)
	if err != nil {
		return nil, err
	}

	if !review.AutoApply {
		return &AccessReviewApplyResult{}, nil
	}
	return ApplyAccessReview(review, actor)
}

// AccessReviewApplyResult counts the revocations applied
type AccessReviewApplyResult struct {
	Applied int      `json:"applied"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

// ApplyAccessReview removes the users from the roles, groups and permissions revoked in the current round and
// writes each change to the audit records
func ApplyAccessReview(review *AccessReview, actor string) (*AccessReviewApplyResult, error) {
	result := &AccessReviewApplyResult{Errors: []string{}}
	items := []*AccessReviewItem{}
	err := ormer.Engine.Where("owner = ? and review = ? and round = ? and decision = ? and applied = ?", review.Owner, review.Name, review.Round, AccessReviewDecisionRevoke, false).Find(&items)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		message, err := revokeAccessReviewItem(item)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s %s %s: %s", item.User, item.EntitlementType, item.Entitlement, err.Error()))
			item.ApplyResult = err.Error()
			_, _ = ormer.Engine.ID(core.PK{item.Owner, item.Name}).Cols("apply_result").Update(item)
			continue
		}

		item.Applied = true
		item.AppliedTime = util.GetCurrentTime()
		item.ApplyResult = message
		_, err = ormer.Engine.ID(core.PK{item.Owner, item.Name}).Cols("applied", "applied_time", "apply_result").Update(item)
		if err != nil {
			return nil, err
		}
		result.Applied++
		writeAccessReviewRecord(review, item, actor)
	}
	return result, nil
}

func revokeAccessReviewItem(item *AccessReviewItem) (string, error) {
	userId := util.GetId(item.Owner, item.User)
	switch item.EntitlementType {
	case EntitlementTypeRole:
		role, err := GetRole(item.Entitlement)
		if err != nil {
			return "", err
		}
		if role == nil || !util.InSlice(role.Users, userId) {
			return "not held any more", nil
		}
		role.Users = util.DeleteVal(append([]string{}, role.Users...), userId)
		_, err = UpdateRole(item.Entitlement, role, true, "en")
		if err != nil {
			return "", err
		}
	case EntitlementTypeGroup:
		user, err := getUser(item.Owner, item.User)
		if err != nil {
			return "", err
		}
		if user == nil || !util.InSlice(user.Groups, item.Entitlement) {
			return "not held any more", nil
		}
		_, err = DeleteGroupForUser(userId, item.Entitlement)
		if err != nil {
			return "", err
		}
	case EntitlementTypePermission, EntitlementTypeApplication:
		permission, err := GetPermission(item.Entitlement)
		if err != nil {
			return "", err
		}
		if permission == nil || !util.InSlice(permission.Users, userId) {
			return "not held any more", nil
		}
		permission.Users = util.DeleteVal(append([]string{}, permission.Users...), userId)
		_, err = UpdatePermission(item.Entitlement, permission)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unknown entitlement type: %s", item.EntitlementType)
	}
	return "revoked", nil
}

func writeAccessReviewRecord(review *AccessReview, item *AccessReviewItem, actor string) {
	user := actor
	if user == "" {
		user = AccessReviewDeadlineDecider
	}
	record := &Record{
		Name:         util.GenerateId(),
		CreatedTime:  util.GetCurrentTime(),
		Organization: review.Owner,
		User:         user,
		Method:       "POST",
		RequestUri:   "/api/apply-access-review?id=" + review.GetId(),
		Action:       "apply-access-review",
		Object: util.StructToJson(map[string]interface{}{
			"review":          review.GetId(),
			"round":           item.Round,
			"user":            util.GetId(item.Owner, item.User),
			"entitlementType": item.EntitlementType,
			"entitlement":     item.Entitlement,
			"decision":        item.Decision,
			"reviewer":        item.Reviewer,
			"decidedBy":       item.DecidedBy,
			"comment":         item.Comment,
		}),
		StatusCode: 200,
		Response:   "{status:\"ok\", msg:\"revoked\"}",
	}
	util.SafeGoroutine(func() {
		AddRecord(record)
	})
}

// GetAccessReviewItems lists the items of a round of a campaign, optionally of one reviewer or decision
func GetAccessReviewItems(owner string, review string, round int, reviewer string, decision string) ([]*AccessReviewItem, error) {
	items := []*AccessReviewItem{}
	session := ormer.Engine.Where("owner = ? and review = ? and round = ?", owner, review, round)
	if reviewer != "" {
		session = session.And("reviewer = ?", reviewer)
	}
	if decision != "" {
		session = session.And("decision = ?", decision)
	}
	err := session.Asc("user", "entitlement").Find(&items)
	return items, err
}

// GetReviewerItems lists the items of the running campaigns the user reviews, for the "My access reviews" page
func GetReviewerItems(owner string, reviewer string, pendingOnly bool) ([]*AccessReviewItem, error) {
	reviews := []*AccessReview{}
	err := ormer.Engine.Where("owner = ? and state = ?", owner, AccessReviewStateActive).Find(&reviews)
	if err != nil {
		return nil, err
	}

	res := []*AccessReviewItem{}
	for _, review := range reviews {
		decision := ""
		if pendingOnly {
			decision = AccessReviewDecisionPending
		}
		items, err := GetAccessReviewItems(owner, review.Name, review.Round, reviewer, decision)
		if err != nil {
			return nil, err
		}
		displayName := review.DisplayName
		if displayName == "" {
			displayName = review.Name
		}
		for _, item := range items {
			item.ReviewDisplayName = displayName
			item.EndTime = review.EndTime
		}
		res = append(res, items...)
	}
	return res, nil
}

// GetPendingReviewItemCount counts the undecided items of the running campaigns the user reviews, 0 without an
// enterprise edition license
func GetPendingReviewItemCount(owner string, reviewer string) (int64, error) {
	if !edition.Has(edition.Enterprise) {
		return 0, nil
	}
	reviews := []*AccessReview{}
	err := ormer.Engine.Cols("name", "round").Where("owner = ? and state = ?", owner, AccessReviewStateActive).Find(&reviews)
	if err != nil {
		return 0, err
	}
	var count int64
	for _, review := range reviews {
		n, err := ormer.Engine.Where("owner = ? and review = ? and round = ? and reviewer = ? and decision = ?", owner, review.Name, review.Round, reviewer, AccessReviewDecisionPending).Count(&AccessReviewItem{})
		if err != nil {
			return 0, err
		}
		count += n
	}
	return count, nil
}

// DecideAccessReviewItems records the decisions of a reviewer, an admin of the organization may decide any item,
// nobody decides on their own access
func DecideAccessReviewItems(decider *User, isOrgAdmin bool, names []string, decision string, comment string, lang string) (int, error) {
	if decision != AccessReviewDecisionKeep && decision != AccessReviewDecisionRevoke {
		return 0, fmt.Errorf("unknown decision: %s", decision)
	}
	if len(names) == 0 {
		return 0, nil
	}
	if len(comment) > 2000 {
		comment = comment[:2000]
	}

	items := []*AccessReviewItem{}
	err := ormer.Engine.Where("owner = ?", decider.Owner).In("name", names).Find(&items)
	if err != nil {
		return 0, err
	}
	if len(items) != len(names) {
		return 0, errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
	}

	reviews := map[string]*AccessReview{}
	for _, item := range items {
		review, ok := reviews[item.Review]
		if !ok {
			review, err = getAccessReview(item.Owner, item.Review)
			if err != nil {
				return 0, err
			}
			reviews[item.Review] = review
		}
		if review == nil || review.State != AccessReviewStateActive || review.Round != item.Round {
			return 0, errors.New(i18n.Translate(lang, "enterprise:The access review is not running"))
		}
		if item.User == decider.Name {
			return 0, errors.New(i18n.Translate(lang, "enterprise:You cannot review your own access"))
		}
		if item.Reviewer != decider.Name && !isOrgAdmin {
			return 0, errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
		}
		if item.Applied {
			return 0, errors.New(i18n.Translate(lang, "enterprise:The revocation is already applied"))
		}
	}

	now := util.GetCurrentTime()
	for _, item := range items {
		item.Decision = decision
		item.Comment = comment
		item.DecidedBy = decider.Name
		item.DecidedTime = now
		_, err = ormer.Engine.ID(core.PK{item.Owner, item.Name}).Cols("decision", "comment", "decided_by", "decided_time").Update(item)
		if err != nil {
			return 0, err
		}
	}
	return len(items), nil
}

// AccessReviewReminderResult tells whom the reminder reached
type AccessReviewReminderResult struct {
	Reviewers  []string `json:"reviewers"`
	Sent       int      `json:"sent"`
	Failed     []string `json:"failed"`
	Unassigned int      `json:"unassigned"`
}

// RemindAccessReviewers emails the reviewers who still have undecided items, a reviewer without an email or a
// failing email provider is reported, never an error
func RemindAccessReviewers(review *AccessReview, lang string) (*AccessReviewReminderResult, error) {
	result := &AccessReviewReminderResult{Reviewers: []string{}, Failed: []string{}}
	if review.State != AccessReviewStateActive {
		return nil, errors.New(i18n.Translate(lang, "enterprise:The access review is not running"))
	}
	items, err := GetAccessReviewItems(review.Owner, review.Name, review.Round, "", AccessReviewDecisionPending)
	if err != nil {
		return nil, err
	}
	pending := map[string]int{}
	for _, item := range items {
		if item.Reviewer == "" {
			result.Unassigned++
			continue
		}
		pending[item.Reviewer]++
	}
	organization, err := getOrganization("admin", review.Owner)
	if err != nil {
		return nil, err
	}
	if organization == nil {
		return nil, fmt.Errorf("the organization: %s is not found", review.Owner)
	}

	reviewers := make([]string, 0, len(pending))
	for reviewer := range pending {
		reviewers = append(reviewers, reviewer)
	}
	sort.Strings(reviewers)
	displayName := review.DisplayName
	if displayName == "" {
		displayName = review.Name
	}
	for _, reviewer := range reviewers {
		result.Reviewers = append(result.Reviewers, reviewer)
		user, err := getUser(review.Owner, reviewer)
		if err != nil || user == nil || user.Email == "" {
			result.Failed = append(result.Failed, reviewer)
			continue
		}
		title := fmt.Sprintf(i18n.Translate(lang, "enterprise:Access review reminder - %s"), displayName)
		content := fmt.Sprintf(i18n.Translate(lang, "enterprise:You have %d access items to review by %s. Open My access reviews in the console to keep or revoke them."), pending[reviewer], review.EndTime)
		err = sendOrganizationEmail(organization, title, content, []string{user.Email})
		if err != nil {
			logs.Warn("the access review reminder to %s is not sent: %s", user.GetId(), err.Error())
			result.Failed = append(result.Failed, reviewer)
			continue
		}
		result.Sent++
	}
	return result, nil
}

// ExportAccessReviewCsv is the round of the campaign as a CSV for the auditors
func ExportAccessReviewCsv(review *AccessReview, round int) ([]byte, error) {
	items, err := GetAccessReviewItems(review.Owner, review.Name, round, "", "")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.WriteString("\uFEFF")
	writer := csv.NewWriter(&buffer)
	rows := [][]string{{"Campaign", "Round", "Start time", "End time", "User", "User display name", "Entitlement type", "Entitlement", "Entitlement display name", "Reviewer", "Decision", "Decided by", "Decided time", "Comment", "Applied", "Applied time", "Apply result"}}
	startTime, endTime := "", ""
	if round == review.Round {
		startTime, endTime = review.StartTime, review.EndTime
	}
	for _, item := range items {
		rows = append(rows, []string{
			review.GetId(), fmt.Sprintf("%d", item.Round), startTime, endTime,
			item.User, item.UserDisplayName, item.EntitlementType, item.Entitlement, item.EntitlementDisplayName,
			item.Reviewer, item.Decision, item.DecidedBy, item.DecidedTime, item.Comment,
			fmt.Sprintf("%t", item.Applied), item.AppliedTime, item.ApplyResult,
		})
	}
	for _, row := range rows {
		for i, cell := range row {
			// keep a spreadsheet from running a cell as a formula
			if cell != "" && strings.ContainsRune("=+-@", rune(cell[0])) {
				row[i] = "'" + cell
			}
		}
	}
	if err = writer.WriteAll(rows); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// RunAccessReviewJob closes the rounds past their deadline and starts the next round of the recurring campaigns
func RunAccessReviewJob() error {
	if !edition.Has(edition.Enterprise) {
		return nil
	}
	reviews := []*AccessReview{}
	err := ormer.Engine.Where("state = ? or (state = ? and next_start_time <> ?)", AccessReviewStateActive, AccessReviewStateCompleted, "").Find(&reviews)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, review := range reviews {
		switch review.State {
		case AccessReviewStateActive:
			endTime, ok := parseAnyTime(review.EndTime)
			if !ok || now.Before(endTime) {
				continue
			}
			result, err := CompleteAccessReview(review, "", "en")
			if err != nil {
				logs.Error("failed to complete the access review: %s: %v", review.GetId(), err)
				continue
			}
			logs.Info("completed the access review: %s, %d revocations applied, %d failed", review.GetId(), result.Applied, result.Failed)
		case AccessReviewStateCompleted:
			nextStartTime, ok := parseAnyTime(review.NextStartTime)
			if !ok || now.Before(nextStartTime) || review.RecurrenceDays <= 0 {
				continue
			}
			count, err := StartAccessReview(review, "en")
			if err != nil {
				logs.Error("failed to start the next round of the access review: %s: %v", review.GetId(), err)
				continue
			}
			logs.Info("started round %d of the access review: %s with %d items", review.Round, review.GetId(), count)
		}
	}
	return nil
}

// InitAccessReviewJobs checks the campaigns every 10 minutes
func InitAccessReviewJobs() {
	cronJob := cron.New()
	_, err := cronJob.AddFunc("*/10 * * * *", func() {
		if err := RunAccessReviewJob(); err != nil {
			logs.Error("the access review job failed: %v", err)
		}
	})
	if err != nil {
		logs.Warn("the access review job is not scheduled: %s", err.Error())
		return
	}
	cronJob.Start()
}
