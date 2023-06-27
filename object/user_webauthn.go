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
	"fmt"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func GetWebAuthnObject(host string) (*webauthn.WebAuthn, error) {
	var err error

	_, originBackend := getOriginFromHost(host)

	localUrl, err := url.Parse(originBackend)
	if err != nil {
		return nil, fmt.Errorf("error when parsing origin:" + err.Error())
	}

	webAuthn, err := webauthn.New(&webauthn.Config{
		RPDisplayName: conf.GetConfigString("appname"),      // Display Name for your site
		RPID:          strings.Split(localUrl.Host, ":")[0], // Generally the domain name for your site, it's ok because splits cannot return empty array
		RPOrigin:      originBackend,                        // The origin URL for WebAuthn requests
		// RPIcon:     "https://duo.com/logo.png",           // Optional icon URL for your site
	})
	if err != nil {
		return nil, err
	}

	return webAuthn, nil
}

// WebAuthnID
// implementation of webauthn.User interface
func (user *User) WebAuthnID() []byte {
	return []byte(user.GetId())
}

func (user *User) WebAuthnName() string {
	return user.Name
}

func (user *User) WebAuthnDisplayName() string {
	return user.DisplayName
}

func (user *User) WebAuthnCredentials() []webauthn.Credential {
	return user.WebauthnCredentials
}

func (user *User) WebAuthnIcon() string {
	return user.Avatar
}

// CredentialExcludeList returns a CredentialDescriptor array filled with all the user's credentials
func (user *User) CredentialExcludeList() []protocol.CredentialDescriptor {
	credentials := user.WebAuthnCredentials()
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

func (user *User) AddCredentials(credential webauthn.Credential, isGlobalAdmin bool) (bool, error) {
	user.WebauthnCredentials = append(user.WebauthnCredentials, credential)
	return UpdateUser(user.GetId(), user, []string{"webauthnCredentials"}, isGlobalAdmin)
}

func (user *User) DeleteCredentials(credentialIdBase64 string) (bool, error) {
	for i, credential := range user.WebauthnCredentials {
		if base64.StdEncoding.EncodeToString(credential.ID) == credentialIdBase64 {
			user.WebauthnCredentials = append(user.WebauthnCredentials[0:i], user.WebauthnCredentials[i+1:]...)
			return UpdateUserForAllFields(user.GetId(), user)
		}
	}
	return false, nil
}
