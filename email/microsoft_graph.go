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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/proxy"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	DefaultMicrosoftGraphScope    = "https://graph.microsoft.com/.default"
	DefaultMicrosoftGraphEndpoint = "https://graph.microsoft.com/v1.0"
)

// MicrosoftGraphEmailProvider sends with the sendMail API of Microsoft Graph and an app-only token (the
// Mail.Send application permission), for tenants where SMTP AUTH is turned off
type MicrosoftGraphEmailProvider struct {
	config      *clientcredentials.Config
	endpoint    string
	enableProxy bool
	httpClient  *http.Client
}

func NewMicrosoftGraphEmailProvider(clientId string, clientSecret string, tokenUrl string, scopes string, endpoint string, enableProxy bool) *MicrosoftGraphEmailProvider {
	scopeList := strings.Fields(scopes)
	if len(scopeList) == 0 {
		scopeList = []string{DefaultMicrosoftGraphScope}
	}

	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		endpoint = DefaultMicrosoftGraphEndpoint
	}

	return &MicrosoftGraphEmailProvider{
		config: &clientcredentials.Config{
			ClientID:     strings.TrimSpace(clientId),
			ClientSecret: strings.TrimSpace(clientSecret),
			TokenURL:     strings.TrimSpace(tokenUrl),
			Scopes:       scopeList,
			AuthStyle:    oauth2.AuthStyleInParams,
		},
		endpoint:    endpoint,
		enableProxy: enableProxy,
	}
}

// SetHttpClient replaces the client that calls Microsoft, nil keeps it
func (p *MicrosoftGraphEmailProvider) SetHttpClient(httpClient *http.Client) {
	p.httpClient = httpClient
}

func (p *MicrosoftGraphEmailProvider) getHttpClient() *http.Client {
	if p.httpClient != nil {
		return p.httpClient
	}
	if p.enableProxy {
		return proxy.ProxyHttpClient
	}
	return proxy.DefaultHttpClient
}

// GetToken fetches an app-only token, for testing the provider without sending
func (p *MicrosoftGraphEmailProvider) GetToken(mailbox string) (*oauth2.Token, error) {
	if p.config.TokenURL == "" || strings.Contains(p.config.TokenURL, "<") {
		return nil, errors.New("the token URL of the Microsoft Graph email provider is required, e.g. https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/token")
	}
	if !strings.Contains(mailbox, "@") {
		return nil, errors.New("the from address of the Microsoft Graph email provider must be the mailbox that sends the emails")
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, p.getHttpClient())
	return p.config.Token(ctx)
}

type graphEmailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

type graphRecipient struct {
	EmailAddress graphEmailAddress `json:"emailAddress"`
}

func (p *MicrosoftGraphEmailProvider) Send(fromAddress string, fromName string, toAddresses []string, subject string, content string) error {
	token, err := p.GetToken(fromAddress)
	if err != nil {
		return err
	}

	recipients := []graphRecipient{}
	for _, toAddress := range toAddresses {
		recipients = append(recipients, graphRecipient{EmailAddress: graphEmailAddress{Address: toAddress}})
	}

	message := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]string{
				"contentType": "HTML",
				"content":     content,
			},
			"from":         graphRecipient{EmailAddress: graphEmailAddress{Address: fromAddress, Name: fromName}},
			"toRecipients": recipients,
		},
		"saveToSentItems": false,
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	sendUrl := fmt.Sprintf("%s/users/%s/sendMail", p.endpoint, url.PathEscape(fromAddress))
	req, err := http.NewRequest(http.MethodPost, sendUrl, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.getHttpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("Microsoft Graph sendMail failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}
