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
	"regexp"
	"strings"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Provider struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk unique" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName       string            `xorm:"varchar(100)" json:"displayName"`
	Category          string            `xorm:"varchar(100)" json:"category"`
	Type              string            `xorm:"varchar(100)" json:"type"`
	SubType           string            `xorm:"varchar(100)" json:"subType"`
	Method            string            `xorm:"varchar(100)" json:"method"`
	ClientId          string            `xorm:"varchar(200)" json:"clientId"`
	ClientSecret      string            `xorm:"varchar(3000)" json:"clientSecret"`
	ClientId2         string            `xorm:"varchar(100)" json:"clientId2"`
	ClientSecret2     string            `xorm:"varchar(500)" json:"clientSecret2"`
	Cert              string            `xorm:"varchar(100)" json:"cert"`
	CustomAuthUrl     string            `xorm:"varchar(200)" json:"customAuthUrl"`
	CustomTokenUrl    string            `xorm:"varchar(200)" json:"customTokenUrl"`
	CustomUserInfoUrl string            `xorm:"varchar(200)" json:"customUserInfoUrl"`
	CustomLogo        string            `xorm:"varchar(200)" json:"customLogo"`
	Scopes            string            `xorm:"varchar(100)" json:"scopes"`
	UserMapping       map[string]string `xorm:"varchar(500)" json:"userMapping"`
	HttpHeaders       map[string]string `xorm:"varchar(500)" json:"httpHeaders"`

	Host       string `xorm:"varchar(100)" json:"host"`
	Port       int    `json:"port"`
	DisableSsl bool   `json:"disableSsl"` // If the provider type is WeChat, DisableSsl means EnableQRCode, if type is Google, it means sync phone number
	Title      string `xorm:"varchar(100)" json:"title"`
	Content    string `xorm:"varchar(2000)" json:"content"` // If provider type is WeChat, Content means QRCode string by Base64 encoding
	Receiver   string `xorm:"varchar(100)" json:"receiver"`

	RegionId     string `xorm:"varchar(100)" json:"regionId"`
	SignName     string `xorm:"varchar(100)" json:"signName"`
	TemplateCode string `xorm:"varchar(100)" json:"templateCode"`
	AppId        string `xorm:"varchar(100)" json:"appId"`

	Endpoint         string `xorm:"varchar(1000)" json:"endpoint"`
	IntranetEndpoint string `xorm:"varchar(100)" json:"intranetEndpoint"`
	Domain           string `xorm:"varchar(100)" json:"domain"`
	Bucket           string `xorm:"varchar(100)" json:"bucket"`
	PathPrefix       string `xorm:"varchar(100)" json:"pathPrefix"`

	Metadata               string `xorm:"mediumtext" json:"metadata"`
	IdP                    string `xorm:"mediumtext" json:"idP"`
	IssuerUrl              string `xorm:"varchar(100)" json:"issuerUrl"`
	EnableSignAuthnRequest bool   `json:"enableSignAuthnRequest"`
	EmailRegex             string `xorm:"varchar(200)" json:"emailRegex"`

	ProviderUrl string `xorm:"varchar(200)" json:"providerUrl"`
}

func GetMaskedProvider(provider *Provider, isMaskEnabled bool) *Provider {
	if !isMaskEnabled {
		return provider
	}

	if provider == nil {
		return nil
	}

	if provider.ClientSecret != "" {
		provider.ClientSecret = "***"
	}

	if provider.Category != "Email" {
		if provider.ClientSecret2 != "" {
			provider.ClientSecret2 = "***"
		}
	}

	return provider
}

func GetMaskedProviders(providers []*Provider, isMaskEnabled bool) []*Provider {
	if !isMaskEnabled {
		return providers
	}

	for _, provider := range providers {
		provider = GetMaskedProvider(provider, isMaskEnabled)
	}
	return providers
}

func GetProviderCount(owner, field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Where("owner = ? or owner = ? ", "admin", owner).Count(&Provider{})
}

func GetGlobalProviderCount(field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Count(&Provider{})
}

func GetProviders(owner string) ([]*Provider, error) {
	providers := []*Provider{}
	err := ormer.Engine.Where("owner = ? or owner = ? ", "admin", owner).Desc("created_time").Find(&providers, &Provider{})
	if err != nil {
		return providers, err
	}

	return providers, nil
}

func GetGlobalProviders() ([]*Provider, error) {
	providers := []*Provider{}
	err := ormer.Engine.Desc("created_time").Find(&providers)
	if err != nil {
		return providers, err
	}

	return providers, nil
}

func GetPaginationProviders(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Provider, error) {
	providers := []*Provider{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Where("owner = ? or owner = ? ", "admin", owner).Find(&providers)
	if err != nil {
		return providers, err
	}

	return providers, nil
}

func GetPaginationGlobalProviders(offset, limit int, field, value, sortField, sortOrder string) ([]*Provider, error) {
	providers := []*Provider{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&providers)
	if err != nil {
		return providers, err
	}

	return providers, nil
}

func getProvider(owner string, name string) (*Provider, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	provider := Provider{Name: name}
	existed, err := ormer.Engine.Get(&provider)
	if err != nil {
		return &provider, err
	}

	if existed {
		return &provider, nil
	} else {
		return nil, nil
	}
}

func GetProvider(id string) (*Provider, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getProvider(owner, name)
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

func UpdateProvider(id string, provider *Provider) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if p, err := getProvider(owner, name); err != nil {
		return false, err
	} else if p == nil {
		return false, nil
	}

	if provider.EmailRegex != "" {
		_, err := regexp.Compile(provider.EmailRegex)
		if err != nil {
			return false, err
		}
	}

	if name != provider.Name {
		err := providerChangeTrigger(name, provider.Name)
		if err != nil {
			return false, err
		}
	}

	session := ormer.Engine.ID(core.PK{owner, name}).AllCols()
	if provider.ClientSecret == "***" {
		session = session.Omit("client_secret")
	}
	if provider.ClientSecret2 == "***" {
		session = session.Omit("client_secret2")
	}

	if provider.Type == "Tencent Cloud COS" {
		provider.Endpoint = util.GetEndPoint(provider.Endpoint)
		provider.IntranetEndpoint = util.GetEndPoint(provider.IntranetEndpoint)
	}

	affected, err := session.Update(provider)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddProvider(provider *Provider) (bool, error) {
	if provider.Type == "Tencent Cloud COS" {
		provider.Endpoint = util.GetEndPoint(provider.Endpoint)
		provider.IntranetEndpoint = util.GetEndPoint(provider.IntranetEndpoint)
	}

	if provider.EmailRegex != "" {
		_, err := regexp.Compile(provider.EmailRegex)
		if err != nil {
			return false, err
		}
	}

	affected, err := ormer.Engine.Insert(provider)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteProvider(provider *Provider) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{provider.Owner, provider.Name}).Delete(&Provider{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func GetPaymentProvider(p *Provider) (pp.PaymentProvider, error) {
	cert := &Cert{}
	if p.Cert != "" {
		var err error
		cert, err = GetCert(util.GetId(p.Owner, p.Cert))
		if err != nil {
			return nil, err
		}

		if cert == nil {
			return nil, fmt.Errorf("the cert: %s does not exist", p.Cert)
		}
	}
	typ := p.Type
	if typ == "Dummy" {
		pp, err := pp.NewDummyPaymentProvider()
		if err != nil {
			return nil, err
		}
		return pp, nil
	} else if typ == "Alipay" {
		if p.Metadata != "" {
			// alipay provider store rootCert's name in metadata
			rootCert, err := GetCert(util.GetId(p.Owner, p.Metadata))
			if err != nil {
				return nil, err
			}
			if rootCert == nil {
				return nil, fmt.Errorf("the cert: %s does not exist", p.Metadata)
			}
			pp, err := pp.NewAlipayPaymentProvider(p.ClientId, cert.Certificate, cert.PrivateKey, rootCert.Certificate, rootCert.PrivateKey)
			if err != nil {
				return nil, err
			}
			return pp, nil
		} else {
			return nil, fmt.Errorf("the metadata of alipay provider is empty")
		}
	} else if typ == "GC" {
		return pp.NewGcPaymentProvider(p.ClientId, p.ClientSecret, p.Host), nil
	} else if typ == "WeChat Pay" {
		pp, err := pp.NewWechatPaymentProvider(p.ClientId, p.ClientSecret, p.ClientId2, cert.Certificate, cert.PrivateKey)
		if err != nil {
			return nil, err
		}
		return pp, nil
	} else if typ == "PayPal" {
		pp, err := pp.NewPaypalPaymentProvider(p.ClientId, p.ClientSecret)
		if err != nil {
			return nil, err
		}
		return pp, nil
	} else if typ == "Stripe" {
		pp, err := pp.NewStripePaymentProvider(p.ClientId, p.ClientSecret)
		if err != nil {
			return nil, err
		}
		return pp, nil
	} else if typ == "AirWallex" {
		pp, err := pp.NewAirwallexPaymentProvider(p.ClientId, p.ClientSecret)
		if err != nil {
			return nil, err
		}
		return pp, nil
	} else if typ == "Balance" {
		pp, err := pp.NewBalancePaymentProvider()
		if err != nil {
			return nil, err
		}
		return pp, nil
	} else {
		return nil, fmt.Errorf("the payment provider type: %s is not supported", p.Type)
	}
}

func (p *Provider) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}

func GetCaptchaProviderByOwnerName(applicationId, lang string) (*Provider, error) {
	owner, name := util.GetOwnerAndNameFromId(applicationId)
	provider := Provider{Owner: owner, Name: name, Category: "Captcha"}
	existed, err := ormer.Engine.Get(&provider)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, fmt.Errorf(i18n.Translate(lang, "provider:the provider: %s does not exist"), applicationId)
	}

	return &provider, nil
}

func GetCaptchaProviderByApplication(applicationId, isCurrentProvider, lang string) (*Provider, error) {
	if isCurrentProvider == "true" {
		return GetCaptchaProviderByOwnerName(applicationId, lang)
	}
	application, err := GetApplication(applicationId)
	if err != nil {
		return nil, err
	}

	if application == nil || len(application.Providers) == 0 {
		return nil, fmt.Errorf(i18n.Translate(lang, "provider:Invalid application id"))
	}
	for _, provider := range application.Providers {
		if provider.Provider == nil {
			continue
		}
		if provider.Provider.Category == "Captcha" {
			return GetCaptchaProviderByOwnerName(util.GetId(provider.Provider.Owner, provider.Provider.Name), lang)
		}
	}
	return nil, nil
}

func GetFaceIdProviderByOwnerName(applicationId, lang string) (*Provider, error) {
	owner, name := util.GetOwnerAndNameFromId(applicationId)
	provider := Provider{Owner: owner, Name: name, Category: "Face ID"}
	existed, err := ormer.Engine.Get(&provider)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, fmt.Errorf(i18n.Translate(lang, "provider:the provider: %s does not exist"), applicationId)
	}

	return &provider, nil
}

func GetFaceIdProviderByApplication(applicationId, isCurrentProvider, lang string) (*Provider, error) {
	if isCurrentProvider == "true" {
		return GetFaceIdProviderByOwnerName(applicationId, lang)
	}
	application, err := GetApplication(applicationId)
	if err != nil {
		return nil, err
	}

	if application == nil || len(application.Providers) == 0 {
		return nil, fmt.Errorf(i18n.Translate(lang, "provider:Invalid application id"))
	}
	for _, provider := range application.Providers {
		if provider.Provider == nil {
			continue
		}
		if provider.Provider.Category == "Face ID" {
			return GetFaceIdProviderByOwnerName(util.GetId(provider.Provider.Owner, provider.Provider.Name), lang)
		}
	}
	return nil, nil
}

func providerChangeTrigger(oldName string, newName string) error {
	session := ormer.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	var applications []*Application
	err = ormer.Engine.Find(&applications)
	if err != nil {
		return err
	}
	for i := 0; i < len(applications); i++ {
		providers := applications[i].Providers
		for j := 0; j < len(providers); j++ {
			if providers[j].Name == oldName {
				providers[j].Name = newName
			}
		}
		applications[i].Providers = providers
		_, err = session.Where("name=?", applications[i].Name).Update(applications[i])
		if err != nil {
			return err
		}
	}

	resource := new(Resource)
	resource.Provider = newName
	_, err = session.Where("provider=?", oldName).Update(resource)
	if err != nil {
		return err
	}

	return session.Commit()
}

func FromProviderToIdpInfo(ctx *context.Context, provider *Provider) *idp.ProviderInfo {
	providerInfo := &idp.ProviderInfo{
		Type:          provider.Type,
		SubType:       provider.SubType,
		ClientId:      provider.ClientId,
		ClientSecret:  provider.ClientSecret,
		ClientId2:     provider.ClientId2,
		ClientSecret2: provider.ClientSecret2,
		AppId:         provider.AppId,
		HostUrl:       provider.Host,
		TokenURL:      provider.CustomTokenUrl,
		AuthURL:       provider.CustomAuthUrl,
		UserInfoURL:   provider.CustomUserInfoUrl,
		UserMapping:   provider.UserMapping,
	}

	if provider.Type == "WeChat" {
		if ctx != nil && strings.Contains(ctx.Request.UserAgent(), "MicroMessenger") {
			providerInfo.ClientId = provider.ClientId2
			providerInfo.ClientSecret = provider.ClientSecret2
		}
	} else if provider.Type == "ADFS" || provider.Type == "AzureAD" || provider.Type == "AzureADB2C" || provider.Type == "Casdoor" || provider.Type == "Okta" {
		providerInfo.HostUrl = provider.Domain
	}

	return providerInfo
}
