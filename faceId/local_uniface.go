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

package faceId

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LocalUniFaceProvider struct {
	Endpoint string
	ApiKey   string
	Client   *http.Client
}

type localUniFaceCompareRequest struct {
	ImageA string `json:"imageA"`
	ImageB string `json:"imageB"`
}

type localUniFaceCompareResponse struct {
	Matched bool    `json:"matched"`
	Score   float64 `json:"score"`
	Reason  string  `json:"reason"`
	Detail  string  `json:"detail"`
}

type LocalUniFaceFace struct {
	Confidence float64   `json:"confidence"`
	Bbox       []float64 `json:"bbox"`
}

type localUniFaceDetectRequest struct {
	Image string `json:"image"`
}

type localUniFaceDetectResponse struct {
	Faces []LocalUniFaceFace `json:"faces"`
}

func NewLocalUniFaceProvider(endpoint string, apiKey string) *LocalUniFaceProvider {
	return &LocalUniFaceProvider{
		Endpoint: strings.TrimRight(endpoint, "/"),
		ApiKey:   apiKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (provider *LocalUniFaceProvider) Check(base64ImageA string, base64ImageB string) (bool, error) {
	if provider.Endpoint == "" {
		return false, fmt.Errorf("Local UniFace endpoint is empty")
	}

	body, err := json.Marshal(localUniFaceCompareRequest{
		ImageA: base64ImageA,
		ImageB: base64ImageB,
	})
	if err != nil {
		return false, err
	}

	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/compare", provider.Endpoint), bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	request.Header.Set("Content-Type", "application/json")
	if provider.ApiKey != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.ApiKey))
	}

	client := provider.Client
	if client == nil {
		client = http.DefaultClient
	}

	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return false, fmt.Errorf("Local UniFace compare failed with status %d: %s", response.StatusCode, string(responseBody))
	}

	var compareResponse localUniFaceCompareResponse
	if err = json.Unmarshal(responseBody, &compareResponse); err != nil {
		return false, err
	}

	return compareResponse.Matched, nil
}

func (provider *LocalUniFaceProvider) Detect(base64Image string) ([]LocalUniFaceFace, error) {
	if provider.Endpoint == "" {
		return nil, fmt.Errorf("Local UniFace endpoint is empty")
	}

	body, err := json.Marshal(localUniFaceDetectRequest{Image: base64Image})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/detect", provider.Endpoint), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if provider.ApiKey != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.ApiKey))
	}

	client := provider.Client
	if client == nil {
		client = http.DefaultClient
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Local UniFace detect failed with status %d: %s", response.StatusCode, string(responseBody))
	}

	var detectResponse localUniFaceDetectResponse
	if err = json.Unmarshal(responseBody, &detectResponse); err != nil {
		return nil, err
	}

	return detectResponse.Faces, nil
}
