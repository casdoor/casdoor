// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Provider struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName       string `xorm:"varchar(100)" json:"displayName"`
	Category          string `xorm:"varchar(100)" json:"category"`
	Type              string `xorm:"varchar(100)" json:"type"`
	SubType           string `xorm:"varchar(100)" json:"subType"`
	Method            string `xorm:"varchar(100)" json:"method"`
	ClientId          string `xorm:"varchar(100)" json:"clientId"`
	ClientSecret      string `xorm:"varchar(2000)" json:"clientSecret"`
	ClientId2         string `xorm:"varchar(100)" json:"clientId2"`
	ClientSecret2     string `xorm:"varchar(100)" json:"clientSecret2"`
	Cert              string `xorm:"varchar(100)" json:"cert"`
	CustomAuthUrl     string `xorm:"varchar(200)" json:"customAuthUrl"`
	CustomScope       string `xorm:"varchar(200)" json:"customScope"`
	CustomTokenUrl    string `xorm:"varchar(200)" json:"customTokenUrl"`
	CustomUserInfoUrl string `xorm:"varchar(200)" json:"customUserInfoUrl"`
	CustomLogo        string `xorm:"varchar(200)" json:"customLogo"`

	Host       string `xorm:"varchar(100)" json:"host"`
	Port       int    `json:"port"`
	DisableSsl bool   `json:"disableSsl"`
	Title      string `xorm:"varchar(100)" json:"title"`
	Content    string `xorm:"varchar(1000)" json:"content"`
	Receiver   string `xorm:"varchar(100)" json:"receiver"`

	RegionId     string `xorm:"varchar(100)" json:"regionId"`
	SignName     string `xorm:"varchar(100)" json:"signName"`
	TemplateCode string `xorm:"varchar(100)" json:"templateCode"`
	AppId        string `xorm:"varchar(100)" json:"appId"`

	Endpoint         string `xorm:"varchar(1000)" json:"endpoint"`
	IntranetEndpoint string `xorm:"varchar(100)" json:"intranetEndpoint"`
	Domain           string `xorm:"varchar(100)" json:"domain"`
	Bucket           string `xorm:"varchar(100)" json:"bucket"`

	Metadata               string `xorm:"mediumtext" json:"metadata"`
	IdP                    string `xorm:"mediumtext" json:"idP"`
	IssuerUrl              string `xorm:"varchar(100)" json:"issuerUrl"`
	EnableSignAuthnRequest bool   `json:"enableSignAuthnRequest"`

	ProviderUrl string `xorm:"varchar(200)" json:"providerUrl"`
}

func GetMaskedProvider(provider *Provider) *Provider {
	if provider == nil {
		return nil
	}

	if provider.ClientSecret != "" {
		provider.ClientSecret = "***"
	}
	if provider.ClientSecret2 != "" {
		provider.ClientSecret2 = "***"
	}

	return provider
}

func GetMaskedProviders(providers []*Provider) []*Provider {
	for _, provider := range providers {
		provider = GetMaskedProvider(provider)
	}
	return providers
}

func GetProviderCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Provider{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetProviders(owner string) []*Provider {
	providers := []*Provider{}
	err := adapter.Engine.Desc("created_time").Find(&providers, &Provider{Owner: owner})
	if err != nil {
		panic(err)
	}

	return providers
}

func GetPaginationProviders(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Provider {
	providers := []*Provider{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&providers)
	if err != nil {
		panic(err)
	}

	return providers
}

func getProvider(owner string, name string) *Provider {
	if owner == "" || name == "" {
		return nil
	}

	provider := Provider{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&provider)
	if err != nil {
		panic(err)
	}

	if existed {
		return &provider
	} else {
		return nil
	}
}

func GetProvider(id string) *Provider {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getProvider(owner, name)
}

func GetDefaultCaptchaProvider() *Provider {
	provider := Provider{Owner: "admin", Category: "Captcha"}
	existed, err := adapter.Engine.Get(&provider)
	if err != nil {
		panic(err)
	}

	if !existed {
		return nil
	}

	return &provider
}

func GetWechatMiniProgramProvider(application *Application) *Provider {
	providers := application.Providers
	for _, provider := range providers {
		if provider.Provider.Type == "WeChatMiniProgram" {
			return provider.Provider
		}
	}
	return nil
}

func UpdateProvider(id string, provider *Provider) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getProvider(owner, name) == nil {
		return false
	}

	session := adapter.Engine.ID(core.PK{owner, name}).AllCols()
	if provider.ClientSecret == "***" {
		session = session.Omit("client_secret")
	}
	if provider.ClientSecret2 == "***" {
		session = session.Omit("client_secret2")
	}
	affected, err := session.Update(provider)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddProvider(provider *Provider) bool {
	affected, err := adapter.Engine.Insert(provider)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteProvider(provider *Provider) bool {
	affected, err := adapter.Engine.ID(core.PK{provider.Owner, provider.Name}).Delete(&Provider{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (p *Provider) getPaymentProvider() (pp.PaymentProvider, *Cert, error) {
	cert := &Cert{}
	if p.Cert != "" {
		cert = getCert(p.Owner, p.Cert)
		if cert == nil {
			return nil, nil, fmt.Errorf("the cert: %s does not exist", p.Cert)
		}
	}

	pProvider := pp.GetPaymentProvider(p.Type, p.ClientId, p.ClientSecret, p.Host, cert.Certificate, cert.PrivateKey, cert.AuthorityPublicKey, cert.AuthorityRootPublicKey)
	if pProvider == nil {
		return nil, cert, fmt.Errorf("the payment provider type: %s is not supported", p.Type)
	}

	return pProvider, cert, nil
}

func (p *Provider) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}

func GetCaptchaProviderByOwnerName(applicationId, lang string) (*Provider, error) {
	owner, name := util.GetOwnerAndNameFromId(applicationId)
	provider := Provider{Owner: owner, Name: name, Category: "Captcha"}
	existed, err := adapter.Engine.Get(&provider)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, fmt.Errorf(i18n.Translate(lang, "ProviderErr.DoNotExist"), applicationId)
	}

	return &provider, nil
}

func GetCaptchaProviderByApplication(applicationId, isCurrentProvider, lang string) (*Provider, error) {
	if isCurrentProvider == "true" {
		return GetCaptchaProviderByOwnerName(applicationId, lang)
	}
	application := GetApplication(applicationId)
	if application == nil || len(application.Providers) == 0 {
		return nil, fmt.Errorf(i18n.Translate(lang, "ApplicationErr.InvalidID"))
	}
	for _, provider := range application.Providers {
		if provider.Provider == nil {
			continue
		}
		if provider.Provider.Category == "Captcha" {
			return GetCaptchaProviderByOwnerName(fmt.Sprintf("%s/%s", provider.Provider.Owner, provider.Provider.Name), lang)
		}
	}
	return nil, nil
}
