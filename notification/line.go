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
	"github.com/casdoor/casdoor/v2/proxy"
	"github.com/casdoor/notify"
	"github.com/casdoor/notify/service/line"
)

func NewLineProvider(channelSecret string, accessToken string, receiver string) (*notify.Notify, error) {
	lineSrv, _ := line.NewWithHttpClient(channelSecret, accessToken, proxy.ProxyHttpClient)

	lineSrv.AddReceivers(receiver)

	notifier := notify.New()
	notifier.UseServices(lineSrv)

	return notifier, nil
}
