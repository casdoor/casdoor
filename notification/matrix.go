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
	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/notify"
	"github.com/casdoor/notify/service/matrix"
	"maunium.net/go/mautrix/id"
)

func NewMatrixProvider(userId string, accessToken string, roomId string, homeServer string) (*notify.Notify, error) {
	matrixSrv, err := matrix.New(id.UserID(userId), id.RoomID(roomId), homeServer, accessToken)
	if err != nil {
		return nil, err
	}

	matrixSrv.SetHttpClient(proxy.ProxyHttpClient)

	notifier := notify.New()
	notifier.UseServices(matrixSrv)

	return notifier, nil
}
