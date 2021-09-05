// Copyright 2021 The casbin Authors. All Rights Reserved.
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

import "github.com/casdoor/go-sms-sender"

func SendSms(provider *Provider, phone string, code string) error {
	client, err := go_sms_sender.NewSmsClient(provider.Type, provider.ClientId, provider.ClientSecret, provider.SignName, provider.RegionId, provider.TemplateCode, provider.AppId)
	if err != nil {
		return err
	}

	param := map[string]string{}
	if provider.Type == go_sms_sender.TencentCloud {
		param["0"] = code
	} else {
		param["code"] = code
	}

	err = client.SendMessage(param, phone)
	return err
}
