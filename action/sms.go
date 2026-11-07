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

package action

type sendSmsReq struct {
	Params       map[string]string `json:"params"`
	PhoneNumbers []string          `json:"phoneNumbers"`
}

// SmsClient is an SMS provider whose code is an action module from the Marketplace.
type SmsClient struct {
	name   string
	module []byte
	hosts  []string
	config *Config
}

func NewSmsClient(name string, module []byte, hosts []string, config *Config) *SmsClient {
	return &SmsClient{name: name, module: module, hosts: hosts, config: config}
}

func (c *SmsClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	return Call(c.name, c.module, c.hosts, "sms.send", c.config, &sendSmsReq{Params: param, PhoneNumbers: targetPhoneNumber}, nil)
}
