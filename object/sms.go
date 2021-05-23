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

import (
	"fmt"

	"github.com/casdoor/go-sms-sender"
)

func SendCodeToPhone(provider *Provider, phone, code string) string {
	client := go_sms_sender.NewSmsClient(provider.Type, provider.ClientId, provider.ClientSecret, provider.SignName, provider.RegionId, provider.TemplateCode, provider.AppId)
	if client == nil {
		return fmt.Sprintf("Unsupported provide type: %s", provider.Type)
	}

	param := make(map[string]string)
	if provider.Type == "tencent" {
		param["0"] = code
	} else {
		param["code"] = code
	}
	client.SendMessage(param, phone)
	return ""
}
