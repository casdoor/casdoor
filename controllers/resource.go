// Copyright 2021 The casbin Authors. All Rights Reserved.
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

	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

func (c *ApiController) GetResources() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetResources(owner)
	c.ServeJSON()
}

func (c *ApiController) GetResource() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetResource(id)
	c.ServeJSON()
}

func (c *ApiController) UpdateResource() {
	id := c.Input().Get("id")

	var resource object.Resource
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &resource)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateResource(id, &resource))
	c.ServeJSON()
}

func (c *ApiController) AddResource() {
	var resource object.Resource
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &resource)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddResource(&resource))
	c.ServeJSON()
}

func (c *ApiController) DeleteResource() {
	var resource object.Resource
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &resource)
	if err != nil {
		panic(err)
	}

	provider, _, ok := c.GetProviderFromContext("Storage")
	if !ok {
		return
	}

	err = object.DeleteFile(provider, resource.Name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteResource(&resource))
	c.ServeJSON()
}

func (c *ApiController) UploadResource() {
	owner := c.Input().Get("owner")
	username := c.Input().Get("user")
	application := c.Input().Get("application")
	tag := c.Input().Get("tag")
	parent := c.Input().Get("parent")
	fullFilePath := c.Input().Get("fullFilePath")

	file, header, err := c.GetFile("file")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	defer file.Close()

	filename := filepath.Base(fullFilePath)
	fileBuffer := bytes.NewBuffer(nil)
	if _, err = io.Copy(fileBuffer, file); err != nil {
		c.ResponseError(err.Error())
		return
	}

	provider, user, ok := c.GetProviderFromContext("Storage")
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

	fileUrl, objectKey, err := object.UploadFile(provider, fullFilePath, fileBuffer)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	fileFormat := filepath.Ext(fullFilePath)
	fileSize := int(header.Size)
	resource := &object.Resource{
		Owner:       owner,
		Name:        objectKey,
		CreatedTime: util.GetCurrentTime(),
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
	}
	object.AddOrUpdateResource(resource)

	switch tag {
	case "avatar":
		if user == nil {
			c.ResponseError("user is nil for tag: \"avatar\"")
			return
		}

		user.Avatar = fileUrl
		object.UpdateUser(user.GetId(), user)
	case "termsOfUse":
		applicationId := fmt.Sprintf("admin/%s", parent)
		app := object.GetApplication(applicationId)
		app.TermsOfUse = fileUrl
		object.UpdateApplication(applicationId, app)
	}

	c.ResponseOk(fileUrl, objectKey)
}
