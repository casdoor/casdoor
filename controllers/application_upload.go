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
	"fmt"
	"os"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// UploadApplications
// @Title UploadApplications
// @Tag Application API
// @Description add the applications of a spreadsheet, keeping their client IDs, client secrets and redirect URIs
// @Param   file formData file true "The .xlsx file"
// @Success 200 {object} object.ApplicationUploadResult The Response object
// @router /upload-applications [post]
func (c *ApiController) UploadApplications() {
	if !c.IsAdmin() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	organization, ok := c.getUploadOrganization()
	if !ok {
		return
	}

	userId := c.GetSessionUsername()
	owner, user, err := util.GetOwnerAndNameFromIdWithError(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	file, header, err := c.Ctx.Request.FormFile("file")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	fileId := fmt.Sprintf("%s_%s_%s", owner, user, util.RemoveExt(header.Filename))
	path := util.GetUploadXlsxPath(fileId)
	defer os.Remove(path)

	err = saveFile(path, &file)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	result, err := object.UploadApplications(path, organization, c.GetAcceptLanguage(), checkQuotaForApplication)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}
