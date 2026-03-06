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
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/conf"
	sender "github.com/casdoor/go-sms-sender"
)

func getSmsClient(provider *Provider) (sender.SmsClient, error) {
	var client sender.SmsClient
	var err error

	switch provider.Type {
	case sender.HuaweiCloud, sender.AzureACS:
		client, err = sender.NewSmsClient(provider.Type, provider.ClientId, provider.ClientSecret, provider.SignName, provider.TemplateCode, provider.ProviderUrl, provider.AppId)
	case "Custom HTTP SMS":
		client, err = newHttpSmsClient(provider.Endpoint, provider.Method, provider.Title, provider.TemplateCode, provider.HttpHeaders, provider.UserMapping, provider.IssuerUrl, provider.EnableProxy)
	case "Alibaba Cloud PNVS SMS":
		client, err = newPnvsSmsClient(provider.ClientId, provider.ClientSecret, provider.SignName, provider.TemplateCode, provider.RegionId)
	case sender.Twilio:
		client, err = sender.NewSmsClient(provider.Type, provider.ClientId, provider.ClientSecret, provider.SignName, "%s", provider.AppId)
	default:
		// 默认初始化方式：适用于阿里云、腾讯云等大多数厂商
		client, err = sender.NewSmsClient(provider.Type, provider.ClientId, provider.ClientSecret, provider.SignName, provider.TemplateCode, provider.AppId)
	}

	if err != nil {
		return nil, err
	}
	return client, nil
}

func SendSms(provider *Provider, content string, phoneNumbers ...string) error {
	client, err := getSmsClient(provider)
	if err != nil {
		return err
	}

	if provider.Type == sender.Aliyun || provider.Type == "Alibaba Cloud PNVS SMS" {
		for i, number := range phoneNumbers {
			phoneNumbers[i] = strings.TrimPrefix(number, "+86")
		}
	}

	if provider.Type == sender.Twilio {
		if provider.AppId != "" {
			phoneNumbers = append([]string{provider.AppId}, phoneNumbers...)
		}
		if strings.Contains(provider.TemplateCode, "%s") {
			content = strings.Replace(provider.TemplateCode, "%s", content, 1)
		}
	}

	params := map[string]string{}

	switch provider.Type {
	case sender.TencentCloud:
		params["0"] = content

	case "Alibaba Cloud PNVS SMS":
		params["code"] = content
		timeoutInMinutes, err := conf.GetConfigInt64("verificationCodeTimeout")
		if err != nil || timeoutInMinutes <= 0 {
			timeoutInMinutes = 10
		}
		params["min"] = strconv.FormatInt(timeoutInMinutes, 10)

	default:
		params["code"] = content
	}

	return client.SendMessage(params, phoneNumbers...)
}