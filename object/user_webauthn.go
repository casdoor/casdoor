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
	"slices"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func GetWebAuthnObject(host string) (*webauthn.WebAuthn, error) {
	_, originBackend := getOriginFromHost(host)

	rpId, rpOrigins, err := getWebAuthnRpIdAndOrigins(originBackend)
	if err != nil {
		return nil, err
	}

	webAuthn, err := webauthn.New(&webauthn.Config{
		RPDisplayName: conf.GetConfigString("appname"), // Display Name for your site
		RPID:          rpId,                            // Generally the domain name for your site
		RPOrigins:     rpOrigins,                       // The origin URLs for WebAuthn requests, without the base path
		// RPIcon:     "https://duo.com/logo.png",           // Optional icon URL for your site
	})
	if err != nil {
		return nil, err
	}

	return webAuthn, nil
}

// getWebAuthnRpIdAndOrigins lets webAuthnRpId and webAuthnRpOrigins widen the defaults taken from the backend origin
func getWebAuthnRpIdAndOrigins(originBackend string) (string, []string, error) {
	localUrl, err := url.Parse(originBackend)
	if err != nil {
		return "", nil, fmt.Errorf("error when parsing origin: %w", err)
	}

	rpId := strings.Split(localUrl.Host, ":")[0]
	rpOrigins := []string{localUrl.Scheme + "://" + localUrl.Host}

	rpIdConf := strings.ToLower(strings.TrimSpace(conf.GetConfigString("webAuthnRpId")))
	rpOriginsConf := strings.TrimSpace(conf.GetConfigString("webAuthnRpOrigins"))
	if rpIdConf == "" && rpOriginsConf == "" {
		return rpId, rpOrigins, nil
	}

	if rpIdConf != "" {
		rpId = rpIdConf
	}
	for _, origin := range strings.Split(rpOriginsConf, ",") {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin != "" && !slices.Contains(rpOrigins, origin) {
			rpOrigins = append(rpOrigins, origin)
		}
	}

	for _, origin := range rpOrigins {
		originUrl, err := url.Parse(origin)
		if err != nil || originUrl.Scheme == "" || originUrl.Host == "" {
			return "", nil, fmt.Errorf("invalid WebAuthn origin: %s", origin)
		}

		hostname := strings.ToLower(originUrl.Hostname())
		if hostname != strings.ToLower(rpId) && !strings.HasSuffix(hostname, "."+strings.ToLower(rpId)) {
			return "", nil, fmt.Errorf("the WebAuthn RP ID: %s is not a registrable suffix of the origin: %s", rpId, origin)
		}
	}

	return rpId, rpOrigins, nil
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
