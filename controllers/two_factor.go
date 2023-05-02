package controllers

import (
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

	if len(userId) == 0 {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	twoFactorUtil := object.GetTwoFactorUtil(authType, nil)
	if twoFactorUtil == nil {
		c.ResponseError("Invalid auth type")
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	issuer := beego.AppConfig.String("appname")
	accountName := user.GetId()

	twoFactorProps, err := twoFactorUtil.Initiate(c.Ctx, issuer, accountName)
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

	if authType == "" || passcode == "" {
		c.ResponseError("missing auth type or passcode")
		return
	}
	twoFactorUtil := object.GetTwoFactorUtil(authType, nil)

	err := twoFactorUtil.SetupVerify(c.Ctx, passcode)
	if err != nil {
		c.ResponseError(err.Error())
	} else {
		c.ResponseOk(http.StatusText(http.StatusOK))
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

	twoFactor := object.GetTwoFactorUtil(authType, nil)
	err := twoFactor.Enable(c.Ctx, user)

	if err != nil {
		c.ResponseError(err.Error())
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

	user := object.GetUser(totpSessionData.UserId)
	if user == nil {
		c.ResponseError("User does not exist")
		return
	}

	twoFactorUtil := object.GetTwoFactorUtil(authType, user.GetPreferTwoFactor())
	err := twoFactorUtil.Verify(passcode)
	if err != nil {
		c.ResponseError(http.StatusText(http.StatusUnauthorized))
	} else {
		if totpSessionData.EnableSession {
			c.SetSessionUsername(totpSessionData.UserId)
		}
		if !totpSessionData.AutoSignIn {
			c.setExpireForSession()
		}
		c.ResponseOk(http.StatusText(http.StatusOK))
	}
}

// TwoFactorAuthRecover
// @Title TwoFactorAuthRecover
// @Tag Totp API
// @Description recover two-factor authentication
// @param	recoveryCode	form	string	true	"recovery code"
// @Success 200 {object}  Response object
// @router /two-factor/auth/recover [post]
func (c *ApiController) TwoFactorAuthRecover() {
	authType := c.Ctx.Request.Form.Get("type")
	recoveryCode := c.Ctx.Request.Form.Get("recoveryCode")

	tfaSessionData := c.getTfaSessionData()
	if tfaSessionData == nil {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	user := object.GetUser(tfaSessionData.UserId)
	if user == nil {
		c.ResponseError("User does not exist")
		return
	}

	ok, err := object.RecoverTfs(user, recoveryCode, authType)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if ok {
		if tfaSessionData.EnableSession {
			c.SetSessionUsername(tfaSessionData.UserId)
		}
		if !tfaSessionData.AutoSignIn {
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
// @router /two-factor/ [delete]
func (c *ApiController) TwoFactorDelete() {
	authType := c.Ctx.Request.Form.Get("type")
	userId := c.Ctx.Request.Form.Get("userId")
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	twoFactorProps := user.TwoFactorAuth[:0]
	i := 0
	for _, twoFactorProp := range twoFactorProps {
		if twoFactorProp.AuthType != authType {
			twoFactorProps[i] = twoFactorProp
			i++
		}
	}
	user.TwoFactorAuth = twoFactorProps
	object.UpdateUser(userId, user, []string{"two_factor_auth"}, user.IsAdminUser())
	c.ResponseOk(user.TwoFactorAuth)
}
