// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/gomail/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const DefaultMicrosoft365Scope = "https://outlook.office365.com/.default"

// Microsoft365EmailProvider sends through the SMTP of Exchange Online with XOAUTH2 and an app-only token
type Microsoft365EmailProvider struct {
	smtp        *SmtpEmailProvider
	config      *clientcredentials.Config
	enableProxy bool
	httpClient  *http.Client
}

func NewMicrosoft365EmailProvider(clientId string, clientSecret string, host string, port int, sslMode string, enableProxy bool, tokenUrl string, scopes string) *Microsoft365EmailProvider {
	scopeList := strings.Fields(scopes)
	if len(scopeList) == 0 {
		scopeList = []string{DefaultMicrosoft365Scope}
	}

	return &Microsoft365EmailProvider{
		smtp: NewSmtpEmailProvider("", "", host, port, "Microsoft 365", sslMode, enableProxy),
		config: &clientcredentials.Config{
			ClientID:     strings.TrimSpace(clientId),
			ClientSecret: strings.TrimSpace(clientSecret),
			TokenURL:     strings.TrimSpace(tokenUrl),
			Scopes:       scopeList,
			AuthStyle:    oauth2.AuthStyleInParams,
		},
		enableProxy: enableProxy,
	}
}

// SetHttpClient replaces the client that fetches the token, nil keeps it
func (p *Microsoft365EmailProvider) SetHttpClient(httpClient *http.Client) {
	p.httpClient = httpClient
}

func (p *Microsoft365EmailProvider) getHttpClient() *http.Client {
	if p.httpClient != nil {
		return p.httpClient
	}
	if p.enableProxy {
		return proxy.ProxyHttpClient
	}
	return proxy.DefaultHttpClient
}

func (p *Microsoft365EmailProvider) authorize(mailbox string) error {
	if p.config.TokenURL == "" || strings.Contains(p.config.TokenURL, "<") {
		return errors.New("the token URL of the Microsoft 365 email provider is required, e.g. https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/token")
	}
	if !strings.Contains(mailbox, "@") {
		return errors.New("the from address of the Microsoft 365 email provider must be the mailbox that sends the emails")
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, p.getHttpClient())
	token, err := p.config.Token(ctx)
	if err != nil {
		return err
	}

	p.smtp.Dialer.Username = mailbox
	p.smtp.Dialer.Auth = &xoauth2Auth{username: mailbox, accessToken: token.AccessToken, host: p.smtp.Dialer.Host}
	return nil
}

// Dial signs in as the mailbox without sending, for testing the provider
func (p *Microsoft365EmailProvider) Dial(mailbox string) (gomail.SendCloser, error) {
	if err := p.authorize(mailbox); err != nil {
		return nil, err
	}
	return p.smtp.Dialer.Dial()
}

func (p *Microsoft365EmailProvider) Send(fromAddress string, fromName string, toAddresses []string, subject string, content string) error {
	if err := p.authorize(fromAddress); err != nil {
		return err
	}
	return p.smtp.Send(fromAddress, fromName, toAddresses, subject, content)
}

type xoauth2Auth struct {
	username    string
	accessToken string
	host        string
}

func (a *xoauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && server.Name != "localhost" && server.Name != "127.0.0.1" && server.Name != "::1" {
		return "", nil, errors.New("unencrypted connection")
	}
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}

	resp := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.username, a.accessToken)
	return "XOAUTH2", []byte(resp), nil
}

func (a *xoauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return []byte{}, nil
	}
	return nil, nil
}
