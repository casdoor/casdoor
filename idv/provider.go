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
	"fmt"
)

// VerificationRequest represents an ID verification request
type VerificationRequest struct {
	// User information
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DateOfBirth string `json:"dateOfBirth"` // Format: YYYY-MM-DD
	Country     string `json:"country"`     // ISO 3166-1 alpha-3 country code

	// Document information
	IdCardType   string `json:"idCardType"` // e.g., "PASSPORT", "ID_CARD", "DRIVING_LICENSE"
	IdCardNumber string `json:"idCardNumber"`

	// Optional fields
	Address string `json:"address,omitempty"`
}

// VerificationResult represents the result of an ID verification
type VerificationResult struct {
	Success       bool              `json:"success"`
	Verified      bool              `json:"verified"`
	TransactionID string            `json:"transactionId"`
	Message       string            `json:"message"`
	Details       map[string]string `json:"details,omitempty"`
}

// IdvProvider defines the interface for ID verification providers
type IdvProvider interface {
	// VerifyIdentity verifies the identity of a user
	VerifyIdentity(request *VerificationRequest) (*VerificationResult, error)

	// GetVerificationStatus retrieves the status of a verification by transaction ID
	GetVerificationStatus(transactionID string) (*VerificationResult, error)

	// TestConnection tests the connection to the IDV provider
	TestConnection() error
}

// GetIdvProvider creates an IDV provider based on the type and configuration
func GetIdvProvider(providerType string, clientId string, clientSecret string, endpoint string) (IdvProvider, error) {
	switch providerType {
	case "Jumio":
		return NewJumioIdvProvider(clientId, clientSecret, endpoint), nil
	default:
		return nil, fmt.Errorf("unsupported IDV provider type: %s", providerType)
	}
}
