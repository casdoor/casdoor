package controllers

import "github.com/casdoor/casdoor/object"

func (c *ApiController) SendVerificationCode() {
	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	remoteAddr := c.Ctx.Request.RemoteAddr

	if len(destType) == 0 || len(dest) == 0 {
		c.Data["json"] = Response{Status: "error", Msg: "Missing parameter."}
		c.ServeJSON()
		return
	}

	ret := "Invalid dest type."
	switch destType {
	case "email":
		ret = object.SendVerificationCodeToEmail(remoteAddr, dest)
	}

	var status string
	if len(ret) == 0 {
		status = "ok"
	} else {
		status = "error"
	}

	c.Data["json"] = Response{Status: status, Msg: ret}
	c.ServeJSON()
}

func (c *ApiController) ResetEmailOrPhone() {
	userId := c.GetSessionUser()
	if len(userId) == 0 {
		c.ResponseError("Please sign in first")
		return
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("No such user.")
		return
	}

	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	code := c.Ctx.Request.Form.Get("code")
	if len(dest) == 0 || len(code) == 0 || len(destType) == 0 {
		c.ResponseError("Missing parameter.")
		return
	}

	if ret := object.CheckVerificationCode(dest, code); len(ret) != 0 {
		c.ResponseError(ret)
		return
	}

	switch destType {
	case "email":
		user.Email = dest
		object.SetUserField(user, "email", user.Email)
	default:
		c.ResponseError("Unknown type.")
		return
	}

	c.Data["json"] = Response{Status: "ok"}
	c.ServeJSON()
}
