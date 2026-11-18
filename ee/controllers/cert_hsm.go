// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package controllers

import (
	"encoding/json"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/ee"
	"github.com/casdoor/casdoor/ee/hsm"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetHsmStatus
// @Title GetHsmStatus
// @Tag Cert API
// @Description get the hardware security module set in app.conf, connecting to it when connect is true
// @Param   connect     query    string  false        "Whether to connect to the HSM"
// @Success 200 {object} hsm.Status The Response object
// @router /get-hsm-status [get]
func (c *ApiController) GetHsmStatus() {
	if !c.RequireGlobalAdmin() {
		return
	}

	status := hsm.GetStatus(false)
	if c.Ctx.Input.Query("connect") == "true" && status.Configured {
		if err := edition.Require(edition.Enterprise, edition.FeatureHsm); err != nil {
			c.ResponseError(err.Error())
			return
		}
		status = hsm.GetStatus(true)
	}
	c.ResponseOk(status)
}

// GenerateHsmCert
// @Title GenerateHsmCert
// @Tag Cert API
// @Description create the self-signed certificate of a cert whose private key is in the HSM, signed in the HSM
// @Param   body    body   object.Cert  true        "The cert with the key reference and the crypto algorithm"
// @Success 200 {object} controllers.Response The Response object
// @router /generate-hsm-cert [post]
func (c *ApiController) GenerateHsmCert() {
	if !c.RequireGlobalAdmin() {
		return
	}

	var cert object.Cert
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cert)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if err = object.CheckCertEdition(&cert); err != nil {
		c.ResponseError(err.Error())
		return
	}

	certificate, err := ee.GenerateHsmCertificate(&cert)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	expireTime, _ := util.GetCertExpireTime(certificate)
	c.ResponseOk(certificate, expireTime)
}

// TestHsmCert
// @Title TestHsmCert
// @Tag Cert API
// @Description sign a test token with the HSM key of a saved cert and verify it with the cert's certificate
// @Param   id     query    string  true        "The id ( owner/name ) of the cert"
// @Success 200 {object} object.HsmTestResult The Response object
// @router /test-hsm-cert [post]
func (c *ApiController) TestHsmCert() {
	id := c.Ctx.Input.Query("id")
	cert, err := object.GetCert(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if cert == nil {
		c.ResponseError(c.T("general:The object does not exist"))
		return
	}

	result, err := ee.TestHsmCert(cert)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}
