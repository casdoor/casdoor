// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package idp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// calculateTelegramHash is a helper function to calculate the hash for test auth data
func calculateTelegramHash(authData map[string]interface{}, botToken string) string {
	var dataCheckArr []string
	for key, value := range authData {
		if key == "hash" {
			continue
		}
		valueStr := formatTelegramValue(value)
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", key, valueStr))
	}
	sort.Strings(dataCheckArr)
	dataCheckString := strings.Join(dataCheckArr, "\n")

	secretKey := sha256.Sum256([]byte(botToken))
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	return hex.EncodeToString(h.Sum(nil))
}

// formatTelegramValue formats a value according to Telegram's expectations
func formatTelegramValue(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return fmt.Sprintf("%d", int64(v))
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// TestTelegramAuthVerification tests the Telegram authentication data verification
func TestTelegramAuthVerification(t *testing.T) {
	// Test bot token
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	// Create test auth data with numeric values (as they come from JSON)
	authData := map[string]interface{}{
		"id":         float64(123456789),
		"first_name": "John",
		"username":   "johndoe",
		"auth_date":  float64(1704467532),
	}

	// Calculate and add the hash
	authData["hash"] = calculateTelegramHash(authData, botToken)

	// Test verification
	idp := NewTelegramIdProvider("", botToken, "")
	err := idp.verifyTelegramAuth(authData)
	if err != nil {
		t.Errorf("verifyTelegramAuth() failed: %v", err)
	}
}

// TestTelegramAuthVerificationWithInvalidHash tests that verification fails with wrong hash
func TestTelegramAuthVerificationWithInvalidHash(t *testing.T) {
	authData := map[string]interface{}{
		"id":         float64(123456789),
		"first_name": "John",
		"username":   "johndoe",
		"auth_date":  float64(1704467532),
		"hash":       "invalid_hash_value",
	}

	idp := NewTelegramIdProvider("", "123456:test", "")
	err := idp.verifyTelegramAuth(authData)
	if err == nil {
		t.Error("verifyTelegramAuth() should fail with invalid hash")
	}
	if err.Error() != "data verification failed" {
		t.Errorf("Expected error 'data verification failed', got '%v'", err)
	}
}

// TestTelegramGetToken tests the complete GetToken flow
func TestTelegramGetToken(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

	// Create test auth data
	authData := map[string]interface{}{
		"id":         float64(123456789),
		"first_name": "John",
		"username":   "johndoe",
		"auth_date":  float64(1704467532),
	}

	// Calculate and add the hash
	authData["hash"] = calculateTelegramHash(authData, botToken)

	// Encode as JSON
	authDataJSON, _ := json.Marshal(authData)

	// Test GetToken
	idp := NewTelegramIdProvider("", botToken, "")
	token, err := idp.GetToken(string(authDataJSON))
	if err != nil {
		t.Errorf("GetToken() failed: %v", err)
	}
	if token == nil {
		t.Error("GetToken() returned nil token")
	}
	if token != nil && token.AccessToken != "telegram_123456789" {
		t.Errorf("Expected access token 'telegram_123456789', got '%s'", token.AccessToken)
	}
}

// TestTelegramNumericFormatting tests that numeric values are formatted correctly
func TestTelegramNumericFormatting(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"small integer", float64(123), "123"},
		{"large integer", float64(123456789), "123456789"},
		{"timestamp", float64(1704467532), "1704467532"},
		{"string", "test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTelegramValue(tt.value)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}

			// Ensure it doesn't use scientific notation
			if strings.Contains(result, "e+") || strings.Contains(result, "E+") {
				t.Errorf("Result contains scientific notation: %s", result)
			}
		})
	}
}
