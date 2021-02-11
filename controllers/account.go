package controllers

import (
	"encoding/json"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type RegisterForm struct {
	Organization string `json:"organization"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

// @Title Register
// @Description register a new user
// @Param   username     formData    string  true        "The username to register"
// @Param   password     formData    string  true        "The password"
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /register [post]
func (c *ApiController) Register() {
	var resp Response

	if c.GetSessionUser() != "" {
		resp = Response{Status: "error", Msg: "please log out first before signing up", Data: c.GetSessionUser()}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	var form RegisterForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		panic(err)
	}

	msg := object.CheckUserRegister(form.Username, form.Password)
	if msg != "" {
		resp = Response{Status: "error", Msg: msg, Data: ""}
	} else {
		user := &object.User{
			Owner:        form.Organization,
			Name:         form.Username,
			CreatedTime:  util.GetCurrentTime(),
			Password:     form.Password,
			PasswordType: "plain",
			DisplayName:  form.Name,
			Email:        form.Email,
			Phone:        form.Phone,
		}
		object.AddUser(user)

		//c.SetSessionUser(user)

		util.LogInfo(c.Ctx, "API: [%s] is registered as new user", user)
		resp = Response{Status: "ok", Msg: "注册成功", Data: user}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Login
// @Description login as a user
// @Param   username     formData    string  true        "The username to login"
// @Param   password     formData    string  true        "The password"
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /login [post]
func (c *ApiController) Login() {
	var resp Response

	if c.GetSessionUser() != "" {
		resp = Response{Status: "error", Msg: "please log out first before signing in", Data: c.GetSessionUser()}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	var form RegisterForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		panic(err)
	}

	user, password := form.Username, form.Password
	msg := object.CheckUserLogin(user, password)

	if msg != "" {
		resp = Response{Status: "error", Msg: msg, Data: ""}
	} else {
		c.SetSessionUser(user)

		util.LogInfo(c.Ctx, "API: [%s] logged in", user)
		resp = Response{Status: "ok", Msg: "", Data: user}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Logout
// @Description logout the current user
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /logout [post]
func (c *ApiController) Logout() {
	var resp Response

	user := c.GetSessionUser()
	util.LogInfo(c.Ctx, "API: [%s] logged out", user)

	c.SetSessionUser("")

	resp = Response{Status: "ok", Msg: "", Data: user}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) GetAccount() {
	var resp Response

	if c.GetSessionUser() == "" {
		resp = Response{Status: "error", Msg: "please sign in first", Data: c.GetSessionUser()}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	username := c.GetSessionUser()
	userObj := object.GetUser(username)
	resp = Response{Status: "ok", Msg: "", Data: util.StructToJson(userObj)}

	c.Data["json"] = resp
	c.ServeJSON()
}
