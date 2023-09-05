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
	"github.com/casdoor/notify"
	"github.com/casdoor/notify/service/dingding"
)

func NewDingTalkProvider(token string, secret string) (notify.Notifier, error) {
	cfg := dingding.Config{
		Token:  token,
		Secret: secret,
	}
	dingtalkSrv := dingding.New(&cfg)

	notifier := notify.New()
	notifier.UseServices(dingtalkSrv)

	return notifier, nil
}
