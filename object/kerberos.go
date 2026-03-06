// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

// ctxKeyKerberosCreds is the context key used by gokrb5 to store credentials.
// This matches the unexported constant in the spnego package.
const ctxKeyKerberosCreds = "github.com/jcmturner/gokrb5/v8/ctxCredentials"

// CheckKerberosToken validates a SPNEGO/Kerberos token and returns the authenticated username.
// The token is the base64-encoded value from the "Authorization: Negotiate <token>" header.
// Returns the username (without realm) on success, or an error on failure.
func CheckKerberosToken(org *Organization, tokenBase64 string) (string, error) {
	if org.KerberosKeytab == "" {
		return "", fmt.Errorf("Kerberos keytab is not configured for organization: %s", org.Name)
	}

	// Decode the base64 keytab
	keytabBytes, err := base64.StdEncoding.DecodeString(org.KerberosKeytab)
	if err != nil {
		return "", fmt.Errorf("failed to decode Kerberos keytab: %v", err)
	}

	// Load the keytab
	kt := keytab.New()
	if err = kt.Unmarshal(keytabBytes); err != nil {
		return "", fmt.Errorf("failed to parse Kerberos keytab: %v", err)
	}

	// Decode the SPNEGO token from base64
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode Kerberos token: %v", err)
	}

	// Create SPNEGO service using the keytab
	settings := []func(*service.Settings){}
	if org.KerberosServiceName != "" {
		settings = append(settings, service.SName(org.KerberosServiceName))
	}
	s := spnego.SPNEGOService(kt, settings...)

	// Unmarshal the SPNEGO token
	var st spnego.SPNEGOToken
	if err = st.Unmarshal(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to parse SPNEGO token: %v", err)
	}

	// Validate the token
	authed, ctx, status := s.AcceptSecContext(&st)
	if status.Code != gssapi.StatusComplete && status.Code != gssapi.StatusContinueNeeded {
		return "", fmt.Errorf("Kerberos authentication failed: %s", status.Message)
	}
	if !authed {
		return "", fmt.Errorf("Kerberos authentication rejected")
	}

	// Extract credentials from context using the gokrb5 internal key
	creds, ok := ctx.Value(ctxKeyKerberosCreds).(*credentials.Credentials)
	if !ok || creds == nil {
		return "", fmt.Errorf("failed to extract credentials from Kerberos context")
	}

	username := creds.UserName()
	// Strip realm suffix if present (e.g., "user@REALM.COM" → "user")
	if idx := strings.Index(username, "@"); idx != -1 {
		username = username[:idx]
	}
	return username, nil
}
