package controllers

import (
	"fmt"
	"os"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) UploadGroups() {
	userId := c.GetSessionUsername()
	owner, user := util.GetOwnerAndNameFromId(userId)

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

	affected, err := object.UploadGroups(owner, path)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if affected {
		c.ResponseOk()
	} else {
		c.ResponseError(c.T("group_upload:Failed to import groups"))
	}
}
