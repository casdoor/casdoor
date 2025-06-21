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
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/wneessen/go-mail"
	"golang.org/x/net/proxy"
)

type SmtpEmailProvider struct {
	Client *mail.Client
}

func NewSmtpEmailProvider(userName string, password string, host string, port int, typ string, disableSsl bool) (*SmtpEmailProvider, error) {
	client, err := mail.NewClient(host, mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover), mail.WithUsername(userName), mail.WithPassword(password), mail.WithPort(port))
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	if typ == "SUBMAIL" {
		err = client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return nil, err
		}
	}

	client.SetSSL(!disableSsl)

	if strings.HasSuffix(host, ".amazonaws.com") {
		socks5Proxy := conf.GetConfigString("socks5Proxy")
		if socks5Proxy != "" {
			dialSocksProxy, err := proxy.SOCKS5("tcp", socks5Proxy, nil, proxy.Direct)
			if err != nil {
				return nil, err
			}

			dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialSocksProxy.Dial(network, addr)
			}

			err = mail.WithDialContextFunc(dialContext)(client)
			if err != nil {
				return nil, err
			}
		}
	}

	return &SmtpEmailProvider{Client: client}, nil
}

func (s *SmtpEmailProvider) Send(fromAddress string, fromName string, toAddress string, subject string, content string) error {
	message := mail.NewMsg()

	err := message.FromFormat(fromName, fromAddress)
	if err != nil {
		return err
	}

	err = message.To(toAddress)
	if err != nil {
		return err
	}

	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, content)

	return s.Client.DialAndSend(message)
}
