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

import "github.com/casdoor/notify"

func GetNotificationProvider(typ string, clientId string, clientSecret string, clientId2 string, clientSecret2 string, appId string, receiver string, method string, title string, metaData string) (notify.Notifier, error) {
	if typ == "Telegram" {
		return NewTelegramProvider(clientSecret, receiver)
	} else if typ == "Custom HTTP" {
		return NewCustomHttpProvider(receiver, method, title)
	} else if typ == "DingTalk" {
		return NewDingTalkProvider(clientId, clientSecret)
	} else if typ == "Lark" {
		return NewLarkProvider(clientSecret)
	} else if typ == "Microsoft Teams" {
		return NewMicrosoftTeamsProvider(clientSecret)
	} else if typ == "Bark" {
		return NewBarkProvider(clientSecret)
	} else if typ == "Pushover" {
		return NewPushoverProvider(clientSecret, receiver)
	} else if typ == "Pushbullet" {
		return NewPushbulletProvider(clientSecret, receiver)
	} else if typ == "Slack" {
		return NewSlackProvider(clientSecret, receiver)
	} else if typ == "Webpush" {
		return NewWebpushProvider(clientId, clientSecret, receiver)
	} else if typ == "Discord" {
		return NewDiscordProvider(clientSecret, receiver)
	} else if typ == "Google Chat" {
		return NewGoogleChatProvider(metaData)
	} else if typ == "Line" {
		return NewLineProvider(clientSecret, appId, receiver)
	} else if typ == "Matrix" {
		return NewMatrixProvider(clientId, clientSecret, appId, receiver)
	} else if typ == "Twitter" {
		return NewTwitterProvider(clientId, clientSecret, clientId2, clientSecret2, receiver)
	} else if typ == "Reddit" {
		return NewRedditProvider(clientId, clientSecret, clientId2, clientSecret2, receiver)
	} else if typ == "Rocket Chat" {
		return NewRocketChatProvider(clientId, clientSecret, appId, receiver)
	} else if typ == "Viber" {
		return NewViberProvider(clientId, clientSecret, appId, receiver)
	}

	return nil, nil
}
