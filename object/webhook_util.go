// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package object

import (
	"io"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
)

func sendWebhook(webhook *Webhook, record *casvisorsdk.Record, extendedUser *User) (int, string, error) {
	client := &http.Client{}

	type RecordEx struct {
		casvisorsdk.Record
		ExtendedUser *User `xorm:"-" json:"extendedUser"`
	}
	recordEx := &RecordEx{
		Record:       *record,
		ExtendedUser: extendedUser,
	}

	body := strings.NewReader(util.StructToJson(recordEx))

	req, err := http.NewRequest(webhook.Method, webhook.Url, body)
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Content-Type", webhook.ContentType)

	for _, header := range webhook.Headers {
		req.Header.Set(header.Name, header.Value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(bodyBytes), err
}
