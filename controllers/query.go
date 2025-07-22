package controllers

import (
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)



func (c *ApiController) GetQueryParams() *object.QueryParams {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	query := c.Input().Get("query")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	organization := c.Input().Get("organization")
	// 使用用户默认origanization
	if organization == "" {
		userId := c.GetSessionUsername()
		if userId != ""{
			organization, _ = util.GetOwnerAndNameFromId(userId)
		}
	}

	return &object.QueryParams{
		Owner:     owner,
		Limit:     util.ParseInt(limit),
		Page:      util.ParseInt(page),
		Query:     query,
		SortField: sortField,
		SortOrder: sortOrder,
		Organization: organization,
	}
}


func (c *ApiController) Query() {
	params := c.GetQueryParams()
	userId := c.GetSessionUsername()
	var err error

		if limit == "" || page == "" {
		var applications []*object.Application
		if organization == "" {
			applications, err = object.GetApplications(owner)
		} else {
			applications, err = object.GetOrganizationApplications(owner, organization)
		}
		if err != nil {
			c.ResponseErr(err)
			return
		}
		c.ResponseOk(object.GetMaskedApplications(applications, userId))
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetApplicationCount(owner, field, value)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		application, err := object.GetPaginationApplications(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		applications := object.GetMaskedApplications(application, userId)
		c.ResponseOk(applications, paginator.Nums())
	}
}