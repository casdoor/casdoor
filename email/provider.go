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

type EmailProvider interface {
	Send(fromAddress string, fromName, toAddress string, subject string, content string) error
}

func GetEmailProvider(typ string, clientId string, clientSecret string, host string, port int, disableSsl bool, endpoint string, method string) EmailProvider {
	if typ == "Azure ACS" {
		return NewAzureACSEmailProvider(clientSecret, host)
	} else if typ == "Custom HTTP Email" {
		return NewHttpEmailProvider(endpoint, method)
	} else if typ == "SendGrid" {
		return NewSendgridEmailProvider(clientSecret, host, endpoint)
	} else {
		return NewSmtpEmailProvider(clientId, clientSecret, host, port, typ, disableSsl)
	}
}
