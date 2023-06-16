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
	"bytes"
	"strings"

	"github.com/casdoor/casdoor/proxy"
)

func (user *User) refreshAvatar() (bool, error) {
	var err error
	var fileBuffer *bytes.Buffer
	var ext string

	// Gravatar + Identicon
	if strings.Contains(user.Avatar, "Gravatar") && user.Email != "" {
		client := proxy.ProxyHttpClient
		has, err := hasGravatar(client, user.Email)
		if err != nil {
			return false, err
		}

		if has {
			fileBuffer, ext, err = getGravatarFileBuffer(client, user.Email)
			if err != nil {
				return false, err
			}
		}
	}

	if fileBuffer == nil && strings.Contains(user.Avatar, "Identicon") {
		fileBuffer, ext, err = getIdenticonFileBuffer(user.Name)
		if err != nil {
			return false, err
		}
	}

	if fileBuffer != nil {
		avatarUrl, err := getPermanentAvatarUrlFromBuffer(user.Owner, user.Name, fileBuffer, ext, true)
		if err != nil {
			return false, err
		}
		user.Avatar = avatarUrl
		return true, nil
	}

	return false, nil
}
