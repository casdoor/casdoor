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

// modified from https://github.com/casbin/casnode/blob/master/service/mail.go

package object

import (
	"crypto/tls"

	"github.com/casdoor/gomail/v2"
)

func getDialer(provider *Provider) *gomail.Dialer {
	dialer := &gomail.Dialer{}
	if provider.Type == "SUBMAIL" {
		dialer = gomail.NewDialer(provider.Host, provider.Port, provider.AppId, provider.ClientSecret)
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		dialer = gomail.NewDialer(provider.Host, provider.Port, provider.ClientId, provider.ClientSecret)
	}

	dialer.SSL = !provider.DisableSsl

	return dialer
}

func SendEmail(provider *Provider, title string, content string, dest string, sender string) error {
	dialer := getDialer(provider)

	message := gomail.NewMessage()
	message.SetAddressHeader("From", provider.ClientId, sender)
	message.SetHeader("To", dest)
	message.SetHeader("Subject", title)
	message.SetBody("text/html", content)

	if provider.Type == "Mailtrap" {
		message.SkipUsernameCheck = true
	}

	return dialer.DialAndSend(message)
}

// DailSmtpServer Dail Smtp server
func DailSmtpServer(provider *Provider) error {
	dialer := getDialer(provider)

	sender, err := dialer.Dial()
	if err != nil {
		return err
	}
	defer sender.Close()

	return nil
}
