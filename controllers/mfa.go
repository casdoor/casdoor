// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"net/http"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/google/uuid"
)

const (
	MfaRecoveryCodesSession = "mfa_recovery_codes"
	MfaCountryCodeSession   = "mfa_country_code"
	MfaDestSession          = "mfa_dest"
	MfaTotpSecretSession    = "mfa_totp_secret"
)

// MfaSetupInitiate
// @Title MfaSetupInitiate
// @Tag MFA API
// @Description setup MFA
// @param owner	form	string	true	"owner of user"
// @param name	form	string	true	"name of user"
// @param type	form	string	true	"MFA auth type"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/setup/initiate [post]
func (c *ApiController) MfaSetupInitiate() {
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	mfaType := c.Ctx.Request.Form.Get("mfaType")
	stateLess := c.Ctx.Request.Form.Get("stateLess")
	userId := util.GetId(owner, name)

	if len(userId) == 0 {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	MfaUtil := object.GetMfaUtil(mfaType, nil)
	if MfaUtil == nil {
		c.ResponseError("Invalid auth type")
	}

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	mfaProps, err := MfaUtil.Initiate(user.GetId())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	recoveryCode := uuid.NewString()
	mfaCacheKey := ""
	if stateLess == "true" {
		mfaCacheKey = c.Ctx.Input.CruSession.SessionID() + util.GenerateSimpleTimeId()
		mfaCacheVal := make(map[string]string)
		mfaCacheVal[MfaRecoveryCodesSession] = recoveryCode
		if mfaType == object.TotpType {
			mfaCacheVal[MfaTotpSecretSession] = mfaProps.Secret
		}
		mfaCacheVal["verifiedUserId"] = userId
		object.MfaCache.Store(mfaCacheKey, mfaCacheVal)
	} else {
		c.SetSession(MfaRecoveryCodesSession, recoveryCode)
		if mfaType == object.TotpType {
			c.SetSession(MfaTotpSecretSession, mfaProps.Secret)
		}
	}

	mfaProps.RecoveryCodes = []string{recoveryCode}

	resp := mfaProps
	c.ResponseOk(resp, mfaCacheKey)
}

// MfaSetupVerify
// @Title MfaSetupVerify
// @Tag MFA API
// @Description setup verify totp
// @param	secret		form	string	true	"MFA secret"
// @param	passcode	form 	string 	true	"MFA passcode"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/setup/verify [post]
func (c *ApiController) MfaSetupVerify() {
	mfaType := c.Ctx.Request.Form.Get("mfaType")
	passcode := c.Ctx.Request.Form.Get("passcode")
	mfaCacheKey := c.Ctx.Request.Form.Get("mfaCacheKey")

	if mfaType == "" || passcode == "" {
		c.ResponseError("missing auth type or passcode")
		return
	}

	config := &object.MfaProps{
		MfaType: mfaType,
	}
	if mfaType == object.TotpType {
		secret := object.GetPropsFromContext(MfaTotpSecretSession, c.Ctx.Input.CruSession, mfaCacheKey)
		if secret == "" {
			c.ResponseError("totp secret is missing")
			return
		}
		config.Secret = secret
	} else if mfaType == object.SmsType {
		dest := object.GetPropsFromContext(MfaDestSession, c.Ctx.Input.CruSession, mfaCacheKey)
		if dest == "" {
			c.ResponseError("destination is missing")
			return
		}
		config.Secret = dest
		countryCode := object.GetPropsFromContext(MfaCountryCodeSession, c.Ctx.Input.CruSession, mfaCacheKey)
		if countryCode == "" {
			c.ResponseError("country code is missing")
			return
		}
		config.CountryCode = countryCode
	} else if mfaType == object.EmailType {
		dest := object.GetPropsFromContext(MfaDestSession, c.Ctx.Input.CruSession, mfaCacheKey)
		if dest == "" {
			c.ResponseError("destination is missing")
			return
		}
		config.Secret = dest
	}

	mfaUtil := object.GetMfaUtil(mfaType, config)
	if mfaUtil == nil {
		c.ResponseError("Invalid multi-factor authentication type")
		return
	}

	err := mfaUtil.SetupVerify(passcode)
	if err != nil {
		c.ResponseError(err.Error())
	} else {
		c.ResponseOk(http.StatusText(http.StatusOK))
	}
}

// MfaSetupEnable
// @Title MfaSetupEnable
// @Tag MFA API
// @Description enable totp
// @param owner	form	string	true	"owner of user"
// @param name	form	string	true	"name of user"
// @param type	form	string	true	"MFA auth type"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/setup/enable [post]
func (c *ApiController) MfaSetupEnable() {
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	mfaType := c.Ctx.Request.Form.Get("mfaType")
	mfaCacheKey := c.Ctx.Request.Form.Get("mfaCacheKey")

	user, err := object.GetUser(util.GetId(owner, name))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	config := &object.MfaProps{
		MfaType: mfaType,
	}

	if mfaType == object.TotpType {
		secret := object.GetPropsFromContext(MfaTotpSecretSession, c.Ctx.Input.CruSession, mfaCacheKey)
		if secret == "" {
			c.ResponseError("totp secret is missing")
			return
		}
		config.Secret = secret
	} else if mfaType == object.EmailType {
		if user.Email == "" {
			dest := object.GetPropsFromContext(MfaDestSession, c.Ctx.Input.CruSession, mfaCacheKey)
			if dest == "" {
				c.ResponseError("destination is missing")
				return
			}
			user.Email = dest
		}
	} else if mfaType == object.SmsType {
		if user.Phone == "" {
			dest := object.GetPropsFromContext(MfaDestSession, c.Ctx.Input.CruSession, mfaCacheKey)
			if dest == "" {
				c.ResponseError("destination is missing")
				return
			}
			user.Phone = dest
			countryCode := object.GetPropsFromContext(MfaCountryCodeSession, c.Ctx.Input.CruSession, mfaCacheKey)
			if countryCode == "" {
				c.ResponseError("country code is missing")
				return
			}
			user.CountryCode = countryCode
		}
	}
	recoveryCodes := object.GetPropsFromContext(MfaRecoveryCodesSession, c.Ctx.Input.CruSession, mfaCacheKey)
	if recoveryCodes == "" {
		c.ResponseError("recovery codes is missing")
		return
	}
	config.RecoveryCodes = []string{recoveryCodes}

	mfaUtil := object.GetMfaUtil(mfaType, config)
	if mfaUtil == nil {
		c.ResponseError("Invalid multi-factor authentication type")
		return
	}

	err = mfaUtil.Enable(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if mfaCacheKey == "" {
		c.DelSession(MfaRecoveryCodesSession)
		if mfaType == object.TotpType {
			c.DelSession(MfaTotpSecretSession)
		} else {
			c.DelSession(MfaCountryCodeSession)
			c.DelSession(MfaDestSession)
		}
	} else {
		object.MfaCache.Delete(mfaCacheKey)
	}

	c.ResponseOk(http.StatusText(http.StatusOK))
}

// DeleteMfa
// @Title DeleteMfa
// @Tag MFA API
// @Description: Delete MFA
// @param owner	form	string	true	"owner of user"
// @param name	form	string	true	"name of user"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-mfa/ [post]
func (c *ApiController) DeleteMfa() {
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	userId := util.GetId(owner, name)

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	err = object.DisabledMultiFactorAuth(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(object.GetAllMfaProps(user, true))
}

// SetPreferredMfa
// @Title SetPreferredMfa
// @Tag MFA API
// @Description: Set specific Mfa Preferred
// @param owner	form	string	true	"owner of user"
// @param name	form	string	true	"name of user"
// @param id	form	string	true	"id of user's MFA props"
// @Success 200 {object} controllers.Response The Response object
// @router /set-preferred-mfa [post]
func (c *ApiController) SetPreferredMfa() {
	mfaType := c.Ctx.Request.Form.Get("mfaType")
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	userId := util.GetId(owner, name)

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	err = object.SetPreferredMultiFactorAuth(user, mfaType)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(object.GetAllMfaProps(user, true))
}
