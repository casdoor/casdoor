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

import "github.com/casdoor/casdoor/email"

// TestSmtpServer Test the SMTP server
func TestSmtpServer(provider *Provider) error {
	smtpEmailProvider := email.NewSmtpEmailProvider(provider.ClientId, provider.ClientSecret, provider.Host, provider.Port, provider.Type, provider.DisableSsl)
	sender, err := smtpEmailProvider.Dialer.Dial()
	if err != nil {
		return err
	}
	defer sender.Close()

	return nil
}

func SendEmail(provider *Provider, title string, content string, dest string, sender string) error {
	emailProvider := email.GetEmailProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.Host, provider.Port, provider.DisableSsl, provider.Endpoint, provider.Method, provider.HttpHeaders, provider.UserMapping, provider.IssuerUrl)

	fromAddress := provider.ClientId2
	if fromAddress == "" {
		fromAddress = provider.ClientId
	}

	fromName := provider.ClientSecret2
	if fromName == "" {
		fromName = sender
	}

	return emailProvider.Send(fromAddress, fromName, dest, title, content)
}
