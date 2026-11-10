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
	"encoding/json"
	"errors"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/hsm"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// checkHsmCertChange lets only a global admin point a cert at a key in the HSM, which holds the keys of every
// organization; an organization admin may keep the HSM key a cert already has
func (c *ApiController) checkHsmCertChange(cert *object.Cert, oldId string) error {
	if !cert.IsHsmKey() || c.IsGlobalAdmin() {
		return nil
	}

	if oldId != "" {
		oldCert, err := object.GetCert(oldId)
		if err != nil {
			return err
		}
		if oldCert != nil && oldCert.IsHsmKey() && oldCert.KeyReference == cert.KeyReference {
			return nil
		}
	}
	return errors.New(c.T("cert:Only a global admin can choose a key in the HSM"))
}

// GetHsmStatus
// @Title GetHsmStatus
// @Tag Cert API
// @Description get the hardware security module set in app.conf, connecting to it when connect is true
// @Param   connect     query    string  false        "Whether to connect to the HSM"
// @Success 200 {object} hsm.Status The Response object
// @router /get-hsm-status [get]
func (c *ApiController) GetHsmStatus() {
	if !c.requireGlobalAdmin() {
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
	if !c.requireGlobalAdmin() {
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

	certificate, err := object.GenerateHsmCertificate(&cert)
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

	result, err := object.TestHsmCert(cert)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(result)
}
