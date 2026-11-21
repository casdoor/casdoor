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
	"fmt"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	dsig "github.com/russellhaering/goxmldsig"
)

// The commercial and enterprise features in ee/ set these hooks when they are compiled in, the core works without them

var editionTables []interface{}

// RegisterTables adds tables that are created and synced with the core ones
func RegisterTables(tables ...interface{}) {
	editionTables = append(editionTables, tables...)
}

var (
	EncryptUserFieldsHook                 func(user *User)
	DecryptUserFieldsHook                 func(user *User)
	KeepMaskedAccountSecretsHook          func(user *User, oldUser *User)
	PopulateHsmContentHook                func(cert *Cert) error
	GetHsmSecretHook                      func(cert *Cert, purpose string) (string, error)
	GetJwtSigningKeyHook                  func(method jwt.SigningMethod, cert *Cert) (interface{}, error)
	GetJwtVerifyKeyHook                   func(token *jwt.Token, certificate string) (interface{}, error)
	SignJwtTokenHook                      func(token *jwt.Token, key interface{}) (string, error)
	NewCertXmlSigningContextHook          func(cert *Cert, certificate string) (*dsig.SigningContext, error)
	GetPreviousJsonWebKeyHook             func(cert *Cert) *jose.JSONWebKey
	GetCachedSecurityPolicyHook           func(owner string) *SecurityPolicy
	GetEnterprisePolicyHook               func(organization *Organization) *EnterprisePolicy
	CheckPasswordSecurityPolicyHook       func(organization *Organization, password string, username string, lang string) string
	CheckSmsSecurityPolicyHook            func(organization *Organization, phone string, lang string) error
	CheckSmsVolumeHook                    func(organization *Organization)
	AddChainedRecordHook                  func(record *Record) (int64, error)
	ConsumePushChallengeHook              func(id string, provider string, receiver string) error
	CheckProviderEditionHook              func(provider *Provider) error
	MergePaidIntegrationsHook             func(index *MarketplaceIndex)
	NewPluginSyncerProviderHook           func(syncer *Syncer) SyncerProvider
	GetUserByWechatUnionIdHook            func(owner string, unionId string) (*User, error)
	CheckProvisioningConfigHook           func(application *Application) error
	ClearProvisioningApplicationCacheHook func()
	IsProvisionedUserUpdateHook           func(columns []string) bool
	ProvisionUserHook                     func(user *User)
	RequestOrganizationProvisioningHook   func(organization string)
	IsTrustedOrganizationHook             func(name string) bool
)

var errEnterpriseFeatures = fmt.Errorf("this needs the enterprise features of Casdoor")

// The xorm processors keep the encrypted personal fields encrypted in the database and in plaintext everywhere else

func (user *User) BeforeInsert() {
	if EncryptUserFieldsHook != nil {
		EncryptUserFieldsHook(user)
	}
}

func (user *User) AfterInsert() {
	if DecryptUserFieldsHook != nil {
		DecryptUserFieldsHook(user)
	}
}

func (user *User) BeforeUpdate() {
	if EncryptUserFieldsHook != nil {
		EncryptUserFieldsHook(user)
	}
}

func (user *User) AfterUpdate() {
	if DecryptUserFieldsHook != nil {
		DecryptUserFieldsHook(user)
	}
}

func (user *User) AfterLoad() {
	if DecryptUserFieldsHook != nil {
		DecryptUserFieldsHook(user)
	}
}

func keepMaskedAccountSecrets(user *User, oldUser *User) {
	if KeepMaskedAccountSecretsHook != nil {
		KeepMaskedAccountSecretsHook(user, oldUser)
	}
}

func getHsmSecret(cert *Cert, purpose string) (string, error) {
	if GetHsmSecretHook == nil {
		return "", errEnterpriseFeatures
	}
	return GetHsmSecretHook(cert, purpose)
}

func getJwtSigningKey(method jwt.SigningMethod, cert *Cert) (interface{}, error) {
	if GetJwtSigningKeyHook != nil {
		return GetJwtSigningKeyHook(method, cert)
	}
	return ParseJwtSigningKey(method, cert)
}

func getJwtVerifyKey(token *jwt.Token, certificate string) (interface{}, error) {
	if GetJwtVerifyKeyHook != nil {
		return GetJwtVerifyKeyHook(token, certificate)
	}
	return ParseJwtVerifyKey(token, certificate)
}

func signJwtToken(token *jwt.Token, key interface{}) (string, error) {
	if SignJwtTokenHook != nil {
		return SignJwtTokenHook(token, key)
	}
	return token.SignedString(key)
}

func newCertXmlSigningContext(cert *Cert, certificate string) (*dsig.SigningContext, error) {
	if NewCertXmlSigningContextHook != nil {
		return NewCertXmlSigningContextHook(cert, certificate)
	}
	return dsig.NewDefaultSigningContext(&X509Key{PrivateKey: cert.PrivateKey, X509Certificate: certificate}), nil
}

func getPreviousJsonWebKey(cert *Cert) *jose.JSONWebKey {
	if GetPreviousJsonWebKeyHook == nil {
		return nil
	}
	return GetPreviousJsonWebKeyHook(cert)
}

// GetCachedSecurityPolicy is the security policy of an organization, nil without the commercial features
func GetCachedSecurityPolicy(owner string) *SecurityPolicy {
	if GetCachedSecurityPolicyHook == nil {
		return nil
	}
	return GetCachedSecurityPolicyHook(owner)
}

// GetEnterprisePolicy is the enterprise policy of an organization, nil without the enterprise features
func GetEnterprisePolicy(organization *Organization) *EnterprisePolicy {
	if GetEnterprisePolicyHook == nil {
		return nil
	}
	return GetEnterprisePolicyHook(organization)
}

func checkPasswordSecurityPolicy(organization *Organization, password string, username string, lang string) string {
	if CheckPasswordSecurityPolicyHook == nil {
		return ""
	}
	return CheckPasswordSecurityPolicyHook(organization, password, username, lang)
}

func checkSmsSecurityPolicy(organization *Organization, phone string, lang string) error {
	if CheckSmsSecurityPolicyHook == nil {
		return nil
	}
	return CheckSmsSecurityPolicyHook(organization, phone, lang)
}

func checkSmsVolume(organization *Organization) {
	if CheckSmsVolumeHook != nil {
		CheckSmsVolumeHook(organization)
	}
}

func addChainedRecord(record *Record) (int64, error) {
	if AddChainedRecordHook == nil {
		return addRecord(record)
	}
	return AddChainedRecordHook(record)
}

func consumePushChallenge(id string, provider string, receiver string) error {
	if ConsumePushChallengeHook == nil {
		return errEnterpriseFeatures
	}
	return ConsumePushChallengeHook(id, provider, receiver)
}

func CheckProviderEdition(provider *Provider) error {
	if CheckProviderEditionHook == nil {
		return nil
	}
	return CheckProviderEditionHook(provider)
}

func mergePaidIntegrations(index *MarketplaceIndex) {
	if MergePaidIntegrationsHook != nil {
		MergePaidIntegrationsHook(index)
	}
}

func GetUserByWechatUnionId(owner string, unionId string) (*User, error) {
	if GetUserByWechatUnionIdHook == nil {
		return nil, nil
	}
	return GetUserByWechatUnionIdHook(owner, unionId)
}

func CheckProvisioningConfig(application *Application) error {
	if CheckProvisioningConfigHook == nil {
		return nil
	}
	return CheckProvisioningConfigHook(application)
}

func ClearProvisioningApplicationCache() {
	if ClearProvisioningApplicationCacheHook != nil {
		ClearProvisioningApplicationCacheHook()
	}
}

func isProvisionedUserUpdate(columns []string) bool {
	return IsProvisionedUserUpdateHook != nil && IsProvisionedUserUpdateHook(columns)
}

func provisionUser(user *User) {
	if ProvisionUserHook != nil {
		ProvisionUserHook(user)
	}
}

func requestOrganizationProvisioning(organization string) {
	if RequestOrganizationProvisioningHook != nil {
		RequestOrganizationProvisioningHook(organization)
	}
}

func isTrustedOrganization(name string) bool {
	return IsTrustedOrganizationHook != nil && IsTrustedOrganizationHook(name)
}
