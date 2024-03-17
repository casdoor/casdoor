package controllers

import (
	"fmt"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// FaceIDSigninBegin
// @Title FaceIDSigninBegin
// @Tag Login API
// @Description FaceId Login Flow 1st stage
// @Param   owner     query    string  true        "owner"
// @Param   name     query    string  true        "name"
// @Success 200 {object} controllers.Response The Response object
// @router /faceid-signin-begin [get]
func (c *ApiController) FaceIDSigninBegin() {
	userOwner := c.Input().Get("owner")
	userName := c.Input().Get("name")
	user, err := object.GetUserByFields(userOwner, userName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(userOwner, userName)))
		return
	}
	if len(user.FaceIds) == 0 {
		c.ResponseError(c.T("check:Face data does not exist, cannot log in"))
		return
	}

	resp := &Response{Status: "ok", Msg: ""}
	c.Data["json"] = resp
	c.ServeJSON()
}
