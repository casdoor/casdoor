// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

package idv

import (
"bytes"
"encoding/json"
"fmt"
"io"
"net/http"
)

type JumioIdvProvider struct {
ClientId     string
ClientSecret string
Endpoint     string
}

func NewJumioIdvProvider(clientId string, clientSecret string, endpoint string) *JumioIdvProvider {
return &JumioIdvProvider{
ClientId:     clientId,
ClientSecret: clientSecret,
Endpoint:     endpoint,
}
}

func (p *JumioIdvProvider) VerifyIdentification(idCardType string, idCard string, realName string) (bool, error) {
if p.Endpoint == "" {
p.Endpoint = "https://netverify.com/api/netverify/v2/performNetverify"
}

requestBody := map[string]interface{}{
"type":     idCardType,
"number":   idCard,
"name":     realName,
"country":  "US",
"callback": "https://yourserver.com/callback",
}

jsonData, err := json.Marshal(requestBody)
if err != nil {
return false, err
}

req, err := http.NewRequest("POST", p.Endpoint, bytes.NewBuffer(jsonData))
if err != nil {
return false, err
}

req.Header.Set("Content-Type", "application/json")
req.Header.Set("Accept", "application/json")
req.SetBasicAuth(p.ClientId, p.ClientSecret)

client := &http.Client{}
resp, err := client.Do(req)
if err != nil {
return false, err
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
return false, err
}

if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
return false, fmt.Errorf("jumio API error: %d, %s", resp.StatusCode, string(body))
}

var result map[string]interface{}
err = json.Unmarshal(body, &result)
if err != nil {
return false, err
}

if status, ok := result["status"].(string); ok {
return status == "APPROVED" || status == "SUCCESS", nil
}

return true, nil
}
