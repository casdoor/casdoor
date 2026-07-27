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

package email

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casdoor/casdoor/proxy"
	"github.com/stretchr/testify/assert"
)

func TestHttpEmailProviderSendForwardsHttpHeaders(t *testing.T) {
	proxy.InitHttpClient()

	scenarios := []struct {
		description string
		method      string
		contentType string
	}{
		{"Should forward the headers on a form POST", http.MethodPost, ""},
		{"Should forward the headers on a JSON POST", http.MethodPost, "application/json"},
		{"Should forward the headers on a GET", http.MethodGet, ""},
	}

	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			var received http.Header
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received = r.Header.Clone()
			}))
			defer server.Close()

			httpHeaders := map[string]string{"Accept-Language": "ru", "X-Token": "123"}
			provider := NewHttpEmailProvider(server.URL, scenery.method, httpHeaders, nil, scenery.contentType)

			err := provider.Send("from@example.com", "Casdoor", []string{"to@example.com"}, "subject", "content")
			assert.Nil(t, err)

			assert.Equal(t, "ru", received.Get("Accept-Language"))
			assert.Equal(t, "123", received.Get("X-Token"))
		})
	}
}

func TestHttpEmailProviderSendUnsupportedMethod(t *testing.T) {
	provider := NewHttpEmailProvider("http://localhost", http.MethodPatch, nil, nil, "")

	err := provider.Send("from@example.com", "Casdoor", []string{"to@example.com"}, "subject", "content")
	assert.EqualError(t, err, "HttpEmailProvider's Send() error, unsupported method: PATCH")
}
