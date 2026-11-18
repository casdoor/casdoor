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
	"errors"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/ee"
	"github.com/casdoor/casdoor/form"
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

// checkCertSignin signs in with the ticket of /api/login/cert, or a certificate and its signature of a challenge
func (c *ApiController) checkCertSignin(authForm *form.AuthForm) (*object.User, error) {
	application, err := c.GetCertLoginApplication(authForm.Application)
	if err != nil {
		return nil, err
	}
	lang := c.GetAcceptLanguage()

	var user *object.User
	if authForm.Code != "" {
		ticket, err := ee.ConsumeCertLoginTicket(ee.CertLoginKindTicket, authForm.Code, application, lang)
		if err != nil {
			return nil, err
		}
		user, err = object.GetUser(ticket.User)
		if err != nil {
			return nil, err
		}
		if user == nil || user.IsDeleted || user.IsForbidden {
			return nil, errors.New(c.T("check:The user is forbidden to sign in, please contact the administrator"))
		}
	} else {
		ders, err := ee.ReadCertificateDers(authForm.Certificate)
		if err != nil {
			return nil, fmt.Errorf(c.T("enterprise:The certificate is invalid: %s"), err.Error())
		}
		signature, err := ee.DecodeSignature(authForm.Signature)
		if err != nil {
			return nil, errors.New(c.T("enterprise:The signature of the challenge is invalid"))
		}
		// the challenge is used up by any attempt, a wrong signature included
		_, err = ee.ConsumeCertLoginTicket(ee.CertLoginKindChallenge, authForm.Nonce, application, lang)
		if err != nil {
			return nil, err
		}
		input := &ee.CertLoginInput{Ders: ders, Message: []byte(authForm.Nonce), Signature: signature}
		user, _, err = ee.VerifyCertLogin(ee.GetCertLoginOrganization(application, authForm.Organization), input, lang)
		if err != nil {
			return nil, err
		}
	}

	authForm.Organization = user.Owner
	authForm.Username = user.Name
	return user, nil
}

// checkEnterprisePolicyChange validates a changed enterprise policy and stops the admins whose duty doesn't cover it
func (c *ApiController) checkEnterprisePolicyChange(oldOrganization *object.Organization, organization *object.Organization) error {
	editor := c.GetCurrentUser()
	if oldOrganization == nil {
		if err := ee.ValidateDelegatedAdmins(organization.Name, organization.EnterprisePolicy, c.GetAcceptLanguage()); err != nil {
			return err
		}
		return ee.ValidateEnterprisePolicy(organization.EnterprisePolicy, nil, c.GetAcceptLanguage())
	}

	err := ee.CheckOrganizationDutyChange(oldOrganization, organization, editor)
	if err != nil {
		return err
	}
	if util.StructToJson(oldOrganization.EnterprisePolicy) == util.StructToJson(organization.EnterprisePolicy) {
		return nil
	}
	if err = ee.ValidateDelegatedAdmins(organization.Name, organization.EnterprisePolicy, c.GetAcceptLanguage()); err != nil {
		return err
	}
	// whoever separates the admins of their own organization has to stay a security admin
	if editor != nil && editor.Owner != organization.Name {
		editor = nil
	}
	return ee.ValidateEnterprisePolicy(organization.EnterprisePolicy, editor, c.GetAcceptLanguage())
}

// maskPersonalInfo masks the phone, email and ID card of other users for the console, under the organization's
// security policy. The applications calling the API with their credentials get the values.
func (c *ApiController) maskPersonalInfo(users ...*object.User) {
	requester := c.GetSessionUsername()
	if strings.HasPrefix(requester, "app/") {
		return
	}
	for _, user := range users {
		if user != nil && user.GetId() != requester {
			ee.MaskPersonalInfo(user)
		}
	}
}

func (c *ApiController) GetCertLoginApplication(name string) (*object.Application, error) {
	application, err := object.GetApplication(util.GetId("admin", name))
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, fmt.Errorf(c.T("auth:The application: %s does not exist"), name)
	}
	if !application.IsCertLoginEnabled() {
		return nil, errors.New(c.T("enterprise:The login method: login with certificate is not enabled for the application"))
	}
	if err = edition.Require(edition.Enterprise, edition.FeatureCertLogin); err != nil {
		return nil, err
	}
	return application, nil
}
