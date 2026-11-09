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
	"fmt"
	"net/smtp"
	"strings"

	"github.com/casdoor/gomail/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const defaultXOauth2Scope = "https://outlook.office365.com/.default"

type XOauth2EmailProvider struct {
	ClientId     string
	ClientSecret string
	Host         string
	Port         int
	SSLMode      string
	Scopes       []string
	TokenURI     string
}

func NewXOauth2EmailProvider(clientId string, clientSecret string, host string, port int, sslMode string, scopes string, tokenUrl string) *XOauth2EmailProvider {
	clientId = strings.TrimSpace(clientId)
	clientSecret = strings.TrimSpace(clientSecret)
	tokenUrl = strings.TrimSpace(tokenUrl)

	scopeList := strings.Fields(scopes)
	if len(scopeList) == 0 {
		scopeList = []string{defaultXOauth2Scope}
	}

	return &XOauth2EmailProvider{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Host:         host,
		Port:         port,
		SSLMode:      sslMode,
		Scopes:       scopeList,
		TokenURI:     tokenUrl,
	}
}

type xoauth2Auth struct {
	username    string
	accessToken string
}

func (a *xoauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	resp := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.username, a.accessToken)
	return "XOAUTH2", []byte(resp), nil
}

func (a *xoauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("unexpected server challenge: %s", string(fromServer))
	}
	return nil, nil
}

func (s *XOauth2EmailProvider) Send(fromAddress string, fromName string, toAddresses []string, subject string, content string) error {
	if s.TokenURI == "" {
		return fmt.Errorf("XOauth2 token URL is required (e.g. https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token)")
	}

	cfg := &clientcredentials.Config{
		ClientID:     s.ClientId,
		ClientSecret: s.ClientSecret,
		TokenURL:     s.TokenURI,
		Scopes:       s.Scopes,
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	token, err := cfg.TokenSource(context.Background()).Token()
	if err != nil {
		return err
	}

	dialer := gomail.NewDialer(s.Host, s.Port, fromAddress, "")
	dialer.Auth = &xoauth2Auth{username: fromAddress, accessToken: token.AccessToken}

	switch s.SSLMode {
	case "Enable":
		dialer.SSL = true
	case "Disable":
		dialer.SSL = false
	}

	message := gomail.NewMessage()
	message.SetAddressHeader("From", fromAddress, fromName)
	var addresses []string
	for _, address := range toAddresses {
		addresses = append(addresses, message.FormatAddress(address, ""))
	}
	message.SetHeader("To", addresses...)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", content)
	message.SkipUsernameCheck = true

	return dialer.DialAndSend(message)
}
