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
	"fmt"
	"strings"

	"github.com/casdoor/notify"
	"github.com/casdoor/notify/service/rocketchat"
)

func NewRocketChatProvider(clientId string, clientSecret string, endpoint string, channelName string) (notify.Notifier, error) {
	parts := strings.Split(endpoint, "://")

	var scheme, serverURL string
	if len(parts) >= 2 {
		scheme = parts[0]
		serverURL = parts[1]
	} else {
		return nil, fmt.Errorf("parse endpoint error")
	}

	rocketChatSrv, err := rocketchat.New(serverURL, scheme, clientId, clientSecret)
	if err != nil {
		return nil, err
	}

	rocketChatSrv.AddReceivers(channelName)

	notifier := notify.New()
	notifier.UseServices(rocketChatSrv)

	return notifier, nil
}
