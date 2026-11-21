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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/gomail/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const GmailScope = "https://mail.google.com/"

// GmailEmailProvider sends through the SMTP of Gmail with XOAUTH2, so no app password is needed. The
// credentials are a Google JSON key: a service account with domain-wide delegation, which signs in as the
// from address of a Google Workspace domain, or an "authorized_user" one (client ID, secret and refresh
// token) of the mailbox itself, which also works for a personal Gmail account.
type GmailEmailProvider struct {
	smtp            *SmtpEmailProvider
	credentialsJson string
	enableProxy     bool
	httpClient      *http.Client
}

func NewGmailEmailProvider(credentialsJson string, host string, port int, sslMode string, enableProxy bool) *GmailEmailProvider {
	if host == "" {
		host = "smtp.gmail.com"
	}
	if port == 0 {
		port = 587
	}

	return &GmailEmailProvider{
		smtp:            NewSmtpEmailProvider("", "", host, port, "Gmail", sslMode, enableProxy),
		credentialsJson: strings.TrimSpace(credentialsJson),
		enableProxy:     enableProxy,
	}
}

// SetHttpClient replaces the client that fetches the token, nil keeps it
func (p *GmailEmailProvider) SetHttpClient(httpClient *http.Client) {
	p.httpClient = httpClient
}

func (p *GmailEmailProvider) getHttpClient() *http.Client {
	if p.httpClient != nil {
		return p.httpClient
	}
	if p.enableProxy {
		return proxy.ProxyHttpClient
	}
	return proxy.DefaultHttpClient
}

func (p *GmailEmailProvider) authorize(mailbox string) error {
	if !strings.Contains(mailbox, "@") {
		return errors.New("the from address of the Gmail email provider must be the mailbox that sends the emails")
	}

	var key struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(p.credentialsJson), &key); err != nil {
		return fmt.Errorf("the credentials of the Gmail email provider must be a Google JSON key: %w", err)
	}

	credType := google.CredentialsType(key.Type)
	if credType != google.ServiceAccount && credType != google.AuthorizedUser {
		return fmt.Errorf("the credentials of the Gmail email provider must be a %q or %q JSON key, got %q", google.ServiceAccount, google.AuthorizedUser, key.Type)
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, p.getHttpClient())
	params := google.CredentialsParams{Scopes: []string{GmailScope}}
	if credType == google.ServiceAccount {
		params.Subject = mailbox
	}
	credentials, err := google.CredentialsFromJSONWithTypeAndParams(ctx, []byte(p.credentialsJson), credType, params)
	if err != nil {
		return err
	}

	token, err := credentials.TokenSource.Token()
	if err != nil {
		return err
	}

	p.smtp.Dialer.Username = mailbox
	p.smtp.Dialer.Auth = &xoauth2Auth{username: mailbox, accessToken: token.AccessToken, host: p.smtp.Dialer.Host}
	return nil
}

// Dial signs in as the mailbox without sending, for testing the provider
func (p *GmailEmailProvider) Dial(mailbox string) (gomail.SendCloser, error) {
	if err := p.authorize(mailbox); err != nil {
		return nil, err
	}
	return p.smtp.Dialer.Dial()
}

func (p *GmailEmailProvider) Send(fromAddress string, fromName string, toAddresses []string, subject string, content string) error {
	if err := p.authorize(fromAddress); err != nil {
		return err
	}
	return p.smtp.Send(fromAddress, fromName, toAddresses, subject, content)
}
