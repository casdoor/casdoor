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

package object

import (
	"context"

	"github.com/casdoor/casdoor/notification"
	"github.com/casdoor/notify"
)

func getNotificationClient(provider *Provider) (notify.Notifier, error) {
	var client notify.Notifier
	client, err := notification.GetNotificationProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.ClientId2, provider.ClientSecret2, provider.AppId, provider.Receiver, provider.Method, provider.Title, provider.Metadata, provider.RegionId)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func SendNotification(provider *Provider, content string) error {
	client, err := getNotificationClient(provider)
	if err != nil {
		return err
	}

	err = client.Send(context.Background(), "", content)
	return err
}
