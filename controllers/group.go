// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
)

// GetGroups
// @Title GetGroups
// @Tag Group API
// @Description 获取分组列表
// @Param   owner     query    string  true        "组织ID"
// @Param   pageSize  query    int     false        "分页大小"
// @Param   p         query    int     false        "分页"
// @Param   query     query    string     false        "查询内容（名称）"
// @Param   sortField     query    string     false        "排序字段"
// @Param   sortOrder     query    string     false        "排序方式: asc, desc"
// @Success 200 {array} object.Group The Response object
// @router /api/groups [get]
func (c *ApiController) GetGroups() {
	// owner := c.Input().Get("owner")
	// limit := c.Input().Get("pageSize")
	// page := c.Input().Get("p")
	// query := c.Input().Get("query")
	// sortField := c.Input().Get("sortField")
	// sortOrder := c.Input().Get("sortOrder")

	params := c.GetQueryParams()
	withTree := c.Input().Get("withTree")

	if params.Limit == 0 || params.Page == 0 {
		groups, err := object.GetGroups(params.Owner)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		err = object.ExtendGroupsWithUsers(groups)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		if withTree == "true" {
			c.ResponseOk(object.ConvertToTreeData(groups, params.Owner))
			return
		}

		c.ResponseOk(groups)
	} else {
		limit := params.Limit
		count, err := object.GetGroupCount(params.Owner, params.Query)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		groups, err := object.GetPaginationGroups(params.Owner, params)
		if err != nil {
			c.ResponseErr(err)
			return
		}
		groupsHaveChildrenMap, err := object.GetGroupsHaveChildrenMap(groups)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		for _, group := range groups {
			_, ok := groupsHaveChildrenMap[group.GetId()]
			if ok {
				group.HaveChildren = true
			}

			parent, ok := groupsHaveChildrenMap[fmt.Sprintf("%s/%s", group.Owner, group.ParentId)]
			if ok {
				group.ParentName = parent.DisplayName
			}
		}

		err = object.ExtendGroupsWithUsers(groups)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		c.ResponseOk(groups, paginator.Nums())

	}
}

// GetGroup
// @Title GetGroup
// @Tag Group API
// @Description get group
// @Param   id     query    string  true        "The id ( owner/name ) of the group"
// @Success 200 {object} object.Group The Response object
// @router /get-group [get]
func (c *ApiController) GetGroup() {
	id := c.Input().Get("id")

	group, err := object.GetGroup(id)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	err = object.ExtendGroupWithUsers(group)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.ResponseOk(group)
}

// UpdateGroup
// @Title UpdateGroup
// @Tag Group API
// @Description update group
// @Param   id     query    string  true        "The id ( owner/name ) of the group"
// @Param   body    body   object.Group  true        "The details of the group"
// @Success 200 {object} controllers.Response The Response object
// @router /update-group [post]
func (c *ApiController) UpdateGroup() {
	id := c.Input().Get("id")

	var group object.Group
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &group)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateGroup(id, &group))
	c.ServeJSON()
}

// AddGroup
// @Title AddGroup
// @Tag Group API
// @Description add group
// @Param   body    body   object.Group  true      "The details of the group"
// @Success 200 {object} controllers.Response The Response object
// @router /add-group [post]
func (c *ApiController) AddGroup() {
	var group object.Group
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &group)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddGroup(&group))
	c.ServeJSON()
}

// DeleteGroup
// @Title DeleteGroup
// @Tag Group API
// @Description delete group
// @Param   body    body   object.Group  true        "The details of the group"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-group [post]
func (c *ApiController) DeleteGroup() {
	var group object.Group
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &group)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteGroup(&group))
	c.ServeJSON()
}

func (c *ApiController) UpdateGroupUser() {

}
