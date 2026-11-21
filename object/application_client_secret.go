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

package object

import (
	"crypto/subtle"
	"time"
)

// IsClientSecretValid checks a client secret against the current one and, until it expires, the previous one
func (application *Application) IsClientSecretValid(clientSecret string) bool {
	if subtle.ConstantTimeCompare([]byte(application.ClientSecret), []byte(clientSecret)) == 1 {
		return true
	}

	if clientSecret == "" || application.PreviousClientSecret == "" || application.PreviousClientSecret == "***" {
		return false
	}

	expireTime, err := time.Parse(time.RFC3339, application.PreviousClientSecretExpireTime)
	if err != nil || !time.Now().Before(expireTime) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(application.PreviousClientSecret), []byte(clientSecret)) == 1
}
