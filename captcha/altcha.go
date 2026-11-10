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

package captcha

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
)

const altchaChallengeTtl = 10 * time.Minute

// AltchaProvider verifies ALTCHA proof-of-work payloads. Without an endpoint,
// Casdoor issues the challenges itself and the secret key is the HMAC key;
// with one, the endpoint is a GateCHA-compatible server (".../api/v1") and the
// site key is its API key.
type AltchaProvider struct {
	endpoint string
}

func NewAltchaProvider(endpoint string) *AltchaProvider {
	return &AltchaProvider{endpoint: strings.TrimRight(strings.TrimSpace(endpoint), "/")}
}

func GetAltchaChallengeUrl(endpoint string, clientId string) string {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	return fmt.Sprintf("%s/challenge?apiKey=%s", endpoint, url.QueryEscape(clientId))
}

func CreateAltchaChallenge(secret string) (*altcha.Challenge, error) {
	if secret == "" {
		return nil, errors.New("the secret key of the ALTCHA provider is empty")
	}

	counter := 5000 + rand.IntN(5000)
	expiresAt := time.Now().Add(altchaChallengeTtl)
	challenge, err := altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm:           "PBKDF2/SHA-256",
		DeriveKey:           altcha.DeriveKeyPBKDF2(),
		HMACSignatureSecret: secret,
		Cost:                5000,
		KeyLength:           32,
		Counter:             &counter,
		ExpiresAt:           &expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (captcha *AltchaProvider) VerifyCaptcha(token, clientId, clientSecret, clientId2 string) (bool, error) {
	if token == "" {
		return false, nil
	}

	if captcha.endpoint != "" {
		return verifyAltchaByServer(captcha.endpoint, clientId, token)
	}
	return verifyAltchaLocally(token, clientSecret)
}

func verifyAltchaLocally(token string, secret string) (bool, error) {
	if secret == "" {
		return false, errors.New("the secret key of the ALTCHA provider is empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false, nil
	}

	var payload altcha.Payload
	err = json.Unmarshal(decoded, &payload)
	if err != nil {
		return false, nil
	}

	// challenges are always issued with an expiry, an unbounded one was not ours
	if payload.Challenge.Parameters.ExpiresAt == 0 {
		return false, nil
	}

	result, err := altcha.VerifySolution(altcha.VerifySolutionOptions{
		Challenge:           payload.Challenge,
		Solution:            payload.Solution,
		DeriveKey:           altcha.DeriveKeyPBKDF2(),
		HMACSignatureSecret: secret,
	})
	if err != nil || !result.Verified {
		return false, nil
	}

	return consumeAltchaChallenge(payload.Challenge.Signature, payload.Challenge.Parameters.ExpiresAt), nil
}

var altchaUsedChallenges = struct {
	sync.Mutex
	expiresAt map[string]int64
}{expiresAt: map[string]int64{}}

// consumeAltchaChallenge lets a solved challenge pass only once until it expires
func consumeAltchaChallenge(signature string, expiresAt int64) bool {
	altchaUsedChallenges.Lock()
	defer altchaUsedChallenges.Unlock()

	now := time.Now().Unix()
	for key, value := range altchaUsedChallenges.expiresAt {
		if value < now {
			delete(altchaUsedChallenges.expiresAt, key)
		}
	}

	if _, ok := altchaUsedChallenges.expiresAt[signature]; ok {
		return false
	}
	altchaUsedChallenges.expiresAt[signature] = expiresAt
	return true
}

func verifyAltchaByServer(endpoint string, apiKey string, token string) (bool, error) {
	body, err := json.Marshal(map[string]string{"payload": token})
	if err != nil {
		return false, err
	}

	verifyUrl := fmt.Sprintf("%s/verify?apiKey=%s", endpoint, url.QueryEscape(apiKey))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(verifyUrl, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, err
	}

	var verifyResp struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}
	err = json.Unmarshal(respBody, &verifyResp)
	if err != nil {
		return false, fmt.Errorf("ALTCHA server returned %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode >= http.StatusInternalServerError {
		return false, fmt.Errorf("ALTCHA server returned %s: %s", resp.Status, verifyResp.Error)
	}
	return verifyResp.Ok, nil
}
