// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package email

import (
	"fmt"

	"github.com/resend/resend-go/v3"
)

type ResendEmailProvider struct {
	Client *resend.Client
}

func NewResendEmailProvider(apiKey string) *ResendEmailProvider {
	client := resend.NewClient(apiKey)
	client.UserAgent += " Casdoor"
	return &ResendEmailProvider{Client: client}
}

func (s *ResendEmailProvider) Send(fromAddress string, fromName string, toAddresses []string, subject string, content string) error {
	from := fromAddress
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, fromAddress)
	}
	params := &resend.SendEmailRequest{
		From:    from,
		To:      toAddresses,
		Subject: subject,
		Html:    content,
	}
	if _, err := s.Client.Emails.Send(params); err != nil {
		return err
	}
	return nil
}
