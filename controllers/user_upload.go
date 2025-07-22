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
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func saveFile(path string, file *multipart.File) (err error) {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, *file)
	if err != nil {
		return err
	}
	return nil
}

func (c *ApiController) UploadUsers() {
	userId := c.GetSessionUsername()
	owner, user := util.GetOwnerAndNameFromId(userId)

	file, header, err := c.Ctx.Request.FormFile("file")
	if err != nil {
		c.ResponseErr(err)
		return
	}

	fileId := fmt.Sprintf("%s_%s_%s", owner, user, util.RemoveExt(header.Filename))
	path := util.GetUploadXlsxPath(fileId)
	defer os.Remove(path)
	err = saveFile(path, &file)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	affected, err := object.UploadUsers(owner, path)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	if affected {
		c.ResponseOk()
	} else {
		c.ResponseError(c.T("general:Failed to import users"))
	}
}
