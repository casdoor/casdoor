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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const MfaRememberCookieName = "casdoor_mfa_remember"

func getMfaRememberSignature(userId string, deadline string) (string, error) {
	cert, err := getCert("admin", "cert-built-in")
	if err != nil {
		return "", err
	}
	if cert == nil || cert.PrivateKey == "" {
		return "", fmt.Errorf("the cert: admin/cert-built-in is not found")
	}

	key := sha256.Sum256([]byte("mfa-remember:" + cert.PrivateKey))
	mac := hmac.New(sha256.New, key[:])
	mac.Write([]byte(userId + "|" + deadline))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func GetMfaRememberToken(user *User, deadline string) (string, error) {
	signature, err := getMfaRememberSignature(user.GetId(), deadline)
	if err != nil {
		return "", err
	}

	payload := base64.RawURLEncoding.EncodeToString([]byte(user.GetId() + "|" + deadline))
	return payload + "." + signature, nil
}

func IsMfaRemembered(user *User, token string) bool {
	userDeadline, err := time.Parse(time.RFC3339, user.MfaRememberDeadline)
	if err != nil || !userDeadline.After(time.Now()) {
		return false
	}

	payload, signature, ok := strings.Cut(token, ".")
	if !ok {
		return false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return false
	}
	userId, deadline, ok := strings.Cut(string(payloadBytes), "|")
	if !ok || userId != user.GetId() {
		return false
	}

	expectedSignature, err := getMfaRememberSignature(userId, deadline)
	if err != nil || !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return false
	}

	deadlineTime, err := time.Parse(time.RFC3339, deadline)
	return err == nil && deadlineTime.After(time.Now()) && !deadlineTime.After(userDeadline)
}
