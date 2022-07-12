// Copyright 2022 The casbin Authors. All Rights Reserved.
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
	"net/url"
	"strings"

	"github.com/astaxie/beego"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
)

func GetWebAuthnObject(host string) *webauthn.WebAuthn {
	var err error

	origin := beego.AppConfig.String("origin")
	if origin == "" {
		_, origin = getOriginFromHost(host)
	}

	localUrl, err := url.Parse(origin)
	if err != nil {
		panic("error when parsing origin:" + err.Error())
	}

	webAuthn, err := webauthn.New(&webauthn.Config{
		RPDisplayName: beego.AppConfig.String("appname"),    // Display Name for your site
		RPID:          strings.Split(localUrl.Host, ":")[0], // Generally the domain name for your site, it's ok because splits cannot return empty array
		RPOrigin:      origin,                               // The origin URL for WebAuthn requests
		// RPIcon:     "https://duo.com/logo.png",           // Optional icon URL for your site
	})
	if err != nil {
		panic(err)
	}

	return webAuthn
}

// implementation of webauthn.User interface
func (u *User) WebAuthnID() []byte {
	return []byte(u.GetId())
}

func (u *User) WebAuthnName() string {
	return u.Name
}

func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.WebauthnCredentials
}

func (u *User) WebAuthnIcon() string {
	return u.Avatar
}

// CredentialExcludeList returns a CredentialDescriptor array filled with all the user's credentials
func (u *User) CredentialExcludeList() []protocol.CredentialDescriptor {
	credentials := u.WebAuthnCredentials()
	credentialExcludeList := []protocol.CredentialDescriptor{}
	for _, cred := range credentials {
		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.ID,
		}
		credentialExcludeList = append(credentialExcludeList, descriptor)
	}

	return credentialExcludeList
}

func (u *User) AddCredentials(credential webauthn.Credential, isGlobalAdmin bool) bool {
	u.WebauthnCredentials = append(u.WebauthnCredentials, credential)
	return UpdateUser(u.GetId(), u, []string{"webauthnCredentials"}, isGlobalAdmin)
}

func (u *User) DeleteCredentials(credentialIdBase64 string) bool {
	for i, credential := range u.WebauthnCredentials {
		if base64.StdEncoding.EncodeToString(credential.ID) == credentialIdBase64 {
			u.WebauthnCredentials = append(u.WebauthnCredentials[0:i], u.WebauthnCredentials[i+1:]...)
			return UpdateUserForAllFields(u.GetId(), u)
		}
	}
	return false
}
