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
	"time"
)

// JumioIdvProvider implements the IdvProvider interface for Jumio
type JumioIdvProvider struct {
	ClientId     string
	ClientSecret string
	Endpoint     string
	httpClient   *http.Client
}

// NewJumioIdvProvider creates a new Jumio IDV provider
func NewJumioIdvProvider(clientId string, clientSecret string, endpoint string) *JumioIdvProvider {
	if endpoint == "" {
		endpoint = "https://api.jumio.com"
	}

	return &JumioIdvProvider{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VerifyIdentity initiates an ID verification with Jumio
func (p *JumioIdvProvider) VerifyIdentity(request *VerificationRequest) (*VerificationResult, error) {
	// Prepare the Jumio API request
	jumioRequest := map[string]interface{}{
		"customerInternalReference": request.IdCardNumber,
		"userReference":             fmt.Sprintf("%s_%s", request.FirstName, request.LastName),
	}

	// Add document verification workflow
	jumioRequest["workflowDefinition"] = map[string]interface{}{
		"key": 1,
		"credentials": []map[string]interface{}{
			{
				"category": "ID",
				"type": map[string]interface{}{
					"values": []string{request.IdCardType},
				},
				"country": map[string]interface{}{
					"values": []string{request.Country},
				},
			},
		},
	}

	reqBody, err := json.Marshal(jumioRequest)
	if err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to marshal request: %v", err),
		}, err
	}

	// Create HTTP request
	apiURL := fmt.Sprintf("%s/api/v1/accounts", p.Endpoint)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to create request: %v", err),
		}, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Casdoor/1.0")
	req.SetBasicAuth(p.ClientId, p.ClientSecret)

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to send request: %v", err),
		}, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to read response: %v", err),
		}, err
	}

	// Parse response
	var jumioResponse map[string]interface{}
	if err := json.Unmarshal(body, &jumioResponse); err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to parse response: %v", err),
		}, err
	}

	// Check if verification was successful
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	// Extract transaction ID
	transactionID := ""
	if id, ok := jumioResponse["account"].(map[string]interface{})["id"].(string); ok {
		transactionID = id
	} else if id, ok := jumioResponse["transactionReference"].(string); ok {
		transactionID = id
	}

	return &VerificationResult{
		Success:       true,
		Verified:      false, // Verification is async, needs to be checked later
		TransactionID: transactionID,
		Message:       "Verification initiated successfully",
		Details: map[string]string{
			"status": "PENDING",
		},
	}, nil
}

// GetVerificationStatus retrieves the status of a verification
func (p *JumioIdvProvider) GetVerificationStatus(transactionID string) (*VerificationResult, error) {
	if transactionID == "" {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  "Transaction ID is required",
		}, fmt.Errorf("transaction ID is required")
	}

	// Create HTTP request to get verification status
	apiURL := fmt.Sprintf("%s/api/v1/accounts/%s", p.Endpoint, transactionID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to create request: %v", err),
		}, err
	}

	// Set headers
	req.Header.Set("User-Agent", "Casdoor/1.0")
	req.SetBasicAuth(p.ClientId, p.ClientSecret)

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to send request: %v", err),
		}, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to read response: %v", err),
		}, err
	}

	// Parse response
	var jumioResponse map[string]interface{}
	if err := json.Unmarshal(body, &jumioResponse); err != nil {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("Failed to parse response: %v", err),
		}, err
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return &VerificationResult{
			Success:  false,
			Verified: false,
			Message:  fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	// Determine verification status
	status := "UNKNOWN"
	verified := false

	if workflow, ok := jumioResponse["workflow"].(map[string]interface{}); ok {
		if s, ok := workflow["status"].(string); ok {
			status = s
			verified = (s == "APPROVED" || s == "PASSED")
		}
	}

	return &VerificationResult{
		Success:       true,
		Verified:      verified,
		TransactionID: transactionID,
		Message:       fmt.Sprintf("Verification status: %s", status),
		Details: map[string]string{
			"status": status,
		},
	}, nil
}

// TestConnection tests the connection to Jumio API
func (p *JumioIdvProvider) TestConnection() error {
	// Test with a simple API call to check credentials
	apiURL := fmt.Sprintf("%s/api/v1/health", p.Endpoint)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %v", err)
	}

	req.Header.Set("User-Agent", "Casdoor/1.0")
	req.SetBasicAuth(p.ClientId, p.ClientSecret)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Jumio API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid credentials")
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
