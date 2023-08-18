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

package notification

import (
	"strconv"

	"github.com/casdoor/casdoor/proxy"
	api "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/telegram"
)

func NewTelegramProvider(apiToken string, chatIdStr string) (notify.Notifier, error) {
	client, err := api.NewBotAPIWithClient(apiToken, proxy.ProxyHttpClient)
	if err != nil {
		return nil, err
	}
	t := &telegram.Telegram{}
	t.SetClient(client)

	chatId, err := strconv.ParseInt(chatIdStr, 10, 64)
	if err != nil {
		return nil, err
	}

	t.AddReceivers(chatId)

	return t, nil
}
