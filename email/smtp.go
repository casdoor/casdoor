// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package email

import (
	"crypto/tls"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/gomail/v2"
)

type SmtpEmailProvider struct {
	Dialer *gomail.Dialer
}

func NewSmtpEmailProvider(userName string, password string, host string, port int, typ string, disableSsl bool) *SmtpEmailProvider {
	dialer := gomail.NewDialer(host, port, userName, password)
	if typ == "SUBMAIL" {
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	dialer.SSL = !disableSsl

	if strings.HasSuffix(host, ".amazonaws.com") {
		socks5Proxy := conf.GetConfigString("socks5Proxy")
		if socks5Proxy != "" {
			dialer.SetSocks5Proxy(socks5Proxy)
		}
	}

	return &SmtpEmailProvider{Dialer: dialer}
}

func (s *SmtpEmailProvider) Send(fromAddress string, fromName string, toAddress string, subject string, content string) error {
	message := gomail.NewMessage()

	message.SetAddressHeader("From", fromAddress, fromName)
	message.SetHeader("To", toAddress)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", content)

	message.SkipUsernameCheck = true
	return s.Dialer.DialAndSend(message)
}
