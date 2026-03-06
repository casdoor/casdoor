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

	if provider.Type == sender.HuaweiCloud || provider.Type == sender.AzureACS {
		client, err = sender.NewSmsClient(provider.Type, provider.ClientId, provider.ClientSecret, provider.SignName, provider.TemplateCode, provider.ProviderUrl, provider.AppId)
	} else if provider.Type == "Custom HTTP SMS" {
		client, err = newHttpSmsClient(provider.Endpoint, provider.Method, provider.Title, provider.TemplateCode, provider.HttpHeaders, provider.UserMapping, provider.IssuerUrl, provider.EnableProxy)
	} else if provider.Type == "Alibaba Cloud PNVS SMS" {
		client, err = newPnvsSmsClient(provider.ClientId, provider.ClientSecret, provider.SignName, provider.TemplateCode, provider.RegionId)
	} else if provider.Type == sender.Twilio {
		// For Twilio, the message body is pre-formatted in SendSms using the template.
		// Pass "%s" as the template so go-sms-sender's fmt.Sprintf passes the content through unchanged.
		client, err = sender.NewSmsClient(provider.Type, provider.ClientId, provider.ClientSecret, provider.SignName, "%s", provider.AppId)
	} else {
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

	if provider.Type == sender.Twilio {
		if provider.AppId != "" {
			phoneNumbers = append([]string{provider.AppId}, phoneNumbers...)
		}
		// Pre-format the message body using the provider's template.
		// If the template contains "%s", substitute the code; otherwise use the content (code) directly.
		if strings.Contains(provider.TemplateCode, "%s") {
			content = strings.Replace(provider.TemplateCode, "%s", content, 1)
		}
	} else if provider.Type == sender.Aliyun || provider.Type == "Alibaba Cloud PNVS SMS" {
		for i, number := range phoneNumbers {
			phoneNumbers[i] = strings.TrimPrefix(number, "+86")
		}
	}

	params := map[string]string{}
	if provider.Type == sender.TencentCloud {
		params["0"] = content
	} else {
		params["code"] = content
		if provider.Type == "Alibaba Cloud PNVS SMS" {
			timeoutInMinutes, err := conf.GetConfigInt64("verificationCodeTimeout")
			if err != nil || timeoutInMinutes <= 0 {
				timeoutInMinutes = 10
			}
			params["min"] = strconv.FormatInt(timeoutInMinutes, 10)
		}
	}

	err = client.SendMessage(params, phoneNumbers...)
	return err
}
