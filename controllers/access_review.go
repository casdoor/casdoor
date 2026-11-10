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
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// isAdminOfOwner tells whether the current user administers the organization: a global admin or an admin of it
func (c *ApiController) isAdminOfOwner(owner string) bool {
	user := c.getCurrentUser()
	return owner != "" && user != nil && (user.IsGlobalAdmin() || (user.IsAdmin && user.Owner == owner))
}

// requireAccessReview checks the license and that the current user administers the organization of the campaign in
// the query parameter "id", then returns the campaign
func (c *ApiController) requireAccessReview() (*object.AccessReview, bool) {
	if err := edition.Require(edition.Enterprise, edition.FeatureAccessReview); err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	id := c.Ctx.Input.Query("id")
	owner, _, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if !c.isAdminOfOwner(owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	review, err := object.GetAccessReview(id)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if review == nil {
		c.ResponseError(fmt.Sprintf(c.T("enterprise:The access review: %s doesn't exist"), id))
		return nil, false
	}
	return review, true
}

func (c *ApiController) currentUserName() string {
	user := c.getCurrentUser()
	if user == nil {
		return ""
	}
	return user.Name
}

// GetAccessReviews
// @Title GetAccessReviews
// @Tag Enterprise API
// @Description list the access review campaigns of an organization with the progress of their current rounds
// @Param   owner     query    string  true        "The organization"
// @Success 200 {array} object.AccessReview The Response object
// @router /get-access-reviews [get]
func (c *ApiController) GetAccessReviews() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureAccessReview)
	if !ok {
		return
	}
	limit := util.ParseInt(c.Ctx.Input.Query("pageSize"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	count, err := object.GetAccessReviewCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	reviews, err := object.GetPaginationAccessReviews(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(reviews, paginator.Nums())
}

// GetAccessReview
// @Title GetAccessReview
// @Tag Enterprise API
// @Description get an access review campaign
// @Param   id     query    string  true        "The id ( owner/name ) of the campaign"
// @Success 200 {object} object.AccessReview The Response object
// @router /get-access-review [get]
func (c *ApiController) GetAccessReview() {
	review, ok := c.requireAccessReview()
	if !ok {
		return
	}
	c.ResponseOk(review)
}

// AddAccessReview
// @Title AddAccessReview
// @Tag Enterprise API
// @Description add an access review campaign as a draft
// @Param   body    body   object.AccessReview  true        "The campaign"
// @Success 200 {object} controllers.Response The Response object
// @router /add-access-review [post]
func (c *ApiController) AddAccessReview() {
	if err := edition.Require(edition.Enterprise, edition.FeatureAccessReview); err != nil {
		c.ResponseError(err.Error())
		return
	}
	var review object.AccessReview
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &review)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !c.isAdminOfOwner(review.Owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.Data["json"] = wrapActionResponse(object.AddAccessReview(&review, c.GetAcceptLanguage()))
	c.ServeJSON()
}

// UpdateAccessReview
// @Title UpdateAccessReview
// @Tag Enterprise API
// @Description change the settings of an access review campaign that isn't running
// @Param   id     query    string  true        "The id ( owner/name ) of the campaign"
// @Param   body    body   object.AccessReview  true        "The campaign"
// @Success 200 {object} controllers.Response The Response object
// @router /update-access-review [post]
func (c *ApiController) UpdateAccessReview() {
	review, ok := c.requireAccessReview()
	if !ok {
		return
	}
	var newReview object.AccessReview
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &newReview)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = wrapActionResponse(object.UpdateAccessReview(review.GetId(), &newReview, c.GetAcceptLanguage()))
	c.ServeJSON()
}

// DeleteAccessReview
// @Title DeleteAccessReview
// @Tag Enterprise API
// @Description delete an access review campaign that isn't running with its items
// @Param   body    body   object.AccessReview  true        "The campaign"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-access-review [post]
func (c *ApiController) DeleteAccessReview() {
	if err := edition.Require(edition.Enterprise, edition.FeatureAccessReview); err != nil {
		c.ResponseError(err.Error())
		return
	}
	var review object.AccessReview
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &review)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !c.isAdminOfOwner(review.Owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.Data["json"] = wrapActionResponse(object.DeleteAccessReview(&review, c.GetAcceptLanguage()))
	c.ServeJSON()
}

// StartAccessReview
// @Title StartAccessReview
// @Tag Enterprise API
// @Description start the next round of an access review campaign
// @Param   id     query    string  true        "The id ( owner/name ) of the campaign"
// @Success 200 {object} controllers.Response The Response object
// @router /start-access-review [post]
func (c *ApiController) StartAccessReview() {
	review, ok := c.requireAccessReview()
	if !ok {
		return
	}
	count, err := object.StartAccessReview(review, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(count)
}

// CloseAccessReview
// @Title CloseAccessReview
// @Tag Enterprise API
// @Description close the running round of an access review campaign before its deadline
// @Param   id     query    string  true        "The id ( owner/name ) of the campaign"
// @Success 200 {object} object.AccessReviewApplyResult The Response object
// @router /close-access-review [post]
func (c *ApiController) CloseAccessReview() {
	review, ok := c.requireAccessReview()
	if !ok {
		return
	}
	result, err := object.CompleteAccessReview(review, c.currentUserName(), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}

// ApplyAccessReview
// @Title ApplyAccessReview
// @Tag Enterprise API
// @Description remove the users from the roles, groups and permissions revoked in the current round
// @Param   id     query    string  true        "The id ( owner/name ) of the campaign"
// @Success 200 {object} object.AccessReviewApplyResult The Response object
// @router /apply-access-review [post]
func (c *ApiController) ApplyAccessReview() {
	review, ok := c.requireAccessReview()
	if !ok {
		return
	}
	result, err := object.ApplyAccessReview(review, c.currentUserName())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}

// RemindAccessReview
// @Title RemindAccessReview
// @Tag Enterprise API
// @Description email the reviewers who still have items to decide
// @Param   id     query    string  true        "The id ( owner/name ) of the campaign"
// @Success 200 {object} object.AccessReviewReminderResult The Response object
// @router /remind-access-review [post]
func (c *ApiController) RemindAccessReview() {
	review, ok := c.requireAccessReview()
	if !ok {
		return
	}
	result, err := object.RemindAccessReviewers(review, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}

// GetAccessReviewItems
// @Title GetAccessReviewItems
// @Tag Enterprise API
// @Description list the items of a round of an access review campaign
// @Param   owner     query    string  true        "The organization"
// @Param   review    query    string  true        "The name of the campaign"
// @Param   round     query    int     false       "The round, the current one by default"
// @Param   reviewer  query    string  false       "Only the items of this reviewer"
// @Param   decision  query    string  false       "Only the items with this decision"
// @Success 200 {array} object.AccessReviewItem The Response object
// @router /get-access-review-items [get]
func (c *ApiController) GetAccessReviewItems() {
	owner, ok := c.requireEnterpriseAdmin(edition.FeatureAccessReview)
	if !ok {
		return
	}
	review, err := object.GetAccessReview(util.GetId(owner, c.Ctx.Input.Query("review")))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if review == nil {
		c.ResponseOk([]*object.AccessReviewItem{})
		return
	}
	round := util.ParseInt(c.Ctx.Input.Query("round"))
	if round <= 0 {
		round = review.Round
	}
	items, err := object.GetAccessReviewItems(owner, review.Name, round, c.Ctx.Input.Query("reviewer"), c.Ctx.Input.Query("decision"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(items)
}

// ExportAccessReview
// @Title ExportAccessReview
// @Tag Enterprise API
// @Description download a round of an access review campaign as a CSV for the auditors
// @Param   id     query    string  true        "The id ( owner/name ) of the campaign"
// @Param   round  query    int     false       "The round, the current one by default"
// @Success 200 {string} string "The CSV file"
// @router /export-access-review [get]
func (c *ApiController) ExportAccessReview() {
	review, ok := c.requireAccessReview()
	if !ok {
		return
	}
	round := util.ParseInt(c.Ctx.Input.Query("round"))
	if round <= 0 || round > review.Round {
		round = review.Round
	}
	data, err := object.ExportAccessReviewCsv(review, round)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Ctx.Output.Header("Content-Type", "text/csv; charset=utf-8")
	c.Ctx.Output.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"access-review-%s-%s-round-%d.csv\"", review.Owner, review.Name, round))
	err = c.Ctx.Output.Body(data)
	if err != nil {
		c.ResponseError(err.Error())
	}
}

// GetMyAccessReviewItems
// @Title GetMyAccessReviewItems
// @Tag Enterprise API
// @Description list the items of the running access reviews that the current user reviews
// @Param   pending     query    string  false        "1 for the undecided items only"
// @Success 200 {array} object.AccessReviewItem The Response object
// @router /get-my-access-review-items [get]
func (c *ApiController) GetMyAccessReviewItems() {
	if err := edition.Require(edition.Enterprise, edition.FeatureAccessReview); err != nil {
		c.ResponseError(err.Error())
		return
	}
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}
	items, err := object.GetReviewerItems(user.Owner, user.Name, c.Ctx.Input.Query("pending") == "1")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(items)
}

// DecideAccessReviewItems
// @Title DecideAccessReviewItems
// @Tag Enterprise API
// @Description keep or revoke access review items, one by one or in bulk
// @Param   body    body   object.AccessReviewItem  true        "{items: [names], decision: Keep or Revoke, comment}"
// @Success 200 {object} controllers.Response The Response object
// @router /decide-access-review-items [post]
func (c *ApiController) DecideAccessReviewItems() {
	if err := edition.Require(edition.Enterprise, edition.FeatureAccessReview); err != nil {
		c.ResponseError(err.Error())
		return
	}
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}
	if object.IsAppUser(c.GetSessionUsername()) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	var request struct {
		Items    []string `json:"items"`
		Decision string   `json:"decision"`
		Comment  string   `json:"comment"`
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	count, err := object.DecideAccessReviewItems(user, user.IsAdmin, request.Items, request.Decision, request.Comment, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(count)
}
