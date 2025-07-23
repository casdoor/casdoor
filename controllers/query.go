package controllers

import (
	"github.com/beego/beego/utils/pagination"
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
		if userId != "" {
			organization, _ = util.GetOwnerAndNameFromId(userId)
		}
	}

	return &object.QueryParams{
		Owner:        owner,
		Limit:        util.ParseInt(limit),
		Page:         util.ParseInt(page),
		Query:        query,
		SortField:    sortField,
		SortOrder:    sortOrder,
		Organization: organization,
	}
}

func Query[T any](
	c *ApiController,
	matchFilters object.Cond,
	queryFilters object.Cond,
	params *object.QueryParams,
) (data []*T, count int64, err error) {
	if params.Limit == 0 || params.Page == 0 {
		resources, err := object.QueryResources[T](params.Owner,params,  queryFilters, matchFilters)
		if err != nil {
			return nil, -1, err
		}
		return resources, int64(len(resources)), nil
	}
	limit := params.Limit
	count, err = object.GetResourcesCount[T](params.Owner, params, queryFilters, matchFilters)
	if err != nil {
		return nil, -1, err
	}

	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	resources, err := object.QueryResources[T](params.Owner, params,  queryFilters, matchFilters)
	if err != nil {
		return nil, -1, err
	}
	return resources, paginator.Nums(), nil

}
