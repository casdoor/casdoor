// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetResources
// @router /get-resources [get]
// @Tag Resource API
// @Title GetResources
func (c *ApiController) GetResources() {
	owner := c.Input().Get("owner")
	user := c.Input().Get("user")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	userObj, ok := c.RequireSignedInUser()
	if !ok {
		return
	}
	if userObj.IsAdmin {
		user = ""
	}

	if limit == "" || page == "" {
		c.Data["json"] = object.GetResources(owner, user)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetResourceCount(owner, user, field, value)))
		resources := object.GetPaginationResources(owner, user, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(resources, paginator.Nums())
	}
}

// GetResource
// @Tag Resource API
// @Title GetResource
// @router /get-resource [get]
func (c *ApiController) GetResource() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetResource(id)
	c.ServeJSON()
}

// UpdateResource
// @Tag Resource API
// @Title UpdateResource
// @router /update-resource [post]
func (c *ApiController) UpdateResource() {
	id := c.Input().Get("id")

	var resource object.Resource
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &resource)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateResource(id, &resource))
	c.ServeJSON()
}

// AddResource
// @Tag Resource API
// @Title AddResource
// @router /add-resource [post]
func (c *ApiController) AddResource() {
	var resource object.Resource
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &resource)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddResource(&resource))
	c.ServeJSON()
}

// DeleteResource
// @Tag Resource API
// @Title DeleteResource
// @router /delete-resource [post]
func (c *ApiController) DeleteResource() {
	var resource object.Resource
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &resource)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	provider, _, ok := c.GetProviderFromContext("Storage")
	if !ok {
		return
	}

	err = object.DeleteFile(provider, resource.Name, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteResource(&resource))
	c.ServeJSON()
}

// UploadResource
// @Tag Resource API
// @Title UploadResource
// @router /upload-resource [post]
func (c *ApiController) UploadResource() {
	owner := c.Input().Get("owner")
	username := c.Input().Get("user")
	application := c.Input().Get("application")
	tag := c.Input().Get("tag")
	parent := c.Input().Get("parent")
	fullFilePath := c.Input().Get("fullFilePath")
	createdTime := c.Input().Get("createdTime")
	description := c.Input().Get("description")

	file, header, err := c.GetFile("file")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	defer file.Close()

	if username == "" || fullFilePath == "" {
		c.ResponseError(fmt.Sprintf(c.T("resource:Username or fullFilePath is empty: username = %s, fullFilePath = %s"), username, fullFilePath))
		return
	}

	filename := filepath.Base(fullFilePath)
	fileBuffer := bytes.NewBuffer(nil)
	if _, err = io.Copy(fileBuffer, file); err != nil {
		c.ResponseError(err.Error())
		return
	}

	provider, _, ok := c.GetProviderFromContext("Storage")
	if !ok {
		return
	}

	fileType := "unknown"
	contentType := header.Header.Get("Content-Type")
	fileType, _ = util.GetOwnerAndNameFromId(contentType)

	if fileType != "image" && fileType != "video" {
		ext := filepath.Ext(filename)
		mimeType := mime.TypeByExtension(ext)
		fileType, _ = util.GetOwnerAndNameFromId(mimeType)
	}

	fullFilePath = object.GetTruncatedPath(provider, fullFilePath, 175)
	if tag != "avatar" && tag != "termsOfUse" {
		ext := filepath.Ext(filepath.Base(fullFilePath))
		index := len(fullFilePath) - len(ext)
		for i := 1; ; i++ {
			_, objectKey := object.GetUploadFileUrl(provider, fullFilePath, true)
			if object.GetResourceCount(owner, username, "name", objectKey) == 0 {
				break
			}

			// duplicated fullFilePath found, change it
			fullFilePath = fullFilePath[:index] + fmt.Sprintf("-%d", i) + ext
		}
	}

	fileUrl, objectKey, err := object.UploadFileSafe(provider, fullFilePath, fileBuffer, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if createdTime == "" {
		createdTime = util.GetCurrentTime()
	}
	fileFormat := filepath.Ext(fullFilePath)
	fileSize := int(header.Size)
	resource := &object.Resource{
		Owner:       owner,
		Name:        objectKey,
		CreatedTime: createdTime,
		User:        username,
		Provider:    provider.Name,
		Application: application,
		Tag:         tag,
		Parent:      parent,
		FileName:    filename,
		FileType:    fileType,
		FileFormat:  fileFormat,
		FileSize:    fileSize,
		Url:         fileUrl,
		Description: description,
	}
	object.AddOrUpdateResource(resource)

	switch tag {
	case "avatar":
		user := object.GetUserNoCheck(util.GetId(owner, username))
		if user == nil {
			c.ResponseError(c.T("resource:User is nil for tag: avatar"))
			return
		}

		user.Avatar = fileUrl
		object.UpdateUser(user.GetId(), user, []string{"avatar"}, false)
	case "termsOfUse":
		user := object.GetUserNoCheck(util.GetId(owner, username))
		if user == nil {
			c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(owner, username)))
			return
		}

		if !user.IsAdminUser() {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}

		_, applicationId := util.GetOwnerAndNameFromIdNoCheck(strings.TrimRight(fullFilePath, ".html"))
		applicationObj := object.GetApplication(applicationId)
		applicationObj.TermsOfUse = fileUrl
		object.UpdateApplication(applicationId, applicationObj)
	}

	c.ResponseOk(fileUrl, objectKey)
}
