package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/object"
)

// TwoFactorSetupInitiate
// @Title TwoFactorSetupInitiate
// @Tag Two-Factor API
// @Description setup totp
// @param userId	form	string	true	" "<owner>/<name>" of user"
// @Success 200 {object}   The Response object
// @router /two-factor/setup/initiate [post]
func (c *ApiController) TwoFactorSetupInitiate() {
	userId := c.Ctx.Request.Form.Get("userId")
	authType := c.Ctx.Request.Form.Get("type")
	application := c.Ctx.Request.Form.Get("application")
	if len(userId) == 0 {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	twoFactorUtil := object.GetTwoFactorUtil(authType)
	if twoFactorUtil == nil {
		c.ResponseError("Invalid auth type")
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	issuer := beego.AppConfig.String("appname")
	accountName := fmt.Sprintf("%s/%s", application, user.Name)

	twoFactorProps, err := twoFactorUtil.Initiate(issuer, accountName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := twoFactorProps
	c.ResponseOk(resp)
}

// TwoFactorSetupVerity
// @Title TwoFactorSetupVerity
// @Tag Two-Factor API
// @Description setup verity totp
// @param	secret		form	string	true	"totp secret"
// @param	passcode	form 	string 	true	"totp passcode"
// @Success 200 {object}  Response object
// @router /two-factor/setup/totp/verity [post]
func (c *ApiController) TwoFactorSetupVerity() {
	authType := c.Ctx.Request.Form.Get("type")
	passcode := c.Ctx.Request.Form.Get("passcode")
	twoFactorUtil := object.GetTwoFactorUtil(authType)

	ok := twoFactorUtil.Verify(c.Ctx, passcode)
	if ok {
		c.ResponseOk(http.StatusText(http.StatusOK))
	} else {
		c.ResponseError(http.StatusText(http.StatusUnauthorized))
	}
}

// TwoFactorSetupEnable
// @Title TwoFactorSetupEnable
// @Tag Two-Factor API
// @Description enable totp
// @param	userId		form	string	true	"Id of user"
// @param  	secret		form	string	true	"totp secret"
// @Success 200 {object}  Response object
// @router /two-factor/setup/enable [post]
func (c *ApiController) TwoFactorSetupEnable() {
	userId := c.Ctx.Request.Form.Get("userId")
	authType := c.Ctx.Request.Form.Get("type")
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	twoFactor := object.GetTwoFactorUtil(authType)
	err := twoFactor.Enable(c.Ctx)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	props, err := json.Marshal(twoFactor.GetProps())
	if err != nil {
		return
	}

	if !object.SetUserField(user, "two_factor_authentications", string(props)) {
		c.ResponseError("Failed to enable two factor authentication")
		return
	}

	c.ResponseOk(http.StatusText(http.StatusOK))
}

// TwoFactorAuthVerity
// @Title TwoFactorAuthVerity
// @Tag Totp API
// @Description Auth Totp
// @param	passcode	form	string	true	"totp passcode"
// @Success 200 {object}  Response object
// @router /two-factor/auth/verity [post]
func (c *ApiController) TwoFactorAuthVerity() {
	authType := c.Ctx.Request.Form.Get("type")
	passcode := c.Ctx.Request.Form.Get("passcode")
	totpSessionData := c.getTfaSessionData()
	if totpSessionData == nil {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	twoFactorUtil := object.GetTwoFactorUtil(authType)
	user := object.GetUser(totpSessionData.UserId)
	if user == nil {
		c.ResponseError("User does not exist")
		return
	}

	ok := twoFactorUtil.Verify(c.Ctx, passcode)
	if ok {
		if totpSessionData.EnableSession {
			c.SetSessionUsername(totpSessionData.UserId)
		}
		if !totpSessionData.AutoSignIn {
			c.setExpireForSession()
		}
		c.ResponseOk(http.StatusText(http.StatusOK))
	} else {
		c.ResponseError(http.StatusText(http.StatusUnauthorized))
	}
}

// TwoFactorDelete
// @Title TwoFactorDelete
// @Tag Two-Factor API
// @Description: Remove Totp
// @param	userId	form	string	true	"Id of user"
// @Success 200 {object}  Response object
// @router /two-factor/delete [post]
func (c *ApiController) TwoFactorDelete() {
	userId := c.Ctx.Request.Form.Get("userId")
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	object.SetUserField(user, "totp_secret", "")
	c.ResponseOk(http.StatusText(http.StatusOK))
}
