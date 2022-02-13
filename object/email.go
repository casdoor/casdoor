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

import "github.com/go-gomail/gomail"

func SendEmail(provider *Provider, title string, content string, dest string, sender string) error {
	dialer := gomail.NewDialer(provider.Host, provider.Port, provider.ClientId, provider.ClientSecret)

	message := gomail.NewMessage()
	message.SetAddressHeader("From", provider.ClientId, sender)
	message.SetHeader("To", dest)
	message.SetHeader("Subject", title)
	message.SetBody("text/html", content)

	return dialer.DialAndSend(message)
}
