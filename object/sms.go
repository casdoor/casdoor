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
	"github.com/astaxie/beego"
	"github.com/casdoor/go-sms-sender"
)

var client go_sms_sender.SmsClient
var provider string

func InitSmsClient() {
	provider = beego.AppConfig.String("smsProvider")
	accessId := beego.AppConfig.String("smsAccessId")
	accessKey := beego.AppConfig.String("smsAccessKey")
	appId := beego.AppConfig.String("smsAppId")
	sign := beego.AppConfig.String("smsSign")
	region := beego.AppConfig.String("smsRegion")
	templateId := beego.AppConfig.String("smsTemplateId")
	client = go_sms_sender.NewSmsClient(provider, accessId, accessKey, sign, region, templateId, appId)
}

func SendCodeToPhone(phone, code string) string {
	if client == nil {
		InitSmsClient()
		if client == nil {
			return "SMS config error"
		}
	}

	param := make(map[string]string)
	if provider == "tencent" {
		param["0"] = code
	} else {
		param["code"] = code
	}
	client.SendMessage(param, phone)
	return ""
}
