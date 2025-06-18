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
	"log"
	"net"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/wneessen/go-mail"
	"golang.org/x/net/proxy"
)

type SmtpEmailProvider struct {
	Client *mail.Client
}

func NewSmtpEmailProvider(userName string, password string, host string, port int, typ string, disableSsl bool) *SmtpEmailProvider {
	client, err := mail.NewClient(host, mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(userName), mail.WithPassword(password), mail.WithPort(port))
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	if client == nil {
		log.Println("client is nil")
		return nil
	}

	if typ == "SUBMAIL" {
		err = client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			log.Println(err.Error())
			return nil
		}
	}

	client.SetSSL(!disableSsl)

	if strings.HasSuffix(host, ".amazonaws.com") {
		socks5Proxy := conf.GetConfigString("socks5Proxy")
		if socks5Proxy != "" {
			customDialer := func(_ context.Context, network, address string) (net.Conn, error) {
				dialSocksProxy, err := proxy.SOCKS5(network, socks5Proxy, nil, proxy.Direct)
				if err != nil {
					return nil, err
				}
				conn, err := dialSocksProxy.Dial(network, address)
				return conn, err
			}

			err = mail.WithDialContextFunc(customDialer)(client)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}

	return &SmtpEmailProvider{Client: client}
}

func (s *SmtpEmailProvider) Send(fromAddress string, fromName string, toAddress string, subject string, content string) error {
	message := mail.NewMsg()

	err := message.SetAddrHeader("From", fromAddress, fromName)
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
