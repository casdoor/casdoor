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

package object

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/gssapi"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/jcmturner/gokrb5/v8/service"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

// ctxCredentials is the SPNEGO context key holding the Kerberos credentials.
// This must match the value used internally by gokrb5's spnego package.
// If the gokrb5 library changes this internal constant in a future version,
// this value will need to be updated accordingly.
const ctxCredentials = "github.com/jcmturner/gokrb5/v8/ctxCredentials"

// ValidateKerberosToken validates a base64-encoded SPNEGO token from the
// Authorization header and returns the authenticated Kerberos username.
func ValidateKerberosToken(organization *Organization, spnegoTokenBase64 string) (string, error) {
	if organization.KerberosRealm == "" || organization.KerberosKdcHost == "" || organization.KerberosKeytab == "" {
		return "", fmt.Errorf("kerberos configuration is incomplete for organization: %s", organization.Name)
	}

	keytabData, err := base64.StdEncoding.DecodeString(organization.KerberosKeytab)
	if err != nil {
		return "", fmt.Errorf("failed to decode keytab: %w", err)
	}

	kt := keytab.New()
	err = kt.Unmarshal(keytabData)
	if err != nil {
		return "", fmt.Errorf("failed to parse keytab: %w", err)
	}

	servicePrincipal := organization.KerberosServiceName
	if servicePrincipal == "" {
		servicePrincipal = "HTTP"
	}

	spnegoSvc := spnego.SPNEGOService(kt, service.KeytabPrincipal(servicePrincipal))

	tokenBytes, err := base64.StdEncoding.DecodeString(spnegoTokenBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode SPNEGO token: %w", err)
	}

	var st spnego.SPNEGOToken
	err = st.Unmarshal(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal SPNEGO token: %w", err)
	}

	authed, ctx, status := spnegoSvc.AcceptSecContext(&st)
	if status.Code != gssapi.StatusComplete && status.Code != gssapi.StatusContinueNeeded {
		return "", fmt.Errorf("SPNEGO validation error: %s", status.Message)
	}
	if status.Code == gssapi.StatusContinueNeeded {
		return "", fmt.Errorf("SPNEGO negotiation requires continuation, which is not supported")
	}
	if !authed {
		return "", fmt.Errorf("SPNEGO token validation failed")
	}

	creds, ok := ctx.Value(ctxCredentials).(*credentials.Credentials)
	if !ok || creds == nil {
		return "", fmt.Errorf("no credentials found in SPNEGO context")
	}

	username := creds.UserName()
	if username == "" {
		return "", fmt.Errorf("no username found in Kerberos ticket")
	}

	return username, nil
}

// GetUserByKerberosName looks up a Casdoor user by their Kerberos principal name.
// It strips the realm part (e.g., "user@REALM.COM" -> "user") and searches by username.
func GetUserByKerberosName(organizationName string, kerberosUsername string) (*User, error) {
	username := kerberosUsername
	if idx := strings.Index(username, "@"); idx >= 0 {
		username = username[:idx]
	}

	user, err := GetUserByFields(organizationName, username)
	if err != nil {
		return nil, err
	}

	return user, nil
}
