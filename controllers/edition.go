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
	"fmt"

	"github.com/casdoor/casdoor/edition"
)

const editionRequiredCode = "edition-required"

// ServeJSON turns the errors of features that need a higher edition into a message in the user's language,
// which names the edition to upgrade to and links to the pricing page
func (c *ApiController) ServeJSON(encoding ...bool) error {
	switch resp := c.Data["json"].(type) {
	case *Response:
		c.localizeEditionError(resp)
	case Response:
		c.localizeEditionError(&resp)
		c.Data["json"] = resp
	}
	return c.Controller.ServeJSON(encoding...)
}

func (c *ApiController) localizeEditionError(resp *Response) {
	if resp == nil || resp.Status != "error" {
		return
	}
	required, start, end := edition.FindRequired(resp.Msg)
	if required == nil {
		return
	}

	text := c.T("general:%s needs a commercial or enterprise edition license, please upgrade (%s)")
	if required.Edition == edition.Enterprise {
		text = c.T("general:%s needs an enterprise edition license, please upgrade (%s)")
	}
	resp.Msg = resp.Msg[:start] + fmt.Sprintf(text, c.getEditionFeatureName(required.Feature), edition.PricingUrl) + resp.Msg[end:]
	if resp.Code == "" {
		resp.Code = editionRequiredCode
	}
}

// getEditionFeatureName translates the names that edition.Require is called with, written out one by one so
// that the i18n generator keeps them
func (c *ApiController) getEditionFeatureName(feature string) string {
	switch feature {
	case edition.FeatureExclusiveSignin:
		return c.T("edition:Exclusive sign-in")
	case edition.FeatureInactiveTimeout:
		return c.T("edition:Inactivity timeout")
	case edition.FeatureMobileOneClick:
		return c.T("edition:Mobile one-click login")
	case edition.FeatureIdJag:
		return c.T("edition:ID-JAG")
	case edition.FeatureMcpGateway:
		return c.T("edition:MCP gateway")
	case edition.FeatureKeycloakSyncer:
		return c.T("edition:Keycloak syncer")
	case edition.FeatureMultipleSyncers:
		return c.T("edition:More than one syncer")
	case edition.FeatureSecurityPolicy:
		return c.T("edition:Security policy")
	case edition.FeatureKeyRotation:
		return c.T("edition:Signing key rotation")
	case edition.FeaturePushMfa:
		return c.T("edition:Push notification MFA")
	case edition.FeaturePaidIntegrations:
		return c.T("edition:Paid integrations")
	case edition.FeatureTracing:
		return c.T("edition:OpenTelemetry tracing")
	case edition.FeatureRbac:
		return c.T("edition:Role-based access control (RBAC)")
	case edition.FeatureAbac:
		return c.T("edition:Attribute-based access control (ABAC)")
	default:
		return feature
	}
}
